/* This small tool is a wrapper around the tree-sitter-cli framework. See
 * the README.md file in this directory for a detailed discussion of the
 * transformations performed and the reasons behind them.
 */

package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/atlantistechnology/sdt/pkg/utils"
)

func main() {
	var body []byte
	var err error
	comments := os.Getenv("TREESIT_COMMENTS") != ""

	if len(os.Args) != 2 {
		utils.Fail("`%s` requires exactly one filename argument", os.Args[0])
	}
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		utils.Info("Tree-sitter based parse tree: may set TREESITE_COMMENTS to retain comments")
		utils.Info("TREESIT_COMMENTS %t", comments)
		return
	}

	// Read in the file for later access to line/offset access
	filename := os.Args[1]
	body, err = os.ReadFile(filename)
	if err != nil {
		utils.Fail("Unable for read file %s", filename)
	}
	srcLines := strings.Split(string(body), "\n")

	// Run `tree-sitter`; return error status 1 if unavailable (for file type)
	cmd := exec.Command("tree-sitter", "parse", filename)
	out, err := cmd.Output()
	if err != nil {
		utils.Fail("No tree-sitter parser was found for %s", filename)
	}

	// Print the modified parse tree that SDT wishes to work with
	reComment := regexp.MustCompile(`\((line_)?comment`)
	rePosSpan := regexp.MustCompile(` \[\d+, \d+\] - \[\d+, \d+\]`)
	reNeedLiteral := regexp.MustCompile(
		`\((identifier |[a-z]+_literal |system_lib_string |operator )`,
	)
	var lineno, left, endline, right int

	fmt.Println("SrcLn | Node")
	parseLines := strings.Split(string(out), "\n")

	for _, line := range parseLines {
		// Should only be blank at final line, could also break
		if line == "" {
			continue
		}
		// Omit comments by default
		if reComment.MatchString(line) && !comments {
			continue
		}

		// Process the offset info then print what is needed
		posSpan := rePosSpan.FindString(line)
		if posSpan == "" {
			utils.Fail("Unable to locate position span for node: %s", line)
		}
		fmt.Sscanf(posSpan, " [%d, %d] - [%d, %d]", &lineno, &left, &endline, &right)
		sub := ""
		if reNeedLiteral.MatchString(line) {
			// TODO: multiline literals
			if lineno == endline {
				sub = " " + srcLines[lineno][left:right]
			}
		}
		line = strings.Replace(line, posSpan, sub, 1)
		fmt.Fprintf(os.Stdout, "%05d | %s\n", lineno+1, line)
	}
}
