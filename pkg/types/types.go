package types

// Structs to hold options and configurations
type (
	Options struct {
		Status    bool
		Semantic  bool
		Glob      string
		Verbose   bool
		Parsetree bool
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
