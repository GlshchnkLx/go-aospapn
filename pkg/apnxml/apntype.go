// # APN Type
//
// File apntype provides bidirectional bitmask-to-string mapping and configurable
// serialization for APN types (Base, Auth, Network, BearerProtocol) to JSON/XML.
//
// Predefined APN types:
//   - BaseType: APN capabilities (default, mms, supl, dun, etc.)
//   - AuthType: Authentication methods (none, pap, chap)
//   - NetworkType: Technologies (gprs, lte, nr, etc.)
//   - BearerProtocol: Protocols (ip, ipv4, ipv6, ppp, etc.)
//
// Each type supports:
//   - String() for human-readable output
//   - JSON/XML marshaling via proxy configuration
//
// Serialization options include:
//   - Arrays vs single values
//   - Case (upper/lower) for strings
//   - Numeric formats (order/index-based)
//   - Custom array separators
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
// APNTypeCoreMap
//--------------------------------------------------------------------------------//

// APNTypeCoreMap maps bitmask indices to strings bidirectionally.
// Enforces bounds via NoneIndex and MaxIndex. Supports bitmask operations.
// Type must be an integer (~int).
type APNTypeCoreMap[Type ~int] struct {
	NoneIndex   Type
	MaxIndex    Type
	IndexArray  []Type
	MapByIndex  map[Type]string
	MapByString map[string]Type
}

// NewAPNTypeCoreMap initializes a core map from index→string mappings.
// Trims string values, validates index bounds, sorts IndexArray.
// Returns pointer to new APNTypeCoreMap.
func NewAPNTypeCoreMap[Type ~int](noneIndex Type, maxIndex Type, mapByIndex map[Type]string) *APNTypeCoreMap[Type] {
	coreMapStorage := &APNTypeCoreMap[Type]{
		NoneIndex:   noneIndex,
		MaxIndex:    maxIndex,
		IndexArray:  []Type{},
		MapByIndex:  map[Type]string{},
		MapByString: map[string]Type{},
	}

	for apnTypeIndex, apnTypeString := range mapByIndex {
		apnTypeString = strings.TrimSpace(apnTypeString)

		if noneIndex < apnTypeIndex && apnTypeIndex < maxIndex {
			coreMapStorage.IndexArray = append(coreMapStorage.IndexArray, apnTypeIndex)
		}

		coreMapStorage.MapByIndex[apnTypeIndex] = apnTypeString
		coreMapStorage.MapByString[apnTypeString] = apnTypeIndex
	}

	sort.Slice(coreMapStorage.IndexArray, func(i, j int) bool {
		return coreMapStorage.IndexArray[i] < coreMapStorage.IndexArray[j]
	})

	return coreMapStorage
}

// GetIndex returns index if valid, else NoneIndex.
func (coreMapStorage *APNTypeCoreMap[Type]) GetIndex(apnTypeValue Type) Type {
	if _, ok := coreMapStorage.MapByIndex[apnTypeValue]; ok {
		return apnTypeValue
	}

	return coreMapStorage.NoneIndex
}

// SetIndex assigns target to index if valid. Returns error if invalid.
func (coreMapStorage *APNTypeCoreMap[Type]) SetIndex(apnTypeValue *Type, apnTypeIndex Type) error {
	if _, ok := coreMapStorage.MapByIndex[apnTypeIndex]; !ok {
		return fmt.Errorf("apn type has incorrect index: %d", apnTypeIndex)
	}

	*apnTypeValue = apnTypeIndex

	return nil
}

// GetIndexArray returns indices with bits set in value. Returns [NoneIndex] if none set.
func (coreMapStorage *APNTypeCoreMap[Type]) GetIndexArray(apnTypeValue Type) []Type {
	apnTypeIndexArray := []Type{}
	for _, apnTypeIndex := range coreMapStorage.IndexArray {
		if apnTypeValue&apnTypeIndex == apnTypeIndex {
			apnTypeIndexArray = append(apnTypeIndexArray, apnTypeIndex)
		}
	}

	if len(apnTypeIndexArray) == 0 {
		apnTypeIndexArray = append(apnTypeIndexArray, coreMapStorage.NoneIndex)
	}

	return apnTypeIndexArray
}

