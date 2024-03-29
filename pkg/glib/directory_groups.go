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

package glib

import (
	"context"
	"sort"
	"strings"

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
			return nil, err
		}

		groups = append(groups, response.Groups...)

		token = response.NextPageToken
		if token == "" {
			break
		}
	}

	sort.SliceStable(groups, func(i, j int) bool {
		return strings.ToLower(groups[i].Name) < strings.ToLower(groups[j].Name)
	})

	return groups, nil
}

// CreateGroup creates a new group in GSuite via their API
func (ds *DirectoryService) CreateGroup(ctx context.Context, group *directoryv1.Group) (*directoryv1.Group, error) {
	updatedGroup, err := ds.Groups.Insert(group).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return updatedGroup, nil
}

// DeleteGroup deletes a group in GSuite via their API
func (ds *DirectoryService) DeleteGroup(ctx context.Context, group *directoryv1.Group) error {
	if err := ds.Groups.Delete(group.Email).Context(ctx).Do(); err != nil {
		return err
	}

	return nil
}

// UpdateGroup updates the remote group with config
func (ds *DirectoryService) UpdateGroup(ctx context.Context, oldGroup *directoryv1.Group, newGroup *directoryv1.Group) (*directoryv1.Group, error) {
	updatedGroup, err := ds.Groups.Update(oldGroup.Email, newGroup).Context(ctx).Do()
	if err != nil {
		return nil, err
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
			return nil, err
		}

		members = append(members, response.Members...)

		token = response.NextPageToken
		if token == "" {
			break
		}
	}

	sort.SliceStable(members, func(i, j int) bool {
		return members[i].Email < members[j].Email
	})

	return members, nil
}

// AddNewMember adds a new member to a group in GSuite
func (ds *DirectoryService) AddNewMember(ctx context.Context, group *directoryv1.Group, member *directoryv1.Member) error {
	if _, err := ds.Members.Insert(group.Email, member).Context(ctx).Do(); err != nil {
		return err
	}

	return nil
}

// RemoveMember removes a member from a group in Gsuite
func (ds *DirectoryService) RemoveMember(ctx context.Context, group *directoryv1.Group, member *directoryv1.Member) error {
	// do NOT use the member email here, as it will lead to "Error 404: Resource Not Found: email, notFound." errors
	if err := ds.Members.Delete(group.Id, member.Id).Context(ctx).Do(); err != nil {
		return err
	}

	return nil
}

// UpdateMembership changes the role of the member
func (ds *DirectoryService) UpdateMembership(ctx context.Context, group *directoryv1.Group, member *directoryv1.Member) error {
	// do NOT use the member email here, as it will lead to "Error 404: Resource Not Found: email, notFound." errors
	if _, err := ds.Members.Update(group.Id, member.Id, member).Context(ctx).Do(); err != nil {
		return err
	}

	return nil
}
