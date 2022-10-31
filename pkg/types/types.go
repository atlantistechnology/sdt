package types

// Structs to hold options and configurations
type (
	Options struct {
		Status      bool
		Semantic    bool
		Parsetree   bool
		Glob        string
		Dumbterm    bool
		Verbose     bool
		Source      string
		Destination string
	}

	Config struct {
		Description string             `toml:"description"`
		Commands    map[string]Command `toml:"commands"`
		Glob        string             `toml:"glob"`
	}

	Command struct {
		Executable string   `toml:"executable"`
		Switches   []string `toml:"switches"`
	}
)

// In places, github.com/fatih/color is used, but raw ANSI is easier
// for writing custom reports based on sergi/go-diff/diffmatchpatch
var CYAN string = "\x1b[36m"
var YELLOW string = "\x1b[33m"
var GREEN string = "\x1b[32m"
var RED string = "\x1b[31m"
var CLEAR string = "\x1b[0m"

// Keep "enum" of types of parse trees we can handle here (<=256 for now)
type ParseType uint8

const (
	Ruby ParseType = iota
	Python
	JavaScript
	Golang
	SomeOtherLanguage
)
