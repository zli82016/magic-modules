package resource

func FilterTerraformAttributionLabel(raw interface{}) map[string]interface{} {
	if raw == nil {
		return nil
	}
	labels := raw.(map[string]interface{})
	delete(labels, "goog-terraform-provisioned")
	return labels
}
