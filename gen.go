package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
)

func init() {
	genCmd.Flags().StringP("import-path", "n", "", "Name to use for renaming")
	genCmd.MarkFlagRequired("import-path")
	genCmd.Flags().StringP("title-name", "t", "", "Name for the app in title case")
	genCmd.MarkFlagRequired("title-name")
	genCmd.Flags().String("template-repository", "https://github.com/katabole/kbexample", "Git repository URL to clone as a template")
	genCmd.Flags().String("template-ref", "", "Git reference (commit hash or tag) to check out after cloning")
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
				URL:          templateRepo,
				SingleBranch: true,
				Depth:        1,
			})
			if err != nil {
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

				data = bytes.ReplaceAll(data, []byte("github.com/katabole/kbexample"), []byte(importPath))
				data = bytes.ReplaceAll(data, []byte("KBExample"), []byte(titleName))
				data = bytes.ReplaceAll(data, []byte("kbexample"), []byte(repoName))
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

			fmt.Printf(`Done!

Next:
	cd ` + repoName + ` && task setup

Ensure you have the following installed:
  - Go https://go.dev/doc/install
  - Task https://taskfile.dev/installation/
  - Docker https://www.docker.com/products/docker-desktop/
`)

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
		Branch: plumbing.NewTagReferenceName(templateRef),
	}); err == nil {
		return nil
	}

	return fmt.Errorf("failed to checkout %s: not a valid hash, branch, or tag", templateRef)
}

// Calling checkApp() to find out if go, docker, postsql, task, node are installed already to before running set up
func checkNecessaryBinaries() error {
	if err := checkApp("go", "version"); err != nil {
		return fmt.Errorf("go is not installed, to install it see https://go.dev/doc/install ")
	}

	if err := checkApp("docker", "version"); err != nil {
		return fmt.Errorf("docker is not installed, to install it see https://docs.docker.com/get-docker")
	}

	if err := checkApp("node", "-v"); err != nil {
		return fmt.Errorf("node is not installed, to install it see https://docs.npmjs.com/downloading-and-installing-node-js-and-npm")
	}

	if err := checkApp("psql", "-V"); err != nil {
		return fmt.Errorf("psql is not installed, to install it see https://www.postgresql.org/download/")

	}
	if err := checkApp("task", "-V"); err != nil {
		return fmt.Errorf("task is not installed, to install it see https://taskfile.dev/installation/")
	}
	return nil
}

// check if the necessary apps are installed to run $task setup
func checkApp(app, command string) error {
	path, err := exec.LookPath(app)
	if err != nil {
		return fmt.Errorf(" %s not found : %v", app, err)
	}

	fmt.Printf("%s is installed: %s\n", app, path)
	return nil
}
