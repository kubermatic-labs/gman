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

func ExportConfiguration(ctx context.Context, organization string, clientService *admin.Service, groupService *groupssettings.Service) (*config.Config, error) {
	cfg := &config.Config{
		Organization: organization,
	}

	if err := exportOrgUnits(ctx, clientService, cfg); err != nil {
		return cfg, fmt.Errorf("failed to export org units: %v", err)
	}

	if err := exportUsers(ctx, clientService, cfg); err != nil {
		return cfg, fmt.Errorf("failed to export users: %v", err)
	}

	if err := exportGroups(ctx, clientService, groupService, cfg); err != nil {
		return cfg, fmt.Errorf("failed to export groups: %v", err)
	}

	return cfg, nil
}

func exportUsers(ctx context.Context, clientService *admin.Service, cfg *config.Config) error {
	log.Println("⇄ Exporting users from GSuite...")
	// get the users array
	users, _ := glib.GetListOfUsers(*clientService)

	// save to file
	if len(users) == 0 {
		log.Println("⚠ No users found.")
	} else {
		for _, u := range users {
			// get emails
			//primaryEmail, secondaryEmail := glib.GetUserEmails(u)
			usr := glib.CreateConfigUserFromGSuite(u)
			cfg.Users = append(cfg.Users, usr)
		}
	}

	return nil
}

func exportGroups(ctx context.Context, clientService *admin.Service, groupService *groupssettings.Service, cfg *config.Config) error {
	log.Println("⇄ Exporting groups from GSuite...")
	// get the groups array
	groups, _ := glib.GetListOfGroups(clientService)
	var members []*admin.Member

	// save to file
	if len(groups) == 0 {
		log.Println("⚠ No groups found.")
	} else {
		for _, g := range groups {
			members, _ = glib.GetListOfMembers(clientService, g)
			gSettings, _ := glib.GetSettingOfGroup(groupService, g.Email)
			thisGroup := config.GroupConfig{
				Name:                 g.Name,
				Email:                g.Email,
				Description:          g.Description,
				WhoCanContactOwner:   gSettings.WhoCanContactOwner,
				WhoCanViewMembership: gSettings.WhoCanViewMembership,
				WhoCanApproveMembers: gSettings.WhoCanApproveMembers,
				WhoCanPostMessage:    gSettings.WhoCanPostMessage,
				WhoCanJoin:           gSettings.WhoCanJoin,
				AllowExternalMembers: gSettings.AllowExternalMembers,
				Members:              []config.MemberConfig{},
			}
			for _, m := range members {
				thisGroup.Members = append(thisGroup.Members, config.MemberConfig{
					Email: m.Email,
					Role:  m.Role,
				})

			}
			cfg.Groups = append(cfg.Groups, thisGroup)
		}
	}

	return nil
}

func exportOrgUnits(ctx context.Context, clientService *admin.Service, cfg *config.Config) error {
	log.Println("⇄ Exporting organizational units from GSuite...")
	// get the users array
	orgUnits, _ := glib.GetListOfOrgUnits(clientService)

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
