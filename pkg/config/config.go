package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/kubermatic-labs/gman/pkg/util"
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
	AllowExternalMembers bool     `yaml:"allowExternalMembers"`
	IsArchived           bool     `yaml:"isArchived"`
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
	config := &Config{} // create config structure

	f, err := os.Open(filename) // open file config
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(config); err != nil {
		return nil, err
	}

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

	if err := encoder.Encode(config); err != nil {
		return err
	}

	return nil
}

// validateEmailFormat is a helper function that checks for existance of '@' and the length of the address
func validateEmailFormat(email string) bool {
	return (len(email) < 129 && strings.Contains(email, "@"))
}

func (c *Config) ValidateUsers() []error {
	var allTheErrors []error
	re164 := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

	// validate organization
	if c.Organization == "" {
		allTheErrors = append(allTheErrors, errors.New("no organization configured"))
	}

	//validate users
	userEmails := []string{}
	for _, user := range c.Users {
		if util.StringSliceContains(userEmails, user.PrimaryEmail) {
			allTheErrors = append(allTheErrors, fmt.Errorf("duplicate user defined (user: %s)", user.PrimaryEmail))
		}

		if user.PrimaryEmail == "" {
			allTheErrors = append(allTheErrors, fmt.Errorf("primary email is required (user: %s)", user.LastName))
		} else {
			if user.PrimaryEmail == user.SecondaryEmail {
				allTheErrors = append(allTheErrors, fmt.Errorf("user has defined the same primary and secondary email (user: %s)", user.PrimaryEmail))
			}
			if !validateEmailFormat(user.PrimaryEmail) {
				allTheErrors = append(allTheErrors, fmt.Errorf("primary email is not a valid email-address (user: %s)", user.PrimaryEmail))
			}
		}

		if user.FirstName == "" || user.LastName == "" {
			allTheErrors = append(allTheErrors, fmt.Errorf("given and family names are required (user: %s)", user.PrimaryEmail))
		}

		if user.SecondaryEmail != "" {
			if !validateEmailFormat(user.SecondaryEmail) {
				allTheErrors = append(allTheErrors, fmt.Errorf("secondary email is not a valid email-address (user: %s)", user.PrimaryEmail))
			}
		}

		if user.RecoveryEmail != "" {
			if !validateEmailFormat(user.RecoveryEmail) {
				allTheErrors = append(allTheErrors, fmt.Errorf("recovery email is not a valid email-address (user: %s)", user.PrimaryEmail))
			}
		}

		if len(user.Aliases) > 0 {
			for _, alias := range user.Aliases {
				if !validateEmailFormat(alias) {
					allTheErrors = append(allTheErrors, fmt.Errorf("alias email is not a valid email-address (user: %s)", user.PrimaryEmail))
				}
			}
		}

		if user.Employee.ManagerEmail != "" {
			if !validateEmailFormat(user.Employee.ManagerEmail) {
				allTheErrors = append(allTheErrors, fmt.Errorf("manager's email is not a valid email-address (user: %s)", user.PrimaryEmail))
			}
		}

		if user.RecoveryPhone != "" {
			if !re164.MatchString(user.RecoveryPhone) {
				allTheErrors = append(allTheErrors, fmt.Errorf("invalid format of recovery phone (user: %s). The phone number must be in the E.164 format, starting with the plus sign (+). Example: +16506661212.", user.PrimaryEmail))
			}
		}

		if len(user.Licenses) > 0 {
			for _, license := range user.Licenses {
				found := false
				for _, permLicense := range AllLicenses {
					if license == permLicense.Name {
						found = true
					}
				}
				if !found {
					allTheErrors = append(allTheErrors, fmt.Errorf("wrong value specified for the user license (user: %s, license: %s)", user.PrimaryEmail, license))
				}
			}
		}

		userEmails = append(userEmails, user.PrimaryEmail)
	}

	if allTheErrors != nil {
		return allTheErrors
	}

	return nil
}

