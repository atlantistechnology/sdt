/* The purpose of this small program is to generate an AST for a Golang
 * package.  It provides no options and always requires exactly one filename
 * as an argument.
 */
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/atlantistechnology/sdt/pkg/utils"
)

func main() {
	var err error
	var code []byte
	raw := os.Getenv("GOTREE_RAW") != ""

	if len(os.Args) != 2 {
		utils.Fail("`%s` requires exactly one filename argument", os.Args[0])
	}
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		utils.Info("Golang parse tree: may set GOTREE_RAW for unmassaged AST")
		return
	}

	filename := os.Args[1]
	code, err = os.ReadFile(filename)
	if err != nil {
		utils.Fail("Unable for read file %s", filename)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", code, parser.AllErrors)
	if err != nil {
		utils.Fail("Unable for parse Golang file %s (%s)", filename, err)
	}

	if raw {
		err = ast.Print(fset, f)
		if err != nil {
			utils.Fail("Unable for to print parsed Golang file %s (%s)", filename, err)
		}
		return
	}

	var buf bytes.Buffer
	err = ast.Fprint(&buf, fset, f, nil)
	if err != nil {
		utils.Fail("Unable for to print parsed Golang file %s (%s)", filename, err)
	}

	lineMark := regexp.MustCompile(
		`^( *)((` +
			`Struct|Defer|Map|Interface|Switch|Case|For|Return|Package|Func|If|` +
			`Assign|Arrow|Go|Begin|Select|Opening|Closing|Star|Colon|Ellipsis|` +
			`.*Pos|[LR]paren|[LR]brace|[LR]brack` +
			`): )(.*)`)
	fmt.Println("SrcLn | Node")
	lines := strings.Split(buf.String(), "\n")
	lineno := 0
	for _, line := range lines {
		line = line[utils.Min(8, len(line)):]
		line = strings.Replace(line, ".  ", "  ", -1)

		// The final parts of the tree are not line-by-line of interest
		if strings.HasPrefix(line, "Scope: ") ||
			strings.HasPrefix(line, "Imports: ") ||
			strings.HasPrefix(line, "Unresolved: ") {
			break
		}

		// Several node types announce position
		if lineMark.MatchString(line) {
			parts := strings.Split(line, ":")
			if len(parts) == 3 {
				lineno, _ = strconv.Atoi(strings.Replace(parts[1], " ", "0", -1))
			}
			justNode := lineMark.ReplaceAllString(line, "$1$2")
			fmt.Fprintf(os.Stdout, "%05d | %s?\n", lineno, justNode)
		} else {
			fmt.Fprintf(os.Stdout, "%05d | %s\n", lineno, line)
		}
	}
}
