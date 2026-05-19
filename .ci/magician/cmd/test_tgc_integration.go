package cmd

import (
	"fmt"
	"magician/exec"
	"magician/github"
	"magician/source"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var testTGCIntegrationCmd = &cobra.Command{
	Use:   "test-tgc-integration",
	Short: "Run tgc integration tests via workflow dispatch",
	Long: `This command runs tgc unit tests via workflow dispatch

	The following PR details are expected as environment variables:
	1. GOPATH
	2. GITHUB_TOKEN_MAGIC_MODULES
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		goPath, ok := os.LookupEnv("GOPATH")
		if !ok {
			return fmt.Errorf("did not provide GOPATH environment variable")
		}

		githubToken, ok := lookupGithubTokenOrFallback("GITHUB_TOKEN_MAGIC_MODULES")
		if !ok {
			return fmt.Errorf("did not provide GITHUB_TOKEN_MAGIC_MODULES or GITHUB_TOKEN environment variables")
		}

		rnr, err := exec.NewRunner()
		if err != nil {
			return fmt.Errorf("error creating runner: %w", err)
		}

		ctlr := source.NewController(goPath, "modular-magician", githubToken, rnr)

		gh := github.NewClient(githubToken)

		return execTestTGCIntegration(args[0], args[1], args[2], args[3], args[4], args[5], "modular-magician", rnr, ctlr, gh)
	},
}

func execTestTGCIntegration(prNumber, mmCommit, buildID, projectID, buildStep, ghRepo, githubUsername string, rnr ExecRunner, ctlr *source.Controller, gh GithubClient) error {
	newBranch := "auto-pr-" + prNumber
	repo := &source.Repo{
		Name:   ghRepo,
		Branch: newBranch,
	}
	ctlr.SetPath(repo)
	if err := ctlr.Clone(repo); err != nil {
		return fmt.Errorf("error cloning repo: %w", err)
	}
	if err := rnr.PushDir(repo.Path); err != nil {
		return fmt.Errorf("error changing to repo dir: %w", err)
	}
	diffs, err := rnr.Run("git", []string{"diff", "--name-only", "HEAD~1"}, nil)
	if err != nil {
		return fmt.Errorf("error diffing repo: %w", err)
	}

	// Convert the raw diff string into a slice of strings
	changedFiles := strings.Split(strings.TrimSpace(diffs), "\n")

	runTests, testPaths := shouldRunTests(changedFiles)
	if !runTests {
		fmt.Println("Skipping tests: No relevant go files changed")
		return nil
	}

	fmt.Println("Running tests: Relevant go files changed!")

	targetURL := fmt.Sprintf("https://console.cloud.google.com/cloud-build/builds;region=global/%s;step=%s?project=%s", buildID, buildStep, projectID)
	if err := gh.PostBuildStatus(prNumber, ghRepo+"-test-integration", "pending", targetURL, mmCommit); err != nil {
		return fmt.Errorf("error posting build status: %w", err)
	}

	if _, err := rnr.Run("go", []string{"mod", "edit", "-replace", fmt.Sprintf("github.com/hashicorp/terraform-provider-google-beta=github.com/%s/terraform-provider-google-beta@%s", githubUsername, newBranch)}, nil); err != nil {
		fmt.Println("Error running go mod edit: ", err)
	}
	if _, err := rnr.Run("go", []string{"mod", "tidy"}, nil); err != nil {
		fmt.Println("Error running go mod tidy: ", err)
	}

	if _, err := rnr.Run("make", []string{"build"}, nil); err != nil {
		fmt.Println("Error running make build: ", err)
	}
	state := "success"
	makeArgs := []string{"test-integration"}
	if len(testPaths) > 0 {
		makeArgs = append(makeArgs, fmt.Sprintf("TESTPATH=%s", strings.Join(testPaths, " ")))
	}
	if _, err := rnr.Run("make", makeArgs, nil); err != nil {
		fmt.Println("Error running make test-integration: ", err)
		state = "failure"
	}

	if err := gh.PostBuildStatus(prNumber, ghRepo+"-test-integration", state, targetURL, mmCommit); err != nil {
		return fmt.Errorf("error posting build status: %w", err)
	}
	return nil
}

func shouldRunTests(changedFiles []string) (bool, []string) {
	runTests := false
	runAllTests := false
	servicePaths := make(map[string]bool)

	for _, file := range changedFiles {
		fmt.Println("current file:", file)
		if !strings.HasSuffix(file, ".go") {
			continue
		}

		// Handle pkg/services/ and its exceptions
		if strings.HasPrefix(file, "pkg/") {
			if strings.HasPrefix(file, "pkg/cai2hcl/") || strings.HasPrefix(file, "pkg/tfplan2cai/") || strings.HasPrefix(file, "pkg/caiasset/") {
				runTests = true
				runAllTests = true
			}
			continue
		}

		// Skip the fully ignored directories
		if strings.HasPrefix(file, "cai2hcl/") || strings.HasPrefix(file, "caiasset/") || strings.HasPrefix(file, "tfplan2cai/") {
			continue
		}

		// If a .go file makes it this far, it needs testing
		runTests = true

		if strings.HasPrefix(file, "test/") {
			if strings.HasPrefix(file, "test/services/") {
				parts := strings.Split(file, "/")
				if len(parts) >= 3 {
					servicePaths["./test/services/"+parts[2]] = true
				} else {
					runAllTests = true
				}
			} else if !strings.Contains(strings.TrimPrefix(file, "test/"), "/") {
				// Direct child file in test/ folder
				runAllTests = true
			} else {
				runAllTests = true
			}
		} else {
			runAllTests = true
		}
	}

	if !runTests {
		return false, nil
	}
	if runAllTests || len(servicePaths) == 0 {
		return true, nil
	}

	var paths []string
	for path := range servicePaths {
		paths = append(paths, path)
	}
	return true, paths
}

func init() {
	rootCmd.AddCommand(testTGCIntegrationCmd)
}
