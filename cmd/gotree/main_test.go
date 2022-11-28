package main_test

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

var goProg string = `package main
import "fmt"

func main() {
	fmt.Println("hello world")
}
`
var parsed string = `SrcLn | Node
00000 | *ast.File
00000 |   Doc: nil
00001 |   Name: *ast.Ident
00001 |     Name: "main"
00001 |     Obj: nil
00001 |   Decls: []ast.Decl (len = 2)
00001 |     0: *ast.GenDecl
00001 |       Doc: nil
00002 |       Tok: import
00002 |       Specs: []ast.Spec (len = 1)
00002 |         0: *ast.ImportSpec
00002 |           Doc: nil
00002 |           Name: nil
00002 |           Path: *ast.BasicLit
00002 |             Kind: STRING
00002 |             Value: "\"fmt\""
00002 |           Comment: nil
00002 |     1: *ast.FuncDecl
00002 |       Doc: nil
00002 |       Recv: nil
00002 |       Name: *ast.Ident
00004 |         Name: "main"
00004 |         Obj: *ast.Object
00004 |           Kind: func
00004 |           Name: "main"
00004 |           Decl: *(obj @ 29)
00004 |           Data: nil
00004 |           Type: nil
00004 |       Type: *ast.FuncType
00004 |         TypeParams: nil
00004 |         Params: *ast.FieldList
00004 |           List: nil
00004 |         Results: nil
00004 |       Body: *ast.BlockStmt
00004 |         List: []ast.Stmt (len = 1)
00004 |           0: *ast.ExprStmt
00004 |             X: *ast.CallExpr
00004 |               Fun: *ast.SelectorExpr
00004 |                 X: *ast.Ident
00005 |                   Name: "fmt"
00005 |                   Obj: nil
00005 |                 Sel: *ast.Ident
00005 |                   Name: "Println"
00005 |                   Obj: nil
00005 |               Args: []ast.Expr (len = 1)
00005 |                 0: *ast.BasicLit
00005 |                   Kind: STRING
00005 |                   Value: "\"hello world\""
`
var parsedLines []string = strings.Split(parsed, "\n")

func TestArithmetic(t *testing.T) {
	if 2+2 != 4 {
		t.Fatalf("Arithmetic doesn't work")
	}
}

func TestParser(t *testing.T) {
	file, err := os.CreateTemp("", "*.go")
	if err != nil {
		log.Fatal(err)
	}
	file.WriteString(goProg)
	defer os.Remove(file.Name())

	cmd := exec.Command("gotree", file.Name())
	body, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(body)
	bodyLines := strings.Split(bodyString, "\n")
	for i, line := range(bodyLines) {
		line = strings.TrimRight(line, " ")
		if line != parsedLines[i] {
			t.Fatalf("Unexpected parse tree:\nGot:  %s\nWant: %s", line, parsedLines[i])
		}
	}
}
