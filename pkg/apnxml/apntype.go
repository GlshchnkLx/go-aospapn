// # APN Type
//
// File apntype provides strongly-typed, configurable enum and bitmask types for APN
// configuration values, with built-in support for XML and JSON marshaling/unmarshaling.
//
// It includes:
//   - APNTypeCoreOption: Configurable formatting options for serialization (XML/JSON/text).
//   - APNTypeCoreEnum: Generic enum type with bidirectional string↔int mapping.
//   - APNTypeCoreBitmask: Generic bitmask type supporting multi-flag combinations.
//   - Predefined types: APNTypeBaseType, APNTypeAuthType, APNTypeBearerProtocol.
//
// All types implement encoding.TextMarshaler, encoding.TextUnmarshaler,
// xml.MarshalerAttr, and xml.UnmarshalerAttr for seamless integration with Go’s standard libraries.
package apnxml

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//--------------------------------------------------------------------------------//
// APNType APNTypeCoreOption
//--------------------------------------------------------------------------------//

// APNTypeCoreOption holds formatting and serialization options for APN types in XML and text (JSON/logs).
// It is embedded into APNTypeCoreEnum and APNTypeCoreBitmask to provide consistent configuration.
type APNTypeCoreOption struct {
	XMLUseNumber  bool
	XMLUseUpper   bool
	XMLSeparator  string
	TextSeparator string
}

// NewAPNTypeCoreOption creates a new option set with default values:
//   - XMLUseNumber: false — serialize as string names, not numbers.
//   - XMLUseUpper: false — preserve original case in XML.
//   - XMLSeparator: "," — separator for bitmask values in XML attributes.
//   - TextSeparator: "," — separator for bitmask values in JSON/text output.
//
// Use Set* methods to customize behavior.
func NewAPNTypeCoreOption() *APNTypeCoreOption {
	return &APNTypeCoreOption{
		XMLUseNumber:  false,
		XMLUseUpper:   false,
		XMLSeparator:  ",",
		TextSeparator: ",",
	}
}

// SetXMLUseNumber enables or disables numeric output in XML marshaling.
// When true, bitmask/enum values are serialized as integers (e.g., "3" instead of "pap,chap").
// Returns the receiver for chaining.
func (apnTypeOption *APNTypeCoreOption) SetXMLUseNumber(value bool) *APNTypeCoreOption {
	apnTypeOption.XMLUseNumber = value
	return apnTypeOption
}

// SetXMLUseUpper enables or disables uppercase conversion for string values in XML.
// When true, string representations are converted to uppercase (e.g., "IPV4V6").
// Returns the receiver for chaining.
func (apnTypeOption *APNTypeCoreOption) SetXMLUseUpper(value bool) *APNTypeCoreOption {
	apnTypeOption.XMLUseUpper = value
	return apnTypeOption
}

// SetXMLSeparator sets the separator string used when marshaling bitmask values to XML attributes.
// Applied only when XMLUseNumber=false. If value is empty, the separator is not changed.
func (apnTypeOption *APNTypeCoreOption) SetXMLSeparator(value string) *APNTypeCoreOption {
	if value != "" {
		apnTypeOption.XMLSeparator = value
	}

	return apnTypeOption
}

// SetTextSeparator sets the separator string used when marshaling bitmask values to text formats (JSON, logs).
// Applied in String(), MarshalJSON(), etc. If value is empty, the separator is not changed.
func (apnTypeOption *APNTypeCoreOption) SetTextSeparator(value string) *APNTypeCoreOption {
	if value != "" {
		apnTypeOption.TextSeparator = value
	}

	return apnTypeOption
}

//--------------------------------------------------------------------------------//
// APNType APNTypeCoreEnum
//--------------------------------------------------------------------------------//

// APNTypeCoreEnum represents an enumeration of APN-related values (e.g., bearer protocol).
// It supports bidirectional mapping between integer indices and string representations,
// with customizable XML/JSON serialization behavior via embedded APNTypeCoreOption.
type APNTypeCoreEnum[Type ~int] struct {
	IndexNone    Type
	IndexDefault Type
	IndexMax     Type

	MapByIndex  map[Type]string
	MapByString map[string]Type

	*APNTypeCoreOption
}

