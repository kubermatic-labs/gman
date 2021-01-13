/*
Copyright 2021 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

func (ds *DirectoryService) UpdateOrgUnit(ctx context.Context, oldUnit *directoryv1.OrgUnit, newUnit *directoryv1.OrgUnit) error {
	if _, err := ds.Orgunits.Update("my_customer", oldUnit.OrgUnitId, newUnit).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to update org unit: %v", err)
	}

	return nil
}
