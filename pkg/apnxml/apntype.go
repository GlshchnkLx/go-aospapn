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
// APNType BaseType
//--------------------------------------------------------------------------------//

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

var (
	_APNTypeBaseTypeIndexArray = []APNTypeBaseType{}

	_APNTypeBaseTypeMapByIndex = map[APNTypeBaseType]string{
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
	}

	_APNTypeBaseTypeMapByString = map[string]APNTypeBaseType{
		"":  APNTYPE_BASE_TYPE_NONE,
		"*": APNTYPE_BASE_TYPE_ALL,
	}
)

func init() {
	for index, value := range _APNTypeBaseTypeMapByIndex {
		if APNTYPE_BASE_TYPE_NONE < index && index < APNTYPE_BASE_TYPE_MAX {
			_APNTypeBaseTypeIndexArray = append(_APNTypeBaseTypeIndexArray, index)
		}

		_APNTypeBaseTypeMapByString[value] = index
	}

	sort.Slice(_APNTypeBaseTypeIndexArray, func(i, j int) bool {
		return _APNTypeBaseTypeIndexArray[i] < _APNTypeBaseTypeIndexArray[j]
	})
}

func (apnBaseType APNTypeBaseType) GetStringArray() []string {
	var apnBaseTypeStringArray []string

	for _, apnBaseTypeIndex := range _APNTypeBaseTypeIndexArray {
		if apnBaseType&apnBaseTypeIndex == apnBaseTypeIndex {
			apnBaseTypeString := _APNTypeBaseTypeMapByIndex[apnBaseTypeIndex]
			apnBaseTypeStringArray = append(apnBaseTypeStringArray, apnBaseTypeString)
		}
	}

	return apnBaseTypeStringArray
}

func (apnBaseType APNTypeBaseType) GetString(separator string) string {
	return strings.Join(apnBaseType.GetStringArray(), separator)
}

func (apnBaseType *APNTypeBaseType) SetStringArray(apnBaseTypeStringArray []string) error {
	*apnBaseType = 0

	for _, apnBaseTypeString := range apnBaseTypeStringArray {
		apnBaseTypeString = strings.ToLower(strings.TrimSpace(apnBaseTypeString))

		apnBaseTypeIndex, ok := _APNTypeBaseTypeMapByString[apnBaseTypeString]
		if !ok {
			return fmt.Errorf("apn base type has uncorrected string: %s", apnBaseTypeString)
		}

		*apnBaseType |= apnBaseTypeIndex
	}

	return nil
}

func (apnBaseType *APNTypeBaseType) SetString(apnBaseTypeString string, separator string) error {
	return apnBaseType.SetStringArray(strings.Split(apnBaseTypeString, separator))
}

func (apnBaseType APNTypeBaseType) String() string {
	return apnBaseType.GetString("|")
}

func (apnBaseType APNTypeBaseType) MarshalXMLAttr(name xml.Name) (attr xml.Attr, err error) {
	attr = xml.Attr{
		Name:  name,
		Value: apnBaseType.GetString(","),
	}

	return
}

func (apnBaseType *APNTypeBaseType) UnmarshalXMLAttr(attr xml.Attr) error {
	if attr.Value == "" {
		*apnBaseType = APNTYPE_BASE_TYPE_NONE
		return nil
	}

	return apnBaseType.SetString(attr.Value, ",")
}

func (apnBaseType APNTypeBaseType) MarshalJSON() ([]byte, error) {
	return json.Marshal(apnBaseType.GetStringArray())
}

func (apnBaseType *APNTypeBaseType) UnmarshalJSON(jsonData []byte) error {
	var (
		apnBaseTypeStringArray []string
		err                    error
	)

	err = json.Unmarshal(jsonData, &apnBaseTypeStringArray)
	if err != nil {
		return err
	}

	return apnBaseType.SetStringArray(apnBaseTypeStringArray)
}

//--------------------------------------------------------------------------------//
// APNType AuthType
//--------------------------------------------------------------------------------//

type APNTypeAuthType int