// NewAPNTypeCoreEnum creates a new enum configuration.
// If option is nil, a default APNTypeCoreOption is used.
//
// Parameters:
//   - indexNone: value representing "none" or "invalid".
//   - indexDefault: default/wildcard value (e.g., "*" maps to this).
//   - indexMax: exclusive upper bound of valid values (used for validation).
//   - mapByIndex: map of valid enum values to their string representations.
//   - option: optional configuration for XML/JSON formatting.
//
// Automatically adds "" → IndexNone and "*" → IndexDefault to MapByString.
func NewAPNTypeCoreEnum[Type ~int](indexNone Type, indexDefault Type, indexMax Type, mapByIndex map[Type]string, option *APNTypeCoreOption) *APNTypeCoreEnum[Type] {
	var (
		apnTypeEnum = &APNTypeCoreEnum[Type]{
			IndexNone:         indexNone,
			IndexDefault:      indexDefault,
			IndexMax:          indexMax,
			MapByIndex:        mapByIndex,
			MapByString:       map[string]Type{},
			APNTypeCoreOption: option,
		}
	)

	if apnTypeEnum.APNTypeCoreOption == nil {
		apnTypeEnum.APNTypeCoreOption = NewAPNTypeCoreOption()
	}

	for apnTypeIndex, apnTypeString := range mapByIndex {
		apnTypeString = strings.TrimSpace(strings.ToLower(apnTypeString))
		apnTypeEnum.MapByString[apnTypeString] = apnTypeIndex
	}

	apnTypeEnum.MapByString[""] = indexNone
	apnTypeEnum.MapByString["*"] = indexDefault

	return apnTypeEnum
}

// GetString returns the string representation of the given enum value.
// If the value is not found, returns the string for IndexNone.
func (apnTypeEnum *APNTypeCoreEnum[Type]) GetString(apnTypeIndex Type) string {
	var (
		apnTypeString string
		ok            bool
	)

	apnTypeString, ok = apnTypeEnum.MapByIndex[apnTypeIndex]
	if !ok {
		return apnTypeEnum.MapByIndex[apnTypeEnum.IndexNone]
	}

	return apnTypeString
}

// SetString sets the enum value by parsing a string.
// The input string is trimmed and converted to lowercase before lookup.
// Returns an error if the string is not recognized.
func (apnTypeEnum *APNTypeCoreEnum[Type]) SetString(apnTypeIndex *Type, apnTypeString string) error {
	var (
		ok bool
	)

	apnTypeString = strings.ToLower(strings.TrimSpace((apnTypeString)))

	*apnTypeIndex, ok = apnTypeEnum.MapByString[apnTypeString]
	if !ok {
		return fmt.Errorf("apn type has incorrect string: %q", apnTypeString)
	}

	return nil
}

// MarshalXMLValue serializes the enum value to an XML attribute.
// Behavior depends on embedded APNTypeCoreOption:
//   - If XMLUseNumber is true, outputs the integer value.
//   - Otherwise, outputs the string representation (optionally uppercased if XMLUseUpper is true).
func (apnTypeEnum *APNTypeCoreEnum[Type]) MarshalXMLValue(apnTypeIndex Type, xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	var (
		apnTypeString string
	)

	if apnTypeEnum.XMLUseNumber {
		apnTypeString = strconv.Itoa(int(apnTypeIndex))
	} else {
		apnTypeString = apnTypeEnum.GetString(apnTypeIndex)
	}

	if !apnTypeEnum.XMLUseNumber && apnTypeEnum.XMLUseUpper {
		apnTypeString = strings.ToUpper(apnTypeString)
	}

	return xml.Attr{
		Name:  xmlAttrName,
		Value: apnTypeString,
	}, nil
}

