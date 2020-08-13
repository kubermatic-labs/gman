package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/kubermatic-labs/gman/pkg/util"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Organization string          `yaml:"organization"`
	OrgUnits     []OrgUnitConfig `yaml:"org_units,omitempty"`
	Users        []UserConfig    `yaml:"users,omitempty"`
	Groups       []GroupConfig   `yaml:"groups,omitempty"`
}

type UserConfig struct {
	FirstName      string         `yaml:"given_name"`
	LastName       string         `yaml:"family_name"`
	PrimaryEmail   string         `yaml:"primary_email"`
	SecondaryEmail string         `yaml:"secondary_email,omitempty"`
	Aliases        []string       `yaml:"aliases,omitempty"`
	Phones         []string       `yaml:"phones,omitempty"`
	RecoveryPhone  string         `yaml:"recovery_phone,omitempty"`
	RecoveryEmail  string         `yaml:"recovery_email,omitempty"`
	OrgUnitPath    string         `yaml:"org_unit_path,omitempty"`
	Employee       EmployeeConfig `yaml:"employee_info,omitempty"`
	Location       LocationConfig `yaml:"location,omitempty"`
	Address        string         `yaml:"addresses,omitempty"`
}

type LocationConfig struct {
	Building     string `yaml:"building,omitempty"`
	Floor        string `yaml:"floor,omitempty"`
	FloorSection string `yaml:"floor_section,omitempty"`
}

type EmployeeConfig struct {
	EmployeeID   string `yaml:"employee_ID,omitempty"`
	Department   string `yaml:"department,omitempty"`
	JobTitle     string `yaml:"job_title,omitempty"`
	Type         string `yaml:"type,omitempty"`
	CostCenter   string `yaml:"cost_center,omitempty"`
	ManagerEmail string `yaml:"manager_email,omitempty"`
}

type GroupConfig struct {
	Name                 string         `yaml:"name"`
	Email                string         `yaml:"email"`
	Description          string         `yaml:"description,omitempty"`
	WhoCanContactOwner   string         `yaml:"who_can_contact_owner,omitempty"`
	WhoCanViewMembership string         `yaml:"who_can_view_members,omitempty"`
	WhoCanApproveMembers string         `yaml:"who_can_approve_members,omitempty"`
	WhoCanPostMessage    string         `yaml:"who_can_post,omitempty"`
	WhoCanJoin           string         `yaml:"who_can_join,omitempty"`
	AllowExternalMembers bool           `yaml:"allow_external_members"`
	IsArchived           bool           `yaml:"is_archived"`
	Members              []MemberConfig `yaml:"members,omitempty"`
}

type MemberConfig struct {
	Email string `yaml:"email"`
	Role  string `yaml:"role,omitempty"`
}

type OrgUnitConfig struct {
	Name              string `yaml:"name"`
	Description       string `yaml:"description,omitempty"`
	ParentOrgUnitPath string `yaml:"parent_org_unit_path,omitempty"`
	OrgUnitPath       string `yaml:"org_unit_path,omitempty"`
	BlockInheritance  bool   `yaml:"block_inheritance,omitempty"`
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
	if err := yaml.NewEncoder(f).Encode(config); err != nil {
		return err
	}

	return nil
}

func (c *Config) Validate() []error {
	var allTheErrors []error
	re164 := regexp.MustCompile("^\\+[1-9]\\d{1,14}$")

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
			if validateEmailFormat(user.PrimaryEmail) == false {
				allTheErrors = append(allTheErrors, fmt.Errorf("primary email is not a valid email-address (user: %s)", user.PrimaryEmail))
			}
		}

		if user.FirstName == "" || user.LastName == "" {
			allTheErrors = append(allTheErrors, fmt.Errorf("given and family names are required (user: %s)", user.PrimaryEmail))
		}

		if user.SecondaryEmail != "" {
			if validateEmailFormat(user.SecondaryEmail) == false {
				allTheErrors = append(allTheErrors, fmt.Errorf("secondary email is not a valid email-address (user: %s)", user.PrimaryEmail))
			}
		}

		if user.RecoveryEmail != "" {
			if validateEmailFormat(user.RecoveryEmail) == false {
				allTheErrors = append(allTheErrors, fmt.Errorf("recovery email is not a valid email-address (user: %s)", user.PrimaryEmail))
			}
		}

		if len(user.Aliases) > 0 {
			for _, alias := range user.Aliases {
				if validateEmailFormat(alias) == false {
					allTheErrors = append(allTheErrors, fmt.Errorf("alias email is not a valid email-address (user: %s)", user.PrimaryEmail))
				}
			}
		}

		if user.Employee.ManagerEmail != "" {
			if validateEmailFormat(user.Employee.ManagerEmail) == false {
				allTheErrors = append(allTheErrors, fmt.Errorf("manager's email is not a valid email-address (user: %s)", user.PrimaryEmail))
			}
		}

		if user.RecoveryPhone != "" {
			if re164.MatchString(user.RecoveryPhone) == false {
				allTheErrors = append(allTheErrors, fmt.Errorf("invalid format of recovery phone (user: %s). The phone number must be in the E.164 format, starting with the plus sign (+). Example: +16506661212.", user.PrimaryEmail))
			}
		}

		userEmails = append(userEmails, user.PrimaryEmail)
	}

	// validate groups
	groupEmails := []string{}
	for _, group := range c.Groups {
		if util.StringSliceContains(groupEmails, group.Email) {
			allTheErrors = append(allTheErrors, fmt.Errorf("duplicate group email defined (%s)", group.Email))
		}

		if validateEmailFormat(group.Email) == false {
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

		if ou.OrgUnitPath == "" {
			allTheErrors = append(allTheErrors, fmt.Errorf("'OrgUnitPath' is not specified (org unit %s)", ou.Name))
		} else {
			if ou.OrgUnitPath[0] != '/' {
				allTheErrors = append(allTheErrors, fmt.Errorf("'OrgUnitPath' must start with a slash (org unit %s)", ou.Name))
			}
		}
	}

	if allTheErrors != nil {
		return allTheErrors
	} else {
		return nil
	}
}

// validateEmailFormat is a helper function that checks for existance of '@' and the length of the address
func validateEmailFormat(email string) bool {
	if (strings.Contains(email, "@")) && len(email) < 129 {
		return true
	} else {
		return false
	}
}
