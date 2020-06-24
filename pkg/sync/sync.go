package sync

import (
	"context"
	"fmt"
	"log"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/glib"
	admin "google.golang.org/api/admin/directory/v1"
)

type UpdatedGsuite struct {
	// users
	usersToDelete []*admin.User
	usersToCreate []config.UserConfig
	usersToUpdate []config.UserConfig
	// groups
	groupsToDelete []*admin.Group
	groupsToCreate []config.GroupConfig
	groupsToUpdate []config.GroupConfig
	// orgUnits
	orgUnitsToDelete []*admin.OrgUnit
	orgUnitsToCreate []config.OrgUnitConfig
	orgUnitsToUpdate []config.OrgUnitConfig
}

func SyncConfiguration(ctx context.Context, cfg *config.Config, clientService *admin.Service, confirm bool) error {
	//var updatedConfig *UpdatedGsuite

	if err := SyncUsers(ctx, clientService, cfg, confirm); err != nil {
		return fmt.Errorf("failed to sync users: %v", err)
	}
	if err := SyncGroups(ctx, clientService, cfg, confirm); err != nil {
		return fmt.Errorf("failed to sync groups: %v", err)
	}

	return nil
}

// TODO: SWAP FOR SLICE CHECK ??
func SyncUsers(ctx context.Context, clientService *admin.Service, cfg *config.Config, confirm bool) error {
	var (
		usersToDelete []*admin.User
		usersToCreate []config.UserConfig
		usersToUpdate []config.UserConfig
	)

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
		// GET USERS TO DELETE & UPDATE
		for _, currentUser := range currentUsers {
			found := false
			for _, configUser := range configUsers {
				if configUser.PrimaryEmail == currentUser.PrimaryEmail {
					found = true
					// user is existing & should exist, so check if needs an update
					_, workEmail := glib.GetUserEmails(currentUser)
					if configUser.LastName != currentUser.Name.FamilyName ||
						configUser.FirstName != currentUser.Name.GivenName ||
						configUser.SecondaryEmail != workEmail {
						usersToUpdate = append(usersToUpdate, configUser)
					}
					break
				}
			}
			if !found {
				usersToDelete = append(usersToDelete, currentUser)
			}
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

	if confirm {
		if usersToCreate != nil {
			log.Println("✎ Creating...")
			for _, user := range usersToCreate {
				glib.CreateNewUser(*clientService, &user)
				log.Printf("\t+ user: %s\n", user.PrimaryEmail)
			}
		}
		if usersToDelete != nil {
			log.Println("✁ Deleting...")
			for _, user := range usersToDelete {
				glib.DeleteUser(*clientService, user)
				log.Printf("\t- user: %s\n", user.PrimaryEmail)
			}
		}
		if usersToUpdate != nil {
			log.Println("✎ Updating...")
			for _, user := range usersToUpdate {
				glib.UpdateUser(*clientService, &user)
				log.Printf("\t~ user: %s\n", user.PrimaryEmail)
			}
		}
	} else {
		if usersToDelete == nil {
			log.Println("✁ There is no users to delete.")
		} else {
			log.Println("✁ Found users to delete: ")
			for _, u := range usersToDelete {
				log.Printf("\t- %s %s\n", u.Name.GivenName, u.Name.FamilyName)
			}
		}

		if usersToCreate == nil {
			log.Println("✎ There is no users to create.")
		} else {
			log.Println("✎ Found users to create: ")
			for _, u := range usersToCreate {
				log.Printf("\t+ %s %s\n", u.FirstName, u.LastName)
			}
		}

		if usersToUpdate == nil {
			log.Println("✎ There is no users to update.")
		} else {
			log.Println("✎ Found users to update: ")
			for _, u := range usersToUpdate {
				log.Printf("\t~ %s %s\n", u.FirstName, u.LastName)
			}
		}
	}

	return nil
}

// groupUpdate holds a group config to update
// (Members array is not bounded to the Group object in the API)
// helper to avoid global vars
type groupUpdate struct {
	groupToUpdate   config.GroupConfig
	membersToAdd    []*config.MemberConfig
	membersToRemove []*admin.Member
}

// SyncGroups
func SyncGroups(ctx context.Context, clientService *admin.Service, cfg *config.Config, confirm bool) error {
	var (
		groupsToDelete []*admin.Group
		groupsToCreate []config.GroupConfig
		groupsToUpdate []groupUpdate
	)

	log.Println("⇄ Syncing groups")
	// get the current groups array
	currentGroups, err := glib.GetListOfGroups(clientService)

	if err != nil {
		return fmt.Errorf("⚠ failed to get current groups: %v", err)
	}
	// config defined groups
	configGroups := cfg.Groups

	if len(currentGroups) == 0 {
		log.Println("⚠ No groups found.")
	} else {
		// GET GROUPS TO DELETE & UPDATE
		for _, currGroup := range currentGroups {
			found := false
			for _, cfgGroup := range configGroups {
				if cfgGroup.Email == currGroup.Email {
					found = true
					// group is existing & should exist, so check if needs an update
					var upGroup groupUpdate
					upGroup.membersToAdd, upGroup.membersToRemove = SyncMembers(ctx, clientService, &cfgGroup, currGroup)
					if cfgGroup.Name != currGroup.Name ||
						cfgGroup.Description != currGroup.Description ||
						upGroup.membersToAdd != nil || upGroup.membersToRemove != nil {
						upGroup.groupToUpdate = cfgGroup
						groupsToUpdate = append(groupsToUpdate, upGroup)
					}
					break
				}
			}
			if !found {
				groupsToDelete = append(groupsToDelete, currGroup)
			}
		}

	}
	// GET GROUPS TO CREATE
	for _, cfgGroup := range configGroups {
		found := false
		for _, currGroup := range currentGroups {
			if currGroup.Email == cfgGroup.Email {
				found = true
				break
			}
		}
		if !found {
			groupsToCreate = append(groupsToCreate, cfgGroup)
		}

	}

	if confirm {
		if groupsToCreate != nil {
			log.Println("✎ Creating...")
			for _, gr := range groupsToCreate {
				glib.CreateGroup(*clientService, &gr)
				log.Printf("\t+ group: %s\n", gr.Name)
			}
		}
		if groupsToDelete != nil {
			log.Println("✁ Deleting...")
			for _, gr := range groupsToDelete {
				glib.DeleteGroup(*clientService, gr)
				log.Printf("\t- group: %s\n", gr.Name)
			}
		}
		if groupsToUpdate != nil {
			log.Println("✎ Updating...")
			for _, gr := range groupsToUpdate {
				//glib.UpdateGroup(*clientService, &g)
				log.Printf("\t~ group: %s\n", gr.groupToUpdate.Name)
				for _, mem := range gr.membersToAdd {
					log.Printf("\t\t+ %s \n", mem.Email)
					glib.AddNewMember(*clientService, gr.groupToUpdate.Email, mem)

				}
				for _, mem := range gr.membersToRemove {
					log.Printf("\t\t- %s \n", mem.Email)
					glib.RemoveMember(*clientService, gr.groupToUpdate.Email, mem)
				}
			}
		}
	} else {
		if groupsToDelete == nil {
			log.Println("✁ There is no groups to delete.")
		} else {
			log.Println("✁ Found groups to delete: ")
			for _, g := range groupsToDelete {
				log.Printf("\t- %s \n", g.Name)
			}
		}

		if groupsToCreate == nil {
			log.Println("✎ There is no groups to create.")
		} else {
			log.Println("✎ Found groups to create: ")
			for _, g := range groupsToCreate {
				log.Printf("\t+ %s\n", g.Name)
			}
		}

		if groupsToUpdate == nil {
			log.Println("✎ There is no groups to update.")
		} else {
			log.Println("✎ Found groups to update: ")
			for _, g := range groupsToUpdate {
				log.Printf("\t~ %s \n", g.groupToUpdate.Name)
				for _, mem := range g.membersToAdd {
					log.Printf("\t\t+ %s \n", mem.Email)
				}
				for _, mem := range g.membersToRemove {
					log.Printf("\t\t- %s \n", mem.Email)
				}
			}
		}

	}

	return nil
}

func SyncMembers(ctx context.Context, clientService *admin.Service, cfgGr *config.GroupConfig, curGr *admin.Group) ([]*config.MemberConfig, []*admin.Member) {
	var memToAdd []*config.MemberConfig
	var memToRemove []*admin.Member
	currentMembers, _ := glib.GetListOfMembers(clientService, curGr)
	// check members to add
	for _, member := range cfgGr.Members {
		foundMem := false
		for _, currMember := range currentMembers {
			if currMember.Email == member.Email {
				foundMem = true
				break
			}
		}
		if !foundMem {
			memToAdd = append(memToAdd, &member)
		}
	}

	// check members to remove
	for _, currMember := range currentMembers {
		foundMem := false
		for _, member := range cfgGr.Members {
			if currMember.Email == member.Email {
				foundMem = true
			}
		}
		if !foundMem {
			memToRemove = append(memToRemove, currMember)
		}
	}

	return memToAdd, memToRemove
}