// UnmarshalXMLValue parses an XML attribute into the enum value.
// If XMLUseNumber is true, expects an integer; otherwise, expects a string.
// Empty values are treated as IndexNone.
func (apnTypeEnum *APNTypeCoreEnum[Type]) UnmarshalXMLValue(apnTypeIndex *Type, xmlAttr xml.Attr) error {
	var (
		_apnTypeIndex int
		err           error
	)

	if xmlAttr.Value == "" {
		*apnTypeIndex = apnTypeEnum.IndexNone
	} else {
		if apnTypeEnum.XMLUseNumber {
			_apnTypeIndex, err = strconv.Atoi(xmlAttr.Value)
			if err != nil {
				return fmt.Errorf("apn auth type has invalid number: %v", err)
			}

			if _apnTypeIndex < int(apnTypeEnum.IndexNone) || int(apnTypeEnum.IndexMax) <= _apnTypeIndex {
				return fmt.Errorf("apn auth type has out of range number: %d", _apnTypeIndex)
			}

			*apnTypeIndex = Type(_apnTypeIndex)
		} else {
			return apnTypeEnum.SetString(apnTypeIndex, xmlAttr.Value)
		}
	}

	return nil
}

// MarshalJSONValue serializes the enum value to JSON as its string representation.
func (apnTypeEnum *APNTypeCoreEnum[Type]) MarshalJSONValue(apnTypeIndex Type) (jsonByte []byte, err error) {
	return json.Marshal(apnTypeEnum.GetString(apnTypeIndex))
}

// UnmarshalJSONValue parses a JSON string into the enum value.
func (apnTypeEnum *APNTypeCoreEnum[Type]) UnmarshalJSONValue(apnTypeIndex *Type, jsonByte []byte) error {
	var (
		apnTypeString string
		err           error
	)

	err = json.Unmarshal(jsonByte, &apnTypeString)
	if err != nil {
		return err
	}

	return apnTypeEnum.SetString(apnTypeIndex, apnTypeString)
}

//--------------------------------------------------------------------------------//
// APNType APNTypeCoreBitmask
//--------------------------------------------------------------------------------//

// APNTypeCoreBitmask represents a bitmask of APN-related flags (e.g., APN types, auth types).
// It supports combining multiple flags, with customizable string serialization via separators.
// XML/JSON formatting is controlled via embedded APNTypeCoreOption.
//
// Flags are expected to be powers of two (1, 2, 4, 8, ...).
// Supports parsing from comma/pipe-separated strings or JSON arrays.
type APNTypeCoreBitmask[Type ~int] struct {
	IndexNone    Type
	IndexDefault Type
	IndexAll     Type
	IndexMax     Type

	ArrayIndex  []Type
	MapByIndex  map[Type]string
	MapByString map[string]Type

	*APNTypeCoreOption
}

// NewAPNTypeCoreBitmask creates a new bitmask configuration.
// If option is nil, a default APNTypeCoreOption is used.
//
// Parameters:
//   - indexNone: value representing "no flags".
//   - indexDefault: default/wildcard flag.
//   - indexAll: value representing all flags combined.
//   - indexMax: exclusive upper bound of valid flag values.
//   - mapByIndex: map of valid flag values to their string representations.
//   - option: optional configuration for XML/JSON formatting.
//
// Automatically adds "" → IndexNone and "*" → IndexAll to MapByString.
// Sorts ArrayIndex for consistent iteration order.
func NewAPNTypeCoreBitmask[Type ~int](indexNone Type, indexDefault Type, indexAll Type, indexMax Type, mapByIndex map[Type]string, option *APNTypeCoreOption) *APNTypeCoreBitmask[Type] {
	var (
		apnTypeBitmask = APNTypeCoreBitmask[Type]{
			IndexNone:         indexNone,
			IndexDefault:      indexDefault,
			IndexAll:          indexAll,
			IndexMax:          indexMax,
			ArrayIndex:        []Type{},
			MapByIndex:        mapByIndex,
			MapByString:       map[string]Type{},
			APNTypeCoreOption: option,
		}
	)

	if apnTypeBitmask.APNTypeCoreOption == nil {
		apnTypeBitmask.APNTypeCoreOption = NewAPNTypeCoreOption()
	}

	for apnTypeIndex, apnTypeString := range mapByIndex {
		apnTypeString = strings.TrimSpace(strings.ToLower(apnTypeString))

		if apnTypeBitmask.IndexNone < apnTypeIndex && apnTypeIndex < apnTypeBitmask.IndexMax {
			apnTypeBitmask.ArrayIndex = append(apnTypeBitmask.ArrayIndex, apnTypeIndex)
		}

		apnTypeBitmask.MapByString[apnTypeString] = apnTypeIndex
	}

	apnTypeBitmask.MapByString[""] = indexNone
	apnTypeBitmask.MapByString["*"] = indexAll

	sort.Slice(apnTypeBitmask.ArrayIndex, func(i, j int) bool {
		return apnTypeBitmask.ArrayIndex[i] < apnTypeBitmask.ArrayIndex[j]
	})

	return &apnTypeBitmask
}

