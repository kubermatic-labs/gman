// Package glib contains methods for interactions with GSuite API
package glib

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"golang.org/x/oauth2/google"
	directoryv1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/licensing/v1"
	"google.golang.org/api/option"

	"github.com/kubermatic-labs/gman/pkg/config"
)

type LicensingService struct {
	*licensing.Service

	licenses []config.License
	delay    time.Duration
}

// NewLicensingService() creates a client for communicating with Google Licensing API,
// returns a service object authorized to perform actions in Gsuite.
func NewLicensingService(ctx context.Context, clientSecretFile string, impersonatedUserEmail string, delay time.Duration, licenses []config.License) (*LicensingService, error) {
	jsonCredentials, err := ioutil.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read json credentials: %v", err)
	}

	config, err := google.JWTConfigFromJSON(jsonCredentials, licensing.AppsLicensingScope)
	if err != nil {
		return nil, fmt.Errorf("unable to process credentials: %v", err)
	}
	config.Subject = impersonatedUserEmail

	ts := config.TokenSource(ctx)

	srv, err := licensing.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("unable to create a new licensing service: %v", err)
	}

	licenseService := &LicensingService{
		Service:  srv,
		licenses: licenses,
		delay:    delay,
	}

	return licenseService, nil
}

func (ls *LicensingService) GetLicenses() ([]config.License, error) {
	// in the future, this might be available via the API itself
	return ls.licenses, nil
}

// GetUserLicense returns a list of licenses of a user;
// note that this is extremely slow due to API limitations, consider
// listing all usages per license instead.
func (ls *LicensingService) GetUserLicenses(ctx context.Context, user string) ([]config.License, error) {
	var result []config.License

	for _, license := range ls.licenses {
		_, err := ls.LicenseAssignments.Get(license.ProductId, license.SkuId, user).Context(ctx).Do()

		log.Println("TODO: delay next request")

		if err != nil {
			return nil, fmt.Errorf("unable to retrieve license in domain: %v", err)
		}

		result = append(result, license)
	}

	return result, nil
}

// LicenseUsages lists all user IDs assigned licenses for a specific
// product SKU.
func (ls *LicensingService) LicenseUsages(ctx context.Context, license config.License) ([]string, error) {
	userIDs := []string{}
	token := ""

	for {
		request := ls.LicenseAssignments.ListForProduct(license.ProductId, "my_customer").PageToken(token).Context(ctx)

		response, err := request.Do()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve list of users: %v", err)
		}

		for _, assignment := range response.Items {
			userIDs = append(userIDs, assignment.UserId)
		}

		token = response.NextPageToken
		if token == "" {
			break
		}
	}

	return userIDs, nil
}

func (ls *LicensingService) AssignLicense(ctx context.Context, user *directoryv1.User, license config.License) error {
	op := licensing.LicenseAssignmentInsert{UserId: user.PrimaryEmail}

	if _, err := ls.LicenseAssignments.Insert(license.ProductId, license.SkuId, &op).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to assign license: %v", err)
	}

	return nil
}

func (ls *LicensingService) UnassignLicense(ctx context.Context, user *directoryv1.User, license config.License) error {
	if _, err := ls.LicenseAssignments.Delete(license.ProductId, license.SkuId, user.PrimaryEmail).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to unassign license: %v", err)
	}

	return nil
}

type LicenseStatus struct {
	Assignments map[string][]string
	Licenses    map[string]config.License
}

func (ls *LicensingService) GetLicenseStatus(ctx context.Context) (*LicenseStatus, error) {
	licenses, err := ls.GetLicenses()
	if err != nil {
		return nil, fmt.Errorf("failed to determine list of all available licenses: %v", err)
	}

	status := &LicenseStatus{
		Assignments: make(map[string][]string),
		Licenses:    make(map[string]config.License),
	}

	for _, license := range licenses {
		log.Printf("  %s", license.Name)

		assignments, err := ls.LicenseUsages(ctx, license)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch license usages: %v", err)
		}

		status.Assignments[license.SkuId] = assignments
		status.Licenses[license.SkuId] = license
	}

	return status, nil
}

func (ls *LicenseStatus) GetLicensesForUser(user *directoryv1.User) []config.License {
	result := []config.License{}

	for _, skuId := range ls.Assignments[user.Id] {
		result = append(result, ls.Licenses[skuId])
	}

	return result
}

func (ls *LicenseStatus) GetLicense(identifier string) *config.License {
	license, ok := ls.Licenses[identifier]
	if !ok {
		return nil
	}

	return &license
}
