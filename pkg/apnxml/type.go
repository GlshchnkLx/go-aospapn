package apnxml

import (
	"encoding/xml"
	"strings"
)

//--------------------------------------------------------------------------------//
// ObjectBaseType
//--------------------------------------------------------------------------------//

type ObjectBaseType int

const (
	ObjectBaseTypeNone    ObjectBaseType = 0
	ObjectBaseTypeDefault ObjectBaseType = 1 << (iota - 1)
	ObjectBaseTypeMMS
	ObjectBaseTypeSUPL
	ObjectBaseTypeDUN
	ObjectBaseTypeHIPRI
	ObjectBaseTypeFOTA
	ObjectBaseTypeIMS
	ObjectBaseTypeCBS
	ObjectBaseTypeIA
	ObjectBaseTypeEmergency
	ObjectBaseTypeMCX
	ObjectBaseTypeXCAP
	ObjectBaseTypeVSIM
	ObjectBaseTypeBIP
	ObjectBaseTypeEnterprise
	ObjectBaseTypeRCS
	ObjectBaseTypeOEMPaid
	ObjectBaseTypeOEMPrivate
	ObjectBaseTypeMax
)

var apnTypeBaseTypeStorage = newEnumCodec(
	ObjectBaseTypeNone,
	ObjectBaseTypeMax,
	map[ObjectBaseType]string{
		ObjectBaseTypeNone:       "None",
		ObjectBaseTypeDefault:    "default",
		ObjectBaseTypeMMS:        "mms",
		ObjectBaseTypeSUPL:       "supl",
		ObjectBaseTypeDUN:        "dun",
		ObjectBaseTypeHIPRI:      "hipri",
		ObjectBaseTypeFOTA:       "fota",
		ObjectBaseTypeIMS:        "ims",
		ObjectBaseTypeCBS:        "cbs",
		ObjectBaseTypeIA:         "ia",
		ObjectBaseTypeEmergency:  "emergency",
		ObjectBaseTypeMCX:        "mcx",
		ObjectBaseTypeXCAP:       "xcap",
		ObjectBaseTypeVSIM:       "vsim",
		ObjectBaseTypeBIP:        "bip",
		ObjectBaseTypeEnterprise: "enterprise",
		ObjectBaseTypeRCS:        "rcs",
		ObjectBaseTypeOEMPaid:    "oem_paid",
		ObjectBaseTypeOEMPrivate: "oem_private",
	},
	newEnumCodecOptions().SetJSONIsArray(true).SetXMLIsArray(",").SetXMLIsString(false),
)

func (baseTypeValue ObjectBaseType) String() string {
	return strings.Join(apnTypeBaseTypeStorage.json.GetStringArray(baseTypeValue), "|")
}

func (baseTypeValue ObjectBaseType) MarshalText() (textByte []byte, err error) {
	return apnTypeBaseTypeStorage.marshalText(baseTypeValue)
}

func (baseTypeValue *ObjectBaseType) UnmarshalText(textByte []byte) error {
	return apnTypeBaseTypeStorage.unmarshalText(baseTypeValue, textByte)
}

func (baseTypeValue ObjectBaseType) MarshalJSON() (jsonByte []byte, err error) {
	return apnTypeBaseTypeStorage.marshalJSON(baseTypeValue)
}

func (baseTypeValue *ObjectBaseType) UnmarshalJSON(jsonByte []byte) error {
	return apnTypeBaseTypeStorage.unmarshalJSON(baseTypeValue, jsonByte)
}

func (baseTypeValue ObjectBaseType) MarshalXMLAttr(xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	return apnTypeBaseTypeStorage.marshalXMLAttr(baseTypeValue, xmlAttrName)
}

func (baseTypeValue *ObjectBaseType) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return apnTypeBaseTypeStorage.unmarshalXMLAttr(baseTypeValue, xmlAttr)
}

//--------------------------------------------------------------------------------//
// ObjectAuthType
//--------------------------------------------------------------------------------//

type ObjectAuthType int

const (
	ObjectAuthTypeNone ObjectAuthType = 0
	ObjectAuthTypePAP  ObjectAuthType = 1 << (iota - 1)
	ObjectAuthTypeCHAP
	ObjectAuthTypeMax
)

