// Package glib contains methods for interactions with GSuite API
package glib

import (
	"context"
	"fmt"
	"sort"

	password "github.com/sethvargo/go-password/password"
	directoryv1 "google.golang.org/api/admin/directory/v1"

	"github.com/kubermatic-labs/gman/pkg/config"
)

// ListUsers returns a list of all current users from the API
func (ds *DirectoryService) ListUsers(ctx context.Context) ([]*directoryv1.User, error) {
	users := []*directoryv1.User{}
	token := ""

	for {
		request := ds.Users.List().Customer("my_customer").OrderBy("email").PageToken(token).Context(ctx)

		response, err := request.Do()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve list of users in domain: %v", err)
		}

		users = append(users, response.Users...)

		token = response.NextPageToken
		if token == "" {
			break
		}
	}

	return users, nil
}

// GetUserEmails retrieves primary and secondary (type: work) user email addresses
func GetUserEmails(user *directoryv1.User) (string, string) {
	var primEmail string
	var secEmail string

	for _, email := range user.Emails.([]interface{}) {
		if email.(map[string]interface{})["primary"] == true {
			primEmail = fmt.Sprint(email.(map[string]interface{})["address"])
		}
		if email.(map[string]interface{})["type"] == "work" {
			secEmail = fmt.Sprint(email.(map[string]interface{})["address"])
		}
	}

	return primEmail, secEmail
}

func (ds *DirectoryService) CreateUser(ctx context.Context, user *config.User) (*directoryv1.User, error) {
	// generate a rand password
	pass, err := password.Generate(20, 5, 5, false, false)
	if err != nil {
		return nil, fmt.Errorf("unable to generate password: %v", err)
	}

	newUser := createGSuiteUserFromConfig(user)
	newUser.Password = pass
	newUser.ChangePasswordAtNextLogin = true

	createdUser, err := ds.Users.Insert(newUser).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create user: %v", err)
	}

	// err = HandleUserAliases(ds, newUser, user.Aliases)
	// if err != nil {
	// 	return err
	// }

	return createdUser, nil
}

func (ds *DirectoryService) DeleteUser(ctx context.Context, user *directoryv1.User) error {
	err := ds.Users.Delete(user.PrimaryEmail).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("unable to delete user: %v", err)
	}

	return nil
}

// UpdateUser updates the remote user with config
func (ds *DirectoryService) UpdateUser(ctx context.Context, user *config.User) (*directoryv1.User, error) {
	apiUser := createGSuiteUserFromConfig(user)

	updatedUser, err := ds.Users.Update(user.PrimaryEmail, apiUser).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to update user: %v", err)
	}

	// err = HandleUserAliases(srv, updatedUser, user.Aliases)
	// if err != nil {
	// 	return err
	// }

	return updatedUser, nil
}

// createGSuiteUserFromConfig converts a ConfigUser to (GSuite) directoryv1.User
func createGSuiteUserFromConfig(user *config.User) *directoryv1.User {
	googleUser := &directoryv1.User{
		Name: &directoryv1.UserName{
			GivenName:  user.FirstName,
			FamilyName: user.LastName,
		},
		PrimaryEmail: user.PrimaryEmail,
		OrgUnitPath:  user.OrgUnitPath,
	}

	if len(user.Phones) > 0 {
		phNums := []directoryv1.UserPhone{}
		for _, phone := range user.Phones {
			phNum := directoryv1.UserPhone{
				Value: phone,
				Type:  "home",
			}
			phNums = append(phNums, phNum)
		}
		googleUser.Phones = phNums
	}

	if user.Address != "" {
		addr := []directoryv1.UserAddress{
			{
				Formatted: user.Address,
				Type:      "home",
			},
		}
		googleUser.Addresses = addr
	}

	if user.RecoveryEmail != "" {
		googleUser.RecoveryEmail = user.RecoveryEmail
	}

	if user.RecoveryPhone != "" {
		googleUser.RecoveryPhone = user.RecoveryPhone
	}

	if user.SecondaryEmail != "" {
		workEm := []directoryv1.UserEmail{
			{
				Address: user.SecondaryEmail,
				Type:    "work",
			},
		}
		googleUser.Emails = workEm
	}

	if user.Employee != (config.Employee{}) {
		uOrg := []directoryv1.UserOrganization{
			{
				Department:  user.Employee.Department,
				Title:       user.Employee.JobTitle,
				CostCenter:  user.Employee.CostCenter,
				Description: user.Employee.Type,
			},
		}

		googleUser.Organizations = uOrg

		if user.Employee.ManagerEmail != "" {
			rel := []directoryv1.UserRelation{
				{
					Value: user.Employee.ManagerEmail,
					Type:  "manager",
				},
			}
			googleUser.Relations = rel
		}

		if user.Employee.EmployeeID != "" {
			ids := []directoryv1.UserExternalId{
				{
					Value: user.Employee.EmployeeID,
					Type:  "organization",
				},
			}
			googleUser.ExternalIds = ids
		}
	}

	if user.Location != (config.Location{}) {
		loc := []directoryv1.UserLocation{
			{
				Area:         "desk",
				BuildingId:   user.Location.Building,
				FloorName:    user.Location.Floor,
				FloorSection: user.Location.FloorSection,
				Type:         "desk",
			},
		}
		googleUser.Locations = loc
	}

	return googleUser
}

