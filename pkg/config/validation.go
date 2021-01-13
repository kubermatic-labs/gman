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
	"errors"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

// validateEmailFormat is a helper function that checks for existance of '@' and the length of the address
func validateEmailFormat(email string) bool {
	return len(email) < 129 && strings.Contains(email, "@")
}

func (c *Config) ValidateUsers() []error {
	var allErrors []error
	re164 := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

	// validate organization
	if c.Organization == "" {
		allErrors = append(allErrors, errors.New("no organization configured"))
	}

	// validate users
	userEmails := sets.NewString()
	for _, user := range c.Users {
		if userEmails.Has(user.PrimaryEmail) {
			allErrors = append(allErrors, fmt.Errorf("duplicate user defined (user: %s)", user.PrimaryEmail))
		}
		userEmails.Insert(user.PrimaryEmail)

		if user.PrimaryEmail == "" {
			allErrors = append(allErrors, fmt.Errorf("primary email is required (user: %s)", user.LastName))
		} else if !validateEmailFormat(user.PrimaryEmail) {
			allErrors = append(allErrors, fmt.Errorf("primary email is not a valid email-address (user: %s)", user.PrimaryEmail))
		}

		if user.FirstName == "" || user.LastName == "" {
			allErrors = append(allErrors, fmt.Errorf("given and family names are required (user: %s)", user.PrimaryEmail))
		}

		if user.RecoveryEmail != "" && !validateEmailFormat(user.RecoveryEmail) {
			allErrors = append(allErrors, fmt.Errorf("recovery email is not a valid email-address (user: %s)", user.PrimaryEmail))
		}

		if user.Employee.ManagerEmail != "" && !validateEmailFormat(user.Employee.ManagerEmail) {
			allErrors = append(allErrors, fmt.Errorf("manager's email is not a valid email-address (user: %s)", user.PrimaryEmail))
		}

		if user.RecoveryPhone != "" && !re164.MatchString(user.RecoveryPhone) {
			allErrors = append(allErrors, fmt.Errorf("invalid format of recovery phone (user: %s). The phone number must be in the E.164 format, starting with the plus sign (+). Example: +16506661212.", user.PrimaryEmail))
		}

		if len(user.Aliases) > 0 {
			for _, alias := range user.Aliases {
				if !validateEmailFormat(alias) {
					allErrors = append(allErrors, fmt.Errorf("alias email is not a valid email-address (user: %s)", user.PrimaryEmail))
				}
			}
		}

		if len(user.Licenses) > 0 {
			for _, license := range user.Licenses {
				found := false
				for _, permLicense := range AllLicenses {
					if license == permLicense.Name {
						found = true
					}
				}
				if !found {
					allErrors = append(allErrors, fmt.Errorf("wrong value specified for the user license (user: %s, license: %s)", user.PrimaryEmail, license))
				}
			}
		}
	}

	return allErrors
}

