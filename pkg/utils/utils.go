package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"golang.org/x/exp/constraints"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/atlantistechnology/sdt/pkg/types"
)

func Fail(msg string, params ...interface{}) {
	msg = types.Colors.Info + "ERROR: " + types.Colors.Clear + msg + "\n"
	fmt.Fprintf(os.Stderr, msg, params...)
	os.Exit(-1)
}

func Info(msg string, params ...interface{}) {
	msg = types.Colors.Info + "INFO: " + types.Colors.Clear + msg + "\n"
	fmt.Fprintf(os.Stderr, msg, params...)
}

type LineOffset struct {
	Start uint32
	End   uint32
	Line  string
}

// We assume that lines are sensibly split on LF, not CR or CRLF
func MakeOffsetsFromString(text string) []LineOffset {
	lines := strings.Split(text, "\n")
	var records []LineOffset
	var start uint32 = 0
	var length uint32

	for i := 0; i < len(lines); i++ {
		length = uint32(len(lines[i]) + 1) // Add back striped LF
		record := LineOffset{Start: start, End: start + length, Line: lines[i]}
		records = append(records, record)
		start += length
	}
	return records
}

// We assume that lines are sensibly split on LF, not CR or CRLF
// Also assume that bytes encode text as UTF-8 not any odd encoding
func MakeOffsetsFromByteArray(text []byte) []LineOffset {
	lines := bytes.Split(text, []byte("\n"))
	var records []LineOffset
	var start uint32 = 0
	var length uint32

	for i := 0; i < len(lines); i++ {
		length = uint32(len(lines[i]) + 1) // Add back stripped LF
		record := LineOffset{Start: start, End: start + length, Line: string(lines[i])}
		records = append(records, record)
		start += length
	}
	return records
}