var apnTypeAuthTypeStorage = newEnumCodec(
	ObjectAuthTypeNone,
	ObjectAuthTypeMax,
	map[ObjectAuthType]string{
		ObjectAuthTypeNone: "None",
		ObjectAuthTypePAP:  "pap",
		ObjectAuthTypeCHAP: "chap",
	},
	newEnumCodecOptions().SetJSONIsArray(true).SetXMLIsNumber(false),
)

func (authTypeValue ObjectAuthType) String() string {
	return strings.Join(apnTypeAuthTypeStorage.json.GetStringArray(authTypeValue), "|")
}

func (authTypeValue ObjectAuthType) MarshalText() (textByte []byte, err error) {
	return apnTypeAuthTypeStorage.marshalText(authTypeValue)
}

func (authTypeValue *ObjectAuthType) UnmarshalText(textByte []byte) error {
	return apnTypeAuthTypeStorage.unmarshalText(authTypeValue, textByte)
}

func (authTypeValue ObjectAuthType) MarshalJSON() (jsonByte []byte, err error) {
	return apnTypeAuthTypeStorage.marshalJSON(authTypeValue)
}

func (authTypeValue *ObjectAuthType) UnmarshalJSON(jsonByte []byte) error {
	return apnTypeAuthTypeStorage.unmarshalJSON(authTypeValue, jsonByte)
}

func (authTypeValue ObjectAuthType) MarshalXMLAttr(xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	return apnTypeAuthTypeStorage.marshalXMLAttr(authTypeValue, xmlAttrName)
}

func (authTypeValue *ObjectAuthType) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return apnTypeAuthTypeStorage.unmarshalXMLAttr(authTypeValue, xmlAttr)
}

//--------------------------------------------------------------------------------//
// ObjectNetworkType
//--------------------------------------------------------------------------------//

type ObjectNetworkType int

const (
	ObjectNetworkTypeNone ObjectNetworkType = 0
	ObjectNetworkTypeGPRS ObjectNetworkType = 1 << (iota - 1)
	ObjectNetworkTypeEDGE
	ObjectNetworkTypeUMTS
	ObjectNetworkTypeCDMA
	ObjectNetworkTypeEVDO0
	ObjectNetworkTypeEVDOA
	ObjectNetworkType1xRTT
	ObjectNetworkTypeHSDPA
	ObjectNetworkTypeHSUPA
	ObjectNetworkTypeHSPA
	ObjectNetworkTypeIDEN
	ObjectNetworkTypeEVDOB
	ObjectNetworkTypeLTE
	ObjectNetworkTypeEHRPD
	ObjectNetworkTypeHSPAP
	ObjectNetworkTypeGSM
	ObjectNetworkTypeTDSCDMA
	ObjectNetworkTypeIWLAN
	ObjectNetworkTypeLTECA
	ObjectNetworkTypeNR
	ObjectNetworkTypeMax
)

var apnTypeNetworkTypeStorage = newEnumCodec(
	ObjectNetworkTypeNone,
	ObjectNetworkTypeMax,
	map[ObjectNetworkType]string{
		ObjectNetworkTypeNone:    "unknown",
		ObjectNetworkTypeGPRS:    "gprs",
		ObjectNetworkTypeEDGE:    "edge",
		ObjectNetworkTypeUMTS:    "umts",
		ObjectNetworkTypeCDMA:    "cdma",
		ObjectNetworkTypeEVDO0:   "evdo_0",
		ObjectNetworkTypeEVDOA:   "evdo_a",
		ObjectNetworkType1xRTT:   "1xrtt",
		ObjectNetworkTypeHSDPA:   "hsdpa",
		ObjectNetworkTypeHSUPA:   "hsupa",
		ObjectNetworkTypeHSPA:    "hspa",
		ObjectNetworkTypeIDEN:    "iden",
		ObjectNetworkTypeEVDOB:   "evdo_b",
		ObjectNetworkTypeLTE:     "lte",
		ObjectNetworkTypeEHRPD:   "ehrpd",
		ObjectNetworkTypeHSPAP:   "hspap",
		ObjectNetworkTypeGSM:     "gsm",
		ObjectNetworkTypeTDSCDMA: "td_scdma",
		ObjectNetworkTypeIWLAN:   "iwlan",
		ObjectNetworkTypeLTECA:   "lte_ca",
		ObjectNetworkTypeNR:      "nr",
	},
	newEnumCodecOptions().SetJSONIsArray(true).SetXMLIsArray("|").SetXMLIsNumber(true),
)

