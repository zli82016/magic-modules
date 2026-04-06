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
