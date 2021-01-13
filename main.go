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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	directoryv1 "google.golang.org/api/admin/directory/v1"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/export"
	"github.com/kubermatic-labs/gman/pkg/glib"
	"github.com/kubermatic-labs/gman/pkg/sync"
)

// These variables are set by goreleaser during build time.
var (
	version = "dev"
	date    = "unknown"
)

type options struct {
	usersConfigFile       string
	groupsConfigFile      string
	orgUnitsConfigFile    string
	usersConfig           *config.Config
	groupsConfig          *config.Config
	orgUnitsConfig        *config.Config
	licenseStatus         *glib.LicenseStatus
	versionAction         bool
	confirm               bool
	validateAction        bool
	exportAction          bool
	clientSecretFile      string
	impersonatedUserEmail string
	throttleRequests      time.Duration
}

func main() {
	var (
		opt options
		err error
	)

	flag.StringVar(&opt.usersConfigFile, "users-config", "", "path to the config.yaml that contains all users")
	flag.StringVar(&opt.groupsConfigFile, "groups-config", "", "path to the config.yaml that contains all groups")
	flag.StringVar(&opt.orgUnitsConfigFile, "orgunits-config", "", "path to the config.yaml that contains all organization units")
	flag.StringVar(&opt.clientSecretFile, "private-key", "", "path to the Service Account secret file (.json) coontaining Keys used for authorization")
	flag.StringVar(&opt.impersonatedUserEmail, "impersonated-email", "", "Admin email used to impersonate Service Account")
	flag.BoolVar(&opt.versionAction, "version", false, "show the GMan version and exit")
	flag.BoolVar(&opt.validateAction, "validate", false, "validate the given configuration and then exit")
	flag.BoolVar(&opt.exportAction, "export", false, "export the state and update the config files (-[user|groups|orgunits]-config flags)")
	flag.BoolVar(&opt.confirm, "confirm", false, "must be set to actually perform any changes")
	flag.DurationVar(&opt.throttleRequests, "throttle-requests", 500*time.Millisecond, "the delay between Enterprise Licensing API requests")
	flag.Parse()

	if opt.versionAction {
		fmt.Printf("GMan %s (built at %s)\n", version, date)
		return
	}

	// open the files
	opt.usersConfig, err = config.LoadFromFile(opt.usersConfigFile)
	if err != nil {
		log.Fatalf("⚠ Failed to load user config from %q: %v.", opt.usersConfigFile, err)
	}

	opt.groupsConfig, err = config.LoadFromFile(opt.groupsConfigFile)
	if err != nil {
		log.Fatalf("⚠ Failed to load group config from %q: %v.", opt.groupsConfigFile, err)
	}

	opt.orgUnitsConfig, err = config.LoadFromFile(opt.orgUnitsConfigFile)
	if err != nil {
		log.Fatalf("⚠ Failed to load org unit config from %q: %v.", opt.orgUnitsConfigFile, err)
	}

	// validate config unless in export mode, where an incomplete configuration is expected
	if !opt.exportAction {
		valid := validateAction(&opt)
		if !valid {
			os.Exit(1)
		}

		log.Println("✓ Configuration is valid.")
		if opt.validateAction {
			return
		}
	}

	orgName := opt.groupsConfig.Organization
	log.Printf("Working with organization %q…", orgName)

	// create glib services
	ctx := context.Background()
	readonly := opt.exportAction || !opt.confirm
	scopes := getScopes(readonly)

	directorySrv, err := glib.NewDirectoryService(ctx, orgName, opt.clientSecretFile, opt.impersonatedUserEmail, opt.throttleRequests, scopes...)
	if err != nil {
		log.Fatalf("⚠ Failed to create GSuite Directory API client: %v", err)
	}

	licensingSrv, err := glib.NewLicensingService(ctx, orgName, opt.clientSecretFile, opt.impersonatedUserEmail, opt.throttleRequests, config.AllLicenses)
	if err != nil {
		log.Fatalf("⚠ Failed to create GSuite Licensing API client: %v", err)
	}

	groupsSettingsSrv, err := glib.NewGroupsSettingsService(ctx, opt.clientSecretFile, opt.impersonatedUserEmail, opt.throttleRequests)
	if err != nil {
		log.Fatalf("⚠ Failed to create GSuite GroupsSettings API client: %v", err)
	}

	// begin actual work
	log.Println("► Fetching license status…")
	opt.licenseStatus, err = licensingSrv.GetLicenseStatus(ctx)
	if err != nil {
		log.Fatalf("⚠ Failed to fetch: %v.", err)
	}

	if opt.exportAction {
		exportAction(ctx, &opt, directorySrv, licensingSrv, groupsSettingsSrv)
	} else {
		syncAction(ctx, &opt, directorySrv, licensingSrv, groupsSettingsSrv)
	}
}

