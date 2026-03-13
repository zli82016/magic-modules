package compute

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceComputeRegionSecurityPolicySpecRulesDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	return false
}
