package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
)

func init() {
	genCmd.Flags().StringP("import-path", "n", "", "Name to use for renaming")
	genCmd.MarkFlagRequired("import-path")
	genCmd.Flags().StringP("title-name", "t", "", "Name for the app in title case")
	genCmd.MarkFlagRequired("title-name")
	genCmd.Flags().String("template-repository", "https://github.com/katabole/kbexample", "Git repository URL to clone as a template")
	genCmd.Flags().String("template-ref", "", "Git reference (commit hash or tag) to check out after cloning")
	genCmd.Flags().Bool("skip-setup", false, "Skip running 'task setup' after generation")
	rootCmd.AddCommand(genCmd)
}

var (
	genCmd = &cobra.Command{
		Use:     "gen",
		Short:   "Generate a new katabole web application",
		Example: "katabole gen --import-path github.com/myuser/myapp --title-name MyApp",
		RunE: func(cmd *cobra.Command, args []string) error {
			importPath, err := cmd.Flags().GetString("import-path")
			if err != nil {
				return err
			}
			if strings.Contains(importPath, " ") {
				return fmt.Errorf("import path must not contain spaces")
			}
			parts := strings.Split(importPath, "/")
			if len(parts) != 3 {
				return fmt.Errorf("import path must be in the form 'github.com/<user>/<app>'")
			}
			repoName := parts[2]
			// reponame only have alphanumerics, dashes or underscores
			if !ValidateRepoName(repoName) {
				return fmt.Errorf("repository name can only contain letters, numbers, dashes or underscores")
			}
			repoNameUnderscores := strings.ReplaceAll(repoName, "-", "_")

			titleName, err := cmd.Flags().GetString("title-name")
			if err != nil {
				return err
			}
			if strings.Contains(titleName, " ") {
				return fmt.Errorf("title name must not contain spaces")
			}
			if titleName == "" {
				titleName = strings.Title(repoName)
			}

			fmt.Printf("Creating %s... ", repoName)

			if _, err := os.Stat(repoName); err == nil {
				return fmt.Errorf("directory '%s' already exists", repoName)
			}

			tmpPath, err := os.MkdirTemp("", "kbexample")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmpPath)

			clonePath := filepath.Join(tmpPath, repoName)

			templateRepo, err := cmd.Flags().GetString("template-repository")
			if err != nil {
				return err
			}
			repo, err := git.PlainClone(clonePath, false, &git.CloneOptions{
				URL:  templateRepo,
				Tags: git.AllTags,
			})
			if err != nil {
				return err
			}

			err = repo.Fetch(&git.FetchOptions{
				RemoteName: "origin",
				Tags:       git.AllTags,
				RefSpecs:   []config.RefSpec{"refs/heads/*:refs/remotes/origin/*"},
				Force:      true,
			})
			if err != nil && err != git.NoErrAlreadyUpToDate {
				return err
			}

			templateRef, err := cmd.Flags().GetString("template-ref")
			if err != nil {
				return err
			}

			// User specified a specific commit(hash), branch, or tag to use
			if templateRef != "" {
				wt, err := repo.Worktree()
				if err != nil {
					return err
				}
				err = checkoutTemplateRef(wt, templateRef)
				if err != nil {
					return fmt.Errorf("error checking out: %v", err)
				}
			}

			os.RemoveAll(filepath.Join(clonePath, ".git"))

			err = filepath.WalkDir(clonePath, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					return nil
				}

				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				info, err := d.Info()
				if err != nil {
					return err
				}

				data = applyReplacements(data, importPath, titleName, repoName, repoNameUnderscores)
				if err := os.WriteFile(path, []byte(data), info.Mode()); err != nil {
					return err
				}

				return nil
			})
			if err != nil {
				return err
			}

			if err := os.Rename(clonePath, repoName); err != nil {
				return err
			}

			if err := checkNecessaryBinaries(); err != nil {
				return err
			}

			skipSetup, err := cmd.Flags().GetBool("skip-setup")
			if err != nil {
				return err
			}

			if !skipSetup {
				if err := runTaskSetup(repoName); err != nil {
					fmt.Printf("\nWarning: 'task setup' failed: %v\n", err)
					fmt.Printf("\nYou can run it manually with:\n")
					fmt.Printf("	cd %s && task setup\n\n", repoName)
				} else {
					fmt.Printf("\nSetup complete! Your application is ready.\n")
					fmt.Printf("\nNext:\n")
					fmt.Printf("	cd %s && task dev\n\n", repoName)
					return nil
				}
			} else {
				fmt.Printf("Done! (setup skipped)\n\n")
			}

			return nil
		},
	}
)

// git-go lacks a direct way to checkout a "ref"
// manually try it as a hash, then a branch, then a tag
func checkoutTemplateRef(wt *git.Worktree, templateRef string) error {
	if err := wt.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(templateRef),
	}); err == nil {
		return nil
	}

	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(templateRef),
	}); err == nil {
		return nil
	}

	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/remotes/origin/" + templateRef),
	}); err == nil {
		return nil
	}

	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewTagReferenceName(templateRef),
	}); err == nil {
		return nil
	}

	return fmt.Errorf("failed to checkout %s: not a valid hash, branch, or tag", templateRef)
}

// checkNecessaryBinaries calls checkApp and returns err
// Tries to find excutable files on user's machine and return an ErrorOrNil
// with links to install if the dependencies' excutables are not found
func checkNecessaryBinaries() error {

	var result *multierror.Error

	if err := checkApp("go", []string{"version"}); err != nil {
		err = errors.New("go is not installed, to install it see https://go.dev/doc/install")
		result = multierror.Append(result, err)
	}

	if err := checkApp("docker", []string{"version"}); err != nil {
		err = errors.New("docker is not installed, to install it see https://docs.docker.com/get-docker")
		result = multierror.Append(result, err)
	}

	if err := checkApp("node", []string{"-v"}); err != nil {
		err = errors.New("node is not installed, to install it see  https://docs.npmjs.com/downloading-and-installing-node-js-and-npm")
		result = multierror.Append(result, err)
	}

	if err := checkApp("psql", []string{"-V"}); err != nil {
		err = errors.New("psql is not installed, to install it see https://www.postgresql.org/download")
		result = multierror.Append(result, err)
	}

	if err := checkApp("task", []string{"--version"}); err != nil {
		err = errors.New("task is not installed, to install it see  https://taskfile.dev/installation")
		result = multierror.Append(result, err)
	}
	return result.ErrorOrNil()
}

// checkApp looks for executables to run on user's local machine
// and return nil if already installed and err otherwise
func checkApp(name string, args []string) error {
	cmd := exec.Command(name, args...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// runTaskSetup executes 'task setup' in the generated project directory
func runTaskSetup(dir string) error {
	fmt.Println("\nRunning 'task setup' in generated project...")
	cmd := exec.Command("task", "setup")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ValidateRepoName returns true if the repo name contains only alphanumerics, dashes, or underscores.
func ValidateRepoName(repoName string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	return re.MatchString(repoName)
}

// applyReplacements runs all of our placeholder swaps on the given content.
func applyReplacements(content []byte, importPath, titleName, repoName, repoNameUnderscores string) []byte {
	content = bytes.ReplaceAll(content, []byte("github.com/katabole/kbexample"), []byte(importPath))
	content = bytes.ReplaceAll(content, []byte("KBExample"), []byte(titleName))
	content = bytes.ReplaceAll(content, []byte("kb_example"), []byte(repoNameUnderscores))
	content = bytes.ReplaceAll(content, []byte("kbexample"), []byte(repoName))
	return content
}
