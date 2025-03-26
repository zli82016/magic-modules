package resourcemanager_test

import (
	"testing"

	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/test"
)

func TestAccProject_labels(t *testing.T) {
	test.AssertTestFile(
		t,
		"TestAccProject_labels",
		"google_project",
		[]string{
			"billing_account",
			"auto_create_network",
			"deletion_policy",
			"tags",
		},
	)
}
func TestAccProject_parentFolder(t *testing.T) {
	test.AssertTestFile(
		t,
		"TestAccProject_parentFolder",
		"google_project",
		[]string{
			"billing_account",
			"auto_create_network",
			"deletion_policy",
			"tags",
		},
	)
}
