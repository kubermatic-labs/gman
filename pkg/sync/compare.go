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
	"reflect"

	directoryv1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/groupssettings/v1"

	"github.com/kubermatic-labs/gman/pkg/config"
)

func orgUnitUpToDate(configured config.OrgUnit, live *directoryv1.OrgUnit) bool {
	converted := config.ToConfigOrgUnit(live)

	return reflect.DeepEqual(configured, converted)
}

func userUpToDate(configured config.User, live *directoryv1.User, liveLicenses []config.License, liveAliases []string) bool {
	converted, err := config.ToConfigUser(live, liveLicenses)
	if err != nil {
		return false
	}

	if converted.Aliases == nil {
		converted.Aliases = []string{}
	}

	if configured.Aliases == nil {
		configured.Aliases = []string{}
	}

	return reflect.DeepEqual(configured, converted)
}

func groupUpToDate(configured config.Group, live *directoryv1.Group, liveMembers []*directoryv1.Member, settings *groupssettings.Groups) bool {
	converted, err := config.ToConfigGroup(live, settings, liveMembers)
	if err != nil {
		return false
	}

	if converted.Members == nil {
		converted.Members = []config.Member{}
	}

	if configured.Members == nil {
		configured.Members = []config.Member{}
	}

	return reflect.DeepEqual(configured, converted)
}

func memberUpToDate(configured config.Member, live *directoryv1.Member) bool {
	converted := config.ToConfigGroupMember(live)

	return reflect.DeepEqual(configured, converted)
}
