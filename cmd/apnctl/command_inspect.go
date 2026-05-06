package main

import "github.com/GlshchnkLx/go-aospapn/pkg/apntool"

func runInspect(args []string) error {
	common, filters, fs := newQueryFlagSet("inspect")
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

	writer, closeOutput, err := outputWriter(common.out)
	if err != nil {
		return err
	}
	defer closeOutput()
	return writeInspect(writer, tool)
}
