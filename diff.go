package main

import (
	"fmt"
	"io/ioutil"

	"github.com/pmezard/go-difflib/difflib"
)

func diffFile(oldfile, newfile string) error {
	aContent, err := ioutil.ReadFile(oldfile)
	if err != nil {
		return err
	}

	bContent, err := ioutil.ReadFile(newfile)
	if err != nil {
		return err
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(aContent)),
		FromFile: oldfile,
		B:        difflib.SplitLines(string(bContent)),
		ToFile:   newfile,
	}

	text, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return err
	}

	if text == "" {
		fmt.Println("file contents are identical")
	} else {
		fmt.Printf(text)
	}
	return nil
}
