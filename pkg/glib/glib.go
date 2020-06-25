// Package glib contains methods for interactions with GSuite API
package glib

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/kubermatic-labs/gman/pkg/config"
	password "github.com/sethvargo/go-password/password"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

// CreateDirectoryService() creates a client for communicating with Google APIs,
// returns an Admin SDK Directory service object authorized with.
func NewDirectoryService(clientSecretFile string, impersonatedUserEmail string) (*admin.Service, error) {
	ctx := context.Background()

	jsonCredentials, err := ioutil.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("ReadFile(clientSecretFile): %v", err)
	}

	config, err := google.JWTConfigFromJSON(jsonCredentials, admin.AdminDirectoryUserScope, admin.AdminDirectoryGroupScope, admin.AdminDirectoryGroupMemberScope, admin.AdminDirectoryOrgunitScope)
	if err != nil {
		return nil, fmt.Errorf("JWTConfigFromJSON: %v", err)
	}
	config.Subject = impersonatedUserEmail

	ts := config.TokenSource(ctx)

	srv, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("NewService: %v", err)
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
		log.Fatalf("Unable to retrieve users in domain: %v", err)
		return nil, err
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

// CreateNewUser creates a new user in GSuite via their API
func CreateNewUser(srv admin.Service, user *config.UserConfig) error {
	// generate a rand password
	pass, err := password.Generate(20, 5, 5, false, false)
	if err != nil {
		log.Fatalf("Unable to generate password: %v", err)
		return err
	}
	newUser := createGSuiteUserFromConfig(user)
	newUser.Password = pass
	newUser.ChangePasswordAtNextLogin = true

	_, err = srv.Users.Insert(newUser).Do()
	if err != nil {
		log.Fatalf("Unable to create a user: %v", err)
		return err
	}
	return nil
}

// DeleteUser deletes a user in GSuite via their API
func DeleteUser(srv admin.Service, user *admin.User) error {
	err := srv.Users.Delete(user.PrimaryEmail).Do()
	if err != nil {
		log.Fatalf("Unable to delete a user: %v", err)
		return err
	}
	return nil
}

// UpdateUser updates the remote user with config
func UpdateUser(srv admin.Service, user *config.UserConfig) error {
	updatedUser := createGSuiteUserFromConfig(user)
	_, err := srv.Users.Update(user.PrimaryEmail, updatedUser).Do()
	if err != nil {
		log.Fatalf("Unable to update a user: %v", err)
		return err
	}
	return nil
}

// createGSuiteUserFromConfig converts a ConfigUser to (Gsuite) admin.User
func createGSuiteUserFromConfig(user *config.UserConfig) *admin.User {
	googleUser := &admin.User{
		Name: &admin.UserName{
			GivenName:  user.FirstName,
			FamilyName: user.LastName,
		},
		PrimaryEmail: user.PrimaryEmail,
		OrgUnitPath:  user.OrgUnitPath,
	}

	if user.SecondaryEmail != "" {
		googleUser.RecoveryEmail = user.SecondaryEmail
		workEm := &[]admin.UserEmail{
			{
				Address: user.SecondaryEmail,
				Type:    "work",
			},
		}
		googleUser.Emails = workEm
	}

	return googleUser
}

//----------------------------------------//
//   Group handling                       //
//----------------------------------------//

// GetListOfGroups returns a list of all current groups from the API
func GetListOfGroups(srv *admin.Service) ([]*admin.Group, error) {
	request, err := srv.Groups.List().Customer("my_customer").OrderBy("email").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve groups in domain: %v", err)
		return nil, err
	}
	return request.Groups, nil
}

// CreateGroup creates a new group in GSuite via their API
func CreateGroup(srv admin.Service, group *config.GroupConfig) error {
	newGroup := createGSuiteGroupFromConfig(group)
	_, err := srv.Groups.Insert(newGroup).Do()
	if err != nil {
		log.Fatalf("Unable to create a group: %v", err)
		return err
	}
	// add the members
	for _, member := range group.Members {
		AddNewMember(srv, newGroup.Email, &member)
	}
	return nil
}

// DeleteGroup deletes a group in GSuite via their API
func DeleteGroup(srv admin.Service, group *admin.Group) error {
	err := srv.Groups.Delete(group.Email).Do()
	if err != nil {
		log.Fatalf("Unable to delete a group: %v", err)
		return err
	}
	return nil
}

// UpdateGroup updates the remote group with config
func UpdateGroup(srv admin.Service, group *config.GroupConfig) error {
	updatedGroup := createGSuiteGroupFromConfig(group)
	_, err := srv.Groups.Update(group.Email, updatedGroup).Do()
	if err != nil {
		log.Fatalf("Unable to update a group: %v", err)
		return err
	}
	return nil
}

