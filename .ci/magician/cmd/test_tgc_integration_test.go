/*
* Copyright 2026 Google LLC. All Rights Reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */
package cmd

import (
	"reflect"
	"sort"
	"testing"
)

func TestShouldRunTests(t *testing.T) {
	cases := []struct {
		name          string
		changedFiles  []string
		expectedRun   bool
		expectedPaths []string
	}{
		{
			name:          "relevant go file in services folder",
			changedFiles:  []string{"test/services/alloydb/alloydb_cluster_generated_test.go"},
			expectedRun:   true,
			expectedPaths: []string{"./test/services/alloydb"},
		},
		{
			name:          "non-go file",
			changedFiles:  []string{"docs/supported_resources.md"},
			expectedRun:   false,
			expectedPaths: nil,
		},
		{
			name:          "ignored directory cai2hcl",
			changedFiles:  []string{"cai2hcl/services/certificatemanager/certificate.go"},
			expectedRun:   false,
			expectedPaths: nil,
		},
		{
			name:          "ignored directory tfplan2cai",
			changedFiles:  []string{"tfplan2cai/converters/google/resources/services/biglakeiceberg"},
			expectedRun:   false,
			expectedPaths: nil,
		},
		{
			name:          "pkg/services file (ignored by default)",
			changedFiles:  []string{"pkg/services/compute/compute_disk.go"},
			expectedRun:   false,
			expectedPaths: nil,
		},
		{
			name:          "pkg/services cai2hcl file (no longer exception)",
			changedFiles:  []string{"pkg/services/compute/compute_disk_cai2hcl.go"},
			expectedRun:   false,
			expectedPaths: nil,
		},
		{
			name:          "pkg/services tfplan2cai file (no longer exception)",
			changedFiles:  []string{"pkg/services/compute/compute_disk_tfplan2cai.go"},
			expectedRun:   false,
			expectedPaths: nil,
		},
		{
			name:          "pkg/cai2hcl file (core exception)",
			changedFiles:  []string{"pkg/cai2hcl/converters/convert_resource.go"},
			expectedRun:   true,
			expectedPaths: nil,
		},
		{
			name:          "pkg/tfplan2cai file (core exception)",
			changedFiles:  []string{"pkg/tfplan2cai/converters/cai/convert.go"},
			expectedRun:   true,
			expectedPaths: nil,
		},
		{
			name:          "pkg/caiasset file (core exception)",
			changedFiles:  []string{"pkg/caiasset/asset.go"},
			expectedRun:   true,
			expectedPaths: nil,
		},
		{
			name:          "multiple files, one relevant core exception",
			changedFiles:  []string{"README.md", "caiasset/asset.go", "pkg/cai2hcl/converters/convert_resource.go"},
			expectedRun:   true,
			expectedPaths: nil,
		},
		{
			name:          "test folder direct child",
			changedFiles:  []string{"test/assert_test_files.go"},
			expectedRun:   true,
			expectedPaths: nil,
		},
		{
			name:          "mix of test services and test folder direct child",
			changedFiles:  []string{"test/services/alloydb/alloydb_cluster_generated_test.go", "test/assert_test_files.go"},
			expectedRun:   true,
			expectedPaths: nil,
		},
		{
			name:          "multiple test service files",
			changedFiles:  []string{"test/services/alloydb/alloydb_cluster_generated_test.go", "test/services/apigee/apigee_test.go"},
			expectedRun:   true,
			expectedPaths: []string{"./test/services/alloydb", "./test/services/apigee"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actualRun, actualPaths := shouldRunTests(tc.changedFiles)
			if actualRun != tc.expectedRun {
				t.Errorf("expected run: %v, got %v", tc.expectedRun, actualRun)
			}
			sort.Strings(actualPaths)
			sort.Strings(tc.expectedPaths)
			if !reflect.DeepEqual(actualPaths, tc.expectedPaths) {
				t.Errorf("expected paths: %v, got %v", tc.expectedPaths, actualPaths)
			}
		})
	}
}
