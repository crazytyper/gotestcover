package main

import (
	"path"
	"sort"

	"golang.org/x/tools/cover"
)

const (
	hotspotUncoveredMin = 20
	maxNumberOfHotspots = 10
)

type aggregate struct {
	Name    string
	Covered int64
	Total   int64
}

func (a aggregate) Uncovered() int64 { return a.Total - a.Covered }
func (a aggregate) Coverage() float64 {
	if a.Total == 0 {
		return 100.0
	}
	return 100.0 * float64(a.Covered) / float64(a.Total)
}

type byUncovered []aggregate

func (p byUncovered) Len() int           { return len(p) }
func (p byUncovered) Less(i, j int) bool { return p[i].Uncovered() > p[j].Uncovered() }
func (p byUncovered) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type byName []aggregate

func (p byName) Len() int           { return len(p) }
func (p byName) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p byName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func add(name string, lhs aggregate, rhs aggregate) aggregate {
	return aggregate{name, lhs.Covered + rhs.Covered, lhs.Total + rhs.Total}
}

func aggregateProfiles(profiles []*cover.Profile) aggregate {
	var total aggregate
	for _, p := range profiles {
		total = add("", total, aggregateProfile(p))
	}
	return total
}

func packageName(profile *cover.Profile) string {
	return path.Dir(profile.FileName)
}

func aggregateProfilesByGroup(profiles []*cover.Profile, grouper func(p *cover.Profile) string) []aggregate {
	grouped := map[string]aggregate{}
	for _, p := range profiles {
		gn := grouper(p)
		grouped[gn] = add(gn, grouped[gn], aggregateProfile(p))
	}

	groups := []aggregate{}
	for _, v := range grouped {
		groups = append(groups, v)
	}

	sort.Sort(byName(groups))
	return groups
}

func aggregateProfile(p *cover.Profile) aggregate {
	var total, covered int64
	for _, b := range p.Blocks {
		total += int64(b.NumStmt)
		if b.Count > 0 {
			covered += int64(b.NumStmt)
		}
	}
	return aggregate{Name: p.FileName, Covered: covered, Total: total}
}

func filterBy(as []aggregate, predicate func(a aggregate) bool) []aggregate {
	fas := []aggregate{}
	for _, a := range as {
		if predicate(a) {
			fas = append(fas, a)
		}
	}
	return fas
}

func isHotspot(a aggregate) bool {
	return a.Uncovered() >= hotspotUncoveredMin
}

func filterHotspots(as []aggregate) []aggregate {
	as = filterBy(as, isHotspot)
	sort.Sort(byUncovered(as))
	if len(as) > maxNumberOfHotspots {
		return as[:maxNumberOfHotspots]
	}
	return as
}
