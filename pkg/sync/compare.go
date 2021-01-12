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

package sync

import (
	directoryv1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/groupssettings/v1"

	"github.com/kubermatic-labs/gman/pkg/config"
)

func orgUnitUpToDate(configured config.OrgUnit, live *directoryv1.OrgUnit) bool {
	return configured.Description == live.Description &&
		configured.ParentOrgUnitPath != live.ParentOrgUnitPath &&
		configured.BlockInheritance != live.BlockInheritance
}

func userUpToDate(configured config.User, live *directoryv1.User, liveLicenses []config.License, liveAliases []string) bool {
	// currentUserConfig := glib.CreateConfigUserFromGSuite(currentUser, currentUserLicenses)
	// if !reflect.DeepEqual(currentUserConfig, configured) {
	// 	usersToUpdate = append(usersToUpdate, configured)
	// }

	return configured.PrimaryEmail == live.PrimaryEmail
}

func groupUpToDate(configured config.Group, live *directoryv1.Group, liveMembers []*directoryv1.Member, settings *groupssettings.Groups) bool {
	// currentUserConfig := glib.CreateConfigUserFromGSuite(currentUser, currentUserLicenses)
	// if !reflect.DeepEqual(currentUserConfig, configured) {
	// 	usersToUpdate = append(usersToUpdate, configured)
	// }

	return configured.Email == live.Email
}

func memberUpToDate(configured config.Member, live *directoryv1.Member) bool {
	// currentUserConfig := glib.CreateConfigUserFromGSuite(currentUser, currentUserLicenses)
	// if !reflect.DeepEqual(currentUserConfig, configured) {
	// 	usersToUpdate = append(usersToUpdate, configured)
	// }

	return configured.Email == live.Email
}