// GetStringByIndex returns the string representation of a single flag.
// If the flag is not found, returns the string for IndexNone.
func (apnTypeBitmask *APNTypeCoreBitmask[Type]) GetStringByIndex(apnTypeIndex Type) string {
	var (
		apnTypeString string
		ok            bool
	)

	apnTypeString, ok = apnTypeBitmask.MapByIndex[apnTypeIndex]
	if !ok {
		apnTypeString = apnTypeBitmask.MapByIndex[apnTypeBitmask.IndexNone]
	}

	return apnTypeString
}

// GetStringArray returns a slice of strings representing all set flags in the bitmask.
// If no flags are set, returns a slice containing the string for IndexNone.
func (apnTypeBitmask *APNTypeCoreBitmask[Type]) GetStringArray(apnTypeIndex Type) []string {
	var (
		apnTypeStringArray []string
	)

	for _, _apnTypeIndex := range apnTypeBitmask.ArrayIndex {
		if apnTypeIndex&_apnTypeIndex == _apnTypeIndex {
			apnTypeStringArray = append(apnTypeStringArray, apnTypeBitmask.GetStringByIndex(_apnTypeIndex))
		}
	}

	if len(apnTypeStringArray) == 0 {
		apnTypeStringArray = append(apnTypeStringArray, apnTypeBitmask.MapByIndex[apnTypeBitmask.IndexNone])
	}

	return apnTypeStringArray
}

// GetString returns a string representation of the bitmask, joining flag strings with the given separator.
func (apnTypeBitmask *APNTypeCoreBitmask[Type]) GetString(apnTypeIndex Type, separator string) string {
	return strings.Join(apnTypeBitmask.GetStringArray(apnTypeIndex), separator)
}

// SetStringArray sets the bitmask value by parsing a slice of strings.
// Each string is trimmed and lowercased before lookup.
// Returns an error if any string is unrecognized.
func (apnTypeBitmask *APNTypeCoreBitmask[Type]) SetStringArray(apnTypeIndex *Type, apnTypeStringArray []string) error {
	*apnTypeIndex = apnTypeBitmask.IndexNone

	for _, apnTypeString := range apnTypeStringArray {
		apnTypeString = strings.ToLower(strings.TrimSpace((apnTypeString)))

		_apnTypeIndex, ok := apnTypeBitmask.MapByString[apnTypeString]
		if !ok {
			return fmt.Errorf("apn type has incorrect string: %q", apnTypeString)
		}

		*apnTypeIndex |= _apnTypeIndex
	}

	return nil
}

// SetString sets the bitmask value by parsing a string with the given separator.
func (apnTypeBitmask *APNTypeCoreBitmask[Type]) SetString(apnTypeIndex *Type, apnTypeString string, separator string) error {
	return apnTypeBitmask.SetStringArray(apnTypeIndex, strings.Split(apnTypeString, separator))
}

// GetText returns the bitmask as a string using the embedded TextSeparator.
func (apnTypeBitmask *APNTypeCoreBitmask[Type]) GetText(apnTypeIndex Type) string {
	return apnTypeBitmask.GetString(apnTypeIndex, apnTypeBitmask.TextSeparator)
}

// SetText sets the bitmask value by parsing a string using the embedded TextSeparator.
func (apnTypeBitmask *APNTypeCoreBitmask[Type]) SetText(apnTypeIndex *Type, apnTypeString string) error {
	return apnTypeBitmask.SetString(apnTypeIndex, apnTypeString, apnTypeBitmask.TextSeparator)
}

