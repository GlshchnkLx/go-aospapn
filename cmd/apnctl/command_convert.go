package main

import "github.com/GlshchnkLx/go-aospapn/pkg/apntool"

func runConvert(args []string) error {
	flags, fs := newCommonFlagSet("convert")
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
	return writeAPNs(flags, tool)
}
