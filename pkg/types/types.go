package types

// Structs to hold options and configurations
type (
	Options struct {
		status    bool
		semantic  bool
		glob      string
		verbose   bool
		parsetree bool
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