// MarshalXMLValue serializes the bitmask to an XML attribute.
// Behavior depends on embedded APNTypeCoreOption:
//   - If XMLUseNumber is true, outputs the integer value.
//   - Otherwise, outputs a string of flags joined by XMLSeparator (optionally uppercased).
func (apnTypeBitmask *APNTypeCoreBitmask[Type]) MarshalXMLValue(apnTypeIndex Type, xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	var (
		apnTypeString string
	)

	if apnTypeBitmask.XMLUseNumber {
		apnTypeString = strconv.Itoa(int(apnTypeIndex))
	} else {
		apnTypeString = apnTypeBitmask.GetString(apnTypeIndex, apnTypeBitmask.XMLSeparator)
	}

	if !apnTypeBitmask.XMLUseNumber && apnTypeBitmask.XMLUseUpper {
		apnTypeString = strings.ToUpper(apnTypeString)
	}

	return xml.Attr{
		Name:  xmlAttrName,
		Value: apnTypeString,
	}, nil
}

// UnmarshalXMLValue parses an XML attribute into the bitmask value.
// If XMLUseNumber is true, expects an integer; otherwise, expects a string split by XMLSeparator.
// Empty values are treated as IndexNone.
func (apnTypeBitmask *APNTypeCoreBitmask[Type]) UnmarshalXMLValue(apnTypeIndex *Type, xmlAttr xml.Attr) error {
	var (
		_apnTypeIndex int
		err           error
	)

	if xmlAttr.Value == "" {
		*apnTypeIndex = apnTypeBitmask.IndexNone
	} else {
		if apnTypeBitmask.XMLUseNumber {
			_apnTypeIndex, err = strconv.Atoi(xmlAttr.Value)
			if err != nil {
				return fmt.Errorf("apn auth type has invalid number: %v", err)
			}

			if _apnTypeIndex < int(apnTypeBitmask.IndexNone) || int(apnTypeBitmask.IndexMax) <= _apnTypeIndex {
				return fmt.Errorf("apn auth type has out of range number: %d", _apnTypeIndex)
			}

			*apnTypeIndex = Type(_apnTypeIndex)
		} else {
			return apnTypeBitmask.SetString(apnTypeIndex, xmlAttr.Value, apnTypeBitmask.XMLSeparator)
		}
	}

	return nil
}

// MarshalJSONValue serializes the bitmask to JSON as an array of flag strings.
func (apnTypeBitmask *APNTypeCoreBitmask[Type]) MarshalJSONValue(apnTypeIndex Type) (jsonByte []byte, err error) {
	return json.Marshal(apnTypeBitmask.GetStringArray(apnTypeIndex))
}

// UnmarshalJSONValue parses a JSON array of strings into the bitmask value.
func (apnTypeBitmask *APNTypeCoreBitmask[Type]) UnmarshalJSONValue(apnTypeIndex *Type, jsonByte []byte) error {
	var (
		apnTypeStringArray []string
		err                error
	)

	err = json.Unmarshal(jsonByte, &apnTypeStringArray)
	if err != nil {
		return err
	}

	return apnTypeBitmask.SetStringArray(apnTypeIndex, apnTypeStringArray)
}

//--------------------------------------------------------------------------------//
// APNType BaseType
//--------------------------------------------------------------------------------//

// APNTypeBaseType represents bitmask flags for APN types (e.g., "default", "mms", "supl").
// Commonly used in Android APN configuration files to specify APN purpose.
type APNTypeBaseType int

const (
	APNTYPE_BASE_TYPE_NONE    APNTypeBaseType = 0
	APNTYPE_BASE_TYPE_DEFAULT APNTypeBaseType = 1 << (iota - 1)
	APNTYPE_BASE_TYPE_MMS
	APNTYPE_BASE_TYPE_SUPL
	APNTYPE_BASE_TYPE_DUN
	APNTYPE_BASE_TYPE_HIPRI
	APNTYPE_BASE_TYPE_FOTA
	APNTYPE_BASE_TYPE_IMS
	APNTYPE_BASE_TYPE_CBS
	APNTYPE_BASE_TYPE_IA
	APNTYPE_BASE_TYPE_EMERGENCY
	APNTYPE_BASE_TYPE_MCX
	APNTYPE_BASE_TYPE_XCAP
	APNTYPE_BASE_TYPE_VSIM
	APNTYPE_BASE_TYPE_BIP
	APNTYPE_BASE_TYPE_ENTERPRISE
	APNTYPE_BASE_TYPE_RCS
	APNTYPE_BASE_TYPE_OEMPAID
	APNTYPE_BASE_TYPE_OEMPRIVATE
	APNTYPE_BASE_TYPE_MAX

	APNTYPE_BASE_TYPE_ALL = APNTYPE_BASE_TYPE_MAX - 1
)