// Return the line number identified.  Use -1 as sentinel for "not found"
// TODO: if we care about speed, we can do a bisection search of the
// well-ordered {start, end, line} structures
func LineAtPosition(records []LineOffset, pos uint32) int {
	for lineNo, record := range records {
		if pos >= record.Start && pos < record.End {
			return lineNo
		}
	}
	return -1
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func BufferToDiff(buff bytes.Buffer, colorLeft bool, dumbterm bool) string {
	if dumbterm {
		// This can possibly be made slightly less special-case
		ret := buff.String()
		reCleanupDumbterm := regexp.MustCompile(`(?m){{_}}`)
		ret = reCleanupDumbterm.ReplaceAllString(ret, "")
		reCleanDumb2 := regexp.MustCompile(`(?Us)({{[+-])([[:space:]]+)(}})`)
		ret = reCleanDumb2.ReplaceAllString(ret, "$2")
		rePrepend := regexp.MustCompile(`(?m)^`)
		ret = rePrepend.ReplaceAllString(ret,
			types.Colors.Header+"| "+types.Colors.Clear)
		return ret
	}

	ret := buff.String()
	rePrepend := regexp.MustCompile(`(?m)^`)
	if colorLeft {
		colorPipe := types.Colors.Header + "| " + types.Colors.Clear
		ret = rePrepend.ReplaceAllString(ret, colorPipe)
	} else {
		ret = rePrepend.ReplaceAllString(ret, "| ")
	}
	return ret
}

func SemanticChanges(
	dmp *diffmatchpatch.DiffMatchPatch,
	diffs []diffmatchpatch.Diff,
	filename string,
	headTree []byte,
	headTreeString string,
	parseType types.ParseType,
	dumbterm bool) string {

	var gitDiff []byte
	var err error
	var file1, file2 string

	// In arguably too much cleverness, a filename like "A -> B" marks
	// comparison of two local files rather than to git branch/revision
	reLocalFiles := regexp.MustCompile(` -> `)
	if reLocalFiles.MatchString(filename) {
		fileNames := strings.Split(filename, " -> ")
		file1 = fileNames[0]
		file2 = fileNames[1]
		// System `diff` exits with non-zero status 1 for diff found
		cmdDiff := exec.Command("diff", "-u", file1, file2)
		// Misnomer of `gitDiff`, but we keep it consistent with other uses
		gitDiff, err = cmdDiff.Output()
		if err != nil && err.Error() != "exit status 1" {
			Fail("Could not perform local diff on %s -> %s", file1, file2)
		}
	} else {
		// What git thinks has changed in actual source since last push
		file1 = filename
		cmdGitDiff := exec.Command("git", "diff", filename)
		gitDiff, err = cmdGitDiff.Output()
		if err != nil {
			Fail("Could not peform git diff against local files")
		}
	}

	// Determine the changes to the respective parse trees
	patch := dmp.PatchToText(dmp.PatchMake(diffs))

	// Only interested in line offsets of the change in parse tree
	reFromTo := regexp.MustCompile(`(?m)^[^@].*$[\r\n]*`)
	ranges := reFromTo.ReplaceAllString(patch, "")
	rangeLines := strings.Split(ranges, "\n")

	var oldStart uint32
	var oldCount uint32
	var newStart uint32
	var newCount uint32
	var lineOfInterest uint32
	var posOfInterest uint32
	var n int

	offsets := MakeOffsetsFromString(headTreeString)
	treeLines := bytes.Split(headTree, []byte("\n"))

	// Successive diffs will add or remove characters
	adjustment := 0
	diffLines := mapset.NewSet[uint32]()
	diffPositions := mapset.NewSet[uint32]()
	// Create set under assumption we'll need to put more in it
	posNotLine := mapset.NewSet[types.ParseType]()
	posNotLine.Add(types.JavaScript)

	for i := 0; i < len(rangeLines); i++ {
		n, _ = fmt.Sscanf(
			rangeLines[i],
			"@@ -%d,%d +%d,%d @@",
			&oldStart, &oldCount, &newStart, &newCount,
		)
		if n == 4 {
			var m int

			tweakedPosition := uint32(int(oldStart) + adjustment)
			adjustment += int(oldCount) - int(newCount)
			parseTreeLineNum := LineAtPosition(offsets, tweakedPosition)

			// Try a few lines before and after the line found for underlying
			// source code position (possible false positives aren't so important)
			// Parse trees tend to have a lot of lines, so looking at 8 of them
			// is not all that likely to grab a lot that is not in the same
			// diff segment.  If it does, the developer can reject the false hit.
			minLine := Max(parseTreeLineNum-4, 0)
			maxLine := Min(parseTreeLineNum+4, len(treeLines))
			switch parseType {
			case types.Ruby:
				for j := minLine; j < maxLine; j++ {
					line := string(treeLines[j])
					lines := strings.Split(line, "(")
					if len(lines) > 1 {
						m, _ = fmt.Sscanf(lines[1], "line: %d", &lineOfInterest)
						if m == 1 {
							diffLines.Add(lineOfInterest)
						}
					}
				}
			case types.Python:
				for j := minLine; j < maxLine; j++ {
					line := string(treeLines[j])
					reLeadingSpace := regexp.MustCompile("^[\n\r\t ]+")
					line = reLeadingSpace.ReplaceAllString(line, "")
					m, _ = fmt.Sscanf(line, "lineno=%d", &lineOfInterest)
					if m == 1 {
						diffLines.Add(lineOfInterest)
					}
					m, _ = fmt.Sscanf(line, "end_lineno=%d", &lineOfInterest)
					if m == 1 {
						diffLines.Add(lineOfInterest)
					}
				}
			case types.JavaScript:
				for j := minLine; j < maxLine; j++ {
					line := string(treeLines[j])
					reLeadingSpace := regexp.MustCompile("^[\n\r\t ]+")
					line = reLeadingSpace.ReplaceAllString(line, "")
					m, _ = fmt.Sscanf(line, `"start": %d"`, &posOfInterest)
					if m == 1 {
						diffPositions.Add(posOfInterest)
					}
					m, _ = fmt.Sscanf(line, `"end": %d"`, &posOfInterest)
					if m == 1 {
						diffPositions.Add(posOfInterest)
					}
				}
			}
		}
	}

	// Some parse types have byte postions in original file, not lines numbers
	if posNotLine.Contains(parseType) {
		data, err := os.ReadFile(file1)
		if err != nil {
			Fail("Unable to read local file %s", filename)
		}
		lineOffsets := MakeOffsetsFromByteArray(data)
		for pos := range diffPositions.Iterator().C {
			diffLines.Add(uint32(LineAtPosition(lineOffsets, pos)))
		}
	}

	if len(ranges) > 0 {
		changedSegments := changedGitSegments(gitDiff, diffLines, dumbterm)
		return changedSegments
	} else {
		return "| No semantic differences detected"
	}
}

// ColorDiff converts (DiffMatchPatch, []Diff) into colored text report
func ColorDiff(
	dmp *diffmatchpatch.DiffMatchPatch,
	diffs []diffmatchpatch.Diff,
	parseType types.ParseType,
	dumbterm bool) string {

	var highlights types.Highlights
	if dumbterm {
		highlights = types.Dumbterm
	} else {
		highlights = types.Colors
	}

	var buff bytes.Buffer
	var transforms []regexp.Regexp
	switch parseType {
	case types.Ruby:
		reComment := regexp.MustCompile(`(?m)^##.*$[\r\n]*`)
		reTreeClean := regexp.MustCompile(`(?m)(\| |\+-)`)
		transforms = append(transforms, *reComment, *reTreeClean)
	case types.Python:
		reLineno := regexp.MustCompile(`(?m)^\s*lineno=\?,$[\r\n]*`)
		reEndlineno := regexp.MustCompile(`(?m)^\s*end_lineno=\?,$[\r\n]*`)
		reColoffset := regexp.MustCompile(`(?m)^\s*col_offset=\?,$[\r\n]*`)
		reEndcoloffset := regexp.MustCompile(`(?m)   end_col_offset=\?`)
		transforms = append(transforms,
			*reLineno, *reEndlineno, *reColoffset, *reEndcoloffset)
	case types.JavaScript:
		reStart := regexp.MustCompile(`(?m)^\s*"start": \?,$[\r\n]*`)
		reEnd := regexp.MustCompile(`(?m)^\s*"end": \?,$[\r\n]*`)
		reBraceOnly := regexp.MustCompile(`(?m)^\s*[\]}],?$[\r\n]*`)
		rePunct := regexp.MustCompile(`[\[{,"]`)
		reBlankln := regexp.MustCompile(`(?m)^\s*$[\r\n]*`)
		transforms = append(transforms,
			*reStart, *reEnd, *reBraceOnly, *rePunct, *reBlankln)
	}

	buff.WriteString("Comparison of parse trees or canonical format\n")

	changed := false
	for _, diff := range diffs {
		text := diff.Text
		for _, transform := range transforms {
			text = transform.ReplaceAllString(text, "")
		}

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			changed = true
			buff.WriteString(highlights.Add)
			buff.WriteString(text)
			buff.WriteString(highlights.Clear)
		case diffmatchpatch.DiffDelete:
			changed = true
			buff.WriteString(highlights.Del)
			buff.WriteString(text)
			buff.WriteString(highlights.Clear)
		case diffmatchpatch.DiffEqual:
			buff.WriteString(highlights.Neutral)
			buff.WriteString(highlights.Clear)
			buff.WriteString(text)
		}
	}
	if changed {
		return BufferToDiff(buff, false, dumbterm)
	}

	return "| No semantic differences detected"
}

