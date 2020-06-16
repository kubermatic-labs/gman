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

	//PrintUsers(request.Users)

	return request.Users, nil
}

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

// TODO: CreateNewUse creates a new user in GSuite via their API
func CreateNewUser(srv admin.Service, user *config.UserConfig) error {
	// gen a pass
	pass, err := password.Generate(20, 5, 5, false, false)
	if err != nil {
		log.Fatalf("Unable to generate password: %v", err)
	}

	fmt.Printf("Create user: %s (%s)\n", user.FirstName, user.PrimaryEmail)
	fmt.Println(pass)
	newUser := admin.User{
		Name: &admin.UserName{
			GivenName:  user.FirstName,
			FamilyName: user.LastName,
		},
		PrimaryEmail: user.PrimaryEmail,
		Emails: &admin.UserEmail{
			Address: user.SecondaryEmail, // FIX IT IT DOESNT WORK
			Primary: false,
			Type:    "work",
		},
		Password:                  pass,
		ChangePasswordAtNextLogin: true,
	}

	request, err := srv.Users.Insert(&newUser).Do()
	if err != nil {
		log.Fatalf("Unable to create a user: %v", err)
	}
	fmt.Println(request)
	//PrintUsers(request.Users)

	return nil
}

// TODO: DeleteUser deletes a user in GSuite via their API
func DeleteUser(user *admin.User) error {
	fmt.Printf(" still test but im pretending to delete %s (%s)\n", user.Name.FullName, user.PrimaryEmail)
	return nil
}
