// # APN Type
//
// File defines core mapping structures (APNTypeCoreMap) and proxy structures
// (APNTypeCoreProxy) that enable bidirectional conversion between integer bitmask
// values and their string representations for JSON/XML marshaling.
//
// Predefined APN types include:
//   - APNTypeBaseType: for base APN capabilities (default, mms, supl, dun, etc.)
//   - APNTypeAuthType: for authentication types (none, pap, chap)
//   - APNTypeNetworkType: for network technologies (gprs, lte, nr, etc.)
//   - APNTypeBearerProtocol: for bearer protocols (ip, ipv4, ipv6, ppp, etc.)
//
// Each type supports:
//   - String() for human-readable output
//   - MarshalJSON/UnmarshalJSON for JSON serialization
//   - MarshalXMLAttr/UnmarshalXMLAttr for XML attribute serialization
//
// The core proxy system allows configuration of serialization formats:
//   - Arrays vs single values
//   - String case (upper/lower)
//   - Numeric representation (order-based or index-based)
//   - Custom separators for array serialization
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
// APNType CoreMap
//--------------------------------------------------------------------------------//

// APNTypeCoreMap provides bidirectional mapping between integer indices and string values
// for APN types. It supports bitmask operations for multi-value types and maintains
// ordered index arrays for consistent serialization.
//
// The structure enforces bounds checking with NoneIndex (invalid/none value) and MaxIndex
// (upper bound). Only indices between NoneIndex and MaxIndex are considered valid.
//
// Type parameter must be an integer type (~int constraint).
type APNTypeCoreMap[Type ~int] struct {
	NoneIndex Type
	MaxIndex  Type

	IndexArray  []Type
	MapByIndex  map[Type]string
	MapByString map[string]Type
}

// NewAPNTypeCoreMap creates a new APNTypeCoreMap with the specified none index, max index,
// and initial mapping from indices to strings.
//
// The function:
//   - Initializes empty arrays and maps
//   - Processes the input map, trimming whitespace from string values
//   - Adds valid indices (between NoneIndex and MaxIndex) to IndexArray
//   - Populates both forward (index→string) and reverse (string→index) maps
//   - Sorts IndexArray in ascending order for consistent iteration
//
// Returns a pointer to the initialized APNTypeCoreMap.
func NewAPNTypeCoreMap[Type ~int](noneIndex Type, maxIndex Type, mapByIndex map[Type]string) *APNTypeCoreMap[Type] {
	apnTypeMap := &APNTypeCoreMap[Type]{
		NoneIndex:   noneIndex,
		MaxIndex:    maxIndex,
		IndexArray:  []Type{},
		MapByIndex:  map[Type]string{},
		MapByString: map[string]Type{},
	}

	for apnTypeIndex, apnTypeString := range mapByIndex {
		apnTypeString = strings.TrimSpace(apnTypeString)

		if noneIndex < apnTypeIndex && apnTypeIndex < maxIndex {
			apnTypeMap.IndexArray = append(apnTypeMap.IndexArray, apnTypeIndex)
		}

		apnTypeMap.MapByIndex[apnTypeIndex] = apnTypeString
		apnTypeMap.MapByString[apnTypeString] = apnTypeIndex
	}

	sort.Slice(apnTypeMap.IndexArray, func(i, j int) bool {
		return apnTypeMap.IndexArray[i] < apnTypeMap.IndexArray[j]
	})

	return apnTypeMap
}

// GetIndex returns the input index if it exists in the map, otherwise returns NoneIndex.
// This provides safe access to validate if an index is defined in the map.
func (apnTypeMap *APNTypeCoreMap[Type]) GetIndex(apnTypeValue Type) Type {
	if _, ok := apnTypeMap.MapByIndex[apnTypeValue]; ok {
		return apnTypeValue
	} else {
		return apnTypeMap.NoneIndex
	}
}

// SetIndex sets the target value to the specified index if it exists in the map.
// Returns an error if the index is not found in MapByIndex.
func (apnTypeMap *APNTypeCoreMap[Type]) SetIndex(apnTypeValue *Type, apnTypeIndex Type) error {
	if _, ok := apnTypeMap.MapByIndex[apnTypeIndex]; !ok {
		return fmt.Errorf("apn type has incorrect index: %d", apnTypeIndex)
	}

	*apnTypeValue = apnTypeIndex

	return nil
}

