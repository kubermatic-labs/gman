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

package export

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/glib"
)

func ExportOrgUnits(ctx context.Context, directorySrv *glib.DirectoryService) ([]config.OrgUnit, error) {
	orgUnits, err := directorySrv.ListOrgUnits(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list org units: %v", err)
	}

	result := []config.OrgUnit{}
	for _, ou := range orgUnits {
		log.Printf("  %s", ou.Name)
		result = append(result, config.ToConfigOrgUnit(ou))
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func ExportUsers(ctx context.Context, directorySrv *glib.DirectoryService, licensingSrv *glib.LicensingService, licenseStatus *glib.LicenseStatus) ([]config.User, error) {
	users, err := directorySrv.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %v", err)
	}

	result := []config.User{}
	for _, user := range users {
		log.Printf("  %s", user.PrimaryEmail)

		userLicenses := licenseStatus.GetLicensesForUser(user)

		configUser, err := config.ToConfigUser(user, userLicenses)
		if err != nil {
			return nil, fmt.Errorf("failed to convert user: %v", err)
		}

		result = append(result, configUser)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].PrimaryEmail < result[j].PrimaryEmail
	})

	return result, nil
}

func ExportGroups(ctx context.Context, directorySrv *glib.DirectoryService, groupsSettingsSrv *glib.GroupsSettingsService) ([]config.Group, error) {
	groups, err := directorySrv.ListGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %v", err)
	}

	result := []config.Group{}
	for _, group := range groups {
		log.Printf("  %s", group.Name)

		settings, err := groupsSettingsSrv.GetSettings(ctx, group.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to get group settings: %v", err)
		}

		members, err := directorySrv.ListMembers(ctx, group)
		if err != nil {
			return nil, fmt.Errorf("failed to list members: %v", err)
		}

		configGroup, err := config.ToConfigGroup(group, settings, members)
		if err != nil {
			return nil, fmt.Errorf("failed to create config group: %v", err)
		}

		result = append(result, configGroup)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}
