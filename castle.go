package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	defaultCastle  = "dotfiles"
	subdirFilename = ".homesick_subdir"
)

var (
	castlePath    = filepath.Join(homeDir, ".homesick/repos")
	githubPattern = regexp.MustCompile(`^([A-Za-z0-9_-]+/[A-Za-z0-9_-]+)$`)
)

type castle string

var errCastleNotExist = errors.New("castle does not exist")

func loadCastle(name string) (castle, error) {
	if _, err := os.Stat(filepath.Join(castlePath, name, ".git")); os.IsNotExist(err) {
		return "", errCastleNotExist
	}

	return castle(name), nil
}

// allCastles returns all directories found in castlePath
func allCastles() ([]castle, error) {
	files, err := ioutil.ReadDir(castlePath)
	if os.IsNotExist(err) {
		return []castle{}, nil
	}
	if err != nil {
		return nil, err
	}

	castles := make([]castle, 0, len(files))
	for _, f := range files {
		// skip files
		if !f.IsDir() {
			continue
		}

		// skip directories that don't have ".git"
		if _, err := os.Stat(filepath.Join(castlePath, f.Name(), ".git")); os.IsNotExist(err) {
			continue
		}

		castles = append(castles, castle(f.Name()))
	}

	return castles, nil
}

// basePath will return the path to the base of the caslte.
func (c castle) basePath() string {
	return filepath.Join(castlePath, string(c))
}

func (c castle) homePath() string {
	return filepath.Join(c.basePath(), "home")
}

// remote returns the remote uri for a given castle
func (c castle) remote() (string, error) {
	return gitRemoteURL(c.basePath())
}

func (c castle) update() error {
	return gitPull(c.basePath())
	// TODO(bbennett) submodules
}

func (c castle) push() error {
	return gitPush(c.basePath())
}

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

func linkables(dir, base string, subdirs []string) ([]string, error) {
	relPath, err := filepath.Rel(base, dir)
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(dir)
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

	for _, subdir := range subdirs {
		subdirLinks, err := linkables(filepath.Join(baseHome, subdir), baseHome, subdirs)
		if err != nil {
			return nil, err
		}
		links = append(links, subdirLinks...)
	}

	return links, nil

}

func (c castle) subdirs() ([]string, error) {
	subdirFile := filepath.Join(c.basePath(), subdirFilename)
	f, err := os.Open(subdirFile)
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
