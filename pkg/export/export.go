package export

import (
	"context"
	"fmt"
	"log"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/glib"
	admin "google.golang.org/api/admin/directory/v1"
)

func ExportConfiguration(ctx context.Context, organization string, clientService *admin.Service) (*config.Config, error) {
	cfg := &config.Config{
		Organization: organization,
	}

	if err := exportUsers(ctx, clientService, cfg); err != nil {
		return cfg, fmt.Errorf("failed to export users: %v", err)
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
			primaryEmail, secondaryEmail := glib.GetUserEmails(u)

			cfg.Users = append(cfg.Users, config.UserConfig{
				FirstName:      u.Name.GivenName,
				LastName:       u.Name.FamilyName,
				PrimaryEmail:   primaryEmail,
				SecondaryEmail: secondaryEmail,
			})
		}
	}

	return nil
}
