package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", path)
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		// Make File
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func setupHomedir(t *testing.T, base string) (string, func()) {
	t.Helper()

	homeZip := filepath.Join("testdata", base+".zip")
	tmpHomeDir, err := ioutil.TempDir("", base)
	if err != nil {
		t.Errorf("failed to create homedir: %v", err)
	}

	if err := unzip(homeZip, tmpHomeDir); err != nil {
		t.Errorf("failed ot unzip homedir: %v", err)
	}

	cleanupFn := func() {
		if err := os.RemoveAll(tmpHomeDir); err != nil {
			t.Logf("failed to remove temp homedir: %v", err)
		}
	}

	t.Logf("setting homedir to %s", tmpHomeDir)
	homeDir = tmpHomeDir

	return homeDir, cleanupFn
}

func TestLoadCastle(t *testing.T) {
	tt := []struct {
		home, castle string
		fail         bool
	}{
		{"emptyHome", "dotfiles", true},
		{"noRepos", "dotfiles", true},
		{"home1", "dotfiles", false},
		{"home1", "private", false},
		{"home1", "nogit", true},
		{"home1", "none", true},
		{"home1", "", true},
	}

	for _, tc := range tt {
		t.Run(tc.home+":"+tc.castle, func(t *testing.T) {
			tmpHomePath, cleanup := setupHomedir(t, tc.home)
			defer cleanup()

			got, err := loadCastle(tc.castle)
			if err != nil {
				if !tc.fail {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if got == nil {
				t.Fatal("got nil castle without an error")
			}

			if got.name != tc.castle {
				t.Errorf("castle name is wrong (want '%s', got '%s')", tc.castle, got.name)
			}

			wantPath, _ := filepath.Abs(filepath.Join(tmpHomePath, ".homesick/repos", tc.castle))
			if got.path != wantPath {
				t.Errorf("castle path wrong (want: '%s', got: '%s)", wantPath, got.path)
			}

		})
	}
}

func TestAllCastles(t *testing.T) {
	tt := []struct {
		home string
		want []string
	}{
		{"emptyHome", []string{}},
		{"noRepos", []string{}},
		{"home1", []string{"dotfiles", "private"}},
	}

	for _, tc := range tt {
		t.Run(tc.home, func(t *testing.T) {
			_, cleanup := setupHomedir(t, tc.home)
			defer cleanup()

			got, err := allCastles()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			gotNames := make([]string, 0, len(got))
			for _, c := range got {
				gotNames = append(gotNames, c.name)
			}

			if !cmp.Equal(tc.want, gotNames) {
				t.Errorf("incorrect castle names:\n%s", cmp.Diff(tc.want, gotNames))
			}
		})
	}
}

func TestIsParentPath(t *testing.T) {
	tt := []struct {
		parent string
		path   string
		want   bool
	}{
		{".config", ".config", true},
		{".config", ".bashrc", false},
		{".local/share", ".local", true},
		{".local/share", ".bashrc", false},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%s,%s", tc.parent, tc.path), func(t *testing.T) {
			got := isParentPath(tc.parent, tc.path)

			if got != tc.want {
				t.Errorf("unexpected result (got: %t, want %t)", got, tc.want)
			}
		})
	}
}

func TestCastleLinkables(t *testing.T) {
	tt := []struct {
		home, castle string
		want         []string
		fail         bool
	}{
		{"home1", "dotfiles",
			[]string{
				".dir1",
				".file1",
				".dir2/.file1",
				".dir2/.file2",
				".dir3/.subdir1/.file1",
				".dir3/.subdir1/.file2",
			}, false},
		{"home1", "private", []string{".file1", ".file2"}, false},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%s:%s", tc.home, tc.castle), func(t *testing.T) {
			_, cleanup := setupHomedir(t, tc.home)
			defer cleanup()

			castle, err := loadCastle(tc.castle)
			if err != nil {
				t.Fatalf("failed to load castle: %v", err)
			}

			got, err := castle.linkables()
			if err != nil && !tc.fail {
				t.Errorf("unexpected error: %v", err)
			}

			if !cmp.Equal(tc.want, got) {
				t.Errorf("wrong linkables returned:\n%s", cmp.Diff(tc.want, got))
			}

		})
	}
}

func TestCastleSubdirs(t *testing.T) {
	tt := []struct {
		home, castle string
		want         []string
		fail         bool
	}{
		{"home1", "dotfiles", []string{".dir2", ".nonexistent", ".dir3/.subdir1"}, false},
		{"home1", "private", []string{}, false},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%s:%s", tc.home, tc.castle), func(t *testing.T) {
			_, cleanup := setupHomedir(t, tc.home)
			defer cleanup()

			castle, err := loadCastle(tc.castle)
			if err != nil {
				t.Fatalf("failed to load castle: %v", err)
			}

			got, err := castle.subdirs()
			if err != nil && !tc.fail {
				t.Errorf("unexpected error: %v", err)
			}

			if !cmp.Equal(tc.want, got) {
				t.Errorf("wrong subdirs returned:\n%s", cmp.Diff(tc.want, got))
			}

		})
	}
}

func TestMain(m *testing.M) {
	// overwrite homedir to make sure tests don't do anything stupid
	tmpHomeDir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(fmt.Sprintf("failed to create failback homedir: %v", err))
	}
	homeDir = tmpHomeDir

	flag.Parse()
	os.Exit(m.Run())
}
