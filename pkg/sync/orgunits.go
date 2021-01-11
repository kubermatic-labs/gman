package sync

import (
	"context"
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/glib"
)

func SyncOrgUnits(
	ctx context.Context,
	directorySrv *glib.DirectoryService,
	cfg *config.Config,
	confirm bool,
) error {
	log.Println("⇄ Syncing organizational units…")

	currentUnits, err := directorySrv.ListOrgUnits(ctx)
	if err != nil {
		return err
	}

	currentNames := sets.NewString()

	for _, current := range currentUnits {
		currentNames.Insert(current.Name)

		found := false

		for _, configured := range cfg.OrgUnits {
			if configured.Name == current.Name {
				found = true

				if orgUnitUpToDate(configured, current) {
					// no update needed
					log.Printf("  ✓ %s", configured.Name)
				} else {
					// update it
					log.Printf("  ✎ %s", configured.Name)

					if confirm {
						err := directorySrv.UpdateOrgUnit(ctx, &configured)
						if err != nil {
							return fmt.Errorf("failed to update org unit: %v", err)
						}
					}
				}

				break
			}
		}

		if !found {
			log.Printf("  ✁ %s", current.Name)

			if confirm {
				err := directorySrv.DeleteOrgUnit(ctx, current)
				if err != nil {
					return fmt.Errorf("failed to delete org unit: %v", err)
				}
			}
		}
	}

	for _, configured := range cfg.OrgUnits {
		if !currentNames.Has(configured.Name) {
			log.Printf("  + %s", configured.Name)

			if confirm {
				err := directorySrv.CreateOrgUnit(ctx, &configured)
				if err != nil {
					return fmt.Errorf("failed to create org unit: %v", err)
				}
			}
		}
	}

	return nil
}
