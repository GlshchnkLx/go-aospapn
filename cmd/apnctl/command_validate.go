package main

import (
	"fmt"

	"github.com/GlshchnkLx/go-aospapn/pkg/apntool"
)

func runValidate(args []string) error {
	common, filters, fs := newQueryFlagSet("validate")
	var strict bool
	common.outputFormat = "summary"
	fs.BoolVar(&strict, "strict", false, "return an error when invalid records exist")
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
	stats := tool.Stats()
	if err := writeStats(common, stats); err != nil {
		return err
	}
	if strict && stats.Invalid > 0 {
		return fmt.Errorf("invalid APN records: %d", stats.Invalid)
	}
	return nil
}
