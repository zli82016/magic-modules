---
trigger: always_on
description: Always-on system prompt for TGC development
---

# TGC development Rules

As an AI agent operating in this repository, you must **ALWAYS** follow these steps before attempting to add a new resource/field to TGC:

1. In the magic-modules repository, don't run command `go test` or `go mod tidy`.

2. In the downstream TGC repository, don't run command `go test`.


3. To fix the failed TGC integration tests
   - **don't** modify the templates in `mmv1/templates/terraform`. It is allowed to modify the templates in `mmv1/templates/tgc_next`.
   - **don't** add ignore_read_extra to example in Resource.yaml
   - **don't** add new fields to mmv1/api/resource/custom_code.go unless it is guided by the user
   - **don't** remove any existing custom_code, including any constants


4. Only commit files under the `mmv1` folder in the branch, and exclude scratch files like `.txt`, `.py`, and `.sh` from commits.

5. DO NOT make changes directly in the downstream repository (`terraform-google-conversion`). All changes must be driven through Magic Modules (`mmv1/`).

6. You must strictly follow the sequence of phases defined in `GEMINI.md` (Session Setup -> Implementation -> Unit Testing -> Integration Testing). Code generation (Phase 2) MUST be performed before unit tests (Phase 3), and unit tests MUST be performed before integration tests (Phase 5). Structure your `task.md` to reflect these phases.

7. For any failure (build, unit test, integration test, or verification), stop and report the error with detailed logs. Analyze the cause and provide a solution instead of attempting automatic fixes.