/*
Copyright 2021 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"encoding/json"
	"fmt"
	"strconv"

	directoryv1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	groupssettingsv1 "google.golang.org/api/groupssettings/v1"

	"github.com/kubermatic-labs/gman/pkg/util"
)

type CustomSchema struct {
	PasswordHash string `json:"passwordHash"`
}

func GetUserSchema(user *directoryv1.User) *CustomSchema {
	customFields, ok := user.CustomSchemas[SchemaName]
	if !ok {
		return nil
	}

	s := &CustomSchema{}
	if err := json.Unmarshal(customFields, s); err != nil {
		return nil
	}

	return s
}

func ToGSuiteUser(user *User, enableInsecurePasswords bool) *directoryv1.User {
	gsuiteUser := &directoryv1.User{
		Name: &directoryv1.UserName{
			GivenName:  user.FirstName,
			FamilyName: user.LastName,
		},
		PrimaryEmail:  user.PrimaryEmail,
		RecoveryEmail: user.RecoveryEmail,
		RecoveryPhone: user.RecoveryPhone,
		OrgUnitPath:   user.OrgUnitPath,

		// set to empty list, because having them as "nil"
		// will not cause proper updates, i.e. orphaned phone numbers
		Phones:        []directoryv1.UserPhone{},
		Addresses:     []directoryv1.UserAddress{},
		Emails:        []directoryv1.UserEmail{},
		Organizations: []directoryv1.UserOrganization{},
		Relations:     []directoryv1.UserRelation{},
		ExternalIds:   []directoryv1.UserExternalId{},
		Locations:     []directoryv1.UserLocation{},
		CustomSchemas: map[string]googleapi.RawMessage{},
	}

	if len(user.Phones) > 0 {
		phNums := []directoryv1.UserPhone{}
		for _, phone := range user.Phones {
			phNums = append(phNums, directoryv1.UserPhone{
				Value: phone,
				Type:  "home",
			})
		}

		gsuiteUser.Phones = phNums
	}

	if user.Address != "" {
		gsuiteUser.Addresses = []directoryv1.UserAddress{
			{
				Formatted: user.Address,
				Type:      "home",
			},
		}
	}

	if !user.Employee.Empty() {
		userOrg := []directoryv1.UserOrganization{
			{
				Department:  user.Employee.Department,
				Title:       user.Employee.JobTitle,
				CostCenter:  user.Employee.CostCenter,
				Description: user.Employee.Type,
			},
		}

		gsuiteUser.Organizations = userOrg

		if user.Employee.ManagerEmail != "" {
			gsuiteUser.Relations = []directoryv1.UserRelation{
				{
					Value: user.Employee.ManagerEmail,
					Type:  "manager",
				},
			}
		}

		if user.Employee.EmployeeID != "" {
			gsuiteUser.ExternalIds = []directoryv1.UserExternalId{
				{
					Value: user.Employee.EmployeeID,
					Type:  "organization",
				},
			}
		}
	}

	if !user.Location.Empty() {
		gsuiteUser.Locations = []directoryv1.UserLocation{
			{
				Area:         "desk",
				BuildingId:   user.Location.Building,
				FloorName:    user.Location.Floor,
				FloorSection: user.Location.FloorSection,
				Type:         "desk",
			},
		}
	}

	if enableInsecurePasswords && user.Password != "" {
		customData := CustomSchema{
			PasswordHash: HashPassword(user.Password),
		}

		encoded, _ := json.Marshal(customData)
		gsuiteUser.CustomSchemas[SchemaName] = encoded
		gsuiteUser.Password = user.Password
	}

	return gsuiteUser
}

// apiUser represents those fields that are not explicitly spec'ed
// out in the GSuite API, but whose we still have to access.
// Re-marshaling into this struct is easier than tons of type assertions
// througout the codebase.
type apiUser struct {
	Emails []struct {
		Address string `json:"address"`
		Primary bool   `json:"primary"`
	} `json:"emails"`

	Phones []struct {
		Value string `json:"value"`
	} `json:"phones"`

	ExternalIds []struct {
		Value string `json:"value"`
		Type  string `json:"type"`
	} `json:"externalIds"`

	Organizations []struct {
		Department  string `json:"department"`
		Title       string `json:"title"`
		Description string `json:"description"`
		CostCenter  string `json:"costCenter"`
	} `json:"organizations"`

	Relations []struct {
		Value string `json:"value"`
		Type  string `json:"type"`
	} `json:"relations"`

	Locations []struct {
		BuildingId   string `json:"buildingId"`
		FloorName    string `json:"floorName"`
		FloorSection string `json:"floorSection"`
	} `json:"locations"`

	Addresses []struct {
		Formatted string `json:"formatted"`
		Type      string `json:"type"`
	} `json:"addresses"`
}

func ToConfigUser(gsuiteUser *directoryv1.User, userLicenses []License) (User, error) {
	apiUser := apiUser{}
	if err := util.ConvertToStruct(gsuiteUser, &apiUser); err != nil {
		return User{}, fmt.Errorf("failed to decode user: %v", err)
	}

	primaryEmail := ""
	for _, email := range apiUser.Emails {
		if email.Primary {
			primaryEmail = email.Address
			break
		}
	}

	user := User{
		FirstName:     gsuiteUser.Name.GivenName,
		LastName:      gsuiteUser.Name.FamilyName,
		PrimaryEmail:  primaryEmail,
		OrgUnitPath:   gsuiteUser.OrgUnitPath,
		RecoveryPhone: gsuiteUser.RecoveryPhone,
		RecoveryEmail: gsuiteUser.RecoveryEmail,
		Aliases:       gsuiteUser.Aliases,
	}

	for _, phone := range apiUser.Phones {
		user.Phones = append(user.Phones, phone.Value)
	}

	for _, extId := range apiUser.ExternalIds {
		if extId.Type == "organization" {
			user.Employee.EmployeeID = extId.Value
		}
	}

	for _, org := range apiUser.Organizations {
		title := org.Title
		if title == "" {
			title = org.Department
		}

		user.Employee.JobTitle = title
		user.Employee.Type = org.Description
		user.Employee.CostCenter = org.CostCenter
	}

	for _, relation := range apiUser.Relations {
		if relation.Type == "manager" {
			user.Employee.ManagerEmail = relation.Value
		}
	}

	for _, location := range apiUser.Locations {
		user.Location.Building = location.BuildingId
		user.Location.Floor = location.FloorName
		user.Location.FloorSection = location.FloorSection
	}

	for _, address := range apiUser.Addresses {
		if address.Type == "home" {
			user.Address = address.Formatted
		}
	}

	if len(userLicenses) > 0 {
		for _, userLicense := range userLicenses {
			user.Licenses = append(user.Licenses, userLicense.Name)
		}
	}

	user.Sort()

	return user, nil
}

func ToGSuiteGroup(group *Group) (*directoryv1.Group, *groupssettingsv1.Groups) {
	gsuiteGroup := &directoryv1.Group{
		Name:        group.Name,
		Email:       group.Email,
		Description: group.Description,
	}

	groupSettings := &groupssettingsv1.Groups{
		WhoCanContactOwner:   group.WhoCanContactOwner,
		WhoCanViewMembership: group.WhoCanViewMembership,
		WhoCanApproveMembers: group.WhoCanApproveMembers,
		WhoCanPostMessage:    group.WhoCanPostMessage,
		WhoCanJoin:           group.WhoCanJoin,
		IsArchived:           strconv.FormatBool(group.IsArchived),
		AllowExternalMembers: strconv.FormatBool(group.AllowExternalMembers),
	}

	return gsuiteGroup, groupSettings
}

func ToConfigGroup(gsuiteGroup *directoryv1.Group, settings *groupssettingsv1.Groups, members []*directoryv1.Member) (Group, error) {
	allowExternalMembers, err := strconv.ParseBool(settings.AllowExternalMembers)
	if err != nil {
		return Group{}, fmt.Errorf("invalid 'AllowExternalMembers' value: %v", err)
	}

	isArchived, err := strconv.ParseBool(settings.IsArchived)
	if err != nil {
		return Group{}, fmt.Errorf("invalid 'IsArchived' value: %v", err)
	}

	group := Group{
		Name:                 gsuiteGroup.Name,
		Email:                gsuiteGroup.Email,
		Description:          gsuiteGroup.Description,
		WhoCanContactOwner:   settings.WhoCanContactOwner,
		WhoCanViewMembership: settings.WhoCanViewMembership,
		WhoCanApproveMembers: settings.WhoCanApproveMembers,
		WhoCanPostMessage:    settings.WhoCanPostMessage,
		WhoCanJoin:           settings.WhoCanJoin,
		AllowExternalMembers: allowExternalMembers,
		IsArchived:           isArchived,
		Members:              []Member{},
	}

	for _, m := range members {
		group.Members = append(group.Members, ToConfigGroupMember(m))
	}

	group.Sort()

	return group, nil
}

func ToGSuiteGroupMember(member *Member, gsuiteMember *directoryv1.Member) *directoryv1.Member {
	result := &directoryv1.Member{
		Email: member.Email,
		Role:  member.Role,
	}

	if gsuiteMember != nil {
		result.Id = gsuiteMember.Id
		result.Etag = gsuiteMember.Etag
	}

	return result
}

func ToConfigGroupMember(gsuiteMember *directoryv1.Member) Member {
	return Member{
		Email: gsuiteMember.Email,
		Role:  gsuiteMember.Role,
	}
}

func ToGSuiteOrgUnit(orgUnit *OrgUnit) *directoryv1.OrgUnit {
	return &directoryv1.OrgUnit{
		Name:              orgUnit.Name,
		Description:       orgUnit.Description,
		ParentOrgUnitPath: orgUnit.ParentOrgUnitPath,
		BlockInheritance:  orgUnit.BlockInheritance,
	}
}

func ToConfigOrgUnit(orgUnit *directoryv1.OrgUnit) OrgUnit {
	return OrgUnit{
		Name:              orgUnit.Name,
		Description:       orgUnit.Description,
		ParentOrgUnitPath: orgUnit.ParentOrgUnitPath,
		BlockInheritance:  orgUnit.BlockInheritance,
	}
}
