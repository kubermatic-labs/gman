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
) error {
	log.Println("⇄ Syncing users…")

	currentUsers, err := directorySrv.ListUsers(ctx)
	if err != nil {
		return err
	}

	currentEmails := sets.NewString()

	for _, current := range currentUsers {
		currentEmails.Insert(current.PrimaryEmail)

		found := false

		for _, configured := range cfg.Users {
			if configured.PrimaryEmail == current.PrimaryEmail {
				found = true

				currentUserLicenses := licenseStatus.GetLicensesForUser(current)

				currentAliases, err := directorySrv.GetUserAliases(ctx, &configured)
				if err != nil {
					return fmt.Errorf("failed to fetch aliases: %v", err)
				}

				if userUpToDate(configured, current, currentUserLicenses, currentAliases) {
					// no update needed
					log.Printf("  ✓ %s", configured.PrimaryEmail)
				} else {
					// update it
					log.Printf("  ✎ %s", configured.PrimaryEmail)

					updatedUser := current
					if confirm {
						updatedUser, err = directorySrv.UpdateUser(ctx, &configured)
						if err != nil {
							return fmt.Errorf("failed to update user: %v", err)
						}
					}

					if err := syncUserAliases(ctx, directorySrv, &configured, updatedUser, currentAliases, confirm); err != nil {
						return fmt.Errorf("failed to sync aliases: %v", err)
					}

					if err := syncUserLicenses(ctx, licensingSrv, &configured, updatedUser, licenseStatus, confirm); err != nil {
						return fmt.Errorf("failed to sync licenses: %v", err)
					}
				}

				break
			}
		}

		if !found {
			log.Printf("  ✁ %s", current.PrimaryEmail)

			if confirm {
				if err := directorySrv.DeleteUser(ctx, current); err != nil {
					return fmt.Errorf("failed to delete user: %v", err)
				}
			}
		}
	}

	for _, configured := range cfg.Users {
		if !currentEmails.Has(configured.PrimaryEmail) {
			log.Printf("  + %s", configured.PrimaryEmail)

			var createdUser *directoryv1.User

			if confirm {
				createdUser, err = directorySrv.CreateUser(ctx, &configured)
				if err != nil {
					return fmt.Errorf("failed to create user: %v", err)
				}
			}

			if err := syncUserAliases(ctx, directorySrv, &configured, createdUser, nil, confirm); err != nil {
				return fmt.Errorf("failed to sync aliases: %v", err)
			}

			if err := syncUserLicenses(ctx, licensingSrv, &configured, createdUser, licenseStatus, confirm); err != nil {
				return fmt.Errorf("failed to sync licenses: %v", err)
			}
		}
	}

	return nil
}

func syncUserAliases(
	ctx context.Context,
	directorySrv *glib.DirectoryService,
	configuredUser *config.User,
	liveUser *directoryv1.User,
	liveAliases []string,
	confirm bool,
) error {
	configuredAliases := sets.NewString(configuredUser.Aliases...)
	liveAliasesSet := sets.NewString(liveAliases...)

	for _, liveAlias := range liveAliases {
		if !configuredAliases.Has(liveAlias) {
			log.Printf("    - alias %s", liveAlias)

			if confirm {
				if err := directorySrv.DeleteUserAlias(ctx, configuredUser, liveAlias); err != nil {
					return fmt.Errorf("unable to delete alias: %v", err)
				}
			}
		}
	}

	for _, configuredAlias := range configuredAliases.List() {
		if !liveAliasesSet.Has(configuredAlias) {
			log.Printf("    + alias %s", configuredAlias)

			if confirm {
				if err := directorySrv.CreateUserAlias(ctx, configuredUser, configuredAlias); err != nil {
					return fmt.Errorf("unable to create alias: %v", err)
				}
			}
		}
	}

	return nil
}