func (c *Config) ValidateGroups() []error {
	var allErrors []error

	// validate organization
	if c.Organization == "" {
		allErrors = append(allErrors, errors.New("no organization configured"))
	}

	// validate groups
	groupEmails := sets.NewString()
	for _, group := range c.Groups {
		if groupEmails.Has(group.Email) {
			allErrors = append(allErrors, fmt.Errorf("[group: %s] duplicate group email defined", group.Email))
		}
		groupEmails.Insert(group.Email)

		if !validateEmailFormat(group.Email) {
			allErrors = append(allErrors, fmt.Errorf("[group: %s] group email is not a valid email address", group.Email))
		}

		if group.WhoCanContactOwner != "" {
			if !allWhoCanContactOwnerOptions.Has(strings.ToUpper(group.WhoCanContactOwner)) {
				allErrors = append(allErrors, fmt.Errorf("[group: %s] invalid value specified for 'whoCanContactOwner' field, must be one of %v", group.Name, allWhoCanContactOwnerOptions.List()))
			}
		}

		if group.WhoCanViewMembership != "" {
			if !allWhoCanViewMembershipOptions.Has(strings.ToUpper(group.WhoCanViewMembership)) {
				allErrors = append(allErrors, fmt.Errorf("[group: %s] invalid value specified for 'whoCanViewMembers' field, must be one of %v", group.Name, allWhoCanViewMembershipOptions.List()))
			}
		}

		if group.WhoCanApproveMembers != "" {
			if !allWhoCanApproveMembersOptions.Has(strings.ToUpper(group.WhoCanApproveMembers)) {
				allErrors = append(allErrors, fmt.Errorf("[group: %s] invalid value specified for 'whoCanApproveMembers' field, must be one of %v", group.Name, allWhoCanApproveMembersOptions.List()))
			}
		}

		if group.WhoCanPostMessage != "" {
			if !allWhoCanPostMessageOptions.Has(strings.ToUpper(group.WhoCanPostMessage)) {
				allErrors = append(allErrors, fmt.Errorf("[group: %s] invalid value specified for 'whoCanPostMessage' field, must be one of %v", group.Name, allWhoCanPostMessageOptions.List()))
			}
		}

		if group.WhoCanJoin != "" {
			if !allWhoCanJoinOptions.Has(strings.ToUpper(group.WhoCanJoin)) {
				allErrors = append(allErrors, fmt.Errorf("[group: %s] invalid value specified for 'whoCanJoin' field, must be one of %v", group.Name, allWhoCanJoinOptions.List()))
			}
		}

		memberEmails := sets.NewString()
		for _, member := range group.Members {
			if memberEmails.Has(member.Email) {
				allErrors = append(allErrors, fmt.Errorf("[group: %s] duplicate member %q defined", group.Name, member.Email))
			}
			memberEmails.Insert(member.Email)

			if !allMemberRoles.Has(member.Role) {
				allErrors = append(allErrors, fmt.Errorf("[group: %s] invalid member role specified for %q, must be one of %v", group.Name, member.Email, allMemberRoles.List()))
			}
		}
	}

	return allErrors
}

func (c *Config) ValidateOrgUnits() []error {
	var allErrors []error

	// validate organization
	if c.Organization == "" {
		allErrors = append(allErrors, errors.New("no organization configured"))
	}

	// validate org units
	unitNames := sets.NewString()
	for _, orgUnit := range c.OrgUnits {
		if unitNames.Has(orgUnit.Name) {
			allErrors = append(allErrors, fmt.Errorf("[org unit: %s] duplicate org unit defined", orgUnit.Name))
		}
		unitNames.Insert(orgUnit.Name)

		if orgUnit.Name == "" {
			allErrors = append(allErrors, fmt.Errorf("[org unit: %s] no name specified", orgUnit.Name))
		}

		if orgUnit.ParentOrgUnitPath == "" {
			allErrors = append(allErrors, fmt.Errorf("[org unit: %s] no parentOrgUnitPath specified", orgUnit.Name))
		} else if !strings.HasPrefix(orgUnit.ParentOrgUnitPath, "/") {
			allErrors = append(allErrors, fmt.Errorf("[org unit: %s] parentOrgUnitPath must start with a slash", orgUnit.Name))
		}

	}

	return allErrors
}

func ValidateLicenses(licenses []License) []error {
	var allErrors []error

	// validate org units
	licenseNames := sets.NewString()
	licenseIdentifiers := sets.NewString()
	for _, license := range licenses {
		identifier := fmt.Sprintf("%s:%s", license.ProductId, license.SkuId)

		if licenseNames.Has(license.Name) {
			allErrors = append(allErrors, fmt.Errorf("[license: %s] duplicate license name defined", license.Name))
		}

		if licenseIdentifiers.Has(identifier) {
			allErrors = append(allErrors, fmt.Errorf("[license: %s] duplicate license productId/skuId combination defined", license.Name))
		}

		licenseNames.Insert(license.Name)
		licenseIdentifiers.Insert(identifier)

		if license.Name == "" {
			allErrors = append(allErrors, fmt.Errorf("[license: %s] no name specified", license.Name))
		}

		if license.ProductId == "" {
			allErrors = append(allErrors, fmt.Errorf("[license: %s] no productId specified", license.Name))
		}

		if license.SkuId == "" {
			allErrors = append(allErrors, fmt.Errorf("[license: %s] no skuId specified", license.Name))
		}
	}

	return allErrors
}