// GetIndexArray returns an array of indices that are set in the bitmask value.
// For each index in IndexArray, it checks if the bit is set in apnTypeValue.
// If no bits are set, returns an array containing only NoneIndex.
func (apnTypeMap *APNTypeCoreMap[Type]) GetIndexArray(apnTypeValue Type) []Type {
	apnTypeIndexArray := []Type{}

	for _, apnTypeIndex := range apnTypeMap.IndexArray {
		if apnTypeValue&apnTypeIndex == apnTypeIndex {
			apnTypeIndexArray = append(apnTypeIndexArray, apnTypeIndex)
		}
	}

	if len(apnTypeIndexArray) == 0 {
		apnTypeIndexArray = append(apnTypeIndexArray, apnTypeMap.NoneIndex)
	}

	return apnTypeIndexArray
}

// SetIndexArray sets the target value by combining all specified indices using bitwise OR.
// Returns an error if any index is not found in MapByIndex.
// The target value is first reset to NoneIndex before applying the OR operations.
func (apnTypeMap *APNTypeCoreMap[Type]) SetIndexArray(apnTypeValue *Type, apnTypeIndexArray []Type) error {
	*apnTypeValue = apnTypeMap.NoneIndex

	for _, apnTypeIndex := range apnTypeIndexArray {
		_, ok := apnTypeMap.MapByIndex[apnTypeIndex]
		if !ok {
			return fmt.Errorf("apn type has incorrect index: %d", apnTypeIndex)
		}

		*apnTypeValue |= apnTypeIndex
	}

	return nil
}

// GetString returns the string representation of the specified index.
// If the index is invalid, returns the string for NoneIndex.
func (apnTypeMap *APNTypeCoreMap[Type]) GetString(apnTypeValue Type) string {
	return apnTypeMap.MapByIndex[apnTypeMap.GetIndex(apnTypeValue)]
}

// SetString sets the target value to the index corresponding to the specified string.
// Returns an error if the string is not found in MapByString.
func (apnTypeMap *APNTypeCoreMap[Type]) SetString(apnTypeValue *Type, apnTypeString string) error {
	apnTypeIndex, ok := apnTypeMap.MapByString[apnTypeString]
	if !ok {
		return fmt.Errorf("apn type has incorrect string: %q", apnTypeString)
	}

	return apnTypeMap.SetIndex(apnTypeValue, apnTypeIndex)
}

// GetStringArray returns an array of strings corresponding to all set bits in the bitmask value.
// Uses GetIndexArray to determine which indices are set, then maps each to its string representation.
func (apnTypeMap *APNTypeCoreMap[Type]) GetStringArray(apnTypeValue Type) []string {
	var (
		apnTypeIndexArray  []Type
		apnTypeStringArray []string
	)

	apnTypeIndexArray = apnTypeMap.GetIndexArray(apnTypeValue)

	for _, apnTypeIndex := range apnTypeIndexArray {
		apnTypeStringArray = append(apnTypeStringArray, apnTypeMap.MapByIndex[apnTypeIndex])
	}

	return apnTypeStringArray
}

// SetStringArray sets the target value based on an array of strings.
// Each string is converted to lowercase and trimmed, then mapped to its index.
// Returns an error if any string is not found in MapByString.
// Uses SetIndexArray to combine all indices.
func (apnTypeMap *APNTypeCoreMap[Type]) SetStringArray(apnTypeValue *Type, apnTypeStringArray []string) error {
	var (
		apnTypeIndexArray []Type
	)

	for _, apnTypeString := range apnTypeStringArray {
		apnTypeString = strings.ToLower(strings.TrimSpace((apnTypeString)))

		apnTypeIndex, ok := apnTypeMap.MapByString[apnTypeString]
		if !ok {
			return fmt.Errorf("apn type has incorrect string: %q", apnTypeString)
		}

		apnTypeIndexArray = append(apnTypeIndexArray, apnTypeIndex)
	}

	return apnTypeMap.SetIndexArray(apnTypeValue, apnTypeIndexArray)
}

// GetValue returns the input value if it's within valid bounds (NoneIndex ≤ value < MaxIndex),
// otherwise returns NoneIndex. This provides bounds checking for raw values.
func (apnTypeMap *APNTypeCoreMap[Type]) GetValue(apnTypeValue Type) Type {
	if apnTypeValue <= apnTypeMap.NoneIndex {
		return apnTypeMap.NoneIndex
	} else if apnTypeMap.MaxIndex <= apnTypeValue {
		return apnTypeMap.NoneIndex
	} else {
		return apnTypeValue
	}
}

