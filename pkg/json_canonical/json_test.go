package json_test

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
	"testing"

	"github.com/atlantistechnology/sdt/pkg/types"
)

type File struct {
	name   string
	digest string
}

func testHash(filename string, digest string) bool {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}
	hashValue := hex.EncodeToString(hash.Sum(nil))
	return hashValue == digest
}

var options = types.Options{
	Status:      false,
	Semantic:    true,
	Parsetree:   false,
	Glob:        "*",
	Dumbterm:    true,
	Verbose:     false,
	Source:      "",
	Destination: "",
}

var config = types.Config{
	Description: "Configuration for SQL tests",
	Glob:        "*",
	Commands:    types.Commands,
}

var file0 = File{name: "file0.sql", digest: "8c7d0bd6d23270b957a41e33da2abeb6"}
var file1 = File{name: "file1.sql", digest: "8a989e81af658fd03c026d741d9eb9fc"}
var file2 = File{name: "file2.sql", digest: "004ad785ad53561f920020c5a7cc2bc9"}

func TestCorrectFiles(t *testing.T) {
	// First make sure that two sample files indeed contain an expected bodies,
	// then make sure that these differences are judged semantically unimportant
	files := []File{file0, file1, file2}
	for _, file := range files {
		if !testHash(file.name, file.digest) {
			t.Fatalf("Test file %s has been changed from expected body", file.name)
		}
	}
}
