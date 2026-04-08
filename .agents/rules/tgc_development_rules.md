---
trigger: always_on
description: Always-on system prompt for TGC development
---

# TGC development Rules

As an AI agent operating in this repository, you must **ALWAYS** follow these steps before attempting to add a new resource/field to TGC:

1. In the magic-modules repository, don't run command `go test` or `go mod tidy`.

2. In the downstream TGC repository, don't run command `go test`. Instead, run `make test-integration-local` for integration tests and `make test-local` for unit tests.


3. To fix the failed TGC integration tests
   - **don't** modify the templates in `mmv1/templates/terraform`. It is allowed to modify the templates in `mmv1/templates/tgc_next`.
   - **don't** add ignore_read_extra to example in Resource.yaml
   - **don't** add new fields to mmv1/api/resource/custom_code.go unless it is guided by the user
   - **don't** remove any existing custom_code, including any constants

4. When running `make tgc` or related generators in magic-modules, always use an explicit `OUTPUT_PATH` specify the downstream GoogleCloudPlatform repository (e.g., `OUTPUT_PATH=$GOPATH/src/github.com/GoogleCloudPlatform/terraform-google-conversion`). This prevents the system from defaulting to root `/tfplan2cai`.

5. When running `tgc-run-integration-tests-skill` or manual integration tests, always set `WRITE_FILES=true` to ensure the framework writes out diagnostic fixtures for comparisons.

6. Only commit files under the `mmv1` folder in the branch, and exclude scratch files like `.txt`, `.py`, and `.sh` from commits.

7. DO NOT make changes directly in the downstream repository (`terraform-google-conversion`). All changes must be driven through Magic Modules (`mmv1/`).

8. You must strictly follow the sequence of phases defined in `GEMINI.md` (Session Setup -> Implementation -> Unit Testing -> CAI Verification -> Integration Testing). Code generation (Phase 2) MUST be performed before unit tests (Phase 3), and unit tests MUST be performed before integration tests (Phase 5). Structure your `task.md` to reflect these phases.

9. For any failure (build, unit test, integration test, or verification), stop and report the error with detailed logs. Analyze the cause and provide a solution instead of attempting automatic fixes.