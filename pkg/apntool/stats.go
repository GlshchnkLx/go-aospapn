package apntool

import (
	"sort"

	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

type Stats struct {
	Groups  int
	Records int
	Invalid int
	ByType  map[apnxml.ObjectBaseType]int
	ByPLMN  map[string]int
}

func (array Array) Stats() Stats {
	stats := Stats{
		Groups:  len(array.data),
		Records: array.data.CountRecords(),
		ByType:  map[apnxml.ObjectBaseType]int{},
		ByPLMN:  map[string]int{},
	}

	for _, record := range flatten(array.data) {
		if record.ObjectRoot == nil || !record.ObjectRoot.Validate() {
			stats.Invalid++
			continue
		}

		stats.ByPLMN[record.GetPLMN()]++
		if record.Base != nil && record.Base.Type != nil {
			stats.ByType[*record.Base.Type]++
		}
	}

	return stats
}

func (array Array) Types() []apnxml.ObjectBaseType {
	typeSet := map[apnxml.ObjectBaseType]bool{}
	for _, record := range flatten(array.data) {
		if record.Base != nil && record.Base.Type != nil {
			typeSet[*record.Base.Type] = true
		}
	}

	result := make([]apnxml.ObjectBaseType, 0, len(typeSet))
	for apnType := range typeSet {
		result = append(result, apnType)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result
}

func (array Array) PLMNs() []string {
	plmnSet := map[string]bool{}
	for _, record := range flatten(array.data) {
		if record.ObjectRoot != nil && record.ObjectRoot.Validate() {
			plmnSet[record.GetPLMN()] = true
		}
	}

	result := make([]string, 0, len(plmnSet))
	for plmn := range plmnSet {
		result = append(result, plmn)
	}

	sort.Strings(result)
	return result
}

func (array Array) CarrierIDs() []int {
	carrierIDSet := map[int]bool{}
	for _, record := range flatten(array.data) {
		if record.CarrierID != nil {
			carrierIDSet[*record.CarrierID] = true
		}
	}

	result := make([]int, 0, len(carrierIDSet))
	for carrierID := range carrierIDSet {
		result = append(result, carrierID)
	}

	sort.Ints(result)
	return result
}
