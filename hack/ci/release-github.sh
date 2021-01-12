#!/usr/bin/env bash

# Copyright 2021 The Kubermatic Kubernetes Platform contributors.
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

# This script is creating release binaries and
# Docker images via goreleaser. It's meant to
# run in the Kubermatic CI environment only,
# as it requires GitHub and quay.io credentials.

set -euo pipefail

cd $(dirname $0)/../..

git remote add origin git@github.com:kubermatic-labs/gman.git
export GITHUB_TOKEN=$(cat /etc/github/oauth | tr -d '\n')

goreleaser release
