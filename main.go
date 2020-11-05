package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/export"
	"github.com/kubermatic-labs/gman/pkg/glib"
	"github.com/kubermatic-labs/gman/pkg/sync"
	admin "google.golang.org/api/admin/directory/v1"
)

// These variables are set by goreleaser during build time.
var (
	version = "dev"
	date    = "unknown"
)

func main() {
	ctx := context.Background()

	var (
		configFile            = ""
		usersConfigFile       = ""
		groupsConfigFile      = ""
		orgunitsConfigFile    = ""
		showVersion           = false
		confirm               = false
		validate              = false
		exportMode            = false
		clientSecretFile      = ""
		impersonatedUserEmail = ""
		throttleRequests      = 500 * time.Millisecond

		splitConfig = false
		userCfg     *config.Config
		groupCfg    *config.Config
		orgunitsCfg *config.Config
		err         error
	)

	flag.StringVar(&configFile, "config", configFile, "path to the config.yaml that manages whole organization; cannot be used together with separated config files for users/groups/organizational units")
	flag.StringVar(&usersConfigFile, "users-config", usersConfigFile, "path to the config.yaml that manages only users in GSuite organization")
	flag.StringVar(&groupsConfigFile, "groups-config", groupsConfigFile, "path to the config.yaml that manages only groups in GSuite organization")
	flag.StringVar(&orgunitsConfigFile, "orgunits-config", orgunitsConfigFile, "path to the config.yaml that manages only organizational units in GSuite organization")
	flag.StringVar(&clientSecretFile, "private-key", clientSecretFile, "path to the Service Account secret file (.json) coontaining Keys used for authorization")
	flag.StringVar(&impersonatedUserEmail, "impersonated-email", impersonatedUserEmail, "Admin email used to impersonate Service Account")
	flag.BoolVar(&showVersion, "version", showVersion, "show the Gman version and exit; does not need config file, API key and impersonated email")
	flag.BoolVar(&confirm, "confirm", confirm, "must be set to actually perform any changes")
	flag.BoolVar(&validate, "validate", validate, "validate the given configuration and then exit; does not need API key and impersonated email")
	flag.BoolVar(&exportMode, "export", exportMode, "export the state and update the config file (-config flag)")
	flag.DurationVar(&throttleRequests, "throttle-requests", throttleRequests, "the delay between Enterprise Licensing API requests")

	flag.Parse()

	if showVersion {
		fmt.Printf("Gman %s (built at %s)\n", version, date)
		return
	}

	splitConfig = usersConfigFile != "" || groupsConfigFile != "" || orgunitsConfigFile != ""

	if splitConfig == true && configFile != "" {
		log.Print("⚠ General configuration file specified (-config); cannot manage resources in separated files as well (-users-config/-groups-config/-orgunits-config).\n\n")
		flag.Usage()
		os.Exit(1)
	} else if splitConfig == false && configFile == "" {
		// general config file must be present if no splitted ones are specified
		configFile = os.Getenv("GMAN_CONFIG_FILE")
		if configFile == "" {
			log.Print("⚠ No configuration file(s) specified.\n\n")
			flag.Usage()
			os.Exit(1)
		}
	} else if splitConfig == false && configFile != "" {
		// open one config file
		cfg, err := config.LoadFromFile(configFile)
		if err != nil {
			log.Fatalf("⚠ Failed to load config %q: %v.", configFile, err)
		}
		userCfg = cfg
		groupCfg = cfg
		orgunitsCfg = cfg
		usersConfigFile = configFile
		groupsConfigFile = configFile
		orgunitsConfigFile = configFile
	} else {
		// open the files
		if usersConfigFile != "" {
			userCfg, err = config.LoadFromFile(usersConfigFile)
			if err != nil {
				log.Fatalf("⚠ Failed to load config %q: %v.", usersConfigFile, err)
			}
		}
		if groupsConfigFile != "" {
			groupCfg, err = config.LoadFromFile(groupsConfigFile)
			if err != nil {
				log.Fatalf("⚠ Failed to load config %q: %v.", groupsConfigFile, err)
			}
		}
		if orgunitsConfigFile != "" {
			orgunitsCfg, err = config.LoadFromFile(orgunitsConfigFile)
			if err != nil {
				log.Fatalf("⚠ Failed to load config %q: %v.", orgunitsConfigFile, err)
			}
		}
	}

	// validate config unless in export mode, where an incomplete configuration is expected
	if !exportMode {
		valid := true
		if usersConfigFile != "" || configFile != "" {
			if errs := userCfg.ValidateUsers(); errs != nil {
				log.Println("Users configuration is invalid:")
				for _, e := range errs {
					log.Printf(" ⚠  %v\n", e)
				}
				valid = false
			}
		}
		if groupsConfigFile != "" || configFile != "" {
			if errs := groupCfg.ValidateGroups(); errs != nil {
				log.Println("Groups configuration is invalid:")
				for _, e := range errs {
					log.Printf(" ⚠  %v\n", e)
				}
				valid = false
			}
		}
		if orgunitsConfigFile != "" || configFile != "" {
			if errs := orgunitsCfg.ValidateOrgUnits(); errs != nil {
				log.Println("Organizational units configuration is invalid:")
				for _, e := range errs {
					log.Printf(" ⚠  %v\n", e)
				}
				valid = false
			}
		}
		if valid {
			log.Println("✓ Configuration is valid.")
		} else {
			os.Exit(1)
		}
		// return if in validate mode
		if validate {
			return
		}
	}

	if clientSecretFile == "" {
		clientSecretFile = os.Getenv("GMAN_SERVICE_ACCOUNT_KEY")
		if clientSecretFile == "" {
			log.Print("⚠ No authorization .json file (-private-key) specified.\n\n")
			flag.Usage()
			os.Exit(1)
		}
	}

	if impersonatedUserEmail == "" {
		impersonatedUserEmail = os.Getenv("GMAN_IMPERSONATED_EMAIL")
		if impersonatedUserEmail == "" {
			log.Print("⚠ No impersonated user email (-impersonated-email) specified.\n\n")
			flag.Usage()
			os.Exit(1)
		}
	}

	// open GSuite API Admin service
	var srv *admin.Service
	if exportMode || !confirm {
		srv, err = glib.NewDirectoryService(clientSecretFile, impersonatedUserEmail, admin.AdminDirectoryUserReadonlyScope, admin.AdminDirectoryGroupReadonlyScope, admin.AdminDirectoryOrgunitReadonlyScope, admin.AdminDirectoryGroupMemberReadonlyScope, admin.AdminDirectoryResourceCalendarReadonlyScope)
		if err != nil {
			log.Fatalf("⚠ Failed to create GSuite Directory API client: %v", err)
		}
	} else {
		srv, err = glib.NewDirectoryService(clientSecretFile, impersonatedUserEmail, admin.AdminDirectoryUserScope, admin.AdminDirectoryGroupScope, admin.AdminDirectoryGroupMemberScope, admin.AdminDirectoryOrgunitScope, admin.AdminDirectoryResourceCalendarScope)
		if err != nil {
			log.Fatalf("⚠ Failed to create GSuite Directory API client: %v", err)
		}
	}

	// handle export/sync for org units
	if orgunitsConfigFile != "" || configFile != "" {
		if exportMode {
			log.Printf("► Exporting organizational units in organization %s…", groupCfg.Organization)

			err = export.ExportOrgUnits(ctx, srv, orgunitsCfg)
			if err != nil {
				log.Fatalf("⚠ Failed to export organizational units: %v.", err)
			}
			if err := config.SaveToFile(orgunitsCfg, orgunitsConfigFile); err != nil {
				log.Fatalf("⚠ Failed to update config file: %v.", err)
			}

			log.Println("✓ Export successful.")
		} else {
			log.Printf("► Updating organizational units in organization %s…", userCfg.Organization)

			err = sync.SyncOrgUnits(ctx, srv, orgunitsCfg, confirm)
			if err != nil {
				log.Fatalf("⚠ Failed to sync state: %v.", err)
			}

			if confirm {
				log.Println("✓ Organizational units successfully synchronized.")
			} else {
				log.Println("⚠ Run again with -confirm to apply the changes above.")
			}
		}
	}

	// handle export/sync for users
	if usersConfigFile != "" || configFile != "" {
		licSrv, err := glib.NewLicensingService(clientSecretFile, impersonatedUserEmail, throttleRequests)
		if err != nil {
			log.Fatalf("⚠ Failed to create GSuite Licensing API client: %v", err)
		}

		if exportMode {
			log.Printf("► Exporting users in organization %s…", userCfg.Organization)

			err := export.ExportUsers(ctx, srv, licSrv, userCfg)
			if err != nil {
				log.Fatalf("⚠ Failed to export users: %v.", err)
			}

			if err := config.SaveToFile(userCfg, usersConfigFile); err != nil {
				log.Fatalf("⚠ Failed to update config file: %v.", err)
			}
			log.Println("✓ Export of users successful.")
		} else {
			log.Printf("► Updating users in organization %s…", userCfg.Organization)

			err = sync.SyncUsers(ctx, srv, licSrv, userCfg, confirm)
			if err != nil {
				log.Fatalf("⚠ Failed to sync state: %v.", err)
			}

			if confirm {
				log.Println("✓ Users successfully synchronized.")
			} else {
				log.Println("⚠ Run again with -confirm to apply the changes above.")
			}
		}
	}
	// handle export/sync for groups
	if groupsConfigFile != "" || configFile != "" {
		grSrv, err := glib.NewGroupsService(clientSecretFile, impersonatedUserEmail)
		if err != nil {
			log.Fatalf("⚠ Failed to create GSuite Groupssettings API client: %v", err)
		}

		if exportMode {
			log.Printf("► Exporting groups in organization %s…", groupCfg.Organization)

			err = export.ExportGroups(ctx, srv, grSrv, groupCfg)
			if err != nil {
				log.Fatalf("⚠ Failed to export groups: %v.", err)
			}
			if err := config.SaveToFile(groupCfg, groupsConfigFile); err != nil {
				log.Fatalf("⚠ Failed to update config file: %v.", err)
			}

			log.Println("✓ Export successful.")
		} else {
			log.Printf("► Updating groups in organization %s…", userCfg.Organization)

			err = sync.SyncGroups(ctx, srv, grSrv, groupCfg, confirm)
			if err != nil {
				log.Fatalf("⚠ Failed to sync state: %v.", err)
			}

			if confirm {
				log.Println("✓ Groups successfully synchronized.")
			} else {
				log.Println("⚠ Run again with -confirm to apply the changes above.")
			}
		}
	}
}
