package utils

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/atlantistechnology/sdt/pkg/types"
	"golang.org/x/exp/constraints"
)

type lineOffset struct {
	start uint32
	end   uint32
	line  string
}

// We assume that lines are sensibly split on LF, not CR or CRLF
func MakeOffsetsFromString(text string) []lineOffset {
	lines := strings.Split(text, "\n")
	var records []lineOffset
	var start uint32 = 0
	var length uint32

	for i := 0; i < len(lines); i++ {
		length = uint32(len(lines[i]) + 1) // Add back striped LF
		record := lineOffset{start: start, end: start + length, line: lines[i]}
		records = append(records, record)
		start += length
	}
	return records
}

// We assume that lines are sensibly split on LF, not CR or CRLF
// Also assume that bytes encode text as UTF-8 not any odd encoding
func MakeOffsetsFromByteArray(text []byte) []lineOffset {
	lines := bytes.Split(text, []byte("\n"))
	var records []lineOffset
	var start uint32 = 0
	var length uint32

	for i := 0; i < len(lines); i++ {
		length = uint32(len(lines[i]) + 1) // Add back stripped LF
		record := lineOffset{start: start, end: start + length, line: string(lines[i])}
		records = append(records, record)
		start += length
	}
	return records
}

// Return the line number identified.  Use -1 as sentinel for "not found"
// TODO: if we care about speed, we can do a bisection search of the
// well-ordered {start, end, line} structures
func LineAtPosition(records []lineOffset, pos uint32) int {
	for lineNo, record := range records {
		if pos >= record.start && pos < record.end {
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

func BufferToDiff(buff bytes.Buffer, colorLeft bool) string {
	ret := buff.String()
	rePrepend := regexp.MustCompile(`(?m)^`)
	if colorLeft {
		colorPipe := types.YELLOW + "| " + types.CLEAR
		ret = rePrepend.ReplaceAllString(ret, colorPipe)
	} else {
		ret = rePrepend.ReplaceAllString(ret, "| ")
	}
	return ret
}