// SetIndexArray sets target by OR-ing all indices. Returns error if any index invalid.
func (coreMapStorage *APNTypeCoreMap[Type]) SetIndexArray(apnTypeValue *Type, apnTypeIndexArray []Type) error {
	*apnTypeValue = coreMapStorage.NoneIndex
	for _, apnTypeIndex := range apnTypeIndexArray {
		if _, ok := coreMapStorage.MapByIndex[apnTypeIndex]; !ok {
			return fmt.Errorf("apn type has incorrect index: %d", apnTypeIndex)
		}

		*apnTypeValue |= apnTypeIndex
	}

	return nil
}

// GetString returns string for index. Returns NoneIndex string if invalid.
func (coreMapStorage *APNTypeCoreMap[Type]) GetString(apnTypeValue Type) string {
	return coreMapStorage.MapByIndex[coreMapStorage.GetIndex(apnTypeValue)]
}

// SetString assigns target to index of string. Returns error if string not found.
func (coreMapStorage *APNTypeCoreMap[Type]) SetString(apnTypeValue *Type, apnTypeString string) error {
	apnTypeIndex, ok := coreMapStorage.MapByString[apnTypeString]
	if !ok {
		return fmt.Errorf("apn type has incorrect string: %q", apnTypeString)
	}

	return coreMapStorage.SetIndex(apnTypeValue, apnTypeIndex)
}

// GetStringArray returns strings for all set bits in value.
func (coreMapStorage *APNTypeCoreMap[Type]) GetStringArray(apnTypeValue Type) []string {
	var apnTypeStringArray []string
	for _, apnTypeIndex := range coreMapStorage.GetIndexArray(apnTypeValue) {
		apnTypeStringArray = append(apnTypeStringArray, coreMapStorage.MapByIndex[apnTypeIndex])
	}

	return apnTypeStringArray
}

// SetStringArray assigns target from lowercase-trimmed strings. Returns error if any string invalid.
func (coreMapStorage *APNTypeCoreMap[Type]) SetStringArray(apnTypeValue *Type, apnTypeStringArray []string) error {
	var apnTypeIndexArray []Type
	for _, apnTypeString := range apnTypeStringArray {
		apnTypeString = strings.ToLower(strings.TrimSpace(apnTypeString))
		apnTypeIndex, ok := coreMapStorage.MapByString[apnTypeString]
		if !ok {
			return fmt.Errorf("apn type has incorrect string: %q", apnTypeString)
		}

		apnTypeIndexArray = append(apnTypeIndexArray, apnTypeIndex)
	}

	return coreMapStorage.SetIndexArray(apnTypeValue, apnTypeIndexArray)
}

// GetValue returns value if in bounds [NoneIndex, MaxIndex), else NoneIndex.
func (coreMapStorage *APNTypeCoreMap[Type]) GetValue(apnTypeValue Type) Type {
	if apnTypeValue <= coreMapStorage.NoneIndex || coreMapStorage.MaxIndex <= apnTypeValue {
		return coreMapStorage.NoneIndex
	}

	return apnTypeValue
}

// SetValue assigns target if in bounds [NoneIndex, MaxIndex). Returns error if out of bounds.
func (coreMapStorage *APNTypeCoreMap[Type]) SetValue(apnTypeValue *Type, apnTypeIndex Type) error {
	if !(coreMapStorage.NoneIndex <= apnTypeIndex && apnTypeIndex < coreMapStorage.MaxIndex) {
		return fmt.Errorf("apn type has incorrect value: %d", apnTypeIndex)
	}

	*apnTypeValue = apnTypeIndex

	return nil
}

//--------------------------------------------------------------------------------//
// APNTypeCoreProxyOption
//--------------------------------------------------------------------------------//

// APNTypeCoreProxyOption configures serialization format for JSON/XML.
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

// NewAPNTypeCoreProxyOption returns default options:
// JSON: single value; XML: lowercase string.
func NewAPNTypeCoreProxyOption() APNTypeCoreProxyOption {
	return APNTypeCoreProxyOption{
		jsonIsArray: false,
		xmlIsArray:  false,
		xmlIsString: true,
	}
}

// SetJSONIsArray configures JSON output as array (true) or single string (false).
func (coreProxyOption APNTypeCoreProxyOption) SetJSONIsArray(jsonIsArray bool) APNTypeCoreProxyOption {
	coreProxyOption.jsonIsArray = jsonIsArray
	return coreProxyOption
}

