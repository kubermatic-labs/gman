package sync

import (
	"context"
	"fmt"
	"log"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/glib"
	admin "google.golang.org/api/admin/directory/v1"
)

// TODO: everything ;_;

func SyncConfiguration(ctx context.Context, cfg *config.Config, clientService *admin.Service, confirm bool) error {

	if err := SyncUsers(ctx, clientService, cfg, confirm); err != nil {
		return fmt.Errorf("failed to export users: %v", err)
	}

	return nil
}

// TODO: SWAP FOR SLICE CHECK ??
func SyncUsers(ctx context.Context, clientService *admin.Service, cfg *config.Config, confirm bool) error {
	var usersToDelete []*admin.User
	var usersToCreate []config.UserConfig

	log.Println("⇄ Syncing users")
	// get the current users array
	currentUsers, err := glib.GetListOfUsers(*clientService)
	if err != nil {
		return fmt.Errorf("⚠ failed to get current users: %v", err)
	}
	// config defined users
	configUsers := cfg.Users

	if len(currentUsers) == 0 {
		log.Println("⚠ No users found.")
	} else {
		// GET USERS TO DELETE
		for _, currentUser := range currentUsers {
			found := false
			for _, configUser := range configUsers {
				if configUser.PrimaryEmail == currentUser.PrimaryEmail {
					found = true
					break
				}
			}

			if !found {
				usersToDelete = append(usersToDelete, currentUser)
			}
		}

		// GET USERS TO CREATE
		for _, configUser := range configUsers {
			found := false
			for _, currentUser := range currentUsers {
				if currentUser.PrimaryEmail == configUser.PrimaryEmail {
					found = true
					break
				}
			}
			if !found {
				usersToCreate = append(usersToCreate, configUser)
			}
		}
	}

	log.Println("Found users to delete: ")
	for _, u := range usersToDelete {
		fmt.Printf("  - %s %s\n", u.Name.GivenName, u.Name.FamilyName)
	}

	log.Println("Found users to create: ")
	for _, u := range usersToCreate {
		fmt.Printf("  + %s %s\n", u.FirstName, u.LastName)
	}

	if confirm {
		for _, user := range usersToCreate {
			glib.CreateNewUser(*clientService, &user)
		}
		for _, user := range usersToDelete {
			glib.DeleteUser(*clientService, user)
		}

	}

	return nil
}
