package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
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
	genCmd.Flags().StringP("template-repository", "r", "https://github.com/katabole/kbexample", "Git repository URL to clone as a template")
	genCmd.Flags().StringP("template-ref", "f", "", "Git reference (commit hash or tag) to check out after cloning")
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

			// User specified a specific commit to use
			if templateRef != "" {
				wt, err := repo.Worktree()
				if err != nil {
					return err
				}
				err = wt.Checkout(&git.CheckoutOptions{
					Hash: plumbing.NewHash(templateRef),
				})
				if err != nil {
					return fmt.Errorf("failed to checkout ref %s: %w", templateRef, err)
				}

				fmt.Printf("Checked out template repository at %s\n", templateRef)
			}

			fmt.Printf("Creating %s from template %s...\n", repoName, templateRepo)
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
