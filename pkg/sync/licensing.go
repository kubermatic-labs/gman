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

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/glib"
)

func userHasLicense(u *config.User, l config.License) bool {
	for _, assigned := range u.Licenses {
		if assigned == l.SkuId {
			return true
		}
	}

	return false
}

func sliceContainsLicense(licenses []config.License, identifier string) bool {
	for _, license := range licenses {
		if license.SkuId == identifier {
			return true
		}
	}

	return false
}

// syncUserLicenses provides logic for creating/deleting/updating licenses according to config file
func syncUserLicenses(
	ctx context.Context,
	licenseSrv *glib.LicensingService,
	expectedUser *config.User,
	liveUser *directoryv1.User,
	licenseStatus *glib.LicenseStatus,
	confirm bool,
) error {
	expectedLicenses := expectedUser.Licenses
	liveLicenses := []config.License{}

	// in dry-run mode, there can be cases where there is no live user yet
	if liveUser != nil {
		liveLicenses = licenseStatus.GetLicensesForUser(liveUser)
	}

	for _, liveLicense := range liveLicenses {
		if !userHasLicense(expectedUser, liveLicense) {
			log.Printf("    - license %s", liveLicense.Name)

			if confirm {
				if err := licenseSrv.UnassignLicense(ctx, liveUser, liveLicense); err != nil {
					return fmt.Errorf("unable to assign license: %v", err)
				}
			}
		}
	}

	for _, expectedLicense := range expectedLicenses {
		if !sliceContainsLicense(liveLicenses, expectedLicense) {
			license := licenseStatus.GetLicense(expectedLicense)
			log.Printf("    + license %s", license.Name)

			if confirm {
				if err := licenseSrv.AssignLicense(ctx, liveUser, *license); err != nil {
					return fmt.Errorf("unable to assign license: %v", err)
				}
			}
		}
	}

	return nil
}
