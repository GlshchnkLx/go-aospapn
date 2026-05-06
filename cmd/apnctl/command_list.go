package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/GlshchnkLx/go-aospapn/pkg/apntool"
	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

func runList(args []string) error {
	common, filters, fs := newQueryFlagSet("list")
	var kind string
	fs.StringVar(&kind, "kind", "plmn", "list kind: plmn, type, carrier-id, carrier, apn")
	common.outputFormat = "text"
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

	values, err := listValues(tool, kind)
	if err != nil {
		return err
	}
	return writeList(common, values)
}

func listValues(tool apntool.Array, kind string) ([]string, error) {
	switch strings.ToLower(kind) {
	case "plmn", "plmns":
		return tool.PLMNs(), nil
	case "type", "types":
		types := tool.Types()
		values := make([]string, 0, len(types))
		for _, apnType := range types {
			values = append(values, apnType.String())
		}
		sort.Strings(values)
		return values, nil
	case "carrier-id", "carrier-ids", "carrierid", "carrierids":
		ids := tool.CarrierIDs()
		values := make([]string, 0, len(ids))
		for _, id := range ids {
			values = append(values, strconv.Itoa(id))
		}
		return values, nil
	case "carrier", "carriers":
		set := map[string]bool{}
		_ = tool.ForEachGroup(func(group apnxml.Object) error {
			if group.Carrier != "" {
				set[group.Carrier] = true
			}
			return nil
		})
		return setKeys(set), nil
	case "apn", "apns":
		set := map[string]bool{}
		_ = tool.ForEach(func(record apnxml.Object) error {
			if record.Base != nil && record.Base.Apn != nil && *record.Base.Apn != "" {
				set[*record.Base.Apn] = true
			}
			return nil
		})
		return setKeys(set), nil
	default:
		return nil, fmt.Errorf("unsupported list kind: %s", kind)
	}
}

func setKeys(set map[string]bool) []string {
	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}
