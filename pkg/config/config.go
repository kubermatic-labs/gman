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
	AllowExternalMembers string         `yaml:"allow_external_members,omitempty"`
	IsArchived           string         `yaml:"is_archived,omitempty"`
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
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	re164 := regexp.MustCompile("^\\+[1-9]\\d{1,14}$")

	// validate organization
	if c.Organization == "" {
		//return errors.New("no organization configured")
		allTheErrors = append(allTheErrors, errors.New("no organization configured"))
	}

	//validate users
	userEmails := []string{}
	for _, user := range c.Users {
		if util.StringSliceContains(userEmails, user.PrimaryEmail) {
			//return fmt.Errorf("duplicate user defined (user: %s)", user.PrimaryEmail)
			allTheErrors = append(allTheErrors, fmt.Errorf("duplicate user defined (user: %s)", user.PrimaryEmail))
		}

		if user.PrimaryEmail == user.SecondaryEmail {
			//return fmt.Errorf("user has defined the same primary and secondary email (user: %s)", user.PrimaryEmail)
			allTheErrors = append(allTheErrors, fmt.Errorf("user has defined the same primary and secondary email (user: %s)", user.PrimaryEmail))
		}

		if re.MatchString(user.PrimaryEmail) == false {
			//return fmt.Errorf("invalid format of primary email (user: %s)", user.PrimaryEmail)
			allTheErrors = append(allTheErrors, fmt.Errorf("primary email is not a valid email-address (user: %s)", user.PrimaryEmail))
		}

		if user.SecondaryEmail != "" {
			if re.MatchString(user.SecondaryEmail) == false {
				//return fmt.Errorf("invalid format of secondary email (user: %s)", user.PrimaryEmail)
				allTheErrors = append(allTheErrors, fmt.Errorf("secondary email is not a valid email-address (user: %s)", user.PrimaryEmail))
			}
		}

		if user.RecoveryEmail != "" {
			if re.MatchString(user.RecoveryEmail) == false {
				//return fmt.Errorf("invalid format of recovery email (user: %s)", user.PrimaryEmail)
				allTheErrors = append(allTheErrors, fmt.Errorf("recovery email is not a valid email-address (user: %s)", user.PrimaryEmail))
			}
		}

		if len(user.Aliases) > 0 {
			for _, alias := range user.Aliases {
				if re.MatchString(alias) == false {
					//return fmt.Errorf("invalid format of alias email (user: %s)", user.PrimaryEmail)
					allTheErrors = append(allTheErrors, fmt.Errorf("alias email is not a valid email-address (user: %s)", user.PrimaryEmail))
				}
			}
		}

		if user.Employee.ManagerEmail != "" {
			if re.MatchString(user.Employee.ManagerEmail) == false {
				//return fmt.Errorf("invalid format of the manager's email (user: %s)", user.PrimaryEmail)
				allTheErrors = append(allTheErrors, fmt.Errorf("manager's email is not a valid email-address (user: %s)", user.PrimaryEmail))
			}
		}

		if user.RecoveryPhone != "" {
			if re164.MatchString(user.RecoveryPhone) == false {
				//return fmt.Errorf("invalid format of recovery phone (user: %s). The phone number must be in the E.164 format, starting with the plus sign (+). Example: +16506661212.", user.PrimaryEmail)
				allTheErrors = append(allTheErrors, fmt.Errorf("invalid format of recovery phone (user: %s). The phone number must be in the E.164 format, starting with the plus sign (+). Example: +16506661212.", user.PrimaryEmail))
			}
		}

		userEmails = append(userEmails, user.PrimaryEmail)
	}

	// validate groups
	groupEmails := []string{}
	for _, group := range c.Groups {
		if util.StringSliceContains(groupEmails, group.Email) {
			//return fmt.Errorf("duplicate group email defined (%s)", group.Email)
			allTheErrors = append(allTheErrors, fmt.Errorf("duplicate group email defined (%s)", group.Email))
		}

		if re.MatchString(group.Email) == false {
			//return fmt.Errorf("invalid group email (%s)", group.Email)
			allTheErrors = append(allTheErrors, fmt.Errorf("group email is not a valid email-address (%s)", group.Email))
		}

		if !(strings.EqualFold(group.WhoCanContactOwner, "ALL_IN_DOMAIN_CAN_CONTACT") || strings.EqualFold(group.WhoCanContactOwner, "ALL_MANAGERS_CAN_CONTACT") || strings.EqualFold(group.WhoCanContactOwner, "ALL_MEMBERS_CAN_CONTACT") || strings.EqualFold(group.WhoCanContactOwner, "ANYONE_CAN_CONTACT")) {
			//return fmt.Errorf("wrong value specified for 'who_can_contact_owner' field (group: %s)", group.Name)
			allTheErrors = append(allTheErrors, fmt.Errorf("wrong value specified for 'who_can_contact_owner' field (group: %s)", group.Name))
		}

		if !(strings.EqualFold(group.WhoCanViewMembership, "ALL_IN_DOMAIN_CAN_VIEW") || strings.EqualFold(group.WhoCanViewMembership, "ALL_MEMBERS_CAN_VIEW") || strings.EqualFold(group.WhoCanViewMembership, "ALL_MANAGERS_CAN_VIEW")) {
			//return fmt.Errorf("wrong value specified for 'who_can_view_members' field (group: %s)", group.Name)
			allTheErrors = append(allTheErrors, fmt.Errorf("wrong value specified for 'who_can_view_members' field (group: %s)", group.Name))
		}

		if !(strings.EqualFold(group.WhoCanApproveMembers, "ALL_MEMBERS_CAN_APPROVE") || strings.EqualFold(group.WhoCanApproveMembers, "ALL_MANAGERS_CAN_APPROVE") || strings.EqualFold(group.WhoCanApproveMembers, "ALL_OWNERS_CAN_APPROVE") || strings.EqualFold(group.WhoCanApproveMembers, "NONE_CAN_APPROVE")) {
			//return fmt.Errorf("wrong value specified for 'who_can_approve_members' field (group: %s)", group.Name)
			allTheErrors = append(allTheErrors, fmt.Errorf("wrong value specified for 'who_can_approve_members' field (group: %s)", group.Name))
		}

		if !(strings.EqualFold(group.WhoCanPostMessage, "NONE_CAN_POST") || strings.EqualFold(group.WhoCanPostMessage, "ALL_MANAGERS_CAN_POST") || strings.EqualFold(group.WhoCanPostMessage, "ALL_MEMBERS_CAN_POST") || strings.EqualFold(group.WhoCanPostMessage, "ALL_OWNERS_CAN_POST") || strings.EqualFold(group.WhoCanPostMessage, "ALL_IN_DOMAIN_CAN_POST") || strings.EqualFold(group.WhoCanPostMessage, "ANYONE_CAN_POST")) {
			//return fmt.Errorf("wrong value specified for 'who_can_post' field (group: %s)", group.Name)
			allTheErrors = append(allTheErrors, fmt.Errorf("wrong value specified for 'who_can_post' field (group: %s)", group.Name))
		}

		if !(strings.EqualFold(group.WhoCanJoin, "CAN_REQUEST_TO_JOIN") || strings.EqualFold(group.WhoCanJoin, "INVITED_CAN_JOIN") || strings.EqualFold(group.WhoCanJoin, "ALL_IN_DOMAIN_CAN_JOIN") || strings.EqualFold(group.WhoCanJoin, "ANYONE_CAN_JOIN")) {
			//return fmt.Errorf("wrong value specified for 'who_can_contact_owner' field (group: %s)", group.Name)
			allTheErrors = append(allTheErrors, fmt.Errorf("wrong value specified for 'who_can_contact_owner' field (group: %s)", group.Name))
		}

		if !(strings.EqualFold(group.AllowExternalMembers, "true") || strings.EqualFold(group.AllowExternalMembers, "false")) {
			//return fmt.Errorf("wrong value specified for 'allow_external_members' field (group: %s)", group.Name)
			allTheErrors = append(allTheErrors, fmt.Errorf("wrong value specified for 'allow_external_members' field (group: %s)", group.Name))
		}

		memberEmails := []string{}
		for _, member := range group.Members {
			if util.StringSliceContains(memberEmails, member.Email) {
				//return fmt.Errorf("duplicate member defined in a group (group: %s, member: %s)", group.Name, member.Email)
				allTheErrors = append(allTheErrors, fmt.Errorf("duplicate member defined in a group (group: %s, member: %s)", group.Name, member.Email))
			}

			if !(strings.EqualFold(member.Role, "OWNER") || strings.EqualFold(member.Role, "MANAGER") || strings.EqualFold(member.Role, "MEMBER")) {
				//return fmt.Errorf("wrong member role specified (group: %s, member: %s)", group.Name, member.Email)
				allTheErrors = append(allTheErrors, fmt.Errorf("wrong member role specified (group: %s, member: %s). Permitted values are OWNER, MEMBER or MANAGER.", group.Name, member.Email))
			}
		}
	}

	// validate org_units
	ouNames := []string{}
	for _, ou := range c.OrgUnits {
		if util.StringSliceContains(ouNames, ou.Name) {
			//return fmt.Errorf("duplicate org unit defined (%s)", ou.Name)
			allTheErrors = append(allTheErrors, fmt.Errorf("duplicate org unit defined (%s)", ou.Name))
		}

		if ou.ParentOrgUnitPath[0] != '/' {
			//return fmt.Errorf("wrong ParentOrgUnitPath specified for org unit (%s)", ou.Name)
			allTheErrors = append(allTheErrors, fmt.Errorf("wrong ParentOrgUnitPath specified for org unit (%s)", ou.Name))

		}
		if ou.OrgUnitPath[0] != '/' {
			//return fmt.Errorf("wrong OrgUnitPath specified for org unit (%s)", ou.Name)
			allTheErrors = append(allTheErrors, fmt.Errorf("wrong OrgUnitPath specified for org unit (%s)", ou.Name))
		}
	}

	if allTheErrors != nil {
		return allTheErrors
	} else {
		return nil
	}
}
