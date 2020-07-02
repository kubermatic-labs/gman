package sync

import (
	"context"
	"fmt"
	"log"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/glib"
	admin "google.golang.org/api/admin/directory/v1"
)

func SyncConfiguration(ctx context.Context, cfg *config.Config, clientService *admin.Service, confirm bool) error {

	if err := SyncOrgUnits(ctx, clientService, cfg, confirm); err != nil {
		return fmt.Errorf("failed to sync users: %v", err)
	}
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
						configUser.SecondaryEmail != workEmail ||
						configUser.OrgUnitPath != currentUser.OrgUnitPath {
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
				glib.CreateUser(*clientService, &user)
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
	membersToUpdate []*config.MemberConfig
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
					upGroup.membersToAdd, upGroup.membersToRemove, upGroup.membersToUpdate = SyncMembers(ctx, clientService, &cfgGroup, currGroup)
					if cfgGroup.Name != currGroup.Name ||
						cfgGroup.Description != currGroup.Description ||
						upGroup.membersToAdd != nil || upGroup.membersToRemove != nil ||
						upGroup.membersToUpdate != nil {
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
				glib.UpdateGroup(*clientService, &gr.groupToUpdate)
				log.Printf("\t~ group: %s\n", gr.groupToUpdate.Name)

				for _, mem := range gr.membersToAdd {
					log.Printf("\t\t+ %s \n", mem.Email)
					glib.AddNewMember(*clientService, gr.groupToUpdate.Email, mem)

				}
				for _, mem := range gr.membersToRemove {
					log.Printf("\t\t- %s \n", mem.Email)
					glib.RemoveMember(*clientService, gr.groupToUpdate.Email, mem)
				}
				for _, mem := range gr.membersToUpdate {
					log.Printf("\t\t~ %s \n", mem.Email)
					glib.UpdateMembership(*clientService, gr.groupToUpdate.Email, mem)

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

func SyncMembers(ctx context.Context, clientService *admin.Service, cfgGr *config.GroupConfig, curGr *admin.Group) ([]*config.MemberConfig, []*admin.Member, []*config.MemberConfig) {
	var memToAdd []*config.MemberConfig
	var memToUpdate []*config.MemberConfig
	var memToRemove []*admin.Member
	currentMembers, _ := glib.GetListOfMembers(clientService, curGr)
	// check members to add
	for _, member := range cfgGr.Members {
		foundMem := false
		for _, currMember := range currentMembers {
			if currMember.Email == member.Email {
				foundMem = true
				// check for update
				if currMember.Role != member.Role {
					memToUpdate = append(memToUpdate, &member)
				}
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

	return memToAdd, memToRemove, memToUpdate
}

func SyncOrgUnits(ctx context.Context, clientService *admin.Service, cfg *config.Config, confirm bool) error {
	var (
		ouToDelete []*admin.OrgUnit
		ouToCreate []config.OrgUnitConfig
		ouToUpdate []config.OrgUnitConfig
	)

	log.Println("⇄ Syncing organizational units")
	// get the current users array
	currentOus, err := glib.GetListOfOrgUnits(clientService)
	if err != nil {
		return fmt.Errorf("⚠ failed to get current org units: %v", err)
	}
	// config defined users
	configOus := cfg.OrgUnits

	if len(currentOus) == 0 {
		log.Println("⚠ No organizational units found.")
	} else {
		// GET ORG UNITS TO DELETE & UPDATE
		for _, currentOu := range currentOus {
			found := false
			for _, configOu := range configOus {
				if configOu.Name == currentOu.Name {
					found = true
					// OU is existing & should exist, so check if needs an update
					if configOu.Description != currentOu.Description ||
						configOu.ParentOrgUnitPath != currentOu.ParentOrgUnitPath ||
						configOu.BlockInheritance != currentOu.BlockInheritance {
						ouToUpdate = append(ouToUpdate, configOu)
					}
					break
				}
			}
			if !found {
				ouToDelete = append(ouToDelete, currentOu)
			}
		}
	}
	// GET ORG UNITS TO CREATE
	for _, configOu := range configOus {
		found := false
		for _, currentOu := range currentOus {
			if currentOu.Name == configOu.Name {
				found = true
				break
			}
		}
		if !found {
			ouToCreate = append(ouToCreate, configOu)
		}
	}

	if confirm {
		if ouToCreate != nil {
			log.Println("✎ Creating...")
			for _, ou := range ouToCreate {
				glib.CreateOrgUnit(*clientService, &ou)
				log.Printf("\t+ org unit: %s\n", ou.Name)
			}
		}
		if ouToDelete != nil {
			log.Println("✁ Deleting...")
			for _, ou := range ouToDelete {
				glib.DeleteOrgUnit(*clientService, ou)
				log.Printf("\t- org unit: %s\n", ou.Name)
			}
		}
		if ouToUpdate != nil {
			log.Println("✎ Updating...")
			for _, ou := range ouToUpdate {
				glib.UpdateOrgUnit(*clientService, &ou)
				log.Printf("\t~ org unit: %s \n", ou.Name)
			}
		}
	} else {
		if ouToDelete == nil {
			log.Println("✁ There is no org units to delete.")
		} else {
			log.Println("✁ Found org units to delete: ")
			for _, ou := range ouToDelete {
				log.Printf("\t- %s \n", ou.Name)
			}
		}

		if ouToCreate == nil {
			log.Println("✎ There is no org units to create.")
		} else {
			log.Println("✎ Found org units to create: ")
			for _, ou := range ouToCreate {
				log.Printf("\t+ %s \n", ou.Name)
			}
		}

		if ouToUpdate == nil {
			log.Println("✎ There is no org units to update.")
		} else {
			log.Println("✎ Found org units to update: ")
			for _, ou := range ouToUpdate {
				log.Printf("\t~ %s\n", ou.Name)
			}
		}
	}
	return nil

}
