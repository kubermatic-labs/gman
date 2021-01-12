package config

import (
	"os"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/sets"
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
}

type User struct {
	FirstName      string   `yaml:"givenName"`
	LastName       string   `yaml:"familyName"`
	PrimaryEmail   string   `yaml:"primaryEmail"`
	SecondaryEmail string   `yaml:"secondaryEmail,omitempty"`
	Aliases        []string `yaml:"aliases,omitempty"`
	Phones         []string `yaml:"phones,omitempty"`
	RecoveryPhone  string   `yaml:"recoveryPhone,omitempty"`
	RecoveryEmail  string   `yaml:"recoveryEmail,omitempty"`
	OrgUnitPath    string   `yaml:"orgUnitPath,omitempty"`
	Licenses       []string `yaml:"licenses,omitempty"`
	Employee       Employee `yaml:"employeeInfo,omitempty"`
	Location       Location `yaml:"location,omitempty"`
	Address        string   `yaml:"addresses,omitempty"`
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

type Member struct {
	Email string `yaml:"email"`
	Role  string `yaml:"role,omitempty"`
}

type OrgUnit struct {
	Name              string `yaml:"name"`
	Description       string `yaml:"description,omitempty"`
	ParentOrgUnitPath string `yaml:"parentOrgUnitPath,omitempty"`
	BlockInheritance  bool   `yaml:"blockInheritance,omitempty"`
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

	if err := encoder.Encode(config); err != nil {
		return err
	}

	return nil
}
