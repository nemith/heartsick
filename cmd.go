package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	flagAll bool
)

func init() {
	cloneCmd := &cobra.Command{
		Use:   "clone URI CASTLE_NAME",
		Short: "clone +uri+ as a castle with name CASTLE_NAME for homesick",
		Run:   cmdClone,
		Args:  cobra.RangeArgs(1, 2),
	}

	commitCmd := &cobra.Command{
		Use:   "commit CASTLE MESSAGE",
		Short: "commit the specified castle's changes",
		Run:   cmdCommit,
	}

	// TODO(bbennett): cmdDestory

	diffCmd := &cobra.Command{
		Use:   "diff CASTLE",
		Short: "shows the git diff of uncommitted changes in a castle",
		Run:   cmdDiff,
	}

	execCmd := &cobra.Command{
		Use:   "exec CASTLE|all COMMAND",
		Short: "execute a single shell command inside the root of a castle",
		Run:   cmdExec,
		Args:  cobra.MinimumNArgs(2),
	}

	execAllCmd := &cobra.Command{
		Use:   "exec_all COMMAND",
		Short: "execute a single shell command inside the root of every cloned castle",
		Run:   cmdExecAll,
		Args:  cobra.MinimumNArgs(1),
	}

	generateCmd := &cobra.Command{
		Use:   "generate PATH",
		Short: "generate a homesick-ready git repo at PATH",
		Run:   cmdGenerate,
		Args:  cobra.ExactArgs(1),
	}

	linkCmd := &cobra.Command{
		Use:     "link",
		Aliases: []string{"symlink"},
		Short:   "symlinks all dotfiles from the specified castle",
		Run:     cmdLink,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list cloned castles",
		Run:   cmdList,
		Args:  cobra.NoArgs,
	}

	openCmd := &cobra.Command{
		Use:     "open CASTLE",
		Aliases: []string{"edit"},
		Short:   "open your default editor in the root of the given castle",
		Run:     cmdOpen,
	}

	pathCmd := &cobra.Command{
		Use:     "show_path CASTLE",
		Aliases: []string{"path"},
		Short:   "prints the path of a castle",
		Run:     cmdPath,
	}

	pullCmd := &cobra.Command{
		Use:   "pull CASTLE",
		Short: "update the specified castle",
		Run:   cmdPull,
	}
	pullCmd.PersistentFlags().BoolVarP(&flagAll, "all", "", false, "update all cloned castles")

	pushCmd := &cobra.Command{
		Use:   "push CASTLE",
		Short: "push the specified castle",
		Run:   cmdPush,
	}

	rcCmd := &cobra.Command{
		Use:   "rc CASTLE",
		Short: "un the .homesickrc for the specified castle",
		Run:   cmdRC,
	}

	shellCmd := &cobra.Command{
		Use:     "cd CASTLE",
		Aliases: []string{"shell"},
		Short:   "open a new shell in the root of the given castle",
		Run:     cmdShell,
	}

	statusCmd := &cobra.Command{
		Use:   "status CASTLE",
		Short: "shows the git status of a castle",
		Run:   cmdStatus,
	}

	trackCmd := &cobra.Command{
		Use:   "track FILE CASTLE",
		Short: "add a file to a castle",
		Run:   cmdTrack,
		Args:  cobra.MinimumNArgs(1),
	}

	// TODO(bbennett): unlinkCmd

	versionCmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"ver"},
		Short:   "display the current version of homesick",
		Run:     cmdVersion,
		Args:    cobra.NoArgs,
	}

	rootCmd.AddCommand(
		cloneCmd,
		commitCmd,
		diffCmd,
		execAllCmd,
		execCmd,
		generateCmd,
		linkCmd,
		listCmd,
		openCmd,
		pathCmd,
		pullCmd,
		pushCmd,
		rcCmd,
		shellCmd,
		statusCmd,
		trackCmd,
		versionCmd,
	)
}

func castleFromArgs(args []string) *castle {
	var name = defaultCastle
	if len(args) > 0 {
		name = args[0]
	}
	castle, err := loadCastle(name)
	if err != nil {
		fatalf("failed to load castle: %s", err)
	}
	return castle
}

func mustAllCastles() []*castle {
	castles, err := allCastles()
	if err != nil {
		fatalf("failed to find castles: %s", err)
	}
	return castles
}

var githubPattern = regexp.MustCompile(`^([A-Za-z0-9_-]+/[A-Za-z0-9_-]+)$`)