// createGSuiteGroupFromConfig converts a ConfigGroup to (Gsuite) admin.Group
func createGSuiteGroupFromConfig(group *config.GroupConfig) *admin.Group {
	googleGroup := &admin.Group{
		Name:  group.Name,
		Email: group.Email,
	}
	if group.Description != "" {
		googleGroup.Description = group.Description
	}
	return googleGroup
}

//----------------------------------------//
//   Group Member handling                //
//----------------------------------------//

// GetListOfMembers returns a list of all current group members form the API
func GetListOfMembers(srv *admin.Service, group *admin.Group) ([]*admin.Member, error) {
	request, err := srv.Members.List(group.Email).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve members in group %s: %v", group.Name, err)
		return nil, err
	}
	return request.Members, nil
}

// AddNewMember adds a new member to a group in GSuite
func AddNewMember(srv admin.Service, groupEmail string, member *config.MemberConfig) error {
	newMember := createGSuiteGroupMemberFromConfig(member)
	_, err := srv.Members.Insert(groupEmail, newMember).Do()
	if err != nil {
		log.Fatalf("Unable to add a member: %v", err)
		return err
	}
	return nil
}

// RemoveMember removes a member from a group in Gsuite
func RemoveMember(srv admin.Service, groupEmail string, member *admin.Member) error {
	err := srv.Members.Delete(groupEmail, member.Email).Do()
	if err != nil {
		log.Fatalf("Unable to delete a member: %v", err)
		return err
	}
	return nil
}

// MemberExists checks if member exists in group
func MemberExists(srv admin.Service, group *admin.Group, member *config.MemberConfig) bool {
	exists, err := srv.Members.HasMember(group.Email, member.Email).Do()
	if err != nil {
		log.Fatalf("Unable to check if member %s exists in a group %s: %v", member.Email, group.Name, err)
		return false
	}
	return exists.IsMember
}

// UpdateMembership changes the role of the member
// Update(groupKey string, memberKey string, member *Member)
func UpdateMembership(srv admin.Service, groupEmail string, member *config.MemberConfig) error {
	newMember := createGSuiteGroupMemberFromConfig(member)
	_, err := srv.Members.Update(groupEmail, member.Email, newMember).Do()
	if err != nil {
		log.Fatalf("Unable to delete a member: %v", err)
		return err
	}
	return nil
}

// createGSuiteGroupMemberFromConfig converts a ConfigMember to (Gsuite) admin.Member
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
		log.Fatalf("Unable to retrieve OrgUnits in domain: %v", err)
		return nil, err
	}
	return request.OrganizationUnits, nil
}

// CreateOU creates a new org unit in GSuite via their API
func CreateOU(srv admin.Service, ou *config.OrgUnitConfig) error {
	newOU := createGSuiteOUFromConfig(ou)
	_, err := srv.Orgunits.Insert("my_customer", newOU).Do()
	if err != nil {
		log.Fatalf("Unable to create an org unit: %v", err)
		return err
	}
	return nil
}

// DeleteGroup deletes a group in GSuite via their API
func DeleteOU(srv admin.Service, ou *admin.OrgUnit) error {
	// the Orgunits.Delete function takes as an argument the full org unit path, but without first slash...
	var orgUPath []string
	if ou.OrgUnitPath[0] == '/' {
		orgUPath = append([]string{}, ou.OrgUnitPath[1:])
	} else {
		orgUPath = append([]string{}, ou.OrgUnitPath)
	}

	err := srv.Orgunits.Delete("my_customer", orgUPath).Do()
	if err != nil {
		log.Fatalf("Unable to delete an org unit: %v", err)
		return err
	}
	return nil
}

// UpdateGroup updates the remote group with config
func UpdateOU(srv admin.Service, ou *config.OrgUnitConfig) error {
	updatedOu := createGSuiteOUFromConfig(ou)
	//orgUPath := append([]string{}, ou.OrgUnitPath)
	// the Orgunits.Update function takes as an argument the full org unit path, but without first slash...
	var orgUPath []string
	if ou.OrgUnitPath[0] == '/' {
		orgUPath = append([]string{}, ou.OrgUnitPath[1:])
	} else {
		orgUPath = append([]string{}, ou.OrgUnitPath)
	}

	_, err := srv.Orgunits.Update("my_customer", orgUPath, updatedOu).Do()
	if err != nil {
		log.Fatalf("Unable to update an org unit: %v", err)
		return err
	}
	return nil
}

// createGSuiteGroupFromConfig converts a ConfigGroup to (Gsuite) admin.Group
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
