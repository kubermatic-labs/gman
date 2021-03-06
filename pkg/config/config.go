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
	"crypto/sha256"
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	SchemaName              = "gman"
	PasswordHashCustomField = "passwordHash"
)

const (
	// WhoCanContactOwner
	GroupOptionAllManagersCanContact     = "ALL_MANAGERS_CAN_CONTACT"
	GroupOptionAllMembersCanContact      = "ALL_MEMBERS_CAN_CONTACT"
	GroupOptionAllInDomainCanContact     = "ALL_IN_DOMAIN_CAN_CONTACT"
	GroupOptionAnyoneCanContact          = "ANYONE_CAN_CONTACT"
	GroupOptionWhoCanContactOwnerDefault = GroupOptionAllInDomainCanContact

	// WhoCanViewMembership
	GroupOptionAllManagersCanViewMembership = "ALL_MANAGERS_CAN_VIEW"
	GroupOptionAllMembersCanViewMembership  = "ALL_MEMBERS_CAN_VIEW"
	GroupOptionAllInDomainCanViewMembership = "ALL_IN_DOMAIN_CAN_VIEW"
	GroupOptionWhoCanViewMembershipDefault  = GroupOptionAllMembersCanViewMembership

	// WhoCanApproveMembers
	GroupOptionAllManagersCanApproveMembers = "ALL_MANAGERS_CAN_APPROVE"
	GroupOptionAllOwnersCanApproveMembers   = "ALL_OWNERS_CAN_APPROVE"
	GroupOptionAllMembersCanApproveMembers  = "ALL_MEMBERS_CAN_APPROVE"
	GroupOptionNoneCanApproveMembers        = "NONE_CAN_APPROVE"
	GroupOptionWhoCanApproveMembersDefault  = GroupOptionAllManagersCanApproveMembers

	// WhoCanPostMessage
	GroupOptionNoneCanPostMessage        = "NONE_CAN_POST"
	GroupOptionAllOwnersCanPostMessage   = "ALL_OWNERS_CAN_POST"
	GroupOptionAllManagersCanPostMessage = "ALL_MANAGERS_CAN_POST"
	GroupOptionAllMembersCanPostMessage  = "ALL_MEMBERS_CAN_POST"
	GroupOptionAllInDomainCanPostMessage = "ALL_IN_DOMAIN_CAN_POST"
	GroupOptionAnyoneCanPostMessage      = "ANYONE_CAN_POST"
	GroupOptionWhoCanPostMessageDefault  = GroupOptionAllMembersCanPostMessage

	// WhoCanJoin
	GroupOptionInvitedCanJoin     = "INVITED_CAN_JOIN"
	GroupOptionCanRequestToJoin   = "CAN_REQUEST_TO_JOIN"
	GroupOptionAllInDomainCanJoin = "ALL_IN_DOMAIN_CAN_JOIN"
	GroupOptionAnyoneCanJoin      = "ANYONE_CAN_JOIN"
	GroupOptionWhoCanJoinDefault  = GroupOptionInvitedCanJoin

	// membership roles
	MemberRoleOwner   = "OWNER"
	MemberRoleManager = "MANAGER"
	MemberRoleMember  = "MEMBER"
)

var (
	allWhoCanContactOwnerOptions = sets.NewString(
		GroupOptionAllManagersCanContact,
		GroupOptionAllMembersCanContact,
		GroupOptionAllInDomainCanContact,
		GroupOptionAnyoneCanContact,
	)

	allWhoCanViewMembershipOptions = sets.NewString(
		GroupOptionAllManagersCanViewMembership,
		GroupOptionAllMembersCanViewMembership,
		GroupOptionAllInDomainCanViewMembership,
	)

	allWhoCanApproveMembersOptions = sets.NewString(
		GroupOptionAllManagersCanApproveMembers,
		GroupOptionAllOwnersCanApproveMembers,
		GroupOptionAllMembersCanApproveMembers,
		GroupOptionNoneCanApproveMembers,
	)

	allWhoCanPostMessageOptions = sets.NewString(
		GroupOptionNoneCanPostMessage,
		GroupOptionAllOwnersCanPostMessage,
		GroupOptionAllManagersCanPostMessage,
		GroupOptionAllMembersCanPostMessage,
		GroupOptionAllInDomainCanPostMessage,
		GroupOptionAnyoneCanPostMessage,
	)

	allWhoCanJoinOptions = sets.NewString(
		GroupOptionInvitedCanJoin,
		GroupOptionCanRequestToJoin,
		GroupOptionAllInDomainCanJoin,
		GroupOptionAnyoneCanJoin,
	)

	allMemberRoles = sets.NewString(
		MemberRoleOwner,
		MemberRoleManager,
		MemberRoleMember,
	)
)