const (
	APNTYPE_AUTH_TYPE_NONE APNTypeAuthType = 0
	APNTYPE_AUTH_TYPE_PAP  APNTypeAuthType = 1 << (iota - 1)
	APNTYPE_AUTH_TYPE_CHAP
	APNTYPE_AUTH_TYPE_MAX

	APNTYPE_AUTH_TYPE_ALL = APNTYPE_AUTH_TYPE_MAX - 1
)

var (
	_APNTypeAuthTypeIndexArray = []APNTypeAuthType{}

	_APNTypeAuthTypeMapByIndex = map[APNTypeAuthType]string{
		APNTYPE_AUTH_TYPE_NONE: "none",
		APNTYPE_AUTH_TYPE_PAP:  "pap",
		APNTYPE_AUTH_TYPE_CHAP: "chap",
	}

	_APNTypeAuthTypeMapByString = map[string]APNTypeAuthType{
		"":  APNTYPE_AUTH_TYPE_NONE,
		"*": APNTYPE_AUTH_TYPE_ALL,
	}
)

func init() {
	for index, value := range _APNTypeAuthTypeMapByIndex {
		if APNTYPE_AUTH_TYPE_NONE < index && index < APNTYPE_AUTH_TYPE_MAX {
			_APNTypeAuthTypeIndexArray = append(_APNTypeAuthTypeIndexArray, index)
		}

		_APNTypeAuthTypeMapByString[value] = index
	}

	sort.Slice(_APNTypeAuthTypeIndexArray, func(i, j int) bool {
		return _APNTypeAuthTypeIndexArray[i] < _APNTypeAuthTypeIndexArray[j]
	})
}

func (apnAuthType APNTypeAuthType) GetStringArray() []string {
	var apnAuthTypeStringArray []string

	for _, apnAuthTypeIndex := range _APNTypeAuthTypeIndexArray {
		if apnAuthType&apnAuthTypeIndex == apnAuthTypeIndex {
			apnAuthTypeString := _APNTypeAuthTypeMapByIndex[apnAuthTypeIndex]
			apnAuthTypeStringArray = append(apnAuthTypeStringArray, apnAuthTypeString)
		}
	}

	return apnAuthTypeStringArray
}

func (apnAuthType APNTypeAuthType) GetString(separator string) string {
	return strings.Join(apnAuthType.GetStringArray(), separator)
}

func (apnAuthType APNTypeAuthType) GetNumber() string {
	return strconv.Itoa(int(apnAuthType))
}

func (apnAuthType *APNTypeAuthType) SetStringArray(apnAuthTypeStringArray []string) error {
	*apnAuthType = 0

	for _, apnAuthTypeString := range apnAuthTypeStringArray {
		apnAuthTypeString = strings.ToLower(strings.TrimSpace(apnAuthTypeString))

		apnAuthTypeIndex, ok := _APNTypeAuthTypeMapByString[apnAuthTypeString]
		if !ok {
			return fmt.Errorf("apn auth type has uncorrected string: %s", apnAuthTypeString)
		}

		*apnAuthType |= apnAuthTypeIndex
	}

	return nil
}

func (apnAuthType *APNTypeAuthType) SetString(apnAuthTypeString string, separator string) error {
	return apnAuthType.SetStringArray(strings.Split(apnAuthTypeString, separator))
}

func (apnAuthType *APNTypeAuthType) SetNumber(apnAuthTypeNumber string) error {
	apnAuthTypeInt, err := strconv.Atoi(apnAuthTypeNumber)
	if err != nil {
		return err
	}

	_, ok := _APNTypeAuthTypeMapByIndex[APNTypeAuthType(apnAuthTypeInt)]
	if !ok {
		if APNTYPE_AUTH_TYPE_NONE <= APNTypeAuthType(apnAuthTypeInt) && APNTypeAuthType(apnAuthTypeInt) < APNTYPE_AUTH_TYPE_MAX {
			*apnAuthType = APNTypeAuthType(apnAuthTypeInt)
		} else {
			*apnAuthType = APNTYPE_AUTH_TYPE_NONE
		}
	} else {
		*apnAuthType = APNTypeAuthType(apnAuthTypeInt)
	}

	return nil
}

func (apnAuthType APNTypeAuthType) String() string {
	return apnAuthType.GetString("|")
}

