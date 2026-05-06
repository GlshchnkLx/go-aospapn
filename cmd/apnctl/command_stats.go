package main

import "github.com/GlshchnkLx/go-aospapn/pkg/apntool"

func runStats(args []string) error {
	flags, fs := newCommonFlagSet("stats")
	flags.outputFormat = "summary"
	if err := fs.Parse(args); err != nil {
		return err
	}

	data, err := loadAPNs(flags)
	if err != nil {
		return err
	}
	tool, err := process(apntool.From(data), flags)
	if err != nil {
		return err
	}
	return writeStats(flags, tool.Stats())
}
