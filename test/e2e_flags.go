/*
Copyright 2019 The Knative Authors
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

// This file contains logic to encapsulate flags which are needed to specify
// what cluster, etc. to use for e2e tests.

package test

import (
	pkgTest "knative.dev/pkg/test"
	"knative.dev/pkg/test/logging"
)

// EventingSourcesFlags holds the command line flags specific to knative/eventing-contrib
var EventingSourcesFlags = initializeEventingSourcesFlags()

// EventingSourcesEnvironmentFlags holds the e2e flags needed only by the eventing-contrib repo
type EventingSourcesEnvironmentFlags struct {
}

// initializeEventingSourcesFlags registers flags used by e2e tests, calling flag.Parse() here would fail in
// go1.13+, see https://github.com/knative/test-infra/issues/1329 for details
func initializeEventingSourcesFlags() *EventingSourcesEnvironmentFlags {
	var f EventingSourcesEnvironmentFlags

	logging.InitializeLogger(pkgTest.Flags.LogVerbose)

	if pkgTest.Flags.EmitMetrics {
		logging.InitializeMetricExporter("eventing-sources")
	}

	return &f
}
