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

	directoryv1 "google.golang.org/api/admin/directory/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/glib"
)

func SyncUsers(
	ctx context.Context,
	directorySrv *glib.DirectoryService,
	licensingSrv *glib.LicensingService,
	cfg *config.Config,
	licenseStatus *glib.LicenseStatus,
	confirm bool,
) (bool, error) {
	changes := false

	log.Println("⇄ Syncing users…")

	liveUsers, err := directorySrv.ListUsers(ctx)
	if err != nil {
		return changes, err
	}

	liveEmails := sets.NewString()

	for _, liveUser := range liveUsers {
		liveEmails.Insert(liveUser.PrimaryEmail)

		found := false

		for _, expectedUser := range cfg.Users {
			if expectedUser.PrimaryEmail == liveUser.PrimaryEmail {
				found = true

				currentUserLicenses := licenseStatus.GetLicensesForUser(liveUser)

				currentAliases, err := directorySrv.GetUserAliases(ctx, liveUser)
				if err != nil {
					return changes, fmt.Errorf("failed to fetch aliases: %v", err)
				}

				if userUpToDate(expectedUser, liveUser, currentUserLicenses, currentAliases) {
					// no update needed
					log.Printf("  ✓ %s", expectedUser.PrimaryEmail)
				} else {
					// update it
					changes = true
					log.Printf("  ✎ %s", expectedUser.PrimaryEmail)

					updatedUser := liveUser
					if confirm {
						apiUser := config.ToGSuiteUser(&expectedUser)
						updatedUser, err = directorySrv.UpdateUser(ctx, liveUser, apiUser)
						if err != nil {
							return changes, fmt.Errorf("failed to update user: %v", err)
						}
					}

					if err := syncUserAliases(ctx, directorySrv, &expectedUser, updatedUser, currentAliases, confirm); err != nil {
						return changes, fmt.Errorf("failed to sync aliases: %v", err)
					}

					if err := syncUserLicenses(ctx, licensingSrv, &expectedUser, updatedUser, licenseStatus, confirm); err != nil {
						return changes, fmt.Errorf("failed to sync licenses: %v", err)
					}
				}

				break
			}
		}

		if !found {
			changes = true
			log.Printf("  - %s", liveUser.PrimaryEmail)

			if confirm {
				if err := directorySrv.DeleteUser(ctx, liveUser); err != nil {
					return changes, fmt.Errorf("failed to delete user: %v", err)
				}
			}
		}
	}

	for _, expectedUser := range cfg.Users {
		if !liveEmails.Has(expectedUser.PrimaryEmail) {
			changes = true
			log.Printf("  + %s", expectedUser.PrimaryEmail)

			var createdUser *directoryv1.User

			if confirm {
				apiUser := config.ToGSuiteUser(&expectedUser)
				createdUser, err = directorySrv.CreateUser(ctx, apiUser)
				if err != nil {
					return changes, fmt.Errorf("failed to create user: %v", err)
				}
			}

			if err := syncUserAliases(ctx, directorySrv, &expectedUser, createdUser, nil, confirm); err != nil {
				return changes, fmt.Errorf("failed to sync aliases: %v", err)
			}

			if err := syncUserLicenses(ctx, licensingSrv, &expectedUser, createdUser, licenseStatus, confirm); err != nil {
				return changes, fmt.Errorf("failed to sync licenses: %v", err)
			}
		}
	}

	return changes, nil
}

func syncUserAliases(
	ctx context.Context,
	directorySrv *glib.DirectoryService,
	expectedUser *config.User,
	liveUser *directoryv1.User,
	liveAliases []string,
	confirm bool,
) error {
	expectedAliases := sets.NewString(expectedUser.Aliases...)
	liveAliasesSet := sets.NewString(liveAliases...)

	for _, liveAlias := range liveAliases {
		if !expectedAliases.Has(liveAlias) {
			log.Printf("    - alias %s", liveAlias)

			if confirm {
				if err := directorySrv.DeleteUserAlias(ctx, liveUser, liveAlias); err != nil {
					return fmt.Errorf("unable to delete alias: %v", err)
				}
			}
		}
	}

	for _, expectedAlias := range expectedAliases.List() {
		if !liveAliasesSet.Has(expectedAlias) {
			log.Printf("    + alias %s", expectedAlias)

			if confirm {
				if err := directorySrv.CreateUserAlias(ctx, liveUser, expectedAlias); err != nil {
					return fmt.Errorf("unable to create alias: %v", err)
				}
			}
		}
	}

	return nil
}
