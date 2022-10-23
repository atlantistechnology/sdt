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
	"github.com/atlantistechnology/ast-diff/pkg/ruby"
	"github.com/fatih/color"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Options struct {
	status    bool
	semantic  bool
	glob      string
	verbose   bool
	parsetree bool
}

const usage = `Usage of ast-dff:
  -s, --status     List all analyzable files modified since last git commit
  -l, --semantic   List semantically meaningful changes since last git commit
  -g, --glob       Limit compared files by a glob pattern
  -p, --parsetree  Full syntax tree differences
  -v, --verbose    Show verbose output on STDERR
  -h, --help       Display this help screen
`

func ASTCompare(line string, options Options) {
	info := strings.TrimSpace(line)
	fileLine := strings.SplitN(info, ":   ", 2)
	status := fileLine[0]
	filename := fileLine[1]
	ext := filepath.Ext(line)
	diffColor := color.New(color.FgYellow)

	if status == "modified" {
		switch ext {
		case ".rb":
			diffColor.Println(ruby.Diff(filename, options.semantic, options.parsetree))
		case ".py":
			// Something with `ast` module
			diffColor.Println("| Comparison of Python ASTs")
		case ".sql":
			// Dunno, find a nice parser
			diffColor.Println("| Comparison with SQL canonicalizer")
		case ".js":
			// Probably eslint parsing
			diffColor.Println("| Comparison with JS syntax tree")
		default:
			diffColor.Println("| No available semantic analyzer for this format")
		}
	}
}

type GitStatus int8

const (
	Preamble GitStatus = iota
	Staged
	Unstaged
	Untracked
)

func ParseGitStatus(status []byte, options Options) {
	var section GitStatus = Preamble
	lines := bytes.Split(status, []byte("\n"))

	header := color.New(color.FgWhite, color.Bold)
	staged := color.New(color.FgGreen)
	unstaged := color.New(color.FgRed)
	untracked := color.New(color.FgCyan)

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
			fstatus := strings.Replace(line, "\t", "  ", 1)
			switch section {
			case Staged:
				staged.Println(fstatus)
				if options.semantic || options.parsetree {
					ASTCompare(line, options)
				}
			case Unstaged:
				unstaged.Println(fstatus)
				if options.semantic || options.parsetree {
					ASTCompare(line, options)
				}
			case Untracked:
				untracked.Println(fstatus)
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

	var parsetree bool
	flag.BoolVar(&parsetree, "parsetree", false, "Full syntax tree differences")
	flag.BoolVar(&parsetree, "p", false, "Full syntax tree differences")

	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "Show verbose output on STDERR")
	flag.BoolVar(&verbose, "v", false, "Show verbose output on STDERR")

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	options := Options{
		status:    status,
		semantic:  semantic,
		glob:      glob,
		verbose:   verbose,
		parsetree: parsetree,
	}

	var out []byte
	var err error
	if status || semantic || parsetree {
		cmd := exec.Command("git", "status")
		out, err = cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
		ParseGitStatus(out, options)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "status: %t\n", status)
		fmt.Fprintf(os.Stderr, "semantic: %t\n", semantic)
		fmt.Fprintf(os.Stderr, "parsetree: %t\n", parsetree)
		fmt.Fprintf(os.Stderr, "glob: %s\n", glob)
	}
}
