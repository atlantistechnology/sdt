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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"

	"github.com/atlantistechnology/sdt/pkg/javascript"
	"github.com/atlantistechnology/sdt/pkg/python"
	"github.com/atlantistechnology/sdt/pkg/ruby"
	"github.com/atlantistechnology/sdt/pkg/sql"
	"github.com/atlantistechnology/sdt/pkg/types"
	"github.com/atlantistechnology/sdt/pkg/utils"
)

const usage = `Usage of Semantic Diff Tool (sdt):

  Names indicated without dashes are subcommands rather than switches.
  If switches are used, they must follow subcommands (if any).

  status, -s      List all analyzable files modified since last git commit
  semantic, -l    List semantically meaningful changes (default viz HEAD:)
  parsetree, -p   Full syntax tree differences (where applicable)
  -g, --glob      Limit compared files by a glob pattern
  -v, --verbose   Show verbose output on STDERR
  -d, --dumbterm  Monochrome/pipe compatible output (also env CI=true)
  -h, --help      Display this help screen

  If not specified, comparisons are between current changes and HEAD.

  -A, --src       File, branch, or revision of source (colon for branch/rev)
  -B, --dst       File, branch, or rev of destination (current if omitted)

  Examples:

    sdt semantic -A 0e904fa3:  # Compare all current files to this revision
    sdt parsetree --src test-branch: --dst HEAD:
    sdt semantic -src my-file.go -dst /path/to/other.go  # Files not in git

`

func compareFileType(
	ext string,
	filename string,
	options types.Options,
	config types.Config,
) {
	diffColor := color.New(color.FgYellow)
	switch ext {
	case ".rb":
		diffColor.Println(ruby.Diff(filename, options, config))
	case ".py":
		diffColor.Println(python.Diff(filename, options, config))
	case ".sql":
		diffColor.Println(sql.Diff(filename, options, config))
	case ".js":
		diffColor.Println(javascript.Diff(filename, options, config))
	case ".go":
		// TODO: Need to investigate AST tools
		diffColor.Println("| Comparison with Golang syntax tree or canonicalization")
	default:
		diffColor.Println("| No available semantic analyzer for this format")
	}
}

func compare(
	line string,
	options types.Options,
	config types.Config,
	lineType types.LineType,
) {
	switch lineType {
	case types.Status:
		info := strings.TrimSpace(line)
		fileLine := strings.SplitN(info, ":   ", 2)
		status := fileLine[0]
		filename := fileLine[1]
		ext := filepath.Ext(line)

		if status == "modified" {
			compareFileType(ext, filename, options, config)
		}
	case types.CompactDiff:
		// TODO
	case types.RawNames:
		ext := filepath.Ext(options.Source)
		ext2 := filepath.Ext(options.Destination)
		if ext != ext2 {
			utils.Info(
				"File extensions mismatch, assuming source type '%s', not '%s'",
				ext, ext2)
		}
		// We allow a slight cleverness of an empty filename meaning that
		// the comparison is between options.Source and options.Destination
		// which will by filepaths not branches/revisions
		compareFileType(ext, "", options, config)
	}
}

type gitStatus int8

const (
	Preamble gitStatus = iota
	Staged
	Unstaged
	Untracked
)

func parseGitStatus(status []byte, options types.Options, config types.Config) {
	var section gitStatus = Preamble
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
			fstatus := strings.Replace(line, "\t", "    ", 1)
			switch section {
			case Staged:
				staged.Println(fstatus)
				if options.Semantic || options.Parsetree {
					compare(line, options, config, types.Status)
				}
			case Unstaged:
				unstaged.Println(fstatus)
				if options.Semantic || options.Parsetree {
					compare(line, options, config, types.Status)
				}
			case Untracked:
				untracked.Println(fstatus)
			}
		}
	}
}

