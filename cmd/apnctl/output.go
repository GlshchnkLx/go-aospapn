package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/GlshchnkLx/go-aospapn/pkg/apntool"
	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

func writeAPNs(flags *commonFlags, tool apntool.Array) error {
	switch strings.ToLower(flags.outputFormat) {
	case "json":
		return writeData(flags.out, func(writer io.Writer) error {
			return apnxml.ExportToWriter(tool.Data(), writer, apnxml.FormatJSON)
		})
	case "xml":
		return writeData(flags.out, func(writer io.Writer) error {
			return apnxml.ExportToWriter(tool.Data(), writer, apnxml.FormatXML)
		})
	case "table", "text":
		return writeData(flags.out, func(writer io.Writer) error {
			return writeTable(writer, tool)
		})
	case "csv":
		return writeData(flags.out, func(writer io.Writer) error {
			return writeCSV(writer, tool)
		})
	case "summary":
		return writeStats(flags, tool.Stats())
	default:
		return fmt.Errorf("unsupported output format: %s", flags.outputFormat)
	}
}

func writeStats(flags *commonFlags, stats apntool.Stats) error {
	if strings.EqualFold(flags.outputFormat, "json") {
		return writeJSON(flags.out, stats)
	}
	return writeData(flags.out, func(writer io.Writer) error {
		fmt.Fprintf(writer, "groups: %d\nrecords: %d\ninvalid: %d\n", stats.Groups, stats.Records, stats.Invalid)
		writeStringIntMap(writer, "by_plmn", stats.ByPLMN)
		typeStats := map[string]int{}
		for apnType, count := range stats.ByType {
			typeStats[apnType.String()] = count
		}
		writeStringIntMap(writer, "by_type", typeStats)
		return nil
	})
}

func writeList(flags *commonFlags, values []string) error {
	sort.Strings(values)
	switch strings.ToLower(flags.outputFormat) {
	case "json":
		return writeJSON(flags.out, values)
	case "csv":
		return writeData(flags.out, func(writer io.Writer) error {
			csvWriter := csv.NewWriter(writer)
			for _, value := range values {
				if err := csvWriter.Write([]string{value}); err != nil {
					return err
				}
			}
			csvWriter.Flush()
			return csvWriter.Error()
		})
	case "text", "table", "summary":
		return writeData(flags.out, func(writer io.Writer) error {
			for _, value := range values {
				fmt.Fprintln(writer, value)
			}
			return nil
		})
	default:
		return fmt.Errorf("unsupported output format for list: %s", flags.outputFormat)
	}
}

func writeStringIntMap(writer io.Writer, title string, values map[string]int) {
	fmt.Fprintf(writer, "%s:\n", title)
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(writer, "  %s: %d\n", key, values[key])
	}
}

func writeTable(writer io.Writer, tool apntool.Array) error {
	fmt.Fprintln(writer, "PLMN\tCarrier\tCarrierID\tType\tAPN\tProtocol\tRoamingProtocol\tNetwork\tProfileID\tEnabled\tVisible\tEditable")
	return tool.ForEach(func(record apnxml.Object) error {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			record.GetPLMN(),
			record.Carrier,
			intPtrString(record.CarrierID),
			baseTypeString(record.Base),
			apnString(record.Base),
			protocolString(record.Bearer),
			roamingProtocolString(record.Bearer),
			networkString(record.Other),
			profileIDString(record.Base),
			boolPtrString(otherBool(record.Other, "enabled")),
			boolPtrString(otherBool(record.Other, "visible")),
			boolPtrString(otherBool(record.Other, "editable")),
		)
		return nil
	})
}

