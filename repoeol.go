package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var verbose, cs bool
	var crlfs, lfs string
	flag.BoolVar(&verbose, "v", false, "list all file statuses")
	flag.BoolVar(&cs, "s", false, "case sensitive extension checking")
	flag.StringVar(&crlfs, "crlf", "", "comma separated list of file name extensions that must have crlf")
	flag.StringVar(&lfs, "lf", "", "comma separated list of file name extensions that must have lf")
	flag.Parse()

	csf := func(s string) string {
		if cs {
			return strings.ToLower(s)
		}
		return s
	}

	xs := []extStat{
		{splitExts(csf(crlfs)), EOLStat.IsCRLF},
		{splitExts(csf(lfs)), EOLStat.IsLF},
	}

	files, err := IndexFilesAdded()
	chk(err)

	nerr := 0
	for _, f := range files {
		fs, err := fileEOLStat(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			nerr++
			continue
		}
		bad := fs.NUL == 0 && (fs.IsMixed() ||
			!checkEOLStat(csf(f.Name()), fs, xs))
		if bad {
			nerr++
		}
		if bad || verbose {
			fmt.Fprintf(os.Stderr, "%s: %s\n", f.Name(), fs)
		}
	}
	if nerr != 0 {
		fmt.Fprintln(os.Stderr, nerr, "error(s)")
		os.Exit(1)
	}
}

func chk(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type extStat struct {
	ext []string // extensions f must return true for
	f   func(EOLStat) bool
}

func checkEOLStat(fn string, s EOLStat, v []extStat) bool {
	for _, es := range v {
		if extIn(fn, es.ext) && !es.f(s) {
			return false
		}
	}
	return true
}

func fileEOLStat(f File) (EOLStat, error) {
	rc, err := f.Open()
	if err != nil {
		return EOLStat{}, err
	}
	defer rc.Close()
	return CalcEOLStatReader(rc)
}

func extIn(s string, vx []string) bool {
	ext := filepath.Ext(s)
	if ext == s {
		return false
	}
	for _, x := range vx {
		if x == ext {
			return true
		}
	}
	return false
}

func splitExts(s string) []string {
	v := strings.Split(s, ",")
	for i, x := range v {
		if !strings.HasPrefix(x, ".") {
			v[i] = "." + x
		}
	}
	return v
}

func timerFunc() func(what string) {
	t0 := time.Now()
	return func(what string) {
		t1 := time.Now()
		delta := t1.Sub(t0)
		t0 = t1
		fmt.Printf("%s: %v\n", what, delta)
	}
}