func (c *Config) ValidateGroups() []error {
	var allTheErrors []error

	// validate organization
	if c.Organization == "" {
		allTheErrors = append(allTheErrors, errors.New("no organization configured"))
	}

	// validate groups
	groupEmails := []string{}
	for _, group := range c.Groups {
		if util.StringSliceContains(groupEmails, group.Email) {
			allTheErrors = append(allTheErrors, fmt.Errorf("duplicate group email defined (%s)", group.Email))
		}

		if !validateEmailFormat(group.Email) {
			allTheErrors = append(allTheErrors, fmt.Errorf("group email is not a valid email-address (%s)", group.Email))
		}

		if group.WhoCanContactOwner != "" {
			if !(strings.Compare(group.WhoCanContactOwner, "ALL_IN_DOMAIN_CAN_CONTACT") == 0 || strings.Compare(group.WhoCanContactOwner, "ALL_MANAGERS_CAN_CONTACT") == 0 || strings.Compare(group.WhoCanContactOwner, "ALL_MEMBERS_CAN_CONTACT") == 0 || strings.Compare(group.WhoCanContactOwner, "ANYONE_CAN_CONTACT") == 0) {
				allTheErrors = append(allTheErrors, fmt.Errorf("wrong value specified for 'who_can_contact_owner' field (group: %s). For the list of possible values, please refer to example config. Fields are case sensitive.", group.Name))
			}
		}

		if group.WhoCanViewMembership != "" {
			if !(strings.Compare(group.WhoCanViewMembership, "ALL_IN_DOMAIN_CAN_VIEW") == 0 || strings.Compare(group.WhoCanViewMembership, "ALL_MEMBERS_CAN_VIEW") == 0 || strings.Compare(group.WhoCanViewMembership, "ALL_MANAGERS_CAN_VIEW") == 0) {
				allTheErrors = append(allTheErrors, fmt.Errorf("wrong value specified for 'who_can_view_members' field (group: %s)", group.Name))
			}
		}
		if group.WhoCanApproveMembers != "" {
			if !(strings.Compare(group.WhoCanApproveMembers, "ALL_MEMBERS_CAN_APPROVE") == 0 || strings.Compare(group.WhoCanApproveMembers, "ALL_MANAGERS_CAN_APPROVE") == 0 || strings.Compare(group.WhoCanApproveMembers, "ALL_OWNERS_CAN_APPROVE") == 0 || strings.Compare(group.WhoCanApproveMembers, "NONE_CAN_APPROVE") == 0) {
				allTheErrors = append(allTheErrors, fmt.Errorf("wrong value specified for 'who_can_approve_members' field (group: %s)", group.Name))
			}
		}

		if group.WhoCanPostMessage != "" {
			if !(strings.Compare(group.WhoCanPostMessage, "NONE_CAN_POST") == 0 || strings.Compare(group.WhoCanPostMessage, "ALL_MANAGERS_CAN_POST") == 0 || strings.Compare(group.WhoCanPostMessage, "ALL_MEMBERS_CAN_POST") == 0 || strings.Compare(group.WhoCanPostMessage, "ALL_OWNERS_CAN_POST") == 0 || strings.Compare(group.WhoCanPostMessage, "ALL_IN_DOMAIN_CAN_POST") == 0 || strings.Compare(group.WhoCanPostMessage, "ANYONE_CAN_POST") == 0) {
				allTheErrors = append(allTheErrors, fmt.Errorf("wrong value specified for 'who_can_post' field (group: %s)", group.Name))
			}
		}
		if group.WhoCanJoin != "" {
			if !(strings.Compare(group.WhoCanJoin, "CAN_REQUEST_TO_JOIN") == 0 || strings.Compare(group.WhoCanJoin, "INVITED_CAN_JOIN") == 0 || strings.Compare(group.WhoCanJoin, "ALL_IN_DOMAIN_CAN_JOIN") == 0 || strings.Compare(group.WhoCanJoin, "ANYONE_CAN_JOIN") == 0) {
				allTheErrors = append(allTheErrors, fmt.Errorf("wrong value specified for 'who_can_contact_owner' field (group: %s)", group.Name))
			}
		}

		memberEmails := []string{}
		for _, member := range group.Members {
			if util.StringSliceContains(memberEmails, member.Email) {
				allTheErrors = append(allTheErrors, fmt.Errorf("duplicate member defined in a group (group: %s, member: %s)", group.Name, member.Email))
			}

			if !(strings.Compare(member.Role, "OWNER") == 0 || strings.Compare(member.Role, "MANAGER") == 0 || strings.Compare(member.Role, "MEMBER") == 0) {
				allTheErrors = append(allTheErrors, fmt.Errorf("wrong member role specified (group: %s, member: %s). Permitted values are OWNER, MEMBER or MANAGER.", group.Name, member.Email))
			}
		}
	}

	if allTheErrors != nil {
		return allTheErrors
	}

	return nil
}

func (c *Config) ValidateOrgUnits() []error {
	var allTheErrors []error

	// validate organization
	if c.Organization == "" {
		allTheErrors = append(allTheErrors, errors.New("no organization configured"))
	}

	// validate org_units
	ouNames := []string{}
	for _, ou := range c.OrgUnits {
		if util.StringSliceContains(ouNames, ou.Name) {
			allTheErrors = append(allTheErrors, fmt.Errorf("duplicate org unit defined (%s)", ou.Name))
		}

		if ou.Name == "" {
			allTheErrors = append(allTheErrors, fmt.Errorf("'Name' is not specified (org unit %s)", ou.Name))
		}

		if ou.ParentOrgUnitPath == "" {
			allTheErrors = append(allTheErrors, fmt.Errorf("'ParentOrgUnitPath' is not specified (org unit %s)", ou.Name))
		} else {
			if ou.ParentOrgUnitPath[0] != '/' {
				allTheErrors = append(allTheErrors, fmt.Errorf("'ParentOrgUnitPath' must start with a slash (org unit %s)", ou.Name))
			}
		}
	}

	if allTheErrors != nil {
		return allTheErrors
	}

	return nil
}
