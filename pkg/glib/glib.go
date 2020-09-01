// Package glib contains methods for interactions with GSuite API
package glib

import (
	"context"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/kubermatic-labs/gman/pkg/config"
	password "github.com/sethvargo/go-password/password"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	groupssettings "google.golang.org/api/groupssettings/v1"
	"google.golang.org/api/licensing/v1"
	"google.golang.org/api/option"
)

// NewDirectoryService() creates a client for communicating with Google Directory API,
// returns a service object authorized to perform actions in Gsuite.
func NewDirectoryService(clientSecretFile string, impersonatedUserEmail string, scopes ...string) (*admin.Service, error) {
	ctx := context.Background()

	jsonCredentials, err := ioutil.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read json credentials (clientSecretFile): %v", err)
	}

	config, err := google.JWTConfigFromJSON(jsonCredentials, scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to process credentials: %v", err)
	}
	config.Subject = impersonatedUserEmail

	ts := config.TokenSource(ctx)

	srv, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("unable to create a new Admin Service: %v", err)
	}
	return srv, nil
}

// NewGroupsService() creates a client for communicating with Google Groupssettings API,
// returns a service object authorized to perform actions in Gsuite.
func NewGroupsService(clientSecretFile string, impersonatedUserEmail string) (*groupssettings.Service, error) {
	ctx := context.Background()

	jsonCredentials, err := ioutil.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read json credentials (clientSecretFile): %v", err)
	}

	config, err := google.JWTConfigFromJSON(jsonCredentials, groupssettings.AppsGroupsSettingsScope)
	if err != nil {
		return nil, fmt.Errorf("unable to process credentials: %v", err)
	}
	config.Subject = impersonatedUserEmail

	ts := config.TokenSource(ctx)

	srv, err := groupssettings.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("unable to create a new Groupssettings Service: %v", err)
	}
	return srv, nil
}

// NewLicensingService() creates a client for communicating with Google Licensing API,
// returns a service object authorized to perform actions in Gsuite.
func NewLicensingService(clientSecretFile string, impersonatedUserEmail string) (*licensing.Service, error) {
	ctx := context.Background()

	jsonCredentials, err := ioutil.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read json credentials (clientSecretFile): %v", err)
	}

	config, err := google.JWTConfigFromJSON(jsonCredentials, licensing.AppsLicensingScope)
	if err != nil {
		return nil, fmt.Errorf("unable to process credentials: %v", err)
	}
	config.Subject = impersonatedUserEmail

	ts := config.TokenSource(ctx)

	srv, err := licensing.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("unable to create a new Licensing Service: %v", err)
	}
	return srv, nil
}

//----------------------------------------//
//   User handling                        //
//----------------------------------------//

// GetListOfUsers returns a list of all current users form the API
func GetListOfUsers(srv admin.Service) ([]*admin.User, error) {
	request, err := srv.Users.List().Customer("my_customer").OrderBy("email").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve list of users in domain: %v", err)
	}
	return request.Users, nil
}

