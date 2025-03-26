// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package tfplan

import (
	"fmt"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
)

func IsCreate(rc *tfjson.ResourceChange) bool {
	return len(rc.Change.Actions) == 1 && rc.Change.Actions[0] == "create"
}

func IsUpdate(rc *tfjson.ResourceChange) bool {
	return len(rc.Change.Actions) == 1 && rc.Change.Actions[0] == "update"
}

func IsDeleteCreate(rc *tfjson.ResourceChange) bool {
	return len(rc.Change.Actions) == 2 && rc.Change.Actions[0] == "delete"
}

func IsDelete(rc *tfjson.ResourceChange) bool {
	return len(rc.Change.Actions) == 1 && rc.Change.Actions[0] == "delete"
}

func IsNoOp(rc *tfjson.ResourceChange) bool {
	return rc.Change.Actions.NoOp()
}

// ReadResourceChanges returns the list of resource changes from a json plan
func ReadResourceChanges(data []byte) ([]*tfjson.ResourceChange, error) {
	plan := tfjson.Plan{}
	err := plan.UnmarshalJSON(data)
	if err != nil {
		return nil, fmt.Errorf("reading JSON plan: %w", err)
	}

	err = plan.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating JSON plan: %w", err)
	}

	return plan.ResourceChanges, nil
}

func TransformResourceChanges(changes []*tfjson.ResourceChange) []map[string]interface{} {
	resourceChanges := make([]map[string]interface{}, 0)
	for _, rc := range changes {
		// Silently skip non-google resources
		if !strings.HasPrefix(rc.Type, "google_") {
			continue
		}

		if IsCreate(rc) || IsUpdate(rc) || IsDeleteCreate(rc) {
			rChange := map[string]interface{}{
				"type":     rc.Type,
				"change":   rc.Change.After,
				"isDelete": false,
				"address":  rc.Address,
			}
			resourceChanges = append(resourceChanges, rChange)
		} else if IsDelete(rc) {
			rChange := map[string]interface{}{
				"type":     rc.Type,
				"change":   rc.Change.Before,
				"isDelete": true,
				"address":  rc.Address,
			}
			resourceChanges = append(resourceChanges, rChange)
		} else {
			continue
		}
	}

	return resourceChanges
}
