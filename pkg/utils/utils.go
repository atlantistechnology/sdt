package utils

import (
	"bytes"
	"strings"
)

type LineOffset struct {
	start uint32
	end   uint32
	line  string
}

type Options struct {
	status    bool
	semantic  bool
	glob      string
	verbose   bool
	parsetree bool
}

// We assume that lines are sensibly split on LF, not CR or CRLF
func MakeOffsetsFromString(text string) []LineOffset {
    lines := strings.Split(text, "\n")
	var records []LineOffset
	var start uint32 = 0
    var length uint32

	for i := 0; i < len(lines); i++ {
		length = uint32(len(lines[i]) + 1) // Add back striped LF
        record := LineOffset{start: start, end: start+length, line: lines[i]}
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
        record := LineOffset{start: start, end: start+length, line: string(lines[i])}
        records = append(records, record)
		start += length
	}
	return records
}

func LineAtPosition(lines []LineOffset, pos uint32) uint32 {
	var lineNumber uint32
	lineNumber = 0
	return lineNumber
}
