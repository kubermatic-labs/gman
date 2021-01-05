#!/usr/bin/env bash

# Copyright 2020 The Kubermatic Kubernetes Platform contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script creates the Docker images on quay.io. Docker is not handled
# by goreleaser because of our internal CI infrastructure and separation
# of credentials.

set -euo pipefail

cd $(dirname $0)/../..

repo="quay.io/kubermatic-labs/gman"
tag="$(git describe --tags --exact-match)"
image="$repo:$tag"

set -x
docker build -t "$image" .
docker push "$image"