// SetValue sets the target value if it's within valid bounds (NoneIndex ≤ value < MaxIndex).
// Returns an error if the value is out of bounds.
func (apnTypeMap *APNTypeCoreMap[Type]) SetValue(apnTypeValue *Type, apnTypeIndex Type) error {
	if !(apnTypeMap.NoneIndex <= apnTypeIndex && apnTypeIndex < apnTypeMap.MaxIndex) {
		return fmt.Errorf("apn type has incorrect value: %d", apnTypeIndex)
	}

	*apnTypeValue = apnTypeIndex

	return nil
}

//--------------------------------------------------------------------------------//
// APNType CoreProxyOption
//--------------------------------------------------------------------------------//

// APNTypeCoreProxyOption configures how APN types are serialized to/from JSON and XML.
// This structure controls the format of the serialized representation.
type APNTypeCoreProxyOption struct {
	jsonIsArray bool

	xmlIsArray           bool
	xmlArrayHasSeparator string
	xmlIsString          bool
	xmlStringIsUpper     bool
	xmlIsNumber          bool
	xmlNumberIsOrder     bool
	xmlNumberIsIndex     bool
}

// NewAPNTypeCoreProxyOption creates a new APNTypeCoreProxyOption with default values:
//   - JSON: single value (not array)
//   - XML: string representation (not array, not number), lowercase
func NewAPNTypeCoreProxyOption() APNTypeCoreProxyOption {
	return APNTypeCoreProxyOption{
		jsonIsArray: false,

		xmlIsArray:           false,
		xmlArrayHasSeparator: "",
		xmlIsString:          true,
		xmlStringIsUpper:     false,
		xmlIsNumber:          false,
		xmlNumberIsOrder:     false,
		xmlNumberIsIndex:     false,
	}
}

// SetJSONIsArray configures whether JSON serialization should use an array format.
// When true, multiple values are serialized as a JSON array; when false, as a single string.
// Returns a new APNTypeCoreProxyOption with the updated setting.
func (apnTypeProxyOption APNTypeCoreProxyOption) SetJSONIsArray(jsonIsArray bool) APNTypeCoreProxyOption {
	apnTypeProxyOption.jsonIsArray = jsonIsArray

	return apnTypeProxyOption
}

// SetXMLIsArray configures XML serialization to use an array format with the specified separator.
// The separator is used to join multiple values into a single string attribute.
// Returns a new APNTypeCoreProxyOption with xmlIsArray=true and the specified separator.
func (apnTypeProxyOption APNTypeCoreProxyOption) SetXMLIsArray(xmlArrayHasSeparator string) APNTypeCoreProxyOption {
	apnTypeProxyOption.xmlIsArray = true
	apnTypeProxyOption.xmlArrayHasSeparator = xmlArrayHasSeparator

	return apnTypeProxyOption
}

// SetXMLIsString configures XML serialization to use string representation.
// When xmlStringIsUpper is true, strings are converted to uppercase.
// This method also disables number representation (xmlIsNumber=false).
// Returns a new APNTypeCoreProxyOption with the updated settings.
func (apnTypeProxyOption APNTypeCoreProxyOption) SetXMLIsString(xmlStringIsUpper bool) APNTypeCoreProxyOption {
	apnTypeProxyOption.xmlIsNumber = false

	apnTypeProxyOption.xmlIsString = true
	apnTypeProxyOption.xmlStringIsUpper = xmlStringIsUpper

	return apnTypeProxyOption
}

// SetXMLIsNumber configures XML serialization to use number representation.
// When xmlNumberIsOrder is true, numbers represent the order (1-based index) of values.
// When false, numbers represent the actual index value.
// This method also disables string representation (xmlIsString=false).
// Returns a new APNTypeCoreProxyOption with the updated settings.
func (apnTypeProxyOption APNTypeCoreProxyOption) SetXMLIsNumber(xmlNumberIsOrder bool) APNTypeCoreProxyOption {
	apnTypeProxyOption.xmlIsString = false

	apnTypeProxyOption.xmlIsNumber = true
	apnTypeProxyOption.xmlNumberIsOrder = xmlNumberIsOrder
	apnTypeProxyOption.xmlNumberIsIndex = !xmlNumberIsOrder

	return apnTypeProxyOption
}

//--------------------------------------------------------------------------------//
// APNType CoreProxy
//--------------------------------------------------------------------------------//

