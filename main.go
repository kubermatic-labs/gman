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
	"gopkg.in/yaml.v3"

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
	licensesConfigFile    string
	usersConfig           *config.Config
	groupsConfig          *config.Config
	orgUnitsConfig        *config.Config
	licenseStatus         *glib.LicenseStatus
	versionAction         bool
	confirm               bool
	validateAction        bool
	exportAction          bool
	licensesAction        bool
	licensesYAML          bool
	clientSecretFile      string
	impersonatedUserEmail string
	insecurePasswords     bool
	throttleRequests      time.Duration
	licenses              []config.License
}

func main() {
	var (
		opt options
		err error
	)

	flag.StringVar(&opt.usersConfigFile, "users-config", "", "path to the config.yaml that contains all users (if not given, users are not synchronized)")
	flag.StringVar(&opt.groupsConfigFile, "groups-config", "", "path to the config.yaml that contains all groups (if not given, groups are not synchronized)")
	flag.StringVar(&opt.orgUnitsConfigFile, "orgunits-config", "", "path to the config.yaml that contains all organization units (required)")
	flag.StringVar(&opt.licensesConfigFile, "licenses-config", "", "(optional) instead of using the inbuilt license list, this is a config.yaml that contains the relevant licenses")
	flag.StringVar(&opt.clientSecretFile, "private-key", "", "path to the Service Account secret file (.json) coontaining Keys used for authorization")
	flag.StringVar(&opt.impersonatedUserEmail, "impersonated-email", "", "Admin email used to impersonate Service Account")
	flag.BoolVar(&opt.versionAction, "version", false, "show the GMan version and exit")
	flag.BoolVar(&opt.validateAction, "validate", false, "validate the given configuration and then exit")
	flag.BoolVar(&opt.exportAction, "export", false, "export the state and update the config files (-[user|groups|orgunits]-config flags)")
	flag.BoolVar(&opt.licensesAction, "licenses", false, "print the builtin licenses and then exit")
	flag.BoolVar(&opt.licensesYAML, "licenses-yaml", false, "print the builtin licenses as YAML (use together with -licenses)")
	flag.BoolVar(&opt.confirm, "confirm", false, "must be set to actually perform any changes")
	flag.BoolVar(&opt.insecurePasswords, "insecure-passwords", false, "allow configuring static passwords for users")
	flag.DurationVar(&opt.throttleRequests, "throttle-requests", 500*time.Millisecond, "the delay between Enterprise Licensing API requests")
	flag.Parse()

	if opt.versionAction {
		fmt.Printf("GMan %s (built at %s)\n", version, date)
		return
	}

	if opt.licensesAction {
		licenseAction(opt.licensesYAML)
		return
	}

	// open the files
	if opt.usersConfigFile != "" {
		opt.usersConfig, err = config.LoadFromFile(opt.usersConfigFile)
		if err != nil {
			log.Fatalf("⚠ Failed to load user config from %q: %v.", opt.usersConfigFile, err)
		}
	}

	if opt.groupsConfigFile != "" {
		opt.groupsConfig, err = config.LoadFromFile(opt.groupsConfigFile)
		if err != nil {
			log.Fatalf("⚠ Failed to load group config from %q: %v.", opt.groupsConfigFile, err)
		}
	}

	opt.orgUnitsConfig, err = config.LoadFromFile(opt.orgUnitsConfigFile)
	if err != nil {
		log.Fatalf("⚠ Failed to load org unit config from %q: %v.", opt.orgUnitsConfigFile, err)
	}

	// load licenses
	opt.licenses = config.AllLicenses
	if opt.licensesConfigFile != "" {
		licensesConfig, err := config.LoadFromFile(opt.licensesConfigFile)
		if err != nil {
			log.Fatalf("⚠ Failed to load license config from %q: %v.", opt.licensesConfigFile, err)
		}

		opt.licenses = licensesConfig.Licenses
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
	log.Printf("☁ Working with organization %q…", orgName)

	if !opt.exportAction && !opt.confirm {
		log.Println("☞ This is a dry-run, no actual changes are being made.")
	}

	// create glib services
	ctx := context.Background()
	readonly := opt.exportAction || !opt.confirm
	scopes := getScopes(readonly)

	directorySrv, err := glib.NewDirectoryService(ctx, orgName, opt.clientSecretFile, opt.impersonatedUserEmail, opt.throttleRequests, scopes...)
	if err != nil {
		log.Fatalf("⚠ Failed to create GSuite Directory API client: %v", err)
	}

	licensingSrv, err := glib.NewLicensingService(ctx, orgName, opt.clientSecretFile, opt.impersonatedUserEmail, opt.throttleRequests, opt.licenses)
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

func licenseAction(asYAML bool) {
	if asYAML {
		output := struct {
			Licenses []config.License `yaml:"licenses"`
		}{
			Licenses: config.AllLicenses,
		}

		encoder := yaml.NewEncoder(os.Stdout)
		encoder.SetIndent(2)

		encoder.Encode(output)
	} else {
		for _, license := range config.AllLicenses {
			fmt.Printf("- %s (productID %q, SKU %q)\n", license.Name, license.ProductId, license.SkuId)
		}
	}
}

func syncAction(
	ctx context.Context,
	opt *options,
	directorySrv *glib.DirectoryService,
	licensingSrv *glib.LicensingService,
	groupsSettingsSrv *glib.GroupsSettingsService,
) {
	orgUnitChanges, err := sync.SyncOrgUnits(ctx, directorySrv, opt.orgUnitsConfig, opt.confirm)
	if err != nil {
		log.Fatalf("⚠ Failed to sync: %v.", err)
	}

	err = sync.SyncSchema(ctx, directorySrv, opt.confirm)
	if err != nil {
		log.Fatalf("⚠ Failed to sync: %v.", err)
	}

	userChanges := false
	if opt.usersConfig != nil {
		userChanges, err = sync.SyncUsers(ctx, directorySrv, licensingSrv, opt.usersConfig, opt.licenseStatus, opt.insecurePasswords, opt.confirm)
		if err != nil {
			log.Fatalf("⚠ Failed to sync: %v.", err)
		}
	} else {
		log.Println("⚠ No user configuration provided, not synchronizing users.")
	}

	groupChanges := false
	if opt.usersConfig != nil {
		groupChanges, err = sync.SyncGroups(ctx, directorySrv, groupsSettingsSrv, opt.groupsConfig, opt.confirm)
		if err != nil {
			log.Fatalf("⚠ Failed to sync: %v.", err)
		}
	} else {
		log.Println("⚠ No group configuration provided, not synchronizing groups.")
	}

	if opt.confirm {
		log.Println("✓ Organization successfully synchronized.")
	} else if orgUnitChanges || userChanges || groupChanges {
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

	users := []config.User{}
	if opt.usersConfigFile != "" {
		log.Println("► Exporting users…")
		users, err = export.ExportUsers(ctx, directorySrv, licensingSrv, opt.licenseStatus)
		if err != nil {
			log.Fatalf("⚠ Failed to export: %v.", err)
		}
	}

	groups := []config.Group{}
	if opt.groupsConfigFile != "" {
		log.Println("► Exporting groups…")
		groups, err = export.ExportGroups(ctx, directorySrv, groupsSettingsSrv)
		if err != nil {
			log.Fatalf("⚠ Failed to export: %v.", err)
		}
	}

	log.Println("► Updating config files…")

	// read&write the files individually, so that if the user specifies the same
	// file for all three configurations, the file gets incrementally updated

	if err := saveExport(opt.orgUnitsConfigFile, func(cfg *config.Config) { cfg.OrgUnits = orgUnits }); err != nil {
		log.Fatalf("⚠ Failed to update org unit config file: %v.", err)
	}

	if opt.usersConfigFile != "" {
		if err := saveExport(opt.usersConfigFile, func(cfg *config.Config) { cfg.Users = users }); err != nil {
			log.Fatalf("⚠ Failed to update user config file: %v.", err)
		}
	}

	if opt.groupsConfigFile != "" {
		if err := saveExport(opt.groupsConfigFile, func(cfg *config.Config) { cfg.Groups = groups }); err != nil {
			log.Fatalf("⚠ Failed to update group config file: %v.", err)
		}
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

	if errs := config.ValidateLicenses(opt.licenses); errs != nil {
		log.Println("⚠ License configuration is invalid:")
		for _, e := range errs {
			log.Printf("  - %v", e)
		}
		valid = false
	}

	if errs := opt.orgUnitsConfig.ValidateOrgUnits(); errs != nil {
		log.Println("⚠ Org unit configuration is invalid:")
		for _, e := range errs {
			log.Printf("  - %v", e)
		}
		valid = false
	}

	if opt.usersConfig != nil {
		if errs := opt.usersConfig.ValidateUsers(); errs != nil {
			log.Println("⚠ User configuration is invalid:")
			for _, e := range errs {
				log.Printf("  - %v", e)
			}
			valid = false
		}
	}

	if opt.groupsConfig != nil {
		if errs := opt.groupsConfig.ValidateGroups(); errs != nil {
			log.Println("⚠ Group configuration is invalid:")
			for _, e := range errs {
				log.Printf("  - %v", e)
			}
			valid = false
		}
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
			directoryv1.AdminDirectoryUserschemaReadonlyScope,
		}
	}

	return []string{
		directoryv1.AdminDirectoryUserScope,
		directoryv1.AdminDirectoryGroupScope,
		directoryv1.AdminDirectoryOrgunitScope,
		directoryv1.AdminDirectoryGroupMemberScope,
		directoryv1.AdminDirectoryResourceCalendarScope,
		directoryv1.AdminDirectoryUserschemaScope,
	}
}