// SetXMLIsArray configures XML output as joined string using separator.
func (coreProxyOption APNTypeCoreProxyOption) SetXMLIsArray(xmlArrayHasSeparator string) APNTypeCoreProxyOption {
	coreProxyOption.xmlIsArray = true
	coreProxyOption.xmlArrayHasSeparator = xmlArrayHasSeparator
	return coreProxyOption
}

// SetXMLIsString configures XML output as string. Uppercase if xmlStringIsUpper=true.
func (coreProxyOption APNTypeCoreProxyOption) SetXMLIsString(xmlStringIsUpper bool) APNTypeCoreProxyOption {
	coreProxyOption.xmlIsString = true
	coreProxyOption.xmlStringIsUpper = xmlStringIsUpper
	coreProxyOption.xmlIsNumber = false
	return coreProxyOption
}

// SetXMLIsNumber configures XML output as number: order (1-based) or raw index.
func (coreProxyOption APNTypeCoreProxyOption) SetXMLIsNumber(xmlNumberIsOrder bool) APNTypeCoreProxyOption {
	coreProxyOption.xmlIsNumber = true
	coreProxyOption.xmlNumberIsOrder = xmlNumberIsOrder
	coreProxyOption.xmlNumberIsIndex = !xmlNumberIsOrder
	coreProxyOption.xmlIsString = false
	return coreProxyOption
}

//--------------------------------------------------------------------------------//
// APNTypeCoreProxy
//--------------------------------------------------------------------------------//

// APNTypeCoreProxy serializes bitmask values to JSON/XML using configurable formats.
// Wraps separate core maps for JSON and XML with format-specific transformations.
type APNTypeCoreProxy[Type ~int] struct {
	JSONMap *APNTypeCoreMap[Type]
	XMLMap  *APNTypeCoreMap[Type]
	option  APNTypeCoreProxyOption
}

// NewAPNTypeCoreProxy creates proxy with JSON map and XML map (transformed per options).
// Transforms: uppercase strings or numeric strings (order/index-based).
func NewAPNTypeCoreProxy[Type ~int](noneIndex Type, maxIndex Type, mapByIndex map[Type]string, option APNTypeCoreProxyOption) *APNTypeCoreProxy[Type] {
	coreProxyStorage := &APNTypeCoreProxy[Type]{
		JSONMap: NewAPNTypeCoreMap(noneIndex, maxIndex, mapByIndex),
		option:  option,
	}

	xmlMapByIndex := map[Type]string{}

	if option.xmlIsString {
		for apnTypeIndex, apnTypeString := range coreProxyStorage.JSONMap.MapByIndex {
			if option.xmlStringIsUpper {
				apnTypeString = strings.ToUpper(apnTypeString)
			}

			xmlMapByIndex[apnTypeIndex] = apnTypeString
		}
	}

	if option.xmlIsNumber {
		for apnTypeOrder, apnTypeIndex := range coreProxyStorage.JSONMap.IndexArray {
			apnTypeString := strconv.Itoa(apnTypeOrder + 1)
			xmlMapByIndex[apnTypeIndex] = apnTypeString
		}
	}

	coreProxyStorage.XMLMap = NewAPNTypeCoreMap(
		coreProxyStorage.JSONMap.NoneIndex,
		coreProxyStorage.JSONMap.MaxIndex,
		xmlMapByIndex,
	)

	return coreProxyStorage
}

// MarshalTextValue serializes value to text: array (comma-separated) or single string.
func (coreProxyStorage *APNTypeCoreProxy[Type]) MarshalTextValue(apnTypeValue Type) (textByte []byte, err error) {
	if coreProxyStorage.option.jsonIsArray {
		return []byte(strings.Join(coreProxyStorage.JSONMap.GetStringArray(apnTypeValue), ",")), nil
	}

	return []byte(coreProxyStorage.JSONMap.GetString(apnTypeValue)), nil
}

// UnmarshalTextValue deserializes text: splits by comma if array, else single string.
func (coreProxyStorage *APNTypeCoreProxy[Type]) UnmarshalTextValue(apnTypeValue *Type, textByte []byte) error {
	if coreProxyStorage.option.jsonIsArray {
		return coreProxyStorage.JSONMap.SetStringArray(apnTypeValue, strings.Split(string(textByte), ","))
	}

	return coreProxyStorage.JSONMap.SetString(apnTypeValue, string(textByte))
}

