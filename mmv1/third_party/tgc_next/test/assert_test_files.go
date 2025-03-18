package test

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/cai2hcl"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/caiasset"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/tfplan2cai/tfplan"
	"go.uber.org/zap"
)

type _TestCase struct {
	name         string
	sourceFolder string
}

func AssertTestFiles(t *testing.T, folder string, fileNames, excludedFields []string) {
	cases := []_TestCase{}

	for _, name := range fileNames {
		cases = append(cases, _TestCase{name: name, sourceFolder: folder})
	}

	for i := range cases {
		c := cases[i]

		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := assertTestData(t, c, excludedFields)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func assertTestData(t *testing.T, c _TestCase, excludedFields []string) error {
	fileName := c.name
	sourceFolder := c.sourceFolder

	rawChangeMap, err := getRawConfig(t, fileName, sourceFolder)
	if err != nil {
		return err
	}

	prettyJSON, err := json.MarshalIndent(rawChangeMap, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling pretty JSON: %v", err)
	}
	fmt.Println("raw")
	fmt.Println(string(prettyJSON))

	exportConfig, err := convertExportConfig(fileName, sourceFolder)
	if err != nil {
		return err
	}
	exportFileName := fmt.Sprintf("%s_export", fileName)
	exportTfFile := fmt.Sprintf("%s.tf", exportFileName)
	exportTfFilePath := fmt.Sprintf("%s/%s", sourceFolder, exportTfFile)
	err = os.WriteFile(exportTfFilePath, exportConfig, 0644)
	if err != nil {
		return err
	}
	exportChangeMap, err := getRawConfig(t, exportFileName, sourceFolder)
	if err != nil {
		return err
	}

	prettyJSON1, err := json.MarshalIndent(exportChangeMap, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling pretty JSON: %v", err)
	}

	fmt.Println("export")
	fmt.Println(string(prettyJSON1))

	excludedFieldMap := make(map[string]bool, 0)
	for _, f := range excludedFields {
		excludedFieldMap[f] = true
	}

	for address, rawConfig := range rawChangeMap {
		if exportConfig, ok := exportChangeMap[address]; !ok {
			t.Errorf("%s - Missing resource after cai2hcl conversion: %s.", t.Name(), address)
		} else {
			missingKeys := findMissingKeys(rawConfig.(map[string]interface{}), exportConfig.(map[string]interface{}), "", excludedFieldMap)
			if len(missingKeys) > 0 {
				t.Errorf("%s - Missing fields in resource %s after cai2hcl conversion:\n%s", t.Name(), address, missingKeys)
			}
		}
	}
	// t.Errorf("zhenhuatest")
	return nil
}

func convertExportConfig(fileName, sourceFolder string) ([]byte, error) {
	assetFilePath := fmt.Sprintf("%s/%s.json", sourceFolder, fileName)
	assetPayload, err := os.ReadFile(assetFilePath)
	if err != nil {
		return nil, err
	}
	var exportAsset *caiasset.Asset
	if err := json.Unmarshal(assetPayload, &exportAsset); err != nil {
		return nil, err
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	exportAssets := []*caiasset.Asset{exportAsset}
	data, err := cai2hcl.Convert(exportAssets, &cai2hcl.Options{
		ErrorLogger: logger,
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func getRawConfig(t *testing.T, fileName, sourceFolder string) (map[string]interface{}, error) {
	// Create a temporary directory for running terraform.
	tfDir, err := os.MkdirTemp(tmpDir, "terraform")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tfDir)

	tfFile := fmt.Sprintf("%s.tf", fileName)
	generateTestFiles(t, sourceFolder, tfDir, tfFile)

	// Run terraform init and terraform apply to generate tfplan.json files
	terraformWorkflow(t, tfDir, fileName)

	planFile := fmt.Sprintf("%s.tfplan.json", fileName)
	planfilePath := filepath.Join(tfDir, planFile)
	jsonPlan, err := os.ReadFile(planfilePath)
	if err != nil {
		return nil, err
	}

	changes, err := tfplan.ReadResourceChanges(jsonPlan)
	if err != nil {
		return nil, err
	}

	transformed := tfplan.TransformResourceChanges(changes)
	changeMap := make(map[string]interface{}, 0)
	for _, t := range transformed {
		changeMap[t["address"].(string)] = t["change"]
	}
	return changeMap, nil
}

// Finds the all of the keys in map1 are in map2
func findMissingKeys(map1, map2 map[string]interface{}, path string, excludedFields map[string]bool) []string {
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
		if strings.Contains(key, "billing_account") {
			fmt.Printf(" currentPath %s value1 %s value2 %s  is nil %t", currentPath, value1, value2, value2 == nil)
		}
		// fmt.Printf(" currentPath %s value1 %s value2 %s", currentPath, value1, value2)
		fmt.Println()
		switch v1 := value1.(type) {
		case map[string]interface{}:
			v2, _ := value2.(map[string]interface{})
			// if !ok {
			// 	fmt.Printf("Type mismatch for key: %s\n", currentPath)
			// 	continue
			// }
			missingKeys = append(missingKeys, findMissingKeys(v1, v2, currentPath, excludedFields)...)
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
					keys := findMissingKeys(nestedMap1, nestedMap2, fmt.Sprintf("%s[%d]", currentPath, i), excludedFields)
					missingKeys = append(missingKeys, keys...)
				}
			}
		default:
		}
	}

	fmt.Printf("missing keys %#v", missingKeys)
	return missingKeys
}