// APNTypeCoreProxy provides serialization capabilities for APN types to JSON and XML formats.
// It wraps an APNTypeCoreMap for JSON and creates a separate APNTypeCoreMap for XML with
// potentially different string representations based on the configured options.
//
// The proxy handles the conversion between the internal bitmask representation and the
// external serialized formats according to the specified options.
type APNTypeCoreProxy[Type ~int] struct {
	JSONMap *APNTypeCoreMap[Type]
	XMLMap  *APNTypeCoreMap[Type]

	option APNTypeCoreProxyOption
}

// NewAPNTypeCoreProxy creates a new APNTypeCoreProxy with the specified parameters.
// It initializes the JSONMap with the provided mapping, then creates an XMLMap with
// potentially transformed string values based on the proxy options:
//   - If xmlIsString is true, strings may be converted to uppercase
//   - If xmlIsNumber is true, strings are replaced with numeric values (order or index)
//
// Returns a pointer to the initialized APNTypeCoreProxy.
func NewAPNTypeCoreProxy[Type ~int](noneIndex Type, maxIndex Type, mapByIndex map[Type]string, option APNTypeCoreProxyOption) *APNTypeCoreProxy[Type] {
	apnTypeProxy := &APNTypeCoreProxy[Type]{
		JSONMap: NewAPNTypeCoreMap(noneIndex, maxIndex, mapByIndex),
		XMLMap:  nil,
		option:  option,
	}

	xmlMapByIndex := map[Type]string{}

	if option.xmlIsString {
		for apnTypeIndex, apnTypeString := range apnTypeProxy.JSONMap.MapByIndex {
			if option.xmlStringIsUpper {
				apnTypeString = strings.ToUpper(apnTypeString)
			}

			xmlMapByIndex[apnTypeIndex] = apnTypeString
		}
	}

	if option.xmlIsNumber {
		for apnTypeOrder, apnTypeIndex := range apnTypeProxy.JSONMap.IndexArray {
			apnTypeString := strconv.Itoa(apnTypeOrder + 1)
			xmlMapByIndex[apnTypeIndex] = apnTypeString
		}
	}

	apnTypeProxy.XMLMap = NewAPNTypeCoreMap(
		apnTypeProxy.JSONMap.NoneIndex,
		apnTypeProxy.JSONMap.MaxIndex,
		xmlMapByIndex,
	)

	return apnTypeProxy
}

// MarshalJSONValue serializes the APN type value to JSON according to the configured options.
// If jsonIsArray is true, returns a JSON array of strings; otherwise returns a single string.
// Returns the JSON bytes and any error from the json.Marshal call.
func (apnTypeProxy *APNTypeCoreProxy[Type]) MarshalJSONValue(apnTypeValue Type) (jsonByte []byte, err error) {
	if apnTypeProxy.option.jsonIsArray {
		return json.Marshal(apnTypeProxy.JSONMap.GetStringArray(apnTypeValue))
	} else {
		return json.Marshal(apnTypeProxy.JSONMap.GetString(apnTypeValue))
	}
}

// UnmarshalJSONValue deserializes JSON data into the APN type value according to configured options.
// If jsonIsArray is true, expects a JSON array of strings; otherwise expects a single string.
// Returns an error if unmarshaling fails or if any string/index is invalid.
func (apnTypeProxy *APNTypeCoreProxy[Type]) UnmarshalJSONValue(apnTypeValue *Type, jsonByte []byte) error {
	var apnTypeJSON interface{}

	if apnTypeProxy.option.jsonIsArray {
		apnTypeJSON = []string{}
	} else {
		apnTypeJSON = ""
	}

	err := json.Unmarshal(jsonByte, &apnTypeJSON)
	if err != nil {
		return err
	}

	if apnTypeProxy.option.jsonIsArray {
		return apnTypeProxy.JSONMap.SetStringArray(apnTypeValue, apnTypeJSON.([]string))
	} else {
		return apnTypeProxy.JSONMap.SetString(apnTypeValue, apnTypeJSON.(string))
	}
}