// MarshalJSONValue serializes value to JSON: array of strings or single string.
func (coreProxyStorage *APNTypeCoreProxy[Type]) MarshalJSONValue(apnTypeValue Type) (jsonByte []byte, err error) {
	if coreProxyStorage.option.jsonIsArray {
		return json.Marshal(coreProxyStorage.JSONMap.GetStringArray(apnTypeValue))
	}

	return json.Marshal(coreProxyStorage.JSONMap.GetString(apnTypeValue))
}

// UnmarshalJSONValue deserializes JSON: array of strings or single string.
func (coreProxyStorage *APNTypeCoreProxy[Type]) UnmarshalJSONValue(apnTypeValue *Type, jsonByte []byte) error {
	if coreProxyStorage.option.jsonIsArray {
		if jsonByte[0] == '[' {
			var apnTypeStringArray []string

			err := json.Unmarshal(jsonByte, &apnTypeStringArray)
			if err != nil {
				return err
			}

			return coreProxyStorage.JSONMap.SetStringArray(apnTypeValue, apnTypeStringArray)
		} else {
			return coreProxyStorage.UnmarshalTextValue(apnTypeValue, jsonByte[1:len(jsonByte)-1])
		}
	} else {
		var apnTypeString string

		err := json.Unmarshal(jsonByte, &apnTypeString)
		if err != nil {
			return err
		}

		return coreProxyStorage.JSONMap.SetString(apnTypeValue, apnTypeString)
	}
}

// MarshalXMLValue serializes value to XML attribute per options: array, string, or number.
func (coreProxyStorage *APNTypeCoreProxy[Type]) MarshalXMLValue(apnTypeValue Type, xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	var apnTypeString string

	if coreProxyStorage.option.xmlIsArray {
		apnTypeString = strings.Join(coreProxyStorage.XMLMap.GetStringArray(apnTypeValue), coreProxyStorage.option.xmlArrayHasSeparator)
	} else if coreProxyStorage.option.xmlIsString {
		apnTypeString = coreProxyStorage.XMLMap.GetString(apnTypeValue)
	} else if coreProxyStorage.option.xmlIsNumber {
		if coreProxyStorage.option.xmlNumberIsOrder {
			apnTypeString = coreProxyStorage.XMLMap.GetString(apnTypeValue)
		} else if coreProxyStorage.option.xmlNumberIsIndex {
			apnTypeString = strconv.Itoa(int(coreProxyStorage.XMLMap.GetValue(apnTypeValue)))
		}
	}

	return xml.Attr{
		Name:  xmlAttrName,
		Value: apnTypeString,
	}, nil
}

// UnmarshalXMLValue deserializes XML attribute per options: array, string, or number.
func (coreProxyStorage *APNTypeCoreProxy[Type]) UnmarshalXMLValue(apnTypeValue *Type, xmlAttr xml.Attr) error {
	if coreProxyStorage.option.xmlIsArray {
		apnTypeStringArray := strings.Split(xmlAttr.Value, coreProxyStorage.option.xmlArrayHasSeparator)
		return coreProxyStorage.XMLMap.SetStringArray(apnTypeValue, apnTypeStringArray)
	} else if coreProxyStorage.option.xmlIsString {
		return coreProxyStorage.XMLMap.SetString(apnTypeValue, xmlAttr.Value)
	} else if coreProxyStorage.option.xmlIsNumber {
		if coreProxyStorage.option.xmlNumberIsOrder {
			return coreProxyStorage.XMLMap.SetString(apnTypeValue, xmlAttr.Value)
		} else if coreProxyStorage.option.xmlNumberIsIndex {
			apnTypeIndex, err := strconv.Atoi(xmlAttr.Value)
			if err != nil {
				return fmt.Errorf("apn type has invalid number: %v", err)
			}

			if apnTypeIndex < int(coreProxyStorage.XMLMap.NoneIndex) || int(coreProxyStorage.XMLMap.MaxIndex) <= apnTypeIndex {
				return fmt.Errorf("apn type has out of range number: %d", apnTypeIndex)
			}

			*apnTypeValue = Type(apnTypeIndex)
		}
	}
	return nil
}

//--------------------------------------------------------------------------------//
// APNTypeBaseType
//--------------------------------------------------------------------------------//

// APNTypeBaseType represents APN capabilities as bitmask (e.g., default, mms, supl).
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
)

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