// APNTypeBaseTypeConfig is the shared configuration for APNTypeBaseType,
// using comma separators and string values in XML/JSON.
var APNTypeBaseTypeConfig = NewAPNTypeCoreBitmask(
	APNTYPE_BASE_TYPE_NONE,
	APNTYPE_BASE_TYPE_DEFAULT,
	APNTYPE_BASE_TYPE_ALL,
	APNTYPE_BASE_TYPE_MAX,
	map[APNTypeBaseType]string{
		APNTYPE_BASE_TYPE_NONE:       "none",
		APNTYPE_BASE_TYPE_DEFAULT:    "default",
		APNTYPE_BASE_TYPE_MMS:        "mms",
		APNTYPE_BASE_TYPE_SUPL:       "supl",
		APNTYPE_BASE_TYPE_DUN:        "dun",
		APNTYPE_BASE_TYPE_HIPRI:      "hipri",
		APNTYPE_BASE_TYPE_FOTA:       "fota",
		APNTYPE_BASE_TYPE_IMS:        "ims",
		APNTYPE_BASE_TYPE_CBS:        "cbs",
		APNTYPE_BASE_TYPE_IA:         "ia",
		APNTYPE_BASE_TYPE_EMERGENCY:  "emergency",
		APNTYPE_BASE_TYPE_MCX:        "mcx",
		APNTYPE_BASE_TYPE_XCAP:       "xcap",
		APNTYPE_BASE_TYPE_VSIM:       "vsim",
		APNTYPE_BASE_TYPE_BIP:        "bip",
		APNTYPE_BASE_TYPE_ENTERPRISE: "enterprise",
		APNTYPE_BASE_TYPE_RCS:        "rcs",
		APNTYPE_BASE_TYPE_OEMPAID:    "oem_paid",
		APNTYPE_BASE_TYPE_OEMPRIVATE: "oem_private",
	},
	NewAPNTypeCoreOption().SetTextSeparator(",").SetXMLSeparator(","),
)

// GetStringArray returns the list of APN type flags set in this value.
func (apnBaseType APNTypeBaseType) GetStringArray() []string {
	return APNTypeBaseTypeConfig.GetStringArray(apnBaseType)
}

// GetString returns a string representation of the APN type flags, joined by the given separator.
func (apnBaseType APNTypeBaseType) GetString(separator string) string {
	return APNTypeBaseTypeConfig.GetString(apnBaseType, separator)
}

// SetStringArray sets the APN type flags from a slice of strings.
func (apnBaseType *APNTypeBaseType) SetStringArray(apnBaseTypeStringArray []string) error {
	return APNTypeBaseTypeConfig.SetStringArray(apnBaseType, apnBaseTypeStringArray)
}

// SetString sets the APN type flags from a string, split by the given separator.
func (apnBaseType *APNTypeBaseType) SetString(apnBaseTypeString string, separator string) error {
	return APNTypeBaseTypeConfig.SetStringArray(apnBaseType, strings.Split(apnBaseTypeString, separator))
}

// String returns a comma-separated string of APN type flags (e.g., "default,mms").
func (apnBaseType APNTypeBaseType) String() string {
	return APNTypeBaseTypeConfig.GetText(apnBaseType)
}

// MarshalXMLAttr implements xml.MarshalerAttr for use in XML attributes.
func (apnBaseType APNTypeBaseType) MarshalXMLAttr(xmlAttr xml.Name) (attr xml.Attr, err error) {
	return APNTypeBaseTypeConfig.MarshalXMLValue(apnBaseType, xmlAttr)
}

// UnmarshalXMLAttr implements xml.UnmarshalerAttr for parsing from XML attributes.
func (apnBaseType *APNTypeBaseType) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return APNTypeBaseTypeConfig.UnmarshalXMLValue(apnBaseType, xmlAttr)
}

// MarshalJSON implements json.Marshaler for serialization to JSON.
func (apnBaseType APNTypeBaseType) MarshalJSON() ([]byte, error) {
	return APNTypeBaseTypeConfig.MarshalJSONValue(apnBaseType)
}

