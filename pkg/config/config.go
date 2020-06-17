package config

import (
	"errors"
	"log"
	"os"
	"regexp"

	"github.com/kubermatic-labs/gman/pkg/util"
	"gopkg.in/yaml.v3"
)

// TODO: everything ;_;

type Config struct {
	Organization string       `yaml:"organization"`
	Users        []UserConfig `yaml:"users,omitempty"`
}

type UserConfig struct {
	FirstName      string `yaml:"given_name"`
	LastName       string `yaml:"family_name"`
	PrimaryEmail   string `yaml:"primary_email"`
	SecondaryEmail string `yaml:"secondary_email,omitempty"`
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

	//validate users -> check if duplicated users
	userEmails := []string{}
	for _, user := range c.Users {
		if util.StringSliceContains(userEmails, user.PrimaryEmail) {
			log.Fatal("Validation failes: duplicate user defined (user: " + user.PrimaryEmail + ")")
		}

		if user.PrimaryEmail == user.SecondaryEmail {
			log.Fatal("Validation failed: user has defined the same primary and secondary email (user: " + user.PrimaryEmail + ")")
		}

		re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

		if re.MatchString(user.PrimaryEmail) == false {
			log.Fatal("Validation failed: invalid primary email (user: " + user.PrimaryEmail + ")")
		}

		if user.SecondaryEmail != "" {
			if re.MatchString(user.SecondaryEmail) == false {
				log.Fatal("Validation failed: invalid secondary email " + user.SecondaryEmail + " (user: " + user.PrimaryEmail + ")")
			}
		}
		userEmails = append(userEmails, user.PrimaryEmail)
	}

	return nil
}
