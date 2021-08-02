/*
Copyright 2021 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sync

import (
	"context"
	"fmt"
	"log"
	"sort"

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
) (bool, error) {
	changes := false

	log.Println("⇄ Syncing groups…")

	liveGroups, err := directorySrv.ListGroups(ctx)
	if err != nil {
		return changes, err
	}

	liveGroupEmails := sets.NewString()

	sort.Slice(liveGroups, func(i, j int) bool {
		return liveGroups[i].Email < liveGroups[j].Email
	})

	for _, liveGroup := range liveGroups {
		liveGroupEmails.Insert(liveGroup.Email)

		found := false

		for _, expectedGroup := range cfg.Groups {
			if expectedGroup.Email == liveGroup.Email {
				found = true

				liveMembers, err := directorySrv.ListMembers(ctx, liveGroup)
				if err != nil {
					return changes, fmt.Errorf("failed to fetch members: %v", err)
				}

				liveSettings, err := groupsSettingsSrv.GetSettings(ctx, liveGroup.Email)
				if err != nil {
					return changes, fmt.Errorf("failed to fetch group settings: %v", err)
				}

				if groupUpToDate(expectedGroup, liveGroup, liveMembers, liveSettings) {
					// no update needed
					log.Printf("  ✓ %s", expectedGroup.Email)
				} else {
					// update it
					changes = true
					log.Printf("  ✎ %s", expectedGroup.Email)

					group, settings := config.ToGSuiteGroup(&expectedGroup)

					if confirm {
						group, err = directorySrv.UpdateGroup(ctx, liveGroup, group)
						if err != nil {
							return changes, fmt.Errorf("failed to update group: %v", err)
						}

						if _, err := groupsSettingsSrv.UpdateSettings(ctx, group, settings); err != nil {
							return changes, fmt.Errorf("failed to update group settings: %v", err)
						}
					}

					if err := syncGroupMembers(ctx, directorySrv, &expectedGroup, group, liveMembers, confirm); err != nil {
						return changes, fmt.Errorf("failed to sync members: %v", err)
					}
				}

				break
			}
		}

		if !found {
			changes = true
			log.Printf("  - %s", liveGroup.Email)

			if confirm {
				if err := directorySrv.DeleteGroup(ctx, liveGroup); err != nil {
					return changes, fmt.Errorf("failed to delete group: %v", err)
				}
			}
		}
	}

	for _, expectedGroup := range cfg.Groups {
		if !liveGroupEmails.Has(expectedGroup.Email) {
			changes = true
			log.Printf("  + %s", expectedGroup.Email)

			group, settings := config.ToGSuiteGroup(&expectedGroup)

			if confirm {
				group, err = directorySrv.CreateGroup(ctx, group)
				if err != nil {
					return changes, fmt.Errorf("failed to create group: %v", err)
				}

				if _, err := groupsSettingsSrv.UpdateSettings(ctx, group, settings); err != nil {
					return changes, fmt.Errorf("failed to update group settings: %v", err)
				}
			}

			if err := syncGroupMembers(ctx, directorySrv, &expectedGroup, group, nil, confirm); err != nil {
				return changes, fmt.Errorf("failed to sync members: %v", err)
			}
		}
	}

	return changes, nil
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
	expectedGroup *config.Group,
	liveGroup *directoryv1.Group,
	liveMembers []*directoryv1.Member,
	confirm bool,
) error {
	liveMemberEmails := sets.NewString()

	for _, liveMember := range liveMembers {
		liveMemberEmails.Insert(liveMember.Email)

		expectedMember := getConfiguredMember(expectedGroup, liveMember)

		if expectedMember == nil {
			log.Printf("    - %s", liveMember.Email)

			if confirm {
				if err := directorySrv.RemoveMember(ctx, liveGroup, liveMember); err != nil {
					return fmt.Errorf("unable to remove member: %v", err)
				}
			}
		} else if !memberUpToDate(*expectedMember, liveMember) {
			log.Printf("    ✎ %s", liveMember.Email)

			if confirm {
				member := config.ToGSuiteGroupMember(expectedMember, liveMember)
				if err := directorySrv.UpdateMembership(ctx, liveGroup, member); err != nil {
					return fmt.Errorf("unable to update membership: %v", err)
				}
			}
		}
	}

	for _, expectedMember := range expectedGroup.Members {
		if !liveMemberEmails.Has(expectedMember.Email) {
			log.Printf("    + %s", expectedMember.Email)

			if confirm {
				member := config.ToGSuiteGroupMember(&expectedMember, nil)
				if err := directorySrv.AddNewMember(ctx, liveGroup, member); err != nil {
					return fmt.Errorf("unable to add member: %v", err)
				}
			}
		}
	}

	return nil
}
