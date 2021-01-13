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

package config

import (
	"sort"
	"strings"
)

func (c *Config) DefaultUsers() error {
	for idx, user := range c.Users {
		if user.OrgUnitPath == "" {
			user.OrgUnitPath = "/"
		}

		c.Users[idx] = user
	}

	return nil
}

func (c *Config) UndefaultUsers() error {
	for idx, user := range c.Users {
		if user.OrgUnitPath == "/" {
			user.OrgUnitPath = ""
		}

		c.Users[idx] = user
	}

	return nil
}

func (c *Config) DefaultGroups() error {
	for idx, group := range c.Groups {
		if group.WhoCanJoin == "" {
			group.WhoCanJoin = GroupOptionWhoCanJoinDefault
		}

		if group.WhoCanPostMessage == "" {
			group.WhoCanPostMessage = GroupOptionWhoCanPostMessageDefault
		}

		if group.WhoCanApproveMembers == "" {
			group.WhoCanApproveMembers = GroupOptionWhoCanApproveMembersDefault
		}

		if group.WhoCanContactOwner == "" {
			group.WhoCanContactOwner = GroupOptionWhoCanContactOwnerDefault
		}

		if group.WhoCanViewMembership == "" {
			group.WhoCanViewMembership = GroupOptionWhoCanViewMembershipDefault
		}

		group.WhoCanJoin = strings.ToUpper(group.WhoCanJoin)
		group.WhoCanPostMessage = strings.ToUpper(group.WhoCanPostMessage)
		group.WhoCanApproveMembers = strings.ToUpper(group.WhoCanApproveMembers)
		group.WhoCanContactOwner = strings.ToUpper(group.WhoCanContactOwner)
		group.WhoCanViewMembership = strings.ToUpper(group.WhoCanViewMembership)

		for n, member := range group.Members {
			if member.Role == "" {
				member.Role = MemberRoleMember
			}

			member.Role = strings.ToUpper(member.Role)
			group.Members[n] = member
		}

		c.Groups[idx] = group
	}

	return nil
}

func (c *Config) UndefaultGroups() error {
	for idx, group := range c.Groups {
		if group.WhoCanJoin == GroupOptionWhoCanJoinDefault {
			group.WhoCanJoin = ""
		}

		if group.WhoCanPostMessage == GroupOptionWhoCanPostMessageDefault {
			group.WhoCanPostMessage = ""
		}

		if group.WhoCanApproveMembers == GroupOptionWhoCanApproveMembersDefault {
			group.WhoCanApproveMembers = ""
		}

		if group.WhoCanContactOwner == GroupOptionWhoCanContactOwnerDefault {
			group.WhoCanContactOwner = ""
		}

		if group.WhoCanViewMembership == GroupOptionWhoCanViewMembershipDefault {
			group.WhoCanViewMembership = ""
		}

		for n, member := range group.Members {
			if member.Role == MemberRoleMember {
				member.Role = ""
			}

			group.Members[n] = member
		}

		c.Groups[idx] = group
	}

	return nil
}

func (c *Config) DefaultOrgUnits() error {
	for idx, orgUnit := range c.OrgUnits {
		if orgUnit.ParentOrgUnitPath == "" {
			orgUnit.ParentOrgUnitPath = "/"
		}

		c.OrgUnits[idx] = orgUnit
	}

	return nil
}

func (c *Config) UndefaultOrgUnits() error {
	for idx, orgUnit := range c.OrgUnits {
		if orgUnit.ParentOrgUnitPath == "/" {
			orgUnit.ParentOrgUnitPath = ""
		}

		c.OrgUnits[idx] = orgUnit
	}

	return nil
}

func (c *Config) Sort() {
	sort.SliceStable(c.OrgUnits, func(i, j int) bool {
		return strings.ToLower(c.OrgUnits[i].Name) < strings.ToLower(c.OrgUnits[j].Name)
	})

	sort.SliceStable(c.Users, func(i, j int) bool {
		return c.Users[i].PrimaryEmail < c.Users[j].PrimaryEmail
	})

	sort.SliceStable(c.Groups, func(i, j int) bool {
		return strings.ToLower(c.Groups[i].Name) < strings.ToLower(c.Groups[j].Name)
	})

	for idx, user := range c.Users {
		user.Sort()
		c.Users[idx] = user
	}

	for idx, group := range c.Groups {
		group.Sort()
		c.Groups[idx] = group
	}
}