// String returns pipe-separated capability names (e.g., "default|mms").
func (baseTypeValue APNTypeBaseType) String() string {
	return strings.Join(apnTypeBaseTypeStorage.JSONMap.GetStringArray(baseTypeValue), "|")
}

// MarshalText serializes to comma-separated string (e.g., "default,mms").
func (baseTypeValue APNTypeBaseType) MarshalText() (textByte []byte, err error) {
	return apnTypeBaseTypeStorage.MarshalTextValue(baseTypeValue)
}

// UnmarshalText deserializes from comma-separated string (e.g., "default,mms").
func (baseTypeValue *APNTypeBaseType) UnmarshalText(textByte []byte) error {
	fmt.Println(1, textByte, string(textByte))
	return apnTypeBaseTypeStorage.UnmarshalTextValue(baseTypeValue, textByte)
}

// MarshalJSON serializes to JSON array (e.g., ["default", "mms"]).
func (baseTypeValue APNTypeBaseType) MarshalJSON() (jsonByte []byte, err error) {
	return apnTypeBaseTypeStorage.MarshalJSONValue(baseTypeValue)
}

// UnmarshalJSON deserializes from JSON array (e.g., ["default", "mms"]).
func (baseTypeValue *APNTypeBaseType) UnmarshalJSON(jsonByte []byte) error {
	return apnTypeBaseTypeStorage.UnmarshalJSONValue(baseTypeValue, jsonByte)
}

// MarshalXMLAttr serializes to XML attribute per proxy options.
func (baseTypeValue APNTypeBaseType) MarshalXMLAttr(xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	return apnTypeBaseTypeStorage.MarshalXMLValue(baseTypeValue, xmlAttrName)
}

// UnmarshalXMLAttr deserializes from XML attribute per proxy options.
func (baseTypeValue *APNTypeBaseType) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return apnTypeBaseTypeStorage.UnmarshalXMLValue(baseTypeValue, xmlAttr)
}

//--------------------------------------------------------------------------------//
// APNTypeAuthType
//--------------------------------------------------------------------------------//

// APNTypeAuthType represents authentication methods as bitmask (none, pap, chap).
type APNTypeAuthType int

const (
	APNTYPE_AUTH_TYPE_NONE APNTypeAuthType = 0
	APNTYPE_AUTH_TYPE_PAP  APNTypeAuthType = 1 << (iota - 1)
	APNTYPE_AUTH_TYPE_CHAP
	APNTYPE_AUTH_TYPE_MAX
)

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

// String returns pipe-separated auth names (e.g., "pap|chap").
func (authTypeValue APNTypeAuthType) String() string {
	return strings.Join(apnTypeAuthTypeStorage.JSONMap.GetStringArray(authTypeValue), "|")
}

// MarshalText serializes to comma-separated string (e.g., "pap,chap").
func (authTypeValue APNTypeAuthType) MarshalText() (textByte []byte, err error) {
	return apnTypeAuthTypeStorage.MarshalTextValue(authTypeValue)
}

// UnmarshalText deserializes from comma-separated string (e.g., "pap,chap").
func (authTypeValue *APNTypeAuthType) UnmarshalText(textByte []byte) error {
	return apnTypeAuthTypeStorage.UnmarshalTextValue(authTypeValue, textByte)
}

// MarshalJSON serializes to JSON array (e.g., ["pap", "chap"]).
func (authTypeValue APNTypeAuthType) MarshalJSON() (jsonByte []byte, err error) {
	return apnTypeAuthTypeStorage.MarshalJSONValue(authTypeValue)
}

// UnmarshalJSON deserializes from JSON array (e.g., ["pap", "chap"]).
func (authTypeValue *APNTypeAuthType) UnmarshalJSON(jsonByte []byte) error {
	return apnTypeAuthTypeStorage.UnmarshalJSONValue(authTypeValue, jsonByte)
}

// MarshalXMLAttr serializes to XML attribute per proxy options.
func (authTypeValue APNTypeAuthType) MarshalXMLAttr(xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	return apnTypeAuthTypeStorage.MarshalXMLValue(authTypeValue, xmlAttrName)
}

// UnmarshalXMLAttr deserializes from XML attribute per proxy options.
func (authTypeValue *APNTypeAuthType) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return apnTypeAuthTypeStorage.UnmarshalXMLValue(authTypeValue, xmlAttr)
}

