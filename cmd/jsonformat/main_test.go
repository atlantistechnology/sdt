package main_test

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Deliberately slightly ugly JSON
var json string = `
{ "food": "pie", "toppings": [ "whipped cream", 
	"strawberries", "anchovies"]}
`
var pretty string = `{
  "food": "pie",
  "toppings": [
    "whipped cream",
    "strawberries",
    "anchovies"
  ]
}
`
var prettyLines []string = strings.Split(pretty, "\n")

func TestArithmetic(t *testing.T) {
	if 2+2 != 4 {
		t.Fatalf("Arithmetic doesn't work")
	}
}

func TestParser(t *testing.T) {
	file, err := os.CreateTemp("", "*.json")
	if err != nil {
		log.Fatal(err)
	}
	file.WriteString(json)
	defer os.Remove(file.Name())

	cmd := exec.Command("jsonformat", file.Name())
	body, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(body)
	bodyLines := strings.Split(bodyString, "\n")
	for i, line := range bodyLines {
		if line != prettyLines[i] {
			t.Fatalf("Unexpected canonical version:\nGot:  %s\nWant: %s",
				line, prettyLines[i])
		}
	}
}