// MarshalXMLValue serializes the APN type value to an XML attribute according to configured options.
// Handles different formats:
//   - Array: joins strings with specified separator
//   - String: uses string representation (with case conversion if specified)
//   - Number: uses either order-based or index-based numeric representation
//
// Returns an xml.Attr with the specified name and serialized value.
func (apnTypeProxy *APNTypeCoreProxy[Type]) MarshalXMLValue(apnTypeValue Type, xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	var (
		apnTypeString string
	)

	if apnTypeProxy.option.xmlIsArray {
		apnTypeString = strings.Join(apnTypeProxy.XMLMap.GetStringArray(apnTypeValue), apnTypeProxy.option.xmlArrayHasSeparator)
	} else {
		if apnTypeProxy.option.xmlIsString {
			apnTypeString = apnTypeProxy.XMLMap.GetString(apnTypeValue)
		}

		if apnTypeProxy.option.xmlIsNumber {
			if apnTypeProxy.option.xmlNumberIsOrder {
				apnTypeString = apnTypeProxy.XMLMap.GetString(apnTypeValue)
			}

			if apnTypeProxy.option.xmlNumberIsIndex {
				apnTypeString = strconv.Itoa(int(apnTypeProxy.XMLMap.GetValue(apnTypeValue)))
			}
		}
	}

	return xml.Attr{
		Name:  xmlAttrName,
		Value: apnTypeString,
	}, err
}

// UnmarshalXMLValue deserializes an XML attribute into the APN type value according to configured options.
// Handles different formats:
//   - Array: splits by separator and processes each string
//   - String: processes the string directly
//   - Number: converts to integer and validates range (for index-based numbers)
//
// Returns an error if parsing fails or if any value is invalid.
func (apnTypeProxy *APNTypeCoreProxy[Type]) UnmarshalXMLValue(apnTypeValue *Type, xmlAttr xml.Attr) error {
	if apnTypeProxy.option.xmlIsArray {
		apnTypeStringArray := strings.Split(xmlAttr.Value, apnTypeProxy.option.xmlArrayHasSeparator)
		return apnTypeProxy.XMLMap.SetStringArray(apnTypeValue, apnTypeStringArray)
	} else {
		if apnTypeProxy.option.xmlIsString {
			return apnTypeProxy.XMLMap.SetString(apnTypeValue, xmlAttr.Value)
		}

		if apnTypeProxy.option.xmlIsNumber {
			if apnTypeProxy.option.xmlNumberIsOrder {
				return apnTypeProxy.XMLMap.SetString(apnTypeValue, xmlAttr.Value)
			}

			if apnTypeProxy.option.xmlNumberIsIndex {
				apnTypeIndex, err := strconv.Atoi(xmlAttr.Value)
				if err != nil {
					return fmt.Errorf("apn type has invalid number: %v", err)
				}

				if apnTypeIndex < int(apnTypeProxy.XMLMap.NoneIndex) || int(apnTypeProxy.XMLMap.MaxIndex) <= apnTypeIndex {
					return fmt.Errorf("apn type has out of range number: %d", apnTypeIndex)
				}

				*apnTypeValue = Type(apnTypeIndex)
			}
		}
	}

	return nil
}

//--------------------------------------------------------------------------------//
// APNType BaseType
//--------------------------------------------------------------------------------//

// APNTypeBaseType represents the base capabilities of an APN.
// This is a bitmask type where multiple capabilities can be combined.
// Common values include DEFAULT, MMS, SUPL, DUN, HIPRI, etc.
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

// apnTypeBaseTypeStorage is the shared proxy instance for APNTypeBaseType serialization.
// Configured for:
//   - JSON: array format
//   - XML: array format with comma separator
//   - XML: number representation (not string)
var apnTypeBaseTypeStorage = NewAPNTypeCoreProxy(
	APNTYPE_BASE_TYPE_NONE,
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
	NewAPNTypeCoreProxyOption().SetJSONIsArray(true).SetXMLIsArray(",").SetXMLIsString(false),
)

// String returns a pipe-separated string of all set capabilities.
// For example: "default|mms|supl"
func (apnBaseType APNTypeBaseType) String() string {
	return strings.Join(apnTypeBaseTypeStorage.JSONMap.GetStringArray(apnBaseType), "|")
}

// MarshalJSON serializes the APNTypeBaseType to JSON according to the configured options.
// Returns the JSON bytes and any error from serialization.
func (apnBaseType APNTypeBaseType) MarshalJSON() ([]byte, error) {
	return apnTypeBaseTypeStorage.MarshalJSONValue(apnBaseType)
}

// UnmarshalJSON deserializes JSON data into the APNTypeBaseType according to configured options.
// Returns an error if unmarshaling fails or if any value is invalid.
func (apnBaseType *APNTypeBaseType) UnmarshalJSON(jsonData []byte) error {
	return apnTypeBaseTypeStorage.UnmarshalJSONValue(apnBaseType, jsonData)
}

