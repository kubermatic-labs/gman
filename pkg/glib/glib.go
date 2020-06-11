// Package glib contains methods for interactions with GSuite API
package glib

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

//TODO: change to read this info from file or sth
const (
	userEmail             = "gman-dev-robot@gman-dev-project.iam.gserviceaccount.com"
	impersonatedUserEmail = "marta@loodse.training" // ???
	clientSecretFile      = "gman-dev-project-bf73cc12b7a5.json"
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
func NewDirectoryService() (*admin.Service, error) {
	ctx := context.Background()

	jsonCredentials, err := ioutil.ReadFile(clientSecretFile)
	if err != nil {
		return nil, err
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
	}

	PrintUsers(request.Users)

	return request.Users, nil
}

func PrintUsers(users []*admin.User) {
	if len(users) == 0 {
		fmt.Print("No users found.\n")
	} else {
		fmt.Print("Users:\n")
		for _, u := range users {
			fmt.Printf("%s (%s)\n", u.PrimaryEmail, u.Name.FullName)
		}
	}
}

func GetUserEmails(user *admin.User) (string, string) {
	var primEmail string
	var secEmail string

	for _, email := range user.Emails.([]interface{}) {
		if email.(map[string]interface{})["primary"] == true {
			primEmail = fmt.Sprint(email.(map[string]interface{})["address"])
		} else {
			secEmail = fmt.Sprint(email.(map[string]interface{})["address"])
		}
	}

	return primEmail, secEmail
}

// TODO: CreateNewUse creates a new user in GSuite via their API
func CreateNewUser(user *GSuiteUser) error {

	return nil
}

// TODO: DeleteUser deletes a user in GSuite via their API
func DeleteUser(users *GSuiteUser) error {

	return nil
}
