package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/GlshchnkLx/go-aospapn/pkg/apntool"
	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

func buildPredicate(filters *filterFlags) (apntool.Predicate, error) {
	var predicates []apntool.Predicate
	if len(filters.plmn) > 0 {
		var plmnPredicates []apntool.Predicate
		for _, plmn := range filters.plmn {
			mcc, mnc, err := parsePLMN(plmn)
			if err != nil {
				return nil, err
			}
			plmnPredicates = append(plmnPredicates, apntool.ByPLMN(mcc, mnc))
		}
		predicates = append(predicates, apntool.Or(plmnPredicates...))
	}
	if filters.mcc >= 0 {
		predicates = append(predicates, apntool.ByMCC(filters.mcc))
	}
	if filters.mnc >= 0 {
		predicates = append(predicates, apntool.ByMNC(filters.mnc))
	}
	if filters.carrierID >= 0 {
		predicates = append(predicates, apntool.ByCarrierID(filters.carrierID))
	}
	if filters.carrier != "" {
		predicates = append(predicates, apntool.ByCarrierName(filters.carrier))
	}
	if filters.apn != "" {
		predicates = append(predicates, apntool.ByAPN(filters.apn))
	}
	if filters.apnContains != "" {
		predicates = append(predicates, apntool.ByAPNContains(filters.apnContains))
	}
	if filters.apnType != "" {
		value, err := apnxml.ParseObjectBaseType(filters.apnType)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, apntool.ByType(value))
	}
	if filters.protocol != "" {
		value, err := apnxml.ParseObjectBearerProtocol(filters.protocol)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, apntool.ByProtocol(value))
	}
	if filters.network != "" {
		value, err := apnxml.ParseObjectNetworkType(filters.network)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, apntool.ByNetwork(value))
	}
	if filters.validOnly && filters.invalidOnly {
		return nil, fmt.Errorf("--valid-only and --invalid-only are mutually exclusive")
	}
	if filters.validOnly {
		predicates = append(predicates, apntool.IsValid)
	}
	if filters.invalidOnly {
		predicates = append(predicates, apntool.Not(apntool.IsValid))
	}
	for _, section := range filters.has {
		predicate, err := hasPredicate(section)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, predicate)
	}
	for _, section := range filters.without {
		predicate, err := hasPredicate(section)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, apntool.Not(predicate))
	}

	predicate := apntool.And(predicates...)
	if filters.invert {
		return apntool.Not(predicate), nil
	}
	return predicate, nil
}

func hasPredicate(value string) (apntool.Predicate, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "root":
		return apntool.HasRoot, nil
	case "valid-root":
		return apntool.HasValidRoot, nil
	case "base":
		return apntool.HasBase, nil
	case "auth":
		return apntool.HasAuth, nil
	case "bearer":
		return apntool.HasBearer, nil
	case "proxy":
		return apntool.HasProxy, nil
	case "mms":
		return apntool.HasMMS, nil
	case "mvno":
		return apntool.HasMVNO, nil
	default:
		return nil, fmt.Errorf("unsupported section: %s", value)
	}
}

func parsePLMN(value string) (int, int, error) {
	value = strings.TrimSpace(value)
	if len(value) != 5 && len(value) != 6 {
		return 0, 0, fmt.Errorf("PLMN must be MCCMNC with 5 or 6 digits")
	}
	mcc, err := strconv.Atoi(value[:3])
	if err != nil {
		return 0, 0, err
	}
	mnc, err := strconv.Atoi(value[3:])
	if err != nil {
		return 0, 0, err
	}
	return mcc, mnc, nil
}