// MarshalXMLAttr serializes the APNTypeBaseType to an XML attribute according to configured options.
// Returns an xml.Attr with the specified name and serialized value.
func (apnBaseType APNTypeBaseType) MarshalXMLAttr(xmlAttrName xml.Name) (attr xml.Attr, err error) {
	return apnTypeBaseTypeStorage.MarshalXMLValue(apnBaseType, xmlAttrName)
}

// UnmarshalXMLAttr deserializes an XML attribute into the APNTypeBaseType according to configured options.
// Returns an error if parsing fails or if any value is invalid.
func (apnBaseType *APNTypeBaseType) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return apnTypeBaseTypeStorage.UnmarshalXMLValue(apnBaseType, xmlAttr)
}

//--------------------------------------------------------------------------------//
// APNType AuthType
//--------------------------------------------------------------------------------//

// APNTypeAuthType represents the authentication type for an APN.
// This is a bitmask type, though typically only one authentication method is used.
// Values include NONE, PAP, and CHAP.
type APNTypeAuthType int

const (
	APNTYPE_AUTH_TYPE_NONE APNTypeAuthType = 0
	APNTYPE_AUTH_TYPE_PAP  APNTypeAuthType = 1 << (iota - 1)
	APNTYPE_AUTH_TYPE_CHAP
	APNTYPE_AUTH_TYPE_MAX

	APNTYPE_AUTH_TYPE_ALL = APNTYPE_AUTH_TYPE_MAX - 1
)

// apnTypeAuthTypeStorage is the shared proxy instance for APNTypeAuthType serialization.
// Configured for:
//   - JSON: array format
//   - XML: number representation (not array, not string)
var apnTypeAuthTypeStorage = NewAPNTypeCoreProxy(
	APNTYPE_AUTH_TYPE_NONE,
	APNTYPE_AUTH_TYPE_MAX,
	map[APNTypeAuthType]string{
		APNTYPE_AUTH_TYPE_NONE: "none",
		APNTYPE_AUTH_TYPE_PAP:  "pap",
		APNTYPE_AUTH_TYPE_CHAP: "chap",
	},
	NewAPNTypeCoreProxyOption().SetJSONIsArray(true).SetXMLIsNumber(false),
)

// String returns a pipe-separated string of all set authentication types.
// For example: "pap|chap" (though typically only one is set)
func (apnAuthType APNTypeAuthType) String() string {
	return strings.Join(apnTypeAuthTypeStorage.JSONMap.GetStringArray(apnAuthType), "|")
}

// MarshalJSON serializes the APNTypeAuthType to JSON according to the configured options.
// Returns the JSON bytes and any error from serialization.
func (apnAuthType APNTypeAuthType) MarshalJSON() ([]byte, error) {
	return apnTypeAuthTypeStorage.MarshalJSONValue(apnAuthType)
}

// UnmarshalJSON deserializes JSON data into the APNTypeAuthType according to configured options.
// Returns an error if unmarshaling fails or if any value is invalid.
func (apnAuthType *APNTypeAuthType) UnmarshalJSON(jsonData []byte) error {
	return apnTypeAuthTypeStorage.UnmarshalJSONValue(apnAuthType, jsonData)
}

// MarshalXMLAttr serializes the APNTypeAuthType to an XML attribute according to configured options.
// Returns an xml.Attr with the specified name and serialized value.
func (apnAuthType APNTypeAuthType) MarshalXMLAttr(xmlAttrName xml.Name) (attr xml.Attr, err error) {
	return apnTypeAuthTypeStorage.MarshalXMLValue(apnAuthType, xmlAttrName)
}

// UnmarshalXMLAttr deserializes an XML attribute into the APNTypeAuthType according to configured options.
// Returns an error if parsing fails or if any value is invalid.
func (apnAuthType *APNTypeAuthType) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return apnTypeAuthTypeStorage.UnmarshalXMLValue(apnAuthType, xmlAttr)
}

//--------------------------------------------------------------------------------//
// APNType NetworkType
//--------------------------------------------------------------------------------//

// APNTypeNetworkType represents the network technologies supported by an APN.
// This is a bitmask type where multiple network types can be combined.
// Values include GPRS, EDGE, UMTS, LTE, NR, etc.
type APNTypeNetworkType int

