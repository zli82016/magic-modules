package test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/cai2hcl"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/cai2hcl/converters/utils"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/caiasset"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/tfplan2cai"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	hcl "github.com/joselitofilho/hcl-parser-go/pkg/parser/hcl"
)

type TestMetadata struct {
	Test      string
	RawConfig string
	Service   string
	Resource  string
	AssetName string
	Asset     caiasset.Asset
}

type _TestCase struct {
	name     string
	resource string
}

func AssertTestFile(t *testing.T, fileName, resource string, excludedFields []string) {
	GlobalSetup()

	c := _TestCase{
		name:     fileName,
		resource: resource,
	}

	t.Run(c.name, func(t *testing.T) {
		t.Parallel()

		err := assertTestData(t, c, excludedFields)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func assertTestData(t *testing.T, c _TestCase, excludedFields []string) error {
	fileName := c.name

	// Create a temporary directory for running terraform.
	tfDir, err := os.MkdirTemp(tmpDir, "terraform")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tfDir)

	testMetadata := TestConfig[c.name]
	exportAsset := testMetadata.Asset

	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	exportAssets := []caiasset.Asset{exportAsset}
	exportConfigData, err := cai2hcl.Convert(exportAssets, &cai2hcl.Options{
		ErrorLogger: logger,
	})
	if err != nil {
		return err
	}

	exportFileName := fmt.Sprintf("%s_export", fileName)
	exportTfFile := fmt.Sprintf("%s.tf", exportFileName)
	exportTfFilePath := fmt.Sprintf("%s/%s", tfDir, exportTfFile)
	err = os.WriteFile(exportTfFilePath, exportConfigData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %#v", exportTfFilePath, err)
	}

	exportConfigMap, err := getConfig(exportTfFilePath, c.resource)
	if err != nil {
		return err
	}

	rawTfFile := fmt.Sprintf("%s.tf", fileName)
	err = os.WriteFile(rawTfFile, []byte(testMetadata.RawConfig), 0644)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %#v", rawTfFile, err)
	}
	rawConfigMap, err := getConfig(rawTfFile, c.resource)
	if err != nil {
		return err
	}
	if len(rawConfigMap) == 0 {
		return fmt.Errorf("raw config for test %s is unavailable", c.name)
	}

	excludedFieldMap := make(map[string]bool, 0)
	for _, f := range excludedFields {
		excludedFieldMap[f] = true
	}

	for address, rawConfig := range rawConfigMap {
		if exportConfig, ok := exportConfigMap[address]; !ok {
			return fmt.Errorf("%s - Missing resource after cai2hcl conversion: %s.", t.Name(), address)
		} else {
			missingKeys := compareHCLFields(rawConfig.(map[string]interface{}), exportConfig.(map[string]interface{}), "", excludedFieldMap)
			if len(missingKeys) > 0 {
				return fmt.Errorf("%s - Missing fields in resource %s after cai2hcl conversion:\n%s", t.Name(), address, missingKeys)
			}
		}
	}

	// Get the ancestry cache for tfplan2cai conversion
	ancestors := exportAsset.Ancestors
	ancestryCache := make(map[string]string, 0)
	if len(ancestors) != 0 {
		var path string
		for i := len(ancestors) - 1; i >= 0; i-- {
			curr := ancestors[i]
			if path == "" {
				path = curr
			} else {
				path = fmt.Sprintf("%s/%s", path, curr)
			}
		}
		ancestryCache[ancestors[0]] = path

		project := utils.ParseFieldValue(exportAsset.Name, "projects")
		projectKey := fmt.Sprintf("projects/%s", project)
		if strings.HasPrefix(ancestors[0], "projects") && ancestors[0] != projectKey {
			ancestryCache[projectKey] = path
		}
	}

	// Convert the export config to roundtrip assets and then convert the roundtrip assets back to roundtrip config
	roundtripConfigData, err := getRoundtripConfig(t, exportFileName, tfDir, ancestryCache)
	if err != nil {
		return err
	}

	roundtripFileName := fmt.Sprintf("%s_roundtrip", fileName)
	roundtripTfFilePath := fmt.Sprintf("%s.tf", roundtripFileName)
	err = os.WriteFile(roundtripTfFilePath, roundtripConfigData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %#v", roundtripTfFilePath, err)
	}
	roundtripConfigMap, err := getConfig(roundtripTfFilePath, c.resource)
	if err != nil {
		return err
	}

	for address, roundtripConfig := range roundtripConfigMap {
		if exportConfig, ok := exportConfigMap[address]; !ok {
			return fmt.Errorf("%s - Missing resource after roundtrip conversion: %s.", t.Name(), address)
		} else {
			missingKeys := compareHCLFields(roundtripConfig.(map[string]interface{}), exportConfig.(map[string]interface{}), "", excludedFieldMap)
			if len(missingKeys) > 0 {
				return fmt.Errorf("%s - Missing fields in resource %s after roundtrip conversion:\n%s", t.Name(), address, missingKeys)
			}
		}
	}

	return nil
}

func getConfig(filePath, target string) (map[string]interface{}, error) {
	files := []string{filePath}

	// Parse Terraform configurations
	config, err := hcl.Parse([]string{}, files)
	if err != nil {
		return nil, err
	}

	configMap := make(map[string]interface{}, 0)
	for _, r := range config.Resources {
		if r.Type != target {
			continue
		}
		addr := fmt.Sprintf("%s.%s", r.Type, r.Name)
		configMap[addr] = r.Attributes
	}
	return configMap, nil
}

// Compares HCL and finds all of the keys in map1 are in map2
func compareHCLFields(map1, map2 map[string]interface{}, path string, excludedFields map[string]bool) []string {
	var missingKeys []string
	for key, value1 := range map1 {
		if value1 == nil {
			continue
		}

		currentPath := path + "." + key
		if path == "" {
			currentPath = key
		}

		if excludedFields[currentPath] {
			continue
		}

		value2, ok := map2[key]
		if !ok || value2 == nil {
			missingKeys = append(missingKeys, currentPath)
			continue
		}

		switch v1 := value1.(type) {
		case map[string]interface{}:
			v2, _ := value2.(map[string]interface{})
			// if !ok {
			// 	fmt.Printf("Type mismatch for key: %s\n", currentPath)
			// 	continue
			// }
			missingKeys = append(missingKeys, compareHCLFields(v1, v2, currentPath, excludedFields)...)
		case []interface{}:
			v2, _ := value2.([]interface{})
			// if !ok {
			// 	fmt.Printf("Type mismatch for key: %s\n", currentPath)
			// 	continue
			// }
			// if len(v1) != len(v2) {
			// 	fmt.Printf("List length mismatch for key: %s\n", currentPath)
			// 	continue
			// }

			for i := 0; i < len(v1); i++ {
				nestedMap1, ok1 := v1[i].(map[string]interface{})
				nestedMap2, ok2 := v2[i].(map[string]interface{})
				if ok1 && ok2 {
					keys := compareHCLFields(nestedMap1, nestedMap2, fmt.Sprintf("%s[%d]", currentPath, i), excludedFields)
					missingKeys = append(missingKeys, keys...)
				}
			}
		default:
		}
	}

	return missingKeys
}

func getRoundtripConfig(t *testing.T, fileName, tfDir string, ancestryCache map[string]string) ([]byte, error) {
	// Run terraform init and terraform apply to generate tfplan.json files
	terraformWorkflow(t, tfDir, fileName)

	planFile := fmt.Sprintf("%s.tfplan.json", fileName)
	planfilePath := filepath.Join(tfDir, planFile)
	jsonPlan, err := os.ReadFile(planfilePath)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	roundtripAssets, err := tfplan2cai.Convert(ctx, jsonPlan, &tfplan2cai.Options{
		ErrorLogger:    zaptest.NewLogger(t),
		Offline:        true,
		DefaultProject: "ci-test-project-188019",
		DefaultRegion:  "",
		DefaultZone:    "",
		UserAgent:      "",
		AncestryCache:  ancestryCache,
	})

	if err != nil {
		return nil, err
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	data, err := cai2hcl.Convert(roundtripAssets, &cai2hcl.Options{
		ErrorLogger: logger,
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}
