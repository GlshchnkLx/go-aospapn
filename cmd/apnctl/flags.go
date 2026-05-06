package main

import (
	"flag"

	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

func newCommonFlagSet(name string) (*commonFlags, *flag.FlagSet) {
	flags := &commonFlags{outputFormat: string(apnxml.FormatJSON)}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.StringVar(&flags.in, "in", "", "input file")
	fs.BoolVar(&flags.stdin, "stdin", false, "read input from stdin")
	fs.StringVar(&flags.url, "url", "", "input URL")
	fs.BoolVar(&flags.base64, "base64", false, "decode base64 input body")
	fs.StringVar(&flags.inputFormat, "input-format", "", "input format: xml or json")
	fs.StringVar(&flags.out, "out", "", "output file")
	fs.StringVar(&flags.outputFormat, "output-format", flags.outputFormat, "output format: xml, json, table, csv, text, summary")
	fs.BoolVar(&flags.flat, "flat", false, "flatten grouped records")
	fs.StringVar(&flags.groupBy, "group-by", "", "group flat records by plmn or identity")
	fs.BoolVar(&flags.normalize, "normalize", false, "normalize records before output")
	fs.StringVar(&flags.dedupeBy, "dedupe-by", "", "dedupe/group by plmn or identity")
	fs.IntVar(&flags.offset, "offset", 0, "skip N materialized records before output")
	fs.IntVar(&flags.limit, "limit", 0, "limit materialized records after filtering")
	return flags, fs
}

func newQueryFlagSet(name string) (*commonFlags, *filterFlags, *flag.FlagSet) {
	common, fs := newCommonFlagSet(name)
	filters := &filterFlags{mcc: -1, mnc: -1, carrierID: -1}
	fs.Var(&filters.plmn, "plmn", "PLMN as MCCMNC; repeatable")
	fs.IntVar(&filters.mcc, "mcc", -1, "MCC")
	fs.IntVar(&filters.mnc, "mnc", -1, "MNC")
	fs.IntVar(&filters.carrierID, "carrier-id", -1, "carrier ID")
	fs.StringVar(&filters.carrier, "carrier", "", "carrier name substring")
	fs.StringVar(&filters.apn, "apn", "", "exact APN")
	fs.StringVar(&filters.apnContains, "apn-contains", "", "APN substring")
	fs.StringVar(&filters.apnType, "type", "", "APN type mask")
	fs.StringVar(&filters.protocol, "protocol", "", "bearer protocol mask")
	fs.StringVar(&filters.network, "network", "", "network bitmask")
	fs.BoolVar(&filters.validOnly, "valid-only", false, "include valid records only")
	fs.BoolVar(&filters.invalidOnly, "invalid-only", false, "include invalid records only")
	fs.Var(&filters.has, "has", "require section: root, valid-root, base, auth, bearer, proxy, mms, mvno")
	fs.Var(&filters.without, "without", "exclude records with section: root, valid-root, base, auth, bearer, proxy, mms, mvno")
	fs.BoolVar(&filters.invert, "not", false, "invert the final predicate")
	return common, filters, fs
}
