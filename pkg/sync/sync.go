package sync

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/glib"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/groupssettings/v1"
	"google.golang.org/api/licensing/v1"
)

func SyncConfiguration(ctx context.Context, cfg *config.Config, clientService *admin.Service, groupService *groupssettings.Service, licensingService *licensing.Service, confirm bool) error {

	if err := SyncOrgUnits(ctx, clientService, cfg, confirm); err != nil {
		return fmt.Errorf("failed to sync users: %v", err)
	}
	if err := SyncUsers(ctx, clientService, licensingService, cfg, confirm); err != nil {
		return fmt.Errorf("failed to sync users: %v", err)
	}
	if err := SyncGroups(ctx, clientService, groupService, cfg, confirm); err != nil {
		return fmt.Errorf("failed to sync groups: %v", err)
	}

	return nil
}

// TODO: SWAP FOR SLICE CHECK ??
func SyncUsers(ctx context.Context, clientService *admin.Service, licensingService *licensing.Service, cfg *config.Config, confirm bool) error {
	var (
		usersToDelete []*admin.User
		usersToCreate []config.UserConfig
		usersToUpdate []config.UserConfig
	)

	log.Println("⇄ Syncing users")
	// get the current users array
	currentUsers, err := glib.GetListOfUsers(*clientService)
	if err != nil {
		return err
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
					// get user licenses
					currentUserLicenses, err := glib.GetUserLicenses(licensingService, currentUser.PrimaryEmail)
					if err != nil {
						return err
					}
					currentUserConfig := glib.CreateConfigUserFromGSuite(currentUser, currentUserLicenses)
					if !reflect.DeepEqual(currentUserConfig, configUser) {
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
			log.Println("Creating...")
			for _, user := range usersToCreate {
				err := glib.CreateUser(*clientService, *licensingService, &user)
				if err != nil {
					return fmt.Errorf("⚠ Failed to create user %s: %v.", user.PrimaryEmail, err)
				} else {
					log.Printf(" ✎  user: %s\n", user.PrimaryEmail)
				}
			}
		}
		if usersToDelete != nil {
			log.Println("Deleting...")
			for _, user := range usersToDelete {
				err := glib.DeleteUser(*clientService, user)
				if err != nil {
					return fmt.Errorf("⚠ Failed to delete user %s: %v.", user.PrimaryEmail, err)
				} else {
					log.Printf(" ✁  user: %s\n", user.PrimaryEmail)
				}
			}
		}
		if usersToUpdate != nil {
			log.Println("Updating...")
			for _, user := range usersToUpdate {
				err := glib.UpdateUser(*clientService, *licensingService, &user)
				if err != nil {
					return fmt.Errorf("⚠ Failed to update user %s: %v.", user.PrimaryEmail, err)
				} else {
					log.Printf(" ✎  user: %s\n", user.PrimaryEmail)
				}
			}
		}
	} else {
		if usersToDelete == nil {
			log.Println("There is no users to delete.")
		} else {
			log.Println("Found users to delete: ")
			for _, u := range usersToDelete {
				log.Printf(" ✁  %s\n", u.PrimaryEmail)
			}
		}

		if usersToCreate == nil {
			log.Println("There is no users to create.")
		} else {
			log.Println("Found users to create: ")
			for _, u := range usersToCreate {
				log.Printf(" ✎  %s\n", u.PrimaryEmail)
			}
		}

		if usersToUpdate == nil {
			log.Println("There is no users to update.")
		} else {
			log.Println("Found users to update: ")
			for _, u := range usersToUpdate {
				log.Printf(" ✎  %s\n", u.PrimaryEmail)
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
func SyncGroups(ctx context.Context, clientService *admin.Service, groupService *groupssettings.Service, cfg *config.Config, confirm bool) error {
	var (
		groupsToDelete []*admin.Group
		groupsToCreate []config.GroupConfig
		groupsToUpdate []groupUpdate
	)

	log.Println("⇄ Syncing groups")
	// get the current groups array
	currentGroups, err := glib.GetListOfGroups(clientService)
	if err != nil {
		return err
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
					currentMembers, err := glib.GetListOfMembers(clientService, currGroup)
					if err != nil {
						return err
					}
					currentSettings, err := glib.GetSettingOfGroup(groupService, currGroup.Email)
					if err != nil {
						return err
					}
					currentGroupConfig, err := glib.CreateConfigGroupFromGSuite(currGroup, currentMembers, currentSettings)
					if err != nil {
						return err
					}
					if !reflect.DeepEqual(currentGroupConfig, cfgGroup) {
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
			log.Println("Creating...")
			for _, gr := range groupsToCreate {
				err := glib.CreateGroup(*clientService, *groupService, &gr)
				if err != nil {
					return fmt.Errorf("⚠ Failed to create a group %s: %v.", gr.Name, err)
				} else {
					log.Printf(" ✎  group: %s\n", gr.Name)
				}
			}
		}
		if groupsToDelete != nil {
			log.Println("Deleting...")
			for _, gr := range groupsToDelete {
				err := glib.DeleteGroup(*clientService, gr)
				if err != nil {
					return fmt.Errorf("⚠ Failed to delete a group %s: %v.", gr.Name, err)
				} else {
					log.Printf(" ✁  group: %s\n", gr.Name)
				}
			}
		}
		if groupsToUpdate != nil {
			log.Println("Updating...")
			for _, gr := range groupsToUpdate {
				err := glib.UpdateGroup(*clientService, *groupService, &gr.groupToUpdate)
				if err != nil {
					return fmt.Errorf("⚠ Failed to update a group: %v.", err)
				} else {
					log.Printf(" ✎  group: %s\n", gr.groupToUpdate.Name)
				}

				for _, mem := range gr.membersToAdd {
					err := glib.AddNewMember(*clientService, gr.groupToUpdate.Email, mem)
					if err != nil {
						return fmt.Errorf("⚠ Failed to add a member to a group: %v.", err)
					} else {
						log.Printf(" ✎  adding member: %s \n", mem.Email)
					}
				}
				for _, mem := range gr.membersToRemove {
					err := glib.RemoveMember(*clientService, gr.groupToUpdate.Email, mem)
					if err != nil {
						return fmt.Errorf("⚠ Failed to add a member to a group: %v.", err)
					} else {
						log.Printf(" ✁  removing member: %s \n", mem.Email)
					}
				}
				for _, mem := range gr.membersToUpdate {
					err := glib.UpdateMembership(*clientService, gr.groupToUpdate.Email, mem)
					if err != nil {
						return fmt.Errorf("⚠ Failed to update membership in a group: %v.", err)
					} else {
						log.Printf(" ✎  updating membership: %s \n", mem.Email)
					}
				}
			}
		}
	} else {
		if groupsToDelete == nil {
			log.Println("There is no groups to delete.")
		} else {
			log.Println("Found groups to delete: ")
			for _, g := range groupsToDelete {
				log.Printf(" ✁  %s \n", g.Name)
			}
		}

		if groupsToCreate == nil {
			log.Println("There is no groups to create.")
		} else {
			log.Println("Found groups to create: ")
			for _, g := range groupsToCreate {
				log.Printf(" ✎  %s\n", g.Name)
			}
		}

		if groupsToUpdate == nil {
			log.Println("There is no groups to update.")
		} else {
			log.Println("Found groups to update: ")
			for _, g := range groupsToUpdate {
				log.Printf(" ✎  %s \n", g.groupToUpdate.Name)
				for _, mem := range g.membersToAdd {
					log.Printf(" ✎  member to add: %s \n", mem.Email)
				}
				for _, mem := range g.membersToRemove {
					log.Printf(" ✁  member to remove: %s \n", mem.Email)
				}
				for _, mem := range g.membersToUpdate {
					log.Printf(" ✎  member to update: %s \n", mem.Email)
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
		return err
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
			log.Println("Creating...")
			for _, ou := range ouToCreate {
				err := glib.CreateOrgUnit(*clientService, &ou)
				if err != nil {
					return err
				}
				log.Printf(" ✎  org unit: %s\n", ou.Name)
			}
		}
		if ouToDelete != nil {
			log.Println("Deleting...")
			for _, ou := range ouToDelete {
				err := glib.DeleteOrgUnit(*clientService, ou)
				if err != nil {
					return err
				}
				log.Printf(" ✁  org unit: %s\n", ou.Name)
			}
		}
		if ouToUpdate != nil {
			log.Println("Updating...")
			for _, ou := range ouToUpdate {
				err := glib.UpdateOrgUnit(*clientService, &ou)
				if err != nil {
					return err
				}
				log.Printf(" ✎  org unit: %s \n", ou.Name)
			}
		}
	} else {
		if ouToDelete == nil {
			log.Println("There is no org units to delete.")
		} else {
			log.Println("Found org units to delete: ")
			for _, ou := range ouToDelete {
				log.Printf(" ✁  %s \n", ou.Name)
			}
		}

		if ouToCreate == nil {
			log.Println("There is no org units to create.")
		} else {
			log.Println("Found org units to create: ")
			for _, ou := range ouToCreate {
				log.Printf(" ✎  %s \n", ou.Name)
			}
		}

		if ouToUpdate == nil {
			log.Println("There is no org units to update.")
		} else {
			log.Println("Found org units to update: ")
			for _, ou := range ouToUpdate {
				log.Printf(" ✎  %s\n", ou.Name)
			}
		}
	}
	return nil

}
