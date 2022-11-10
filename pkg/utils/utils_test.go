package utils_test

import (
	"reflect"
	"testing"

	"github.com/atlantistechnology/sdt/pkg/utils"
)

var sampleString = `Mary had a little lamb
Its fleece as white as snow
And everywhere that Mary went
The lamb was sure to go`

var sampleByteArray = []byte(sampleString)

var nRecords = 4
var startsWant = []uint32{0, 23, 51, 81}
var endsWant = []uint32{23, 51, 81, 105}

func TestOffsetsFromString(t *testing.T) {
	lineOffsets := utils.MakeOffsetsFromString(sampleString)
	if len(lineOffsets) != nRecords {
		t.Fatalf(`MakeOffsetsFromString() does not produce %d records`, nRecords)
	}
}

func TestOffsetsFromStringStartsEnds(t *testing.T) {
	lineOffsets := utils.MakeOffsetsFromString(sampleString)
	var starts []uint32
	var ends []uint32
	for _, record := range lineOffsets {
		starts = append(starts, record.Start)
		ends = append(ends, record.End)
	}
	if !reflect.DeepEqual(starts, startsWant) {
		t.Fatalf(
			`MakeOffsetsFromString() wrong starts: %v not %v`,
			starts, startsWant)
	}
	if !reflect.DeepEqual(ends, endsWant) {
		t.Fatalf(
			`MakeOffsetsFromString() wrong ends: %v not %v`,
			ends, endsWant)
	}
}

func TestOffsetsFromByteArray(t *testing.T) {
	lineOffsets := utils.MakeOffsetsFromByteArray(sampleByteArray)
	if len(lineOffsets) != nRecords {
		t.Fatalf(`MakeOffsetsFromString() does not produce %d records`, nRecords)
	}
}

func TestOffsetsFromByteArrayStartsEnds(t *testing.T) {
	lineOffsets := utils.MakeOffsetsFromByteArray(sampleByteArray)
	var starts []uint32
	var ends []uint32
	for _, record := range lineOffsets {
		starts = append(starts, record.Start)
		ends = append(ends, record.End)
	}
	if !reflect.DeepEqual(starts, startsWant) {
		t.Fatalf(
			`MakeOffsetsFromByteArray() wrong starts: %v not %v`,
			starts, startsWant)
	}
	if !reflect.DeepEqual(ends, endsWant) {
		t.Fatalf(
			`MakeOffsetsFromByteArray() wrong ends: %v not %v`,
			ends, endsWant)
	}
}

func TestLineAtPosition(t *testing.T) {
	offsetsTry := []uint32{3, 25, 26, 70, 90, 999}
	linesWant := []int{0, 1, 1, 2, 3, -1}
	lineOffsets := utils.MakeOffsetsFromString(sampleString)
	for i, offset := range offsetsTry {
		lineNo := utils.LineAtPosition(lineOffsets, offset)
		if lineNo != linesWant[i] {
			t.Fatalf(
				`LineAtPosition(lineOffsets, %v) produces %d rather than %d`,
				offset, lineNo, linesWant[i])
		}
	}
}
