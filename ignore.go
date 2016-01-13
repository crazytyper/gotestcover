package main

import (
	"github.com/monochromegane/go-gitignore"
	"golang.org/x/tools/cover"
	"os"
)

func ignore(ps []*cover.Profile) (keep []*cover.Profile, ignored []*cover.Profile) {
	matcher, err := gitignore.NewGitIgnore(".coverignore", "")
	if os.IsNotExist(err) {
		return ps, []*cover.Profile{}
	} else if err != nil {
		panic(err)
	}

	keep = []*cover.Profile{}
	ignored = []*cover.Profile{}
	for _, p := range ps {
		if matcher.Match(p.FileName, false) {
			ignored = append(ignored, p)
		} else {
			keep = append(keep, p)
		}
	}

	return keep, ignored
}