// UnmarshalJSON implements json.Unmarshaler for parsing from JSON.
func (apnBaseType *APNTypeBaseType) UnmarshalJSON(jsonData []byte) error {
	return APNTypeBaseTypeConfig.UnmarshalJSONValue(apnBaseType, jsonData)
}

//--------------------------------------------------------------------------------//
// APNType AuthType
//--------------------------------------------------------------------------------//

// APNTypeAuthType represents bitmask flags for APN authentication types (PAP, CHAP).
// Used to specify which authentication protocols are enabled for an APN.
type APNTypeAuthType int

const (
	APNTYPE_AUTH_TYPE_NONE APNTypeAuthType = 0
	APNTYPE_AUTH_TYPE_PAP  APNTypeAuthType = 1 << (iota - 1)
	APNTYPE_AUTH_TYPE_CHAP
	APNTYPE_AUTH_TYPE_MAX

	APNTYPE_AUTH_TYPE_ALL = APNTYPE_AUTH_TYPE_MAX - 1
)

// APNTypeAuthTypeConfig is the shared configuration for APNTypeAuthType,
// using numeric values in XML and comma separators in JSON/text.
var APNTypeAuthTypeConfig = NewAPNTypeCoreBitmask(
	APNTYPE_AUTH_TYPE_NONE,
	APNTYPE_AUTH_TYPE_PAP,
	APNTYPE_AUTH_TYPE_ALL,
	APNTYPE_AUTH_TYPE_MAX,
	map[APNTypeAuthType]string{
		APNTYPE_AUTH_TYPE_NONE: "none",
		APNTYPE_AUTH_TYPE_PAP:  "pap",
		APNTYPE_AUTH_TYPE_CHAP: "chap",
	},
	NewAPNTypeCoreOption().
		SetTextSeparator(",").
		SetXMLUseNumber(true).
		SetXMLSeparator("|"),
)

// GetStringArray returns the list of authentication type flags set in this value.
func (apnAuthType APNTypeAuthType) GetStringArray() []string {
	return APNTypeAuthTypeConfig.GetStringArray(apnAuthType)
}

// GetString returns a string representation of the auth type flags, joined by the given separator.
func (apnAuthType APNTypeAuthType) GetString(separator string) string {
	return APNTypeAuthTypeConfig.GetString(apnAuthType, separator)
}

// SetStringArray sets the auth type flags from a slice of strings.
func (apnAuthType *APNTypeAuthType) SetStringArray(apnAuthTypeStringArray []string) error {
	return APNTypeAuthTypeConfig.SetStringArray(apnAuthType, apnAuthTypeStringArray)
}

// SetString sets the auth type flags from a string, split by the given separator.
func (apnAuthType *APNTypeAuthType) SetString(apnAuthTypeString string, separator string) error {
	return APNTypeAuthTypeConfig.SetStringArray(apnAuthType, strings.Split(apnAuthTypeString, separator))
}

// String returns a comma-separated string of auth type flags (e.g., "pap,chap").
func (apnAuthType APNTypeAuthType) String() string {
	return APNTypeAuthTypeConfig.GetText(apnAuthType)
}

// MarshalXMLAttr implements xml.MarshalerAttr for use in XML attributes.
func (apnAuthType APNTypeAuthType) MarshalXMLAttr(xmlAttr xml.Name) (attr xml.Attr, err error) {
	return APNTypeAuthTypeConfig.MarshalXMLValue(apnAuthType, xmlAttr)
}

// UnmarshalXMLAttr implements xml.UnmarshalerAttr for parsing from XML attributes.
func (apnAuthType *APNTypeAuthType) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return APNTypeAuthTypeConfig.UnmarshalXMLValue(apnAuthType, xmlAttr)
}

// MarshalJSON implements json.Marshaler for serialization to JSON.
func (apnAuthType APNTypeAuthType) MarshalJSON() ([]byte, error) {
	return APNTypeAuthTypeConfig.MarshalJSONValue(apnAuthType)
}

// UnmarshalJSON implements json.Unmarshaler for parsing from JSON.
func (apnAuthType *APNTypeAuthType) UnmarshalJSON(jsonData []byte) error {
	return APNTypeAuthTypeConfig.UnmarshalJSONValue(apnAuthType, jsonData)
}

