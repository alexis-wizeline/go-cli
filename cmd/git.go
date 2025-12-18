package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

const githubRepoRegex = `^((https?:\/\/)|(git@))github\.com[\/:]{1}([\w.-]+)\/([\w.-]+?)(\.git)?$`

type gitActions struct{}

func (g gitActions) isRepo() (string, error) {
	cmd := exec.Command("git", "status")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (g gitActions) init(origin string) error {
	_, err := g.isRepo()
	if err == nil {
		return errors.New("current folder is already a repo")
	}

	cmd := exec.Command("git", "init", ".")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("there was an erro while trying to create the repo %e", err)
	}

	if len(origin) <= 0 {
		return errors.New("a origin is required")
	}

	if !rex.Match([]byte(origin)) {
		return errors.New("origin must be valid")
	}

	cmd = exec.Command("git", "remote", "add", "origin", origin)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("uanble to add origin: %e", err)
	}

	// TODO add options for main branch
	cmd = exec.Command("git", "branch", "-M", "main")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("unable to set main branch: %e", err)
	}

	cmd = exec.Command("git", "push", "-u", "origin", "main")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("unable to push to main branch: %e", err)
	}

	return nil
}

func (g gitActions) commitAll(message string) (string, error) {
	if len(strings.TrimSpace(message)) <= 0 {
		message = "added all changes"
	}

	err := g.trackFiles()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "commit", "-am", message)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (g gitActions) trackFiles() error {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")

	filesStr, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	files := strings.Split(string(filesStr), "\n")
	params := []string{"add"}
	folders := make(map[string]struct{})
	for _, file := range files {
		normalized := strings.TrimSpace(file)
		if len(normalized) <= 0 {
			continue
		}

		idx := strings.Index(normalized, "/")
		if idx > 0 {
			key := normalized[:idx+1]

			folders[key] = struct{}{}
			continue
		}

		params = append(params, normalized)
	}

	for folder := range folders {
		params = append(params, folder)
	}

	cmd = exec.Command("git", params...)
	return cmd.Run()
}

func (g gitActions) push() (string, error) {
	cmd := exec.Command("git", "push")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

var (
	commands = []string{
		"commit",
		//"pull",
		//"branch",
		"init",
	}

	rex = regexp.MustCompile(githubRepoRegex)

	actions = gitActions{}

	// gitCMD represents the git commands it will be using for shortcut git commands
	gitCmd = &cobra.Command{
		Use:     "git [action]",
		Short:   "commands for github",
		Long:    ` aliases to work with github commands`,
		Aliases: []string{"g"},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("run with any of these commands", commands)
		},
	}
	commitCmd = &cobra.Command{
		Use:     "commit",
		Short:   "add and commit all changes ot the repo",
		Aliases: []string{"c"},
		Run: func(cmd *cobra.Command, args []string) {
			status, err := actions.isRepo()
			if err != nil {
				fmt.Println(fmt.Sprintf("error: %e, with: %s", err, status))
				return
			}

			message, err := cmd.Flags().GetString("message")
			if err != nil {
				fmt.Println("no message found it wil use default message")
			}
			out, err := actions.commitAll(message)
			if err != nil {
				fmt.Println(fmt.Sprintf("error: %e, with: %s", err, out))
				return
			}

			commit, err := actions.push()
			if err != nil {
				fmt.Println(fmt.Sprintf("error: %e, with: %s", err, commit))
			}

		},
	}

	initCmd = &cobra.Command{
		Use:     "init -o [origin]",
		Short:   "init a repo in the current folder it requires the origin to work",
		Aliases: []string{"i"},
		Run: func(cmd *cobra.Command, args []string) {
			origin, err := cmd.Flags().GetString("origin")
			if err != nil {
				fmt.Println("a problem happens while setting the origin")
				return
			}

			err = actions.init(origin)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println("your folder now is a git repo!!!")
		},
	}
)

func init() {
	gitCmd.AddCommand(commitCmd)
	gitCmd.AddCommand(initCmd)
	rootCmd.AddCommand(gitCmd)

	commitCmd.Flags().StringP("message", "m", "", "the commit message")

	initCmd.Flags().StringP("origin", "o", "", "the origin for the repo to init")
	_ = initCmd.MarkFlagRequired("origin")
}