func changedGitSegments(
	gitDiff []byte,
	diffLines mapset.Set[uint32],
	dumbterm bool) string {

	var highlights types.Highlights
	if dumbterm {
		highlights = types.PlainASCII
	} else {
		highlights = types.Colors
	}

	var buff bytes.Buffer
	lines := bytes.Split(gitDiff, []byte("\n"))
	showSegment := false
	var oldStart uint32
	var oldCount uint32
	var newStart uint32
	var newCount uint32

	buff.WriteString(highlights.Header)
	buff.WriteString("Segments with likely semantic changes\n")

	for _, line := range lines {
		n, _ := fmt.Sscanf(
			string(line),
			"@@ -%d,%d +%d,%d @@",
			&oldStart, &oldCount, &newStart, &newCount,
		)

		if n == 4 {
			minLine := Min(oldStart, newStart)
			maxLine := Max(oldStart+oldCount, newStart+newCount)
			showSegment = false
			for i := minLine; i <= maxLine; i++ {
				if diffLines.Contains(i) {
					showSegment = true
					break
				}
			}
		}

		if showSegment {
			prefix := byte(' ')
			if len(line) > 0 {
				prefix = line[0]
			}
			switch prefix {
			case byte('@'):
				buff.WriteString(highlights.Info)
				buff.Write(line)
				buff.WriteString(highlights.Clear)
			case byte('+'):
				buff.WriteString(highlights.Add)
				buff.Write(line)
				buff.WriteString(highlights.Clear)
			case byte('-'):
				buff.WriteString(highlights.Del)
				buff.Write(line)
				buff.WriteString(highlights.Clear)
			default:
				buff.Write(line)
			}
			buff.WriteString("\n")
		}
	}
	return BufferToDiff(buff, true, dumbterm)
}

