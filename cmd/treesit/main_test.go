package main_test

import (
	"log"
	"os/exec"
	"strings"
	"testing"

	"github.com/atlantistechnology/sdt/pkg/utils"
)

func TestArithmetic(t *testing.T) {
	if 2+2 != 4 {
		t.Fatalf("Arithmetic doesn't work")
	}
}

func TestHelp(t *testing.T) {
	cmd := exec.Command("treesit", "--help")
	body, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	if !strings.Contains(string(body), "Tree-sitter based parse tree:") {
		t.Fatalf("Help message not displayed")
	}
}

func TestNoFile(t *testing.T) {
	cmd := exec.Command("treesit", "nosuch.file")
	body, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected an exit status 255 for missing file")
	}
	if !strings.Contains(string(body), "Unable for read file nosuch.file") {
		t.Fatalf("Missing file warning not displayed")
	}
}

// Let's assume tree-sitter will not get a Brainf*ck grammar
func TestNoGrammar(t *testing.T) {
	cmd := exec.Command("treesit", "hello.bf")
	body, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected an exit status 255 for missing grammar")
	}
	if !strings.Contains(string(body), "No tree-sitter parser") {
		t.Fatalf("Failed to identify missing grammar")
	}
}

// We generally expect tree-sitter and a C grammar to be present
// But do not fail this test if the grammar is not present
func TestGrammar(t *testing.T) {
	cmd := exec.Command("treesit", "hello0.c")
	body, err := cmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(body), "No tree-sitter parser") {
			t.Fatalf("Failed in some manner other than missing grammar: %s", err)
		}
		utils.Info("C grammar for tree-sitter missing, install for more complete test")
		return
	}
	if !strings.Contains(string(body), "SrcLn | Node") {
		t.Fatalf("Failed to produce a tree-sitter parse tree")
	}
}