type Config struct {
	Organization string    `yaml:"organization"`
	OrgUnits     []OrgUnit `yaml:"orgUnits,omitempty"`
	Users        []User    `yaml:"users,omitempty"`
	Groups       []Group   `yaml:"groups,omitempty"`
	Licenses     []License `yaml:"licenses,omitempty"`
}

type OrgUnit struct {
	Name              string `yaml:"name"`
	Description       string `yaml:"description,omitempty"`
	ParentOrgUnitPath string `yaml:"parentOrgUnitPath,omitempty"`
	BlockInheritance  bool   `yaml:"blockInheritance,omitempty"`
}

type User struct {
	FirstName     string   `yaml:"givenName"`
	LastName      string   `yaml:"familyName"`
	PrimaryEmail  string   `yaml:"primaryEmail"`
	Aliases       []string `yaml:"aliases,omitempty"`
	Phones        []string `yaml:"phones,omitempty"`
	RecoveryPhone string   `yaml:"recoveryPhone,omitempty"`
	RecoveryEmail string   `yaml:"recoveryEmail,omitempty"`
	OrgUnitPath   string   `yaml:"orgUnitPath,omitempty"`
	Licenses      []string `yaml:"licenses,omitempty"`
	Employee      Employee `yaml:"employeeInfo,omitempty"`
	Location      Location `yaml:"location,omitempty"`
	Address       string   `yaml:"address,omitempty"`
	Password      string   `yaml:"password,omitempty"`
}

func (u *User) Sort() {
	sort.Strings(u.Aliases)
	sort.Strings(u.Phones)
	sort.Strings(u.Licenses)
}

type Location struct {
	Building     string `yaml:"building,omitempty"`
	Floor        string `yaml:"floor,omitempty"`
	FloorSection string `yaml:"floorSection,omitempty"`
}

func (l *Location) Empty() bool {
	return l.Building == "" && l.Floor == "" && l.FloorSection == ""
}

type Employee struct {
	EmployeeID   string `yaml:"id,omitempty"`
	Department   string `yaml:"department,omitempty"`
	JobTitle     string `yaml:"jobTitle,omitempty"`
	Type         string `yaml:"type,omitempty"`
	CostCenter   string `yaml:"costCenter,omitempty"`
	ManagerEmail string `yaml:"managerEmail,omitempty"`
}

func (e *Employee) Empty() bool {
	return e.EmployeeID == "" && e.Department == "" && e.JobTitle == "" && e.Type == "" && e.CostCenter == "" && e.ManagerEmail == ""
}

type Group struct {
	Name                 string   `yaml:"name"`
	Email                string   `yaml:"email"`
	Description          string   `yaml:"description,omitempty"`
	WhoCanContactOwner   string   `yaml:"whoCanContactOwner,omitempty"`
	WhoCanViewMembership string   `yaml:"whoCanViewMembers,omitempty"`
	WhoCanApproveMembers string   `yaml:"whoCanApproveMembers,omitempty"`
	WhoCanPostMessage    string   `yaml:"whoCanPostMessage,omitempty"`
	WhoCanJoin           string   `yaml:"whoCanJoin,omitempty"`
	AllowExternalMembers bool     `yaml:"allowExternalMembers,omitempty"`
	IsArchived           bool     `yaml:"isArchived,omitempty"`
	Members              []Member `yaml:"members,omitempty"`
}

func (g *Group) Sort() {
	sort.SliceStable(g.Members, func(i, j int) bool {
		return g.Members[i].Email < g.Members[j].Email
	})
}

type Member struct {
	Email string `yaml:"email"`
	Role  string `yaml:"role,omitempty"`
}

func LoadFromFile(filename string) (*Config, error) {
	config := &Config{}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(config); err != nil {
		return nil, err
	}

	// apply default values
	config.DefaultOrgUnits()
	config.DefaultUsers()
	config.DefaultGroups()
	config.Sort()

	return config, nil
}

func SaveToFile(config *Config, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)

	// remove default values so we create a minimal config file
	config.UndefaultOrgUnits()
	config.UndefaultUsers()
	config.UndefaultGroups()
	config.Sort()

	if err := encoder.Encode(config); err != nil {
		return err
	}

	return nil
}

// HashPassword returns an shortened hash for the given password;
// the hash is not meant to validate password inputs, but only to
// compare if the passwords have changed
func HashPassword(password string) string {
	checksum := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", checksum[:16])
}
