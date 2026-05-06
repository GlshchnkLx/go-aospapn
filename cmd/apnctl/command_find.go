package main

import "github.com/GlshchnkLx/go-aospapn/pkg/apntool"

func runFind(args []string) error {
	common, filters, fs := newQueryFlagSet("find")
	if err := fs.Parse(args); err != nil {
		return err
	}

	data, err := loadAPNs(common)
	if err != nil {
		return err
	}
	predicate, err := buildPredicate(filters)
	if err != nil {
		return err
	}
	tool, err := process(apntool.From(data).Filter(predicate), common)
	if err != nil {
		return err
	}
	return writeAPNs(common, tool)
}
