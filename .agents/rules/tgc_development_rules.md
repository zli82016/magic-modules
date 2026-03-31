---
trigger: always_on
description: Always-on system prompt for TGC development
---

# TGC development Rules

As an AI agent operating in this repository, you must **ALWAYS** follow these steps before attempting to add a new resource/field to TGC:

1. In the magic-modules repository, don't run command `go test` or `go mod tidy`.

2. In the downstream TGC repository, don't run command `go test`. Instead, run `make test-integration-local` for integration tests and `make test-local` for unit tests.

Examples:
make test-integration-local TESTPATH=./test/services/alloydb  TESTARGS='-run=TestAccAlloydbBackup' > alloydbBackup.log


make test-local TEST=./pkg/...

make test-local TEST=./pkg/... TESTARGS='-run=TestConvert_iamBinding'

3. To fix the failed TGC integration tests
   - **don't** modify the templates in `mmv1/templates/terraform`. It is allowed to modify the templates in `mmv1/templates/tgc_next`.
   - **don't** add ignore_read_extra to example in Resource.yaml
   - **don't** add new fields to mmv1/api/resource/custom_code.go unless it is guided by the user
   - **don't** remove any existing custom_code, including any constants

4. When running `make tgc` or related generators in magic-modules, always use an explicit `OUTPUT_PATH` specify the downstream GoogleCloudPlatform repository (e.g., `OUTPUT_PATH=$GOPATH/src/github.com/GoogleCloudPlatform/terraform-google-conversion`). This prevents the system from defaulting to root `/tfplan2cai`.

5. When running `tgc-run-integration-tests-skill` or manual integration tests, always set `WRITE_FILES=true` to ensure the framework writes out diagnostic fixtures for comparisons.