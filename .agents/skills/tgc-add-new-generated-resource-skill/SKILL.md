---
name: tgc-add-new-generated-resource-skill
description: Add a new generated resource to TGC. Use when you need to add a new generated resource to TGC.
---

# tgc-add-new-generated-resource-skill

When you need to add a new generated resource to TGC, use this skill.

## When to Use This Skill

- Use this when adding a new generated resource to TGC.
- This is helpful when you need to understand the structural steps and configurations needed to expose a generated resource to the Terraform Google Conversion (TGC) library.

---

## How to Use It

If you added or modified a generated resource, follow the steps below carefully.

### 1. Map and Enable

- **Mapping**: Use a script or command to locate `mmv1/products/.../Resource.yaml` for each `google_` type.
  - **Search pattern**: Look for `name: 'resource_name'` or perform a filename match.
- **Enabling**: For each found YAML file:
  - Ensure `include_in_tgc_next: true` is present at the **top-level**.
  - Place it in the proper order according to the progression of fields in the `mmv1/api/resource.go` file.

### 2. Check for URL Parameters and Asset Name Format

- If the resource has parameters marked as `url_param_only: true` and `required: true`, verify if they can be extracted from the CAI asset name during `cai2hcl` conversion.
- If the `self_link` in the YAML file is just `{{name}}` or does not contain all the required parameters in its pattern, you MUST specify `cai_asset_name_format` at the top-level to define the pattern for extraction.
- Example: `cai_asset_name_format: 'projects/{{project}}/locations/{{location}}/notificationConfigs/{{config_id}}'`


### Troubleshooting Build Failures

### Missing Package Dependency in Shared Templates
- **Symptom**: `go mod tidy` or compilation fails after generation because a package (e.g., `compute`) is not found in the TGC environment.
- **Cause**: Shared templates in `mmv1/templates/terraform/constants` may contain hardcoded imports or functions relying on packages not available in TGC.
- **Solution**: Wrap the problematic code in the template with a compiler condition to exclude it for TGC generation. You can use the helper method `IsTgcCompiler`:
  ```tmpl
  {{- if not $.ResourceMetadata.ProductMetadata.IsTgcCompiler }}
  // Code to exclude for TGC (only included for standard Terraform provider)
  {{- end }}
  ```
  *Note: The exact path to `IsTgcCompiler` may vary depending on the template's context (e.g., `$.IsTgcCompiler` or `$.ProductMetadata.IsTgcCompiler`).*

### No Tests Generated Failure
- **Symptom**: `Error generating resource tests: No TGC tests for resource <ResourceName>`
- **Cause**: All examples in the YAML file have `exclude_test: true`, and the generator cannot find matching handwritten tests (e.g., if they start with lowercase `t` like `testAcc...` instead of `TestAcc`).
- **Solution**: Explicitly list the expected test name (or subtest name like `ParentTest/SubTest` if it is a subtest) in `tgc_tests` in the resource's YAML file to force generation. See Troubleshooting Playbook Item 11 for more details.
