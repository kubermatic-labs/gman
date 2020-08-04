package config

import (
	"errors"
	"log"
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

func (c *Config) Validate() error {
	// validate organization
	if c.Organization == "" {
		return errors.New("no organization configured")
	}

	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	//validate users
	userEmails := []string{}
	for _, user := range c.Users {
		if util.StringSliceContains(userEmails, user.PrimaryEmail) {
			log.Fatalf("Validation failed: duplicate user defined (user: %s)", user.PrimaryEmail)
		}

		if user.PrimaryEmail == user.SecondaryEmail {
			log.Fatalf("Validation failed: user has defined the same primary and secondary email (user: %s)", user.PrimaryEmail)
		}

		if re.MatchString(user.PrimaryEmail) == false {
			log.Fatalf("Validation failed: invalid primary email (user: %s)", user.PrimaryEmail)
		}

		if user.SecondaryEmail != "" {
			if re.MatchString(user.SecondaryEmail) == false {
				log.Fatalf("Validation failed: invalid secondary email (user: %s)", user.PrimaryEmail)
			}
		}
		userEmails = append(userEmails, user.PrimaryEmail)
	}

	// validate groups
	groupEmails := []string{}
	for _, group := range c.Groups {
		if util.StringSliceContains(groupEmails, group.Email) {
			log.Fatalf("Validation failed: duplicate group email defined (%s)", group.Email)
		}

		if re.MatchString(group.Email) == false {
			log.Fatalf("Validation failed: invalid group email (%s)", group.Email)
		}

		memberEmails := []string{}
		for _, member := range group.Members {
			if util.StringSliceContains(memberEmails, member.Email) {
				log.Fatalf("Validation failed: duplicate member defined in a group (group: %s, member: %s)", group.Name, member.Email)
			}

			if !(strings.EqualFold(member.Role, "OWNER") || strings.EqualFold(member.Role, "MANAGER") || strings.EqualFold(member.Role, "MEMBER")) {
				log.Fatalf("Validation failed: wrong member role specified (group: %s, member: %s)", group.Name, member.Email)
			}
		}
	}

	// validate org_units
	ouNames := []string{}
	for _, ou := range c.OrgUnits {
		if util.StringSliceContains(ouNames, ou.Name) {
			log.Fatalf("Validation failed: duplicate org unit defined (%s)", ou.Name)
		}

		if ou.ParentOrgUnitPath[0] != '/' {
			log.Fatalf("Validation failed: wrong ParentOrgUnitPath specified for org unit (%s)", ou.Name)
		}
		if ou.OrgUnitPath[0] != '/' {
			log.Fatalf("Validation failed: wrong OrgUnitPath specified for org unit (%s)", ou.Name)
		}
	}

	return nil
}
