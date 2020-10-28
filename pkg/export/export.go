package export

import (
	"context"
	"fmt"
	"log"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/glib"
	admin "google.golang.org/api/admin/directory/v1"
	groupssettings "google.golang.org/api/groupssettings/v1"
)

func ExportConfiguration(ctx context.Context, organization string, clientService *admin.Service, groupService *groupssettings.Service, licensingService *glib.LicensingService) (*config.Config, error) {
	cfg := &config.Config{
		Organization: organization,
	}

	if err := exportOrgUnits(ctx, clientService, cfg); err != nil {
		return cfg, fmt.Errorf("org units: %v", err)
	}

	if err := exportUsers(ctx, clientService, licensingService, cfg); err != nil {
		return cfg, fmt.Errorf("users: %v", err)
	}

	if err := exportGroups(ctx, clientService, groupService, cfg); err != nil {
		return cfg, fmt.Errorf("groups: %v", err)
	}

	return cfg, nil
}

func exportUsers(ctx context.Context, clientService *admin.Service, licensingService *glib.LicensingService, cfg *config.Config) error {
	log.Println("⇄ Exporting users from GSuite...")
	// get the users array
	users, err := glib.GetListOfUsers(*clientService)
	if err != nil {
		return err
	}

	// save to file
	if len(users) == 0 {
		log.Println("⚠ No users found.")
	} else {
		for _, u := range users {
			// get user licenses
			userLicenses, err := glib.GetUserLicenses(licensingService, u.PrimaryEmail)
			if err != nil {
				return err
			}
			usr := glib.CreateConfigUserFromGSuite(u, userLicenses)
			cfg.Users = append(cfg.Users, usr)

		}
	}

	return nil
}

func exportGroups(ctx context.Context, clientService *admin.Service, groupService *groupssettings.Service, cfg *config.Config) error {
	log.Println("⇄ Exporting groups from GSuite...")
	// get the groups array
	groups, err := glib.GetListOfGroups(clientService)
	if err != nil {
		return err
	}
	var members []*admin.Member

	// save to file
	if len(groups) == 0 {
		log.Println("⚠ No groups found.")
	} else {
		for _, g := range groups {
			members, err = glib.GetListOfMembers(clientService, g)
			if err != nil {
				return err
			}
			gSettings, err := glib.GetSettingOfGroup(groupService, g.Email)
			if err != nil {
				return err
			}
			thisGroup, err := glib.CreateConfigGroupFromGSuite(g, members, gSettings)
			if err != nil {
				return err
			}

			cfg.Groups = append(cfg.Groups, thisGroup)
		}
	}

	return nil
}

func exportOrgUnits(ctx context.Context, clientService *admin.Service, cfg *config.Config) error {
	log.Println("⇄ Exporting organizational units from GSuite...")
	// get the users array
	orgUnits, err := glib.GetListOfOrgUnits(clientService)
	if err != nil {
		return err
	}

	// save to file
	if len(orgUnits) == 0 {
		log.Println("⚠ No OrgUnits found.")
	} else {
		for _, ou := range orgUnits {
			cfg.OrgUnits = append(cfg.OrgUnits, config.OrgUnitConfig{
				Name:              ou.Name,
				Description:       ou.Description,
				ParentOrgUnitPath: ou.ParentOrgUnitPath,
				BlockInheritance:  ou.BlockInheritance,
				OrgUnitPath:       ou.OrgUnitPath,
			})

		}
	}

	return nil
}
