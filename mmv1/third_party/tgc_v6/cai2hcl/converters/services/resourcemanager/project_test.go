package resourcemanager_test

import (
	"testing"

	cai2hclTesting "github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/cai2hcl/testing"
)

func TestComputeInstance(t *testing.T) {
	cai2hclTesting.AssertTestFiles(
		t,
		"./testdata",
		[]string{
			"TestAccProject_create",
		})
}
