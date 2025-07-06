package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var kataboleBin string

func TestMain(m *testing.M) {
	// build the katabole CLI locally
	bin := "katabole"
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build katabole CLI: %v\n", err)
		os.Exit(1)
	}
	// use the built binary path
	kataboleBin = filepath.Join(".", bin)
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

// runCommand is a helper to run commands inside a given directory
func runCommand(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func TestGenerateFromKbexample(t *testing.T) {
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