const (
	APNTYPE_NETWORK_TYPE_NONE APNTypeNetworkType = 0
	APNTYPE_NETWORK_TYPE_GPRS APNTypeNetworkType = 1 << (iota - 1)
	APNTYPE_NETWORK_TYPE_EDGE
	APNTYPE_NETWORK_TYPE_UMTS
	APNTYPE_NETWORK_TYPE_CDMA
	APNTYPE_NETWORK_TYPE_EVDO_0
	APNTYPE_NETWORK_TYPE_EVDO_A
	APNTYPE_NETWORK_TYPE_1xRTT
	APNTYPE_NETWORK_TYPE_HSDPA
	APNTYPE_NETWORK_TYPE_HSUPA
	APNTYPE_NETWORK_TYPE_HSPA
	APNTYPE_NETWORK_TYPE_IDEN
	APNTYPE_NETWORK_TYPE_EVDO_B
	APNTYPE_NETWORK_TYPE_LTE
	APNTYPE_NETWORK_TYPE_EHRPD
	APNTYPE_NETWORK_TYPE_HSPAP
	APNTYPE_NETWORK_TYPE_GSM
	APNTYPE_NETWORK_TYPE_TD_SCDMA
	APNTYPE_NETWORK_TYPE_IWLAN
	APNTYPE_NETWORK_TYPE_LTE_CA
	APNTYPE_NETWORK_TYPE_NR
	APNTYPE_NETWORK_TYPE_MAX

	APNTYPE_NETWORK_TYPE_ALL = APNTYPE_NETWORK_TYPE_MAX - 1
)

// apnTypeNetworkTypeStorage is the shared proxy instance for APNTypeNetworkType serialization.
// Configured for:
//   - JSON: array format
//   - XML: array format with pipe separator
//   - XML: number representation
var apnTypeNetworkTypeStorage = NewAPNTypeCoreProxy(
	APNTYPE_NETWORK_TYPE_NONE,
	APNTYPE_NETWORK_TYPE_MAX,
	map[APNTypeNetworkType]string{
		APNTYPE_NETWORK_TYPE_NONE:     "unknown",
		APNTYPE_NETWORK_TYPE_GPRS:     "gprs",
		APNTYPE_NETWORK_TYPE_EDGE:     "edge",
		APNTYPE_NETWORK_TYPE_UMTS:     "umts",
		APNTYPE_NETWORK_TYPE_CDMA:     "cdma",
		APNTYPE_NETWORK_TYPE_EVDO_0:   "evdo_0",
		APNTYPE_NETWORK_TYPE_EVDO_A:   "evdo_a",
		APNTYPE_NETWORK_TYPE_1xRTT:    "1xrtt",
		APNTYPE_NETWORK_TYPE_HSDPA:    "hsdpa",
		APNTYPE_NETWORK_TYPE_HSUPA:    "hsupa",
		APNTYPE_NETWORK_TYPE_HSPA:     "hspa",
		APNTYPE_NETWORK_TYPE_IDEN:     "iden",
		APNTYPE_NETWORK_TYPE_EVDO_B:   "evdo_b",
		APNTYPE_NETWORK_TYPE_LTE:      "lte",
		APNTYPE_NETWORK_TYPE_EHRPD:    "ehrpd",
		APNTYPE_NETWORK_TYPE_HSPAP:    "hspap",
		APNTYPE_NETWORK_TYPE_GSM:      "gsm",
		APNTYPE_NETWORK_TYPE_TD_SCDMA: "td_scdma",
		APNTYPE_NETWORK_TYPE_IWLAN:    "iwlan",
		APNTYPE_NETWORK_TYPE_LTE_CA:   "lte_ca",
		APNTYPE_NETWORK_TYPE_NR:       "nr",
	},
	NewAPNTypeCoreProxyOption().SetJSONIsArray(true).SetXMLIsArray("|").SetXMLIsNumber(true),
)

// String returns a pipe-separated string of all set network types.
// For example: "lte|nr"
func (apnNetworkType APNTypeNetworkType) String() string {
	return strings.Join(apnTypeNetworkTypeStorage.JSONMap.GetStringArray(apnNetworkType), "|")
}

// MarshalJSON serializes the APNTypeNetworkType to JSON according to the configured options.
// Returns the JSON bytes and any error from serialization.
func (apnNetworkType APNTypeNetworkType) MarshalJSON() ([]byte, error) {
	return apnTypeNetworkTypeStorage.MarshalJSONValue(apnNetworkType)
}

