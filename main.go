package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/export"
	"github.com/kubermatic-labs/gman/pkg/sync"

	"github.com/kubermatic-labs/gman/pkg/glib"
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
		createUsers           = false
		deleteUsers           = false
		clientSecretFile      = ""
		impersonatedUserEmail = ""
	)

	flag.StringVar(&configFile, "config", configFile, "path to the config.yaml")
	flag.StringVar(&clientSecretFile, "private-key", clientSecretFile, "path to the Service Account secret file (.json) coontaining Keys used for authorization")
	flag.StringVar(&impersonatedUserEmail, "impersonated-email", impersonatedUserEmail, "Admin email used to impersonate Service Account")
	flag.BoolVar(&showVersion, "version", showVersion, "show the Gman version and exit")
	flag.BoolVar(&confirm, "confirm", confirm, "must be set to actually perform any changes")
	flag.BoolVar(&validate, "validate", validate, "validate the given configuration and then exit")
	flag.BoolVar(&exportMode, "export", exportMode, "export the state and update the config file (-config flag)")
	flag.BoolVar(&createUsers, "create-users", createUsers, "create repositories listed in the config file but not existing on quay.io yet")
	flag.BoolVar(&deleteUsers, "delete-users", deleteUsers, "delete repositories on quay.io that are not listed in the config file")

	flag.Parse()

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

	if showVersion {
		fmt.Printf("Gman %s (built at %s)\n", version, date)
		return
	}

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
		if err := cfg.Validate(); err != nil {
			log.Fatalf("⚠ Configuration is invalid: %v", err)
		} else {
			log.Println("✓ Configuration is valid.")
		}
		if validate {
			return
		}
	}

	srv, err := glib.NewDirectoryService(clientSecretFile, impersonatedUserEmail)
	if err != nil {
		log.Fatalf("⚠ Failed to create GSuite API client (CreateDirectoryService): %v", err)
	}

	if exportMode {
		log.Printf("► Exporting organization %s…", cfg.Organization)

		newConfig, err := export.ExportConfiguration(ctx, cfg.Organization, srv)
		if err != nil {
			log.Fatalf("⚠ Failed to export: %v.", err)
		}

		if err := config.SaveToFile(newConfig, configFile); err != nil {
			log.Fatalf("⚠ Failed to update config file: %v.", err)
		}

		log.Println("✓ Export successful.")
		return
	}

	log.Printf("► Updating organization %s…", cfg.Organization)

	err = sync.SyncConfiguration(ctx, cfg, srv, confirm)
	if err != nil {
		log.Fatalf("⚠ Failed to sync state: %v.", err)
	}

	if confirm {
		log.Println("✓ Organization successfully synchronized.")
	} else {
		log.Println("⚠ Run again with -confirm to apply the changes above.")
	}
}