func consistentOptions(options types.Options) string {
	// For now we will only allow the following combinations
	//
	//   sdt <subcommand> -A branch:      # -B omitted
	//   sdt <subcommand> -A revision:    # -B omitted
	//   sdt <subcommand> -A branch/rev: -B branch/rev:
	//   sdt <subcommand> -A local-file1 -B local-file2
	//
	// In the future this might be enhanced to allow direct reference
	// to a particular file in a particular revision; but the logic of
	// which revisions do or don't have a specific file has too many
	// edge cases for now.
	src := options.Source
	dst := options.Destination
	if strings.HasSuffix(src, ":") {
		if dst != "" && !strings.HasSuffix(dst, ":") {
			return "You may only compare a branch/revision with another branch/revision"
		}
	} else {
		if dst == "" || strings.HasSuffix(dst, ":") {
			return "A source of a filepath must be matched by a destination filepath"
		} else {
			if _, err := os.Stat(src); err != nil {
				return "The file " + src + " does not exist!"
			}
			if _, err := os.Stat(dst); err != nil {
				return "The file " + dst + " does not exist!"
			}
		}
	}

	// If no subcommand is given, --src and --dst make no sense
	if !options.Status && !options.Semantic && !options.Parsetree {
		if options.Source != "HEAD:" && options.Destination != "" {
			return "Specifying source or destination is meaningless without a subcommand"
		}

	}

	return "HAPPY"
}

func getOptions() types.Options {
	// Manually pull out "subcommand" since we do not actually want
	// different flags for different subcommands
	subcommand := "FLAGS_ONLY"
	if len(os.Args) == 2 && os.Args[1][0] != '-' {
		// Subcommand only
		subcommand = os.Args[1]
		os.Args = os.Args[:1]
	} else if len(os.Args) > 2 && os.Args[1][0] != '-' {
		// Subcommand and extra flags
		subcommand = os.Args[1]
		// Bad attempt at second subcommand
		if os.Args[2][0] != '-' {
			utils.Fail("Only one subcommand may be specified: \n\t%v", os.Args)
		}
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	// Parse flags and switches provided on command line
	var status bool
	flag.BoolVar(&status, "s", false, "Modified since last git commit")

	var semantic bool
	flag.BoolVar(&semantic, "l", false, "Semantically meaningful changes")

	var glob string
	flag.StringVar(&glob, "glob", "*", "Limit compared files by a glob pattern")
	flag.StringVar(&glob, "g", "*", "Limit compared files by glob (short flag)")

	var parsetree bool
	flag.BoolVar(&parsetree, "p", false, "Full syntax tree differences")

	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "Show verbose output on STDERR")
	flag.BoolVar(&verbose, "v", false, "Show verbose output on STDERR")

	var dumbterm bool
	flag.BoolVar(&dumbterm, "dumbterm", false, "Monochrome/pipe compatible output")
	flag.BoolVar(&dumbterm, "d", false, "Monochrome/pipe compatible output")

	var src string
	flag.StringVar(&src, "src", "HEAD:", "File, branch, or revision of source")
	flag.StringVar(&src, "A", "HEAD:", "File, branch, or revision of source")

	var dst string
	flag.StringVar(&dst, "dst", "", "File, branch, or revision of destination")
	flag.StringVar(&dst, "B", "", "File, branch, or revision of destination")

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	switch subcommand {
	case "status":
		status = true
	case "semantic":
		semantic = true
	case "parsetree":
		parsetree = true
	}

	if os.Getenv("CI") == "true" {
		dumbterm = true
	}

	// Create a struct with the command-line configured options
	return types.Options{
		Status:      status,
		Semantic:    semantic,
		Parsetree:   parsetree,
		Glob:        glob,
		Verbose:     verbose,
		Dumbterm:    dumbterm,
		Source:      src,
		Destination: dst,
	}
}

