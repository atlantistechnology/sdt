package ruby

import (
	"bytes"
	"fmt"
	//"github.com/atlantistechnology/ast-diff/pkg/utils"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func simplifyParseTree(parseTree string) string {
	reComment := regexp.MustCompile(`(?m)^##.*$[\r\n]*`)
	mod1 := reComment.ReplaceAllString(parseTree, "")
	reLocation := regexp.MustCompile(`(?m)\(line:.*$[\r\n]*`)
	mod2 := reLocation.ReplaceAllString(mod1, "\n")
	reStrip := regexp.MustCompile(`(?m)^# `)
	mod3 := reStrip.ReplaceAllString(mod2, "")
	return mod3
}

func Diff(filename string, semantic bool, parsetree bool) string {
	var currentTree []byte
	var head []byte
	var headTree []byte
	var err error
	rubyCmd := "ruby" // TODO: Determine executable in more dynamic way

	cmdCurrentTree := exec.Command(rubyCmd, "--dump=parsetree", filename)
	currentTree, err = cmdCurrentTree.Output()
	if err != nil {
		log.Fatal(err)
	}

	cmdHead := exec.Command("git", "show", fmt.Sprintf("HEAD:%s", filename))
	head, err = cmdHead.Output()
	if err != nil {
		log.Fatal(err)
	}

	tmpfile, err := ioutil.TempFile("", "*.rb")
	if err != nil {
		log.Fatal(err)
	}
	tmpfile.Write(head)
	defer os.Remove(tmpfile.Name()) // clean up

	cmdHeadTree := exec.Command(rubyCmd, "--dump=parsetree", tmpfile.Name())
	headTree, err = cmdHeadTree.Output()
	if err != nil {
		log.Fatal(err)
	}

	// Make the trees into slightly simpler string representation
	headTreeString := simplifyParseTree(string(headTree))
	currentTreeString := simplifyParseTree(string(currentTree))

	// Perform the diff between the versions
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(headTreeString, currentTreeString, false)
	patch := dmp.PatchToText(dmp.PatchMake(diffs))

	// Only interested in line offsets of the change in parse tree
	reFromTo := regexp.MustCompile(`(?m)^[^@].*$[\r\n]*`)
	ranges := reFromTo.ReplaceAllString(patch, "")
	rangeLines := strings.Split(ranges, "\n")
	_ = rangeLines

	/*
		var oldStart int
		var oldCount int
		var newStart int
		var newCount int
		var n int
		for i := 0; i < len(rangeLines); i++ {
			n, err = fmt.Sscanf(rangeLines[i],
				"@@ -%d,%d +%d,%d @@",
				&oldStart, &oldCount, &newStart, &newCount)
			if n == 4 {
				fmt.Fprintf(os.Stderr,
					"Start %d Count %d, Start %d Count %d (matches %d)\n",
					oldStart, oldCount, newStart, newCount, n)
				if oldStart+oldCount < len(headTreeString) {
					fmt.Println(headTreeString[oldStart : oldStart+oldCount])
				}
				fmt.Println("--- Old ^^ --- New vvv ---")
				if newStart+newCount < len(currentTreeString) {
					fmt.Println(currentTreeString[newStart : newStart+newCount])
				}
			}
		}
	*/

	linesHeadTree := bytes.Split(headTree, []byte("\n"))
	linesCurrentTree := bytes.Split(currentTree, []byte("\n"))
	_ = linesHeadTree
	_ = linesCurrentTree

	if len(ranges) > 0 {
		return ranges
	} else {
		return "| No semantic differences detected"
	}
}