func writeCSV(writer io.Writer, tool apntool.Array) error {
	csvWriter := csv.NewWriter(writer)
	header := []string{"plmn", "carrier", "carrier_id", "type", "apn", "protocol", "roaming_protocol", "network", "profile_id", "enabled", "visible", "editable"}
	if err := csvWriter.Write(header); err != nil {
		return err
	}
	err := tool.ForEach(func(record apnxml.Object) error {
		return csvWriter.Write([]string{
			record.GetPLMN(),
			record.Carrier,
			intPtrString(record.CarrierID),
			baseTypeString(record.Base),
			apnString(record.Base),
			protocolString(record.Bearer),
			roamingProtocolString(record.Bearer),
			networkString(record.Other),
			profileIDString(record.Base),
			boolPtrString(otherBool(record.Other, "enabled")),
			boolPtrString(otherBool(record.Other, "visible")),
			boolPtrString(otherBool(record.Other, "editable")),
		})
	})
	if err != nil {
		return err
	}
	csvWriter.Flush()
	return csvWriter.Error()
}

func writeInspect(writer io.Writer, tool apntool.Array) error {
	return tool.ForEachGroup(func(group apnxml.Object) error {
		fmt.Fprintf(writer, "PLMN: %s\ncarrier: %s\ncarrier_id: %s\nrecords: %d\n", group.GetPLMN(), group.Carrier, intPtrString(group.CarrierID), group.CountRecords())
		for _, record := range group.Records() {
			materialized := apntool.MaterializeRecord(&group, record)
			fmt.Fprintf(writer, "  type=%s apn=%s profile_id=%s protocol=%s roaming_protocol=%s network=%s enabled=%s visible=%s editable=%s\n",
				baseTypeString(materialized.Base),
				apnString(materialized.Base),
				profileIDString(materialized.Base),
				protocolString(materialized.Bearer),
				roamingProtocolString(materialized.Bearer),
				networkString(materialized.Other),
				boolPtrString(otherBool(materialized.Other, "enabled")),
				boolPtrString(otherBool(materialized.Other, "visible")),
				boolPtrString(otherBool(materialized.Other, "editable")),
			)
		}
		return nil
	})
}

func writeJSON(path string, value any) error {
	return writeData(path, func(writer io.Writer) error {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "\t")
		return encoder.Encode(value)
	})
}

func writeData(path string, write func(io.Writer) error) error {
	writer, closeOutput, err := outputWriter(path)
	if err != nil {
		return err
	}
	defer closeOutput()
	return write(writer)
}

func outputWriter(path string) (io.Writer, func(), error) {
	if path == "" {
		return os.Stdout, func() {}, nil
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	return file, func() { _ = file.Close() }, nil
}

func intPtrString(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func boolPtrString(value *bool) string {
	if value == nil {
		return ""
	}
	return strconv.FormatBool(*value)
}

func baseTypeString(base *apnxml.ObjectBase) string {
	if base == nil || base.Type == nil {
		return ""
	}
	return base.Type.String()
}

func apnString(base *apnxml.ObjectBase) string {
	if base == nil || base.Apn == nil {
		return ""
	}
	return *base.Apn
}

func profileIDString(base *apnxml.ObjectBase) string {
	if base == nil || base.ProfileID == nil {
		return ""
	}
	return strconv.Itoa(*base.ProfileID)
}

func protocolString(bearer *apnxml.ObjectBearer) string {
	if bearer == nil || bearer.Type == nil {
		return ""
	}
	return bearer.Type.String()
}

func roamingProtocolString(bearer *apnxml.ObjectBearer) string {
	if bearer == nil || bearer.TypeRoaming == nil {
		return ""
	}
	return bearer.TypeRoaming.String()
}

func networkString(other *apnxml.ObjectOther) string {
	if other == nil || other.NetworkTypeBitmask == nil {
		return ""
	}
	return other.NetworkTypeBitmask.String()
}

func otherBool(other *apnxml.ObjectOther, name string) *bool {
	if other == nil {
		return nil
	}
	switch name {
	case "enabled":
		return other.CarrierEnabled
	case "visible":
		return other.UserVisible
	case "editable":
		return other.UserEditable
	default:
		return nil
	}
}