func main() {
	// Process all flags and subcommands provided
	options := getOptions()
	checkOpts := consistentOptions(options)
	if checkOpts != "HAPPY" {
		utils.Fail(checkOpts)
	}

	// Configure default tools that might be overrridden by the TOML config
	python := types.Command{
		Executable: "python",
		Switches:   []string{"-m", "ast", "-a"},
		Options:    "",
	}
	ruby := types.Command{
		Executable: "ruby",
		Switches:   []string{"--dump=parsetree"},
		Options:    "",
	}
	sql := types.Command{
		Executable: "sqlformat",
		Switches: []string{
			"--reindent_aligned",
			"--identifiers=lower",
			"--strip-comments",
			"--keywords=upper",
		},
		Options: "",
	}
	jsParse := `'const acorn = require("acorn"); ` +
		`const fs = require("fs"); ` +
		`const source = fs.readFileSync("${FILENAME}", "utf8"); ` +
		`const parse = acorn.parse(source, ${OPTIONS}); ` +
		`console.log(JSON.stringify(parse, null, ""));'`
	js := types.Command{
		Executable: "node",
		Switches:   []string{"-e", jsParse},
		Options:    "{sourceType: 'module', ecmaVersion: 'latest'}",
	}

	// Read the configuration file if it is present
	var out []byte

	cfgMessage := "Read $HOME/.sdt.toml for configuration overrides"
	description := "Default commands for each language type"
	configFile := fmt.Sprintf("%s/.sdt.toml", os.Getenv("HOME"))

	var config types.Config
	_, err := toml.DecodeFile(configFile, &config)

	if err != nil {
		cfgMessage = "No $HOME/.sdt.toml file, using built-in defaults"
	} else {
		// Glob can be defined twice, but command-line rules when different
		if config.Glob != "" && options.Glob == "*" {
			options.Glob = config.Glob
		}
		// $HOME/.sdt.toml may override default description if present
		if config.Description != "" {
			description = config.Description
		}
		// We might override default programming language commands
		if userpython, found := config.Commands["python"]; found {
			python = userpython
		}
		if userruby, found := config.Commands["ruby"]; found {
			ruby = userruby
		}
		if usersql, found := config.Commands["sql"]; found {
			sql = usersql
		}
		if userjs, found := config.Commands["javascript"]; found {
			js = userjs
		}
	}
	// Create userCfg with possibly changed values for Commands
	userCfg := types.Config{
		Description: description,
		Glob:        config.Glob,
		Commands: map[string]types.Command{
			"python":     python,
			"ruby":       ruby,
			"sql":        sql,
			"javascript": js,
		},
	}

	// The call to consistentOptions() has already ruled out cases that are
	// generally impermissible. This limits the if predicates needed here.
	if options.Status || options.Semantic || options.Parsetree {
		if strings.HasSuffix(options.Destination, ":") {
			// Handle case of two branches/revisions given for -A/-B
			cmd := exec.Command("git", "diff", "--compact-summary",
				options.Source, options.Destination,
			)
			out, err = cmd.Output()
			if err != nil {
				utils.Fail(
					"One or both revisions are unavailable: %s, %s",
					options.Source, options.Destination)
			}
			utils.Info(
				"Only committed files will be included in comparison to branch")
			fmt.Println("XXX\n" + string(out))
		} else if options.Source != "HEAD:" && options.Destination == "" {
			// Handle case of comparing source branch/revision to dest HEAD
			cmd := exec.Command("git", "diff", "--compact-summary",
				options.Source, "HEAD:",
			)
			out, err = cmd.Output()
			if err != nil {
				utils.Fail(
					"The indicate source branch/revision is unavailable: %s",
					options.Source)
			}
			utils.Info(
				"Only committed files will be included in comparison to branch")
			fmt.Println("XXX\n" + string(out))
		} else if options.Destination != "" {
			// Handle the case of comparing two local files
			// ...which were verified as existing in an earlier check
			compare("", options, userCfg, types.RawNames)
		} else {
			// Handle default case of comparing HEAD to current files
			cmd := exec.Command("git", "status")
			out, err = cmd.Output()
			if err != nil {
				utils.Fail("%s %s", err, "(probably not in a git directory)")
			}
			parseGitStatus(out, options, userCfg)
		}
	}

	if options.Verbose {
		fmt.Fprintf(os.Stderr, "Description: %s\n", description)
		fmt.Fprintf(os.Stderr, "Config: %s\n", cfgMessage)
		fmt.Fprintf(os.Stderr, "status: %t\n", options.Status)
		fmt.Fprintf(os.Stderr, "semantic: %t\n", options.Semantic)
		fmt.Fprintf(os.Stderr, "parsetree: %t\n", options.Parsetree)
		fmt.Fprintf(os.Stderr, "glob: %s\n", options.Glob)
		fmt.Fprintf(os.Stderr, "source: %s\n", options.Source)
		fmt.Fprintf(os.Stderr, "destination: %s\n", options.Destination)
		fmt.Fprintf(os.Stderr, "dumbterm: %t\n", options.Dumbterm)
		fmt.Fprintf(os.Stderr, "---\n")
		fmt.Fprintf(os.Stderr, "python: %s %s\n", python.Executable, python.Switches)
		fmt.Fprintf(os.Stderr, "ruby: %s %s\n", ruby.Executable, ruby.Switches)
		fmt.Fprintf(os.Stderr, "sql: %s\n  %s\n", sql.Executable, sql.Switches)
		fmt.Fprintf(os.Stderr, "javascript: %s\n  %s\n  %s\n",
			js.Executable, js.Switches, js.Options)
	}
}
