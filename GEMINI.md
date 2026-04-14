# TGC Resource Addition Main Loop

This document defines the main loop for adding a new resource to TGC or fixing a resource conversion.

## The Main Loop

The workflow consists of the following phases, orchestrated by the Parent Agent:

## Required Skills
Before proceeding with the workflow, ensure you are familiar with and read the following skills when prompted in the phases:
- `sync-provider` (Phase 1)
- `implementing-and-triaging-tgc-resources` (Phase 2)
- `fixing-tgc-resource-or-test-failures` (Phase 6)

## Required Subagents
- None

### 1. Session Setup
- **Set Environment**: Ensure `TGC_DIR` environment variable is set to the absolute path of your active TGC downstream workspace. On macOS, you may also need to ensure Go and Terraform are available in your PATH (e.g., `/usr/local/go/bin` for Go and `/opt/homebrew/bin` for Terraform on Apple Silicon).
  ```bash
  export TGC_DIR=/path/to/downstream/workspace
  export PATH=/usr/local/go/bin:/opt/homebrew/bin:$PATH
  ```

- **Use Skill**: Read and follow `sync-provider` skill to synchronize the downstream repository with Magic Modules.

### 2. Implementation (Parent Agent)
- **Read Skill**: Read `implementing-and-triaging-tgc-resources` to identify the task and get guidance on implementation or fixes.

### 3. Generate Code
- **Generate Code**: Use the automation script `./.agents/skills/tgc-build-skill/scripts/build_tgc.sh` to project changes to the downstream repository.
- If build or dependency errors occur, stop and immediately report the error in the conversation using the following template:
 - **Failed Command**: `[The command that failed]`
 - **Detailed Logs**:
   ```
   [Paste the full, relevant error logs here]
   ```
 - **Analysis**: `[Analyze the cause of the failure]`
 - **Proposed Solution**: `[Outline the solution and ask for user approval before applying it]`
- **Troubleshooting**: If failure is due to missing dependencies in shared templates, refer to "Troubleshooting Build Failures" in `tgc-add-new-generated-resource-skill` for solutions like using `IsTgcCompiler` guards.

### 4. Unit Testing
- Run the following command for folders pkg and test. **Do NOT scope to specific services or tests; all unit tests must be run:**
  ```bash
  make test-local TEST=./test
  make test-local TEST=./pkg/...
  ```

### 5. Integration Testing
- Identify the target test name and its specific service directory. 
   - *Example*: Target `TestAccAlloydbBackup` located in `./test/services/alloydb`.
2. Run the test using the script from the `scripts` file, passing the test path and test name:
   ```bash
   .agents/skills/tgc-run-integration-tests-skill/scripts/run_integration_test.sh <test-path> <test-name>
   ```
   **Example**:
   ```bash
   .agents/skills/tgc-run-integration-tests-skill/scripts/run_integration_test.sh ./test/services/alloydb TestAccAlloydbBackup
   ```

### 6. Fix (Parent Agent)
- **Read Skill**: Read `fixing-tgc-resource-or-test-failures` to report and fix failures.
- Apply fixes in MMv1 according to triage rules and troubleshooting playbooks.
> [!IMPORTANT]
> **After ANY fix applied in Step 6, you MUST repeat the full verification loop:**
> 2. **Step 3 (Generate Code)**: Generate code.
> 3. **Step 4 (Unit Testing)**: Run unit tests.
> 4. **Step 5 (Integration Testing)**: Run integration tests.
> Do not skip any of these steps to ensure no new regressions are introduced.

### 7. Finalization
- Ask the user if the task is complete and if you should proceed with committing.
- Commit changes under `mmv1/` folder only.
- Exclude scratch files from commits.
