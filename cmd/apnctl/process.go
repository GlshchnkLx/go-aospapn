package main

import (
	"fmt"
	"strings"

	"github.com/GlshchnkLx/go-aospapn/pkg/apntool"
	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

func process(tool apntool.Array, flags *commonFlags) (apntool.Array, error) {
	if flags.flat {
		tool = tool.Flatten()
	}
	switch strings.ToLower(flags.groupBy) {
	case "":
	case "plmn":
		tool = tool.GroupByPLMN()
	case "identity":
		tool = tool.GroupByIdentity()
	default:
		return apntool.Array{}, fmt.Errorf("unsupported --group-by value: %s", flags.groupBy)
	}
	switch strings.ToLower(flags.dedupeBy) {
	case "":
	case "plmn":
		tool = tool.DedupeByPLMN()
	case "identity":
		tool = tool.DedupeByIdentity()
	default:
		return apntool.Array{}, fmt.Errorf("unsupported --dedupe-by value: %s", flags.dedupeBy)
	}
	if flags.normalize {
		normalized, err := tool.Normalize()
		if err != nil {
			return apntool.Array{}, err
		}
		tool = normalized
	}
	if flags.offset > 0 || flags.limit > 0 {
		tool = sliceRecords(tool, flags.offset, flags.limit, flags.flat)
	}
	return tool, nil
}

func sliceRecords(tool apntool.Array, offset int, limit int, keepFlat bool) apntool.Array {
	flat := tool.Flatten().Data()
	if offset < 0 {
		offset = 0
	}
	if offset >= len(flat) {
		return apntool.From(nil)
	}
	end := len(flat)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	result := apntool.From(apnxml.Array(flat[offset:end]))
	if keepFlat {
		return result
	}
	return result.GroupByIdentity()
}
