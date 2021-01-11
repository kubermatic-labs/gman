package sync

import (
	"context"
	"fmt"
	"log"

	directoryv1 "google.golang.org/api/admin/directory/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/glib"
)

func SyncGroups(
	ctx context.Context,
	directorySrv *glib.DirectoryService,
	groupsSettingsSrv *glib.GroupsSettingsService,
	cfg *config.Config,
	confirm bool,
) error {
	log.Println("⇄ Syncing groups…")

	currentGroups, err := directorySrv.ListGroups(ctx)
	if err != nil {
		return err
	}

	currentGroupEmails := sets.NewString()

	for _, current := range currentGroups {
		currentGroupEmails.Insert(current.Email)

		found := false

		for _, configured := range cfg.Groups {
			if configured.Email == current.Email {
				found = true

				currentMembers, err := directorySrv.ListMembers(ctx, current)
				if err != nil {
					return fmt.Errorf("failed to fetch members: %v", err)
				}

				currentSettings, err := groupsSettingsSrv.GetSettings(ctx, current.Email)
				if err != nil {
					return fmt.Errorf("failed to fetch group settings: %v", err)
				}

				if groupUpToDate(configured, current, currentMembers, currentSettings) {
					// no update needed
					log.Printf("  ✓ %s", configured.Email)
				} else {
					// update it
					log.Printf("  ✎ %s", configured.Email)

					group, settings := glib.CreateGSuiteGroupFromConfig(&configured)

					if confirm {
						group, err = directorySrv.UpdateGroup(ctx, group)
						if err != nil {
							return fmt.Errorf("failed to update group: %v", err)
						}

						if _, err := groupsSettingsSrv.UpdateSettings(ctx, group, settings); err != nil {
							return fmt.Errorf("failed to update group settings: %v", err)
						}
					}

					if err := syncGroupMembers(ctx, directorySrv, &configured, currentMembers, confirm); err != nil {
						return fmt.Errorf("failed to sync members: %v", err)
					}
				}

				break
			}
		}

		if !found {
			log.Printf("  ✁ %s", current.Email)

			if confirm {
				if err := directorySrv.DeleteGroup(ctx, current); err != nil {
					return fmt.Errorf("failed to delete group: %v", err)
				}
			}
		}
	}

	for _, configured := range cfg.Groups {
		if !currentGroupEmails.Has(configured.Email) {
			group, settings := glib.CreateGSuiteGroupFromConfig(&configured)

			log.Printf("  + %s", configured.Email)

			if confirm {
				group, err = directorySrv.CreateGroup(ctx, group)
				if err != nil {
					return fmt.Errorf("failed to create group: %v", err)
				}

				if _, err := groupsSettingsSrv.UpdateSettings(ctx, group, settings); err != nil {
					return fmt.Errorf("failed to update group settings: %v", err)
				}
			}

			if err := syncGroupMembers(ctx, directorySrv, &configured, nil, confirm); err != nil {
				return fmt.Errorf("failed to sync members: %v", err)
			}
		}
	}

	return nil
}

func getConfiguredMember(group *config.Group, member *directoryv1.Member) *config.Member {
	for _, m := range group.Members {
		if m.Email == member.Email {
			return &m
		}
	}

	return nil
}

func syncGroupMembers(
	ctx context.Context,
	directorySrv *glib.DirectoryService,
	configuredGroup *config.Group,
	liveMembers []*directoryv1.Member,
	confirm bool,
) error {
	liveMemberEmails := sets.NewString()

	for _, liveMember := range liveMembers {
		liveMemberEmails.Insert(liveMember.Email)

		expectedMember := getConfiguredMember(configuredGroup, liveMember)

		if expectedMember == nil {
			log.Printf("    - %s", liveMember.Email)

			if confirm {
				if err := directorySrv.RemoveMember(ctx, configuredGroup.Email, liveMember); err != nil {
					return fmt.Errorf("unable to remove member: %v", err)
				}
			}
		} else if !memberUpToDate(*expectedMember, liveMember) {
			log.Printf("    ✎ %s", liveMember.Email)

			if confirm {
				if err := directorySrv.UpdateMembership(ctx, configuredGroup.Email, expectedMember); err != nil {
					return fmt.Errorf("unable to update membership: %v", err)
				}
			}
		}
	}

	for _, configuredMember := range configuredGroup.Members {
		if !liveMemberEmails.Has(configuredMember.Email) {
			log.Printf("    + %s", configuredMember.Email)

			if confirm {
				if err := directorySrv.AddNewMember(ctx, configuredGroup.Email, &configuredMember); err != nil {
					return fmt.Errorf("unable to add member: %v", err)
				}
			}
		}
	}

	return nil
}
