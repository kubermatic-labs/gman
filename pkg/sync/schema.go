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

package sync

import (
	"context"
	"log"

	directoryv1 "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"

	"github.com/kubermatic-labs/gman/pkg/config"
	"github.com/kubermatic-labs/gman/pkg/glib"
)

func SyncSchema(
	ctx context.Context,
	directorySrv *glib.DirectoryService,
	confirm bool,
) error {
	log.Println("⇄ Syncing schema…")

	if !confirm {
		return nil
	}

	desiredSchema := &directoryv1.Schema{
		DisplayName: "GMan",
		SchemaName:  config.SchemaName,
		Fields: []*directoryv1.SchemaFieldSpec{
			{
				FieldName:      config.PasswordHashCustomField,
				FieldType:      "STRING",
				ReadAccessType: "ADMINS_AND_SELF",
				Indexed:        googleapi.Bool(false),
			},
		},
	}

	schema, err := directorySrv.GetSchema(ctx, config.SchemaName)

	if err != nil {
		_, err = directorySrv.CreateSchema(ctx, desiredSchema)
	} else {
		_, err = directorySrv.UpdateSchema(ctx, schema, desiredSchema)
	}

	return err
}
