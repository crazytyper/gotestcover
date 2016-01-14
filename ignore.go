package main

import (
	gitignore "github.com/crazytyper/gotestcover/ignore"
	"golang.org/x/tools/cover"
	"os"
)

func ignore(ps []*cover.Profile) (keep []*cover.Profile, ignored []*cover.Profile) {
	matcher, err := gitignore.CompileIgnoreFile(".coverignore")
	if os.IsNotExist(err) {
		return ps, []*cover.Profile{}
	} else if err != nil {
		panic(err)
	}

	keep = []*cover.Profile{}
	ignored = []*cover.Profile{}
	for _, p := range ps {
		if matcher.MatchesPath(p.FileName) {
			ignored = append(ignored, p)
		} else {
			keep = append(keep, p)
		}
	}

	return keep, ignored
}
