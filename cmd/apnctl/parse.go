package main

import (
	"strconv"
	"strings"
)

func setOptionalBool(target **bool, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	*target = &parsed
	return nil
}
