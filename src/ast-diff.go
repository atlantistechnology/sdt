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
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

const usage = `Usage of ast-dff:
  -s, --status   List all analyzable files modified since last git commit
  -l, --semantic List semantically meaningful changes since last git commit
  -g, --glob     Limit compared files by a glob pattern
  -v, --verbose  Show verbose output on STDERR
`

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
	}

	if semantic {
		// Do the actual semantic diff
	}

	fmt.Fprintf(os.Stdout, "%s\n", out)

	fmt.Fprintf(os.Stderr, "status: %t\n", status)
	fmt.Fprintf(os.Stderr, "semantic: %t\n", semantic)
	fmt.Fprintf(os.Stderr, "glob: %s\n", glob)
}
