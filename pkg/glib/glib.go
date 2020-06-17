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

// GSuiteUser stores the User object returned from the users api
// https://godoc.org/google.golang.org/api/admin/directory/v1#User
type GSuiteUser struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Role      string `json:"role"`
	ID        int    `json:"id"`
}

// CreateDirectoryService() creates a client for communicating with Google APIs,
// returns an Admin SDK Directory service object authorized with.
func NewDirectoryService(clientSecretFile string, impersonatedUserEmail string) (*admin.Service, error) {
	ctx := context.Background()

	jsonCredentials, err := ioutil.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("ReadFile(clientSecretFile): %v", err)
	}

	config, err := google.JWTConfigFromJSON(jsonCredentials, admin.AdminDirectoryUserScope)
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

// GetListOfUsers returns a list of all current users form the API
func GetListOfUsers(srv admin.Service) ([]*admin.User, error) {
	request, err := srv.Users.List().Customer("my_customer").OrderBy("email").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve users in domain: %v", err)
		return nil, err
	}
	return request.Users, nil
}

// helper function for testing / TODO: change it, esrase it, whatever
func PrintUsers(users []*admin.User) {
	if len(users) == 0 {
		fmt.Print("No users found.\n")
	} else {
		fmt.Print("Current users in Gsuite:\n")
		for _, u := range users {
			pri, sec := GetUserEmails(u)
			fmt.Printf("  %s (%s) (secondary: %s) \n", pri, u.Name.FullName, sec)
		}
	}
}

// GetUserEmails retrieves primary and secondary (type: work) user email addresses
// it is impossible to make it nicer ;_;
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

// CreateNewUse creates a new user in GSuite via their API
func CreateNewUser(srv admin.Service, user *config.UserConfig) error {
	// generate a rand password
	pass, err := password.Generate(20, 5, 5, false, false)
	if err != nil {
		log.Fatalf("Unable to generate password: %v", err)
		return err
	}
	newUser := admin.User{
		Name: &admin.UserName{
			GivenName:  user.FirstName,
			FamilyName: user.LastName,
		},
		PrimaryEmail:  user.PrimaryEmail,
		RecoveryEmail: user.SecondaryEmail,
		Emails: []admin.UserEmail{
			{
				Address: user.SecondaryEmail,
				Type:    "work",
			},
		},
		Password:                  pass,
		ChangePasswordAtNextLogin: true,
	}

	_, err = srv.Users.Insert(&newUser).Do()
	if err != nil {
		log.Fatalf("Unable to create a user: %v", err)
		return err
	}
	log.Printf("Created user: %s \n", user.PrimaryEmail)
	return nil
}

// DeleteUser deletes a user in GSuite via their API
func DeleteUser(srv admin.Service, user *admin.User) error {
	err := srv.Users.Delete(user.PrimaryEmail).Do()
	if err != nil {
		log.Fatalf("Unable to delete a user: %v", err)
		return err
	}
	log.Printf("Deleted user: %s \n", user.PrimaryEmail)
	return nil
}

// UpdateUser makes sure that the user in Gsuite is corresponding to user in config
// in case it is not, it updates the remote user with config
func UpdateUser(srv admin.Service, user *config.UserConfig) error {
	updatedUser := &admin.User{
		Name: &admin.UserName{
			GivenName:  user.FirstName,
			FamilyName: user.LastName,
		},
		PrimaryEmail:  user.PrimaryEmail,
		RecoveryEmail: user.SecondaryEmail,
		Emails: []admin.UserEmail{
			{
				Address: user.SecondaryEmail,
				Type:    "work",
			},
		},
	}

	_, err := srv.Users.Update(user.PrimaryEmail, updatedUser).Do()
	if err != nil {
		log.Fatalf("Unable to update a user: %v", err)
		return err
	}
	log.Printf("Updated user: %s \n", user.PrimaryEmail)
	return nil
}
