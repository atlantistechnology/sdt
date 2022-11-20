package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"

	"github.com/atlantistechnology/sdt/pkg/javascript"
	"github.com/atlantistechnology/sdt/pkg/json_canonical"
	"github.com/atlantistechnology/sdt/pkg/python"
	"github.com/atlantistechnology/sdt/pkg/ruby"
	"github.com/atlantistechnology/sdt/pkg/sql"
	"github.com/atlantistechnology/sdt/pkg/types"
	"github.com/atlantistechnology/sdt/pkg/utils"
)

type gitStatus int8

const (
	Preamble gitStatus = iota
	Staged
	Unstaged
	Untracked
)

func CompareFileType(
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
	// Several other extensions are -sometimes- used for JavaScript
	case ".cjs":
		diffColor.Println(javascript.Diff(filename, options, config))
	case ".mjs":
		diffColor.Println(javascript.Diff(filename, options, config))
	case ".es":
		diffColor.Println(javascript.Diff(filename, options, config))
	case ".es6":
		diffColor.Println(javascript.Diff(filename, options, config))
	case ".json":
		diffColor.Println(json_canonical.Diff(filename, options, config))
	case ".go":
		// TODO: Need to investigate AST tools
		diffColor.Println("| Comparison with Golang syntax tree or canonicalization")
	default:
		diffColor.Println("| No available semantic analyzer for this format")
	}
}

func Compare(
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
			CompareFileType(ext, filename, options, config)
		}
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
		CompareFileType(ext, "", options, config)
	}
}

func ParseGitDiffCompact(diff string, options types.Options, config types.Config) {
	// We wish to sort the changes by their type. The display is a hybrid
	// between `git diff` and `git status`.  Untracked files won't be shown.
	// But for empty destination, the on-disk files will be used as target
	// rather than those alredy committed.
	lines := strings.Split(diff, "\n")
	header := color.New(color.FgWhite, color.Bold)
	newFile := color.New(color.FgGreen)
	delFile := color.New(color.FgRed)
	moveFile := color.New(color.FgMagenta)
	changeFile := color.New(color.FgCyan)
	var changed, added, gone, moved []string

	if len(lines) <= 1 {
		header.Println("No changes detected")
		os.Exit(0)
	} else if options.Status {
		utils.Info("git diff --compact-summary %s %s",
			options.Source, options.Destination)
		fmt.Println(diff)
		os.Exit(0)
	}

	lines = lines[:len(lines)-2] // Do not use summary final line
	// If someone names file perversely, we could mis-identify type in diff
	reStripIndicator := regexp.MustCompile(`(?m) +\| .*$`)
	reStripNew := regexp.MustCompile(` \(new\) *$`)
	reStripGone := regexp.MustCompile(` \(gone\) *$`)
	reMoved := regexp.MustCompile(` => `)

	for _, line := range lines {
		line = reStripIndicator.ReplaceAllString(line, "")
		if reStripNew.MatchString(line) {
			added = append(added, "   "+reStripNew.ReplaceAllString(line, ""))
		} else if reStripGone.MatchString(line) {
			gone = append(gone, "   "+reStripGone.ReplaceAllString(line, ""))
		} else if reMoved.MatchString(line) {
			moved = append(moved, "   "+line)
		} else {
			changed = append(changed, strings.TrimLeft(line, " "))
		}
	}

	if len(added) > 0 {
		header.Println("New files created:")
	}
	for _, filename := range added {
		newFile.Println(filename)
	}

	if len(gone) > 0 {
		header.Println("Files removed from branch/revision:")
	}
	for _, filename := range gone {
		delFile.Println(filename)
	}

	if len(moved) > 0 {
		header.Println("Files moved between branches/revisions:")
	}
	for _, filename := range moved {
		moveFile.Println(filename)
	}

	if len(changed) > 0 {
		if options.Destination != "" {
			header.Println("Changes between branches/revisions:")
		} else {
			header.Println("Changes between branch/revision and current:")
		}
	}
	for _, filename := range changed {
		// Prepare the "local" files being used.  Although one or both files
		// will be revisions rather than local, we save them to tempfiles
		// and use the types.RawNames mode for the comparison.
		var src string
		var tmpfile *os.File
		var body []byte
		var err error
		dst := filename // unless revision was indicated, use the local file

		if options.Destination != "" {
			tmpName := "*-" + strings.ReplaceAll(filename, "/", ":")
			tmpfile, err = os.CreateTemp("", tmpName)
			if err != nil {
				utils.Fail("Could not create temporary destination for %s", filename)
			}
			// Retrieve the HEAD version of the file to a temporary filename
			cmdHead := exec.Command("git", "show", options.Destination+filename)
			body, err = cmdHead.Output()
			if err != nil {
				changeFile.Println("    " + filename)
				continue
			}
			tmpfile.Write(body)
			dst = tmpfile.Name()
			defer os.Remove(tmpfile.Name()) // clean up
		}

		if options.Source != "HEAD:" {
			tmpName := "*-" + strings.ReplaceAll(filename, "/", ":")
			tmpfile, err = os.CreateTemp("", tmpName)
			if err != nil {
				utils.Fail("Could not create temporary source for %s", filename)
			}
			// Retrieve the HEAD version of the file to a temporary filename
			cmdHead := exec.Command("git", "show", options.Source+filename)
			body, err = cmdHead.Output()
			if err != nil {
				changeFile.Println("    " + filename)
				continue
			}
			tmpfile.Write(body)
			src = tmpfile.Name()
			defer os.Remove(tmpfile.Name()) // clean up
		}

		perFileOpts := types.Options{
			Status:      options.Status,
			Semantic:    options.Semantic,
			Parsetree:   options.Parsetree,
			Glob:        options.Glob,
			Verbose:     options.Verbose,
			Dumbterm:    options.Dumbterm,
			Source:      src,
			Destination: dst,
		}
		changeFile.Println("    " + filename)
		Compare("", perFileOpts, config, types.RawNames)
	}
}

func ParseGitStatus(status []byte, options types.Options, config types.Config) {
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
					Compare(line, options, config, types.Status)
				}
			case Unstaged:
				unstaged.Println(fstatus)
				if options.Semantic || options.Parsetree {
					Compare(line, options, config, types.Status)
				}
			case Untracked:
				untracked.Println(fstatus)
			}
		}
	}
}
