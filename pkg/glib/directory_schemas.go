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

	directoryv1 "google.golang.org/api/admin/directory/v1"
)

func (ds *DirectoryService) GetSchema(ctx context.Context, name string) (*directoryv1.Schema, error) {
	return ds.Schemas.Get("my_customer", name).Context(ctx).Do()
}

func (ds *DirectoryService) CreateSchema(ctx context.Context, schema *directoryv1.Schema) (*directoryv1.Schema, error) {
	return ds.Schemas.Insert("my_customer", schema).Context(ctx).Do()
}

func (ds *DirectoryService) UpdateSchema(ctx context.Context, oldSchema *directoryv1.Schema, newSchema *directoryv1.Schema) (*directoryv1.Schema, error) {
	return ds.Schemas.Update("my_customer", oldSchema.SchemaId, newSchema).Context(ctx).Do()
}
