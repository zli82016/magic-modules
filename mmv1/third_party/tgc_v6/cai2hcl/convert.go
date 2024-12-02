package cai2hcl

import (
	"fmt"

	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/cai2hcl/converters"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/cai2hcl/hcl"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/caiasset"
	"go.uber.org/zap"
)

// Struct for options so that adding new options does not
// require updating function signatures all along the pipe.
type Options struct {
	ErrorLogger *zap.Logger
}

// Converts CAI Assets into HCL string.
func Convert(assets []*caiasset.Asset, options *Options) ([]byte, error) {
	if options == nil || options.ErrorLogger == nil {
		return nil, fmt.Errorf("logger is not initialized")
	}

	allBlocks := []*hcl.HCLResourceBlock{}
	for _, asset := range assets {
		converter, ok := converters.ConverterMap[asset.Type]
		if !ok {
			continue
		}
		newBlocks, err := converter.Convert(asset)
		if err != nil {
			return nil, err
		}

		allBlocks = append(allBlocks, newBlocks...)
	}

	t, err := hcl.HclWriteBlocks(allBlocks)

	options.ErrorLogger.Debug(string(t))

	return t, err
}
