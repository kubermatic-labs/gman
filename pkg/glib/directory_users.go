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
	"sort"

	password "github.com/sethvargo/go-password/password"
	directoryv1 "google.golang.org/api/admin/directory/v1"

	"github.com/kubermatic-labs/gman/pkg/util"
)

func (ds *DirectoryService) ListUsers(ctx context.Context) ([]*directoryv1.User, error) {
	users := []*directoryv1.User{}
	token := ""

	for {
		request := ds.Users.List().Customer("my_customer").OrderBy("email").PageToken(token).Context(ctx)

		response, err := request.Do()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve list of users in domain: %v", err)
		}

		users = append(users, response.Users...)

		token = response.NextPageToken
		if token == "" {
			break
		}
	}

	return users, nil
}

func (ds *DirectoryService) CreateUser(ctx context.Context, user *directoryv1.User) (*directoryv1.User, error) {
	// generate a rand password
	pass, err := password.Generate(20, 5, 5, false, false)
	if err != nil {
		return nil, fmt.Errorf("unable to generate password: %v", err)
	}

	user.Password = pass
	user.ChangePasswordAtNextLogin = true

	createdUser, err := ds.Users.Insert(user).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create user: %v", err)
	}

	return createdUser, nil
}

func (ds *DirectoryService) DeleteUser(ctx context.Context, user *directoryv1.User) error {
	err := ds.Users.Delete(user.PrimaryEmail).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("unable to delete user: %v", err)
	}

	return nil
}

func (ds *DirectoryService) UpdateUser(ctx context.Context, user *directoryv1.User) (*directoryv1.User, error) {
	updatedUser, err := ds.Users.Update(user.PrimaryEmail, user).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to update user: %v", err)
	}

	return updatedUser, nil
}

type aliases struct {
	Aliases []struct {
		Alias        string `json:"alias"`
		PrimaryEmail string `json:"primaryEmail"`
	} `json:"aliases"`
}

func (ds *DirectoryService) GetUserAliases(ctx context.Context, user *directoryv1.User) ([]string, error) {
	data, err := ds.Users.Aliases.List(user.PrimaryEmail).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to list user aliases: %v", err)
	}

	aliases := aliases{}
	if err := util.ConvertToStruct(data, &aliases); err != nil {
		return nil, fmt.Errorf("failed to parse user aliases: %v", err)
	}

	result := []string{}
	for _, alias := range aliases.Aliases {
		result = append(result, alias.Alias)
	}

	sort.Strings(result)

	return result, nil
}

func (ds *DirectoryService) CreateUserAlias(ctx context.Context, user *directoryv1.User, alias string) error {
	newAlias := &directoryv1.Alias{
		Alias: alias,
	}

	if _, err := ds.Users.Aliases.Insert(user.PrimaryEmail, newAlias).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to create user alias: %v", err)
	}

	return nil
}

func (ds *DirectoryService) DeleteUserAlias(ctx context.Context, user *directoryv1.User, alias string) error {
	if err := ds.Users.Aliases.Delete(user.PrimaryEmail, alias).Context(ctx).Do(); err != nil {
		return fmt.Errorf("unable to delete user alias: %v", err)
	}

	return nil
}
