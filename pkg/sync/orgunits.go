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

	liveOrgUnits, err := directorySrv.ListOrgUnits(ctx)
	if err != nil {
		return err
	}

	liveNames := sets.NewString()

	for _, liveOrgUnit := range liveOrgUnits {
		liveNames.Insert(liveOrgUnit.Name)

		found := false

		for _, expectedOrgUnit := range cfg.OrgUnits {
			if expectedOrgUnit.Name == liveOrgUnit.Name {
				found = true

				if orgUnitUpToDate(expectedOrgUnit, liveOrgUnit) {
					// no update needed
					log.Printf("  ✓ %s", expectedOrgUnit.Name)
				} else {
					// update it
					log.Printf("  ✎ %s", expectedOrgUnit.Name)

					if confirm {
						apiOrgUnit := config.ToGSuiteOrgUnit(&expectedOrgUnit)
						if err := directorySrv.UpdateOrgUnit(ctx, apiOrgUnit); err != nil {
							return fmt.Errorf("failed to update org unit: %v", err)
						}
					}
				}

				break
			}
		}

		if !found {
			log.Printf("  ✁ %s", liveOrgUnit.Name)

			if confirm {
				err := directorySrv.DeleteOrgUnit(ctx, liveOrgUnit)
				if err != nil {
					return fmt.Errorf("failed to delete org unit: %v", err)
				}
			}
		}
	}

	for _, expectedOrgUnit := range cfg.OrgUnits {
		if !liveNames.Has(expectedOrgUnit.Name) {
			log.Printf("  + %s", expectedOrgUnit.Name)

			if confirm {
				apiOrgUnit := config.ToGSuiteOrgUnit(&expectedOrgUnit)
				if err := directorySrv.CreateOrgUnit(ctx, apiOrgUnit); err != nil {
					return fmt.Errorf("failed to create org unit: %v", err)
				}
			}
		}
	}

	return nil
}
