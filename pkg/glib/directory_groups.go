// Package glib contains methods for interactions with GSuite API
package glib

import (
	"context"
	"fmt"

	directoryv1 "google.golang.org/api/admin/directory/v1"
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
func (ds *DirectoryService) AddNewMember(ctx context.Context, group *directoryv1.Group, member *directoryv1.Member) error {
	if _, err := ds.Members.Insert(group.Email, member).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to add member to group: %v", err)
	}

	return nil
}

// RemoveMember removes a member from a group in Gsuite
func (ds *DirectoryService) RemoveMember(ctx context.Context, group *directoryv1.Group, member *directoryv1.Member) error {
	if err := ds.Members.Delete(group.Email, member.Email).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to delete member from group: %v", err)
	}

	return nil
}

// UpdateMembership changes the role of the member
func (ds *DirectoryService) UpdateMembership(ctx context.Context, group *directoryv1.Group, member *directoryv1.Member) error {
	if _, err := ds.Members.Update(group.Email, member.Email, member).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to update member in group: %v", err)
	}

	return nil
}