func (networkTypeValue ObjectNetworkType) String() string {
	return strings.Join(apnTypeNetworkTypeStorage.json.GetStringArray(networkTypeValue), "|")
}

func (networkTypeValue ObjectNetworkType) MarshalText() (textByte []byte, err error) {
	return apnTypeNetworkTypeStorage.marshalText(networkTypeValue)
}

func (networkTypeValue *ObjectNetworkType) UnmarshalText(textByte []byte) error {
	return apnTypeNetworkTypeStorage.unmarshalText(networkTypeValue, textByte)
}

func (networkTypeValue ObjectNetworkType) MarshalJSON() (jsonByte []byte, err error) {
	return apnTypeNetworkTypeStorage.marshalJSON(networkTypeValue)
}

func (networkTypeValue *ObjectNetworkType) UnmarshalJSON(jsonByte []byte) error {
	return apnTypeNetworkTypeStorage.unmarshalJSON(networkTypeValue, jsonByte)
}

func (networkTypeValue ObjectNetworkType) MarshalXMLAttr(xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	return apnTypeNetworkTypeStorage.marshalXMLAttr(networkTypeValue, xmlAttrName)
}

func (networkTypeValue *ObjectNetworkType) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return apnTypeNetworkTypeStorage.unmarshalXMLAttr(networkTypeValue, xmlAttr)
}

//--------------------------------------------------------------------------------//
// ObjectBearerProtocol
//--------------------------------------------------------------------------------//

type ObjectBearerProtocol int

const (
	ObjectBearerProtocolNone ObjectBearerProtocol = iota
	ObjectBearerProtocolIP   ObjectBearerProtocol = 1 << (iota - 1)
	ObjectBearerProtocolIPv4
	ObjectBearerProtocolIPv6
	ObjectBearerProtocolIPv4v6
	ObjectBearerProtocolPPP
	ObjectBearerProtocolNonIP
	ObjectBearerProtocolUnstructured
	ObjectBearerProtocolMax
)

var apnTypeBearerProtocolStorage = newEnumCodec(
	ObjectBearerProtocolNone,
	ObjectBearerProtocolMax,
	map[ObjectBearerProtocol]string{
		ObjectBearerProtocolNone:         "None",
		ObjectBearerProtocolIP:           "ip",
		ObjectBearerProtocolIPv4:         "ipv4",
		ObjectBearerProtocolIPv6:         "ipv6",
		ObjectBearerProtocolIPv4v6:       "ipv4v6",
		ObjectBearerProtocolPPP:          "ppp",
		ObjectBearerProtocolNonIP:        "non-ip",
		ObjectBearerProtocolUnstructured: "unstructured",
	},
	newEnumCodecOptions().SetJSONIsArray(false).SetXMLIsString(true),
)

func (bearerProtocolValue ObjectBearerProtocol) String() string {
	return apnTypeBearerProtocolStorage.json.GetString(bearerProtocolValue)
}

func (bearerProtocolValue ObjectBearerProtocol) MarshalText() (textByte []byte, err error) {
	return apnTypeBearerProtocolStorage.marshalText(bearerProtocolValue)
}

func (bearerProtocolValue *ObjectBearerProtocol) UnmarshalText(textByte []byte) error {
	return apnTypeBearerProtocolStorage.unmarshalText(bearerProtocolValue, textByte)
}

func (bearerProtocolValue ObjectBearerProtocol) MarshalJSON() (jsonByte []byte, err error) {
	return apnTypeBearerProtocolStorage.marshalJSON(bearerProtocolValue)
}

func (bearerProtocolValue *ObjectBearerProtocol) UnmarshalJSON(jsonByte []byte) error {
	return apnTypeBearerProtocolStorage.unmarshalJSON(bearerProtocolValue, jsonByte)
}

func (bearerProtocolValue ObjectBearerProtocol) MarshalXMLAttr(xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	return apnTypeBearerProtocolStorage.marshalXMLAttr(bearerProtocolValue, xmlAttrName)
}

func (bearerProtocolValue *ObjectBearerProtocol) UnmarshalXMLAttr(xmlAttr xml.Attr) error {
	return apnTypeBearerProtocolStorage.unmarshalXMLAttr(bearerProtocolValue, xmlAttr)
}

//--------------------------------------------------------------------------------//
