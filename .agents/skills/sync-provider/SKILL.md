---
name: sync-provider
description: "Synchronize a downstream Terraform provider repository with Magic Modules by aligning commit history and verifying parity."
---

# `sync-provider`

> **Note to AI Agents:** You MUST read the YAML frontmatter above first. Only read the rest of this file if the `description` matches your required task.
> This skill is designed to be completely self-contained and unambiguous for a fresh agent without prior context.

## Prerequisites

- You must be operating relative to the `magic-modules` and downstream provider repositories.
- You must have the absolute path to the downstream repository.
- You must have verified there are no unsaved or uncommitted changes in the downstream provider directory that would be overwritten by code generation.

## Execution Steps

### 1. Update Magic Modules to Latest
Ensure you are on the correct branch and pull the latest changes:
```bash
git checkout <branch-name>
git pull origin <branch-name>
```

### 2. Clean and Update Downstream to Latest
Downstream is generated, so clean local changes first to avoid conflicts, then pull latest:
```bash
cd <downstream-provider-path>
git reset --hard
git clean -fd
git checkout main
git pull origin main
```

### 3. Project Latest Changes
Return to `magic-modules` and run the build script to project all changes to the latest downstream state:
```bash
./.agents/skills/tgc-build-skill/scripts/build_tgc.sh <downstream-provider-path>
```