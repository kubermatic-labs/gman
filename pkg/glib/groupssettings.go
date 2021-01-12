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
	"io/ioutil"
	"time"

	"golang.org/x/oauth2/google"
	directoryv1 "google.golang.org/api/admin/directory/v1"
	groupssettingsv1 "google.golang.org/api/groupssettings/v1"
	"google.golang.org/api/option"
)

type GroupsSettingsService struct {
	*groupssettingsv1.Service

	delay time.Duration
}

// NewGroupsSettingsService() creates a client for communicating with Google Groupssettings API.
func NewGroupsSettingsService(ctx context.Context, clientSecretFile string, impersonatedUserEmail string, delay time.Duration) (*GroupsSettingsService, error) {
	jsonCredentials, err := ioutil.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read JSON credentials: %v", err)
	}

	config, err := google.JWTConfigFromJSON(jsonCredentials, groupssettingsv1.AppsGroupsSettingsScope)
	if err != nil {
		return nil, fmt.Errorf("unable to process credentials: %v", err)
	}
	config.Subject = impersonatedUserEmail

	ts := config.TokenSource(ctx)

	srv, err := groupssettingsv1.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("unable to create a new Groupssettings Service: %v", err)
	}

	groupsService := &GroupsSettingsService{
		Service: srv,
		delay:   delay,
	}

	return groupsService, nil
}

func (gs *GroupsSettingsService) GetSettings(ctx context.Context, groupId string) (*groupssettingsv1.Groups, error) {
	request, err := gs.Groups.Get(groupId).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve group settings: %v", err)
	}

	return request, nil
}

func (gs *GroupsSettingsService) UpdateSettings(ctx context.Context, group *directoryv1.Group, settings *groupssettingsv1.Groups) (*groupssettingsv1.Groups, error) {
	updatedSettings, err := gs.Groups.Update(group.Email, settings).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to update a group settings: %v", err)
	}

	return updatedSettings, nil
}