//--------------------------------------------------------------------------------//
// APNType BearerProtocol
//--------------------------------------------------------------------------------//

// APNTypeBearerProtocol represents an enumeration of APN bearer protocols (e.g., "IP", "IPv6").
// Specifies the IP protocol version or type used for the data connection.
type APNTypeBearerProtocol int

const (
	APNTYPE_BEARER_PROTOCOL_NONE APNTypeBearerProtocol = iota
	APNTYPE_BEARER_PROTOCOL_IP
	APNTYPE_BEARER_PROTOCOL_IPV4
	APNTYPE_BEARER_PROTOCOL_IPV6
	APNTYPE_BEARER_PROTOCOL_IPV4V6
	APNTYPE_BEARER_PROTOCOL_PPP
	APNTYPE_BEARER_PROTOCOL_NONIP
	APNTYPE_BEARER_PROTOCOL_UNSTRUCTURED
	APNTYPE_BEARER_PROTOCOL_MAX
)

// APNTypeBearerProtocolConfig is the shared configuration for APNTypeBearerProtocol,
// using uppercase string values in XML (e.g., "IPV4V6").
var APNTypeBearerProtocolConfig = NewAPNTypeCoreEnum(
	APNTYPE_BEARER_PROTOCOL_NONE,
	APNTYPE_BEARER_PROTOCOL_IPV4V6,
	APNTYPE_BEARER_PROTOCOL_MAX,
	map[APNTypeBearerProtocol]string{
		APNTYPE_BEARER_PROTOCOL_NONE:         "none",
		APNTYPE_BEARER_PROTOCOL_IP:           "ip",
		APNTYPE_BEARER_PROTOCOL_IPV4:         "ipv4",
		APNTYPE_BEARER_PROTOCOL_IPV6:         "ipv6",
		APNTYPE_BEARER_PROTOCOL_IPV4V6:       "ipv4v6",
		APNTYPE_BEARER_PROTOCOL_PPP:          "ppp",
		APNTYPE_BEARER_PROTOCOL_NONIP:        "non-ip",
		APNTYPE_BEARER_PROTOCOL_UNSTRUCTURED: "unstructured",
	},
	NewAPNTypeCoreOption().SetXMLUseUpper(true),
)

// GetString returns the string representation of the bearer protocol.
func (apnBearerProtocol APNTypeBearerProtocol) GetString() string {
	return APNTypeBearerProtocolConfig.GetString(apnBearerProtocol)
}

// SetString sets the bearer protocol by parsing a string (case-insensitive).
func (apnBearerProtocol *APNTypeBearerProtocol) SetString(apnBearerProtocolString string) error {
	return APNTypeBearerProtocolConfig.SetString(apnBearerProtocol, apnBearerProtocolString)
}

// String returns the string representation of the bearer protocol.
func (apnBearerProtocol APNTypeBearerProtocol) String() string {
	return APNTypeBearerProtocolConfig.GetString(apnBearerProtocol)
}

// MarshalXMLAttr implements xml.MarshalerAttr for use in XML attributes.
func (apnBearerProtocol APNTypeBearerProtocol) MarshalXMLAttr(xmlAttr xml.Name) (attr xml.Attr, err error) {
	return APNTypeBearerProtocolConfig.MarshalXMLValue(apnBearerProtocol, xmlAttr)
}

// UnmarshalXMLAttr implements xml.UnmarshalerAttr for parsing from XML attributes.
func (apnBearerProtocol *APNTypeBearerProtocol) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return APNTypeBearerProtocolConfig.UnmarshalXMLValue(apnBearerProtocol, xmlAttr)
}

// MarshalJSON implements json.Marshaler for serialization to JSON.
func (apnBearerProtocol APNTypeBearerProtocol) MarshalJSON() ([]byte, error) {
	return APNTypeBearerProtocolConfig.MarshalJSONValue(apnBearerProtocol)
}

// UnmarshalJSON implements json.Unmarshaler for parsing from JSON.
func (apnBearerProtocol *APNTypeBearerProtocol) UnmarshalJSON(jsonData []byte) error {
	return APNTypeBearerProtocolConfig.UnmarshalJSONValue(apnBearerProtocol, jsonData)
}

//--------------------------------------------------------------------------------//
