package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func main() {
	path := flag.String("repo", ".", "repository path")
	flag.Parse()

	repo, err := git.PlainOpen(*path)
	chk(err)

	c, err := repo.CommitObject(plumbing.NewHash(flag.Arg(0)))
	chk(err)

	ti, err := c.Files()
	chk(err)

	err = ti.ForEach(func(f *object.File) error {
		fmt.Println(f.Name, f.Hash)
		return nil
	})
	chk(err)

}

func chk(err error) {
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
