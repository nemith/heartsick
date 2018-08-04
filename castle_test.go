package main

import (
	"fmt"
	"testing"
)

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
