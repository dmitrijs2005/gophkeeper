// Package buildinfo provides a utility to print embedded build metadata such as version,
// build date, and commit hash. These values are expected to be injected at build time
// using -ldflags.
package buildinfo

import (
	"fmt"
	"io"
)

// buildVersion, buildDate, and buildCommit are populated at build time via -ldflags.
// Example:
//
//	go build -ldflags "-X 'package_path/buildinfo.buildVersion=1.0.0' -X '...'" ...
var buildVersion, buildDate, buildCommit string

// printBuildParam writes a single build parameter to the provided writer.
// If the value is empty, it substitutes it with "N/A".
func printBuildParam(w io.Writer, name string, value string) {
	if value == "" {
		value = "N/A"
	}
	fmt.Fprintf(w, "%s: %s\n", name, value)
}

// PrintBuildData prints the build version, build date, and build commit
// to the provided writer. This is useful for logging or displaying version info.
func PrintBuildData(w io.Writer) {
	printBuildParam(w, "Build version", buildVersion)
	printBuildParam(w, "Build date", buildDate)
	printBuildParam(w, "Build commit", buildCommit)
}
