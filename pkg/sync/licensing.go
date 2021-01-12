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
