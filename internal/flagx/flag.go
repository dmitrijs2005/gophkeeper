package flagx

import (
	"flag"
	"os"
	"strings"
)

// FilterArgs returns a slice of command-line arguments that only contains
// the allowed flags (and their values) specified in allowedFlags.
//
// Supported formats:
//  1. Flag and value as separate arguments:  -c conf.json
//  2. Flag and value combined with '=':      --config=conf.json
//
// Parameters:
//
//	args         — the command-line arguments (usually os.Args[1:])
//	allowedFlags — list of allowed flag names (e.g. []string{"-c", "--config"})
//
// Returns:
//
//	A slice containing the allowed flags and their values (if provided separately).
func FilterArgs(args []string, allowedFlags []string) []string {
	// Convert the list of allowed flags into a map for O(1) lookup
	allowed := make(map[string]struct{}, len(allowedFlags))
	for _, f := range allowedFlags {
		allowed[f] = struct{}{}
	}

	// Initialize the result slice as empty (not nil) so it’s always safe to use
	filtered := make([]string, 0, len(args))

	// Iterate over the arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Case 1: flag in the form "--flag=value" or "-f=value"
		if strings.HasPrefix(arg, "-") && strings.Contains(arg, "=") {
			// Extract the flag name (before the '=')
			name := strings.SplitN(arg, "=", 2)[0]
			// If this flag is allowed, keep the whole "flag=value" argument
			if _, ok := allowed[name]; ok {
				filtered = append(filtered, arg)
			}
			continue
		}

		// Case 2: flag as a separate argument (value might follow)
		if _, ok := allowed[arg]; ok {
			filtered = append(filtered, arg)
			// If the next argument exists and does not look like another flag,
			// treat it as this flag's value and include it
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				filtered = append(filtered, args[i+1])
				i++ // skip the value in the next loop iteration
			}
		}
	}

	return filtered
}

// jsonConfigFlags inspects command-line arguments and extracts the config file
// path provided via the -c or -config flags.
//
// Only these flags are parsed; other arguments are ignored. This allows the
// application to safely parse its own flags without interfering with flags
// defined by other packages.
//
// If neither -c nor -config is present, an empty string is returned.
func JsonConfigFlags() string {
	var config string

	args := FilterArgs(os.Args[1:], []string{"-c", "-config"})

	fs := flag.NewFlagSet("json", flag.ContinueOnError)
	fs.StringVar(&config, "config", "", "Path to config file")
	fs.StringVar(&config, "c", "", "Path to config file (short)")
	_ = fs.Parse(args)

	return config
}
