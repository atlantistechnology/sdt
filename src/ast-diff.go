/* Created by David Mertz

Examine changed files within git revisions, and provide guidance on whether
such changes are likely to represent semantic differences or merely stylistic
changes.

This program will operate by calling the "native" parsers of various
programming languages, or failing that widely available parsers and grammars
used with those languages.
*/
package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"log"
	"os"
	"os/exec"
	"strings"
)

const usage = `Usage of ast-dff:
  -s, --status   List all analyzable files modified since last git commit
  -l, --semantic List semantically meaningful changes since last git commit
  -g, --glob     Limit compared files by a glob pattern
  -v, --verbose  Show verbose output on STDERR
  -h, --help     Display this help screen
`

type GitStatus int8

const (
	Preamble GitStatus = iota
	Staged
	Unstaged
	Untracked
)

func ParseGitStatus(status []byte, semantic bool) {
	lines := bytes.Split(status, []byte("\n"))
	var section GitStatus = Preamble
	header := color.New(color.FgWhite, color.Bold)
	untracked := color.New(color.FgRed).Add(color.Underline)
	for i := 0; i < len(lines); i++ {
		line := string(lines[i])
		if strings.HasPrefix(line, "Changes to be committed") {
			section = Staged
			header.Println(line)
		} else if strings.HasPrefix(line, "Changes not staged for commit") {
			section = Unstaged
			header.Println(line)
		} else if strings.HasPrefix(line, "Untracked files") {
			section = Untracked
			header.Println(line)
		}

		if strings.HasPrefix(line, "\t") {
			switch section {
			case Staged:
				color.Green(line)
				if semantic {
					color.Cyan("... Actual AST comparison here ...")
				}
			case Unstaged:
				color.Red(line)
				if semantic {
					color.Cyan("... Actual AST comparison here ...")
				}
			case Untracked:
				untracked.Println(line)
			}
		}
	}
}

func main() {
	var status bool
	flag.BoolVar(&status, "status", false, "Modified since last git commit")
	flag.BoolVar(&status, "s", false, "Modified since last git commit")

	var semantic bool
	flag.BoolVar(&semantic, "semantic", false, "Semantically meaningful changes")
	flag.BoolVar(&semantic, "l", false, "Semantically meaningful changes")

	var glob string
	flag.StringVar(&glob, "glob", "*", "Limit compared files by a glob pattern")
	flag.StringVar(&glob, "g", "*", "Limit compared files by glob (short flag)")

	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "Show verbose output on STDERR")
	flag.BoolVar(&verbose, "v", false, "Show verbose output on STDERR")

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	var out []byte
	var err error
	if status || semantic {
		cmd := exec.Command("git", "status")
		out, err = cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
		ParseGitStatus(out, semantic)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "status: %t\n", status)
		fmt.Fprintf(os.Stderr, "semantic: %t\n", semantic)
		fmt.Fprintf(os.Stderr, "glob: %s\n", glob)
	}
}
