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
		showVersion           = false
		confirm               = false
		validate              = false
		exportMode            = false
		clientSecretFile      = ""
		impersonatedUserEmail = ""
		throttleRequests      = 500 * time.Millisecond
	)

	flag.StringVar(&configFile, "config", configFile, "path to the config.yaml")
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

	// config file must be present
	if configFile == "" {
		configFile = os.Getenv("GMAN_CONFIG_FILE")
		if configFile == "" {
			log.Print("⚠ No configuration (-config) specified.\n\n")
			flag.Usage()
			os.Exit(1)
		}
	}

	cfg, err := config.LoadFromFile(configFile)
	if err != nil {
		log.Fatalf("⚠ Failed to load config %q: %v.", configFile, err)
	}

	// validate config unless in export mode, where an incomplete configuration is allowed and even expected
	if !exportMode {
		if errs := cfg.Validate(); errs != nil {
			log.Println("Configuration is invalid:")
			for _, e := range errs {
				log.Printf(" ⚠  %v\n", e)
			}
			os.Exit(1)
		} else {
			log.Println("✓ Configuration is valid.")
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

	grSrv, err := glib.NewGroupsService(clientSecretFile, impersonatedUserEmail)
	if err != nil {
		log.Fatalf("⚠ Failed to create GSuite Groupssettings API client: %v", err)
	}

	licnsSrv, err := glib.NewLicensingService(clientSecretFile, impersonatedUserEmail)
	if err != nil {
		log.Fatalf("⚠ Failed to create GSuite Licensing API client: %v", err)
	}

	licSrv := glib.LicensingService{
		Service:          licnsSrv,
		ThrottleRequests: throttleRequests,
	}

	if exportMode {
		log.Printf("► Exporting organization %s…", cfg.Organization)

		newConfig, err := export.ExportConfiguration(ctx, cfg.Organization, srv, grSrv, licSrv)
		if err != nil {
			log.Fatalf("⚠ Failed to export %v.", err)
		}
		if err := config.SaveToFile(newConfig, configFile); err != nil {
			log.Fatalf("⚠ Failed to update config file: %v.", err)
		}

		log.Println("✓ Export successful.")
		return
	}

	log.Printf("► Updating organization %s…", cfg.Organization)

	err = sync.SyncConfiguration(ctx, cfg, srv, grSrv, licSrv, confirm)
	if err != nil {
		log.Fatalf("⚠ Failed to sync state: %v.", err)
	}

	if confirm {
		log.Println("✓ Organization successfully synchronized.")
	} else {
		log.Println("⚠ Run again with -confirm to apply the changes above.")
	}
}
