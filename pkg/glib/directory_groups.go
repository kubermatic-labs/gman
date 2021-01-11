// Package glib contains methods for interactions with GSuite API
package glib

import (
	"context"
	"fmt"
	"strconv"

	directoryv1 "google.golang.org/api/admin/directory/v1"
	groupssettingsv1 "google.golang.org/api/groupssettings/v1"

	"github.com/kubermatic-labs/gman/pkg/config"
)

// ListGroups returns a list of all current groups from the API
func (ds *DirectoryService) ListGroups(ctx context.Context) ([]*directoryv1.Group, error) {
	groups := []*directoryv1.Group{}
	token := ""

	for {
		request := ds.Groups.List().Customer("my_customer").OrderBy("email").Context(ctx).PageToken(token)

		response, err := request.Do()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve list of groups in domain: %v", err)
		}

		groups = append(groups, response.Groups...)

		token = response.NextPageToken
		if token == "" {
			break
		}
	}

	return groups, nil
}

// CreateGroup creates a new group in GSuite via their API
func (ds *DirectoryService) CreateGroup(ctx context.Context, group *directoryv1.Group) (*directoryv1.Group, error) {
	updatedGroup, err := ds.Groups.Insert(group).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create group: %v", err)
	}

	return updatedGroup, nil
}

// DeleteGroup deletes a group in GSuite via their API
func (ds *DirectoryService) DeleteGroup(ctx context.Context, group *directoryv1.Group) error {
	if err := ds.Groups.Delete(group.Email).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to delete group: %v", err)
	}

	return nil
}

// UpdateGroup updates the remote group with config
func (ds *DirectoryService) UpdateGroup(ctx context.Context, group *directoryv1.Group) (*directoryv1.Group, error) {
	updatedGroup, err := ds.Groups.Update(group.Email, group).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to update a group: %v", err)
	}

	return updatedGroup, nil
}

// CreateGSuiteGroupFromConfig converts a ConfigGroup to (GSuite) directoryv1.Group
func CreateGSuiteGroupFromConfig(group *config.Group) (*directoryv1.Group, *groupssettingsv1.Groups) {
	googleGroup := &directoryv1.Group{
		Name:  group.Name,
		Email: group.Email,
	}
	if group.Description != "" {
		googleGroup.Description = group.Description
	}

	groupSettings := &groupssettingsv1.Groups{
		WhoCanContactOwner:   group.WhoCanContactOwner,
		WhoCanViewMembership: group.WhoCanViewMembership,
		WhoCanApproveMembers: group.WhoCanApproveMembers,
		WhoCanPostMessage:    group.WhoCanPostMessage,
		WhoCanJoin:           group.WhoCanJoin,
		IsArchived:           strconv.FormatBool(group.IsArchived),
		ArchiveOnly:          strconv.FormatBool(group.IsArchived),
		AllowExternalMembers: strconv.FormatBool(group.AllowExternalMembers),
	}

	return googleGroup, groupSettings
}

func CreateConfigGroupFromGSuite(googleGroup *directoryv1.Group, members []*directoryv1.Member, gSettings *groupssettingsv1.Groups) (config.Group, error) {
	boolAllowExternalMembers, err := strconv.ParseBool(gSettings.AllowExternalMembers)
	if err != nil {
		return config.Group{}, fmt.Errorf("could not parse 'AllowExternalMembers' value from string to bool: %v", err)
	}

	boolIsArchived, err := strconv.ParseBool(gSettings.IsArchived)
	if err != nil {
		return config.Group{}, fmt.Errorf("could not parse 'IsArchived' value from string to bool: %v", err)
	}

	configGroup := config.Group{
		Name:                 googleGroup.Name,
		Email:                googleGroup.Email,
		Description:          googleGroup.Description,
		WhoCanContactOwner:   gSettings.WhoCanContactOwner,
		WhoCanViewMembership: gSettings.WhoCanViewMembership,
		WhoCanApproveMembers: gSettings.WhoCanApproveMembers,
		WhoCanPostMessage:    gSettings.WhoCanPostMessage,
		WhoCanJoin:           gSettings.WhoCanJoin,
		AllowExternalMembers: boolAllowExternalMembers,
		IsArchived:           boolIsArchived,
		Members:              []config.Member{},
	}

	for _, m := range members {
		configGroup.Members = append(configGroup.Members, config.Member{
			Email: m.Email,
			Role:  m.Role,
		})
	}

	return configGroup, nil
}

//----------------------------------------//
//   Group Member handling                //
//----------------------------------------//

// ListMembers returns a list of all current group members form the API
func (ds *DirectoryService) ListMembers(ctx context.Context, group *directoryv1.Group) ([]*directoryv1.Member, error) {
	members := []*directoryv1.Member{}
	token := ""

	for {
		request := ds.Members.List(group.Email).PageToken(token).Context(ctx)

		response, err := request.Do()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve list of members in group %s: %v", group.Name, err)
		}

		members = append(members, response.Members...)

		token = response.NextPageToken
		if token == "" {
			break
		}
	}

	return members, nil
}

// AddNewMember adds a new member to a group in GSuite
func (ds *DirectoryService) AddNewMember(ctx context.Context, groupEmail string, member *config.Member) error {
	newMember := createGSuiteGroupMemberFromConfig(member)

	if _, err := ds.Members.Insert(groupEmail, newMember).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to add member to group: %v", err)
	}

	return nil
}

// RemoveMember removes a member from a group in Gsuite
func (ds *DirectoryService) RemoveMember(ctx context.Context, groupEmail string, member *directoryv1.Member) error {
	if err := ds.Members.Delete(groupEmail, member.Email).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to delete member from group: %v", err)
	}

	return nil
}

// MemberExists checks if member exists in group
func (ds *DirectoryService) MemberExists(ctx context.Context, group *directoryv1.Group, member *config.Member) (bool, error) {
	exists, err := ds.Members.HasMember(group.Email, member.Email).Context(ctx).Do()
	if err != nil {
		return false, fmt.Errorf("unable to check if member exists in group: %v", err)
	}

	return exists.IsMember, nil
}

// UpdateMembership changes the role of the member
func (ds *DirectoryService) UpdateMembership(ctx context.Context, groupEmail string, member *config.Member) error {
	newMember := createGSuiteGroupMemberFromConfig(member)

	if _, err := ds.Members.Update(groupEmail, member.Email, newMember).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to update member in group: %v", err)
	}

	return nil
}

// createGSuiteGroupMemberFromConfig converts a ConfigMember to (GSuite) admin.Member
func createGSuiteGroupMemberFromConfig(member *config.Member) *directoryv1.Member {
	googleMember := &directoryv1.Member{
		Email: member.Email,
		Role:  member.Role,
	}

	return googleMember
}
