package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