func LocalFileTrees(
	cmd string,
	switches []string,
	options types.Options,
	langName string,
	canonical bool) (string, []byte, []byte) {

	var headTree []byte
	var currentTree []byte
	var err error
	filename := options.Source + " -> " + options.Destination

	cmdHeadTree := exec.Command(cmd, append(switches, options.Source)...)
	headTree, err = cmdHeadTree.Output()
	if err != nil {
		if canonical {
			Fail("Could not create canonical %s for %s (using '%s')",
				langName, options.Source, cmd)
		} else {
			Fail("Could not create %s parse tree for %s (using '%s')",
				langName, options.Source, cmd)
		}
	}

	cmdCurrentTree := exec.Command(cmd, append(switches, options.Destination)...)
	currentTree, err = cmdCurrentTree.Output()
	if err != nil {
		if canonical {
			Fail("Could not create canonical %s for %s (using '%s')",
				langName, options.Destination, cmd)
		} else {
			Fail("Could not create %s parse tree for %s (using '%s')",
				langName, options.Destination, cmd)
		}
	}
	return filename, headTree, currentTree
}

func RevisionToCurrentTree(
	filename string,
	cmd string,
	switches []string,
	options types.Options,
	langName string,
	canonical bool) ([]byte, []byte) {

	var headTree []byte
	var currentTree []byte
	var head []byte
	var err error

	// Get the AST for the current version of the file
	cmdCurrentTree := exec.Command(cmd, append(switches, filename)...)
	currentTree, err = cmdCurrentTree.Output()
	if err != nil {
		if canonical {
			Fail("Could not create canonical %s for %s (using '%s')",
				langName, filename, cmd)
		} else {
			Fail("Could not create %s parse tree for %s (using '%s')",
				langName, filename, cmd)
		}
	}

	// Retrieve the HEAD version of the file to a temporary filename
	cmdHead := exec.Command("git", "show", options.Source+filename)
	head, err = cmdHead.Output()
	if err != nil {
		Fail("Unable to retrieve file %s from branch/revision %s",
			filename, options.Source)
	}

	tmpfile, err := os.CreateTemp("", "*."+langName)
	if err != nil {
		Fail("Could not create a temporary %s file", langName)
	}
	tmpfile.Write(head)
	defer os.Remove(tmpfile.Name()) // clean up

	// Get the AST for the HEAD version of the file
	cmdHeadTree := exec.Command(cmd, append(switches, tmpfile.Name())...)
	headTree, err = cmdHeadTree.Output()
	if err != nil {
		if canonical {
			Fail("Could not create canonical %s for %s (using '%s')",
				langName, tmpfile.Name(), cmd)
		} else {
			Fail("Could not create %s parse tree for %s (using '%s')",
				langName, tmpfile.Name(), cmd)
		}
	}

	return headTree, currentTree
}

func VerifyHash(filename string, digest string) bool {
	file, err := os.Open(filename)
	if err != nil {
		Fail("%s", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		Fail("%s", err)
	}
	hashValue := hex.EncodeToString(hash.Sum(nil))
	return hashValue == digest
}
