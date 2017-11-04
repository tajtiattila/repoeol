// +build none
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"

	"github.com/pkg/errors"
)

type File interface {
	Name() string
	Hash() string
	Open() (io.ReadCloser, error)
}

func IndexFilesAdded() ([]File, error) {
	ref, err := RepoHead()
	if err != nil {
		return nil, err
	}

	o, err := exec.Command("git", "diff-index", "--cached", ref).Output()
	if err != nil {
		return nil, err
	}

	vdl, err := parseDiffLines(o, false)
	if err != nil {
		return nil, err
	}

	var v []File
	for _, dl := range vdl {
		if s := dl.status; s == 'M' || s == 'C' || s == 'R' || s == 'A' {
			v = append(v, &gitFile{dl.path(), dl.dstHash})
		}
	}

	return v, nil
}

type gitFile struct {
	name string
	hash string
}

func (f *gitFile) Name() string { return f.name }
func (f *gitFile) Hash() string { return f.hash }

func (f *gitFile) Open() (io.ReadCloser, error) {
	o, err := exec.Command("git", "cat-file", "blob", f.hash).Output()
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(o)), nil
}

type diffLine struct {
	srcMode string
	dstMode string
	srcHash string
	dstHash string
	status  byte
	score   int
	srcPath string
	dstPath string
}

func (l *diffLine) path() string {
	if l.dstPath != "" {
		return l.dstPath
	}
	return l.srcPath
}

func parseDiffLines(p []byte, zflag bool) ([]diffLine, error) {
	var v []diffLine
	parser := diffLineParser{p: p}
	var fsep, lsep byte
	if zflag {
		fsep, lsep = 0, 0
	} else {
		fsep, lsep = '\t', '\n'
	}
	for !parser.done() {
		var l diffLine
		parser.accept(':')
		l.srcMode = parser.word(' ')
		l.dstMode = parser.word(' ')
		l.srcHash = parser.word(' ')
		l.dstHash = parser.word(' ')
		l.status = parser.byte()
		score := parser.word(fsep)
		if parser.err != nil {
			break
		}
		if score != "" {
			l.score, parser.err = strconv.Atoi(score)
		}
		if l.status == 'C' || l.status == 'R' {
			l.srcPath = parser.word(fsep)
			l.dstPath = parser.word(lsep)
		} else {
			l.srcPath = parser.word(lsep)
		}
		v = append(v, l)
	}
	return v, parser.err
}

type diffLineParser struct {
	p   []byte
	i   int
	err error
}

func (p *diffLineParser) done() bool {
	if p.err != nil {
		return true
	}
	return p.i == len(p.p)
}

func (p *diffLineParser) byte() byte {
	if p.err != nil {
		return 0
	}
	if p.i < len(p.p) {
		b := p.p[p.i]
		p.i++
		return b
	}
	p.err = errors.New("Unexpected EOF")
	return 0
}

func (p *diffLineParser) accept(ch byte) bool {
	if p.err != nil {
		return false
	}
	if p.i < len(p.p) && p.p[p.i] == ch {
		p.i++
		return true
	}
	p.err = errors.Errorf("Missing '%s' at position %d", string(ch), p.i)
	return false
}

func (p *diffLineParser) word(sep byte) string {
	if p.err != nil {
		return ""
	}
	start := p.i
	for p.i < len(p.p) && p.p[p.i] != sep {
		p.i++
	}
	end := p.i
	if p.i == len(p.p) || p.p[p.i] != sep {
		var chname string
		switch {
		case sep == 0:
			chname = "NUL"
		case sep == 7:
			chname = "TAB"
		case sep < 32:
			chname = fmt.Sprintf("'\\x%02x'", sep)
		default:
			chname = fmt.Sprintf("'%s'", string(sep))
		}
		p.err = errors.Errorf("Missing %s at position %d", chname, p.i)
	} else {
		p.i++
	}
	return string(p.p[start:end])
}

func RepoHead() (string, error) {
	err := exec.Command("git", "rev-parse", "--verify", "HEAD").Run()
	if _, ok := err.(*exec.ExitError); ok {
		return "4b825dc642cb6eb9a060e54bf8d69288fbee4904", nil // empty tree
	}
	if err != nil {
		return "", err
	}
	return "HEAD", nil
}