// GetUserEmails retrieves primary and secondary (type: work) user email addresses
func GetUserEmails(user *admin.User) (string, string) {
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

// CreateUser creates a new user in GSuite via their API
func CreateUser(srv admin.Service, licensingSrv licensing.Service, user *config.UserConfig) error {
	// generate a rand password
	pass, err := password.Generate(20, 5, 5, false, false)
	if err != nil {
		return fmt.Errorf("unable to generate password: %v", err)
	}
	newUser := createGSuiteUserFromConfig(srv, user)
	newUser.Password = pass
	newUser.ChangePasswordAtNextLogin = true

	_, err = srv.Users.Insert(newUser).Do()
	if err != nil {
		return fmt.Errorf("unable to insert a user: %v", err)
	}

	err = HandleUserAliases(srv, newUser, user.Aliases)
	if err != nil {
		return err
	}

	err = HandleUserLicenses(licensingSrv, newUser, user.Licenses)
	if err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes a user in GSuite via their API
func DeleteUser(srv admin.Service, user *admin.User) error {
	err := srv.Users.Delete(user.PrimaryEmail).Do()
	if err != nil {
		return fmt.Errorf("unable to delete a user %s: %v", user.PrimaryEmail, err)
	}
	return nil
}

// UpdateUser updates the remote user with config
func UpdateUser(srv admin.Service, licensingSrv licensing.Service, user *config.UserConfig) error {
	updatedUser := createGSuiteUserFromConfig(srv, user)
	_, err := srv.Users.Update(user.PrimaryEmail, updatedUser).Do()
	if err != nil {
		return fmt.Errorf("unable to update a user %s: %v", user.PrimaryEmail, err)
	}

	err = HandleUserAliases(srv, updatedUser, user.Aliases)
	if err != nil {
		return err
	}

	err = HandleUserLicenses(licensingSrv, updatedUser, user.Licenses)
	if err != nil {
		return err
	}

	return nil
}

// HandleUserAliases provides logic for creating/deleting/updating aiases
func HandleUserAliases(srv admin.Service, googleUser *admin.User, configAliases []string) error {
	request, err := srv.Users.Aliases.List(googleUser.PrimaryEmail).Do()
	if err != nil {
		return fmt.Errorf("unable to list user aliases in GSuite: %v", err)
	}

	if len(configAliases) == 0 {
		for _, alias := range request.Aliases {
			err = srv.Users.Aliases.Delete(googleUser.PrimaryEmail, fmt.Sprint(alias.(map[string]interface{})["alias"])).Do()
			if err != nil {
				return fmt.Errorf("unable to delete user alias: %v", err)
			}
		}
	} else {
		// check aliases to delete
		for _, alias := range request.Aliases {
			found := false
			for _, configAlias := range configAliases {
				if alias.(map[string]interface{})["alias"] == configAlias {
					found = true
					break
				}
			}
			if !found {
				// delete
				err = srv.Users.Aliases.Delete(googleUser.PrimaryEmail, fmt.Sprint(alias.(map[string]interface{})["alias"])).Do()
				if err != nil {
					return fmt.Errorf("unable to delete user alias: %v", err)
				}
			}

		}
	}

	// check aliases to add
	for _, configAlias := range configAliases {
		found := false
		for _, alias := range request.Aliases {
			if alias.(map[string]interface{})["alias"] == configAlias {
				found = true
				break
			}
		}
		if !found {
			// add
			newAlias := &admin.Alias{
				Alias: configAlias,
			}
			_, err = srv.Users.Aliases.Insert(googleUser.PrimaryEmail, newAlias).Do()
			if err != nil {
				return fmt.Errorf("unable to add user alias: %v", err)
			}
		}
	}

	return nil
}

// createGSuiteUserFromConfig converts a ConfigUser to (GSuite) admin.User
func createGSuiteUserFromConfig(srv admin.Service, user *config.UserConfig) *admin.User {
	googleUser := &admin.User{
		Name: &admin.UserName{
			GivenName:  user.FirstName,
			FamilyName: user.LastName,
		},
		PrimaryEmail: user.PrimaryEmail,
		OrgUnitPath:  user.OrgUnitPath,
	}

	if len(user.Phones) > 0 {
		phNums := []admin.UserPhone{}
		for _, phone := range user.Phones {
			phNum := admin.UserPhone{
				Value: phone,
				Type:  "home",
			}
			phNums = append(phNums, phNum)
		}
		googleUser.Phones = phNums
	}

	if user.Address != "" {
		addr := []admin.UserAddress{
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
		workEm := []admin.UserEmail{
			{
				Address: user.SecondaryEmail,
				Type:    "work",
			},
		}
		googleUser.Emails = workEm
	}

	if user.Employee != (config.EmployeeConfig{}) {
		uOrg := []admin.UserOrganization{
			{
				Department:  user.Employee.Department,
				Title:       user.Employee.JobTitle,
				CostCenter:  user.Employee.CostCenter,
				Description: user.Employee.Type,
			},
		}

		googleUser.Organizations = uOrg

		if user.Employee.ManagerEmail != "" {
			rel := []admin.UserRelation{
				{
					Value: user.Employee.ManagerEmail,
					Type:  "manager",
				},
			}
			googleUser.Relations = rel
		}

		if user.Employee.EmployeeID != "" {
			ids := []admin.UserExternalId{
				{
					Value: user.Employee.EmployeeID,
					Type:  "organization",
				},
			}
			googleUser.ExternalIds = ids
		}

	}

	if user.Location != (config.LocationConfig{}) {
		loc := []admin.UserLocation{
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

// createConfigUserFromGSuite converts a (GSuite) admin.User to ConfigUser
func CreateConfigUserFromGSuite(googleUser *admin.User, userLicenses []License) config.UserConfig {
	// get emails
	primaryEmail, secondaryEmail := GetUserEmails(googleUser)

	configUser := config.UserConfig{
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
			configUser.Phones = append(configUser.Phones, fmt.Sprint(phone.(map[string]interface{})["value"]))
		}
	}

	if googleUser.ExternalIds != nil {
		for _, id := range googleUser.ExternalIds.([]interface{}) {
			if id.(map[string]interface{})["type"] == "organization" {
				configUser.Employee.EmployeeID = fmt.Sprint(id.(map[string]interface{})["value"])
			}
		}
	}

	if googleUser.Organizations != nil {
		for _, org := range googleUser.Organizations.([]interface{}) {
			if org.(map[string]interface{})["department"] != nil {
				configUser.Employee.Department = fmt.Sprint(org.(map[string]interface{})["department"])
			}
			if org.(map[string]interface{})["title"] != nil {
				configUser.Employee.JobTitle = fmt.Sprint(org.(map[string]interface{})["title"])
			}
			if org.(map[string]interface{})["description"] != nil {
				configUser.Employee.Type = fmt.Sprint(org.(map[string]interface{})["description"])
			}
			if org.(map[string]interface{})["costCenter"] != nil {
				configUser.Employee.CostCenter = fmt.Sprint(org.(map[string]interface{})["costCenter"])
			}
		}
	}

	if googleUser.Relations != nil {
		for _, rel := range googleUser.Relations.([]interface{}) {
			if rel.(map[string]interface{})["type"] == "manager" && rel.(map[string]interface{})["value"] != nil {
				configUser.Employee.ManagerEmail = fmt.Sprint(rel.(map[string]interface{})["value"])
			}
		}
	}

	if googleUser.Locations != nil {
		for _, loc := range googleUser.Locations.([]interface{}) {
			if loc.(map[string]interface{})["buildingId"] != nil {
				configUser.Location.Building = fmt.Sprint(loc.(map[string]interface{})["buildingId"])
			}
			if loc.(map[string]interface{})["floorName"] != nil {
				configUser.Location.Floor = fmt.Sprint(loc.(map[string]interface{})["floorName"])
			}
			if loc.(map[string]interface{})["floorSection"] != nil {
				configUser.Location.FloorSection = fmt.Sprint(loc.(map[string]interface{})["floorSection"])
			}
		}
	}

	if googleUser.Addresses != nil {
		for _, addr := range googleUser.Addresses.([]interface{}) {
			if addr.(map[string]interface{})["type"] == "home" {
				configUser.Address = fmt.Sprint(addr.(map[string]interface{})["formatted"])
			}
		}
	}

	if len(userLicenses) > 0 {
		for _, userLicense := range userLicenses {
			configUser.Licenses = append(configUser.Licenses, userLicense.name)
		}
	}

	return configUser
}

//----------------------------------------//
//   Group handling                       //
//----------------------------------------//

// GetListOfGroups returns a list of all current groups from the API
func GetListOfGroups(srv *admin.Service) ([]*admin.Group, error) {
	request, err := srv.Groups.List().Customer("my_customer").OrderBy("email").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve a list of groups in domain: %v", err)
	}
	return request.Groups, nil
}

// GetSettingOfGroup returns a group settings object from the API
func GetSettingOfGroup(srv *groupssettings.Service, groupId string) (*groupssettings.Groups, error) {
	request, err := srv.Groups.Get(groupId).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve group's (%s) settings: %v", groupId, err)
	}
	return request, nil
}

// CreateGroup creates a new group in GSuite via their API
func CreateGroup(srv admin.Service, grSrv groupssettings.Service, group *config.GroupConfig) error {
	newGroup, groupSettings := CreateGSuiteGroupFromConfig(group)
	_, err := srv.Groups.Insert(newGroup).Do()
	if err != nil {
		return fmt.Errorf("unable to insert a group: %v", err)
	}
	// add the members
	for _, member := range group.Members {
		AddNewMember(srv, newGroup.Email, &member)
	}
	// add the group's settings
	_, err = grSrv.Groups.Update(newGroup.Email, groupSettings).Do()
	if err != nil {
		return fmt.Errorf("unable to set the group settings: %v", err)
	}
	return nil
}

// DeleteGroup deletes a group in GSuite via their API
func DeleteGroup(srv admin.Service, group *admin.Group) error {
	err := srv.Groups.Delete(group.Email).Do()
	if err != nil {
		return fmt.Errorf("unable to delete a group: %v", err)
	}
	return nil
}

// UpdateGroup updates the remote group with config
func UpdateGroup(srv admin.Service, grSrv groupssettings.Service, group *config.GroupConfig) error {
	updatedGroup, groupSettings := CreateGSuiteGroupFromConfig(group)
	_, err := srv.Groups.Update(group.Email, updatedGroup).Do()
	if err != nil {
		return fmt.Errorf("unable to update a group: %v", err)
	}
	// update group's settings
	_, err = grSrv.Groups.Update(group.Email, groupSettings).Do()
	if err != nil {
		return fmt.Errorf("unable to update group settings: %v", err)
	}

	return nil
}

// createGSuiteGroupFromConfig converts a ConfigGroup to (GSuite) admin.Group
func CreateGSuiteGroupFromConfig(group *config.GroupConfig) (*admin.Group, *groupssettings.Groups) {
	googleGroup := &admin.Group{
		Name:  group.Name,
		Email: group.Email,
	}
	if group.Description != "" {
		googleGroup.Description = group.Description
	}

	groupSettings := &groupssettings.Groups{
		WhoCanContactOwner:   group.WhoCanContactOwner,
		WhoCanViewMembership: group.WhoCanViewMembership,
		WhoCanApproveMembers: group.WhoCanApproveMembers,
		WhoCanPostMessage:    group.WhoCanPostMessage,
		WhoCanJoin:           group.WhoCanJoin,
		IsArchived:           strconv.FormatBool(group.IsArchived),
		ArchiveOnly:          strconv.FormatBool(group.IsArchived),
		AllowExternalMembers: strconv.FormatBool(group.AllowExternalMembers),
	}

	return googleGroup, groupSettings
}

func CreateConfigGroupFromGSuite(googleGroup *admin.Group, members []*admin.Member, gSettings *groupssettings.Groups) (config.GroupConfig, error) {

	boolAllowExternalMembers, err := strconv.ParseBool(gSettings.AllowExternalMembers)
	if err != nil {
		return config.GroupConfig{}, fmt.Errorf("could not parse 'AllowExternalMembers' value from string to bool: %v", err)
	}
	boolIsArchived, err := strconv.ParseBool(gSettings.IsArchived)
	if err != nil {
		return config.GroupConfig{}, fmt.Errorf("could not parse 'IsArchived' value from string to bool: %v", err)
	}

	configGroup := config.GroupConfig{
		Name:                 googleGroup.Name,
		Email:                googleGroup.Email,
		Description:          googleGroup.Description,
		WhoCanContactOwner:   gSettings.WhoCanContactOwner,
		WhoCanViewMembership: gSettings.WhoCanViewMembership,
		WhoCanApproveMembers: gSettings.WhoCanApproveMembers,
		WhoCanPostMessage:    gSettings.WhoCanPostMessage,
		WhoCanJoin:           gSettings.WhoCanJoin,
		AllowExternalMembers: boolAllowExternalMembers,
		IsArchived:           boolIsArchived,
		Members:              []config.MemberConfig{},
	}

	for _, m := range members {
		configGroup.Members = append(configGroup.Members, config.MemberConfig{
			Email: m.Email,
			Role:  m.Role,
		})

	}

	return configGroup, nil
}

//----------------------------------------//
//   Group Member handling                //
//----------------------------------------//

// GetListOfMembers returns a list of all current group members form the API
func GetListOfMembers(srv *admin.Service, group *admin.Group) ([]*admin.Member, error) {
	request, err := srv.Members.List(group.Email).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve members in group %s: %v", group.Name, err)
	}
	return request.Members, nil
}

// AddNewMember adds a new member to a group in GSuite
func AddNewMember(srv admin.Service, groupEmail string, member *config.MemberConfig) error {
	newMember := createGSuiteGroupMemberFromConfig(member)
	_, err := srv.Members.Insert(groupEmail, newMember).Do()
	if err != nil {
		return fmt.Errorf("unable to add a member to a group: %v", err)
	}
	return nil
}

// RemoveMember removes a member from a group in Gsuite
func RemoveMember(srv admin.Service, groupEmail string, member *admin.Member) error {
	err := srv.Members.Delete(groupEmail, member.Email).Do()
	if err != nil {
		return fmt.Errorf("unable to delete a member from a group: %v", err)
	}
	return nil
}

// MemberExists checks if member exists in group
func MemberExists(srv admin.Service, group *admin.Group, member *config.MemberConfig) (bool, error) {
	exists, err := srv.Members.HasMember(group.Email, member.Email).Do()
	if err != nil {
		return false, fmt.Errorf("unable to check if member %s exists in a group %s: %v", member.Email, group.Name, err)
	}
	return exists.IsMember, nil
}

// UpdateMembership changes the role of the member
// Update(groupKey string, memberKey string, member *Member)
func UpdateMembership(srv admin.Service, groupEmail string, member *config.MemberConfig) error {
	newMember := createGSuiteGroupMemberFromConfig(member)
	_, err := srv.Members.Update(groupEmail, member.Email, newMember).Do()
	if err != nil {
		return fmt.Errorf("unable to update a member in a group: %v", err)
	}
	return nil
}

// createGSuiteGroupMemberFromConfig converts a ConfigMember to (GSuite) admin.Member
func createGSuiteGroupMemberFromConfig(member *config.MemberConfig) *admin.Member {
	googleMember := &admin.Member{
		Email: member.Email,
		Role:  member.Role,
	}
	return googleMember
}

//----------------------------------------//
//   OrgUnit handling                     //
//----------------------------------------//

// GetListOfOrgUnits returns a list of all current organizational units form the API
func GetListOfOrgUnits(srv *admin.Service) ([]*admin.OrgUnit, error) {
	request, err := srv.Orgunits.List("my_customer").Type("all").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to list OrgUnits in domain: %v", err)
	}
	return request.OrganizationUnits, nil
}

// CreateOrgUnit creates a new org unit in GSuite via their API
func CreateOrgUnit(srv admin.Service, ou *config.OrgUnitConfig) error {
	newOU := createGSuiteOUFromConfig(ou)
	_, err := srv.Orgunits.Insert("my_customer", newOU).Do()
	if err != nil {
		return fmt.Errorf("unable to create an org unit: %v", err)
	}
	return nil
}

// DeleteOrgUnit deletes a group in GSuite via their API
func DeleteOrgUnit(srv admin.Service, ou *admin.OrgUnit) error {
	// the Orgunits.Delete function takes as an argument the full org unit path, but without first slash...
	var orgUPath []string
	if ou.OrgUnitPath[0] == '/' {
		orgUPath = append([]string{}, ou.OrgUnitPath[1:])
	} else {
		orgUPath = append([]string{}, ou.OrgUnitPath)
	}

	err := srv.Orgunits.Delete("my_customer", orgUPath).Do()
	if err != nil {
		return fmt.Errorf("unable to delete an org unit: %v", err)
	}
	return nil
}

// UpdateOrgUnit updates the remote org unit with config
func UpdateOrgUnit(srv admin.Service, ou *config.OrgUnitConfig) error {
	updatedOu := createGSuiteOUFromConfig(ou)
	// the Orgunits.Update function takes as an argument the full org unit path, but without first slash...
	var orgUPath []string
	if ou.OrgUnitPath[0] == '/' {
		orgUPath = append([]string{}, ou.OrgUnitPath[1:])
	} else {
		orgUPath = append([]string{}, ou.OrgUnitPath)
	}

	_, err := srv.Orgunits.Update("my_customer", orgUPath, updatedOu).Do()
	if err != nil {
		return fmt.Errorf("unable to update an org unit: %v", err)
	}
	return nil
}

// createGSuiteOUFromConfig converts a OrgUnitConfig to (GSuite) admin.OrgUnit
func createGSuiteOUFromConfig(ou *config.OrgUnitConfig) *admin.OrgUnit {
	googleOU := &admin.OrgUnit{
		Name: ou.Name,
		//OrgUnitPath:       ou.OrgUnitPath,
		ParentOrgUnitPath: ou.ParentOrgUnitPath,
	}
	if ou.Description != "" {
		googleOU.Description = ou.Description
	}

	return googleOU
}

//----------------------------------------//
//   Licenses handling                    //
//----------------------------------------//

// GetUserLicense returns a list of licenses of a user
func GetUserLicenses(srv *licensing.Service, user string) ([]License, error) {
	var userLicenses []License
	for _, license := range googleLicenses {
		_, err := srv.LicenseAssignments.Get(license.productId, license.skuId, user).Do()
		if err != nil {
			if err.(*googleapi.Error).Code == 404 {
				// license doesnt exists
				break
			} else {
				return nil, fmt.Errorf("unable to retrieve license in domain: %v", err)
			}
		}
		userLicenses = append(userLicenses, license)
	}

	return userLicenses, nil
}

// HandleUserLicenses provides logic for creating/deleting/updating licenses according to config file
func HandleUserLicenses(srv licensing.Service, googleUser *admin.User, configLicenses []string) error {
	var userLicenses []License
	// request the list of user licenses
	for _, license := range googleLicenses {
		_, err := srv.LicenseAssignments.Get(license.productId, license.skuId, googleUser.PrimaryEmail).Do()
		if err != nil {
			// error code 404 - if the user does not have this license, the response has a 'not found' error
			if err.(*googleapi.Error).Code == 404 {
				// check if config includes given google license
				found := false
				for _, configLicense := range configLicenses {
					if configLicense == license.name {
						found = true
						break
					}
				}
				// if config includes it (found in config), add it
				if found == true {
					_, err := srv.LicenseAssignments.Insert(license.productId, license.skuId, &licensing.LicenseAssignmentInsert{UserId: googleUser.PrimaryEmail}).Do()
					if err != nil {
						return fmt.Errorf("unable to insert user license: %v", err)
					}
				}
			} else {
				// license exists in gsuite
				return fmt.Errorf("unable to retrieve user license: %v", err)
			}
		} else {
			userLicenses = append(userLicenses, license)
		}
	}

	// check licenses to delete
	if len(configLicenses) == 0 {
		for _, license := range userLicenses {
			err := srv.LicenseAssignments.Delete(license.productId, license.skuId, googleUser.PrimaryEmail).Do()
			if err != nil {
				return fmt.Errorf("unable to delete user license: %v", err)
			}
		}
	} else {
		for _, license := range userLicenses {
			found := false
			for _, configLicense := range configLicenses {
				if license.name == configLicense {
					found = true
					break
				}
			}
			if !found {
				// delete
				err := srv.LicenseAssignments.Delete(license.productId, license.skuId, googleUser.PrimaryEmail).Do()
				if err != nil {
					return fmt.Errorf("unable to delete user license: %v", err)
				}
			}

		}

	}

	return nil
}
