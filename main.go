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
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/BurntSushi/toml"
    "github.com/fatih/color"

    "github.com/atlantistechnology/sdt/pkg/ruby"
    "github.com/atlantistechnology/sdt/pkg/sql"
    "github.com/atlantistechnology/sdt/pkg/types"
)

const usage = `Usage of Semantic Diff Tool (sdt):

  Names indicated without dashes are subcommands rather than switches

  status, -s      List all analyzable files modified since last git commit
  semantic, -l    List semantically meaningful changes since last git commit
  parsetree, -p   Full syntax tree differences (where applicable)
  -g, --glob      Limit compared files by a glob pattern
  -v, --verbose   Show verbose output on STDERR
  -h, --help      Display this help screen

  If not specified, comparisons are between current changes and HEAD.

  -A, --src       File, branch, or revision of source (colon required for revision)
  -B, --dst       File, branch, or revision of destination (current if omitted)

  Examples:

    sdt semantic -A 0e904fa3:  # Compare all current files to this revision
    sdt parsetree --src test-branch: --dst HEAD:
    sdt semantic -src my-file.go -dst /path/to/other.go  # Files not in git

`

func ASTCompare(line string, options types.Options, config types.Config) {
    info := strings.TrimSpace(line)
    fileLine := strings.SplitN(info, ":   ", 2)
    status := fileLine[0]
    filename := fileLine[1]
    ext := filepath.Ext(line)
    diffColor := color.New(color.FgYellow)

    if status == "modified" {
        switch ext {
        case ".rb":
            diffColor.Println(ruby.Diff(filename, options, config))
        case ".py":
            // Something with `ast` module
            diffColor.Println("| Comparison of Python ASTs")
        case ".sql":
            // sqlformat --reindent_aligned --identifiers lower --strip-comments --keywords upper
            diffColor.Println(sql.Diff(filename, options, config))
        case ".js":
            // Probably eslint parsing
            diffColor.Println("| Comparison with JS syntax tree")
        case ".go":
            // TODO: Need to investigate AST tools
            diffColor.Println("| Comparison with Golang syntax tree or canonicalization")
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

func ParseGitStatus(status []byte, options types.Options, config types.Config) {
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
                if options.Semantic || options.Parsetree {
                    ASTCompare(line, options, config)
                }
            case Unstaged:
                unstaged.Println(fstatus)
                if options.Semantic || options.Parsetree {
                    ASTCompare(line, options, config)
                }
            case Untracked:
                untracked.Println(fstatus)
            }
        }
    }
}

func main() {
    // Parse flags and switches provided on command line
    var status bool
    flag.BoolVar(&status, "s", false, "Modified since last git commit")

    var semantic bool
    flag.BoolVar(&semantic, "l", false, "Semantically meaningful changes")

    var glob string
    flag.StringVar(&glob, "glob", "*.*", "Limit compared files by a glob pattern")
    flag.StringVar(&glob, "g", "*.*", "Limit compared files by glob (short flag)")

    var parsetree bool
    flag.BoolVar(&parsetree, "p", false, "Full syntax tree differences")

    var verbose bool
    flag.BoolVar(&verbose, "verbose", false, "Show verbose output on STDERR")
    flag.BoolVar(&verbose, "v", false, "Show verbose output on STDERR")

    flag.Usage = func() { fmt.Print(usage) }

    if len(os.Args) > 1 {
        switch os.Args[1] {
        case "status":
            status = true
        case "semantic":
            semantic = true
        case "parsetree":
            parsetree = true
        }
    }
    flag.Parse()

    options := types.Options{
        Status:    status,
        Semantic:  semantic,
        Glob:      glob,
        Verbose:   verbose,
        Parsetree: parsetree,
    }

    // Configure default tools that might be overrridden by the TOML config
    description := "Default commands for each language type"
    pythonCmd := types.Command{
        Executable: "python",
        Switches:   []string{"-m", "ast", "-a"},
    }
    rubyCmd := types.Command{
        Executable: "ruby",
        Switches:   []string{"--dump=parsetree"},
    }
    sqlCmd := types.Command{
        Executable: "sqlformat",
        Switches: []string{
            "--reindent_aligned",
            "--identifiers=lower",
            "--strip-comments",
            "--keywords=upper",
        },
    }

    // Read the configuration file if it is present
    var out []byte
    var err error

    configFile := fmt.Sprintf("%s/.sdt.toml", os.Getenv("HOME"))
    var config types.Config
    _, err = toml.DecodeFile(configFile, &config)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    } else {
        if config.Glob != "" {
            glob = config.Glob
        }
        if config.Description != "" {
            description = config.Description
        }
        if userpython, found := config.Commands["python"]; found {
            pythonCmd = userpython
        }
        if userruby, found := config.Commands["ruby"]; found {
            rubyCmd = userruby
        }
        if usersql, found := config.Commands["sql"]; found {
            sqlCmd = usersql
        }
    }
    // Create userCfg with possibly changed values for Commands
    userCfg := types.Config{
        Description: description,
        Glob:        glob,
        Commands: map[string]types.Command{
            "python": pythonCmd,
            "ruby":   rubyCmd,
            "sql":    sqlCmd,
        },
    }

    if status || semantic || parsetree {
        cmd := exec.Command("git", "status")
        out, err = cmd.Output()
        if err != nil {
            log.Fatal(err)
        }
        ParseGitStatus(out, options, userCfg)
    }

    if verbose {
        fmt.Fprintf(os.Stderr, "Description: %s\n", description)
        fmt.Fprintf(os.Stderr, "status: %t\n", status)
        fmt.Fprintf(os.Stderr, "semantic: %t\n", semantic)
        fmt.Fprintf(os.Stderr, "parsetree: %t\n", parsetree)
        fmt.Fprintf(os.Stderr, "glob: %s\n", glob)
        fmt.Fprintf(os.Stderr, "python: %s %s\n", pythonCmd.Executable, pythonCmd.Switches)
        fmt.Fprintf(os.Stderr, "ruby: %s %s\n", rubyCmd.Executable, rubyCmd.Switches)
        fmt.Fprintf(os.Stderr, "sql: %s %s\n", sqlCmd.Executable, sqlCmd.Switches)
    }
}
