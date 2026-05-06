package main

import "strings"

type stringList []string

func (list *stringList) String() string {
	return strings.Join(*list, ",")
}

func (list *stringList) Set(value string) error {
	*list = append(*list, value)
	return nil
}

type commonFlags struct {
	in           string
	out          string
	url          string
	stdin        bool
	base64       bool
	inputFormat  string
	outputFormat string
	flat         bool
	groupBy      string
	normalize    bool
	dedupeBy     string
	offset       int
	limit        int
}

type filterFlags struct {
	plmn        stringList
	mcc         int
	mnc         int
	carrierID   int
	carrier     string
	apn         string
	apnContains string
	apnType     string
	protocol    string
	network     string
	validOnly   bool
	invalidOnly bool
	has         stringList
	without     stringList
	invert      bool
}
