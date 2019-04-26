package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func testHomedir(t *testing.T, base string) {
	t.Helper()
	homeDir = filepath.Join("testdata", base)
	t.Logf("setting homedir to %s", homeDir)
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
			testHomedir(t, tc.home)

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

			wantPath, _ := filepath.Abs(filepath.Join("testdata", tc.home, ".homesick/repos", tc.castle))
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
			testHomedir(t, tc.home)

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
		{"home1", "private", []string{".file1"}, false},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%s:%s", tc.home, tc.castle), func(t *testing.T) {
			testHomedir(t, tc.home)

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
			testHomedir(t, tc.home)

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
