// Package glib contains methods for interactions with GSuite API
package glib

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"golang.org/x/oauth2/google"
	directoryv1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

type DirectoryService struct {
	*directoryv1.Service

	organization string
	delay        time.Duration
}

// NewDirectoryService() creates a client for communicating with Google Directory API.
func NewDirectoryService(ctx context.Context, organization string, clientSecretFile string, impersonatedUserEmail string, delay time.Duration, scopes ...string) (*DirectoryService, error) {
	jsonCredentials, err := ioutil.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read JSON credentials: %v", err)
	}

	config, err := google.JWTConfigFromJSON(jsonCredentials, scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to process credentials: %v", err)
	}
	config.Subject = impersonatedUserEmail

	ts := config.TokenSource(ctx)

	srv, err := directoryv1.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("unable to create a new directory service: %v", err)
	}

	dirService := &DirectoryService{
		Service:      srv,
		organization: organization,
		delay:        delay,
	}

	return dirService, nil
}
