package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/yaml-workflow/yaml"
)

func main() {
	// Configuring hclog like this allows Lyra to handle log levels automatically
	hclog.DefaultOptions = &hclog.LoggerOptions{
		Name:            "Yaml",
		Level:           hclog.LevelFromString(os.Getenv("LYRA_LOG_LEVEL")),
		JSONFormat:      true,
		IncludeLocation: false,
		Output:          os.Stderr,
	}
	if hclog.DefaultOptions.Level <= hclog.Debug {
		// Tell issue reporting to amend all errors with a stack trace.
		issue.IncludeStacktrace(true)
	}
	yaml.Start(`Yaml`)
}
