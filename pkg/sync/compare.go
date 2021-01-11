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