// CreateConfigUserFromGSuite converts a (GSuite) admin.User to ConfigUser
func CreateConfigUserFromGSuite(googleUser *directoryv1.User, userLicenses []config.License) config.User {
	// get emails
	primaryEmail, secondaryEmail := GetUserEmails(googleUser)

	configUser := config.User{
		FirstName:      googleUser.Name.GivenName,
		LastName:       googleUser.Name.FamilyName,
		PrimaryEmail:   primaryEmail,
		SecondaryEmail: secondaryEmail,
		OrgUnitPath:    googleUser.OrgUnitPath,
		RecoveryPhone:  googleUser.RecoveryPhone,
		RecoveryEmail:  googleUser.RecoveryEmail,
	}

	if len(googleUser.Aliases) > 0 {
		for _, alias := range googleUser.Aliases {
			configUser.Aliases = append(configUser.Aliases, string(alias))
		}
	}

	if googleUser.Phones != nil {
		for _, phone := range googleUser.Phones.([]interface{}) {
			if phoneMap, ok := phone.(map[string]interface{}); ok {
				if phoneVal, exists := phoneMap["value"]; exists {
					configUser.Phones = append(configUser.Phones, fmt.Sprint(phoneVal))
				}
			}
		}
	}

	if googleUser.ExternalIds != nil {
		for _, id := range googleUser.ExternalIds.([]interface{}) {
			if idMap, ok := id.(map[string]interface{}); ok {
				if idType := idMap["type"]; idType == "organization" {
					if orgId, exists := idMap["value"]; exists {
						configUser.Employee.EmployeeID = fmt.Sprint(orgId)
					}
				}
			}
		}
	}

	if googleUser.Organizations != nil {
		for _, org := range googleUser.Organizations.([]interface{}) {
			if orgMap, ok := org.(map[string]interface{}); ok {
				if department, exists := orgMap["department"]; exists {
					configUser.Employee.JobTitle = fmt.Sprint(department)
				}
				if title, exists := orgMap["title"]; exists {
					configUser.Employee.JobTitle = fmt.Sprint(title)
				}
				if description, exists := orgMap["description"]; exists {
					configUser.Employee.Type = fmt.Sprint(description)
				}
				if costCenter, exists := orgMap["costCenter"]; exists {
					configUser.Employee.CostCenter = fmt.Sprint(costCenter)
				}
			}
		}
	}

	if googleUser.Relations != nil {
		for _, rel := range googleUser.Relations.([]interface{}) {
			if relMap, ok := rel.(map[string]interface{}); ok {
				if relType := relMap["type"]; relType == "manager" {
					if managerEmail, exists := relMap["value"]; exists {
						configUser.Employee.ManagerEmail = fmt.Sprint(managerEmail)
					}
				}
			}
		}
	}

	if googleUser.Locations != nil {
		for _, loc := range googleUser.Locations.([]interface{}) {
			if locMap, ok := loc.(map[string]interface{}); ok {
				if buildingId, exists := locMap["buildingId"]; exists {
					configUser.Location.Building = fmt.Sprint(buildingId)
				}
				if floorName, exists := locMap["floorName"]; exists {
					configUser.Location.Floor = fmt.Sprint(floorName)
				}
				if floorSection, exists := locMap["floorSection"]; exists {
					configUser.Location.FloorSection = fmt.Sprint(floorSection)
				}
			}
		}
	}

	if googleUser.Addresses != nil {
		for _, addr := range googleUser.Addresses.([]interface{}) {

			if addrMap, ok := addr.(map[string]interface{}); ok {
				if addrType := addrMap["type"]; addrType == "home" {
					if address, exists := addrMap["formatted"]; exists {
						configUser.Address = fmt.Sprint(address)
					}
				}
			}
		}
	}

	if len(userLicenses) > 0 {
		for _, userLicense := range userLicenses {
			configUser.Licenses = append(configUser.Licenses, userLicense.Name)
		}
	}

	return configUser
}

type aliases struct {
	Aliases []struct {
		Alias        string `json:"alias"`
		PrimaryEmail string `json:"primaryEmail"`
	} `json:"aliases"`
}

func (ds *DirectoryService) GetUserAliases(ctx context.Context, user *config.User) ([]string, error) {
	data, err := ds.Users.Aliases.List(user.PrimaryEmail).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to list user aliases: %v", err)
	}

	aliases := aliases{}
	if err := convertToStruct(data, &aliases); err != nil {
		return nil, fmt.Errorf("failed to parse user aliases: %v", err)
	}

	result := []string{}
	for _, alias := range aliases.Aliases {
		result = append(result, alias.Alias)
	}

	sort.Strings(result)

	return result, nil
}

func (ds *DirectoryService) CreateUserAlias(ctx context.Context, user *config.User, alias string) error {
	newAlias := &directoryv1.Alias{
		Alias: alias,
	}

	if _, err := ds.Users.Aliases.Insert(user.PrimaryEmail, newAlias).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to create user alias: %v", err)
	}

	return nil
}

func (ds *DirectoryService) DeleteUserAlias(ctx context.Context, user *config.User, alias string) error {
	if err := ds.Users.Aliases.Delete(user.PrimaryEmail, alias).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to delete user alias: %v", err)
	}

	return nil
}
