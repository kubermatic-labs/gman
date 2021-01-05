package export

import (
	"context"
	"log"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/glib"
	admin "google.golang.org/api/admin/directory/v1"
	groupssettings "google.golang.org/api/groupssettings/v1"
)

func ExportUsers(ctx context.Context, clientService *admin.Service, licensingService *glib.LicensingService, cfg *config.Config) error {
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
			log.Printf("  %s", u.PrimaryEmail)

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

func ExportGroups(ctx context.Context, clientService *admin.Service, groupService *groupssettings.Service, cfg *config.Config) error {
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
			log.Printf("  %s", g.Name)

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

func ExportOrgUnits(ctx context.Context, clientService *admin.Service, cfg *config.Config) error {
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
			log.Printf("  %s", ou.Name)

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
