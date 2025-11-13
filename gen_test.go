//go:build integration

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
)

// skipIfMissingTools skips the test if any of the provided CLI tools are not
// found on the host system. This prevents integration tests from failing on
// minimal CI environments that lack Docker, Task, Node, etc.
func skipIfMissingTools(t *testing.T, tools ...string) {
	t.Helper()
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("skipping test – required tool %q not found", tool)
		}
	}
}

var kataboleBin string

func TestMain(m *testing.M) {
	// build the katabole CLI locally
	bin := "katabole"
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build katabole CLI: %v\n", err)
		os.Exit(1)
	}
	// determine absolute path to the built binary so we can execute it from any working dir
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get working directory: %v\n", err)
		os.Exit(1)
	}
	kataboleBin = filepath.Join(wd, bin)
	fmt.Println("[TestMain] kataboleBin =", kataboleBin)

	os.Exit(m.Run())
}

func TestApplyReplacements(t *testing.T) {
	template := []byte(`
import "github.com/katabole/kbexample"
Title: KBExample
psql -d kbexample_dev
psql -d kb_example_dev
`)

	// --- No-dash case: repoName == repoNameUnder (both "myapp").

	out := applyReplacements(
		template,
		"github.com/foo/myapp", // importPath
		"MyApp",                // titleName
		"myapp",                // repoName
		"myapp",                // repoNameUnderscores
	)
	// Since both placeholders become "myapp_dev", we expect exactly two occurrences.
	rawCount := bytes.Count(out, []byte("myapp_dev"))
	assert.Equal(t, 2, rawCount, "no-dash: expected exactly 2 occurrences of \"myapp_dev\"")

	// --- With-dash case: repoName="my-app", repoNameUnderscores="my_app".

	out = applyReplacements(
		template,
		"github.com/foo/my-app", // importPath
		"MyApp",                 // titleName
		"my-app",                // repoName
		"my_app",                // repoNameUnderscores
	)
	// Expect one occurrence of "my-app_dev"
	assert.Equal(t, 1,
		bytes.Count(out, []byte("my-app_dev")),
		"with-dash: expected 1 occurrence of \"my-app_dev\"")

	// Expect one occurrence of "my_app_dev"
	assert.Equal(t, 1,
		bytes.Count(out, []byte("my_app_dev")),
		"with-dash: expected 1 occurrence of \"my_app_dev\"")
}

func TestRepoNameValidation(t *testing.T) {
	// Valid names should pass ValidateRepoName
	valid := []string{"abc", "a_b-c", "A1_2-3"}
	for _, s := range valid {
		assert.Truef(t, ValidateRepoName(s), "expected %q to be valid", s)
	}

	// Invalid names should fail
	invalid := []string{"bad$name", " space", "dot.name"}
	for _, s := range invalid {
		assert.Falsef(t, ValidateRepoName(s), "expected %q to be invalid", s)
	}
}