func syncAction(
	ctx context.Context,
	opt *options,
	directorySrv *glib.DirectoryService,
	licensingSrv *glib.LicensingService,
	groupsSettingsSrv *glib.GroupsSettingsService,
) {
	// log.Println("► Updating org units…")
	// if err := sync.SyncOrgUnits(ctx, directorySrv, opt.orgUnitsConfig, opt.confirm); err != nil {
	// 	log.Fatalf("⚠ Failed to sync: %v.", err)
	// }

	log.Println("► Updating users…")
	userChanges, err := sync.SyncUsers(ctx, directorySrv, licensingSrv, opt.usersConfig, opt.licenseStatus, opt.confirm)
	if err != nil {
		log.Fatalf("⚠ Failed to sync: %v.", err)
	}

	// log.Println("► Updating groups…")
	// groupChanges, err := sync.SyncGroups(ctx, directorySrv, groupsSettingsSrv, opt.groupsConfig, opt.confirm)
	// if err != nil {
	// 	log.Fatalf("⚠ Failed to sync: %v.", err)
	// }

	if opt.confirm {
		log.Println("✓ Organization successfully synchronized.")
	} else if userChanges /* || groupChanges */ {
		log.Println("⚠ Run again with -confirm to apply the changes above.")
	} else {
		log.Println("✓ No changes necessary, organization is in sync.")
	}
}

func exportAction(
	ctx context.Context,
	opt *options,
	directorySrv *glib.DirectoryService,
	licensingSrv *glib.LicensingService,
	groupsSettingsSrv *glib.GroupsSettingsService,
) {
	log.Println("► Exporting organizational units…")
	orgUnits, err := export.ExportOrgUnits(ctx, directorySrv)
	if err != nil {
		log.Fatalf("⚠ Failed to export: %v.", err)
	}

	log.Println("► Exporting users…")
	users, err := export.ExportUsers(ctx, directorySrv, licensingSrv, opt.licenseStatus)
	if err != nil {
		log.Fatalf("⚠ Failed to export: %v.", err)
	}

	log.Println("► Exporting groups…")
	groups, err := export.ExportGroups(ctx, directorySrv, groupsSettingsSrv)
	if err != nil {
		log.Fatalf("⚠ Failed to export: %v.", err)
	}

	log.Println("► Updating config files…")

	// read&write the files individually, so that if the user specifies the same
	// file for all three configurations, the file gets incrementally updated

	if err := saveExport(opt.orgUnitsConfigFile, func(cfg *config.Config) { cfg.OrgUnits = orgUnits }); err != nil {
		log.Fatalf("⚠ Failed to update org unit config file: %v.", err)
	}

	if err := saveExport(opt.usersConfigFile, func(cfg *config.Config) { cfg.Users = users }); err != nil {
		log.Fatalf("⚠ Failed to update user config file: %v.", err)
	}

	if err := saveExport(opt.groupsConfigFile, func(cfg *config.Config) { cfg.Groups = groups }); err != nil {
		log.Fatalf("⚠ Failed to update group config file: %v.", err)
	}

	log.Println("✓ Export successful.")
}

func saveExport(filename string, patch func(*config.Config)) error {
	cfg, err := config.LoadFromFile(filename)
	if err != nil {
		return err
	}

	patch(cfg)

	return config.SaveToFile(cfg, filename)
}

func validateAction(opt *options) bool {
	valid := true

	if errs := opt.orgUnitsConfig.ValidateOrgUnits(); errs != nil {
		log.Println("⚠ Org unit configuration is invalid:")
		for _, e := range errs {
			log.Printf("  - %v", e)
		}
		valid = false
	}

	if errs := opt.usersConfig.ValidateUsers(); errs != nil {
		log.Println("⚠ User configuration is invalid:")
		for _, e := range errs {
			log.Printf("  - %v", e)
		}
		valid = false
	}

	if errs := opt.groupsConfig.ValidateGroups(); errs != nil {
		log.Println("⚠ Group configuration is invalid:")
		for _, e := range errs {
			log.Printf("  - %v", e)
		}
		valid = false
	}

	return valid
}

func getScopes(readonly bool) []string {
	if readonly {
		return []string{
			directoryv1.AdminDirectoryUserReadonlyScope,
			directoryv1.AdminDirectoryGroupReadonlyScope,
			directoryv1.AdminDirectoryOrgunitReadonlyScope,
			directoryv1.AdminDirectoryGroupMemberReadonlyScope,
			directoryv1.AdminDirectoryResourceCalendarReadonlyScope,
		}
	}

	return []string{
		directoryv1.AdminDirectoryUserScope,
		directoryv1.AdminDirectoryGroupScope,
		directoryv1.AdminDirectoryOrgunitScope,
		directoryv1.AdminDirectoryGroupMemberScope,
		directoryv1.AdminDirectoryResourceCalendarScope,
	}
}