//--------------------------------------------------------------------------------//
// APNTypeNetworkType
//--------------------------------------------------------------------------------//

// APNTypeNetworkType represents network technologies as bitmask (e.g., lte, nr, gprs).
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
)

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

// String returns pipe-separated network names (e.g., "lte|nr").
func (networkTypeValue APNTypeNetworkType) String() string {
	return strings.Join(apnTypeNetworkTypeStorage.JSONMap.GetStringArray(networkTypeValue), "|")
}

// MarshalText serializes to comma-separated string (e.g., "lte,nr").
func (networkTypeValue APNTypeNetworkType) MarshalText() (textByte []byte, err error) {
	return apnTypeNetworkTypeStorage.MarshalTextValue(networkTypeValue)
}

// UnmarshalText deserializes from comma-separated string (e.g., "lte,nr").
func (networkTypeValue *APNTypeNetworkType) UnmarshalText(textByte []byte) error {
	return apnTypeNetworkTypeStorage.UnmarshalTextValue(networkTypeValue, textByte)
}

// MarshalJSON serializes to JSON array (e.g., ["lte", "nr"]).
func (networkTypeValue APNTypeNetworkType) MarshalJSON() (jsonByte []byte, err error) {
	return apnTypeNetworkTypeStorage.MarshalJSONValue(networkTypeValue)
}

// UnmarshalJSON deserializes from JSON array (e.g., ["lte", "nr"]).
func (networkTypeValue *APNTypeNetworkType) UnmarshalJSON(jsonByte []byte) error {
	return apnTypeNetworkTypeStorage.UnmarshalJSONValue(networkTypeValue, jsonByte)
}

// MarshalXMLAttr serializes to XML attribute per proxy options.
func (networkTypeValue APNTypeNetworkType) MarshalXMLAttr(xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	return apnTypeNetworkTypeStorage.MarshalXMLValue(networkTypeValue, xmlAttrName)
}

// UnmarshalXMLAttr deserializes from XML attribute per proxy options.
func (networkTypeValue *APNTypeNetworkType) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return apnTypeNetworkTypeStorage.UnmarshalXMLValue(networkTypeValue, xmlAttr)
}

//--------------------------------------------------------------------------------//
// APNTypeBearerProtocol
//--------------------------------------------------------------------------------//

// APNTypeBearerProtocol represents bearer protocol (typically single-valued: ip, ipv6, ppp).
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

// String returns protocol name (e.g., "ipv4").
func (bearerProtocolValue APNTypeBearerProtocol) String() string {
	return apnTypeBearerProtocolStorage.JSONMap.GetString(bearerProtocolValue)
}

// MarshalText serializes to string (e.g., "ipv4").
func (bearerProtocolValue APNTypeBearerProtocol) MarshalText() (textByte []byte, err error) {
	return apnTypeBearerProtocolStorage.MarshalTextValue(bearerProtocolValue)
}

// UnmarshalText deserializes from string (e.g., "ipv4").
func (bearerProtocolValue *APNTypeBearerProtocol) UnmarshalText(textByte []byte) error {
	return apnTypeBearerProtocolStorage.UnmarshalTextValue(bearerProtocolValue, textByte)
}

// MarshalJSON serializes to JSON string (e.g., "ipv4").
func (bearerProtocolValue APNTypeBearerProtocol) MarshalJSON() (jsonByte []byte, err error) {
	return apnTypeBearerProtocolStorage.MarshalJSONValue(bearerProtocolValue)
}

// UnmarshalJSON deserializes from JSON string (e.g., "ipv4").
func (bearerProtocolValue *APNTypeBearerProtocol) UnmarshalJSON(jsonByte []byte) error {
	return apnTypeBearerProtocolStorage.UnmarshalJSONValue(bearerProtocolValue, jsonByte)
}

// MarshalXMLAttr serializes to XML attribute per proxy options.
func (bearerProtocolValue APNTypeBearerProtocol) MarshalXMLAttr(xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	return apnTypeBearerProtocolStorage.MarshalXMLValue(bearerProtocolValue, xmlAttrName)
}

// UnmarshalXMLAttr deserializes from XML attribute per proxy options.
func (bearerProtocolValue *APNTypeBearerProtocol) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return apnTypeBearerProtocolStorage.UnmarshalXMLValue(bearerProtocolValue, xmlAttr)
}

//--------------------------------------------------------------------------------//