func (apnAuthType APNTypeAuthType) MarshalXMLAttr(name xml.Name) (attr xml.Attr, err error) {
	attr = xml.Attr{
		Name:  name,
		Value: apnAuthType.GetNumber(),
	}

	return
}

func (apnAuthType *APNTypeAuthType) UnmarshalXMLAttr(attr xml.Attr) error {
	if attr.Value == "" {
		*apnAuthType = APNTYPE_AUTH_TYPE_NONE
		return nil
	}

	return apnAuthType.SetNumber(attr.Value)
}

func (apnAuthType APNTypeAuthType) MarshalJSON() ([]byte, error) {
	return json.Marshal(apnAuthType.GetStringArray())
}

func (apnAuthType *APNTypeAuthType) UnmarshalJSON(jsonData []byte) error {
	var (
		apnAuthTypeStringArray []string
		err                    error
	)

	err = json.Unmarshal(jsonData, &apnAuthTypeStringArray)
	if err != nil {
		return err
	}
	return apnAuthType.SetStringArray(apnAuthTypeStringArray)
}

//--------------------------------------------------------------------------------//
// APNType BearerProtocol
//--------------------------------------------------------------------------------//

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

var (
	_APNTypeBearerProtocolIndexArray = []APNTypeBearerProtocol{}

	_APNTypeBearerProtocolMapByIndex = map[APNTypeBearerProtocol]string{
		APNTYPE_BEARER_PROTOCOL_IP:           "ip",
		APNTYPE_BEARER_PROTOCOL_IPV4:         "ipv4",
		APNTYPE_BEARER_PROTOCOL_IPV6:         "ipv6",
		APNTYPE_BEARER_PROTOCOL_IPV4V6:       "ipv4v6",
		APNTYPE_BEARER_PROTOCOL_PPP:          "ppp",
		APNTYPE_BEARER_PROTOCOL_NONIP:        "non-ip",
		APNTYPE_BEARER_PROTOCOL_UNSTRUCTURED: "unstructured",
	}

	_APNTypeBearerProtocolMapByString = map[string]APNTypeBearerProtocol{
		"":     APNTYPE_BEARER_PROTOCOL_NONE,
		"none": APNTYPE_BEARER_PROTOCOL_NONE,
		"*":    APNTYPE_BEARER_PROTOCOL_IPV4V6,
	}
)

func init() {
	for index, value := range _APNTypeBearerProtocolMapByIndex {
		_APNTypeBearerProtocolIndexArray = append(_APNTypeBearerProtocolIndexArray, index)
		_APNTypeBearerProtocolMapByString[value] = index
	}

	sort.Slice(_APNTypeBearerProtocolIndexArray, func(i, j int) bool {
		return _APNTypeBearerProtocolIndexArray[i] < _APNTypeBearerProtocolIndexArray[j]
	})
}

func (apnBearerProtocol APNTypeBearerProtocol) String() string {
	apnBearerProtocolString, ok := _APNTypeBearerProtocolMapByIndex[apnBearerProtocol]
	if !ok {
		return ""
	}

	return apnBearerProtocolString
}

func (apnBearerProtocol APNTypeBearerProtocol) MarshalXMLAttr(name xml.Name) (attr xml.Attr, err error) {
	attr = xml.Attr{
		Name:  name,
		Value: strings.ToUpper(apnBearerProtocol.String()),
	}

	return
}

func (apnBearerProtocol *APNTypeBearerProtocol) UnmarshalXMLAttr(attr xml.Attr) error {
	var ok bool

	*apnBearerProtocol, ok = _APNTypeBearerProtocolMapByString[strings.ToLower(attr.Value)]
	if !ok {
		return fmt.Errorf("apn bearer protocol has uncorrected string: %s", attr.Value)
	}

	return nil
}

func (apnBearerProtocol APNTypeBearerProtocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(apnBearerProtocol.String())
}

func (apnBearerProtocol *APNTypeBearerProtocol) UnmarshalJSON(jsonData []byte) error {
	var ok bool

	*apnBearerProtocol, ok = _APNTypeBearerProtocolMapByString[strings.Trim(string(jsonData), "\"")]
	if !ok {
		return fmt.Errorf("apn bearer protocol has uncorrected string: %s", string(jsonData))
	}

	return nil
}

//--------------------------------------------------------------------------------//
