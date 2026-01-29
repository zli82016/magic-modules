package provider

import "sort"

func ListSupportedTerraformResources() []string {
	resources := make([]string, 0, len(generatedResources)+len(handwrittenResources))
	for k := range generatedResources {
		resources = append(resources, k)
	}
	for k := range handwrittenResources {
		resources = append(resources, k)
	}
	sort.Strings(resources)
	return resources
}
