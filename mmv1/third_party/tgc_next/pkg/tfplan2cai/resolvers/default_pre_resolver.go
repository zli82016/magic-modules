package resolvers

import (
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/tfplan2cai/models"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/tfplan2cai/tfplan"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	provider "github.com/hashicorp/terraform-provider-google-beta/google-beta/provider"
)

var ErrDuplicateAsset = errors.New("duplicate asset")

type DefaultPreResolver struct {
	schema *schema.Provider

	// For logging error / status information that doesn't warrant an outright failure
	errorLogger *zap.Logger
}

func NewDefaultPreResolver(errorLogger *zap.Logger) *DefaultPreResolver {
	return &DefaultPreResolver{
		schema:      provider.Provider(),
		errorLogger: errorLogger,
	}
}

func (r *DefaultPreResolver) Resolve(jsonPlan []byte) map[string][]*models.FakeResourceDataWithMeta {
	changes, err := tfplan.ReadResourceChanges(jsonPlan)
	if err != nil {
		return nil
	}

	transformed := tfplan.TransformResourceChanges(changes)
	return r.AddResourceChanges(transformed)
}

// AddResourceChange processes the resource changes in two stages:
// 1. Process deletions (fetching canonical resources from GCP as necessary)
// 2. Process creates, updates, and no-ops (fetching canonical resources from GCP as necessary)
// This will give us a deterministic end result even in cases where for example
// an IAM Binding and Member conflict with each other, but one is replacing the
// other.
func (r *DefaultPreResolver) AddResourceChanges(changes []map[string]interface{}) map[string][]*models.FakeResourceDataWithMeta {
	resourceDataMap := make(map[string][]*models.FakeResourceDataWithMeta, 0)

	for _, rc := range changes {
		var resourceData *models.FakeResourceDataWithMeta
		t := rc["type"].(string)
		address := rc["address"].(string)
		resource := r.schema.ResourcesMap[t]
		// Skip resources not found in the google beta provider's schema
		if _, ok := r.schema.ResourcesMap[t]; !ok {
			r.errorLogger.Debug(fmt.Sprintf("%s: resource t not found in google beta provider: %s.", address, t))
			continue
		}

		resourceData = models.NewFakeResourceDataWithMeta(
			t,
			resource.Schema,
			rc["change"].(map[string]interface{}),
			rc["isDelete"].(bool),
			address,
		)

		// TODO: handle the address of iam resources
		if exist := resourceDataMap[address]; exist == nil {
			resourceDataMap[address] = make([]*models.FakeResourceDataWithMeta, 0)
		}
		resourceDataMap[address] = append(resourceDataMap[address], resourceData)
	}

	return resourceDataMap
}
