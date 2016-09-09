// Package gotestcover provides multiple packages support for Go test cover.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"sync"

	"golang.org/x/tools/cover"
)

var (
	// go build
	flagA    bool
	flagX    bool
	flagRace bool
	flagTags string

	// go test
	flagV            bool
	flagCount        int
	flagCPU          string
	flagParallel     string
	flagRun          string
	flagShort        bool
	flagTimeout      string
	flagCoverMode    string
	flagCoverProfile string

	// custom
	flagParallelPackages = runtime.GOMAXPROCS(0)

	// reports
	flagReportJSON bool
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	err := parseFlags()
	if err != nil {
		return err
	}

	pkgs, err := getPackages()
	if err != nil {
		return err
	}

	cov, failed := runAllPackageTests(pkgs, func(out string) {
		fmt.Print(out)
	})

	// write the merged profile
	file, err := os.Create(flagCoverProfile)
	if err != nil {
		return err
	}
	writeCoverProfile(cov, file)
	file.Close()

	cov, ignored := ignore(cov)

	if len(ignored) > 0 {
		writeIgnoredReport(ignored, os.Stdout)
	}
	writeHotspotsReport(cov, os.Stdout)
	writeSummaryReport(cov, os.Stdout)

	if failed {
		return fmt.Errorf("test failed")
	}
	return nil
}

func parseFlags() error {
	flag.BoolVar(&flagA, "a", flagA, "see 'go build' help")
	flag.BoolVar(&flagX, "x", flagX, "see 'go build' help")
	flag.BoolVar(&flagRace, "race", flagRace, "see 'go build' help")
	flag.StringVar(&flagTags, "tags", flagTags, "see 'go build' help")

	flag.BoolVar(&flagV, "v", flagV, "see 'go test' help")
	flag.IntVar(&flagCount, "count", flagCount, "see 'go test' help")
	flag.StringVar(&flagCPU, "cpu", flagCPU, "see 'go test' help")
	flag.StringVar(&flagParallel, "parallel", flagParallel, "see 'go test' help")
	flag.StringVar(&flagRun, "run", flagRun, "see 'go test' help")
	flag.BoolVar(&flagShort, "short", flagShort, "see 'go test' help")
	flag.StringVar(&flagTimeout, "timeout", flagTimeout, "see 'go test' help")
	flag.StringVar(&flagCoverMode, "covermode", flagCoverMode, "see 'go test' help")
	flag.StringVar(&flagCoverProfile, "coverprofile", flagCoverProfile, "see 'go test' help")

	flag.IntVar(&flagParallelPackages, "parallelpackages", flagParallelPackages, "Number of package test run in parallel")

	flag.Parse()
	if flagCoverProfile == "" {
		flagCoverProfile = "out.coverprofile"
	}
	if flagParallelPackages < 1 {
		return fmt.Errorf("flag parallelpackages must be greater than or equal to 1")
	}
	return nil
}

func argsBefore(this string) []string {
	args := flag.Args()
	for i, arg := range args {
		if arg == this {
			return args[:i]
		}
	}
	return args
}

func argsAfter(this string) []string {
	args := flag.Args()
	for i, arg := range args {
		if arg == this {
			return args[i+1:]
		}
	}
	return args
}

func getPackages() ([]string, error) {
	cmdArgs := []string{"list"}
	cmdArgs = append(cmdArgs, argsBefore("--")...)
	cmdOut, err := runGoCommand(cmdArgs...)
	if err != nil {
		return nil, err
	}
	var pkgs []string
	sc := bufio.NewScanner(bytes.NewReader(cmdOut))
	for sc.Scan() {
		pkgs = append(pkgs, sc.Text())
	}
	return pkgs, nil
}

func runAllPackageTests(pkgs []string, pf func(string)) ([]*cover.Profile, bool) {
	pkgch := make(chan string)
	type res struct {
		out string
		cov []*cover.Profile
		err error
	}
	resch := make(chan res)
	wg := new(sync.WaitGroup)
	wg.Add(flagParallelPackages)
	go func() {
		for _, pkg := range pkgs {
			pkgch <- pkg
		}
		close(pkgch)
		wg.Wait()
		close(resch)
	}()
	for i := 0; i < flagParallelPackages; i++ {
		go func() {
			for p := range pkgch {
				out, cov, err := runPackageTests(p)
				if len(cov) == 0 {
					// we assume the package contains no test.
					// for the total coverage to be meaningful we still have to take
					// the number of statements in this package into account.
					var profileErr error
					cov, profileErr = emptyProfile(p)
					if err == nil && profileErr != nil {
						err = profileErr
					}
				}
				resch <- res{
					out: out,
					cov: cov,
					err: err,
				}
			}
			wg.Done()
		}()
	}
	failed := false
	cov := []*cover.Profile{}
	for r := range resch {
		for _, p := range r.cov {
			cov = addProfile(cov, p)
		}
		if r.err == nil {
			pf(r.out)
		} else {
			pf(r.err.Error())
			failed = true
		}
	}
	return cov, failed
}

func runPackageTests(pkg string) (out string, cov []*cover.Profile, err error) {
	coverFile, err := ioutil.TempFile("", "gotestcover-")
	if err != nil {
		return "", nil, err
	}
	defer coverFile.Close()
	defer os.Remove(coverFile.Name())
	var args []string
	args = append(args, "test")

	if flagA {
		args = append(args, "-a")
	}
	if flagX {
		args = append(args, "-x")
	}
	if flagRace {
		args = append(args, "-race")
	}
	if flagTags != "" {
		args = append(args, "-tags", flagTags)
	}

	if flagV {
		args = append(args, "-v")
	}
	if flagCount != 0 {
		args = append(args, "-count", fmt.Sprint(flagCount))
	}
	if flagCPU != "" {
		args = append(args, "-cpu", flagCPU)
	}
	if flagParallel != "" {
		args = append(args, "-parallel", flagParallel)
	}
	if flagRun != "" {
		args = append(args, "-run", flagRun)
	}
	if flagShort {
		args = append(args, "-short")
	}
	if flagTimeout != "" {
		args = append(args, "-timeout", flagTimeout)
	}
	args = append(args, "-cover")
	if flagCoverMode != "" {
		args = append(args, "-covermode", flagCoverMode)
	}
	args = append(args, "-coverprofile", coverFile.Name())

	args = append(args, pkg)

	args = append(args, argsAfter("--")...)

	cmdOut, err := runGoCommand(args...)
	if err != nil {
		return "", nil, err
	}

	cov, err = cover.ParseProfiles(coverFile.Name())
	if err != nil {
		return "", nil, err
	}

	return string(cmdOut), cov, nil
}

func runGoCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("go", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command %s: %s\n%s", cmd.Args, err, out)
	}
	return out, nil
}

func removeFirstLine(b []byte) []byte {
	out := new(bytes.Buffer)
	sc := bufio.NewScanner(bytes.NewReader(b))
	firstLine := true
	for sc.Scan() {
		if firstLine {
			firstLine = false
			continue
		}
		fmt.Fprintf(out, "%s\n", sc.Bytes())
	}
	return out.Bytes()
}
