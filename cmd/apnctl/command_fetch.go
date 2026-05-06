package main

import (
	"github.com/GlshchnkLx/go-aospapn/pkg/apntool"
	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

func runFetch(args []string) error {
	flags, fs := newCommonFlagSet("fetch")
	flags.url = aospURL
	flags.base64 = true
	flags.outputFormat = string(apnxml.FormatXML)
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
