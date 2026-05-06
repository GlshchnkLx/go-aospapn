package apnxml

import (
	"fmt"
	"strings"
)

//--------------------------------------------------------------------------------//
// Parse
//--------------------------------------------------------------------------------//

func ParseFormat(value string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "json":
		return FormatJSON, nil
	case "xml":
		return FormatXML, nil
	default:
		return "", fmt.Errorf("unsupported apn format: %s", value)
	}
}

func ParseObjectUpdateMode(value string) (ObjectUpdateMode, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "patch":
		return ObjectUpdatePatch, nil
	case "merge":
		return ObjectUpdateMerge, nil
	case "apply":
		return ObjectUpdateApply, nil
	default:
		return 0, fmt.Errorf("unsupported apn update mode: %s", value)
	}
}

func ParseObjectBaseType(value string) (ObjectBaseType, error) {
	var result ObjectBaseType
	if err := result.UnmarshalText([]byte(value)); err != nil {
		return 0, err
	}
	return result, nil
}

func ParseObjectAuthType(value string) (ObjectAuthType, error) {
	var result ObjectAuthType
	if err := result.UnmarshalText([]byte(value)); err != nil {
		return 0, err
	}
	return result, nil
}

func ParseObjectBearerProtocol(value string) (ObjectBearerProtocol, error) {
	var result ObjectBearerProtocol
	if err := result.UnmarshalText([]byte(value)); err != nil {
		return 0, err
	}
	return result, nil
}

func ParseObjectNetworkType(value string) (ObjectNetworkType, error) {
	var result ObjectNetworkType
	if err := result.UnmarshalText([]byte(value)); err != nil {
		return 0, err
	}
	return result, nil
}

//--------------------------------------------------------------------------------//
