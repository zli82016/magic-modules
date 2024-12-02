package hcl

import (
	"github.com/zclconf/go-cty/cty"
)

// HCLResourceBlock identifies the HCL block's labels and content.
type HCLResourceBlock struct {
	Labels []string
	Value  cty.Value
}