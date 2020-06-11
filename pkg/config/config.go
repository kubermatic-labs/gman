package config

import (
	"errors"
	"os"

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
	Password       string `yaml:"password,omitempty"`
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
	if c.Organization == "" {
		return errors.New("no organization configured")
	}

	//	// TODO OMG EVERITING
	//
	//	userLastNames := []string{}
	//
	//	for _, user := range c.Users {
	//		if util.StringSliceContains(userLastNames, user.Name) {
	//			return fmt.Errorf("duplicate team %q defined", team.Name)
	//		}
	//
	//		teamNames = append(teamNames, team.Name)
	//	}
	//
	//

	return nil
}
