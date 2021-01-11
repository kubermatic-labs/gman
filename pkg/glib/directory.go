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

	delay time.Duration
}

// NewDirectoryService() creates a client for communicating with Google Directory API,
// returns a service object authorized to perform actions in Gsuite.
func NewDirectoryService(ctx context.Context, clientSecretFile string, impersonatedUserEmail string, delay time.Duration, scopes ...string) (*DirectoryService, error) {
	jsonCredentials, err := ioutil.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read json credentials: %v", err)
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
		Service: srv,
		delay:   delay,
	}

	return dirService, nil
}