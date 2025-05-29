package main

import (
	"bytes"
	"regexp"
	"testing"
)

func TestApplyReplacements(t *testing.T) {
	template := []byte(`
import "github.com/katabole/kbexample"
Title: KBExample
psql -d kbexample_dev
psql -d kb_example_dev
`)

	// 1) No-dash case: repoName == repoNameUnder, so we expect two identical replacements.
	out := applyReplacements(template,
		"github.com/foo/myapp", // importPath
		"MyApp",                // titleName
		"myapp",                // repoName
		"myapp",                // repoNameUnderscores
	)
	count := bytes.Count(out, []byte("myapp_dev"))
	if count != 2 {
		t.Errorf("no-dash: expected 2 occurrences of \"myapp_dev\", got %d\nOutput:\n%s",
			count, out)
	}

	// 2) With-dash case: expect one raw and one underscored
	out = applyReplacements(template,
		"github.com/foo/my-app",
		"MyApp",
		"my-app",
		"my_app",
	)
	if c := bytes.Count(out, []byte("my-app_dev")); c != 1 {
		t.Errorf("with-dash: expected 1 occurrence of \"my-app_dev\", got %d\nOutput:\n%s",
			c, out)
	}
	if c := bytes.Count(out, []byte("my_app_dev")); c != 1 {
		t.Errorf("with-dash: expected 1 occurrence of \"my_app_dev\", got %d\nOutput:\n%s",
			c, out)
	}
}

func TestRepoNameValidation(t *testing.T) {
	re := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

	valid := []string{"abc", "a_b-c", "A1_2-3"}
	for _, s := range valid {
		if !re.MatchString(s) {
			t.Errorf("should be valid: %q", s)
		}
	}

	invalid := []string{"bad$name", " space", "dot.name"}
	for _, s := range invalid {
		if re.MatchString(s) {
			t.Errorf("should be invalid: %q", s)
		}
	}
}