func cmdClone(cmd *cobra.Command, args []string) {
	uri := args[0]

	var castleName string
	if len(args) > 1 {
		castleName = args[1]
	}

	// expand out a github path if only user/repo given
	if githubPattern.MatchString(uri) {
		uri = "https://github.com/" + uri + ".git"
	}

	// obtain the castle name from the repo name if one wasn't pas
	if castleName == "" {
		castleName = strings.TrimSuffix(filepath.Base(uri), ".git")
	}

	dest := filepath.Join(castlePath(), castleName)

	if _, err := os.Stat(dest); err == nil {
		status(colorBrBlue, "exist", dest)
		return
	}

	statusf(colorBrGreen, "git clone", "%s to %s", uri, dest)
	if err := gitClone(uri, dest); err != nil {
		fatalf("failed to clone '%s': %v", uri, err)
	}
	// TODO(bbennett): clone/update submodules?
}

func cmdCommit(cmd *cobra.Command, args []string) {
	castle := castleFromArgs(args)

	var commitMsg string
	if len(args) > 1 {
		commitMsg = strings.Join(args[1:], " ")

	}

	status(colorBrGreen, "git commit all", castle.name)
	if err := gitCommitAll(castle.path, commitMsg); err != nil {
		fatalf("failed to commit: %v", err)
	}
}

func cmdDiff(cmd *cobra.Command, args []string) {
	castle := castleFromArgs(args)

	status(colorBrGreen, "git diff", castle.name)
	if err := gitDiff(castle.path); err != nil {
		fatalf("failed to diff: %v", err)
	}
}

func execHelper(castle *castle, args []string) {
	var cmdArgs []string
	if len(args) > 1 {
		cmdArgs = args[1:]
	}
	statusf(colorBrGreen, "exec", "%s %s in castle '%s'", args[0], strings.Join(cmdArgs, " "), castle.name)
	c := exec.Command(args[0], cmdArgs...)
	c.Dir = castle.path
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	if err := c.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			fatalf("failed to run command: %v", err)
		}
	}
}

func cmdExec(cmd *cobra.Command, args []string) {
	castle := castleFromArgs(args)
	execHelper(castle, args[1:])
}

func cmdExecAll(cmd *cobra.Command, args []string) {
	for _, c := range mustAllCastles() {
		execHelper(c, args)
	}
}

func cmdGenerate(cmd *cobra.Command, args []string) {
	path := args[0]

	createDir(path)

	if isGitDir(path) {
		status(colorBrBlue, "git init", "already initalized")
	} else {
		status(colorBrGreen, "git init", path)
		if err := gitInit(path); err != nil {
			fatalf("failed to git init: %v", err)
		}
	}

	// If a github user is defined add it as the default remote.  This is a bit
	// weird but it's what homesick does.
	ghUser, _ := gitConfig("/", "github.user")
	if ghUser != "" {
		url := "https://github.com/" + ghUser + "/" + filepath.Base(path) + ".git"

		if gitRemoteExists(path, "origin") {
			statusf(colorBrBlue, "git remote", "%s already exists", "origin")
		} else {
			statusf(colorBrGreen, "git remote", "add %s %s", "origin", url)
			if err := gitRemoteAdd(path, "origin", url); err != nil {
				fatalf("failed to add remote: %v", err)
			}
		}
	}

	createDir(filepath.Join(path, "home"))
}

const conflictHelp = `	Y - yes, overwrite
	n - no, do not overwrite
	a - all, overwrite this and all other
	q - quit, abort
	d - diff, show the differences between old and new
	h - help, show this help
`

func conflictPrompt(oldfile, newfile string) (skip bool, yesAll bool) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("Overwrite %s? (enter 'h' for help) [Ynaqdh] ", newfile)

		if ok := scanner.Scan(); !ok {
			break
		}

		switch scanner.Text() {
		case "N", "n":
			return true, false
		case "a", "A":
			return false, true
		case "d", "D":
			if err := diffFile(oldfile, newfile); err != nil {
				errorf("failed to diff files: %v", err)
			}
		case "h", "H":
			fmt.Println(conflictHelp)
		// default is 'Y'
		default:
			return false, false
		}
	}
	if err := scanner.Err(); err != nil {
		fatalf("failed to read input: %v", err)
	}

	fatalf("prompt failed specacularly!")
	return false, false
}

