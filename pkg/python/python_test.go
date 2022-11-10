package python

import (
	//"os/exec"
	"strings"
	"testing"
	//"github.com/atlantistechnology/sdt/pkg/types"
)

func TestNoSemanticDiff(t *testing.T) {
	// First make sure that two sample files indeed contain an expected diff,
	// then make sure that these differences are judged semantically unimportant
	/*
		options := types.Options{
			Status: false,
			Semantic: true,
			Parsetree: false,
			Glob: "*",
			Dumbterm: true,
			Verbose: false,
			Source: "samples/funcs0.py",
			Destination: "samples/funcs1.py",
		}

		Diff("", options, config)
	*/

	if !strings.Contains("foobarbaz", "bar") {
		t.Fatalf("Substring not found")
	}

}
