// Package glib contains methods for interactions with GSuite API
package glib

import (
	"context"
	"fmt"

	directoryv1 "google.golang.org/api/admin/directory/v1"
)

func (ds *DirectoryService) ListOrgUnits(ctx context.Context) ([]*directoryv1.OrgUnit, error) {
	// OrgUnits do not use pagination and always return all units in a single API call.
	request, err := ds.Orgunits.List("my_customer").Type("all").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to list OrgUnits in domain: %v", err)
	}

	return request.OrganizationUnits, nil
}

func (ds *DirectoryService) CreateOrgUnit(ctx context.Context, orgUnit *directoryv1.OrgUnit) error {
	if _, err := ds.Orgunits.Insert("my_customer", orgUnit).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to create org unit: %v", err)
	}

	return nil
}

func (ds *DirectoryService) DeleteOrgUnit(ctx context.Context, orgUnit *directoryv1.OrgUnit) error {
	// deletion can happen with the full orgunit's path *OR* it's unique ID
	if err := ds.Orgunits.Delete("my_customer", orgUnit.OrgUnitId).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to delete org unit: %v", err)
	}

	return nil
}

func (ds *DirectoryService) UpdateOrgUnit(ctx context.Context, orgUnit *directoryv1.OrgUnit) error {
	// to update, we need the org unit's ID or its path;
	// we possibly have neither, but since the path is always just "{parent}/{orgunit-name}",
	// we can construct it (there is no encoding/escaping in the paths, amazingly)
	path := orgUnit.OrgUnitPath
	if path == "" {
		path = orgUnit.ParentOrgUnitPath + "/" + orgUnit.Name
	}

	if _, err := ds.Orgunits.Update("my_customer", path, orgUnit).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to update org unit: %v", err)
	}

	return nil
}