func cmdLink(cmd *cobra.Command, args []string) {
	castle := castleFromArgs(args)

	links, subdirs, err := castle.linkables()
	if err != nil {
		fatalf("failed to find links: %v", err)
	}

	for _, subdir := range subdirs {
		subdir := filepath.Join(homeDir, subdir)

		err := os.MkdirAll(subdir, 0755)
		if os.IsExist(err) {
			fi, err := os.Lstat(subdir)
			if err != nil {
				fatalf("failed to read subdir '%s': %v", subdir, err)
			}

			if !fi.IsDir() {
				fatalf("subdir '%s' already exists but isn't a directory", subdir)
			}
			status(colorBrBlue, "exists", subdir)
		} else if err != nil {
			fatalf("failed to create subdir '%s': %v", subdir, err)
		}
		status(colorBrGreen, "mkdir", subdir)
	}

	castleHome := castle.homePath()

	allYes := false

	for _, link := range links {
		oldname := filepath.Join(castleHome, link)
		newname := filepath.Join(homeDir, link)

		var skip bool // skip if there is a conflict

		if fi, err := os.Lstat(newname); err == nil {
			if fi.Mode()&os.ModeSymlink != 0 {
				existingLink, err := os.Readlink(newname)
				if err != nil {
					fatalf("failed to real link: %v", err)
				}
				if existingLink == oldname {
					status(colorBrBlue, "identical", oldname)
					continue
				}
			}

			if !allYes {
				statusf(colorBrRed, "conflict", "%s exists", oldname)
				skip, allYes = conflictPrompt(oldname, newname)
			}
		}

		if !skip {
			statusf(colorBrGreen, "symlink", "%s to %s", oldname, newname)
			if err := os.RemoveAll(newname); err != nil {
				errorf("failed to remove old file: %v", err)
			}

			if err := os.Symlink(oldname, newname); err != nil {
				errorf("failed to symlink file: %v", err)
			}
		}
	}
}

func cmdList(cmd *cobra.Command, args []string) {
	for _, c := range mustAllCastles() {
		remote, err := c.remote()
		if err != nil {
			statusf(colorBrRed, c.name, "failed to get remote uri: %v", err)
			continue
		}
		status(colorBrCyan, c.name, remote)
	}
}

func cmdOpen(cmd *cobra.Command, args []string) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		fatalf("The $EDITOR enviroment variable must be set to use this command")
	}

	castle := castleFromArgs(args)

	statusf(colorBrGreen, editor, "Opening the root directory of castle '%s' in editor '%s'", castle, editor)
	c := exec.Command(editor, castle.path)
	c.Dir = castle.path
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fatalf("failed to open editor: %v", err)
	}

}

func cmdPath(cmd *cobra.Command, args []string) {
	castle := castleFromArgs(args)
	fmt.Println(castle.path)
}

func cmdPull(cmd *cobra.Command, args []string) {
	var castles []*castle
	if flagAll {
		castles = mustAllCastles()
	} else {
		castles = []*castle{castleFromArgs(args)}
	}

	var fail bool
	for _, c := range castles {
		remote, err := c.remote()
		if err != nil {
			errorf("failed to get remote for castle: %v", err)
			fail = true
			continue
		}

		statusf(colorBrGreen, "git pull", "%s to castle '%s'", remote, c)
		if err := c.update(); err != nil {
			errorf("failed to update castle: %v", err)
			fail = true
		}
	}
	if fail {
		os.Exit(1)
	}
}

func cmdPush(cmd *cobra.Command, args []string) {
	castle := castleFromArgs(args)
	if err := castle.push(); err != nil {
		fatalf("failed to push castle: %v", err)
	}
}

func cmdRC(cmd *cobra.Command, args []string) {
	fatalf("rc not implemented yet in heartsick")
}

func cmdShell(cmd *cobra.Command, args []string) {
	castle := castleFromArgs(args)

	statusf(colorBrGreen, "shell", "Opening new shell in '%s To return to the original one exit from the new shell.", castle.path)
	c := exec.Command(os.Getenv("SHELL"))
	c.Dir = castle.path
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fatalf("failed to open shell: %v", err)
	}
}

func cmdStatus(cmd *cobra.Command, args []string) {
	castle := castleFromArgs(args)

	statusf(colorBrGreen, "git status", "%s for castle '%s", castle.path, castle.name)
	if err := gitStatus(castle.path); err != nil {
		fatalf("failed to get status: %v", err)
	}
}

func cmdTrack(cmd *cobra.Command, args []string) {
	path := args[0]

	absPath, err := filepath.Abs(path)
	if err != nil {
		fatalf("failed to get absolute path: %v", err)
	}

	relPath, err := filepath.Rel(homeDir, absPath)
	if err != nil {
		fatalf("failed to get absolute path: %v", err)
	}

	fmt.Println(relPath)

}

func cmdVersion(cmd *cobra.Command, args []string) {
	fmt.Println(heartsickVer)
}

func createDir(path string) {
	if _, err := os.Stat(path); err == nil {
		status(colorBrBlue, "exist", path)
		return
	}

	status(colorBrGreen, "create", path)
	if err := os.Mkdir(path, 0755); err != nil && !os.IsExist(err) {
		fatalf("failed to create directory '%s': %v", path, err)
	}
}