// UnmarshalJSON deserializes JSON data into the APNTypeNetworkType according to configured options.
// Returns an error if unmarshaling fails or if any value is invalid.
func (apnNetworkType *APNTypeNetworkType) UnmarshalJSON(jsonData []byte) error {
	return apnTypeNetworkTypeStorage.UnmarshalJSONValue(apnNetworkType, jsonData)
}

// MarshalXMLAttr serializes the APNTypeNetworkType to an XML attribute according to configured options.
// Returns an xml.Attr with the specified name and serialized value.
func (apnNetworkType APNTypeNetworkType) MarshalXMLAttr(xmlAttrName xml.Name) (attr xml.Attr, err error) {
	return apnTypeNetworkTypeStorage.MarshalXMLValue(apnNetworkType, xmlAttrName)
}

// UnmarshalXMLAttr deserializes an XML attribute into the APNTypeNetworkType according to configured options.
// Returns an error if parsing fails or if any value is invalid.
func (apnNetworkType *APNTypeNetworkType) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return apnTypeNetworkTypeStorage.UnmarshalXMLValue(apnNetworkType, xmlAttr)
}

//--------------------------------------------------------------------------------//
// APNType BearerProtocol
//--------------------------------------------------------------------------------//

// APNTypeBearerProtocol represents the bearer protocol for an APN.
// This is typically a single value (not a bitmask), though defined as int for consistency.
// Values include IP, IPV4, IPV6, IPV4V6, PPP, NONIP, UNSTRUCTURED.
type APNTypeBearerProtocol int

const (
	APNTYPE_BEARER_PROTOCOL_NONE APNTypeBearerProtocol = iota
	APNTYPE_BEARER_PROTOCOL_IP   APNTypeBearerProtocol = 1 << (iota - 1)
	APNTYPE_BEARER_PROTOCOL_IPV4
	APNTYPE_BEARER_PROTOCOL_IPV6
	APNTYPE_BEARER_PROTOCOL_IPV4V6
	APNTYPE_BEARER_PROTOCOL_PPP
	APNTYPE_BEARER_PROTOCOL_NONIP
	APNTYPE_BEARER_PROTOCOL_UNSTRUCTURED
	APNTYPE_BEARER_PROTOCOL_MAX
)

// apnTypeBearerProtocolStorage is the shared proxy instance for APNTypeBearerProtocol serialization.
// Configured for:
//   - JSON: single value (not array)
//   - XML: string representation
var apnTypeBearerProtocolStorage = NewAPNTypeCoreProxy(
	APNTYPE_BEARER_PROTOCOL_NONE,
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
	NewAPNTypeCoreProxyOption().SetJSONIsArray(false).SetXMLIsString(true),
)

// String returns the string representation of the bearer protocol.
// Since this is typically a single value, returns just that string.
func (apnBearerProtocol APNTypeBearerProtocol) String() string {
	return apnTypeBearerProtocolStorage.JSONMap.GetString(apnBearerProtocol)
}

// MarshalJSON serializes the APNTypeBearerProtocol to JSON according to the configured options.
// Returns the JSON bytes and any error from serialization.
func (apnBearerProtocol APNTypeBearerProtocol) MarshalJSON() ([]byte, error) {
	return apnTypeBearerProtocolStorage.MarshalJSONValue(apnBearerProtocol)
}

// UnmarshalJSON deserializes JSON data into the APNTypeBearerProtocol according to configured options.
// Returns an error if unmarshaling fails or if the value is invalid.
func (apnBearerProtocol *APNTypeBearerProtocol) UnmarshalJSON(jsonData []byte) error {
	return apnTypeBearerProtocolStorage.UnmarshalJSONValue(apnBearerProtocol, jsonData)
}

// MarshalXMLAttr serializes the APNTypeBearerProtocol to an XML attribute according to configured options.
// Returns an xml.Attr with the specified name and serialized value.
func (apnBearerProtocol APNTypeBearerProtocol) MarshalXMLAttr(xmlAttrName xml.Name) (attr xml.Attr, err error) {
	return apnTypeBearerProtocolStorage.MarshalXMLValue(apnBearerProtocol, xmlAttrName)
}

// UnmarshalXMLAttr deserializes an XML attribute into the APNTypeBearerProtocol according to configured options.
// Returns an error if parsing fails or if the value is invalid.
func (apnBearerProtocol *APNTypeBearerProtocol) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return apnTypeBearerProtocolStorage.UnmarshalXMLValue(apnBearerProtocol, xmlAttr)
}

//--------------------------------------------------------------------------------//
