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

//TODO: like everything, copy-pasta alert (from aquayman)

// These variables are set by goreleaser during build time.
var (
	version = "dev"
	date    = "unknown"
)

func main() {
	ctx := context.Background()

	var (
		configFile  = ""
		showVersion = false
		//confirm            = false
		//validate           = false
		exportMode  = false
		createUsers = false
		deleteUsers = false
	)

	flag.StringVar(&configFile, "config", configFile, "path to the config.yaml")
	flag.BoolVar(&showVersion, "version", showVersion, "show the Gman version and exit")
	//flag.BoolVar(&confirm, "confirm", confirm, "must be set to actually perform any changes")
	//flag.BoolVar(&validate, "validate", validate, "validate the given configuration and then exit")
	flag.BoolVar(&exportMode, "export", exportMode, "export the state and update the config file (-config flag)")
	flag.BoolVar(&createUsers, "create-users", createUsers, "create repositories listed in the config file but not existing on quay.io yet")
	flag.BoolVar(&deleteUsers, "delete-users", deleteUsers, "delete repositories on quay.io that are not listed in the config file")
	flag.Parse()

	if showVersion {
		fmt.Printf("Gman %s (built at %s)\n", version, date)
		return
	}

	if configFile == "" {
		log.Print("⚠ No configuration (-config) specified.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	cfg, err := config.LoadFromFile(configFile)
	if err != nil {
		log.Fatalf("⚠ Failed to load config %q: %v.", configFile, err)
	}

	srv, err := glib.NewDirectoryService()
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

	//	options := sync.Options{
	//CreateMissingRepositories:  createRepositories,
	//DeleteDanglingRepositories: deleteRepositories,
	//	}

	err = sync.SyncUsers(ctx, srv, cfg)
	if err != nil {
		log.Fatalf("⚠ Failed to sync state: %v.", err)
	}
	//	// TODO: users not repos
	//	log.Printf("► Updating organization %s…", cfg.Organization)
	//
	//	options := sync.Options{
	//		CreateMissingRepositories:  createRepositories,
	//		DeleteDanglingRepositories: deleteRepositories,
	//	}
	//
	//	err = sync.Sync(ctx, cfg, client, options)
	//	if err != nil {
	//		log.Fatalf("⚠ Failed to sync state: %v.", err)
	//	}
	//
	//	if confirm {
	//		log.Println("✓ Permissions successfully synchronized.")
	//	} else {
	//		log.Println("⚠ Run again with -confirm to apply the changes above.")
	//}
}
