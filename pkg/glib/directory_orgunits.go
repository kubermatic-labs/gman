// Package glib contains methods for interactions with GSuite API
package glib

import (
	"context"
	"fmt"

	directoryv1 "google.golang.org/api/admin/directory/v1"

	"github.com/kubermatic-labs/gman/pkg/config"
)

// ListOrgUnits returns a list of all current organizational units from the API
func (ds *DirectoryService) ListOrgUnits(ctx context.Context) ([]*directoryv1.OrgUnit, error) {
	// OrgUnits do not use pagination and always return all units in a single API call.
	request, err := ds.Orgunits.List("my_customer").Type("all").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to list OrgUnits in domain: %v", err)
	}

	return request.OrganizationUnits, nil
}

// CreateOrgUnit creates a new org unit
func (ds *DirectoryService) CreateOrgUnit(ctx context.Context, ou *config.OrgUnit) error {
	newOU := createGSuiteOUFromConfig(ou)

	if _, err := ds.Orgunits.Insert("my_customer", newOU).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to create org unit: %v", err)
	}

	return nil
}

// DeleteOrgUnit deletes an org group
func (ds *DirectoryService) DeleteOrgUnit(ctx context.Context, ou *directoryv1.OrgUnit) error {
	// deletion can happen with the full orgunit's path *OR* it's unique ID
	if err := ds.Orgunits.Delete("my_customer", ou.OrgUnitId).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to delete org unit: %v", err)
	}

	return nil
}

// UpdateOrgUnit updates the remote org unit with config
func (ds *DirectoryService) UpdateOrgUnit(ctx context.Context, ou *config.OrgUnit) error {
	updatedOu := createGSuiteOUFromConfig(ou)

	// to update, we need the org unit's ID or its path;
	// we have neither, but since the path is always just "{parent}/{orgunit-name}",
	// we can construct it (there is no encoding/escaping in the paths, amazingly)
	path := ou.ParentOrgUnitPath + "/" + ou.Name

	if _, err := ds.Orgunits.Update("my_customer", path, updatedOu).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to update org unit: %v", err)
	}

	return nil
}

// createGSuiteOUFromConfig converts a OrgUnitConfig to (GSuite) admin.OrgUnit
func createGSuiteOUFromConfig(ou *config.OrgUnit) *directoryv1.OrgUnit {
	return &directoryv1.OrgUnit{
		Name:              ou.Name,
		Description:       ou.Description,
		ParentOrgUnitPath: ou.ParentOrgUnitPath,
	}
}
