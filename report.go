package main

import (
	"fmt"
	"io"

	"golang.org/x/tools/cover"
)

func writeSummaryReport(profiles []*cover.Profile, out io.Writer) {

	if len(profiles) == 0 {
		fmt.Fprintln(out, "no lines to cover.")
		return
	}

	var total, covered int64
	for _, p := range profiles {
		t, c := totalStatementsCovered(p)
		total += t
		covered += c
	}
	var coverage float64
	if total != 0 {
		coverage = float64(covered) * 100.0 / float64(total)
	}

	fmt.Fprintf(out, `statements total  : %d
statements covered: %d
total coverage    : %.1f%% of statements
`, total, covered, coverage)
}

func writeCoverProfile(profiles []*cover.Profile, out io.Writer) {
	if len(profiles) == 0 {
		return
	}
	fmt.Fprintf(out, "mode: %s\n", profiles[0].Mode)
	for _, p := range profiles {
		for _, b := range p.Blocks {
			fmt.Fprintf(out, "%s:%d.%d,%d.%d %d %d\n", p.FileName, b.StartLine, b.StartCol, b.EndLine, b.EndCol, b.NumStmt, b.Count)
		}
	}
}

func totalStatementsCovered(p *cover.Profile) (total int64, covered int64) {
	for _, b := range p.Blocks {
		total += int64(b.NumStmt)
		if b.Count > 0 {
			covered += int64(b.NumStmt)
		}
	}
	return
}
