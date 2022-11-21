package types

// Structs to hold options and configurations
type (
	Options struct {
		Status      bool
		Semantic    bool
		Parsetree   bool
		Glob        string
		Minimal     bool
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
		Options    string   `toml:"options"`
	}

	Highlights struct {
		Add     string
		Del     string
		Header  string
		Info    string
		Clear   string
		Neutral string
	}
)

type LineType int8

const (
	Status      LineType = iota
	CompactDiff          // Probably deprecated for "fake" RawNames
	RawNames
)

// In places, github.com/fatih/color is used, but raw ANSI is easier
// for writing custom reports based on sergi/go-diff/diffmatchpatch
var Colors Highlights = Highlights{
	Add:     "\x1b[32m", // green
	Del:     "\x1b[31m", // red
	Header:  "\x1b[33m", // yellow
	Info:    "\x1b[36m", // cyan
	Clear:   "\x1b[0m",
	Neutral: "",
}

var Dumbterm Highlights = Highlights{
	Add:     "{{+",
	Del:     "{{-",
	Header:  "",
	Info:    "",
	Clear:   "}}",
	Neutral: "{{_",
}

var PlainASCII Highlights = Highlights{
	Add:     "",
	Del:     "",
	Header:  "",
	Info:    "",
	Clear:   "",
	Neutral: "",
}

// Create "enum" of filetypes we can handle (<=256 langs for now)
type ParseType uint8

const (
	Ruby ParseType = iota
	Python
	JavaScript
	JSON
	Golang
	SomeOtherLanguage
)

var JsSwitches string = `
	const acorn = require("acorn"); 
	const fs = require("fs"); 
	const source = fs.readFileSync(process.argv[1], "utf8");
	const parse = acorn.parse(source, ${OPTIONS});
	console.log(JSON.stringify(parse, null, "  "));
`

var Commands = map[string]Command{
	// Configure default tools that might be overrridden by the TOML config
	"python": {
		Executable: "python",
		Switches:   []string{"-m", "ast", "-a"},
		Options:    "",
	},
	"ruby": {
		Executable: "ruby",
		Switches:   []string{"--dump=parsetree"},
		Options:    "",
	},
	"sql": {
		Executable: "sqlformat",
		Switches: []string{
			"--reindent_aligned",
			"--identifiers=lower",
			"--strip-comments",
			"--keywords=upper",
		},
		Options: "",
	},
	"javascript": {
		Executable: "node",
		Switches:   []string{"-e", JsSwitches},
		Options:    `{sourceType: "module", ecmaVersion: "latest"}`,
	},
	// A tiny and simple tool (within this project is  used by default.
	// For an example of using external tool `jq`, see `samples/.sdt.toml`
	"json": {
		Executable: "jsonformat",
		Switches:   []string{},
		Options:    "",
	},
}