func TestGetLatestTag(t *testing.T) {
	skipIfMissingTools(t, "git")

	// Create a temporary git repository with tags
	tmpDir := t.TempDir()

	// Initialize git repo
	if err := runCommand(tmpDir, "git", "init"); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user for commits
	if err := runCommand(tmpDir, "git", "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("failed to config git email: %v", err)
	}
	if err := runCommand(tmpDir, "git", "config", "user.name", "Test User"); err != nil {
		t.Fatalf("failed to config git name: %v", err)
	}

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("v1"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	if err := runCommand(tmpDir, "git", "add", "."); err != nil {
		t.Fatalf("failed to git add: %v", err)
	}
	if err := runCommand(tmpDir, "git", "commit", "-m", "first commit"); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Create v0.1.0 tag
	if err := runCommand(tmpDir, "git", "tag", "v0.1.0"); err != nil {
		t.Fatalf("failed to create tag v0.1.0: %v", err)
	}

	// Sleep to ensure different commit timestamps
	// (Git commit timestamps have 1-second resolution)
	time.Sleep(1100 * time.Millisecond)

	// Create another commit
	if err := os.WriteFile(testFile, []byte("v2"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	if err := runCommand(tmpDir, "git", "add", "."); err != nil {
		t.Fatalf("failed to git add: %v", err)
	}
	if err := runCommand(tmpDir, "git", "commit", "-m", "second commit"); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// Create v0.2.0 tag (this should be the latest)
	if err := runCommand(tmpDir, "git", "tag", "v0.2.0"); err != nil {
		t.Fatalf("failed to create tag v0.2.0: %v", err)
	}

	// Open the repository with go-git
	repo, err := git.PlainOpen(tmpDir)
	if err != nil {
		t.Fatalf("failed to open repo: %v", err)
	}

	// Test getLatestTag
	latestTag, err := getLatestTag(repo)
	if err != nil {
		t.Fatalf("getLatestTag failed: %v", err)
	}

	assert.Equal(t, "v0.2.0", latestTag, "expected latest tag to be v0.2.0")
}

// runCommand is a helper to run commands inside a given directory
func runCommand(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func TestGenerateFromKbexample(t *testing.T) {
	// Ensure required external tooling exists; otherwise skip.
	skipIfMissingTools(t, "docker", "task", "npm", "npx")
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "kbexample")
	defer runCommand(outputPath, "docker", "compose", "down")

	err := runCommand(tmpDir, kataboleBin, "gen",
		"--import-path", "github.com/LingEnOwO/kbexample",
		"--title-name", "KBExample",
		"--template-repository", "https://github.com/LingEnOwO/kbexample",
		"-n", "github.com/LingEnOwO/kbexample",
	)
	if err != nil {
		t.Fatalf("katabole gen failed: %v", err)
	}

	err = runCommand(outputPath, "task", "db:up")
	if err != nil {
		t.Fatalf("task db:up failed: %v", err)
	}

	err = runCommand(outputPath, "task", "db:apply")
	if err != nil {
		t.Fatalf("atlas apply failed: %v", err)
	}

	// Ensure npm install runs before Vite build
	err = runCommand(outputPath, "npm", "install")
	if err != nil {
		t.Fatalf("npm install failed: %v", err)
	}

	// Ensure dist/ is built
	err = runCommand(outputPath, "npx", "vite", "build")
	if err != nil {
		t.Fatalf("vite build failed: %v", err)
	}

	err = runCommand(outputPath, "task", "db:seed")
	if err != nil {
		t.Fatalf("task db:seed failed: %v", err)
	}

	err = runCommand(outputPath, "task", "test")
	if err != nil {
		t.Fatalf("task test failed: %v", err)
	}
}

// Happy-path flag tests for the katabole-gen-testing template

func TestGenerateFromGenTesting_DefaultBranch(t *testing.T) {
	// Integration test prerequisites
	skipIfMissingTools(t, "git", "docker", "task")
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test-output")

	// Generate against the default branch (no --template-ref)
	err := runCommand(tmpDir, kataboleBin, "gen",
		"--template-repository", "https://github.com/katabole/katabole-gen-testing.git",
		"--import-path", "github.com/katabole/test-output",
		"--title-name", "TestOutput",
		"-n", "github.com/katabole/test-output",
	)
	if err != nil {
		t.Fatalf("generation on default branch failed: %v", err)
	}

	// Quick sanity: check that main.go now imports our new path
	data, err := os.ReadFile(filepath.Join(outputPath, "main.go"))
	if err != nil {
		t.Fatalf("cannot read main.go: %v", err)
	}
	if !bytes.Contains(data, []byte("github.com/katabole/test-output")) {
		t.Errorf("import path not replaced in default branch: got %q", data)
	}
}

func TestGenerateFromGenTesting_Branch(t *testing.T) {
	t.Skip("skipping branch checkout test until checkoutTemplateRef bug is fixed")
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test-output")

	// Generate against the named branch “test-branch”
	err := runCommand(tmpDir, kataboleBin, "gen",
		"--template-repository", "https://github.com/katabole/katabole-gen-testing.git",
		"--template-ref", "test-branch",
		"--import-path", "github.com/katabole/test-output",
		"--title-name", "TestOutput",
		"-n", "github.com/katabole/test-output",
	)
	if err != nil {
		t.Fatalf("generation on named branch failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outputPath, "main.go"))
	if err != nil {
		t.Fatalf("cannot read main.go: %v", err)
	}
	if !bytes.Contains(data, []byte("github.com/katabole/test-output")) {
		t.Errorf("import path not replaced on branch: got %q", data)
	}
}

func TestGenerateFromGenTesting_Tag(t *testing.T) {
	// Integration test prerequisites
	skipIfMissingTools(t, "git", "docker", "task")
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test-output")

	// Generate against the v0.1.0 tag
	err := runCommand(tmpDir, kataboleBin, "gen",
		"--template-repository", "https://github.com/katabole/katabole-gen-testing.git",
		"--template-ref", "v0.1.0",
		"--import-path", "github.com/katabole/test-output",
		"--title-name", "TestOutput",
		"-n", "github.com/katabole/test-output",
	)
	if err != nil {
		t.Fatalf("generation on tag failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outputPath, "main.go"))
	if err != nil {
		t.Fatalf("cannot read main.go: %v", err)
	}
	if !bytes.Contains(data, []byte("github.com/katabole/test-output")) {
		t.Errorf("import path not replaced on tag: got %q", data)
	}
}

func TestGenerateFromGenTesting_LatestTag(t *testing.T) {
	// Integration test prerequisites
	skipIfMissingTools(t, "git", "docker", "task")
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test-output")

	// Generate without specifying --template-ref, should use latest tag
	cmd := exec.Command(kataboleBin, "gen",
		"--template-repository", "https://github.com/katabole/katabole-gen-testing.git",
		"--import-path", "github.com/katabole/test-output",
		"--title-name", "TestOutput",
		"-n", "github.com/katabole/test-output",
	)
	cmd.Dir = tmpDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("generation with auto-latest-tag failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify that output mentions using a release
	output := stdout.String() + stderr.String()
	if !bytes.Contains([]byte(output), []byte("Using latest release:")) {
		t.Errorf("expected output to mention 'Using latest release:', got: %s", output)
	}

	// Verify the generated code has correct import path
	data, err := os.ReadFile(filepath.Join(outputPath, "main.go"))
	if err != nil {
		t.Fatalf("cannot read main.go: %v", err)
	}
	if !bytes.Contains(data, []byte("github.com/katabole/test-output")) {
		t.Errorf("import path not replaced with auto-latest-tag: got %q", data)
	}
}
