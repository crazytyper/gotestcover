package main

import (
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/tools/cover"
)

func writeSummaryReport(profiles []*cover.Profile, out io.Writer) {

	if len(profiles) == 0 {
		fmt.Fprintln(out, "no lines to cover.")
		return
	}
	totals := aggregateProfiles(profiles)

	fmt.Fprintf(out, `statements total  : %d
statements covered: %d
total coverage    : %.1f%% of statements
`, totals.Total, totals.Covered, totals.Coverage())
}

func writeHotspotsReport(profiles []*cover.Profile, out io.Writer) {

	grouped := aggregateProfilesByGroup(profiles, func(p *cover.Profile) string { return packageName(p) })
	hotspots := filterHotspots(grouped)
	if len(hotspots) <= 0 {
		return
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, "Hotspots:\n")
	tw := tablewriter.NewWriter(out)
	tw.SetHeader([]string{"Package", "Uncovered", "Total", "% Coverage"})
	for _, hotspot := range hotspots {
		tw.Append([]string{
			hotspot.Name,
			fmt.Sprintf("%d", hotspot.Uncovered()),
			fmt.Sprintf("%d", hotspot.Total),
			fmt.Sprintf("%.1f", hotspot.Coverage())})
	}
	tw.Render()
	fmt.Fprintln(out)
}

func writeIgnoredReport(profiles []*cover.Profile, out io.Writer) {
	aggregated := aggregateProfilesByGroup(profiles, func(p *cover.Profile) string { return p.FileName })

	fmt.Fprintln(out)
	fmt.Fprintf(out, "Ignored:\n")
	tw := tablewriter.NewWriter(out)
	tw.SetHeader([]string{"File", "Covered", "Total", "% Coverage"})
	for _, agg := range aggregated {
		tw.Append([]string{
			agg.Name,
			fmt.Sprintf("%d", agg.Covered),
			fmt.Sprintf("%d", agg.Total),
			fmt.Sprintf("%.1f", agg.Coverage())})
	}
	tw.Render()
	fmt.Fprintln(out)
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
