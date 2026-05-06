package main

import (
	"fmt"
	"os"

	"github.com/GlshchnkLx/go-aospapn/pkg/apntool"
	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

func runPatch(args []string) error {
	common, filters, fs := newQueryFlagSet("patch")
	var setList stringList
	var modeValue string
	var patchFile string
	var patchFormat string
	var strict bool
	fs.Var(&setList, "set", "set APN field as section.field=value")
	fs.StringVar(&modeValue, "mode", "patch", "update mode: merge, patch, apply")
	fs.StringVar(&patchFile, "patch-file", "", "XML or JSON APN file to merge, patch, or apply")
	fs.StringVar(&patchFormat, "patch-format", "", "patch file format: xml or json")
	fs.BoolVar(&strict, "strict", false, "return an error when --set matches no records")
	if err := fs.Parse(args); err != nil {
		return err
	}

	mode, err := apnxml.ParseObjectUpdateMode(modeValue)
	if err != nil {
		return err
	}
	data, err := loadAPNs(common)
	if err != nil {
		return err
	}

	tool := apntool.From(data)
	if patchFile != "" {
		patchData, err := loadFile(patchFile, patchFormat)
		if err != nil {
			return err
		}
		switch mode {
		case apnxml.ObjectUpdateMerge:
			tool = tool.Merge(patchData)
		case apnxml.ObjectUpdatePatch:
			tool = tool.Patch(patchData)
		case apnxml.ObjectUpdateApply:
			tool = tool.ApplyUpdate(patchData)
		}
	}

	if len(setList) > 0 {
		var patch apnxml.Object
		for _, expr := range setList {
			if err := apntool.SetObjectFieldExpr(&patch, expr); err != nil {
				return err
			}
		}
		predicate, err := buildPredicate(filters)
		if err != nil {
			return err
		}
		result, err := tool.UpdateByFilter(predicate, &patch, mode)
		if err != nil {
			return err
		}
		if strict && result.Matched == 0 {
			return fmt.Errorf("patch matched no records")
		}
		tool = result.Data
		fmt.Fprintf(os.Stderr, "matched=%d changed=%d\n", result.Matched, result.Changed)
	}

	tool, err = process(tool, common)
	if err != nil {
		return err
	}
	return writeAPNs(common, tool)
}
