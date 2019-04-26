package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultCastle  = "dotfiles"
	subdirFilename = ".homesick_subdir"
)

type castle struct {
	name string
	path string
}

func castlePath() string {
	return filepath.Join(homeDir, ".homesick/repos")
}

var errCastleNotExist = errors.New("castle does not exist")

func loadCastle(name string) (*castle, error) {
	if name == "" {
		return nil, errCastleNotExist
	}

	path, err := filepath.Abs(filepath.Join(castlePath(), name))
	if err != nil {
		return nil, fmt.Errorf("cannot find path for castle: %v", err)
	}

	if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
		if os.IsNotExist(err) {
			return nil, errCastleNotExist
		}
		return nil, fmt.Errorf("couldn't open castle: %v", err)
	}

	return &castle{
		name: name,
		path: path,
	}, nil
}

// allCastles returns all directories found in castlePath
func allCastles() ([]*castle, error) {
	files, err := ioutil.ReadDir(castlePath())
	if os.IsNotExist(err) {
		return []*castle{}, nil
	}
	if err != nil {
		return nil, err
	}

	castles := make([]*castle, 0, len(files))
	for _, f := range files {
		// skip files
		if !f.IsDir() {
			continue
		}

		castle, err := loadCastle(f.Name())
		if err != nil {
			continue
		}

		castles = append(castles, castle)
	}

	return castles, nil
}

// homePath will return the `home` directory in the castle.
func (c castle) homePath() string {
	return filepath.Join(c.path, "home")
}

// remote returns the remote uri for a given castle.
func (c castle) remote() (string, error) {
	return gitRemoteURL(c.path)
}

// update will update the castle from git.
func (c castle) update() error {
	return gitPull(c.path)
	// TODO(bbennett) submodules
}

// push will push commited changes to the remote.
func (c castle) push() error {
	return gitPush(c.path)
}

// isParentPath will return true if the path has parent as a parent.
func isParentPath(parent, path string) bool {
	parts := strings.Split(parent, string(filepath.Separator))
	for i := range parts {
		base := strings.Join(parts[:i+1], string(filepath.Separator))
		if base == path {
			return true
		}
	}
	return false
}

// linkables is a helper for castle.linkables() returning a list of files to
// be linked in a directory.  If the file/dir is a parent of any given subdirs
// then it will be excluded.  Results are all relative to the given base
// directory.
func linkables(dir, base string, subdirs []string) ([]string, error) {
	relPath, err := filepath.Rel(base, dir)
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(dir)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	var links []string
Loop:
	for _, f := range files {
		rel := filepath.Join(relPath, f.Name())
		for _, subdir := range subdirs {
			if isParentPath(subdir, rel) {
				continue Loop
			}
		}
		links = append(links, rel)
	}
	return links, nil
}

// linkables will find all files/directories that are eligible to be linked.
// Only top level dir/files are linked a long with any sub-directories found in
// the .homesick_subdir file at the top of the castle.
func (c castle) linkables() ([]string, error) {
	subdirs, err := c.subdirs()
	if err != nil {
		return nil, err
	}

	baseHome := c.homePath()
	links, err := linkables(baseHome, baseHome, subdirs)
	if err != nil {
		return nil, err
	}

	// find all linkables for the subdirs
	for _, subdir := range subdirs {
		subdirLinks, err := linkables(filepath.Join(baseHome, subdir), baseHome, subdirs)
		if err != nil {
			return nil, err
		}
		links = append(links, subdirLinks...)
	}

	return links, nil
}

// subdirs will read the .homesick_subdir file from the castle and return the
// a list of directories.  Directories are defined as one per line.
func (c castle) subdirs() ([]string, error) {
	subdirFile := filepath.Join(c.path, subdirFilename)

	f, err := os.Open(subdirFile)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read subdir file '%s': %v", subdirFile, err)
	}

	var subdirs []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		subdirs = append(subdirs, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read subdir file '%s': %v", subdirFile, err)
	}

	return subdirs, nil
}
