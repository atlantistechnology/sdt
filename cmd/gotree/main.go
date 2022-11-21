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
	//"strconv"
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

	fmt.Println("SrcLn | Node")
	lines := strings.Split(buf.String(), "\n")
	lineno := 0
	//var n int
	for _, line := range lines {
		line = line[utils.Min(8, len(line)):]
		line = strings.Replace(line, ".  ", "  ", -1)
		if strings.Contains(line, "Pos: ") {
			parts := strings.Split(line, ":")
			if len(parts) == 3 {
				fmt.Println("XXX", parts)
			}
			//if n, err = strconv.Atoi(strings.Split(line, ":")[1]); err != nil {
			//	lineno = n
			//}
		}
		fmt.Fprintf(os.Stdout, "%05d  | %s\n", lineno, line)
	}


}

