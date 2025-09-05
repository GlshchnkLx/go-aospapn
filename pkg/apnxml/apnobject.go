package apnxml

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
)

//--------------------------------------------------------------------------------//
// APNObject Interface
//--------------------------------------------------------------------------------//

type APNObjectInterface interface {
	IsExist() bool
}

//--------------------------------------------------------------------------------//
// APNObject Core & Helper
//--------------------------------------------------------------------------------//

type APNObject struct {
	*APNObjectRoot

	Base   *APNObjectBase   `json:"base,omitempty"`
	Auth   *APNObjectAuth   `json:"auth,omitempty"`
	Bearer *APNObjectBearer `json:"bearer,omitempty"`
	Proxy  *APNObjectProxy  `json:"proxy,omitempty"`
	Mms    *APNObjectMms    `json:"mms,omitempty"`
	Mvno   *APNObjectMvno   `json:"mvno,omitempty"`
	Limit  *APNObjectLimit  `json:"limit,omitempty"`
	Other  *APNObjectOther  `json:"other,omitempty"`

	GroupMapByType map[string]APNObject `json:"groupMap,omitempty"`
}

type helperAPNObject struct {
	*APNObjectRoot   `xml:",omitempty"`
	*APNObjectBase   `xml:",omitempty"`
	*APNObjectAuth   `xml:",omitempty"`
	*APNObjectBearer `xml:",omitempty"`
	*APNObjectProxy  `xml:",omitempty"`
	*APNObjectMms    `xml:",omitempty"`
	*APNObjectMvno   `xml:",omitempty"`
	*APNObjectLimit  `xml:",omitempty"`
	*APNObjectOther  `xml:",omitempty"`
}

func (apnObject APNObject) String() string {
	jsonData, err := json.MarshalIndent(apnObject, "", "\t")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return string(jsonData)
}

func (apnObject APNObject) MarshalXML(xmlEncoder *xml.Encoder, xmlStart xml.StartElement) error {
	var _apnObject helperAPNObject

	if apnObject.APNObjectRoot.IsExist() {
		_apnObject.APNObjectRoot = apnObject.APNObjectRoot
	}

	if apnObject.Base.IsExist() {
		_apnObject.APNObjectBase = apnObject.Base
	}

	if apnObject.Auth.IsExist() {
		_apnObject.APNObjectAuth = apnObject.Auth
	}

	if apnObject.Bearer.IsExist() {
		_apnObject.APNObjectBearer = apnObject.Bearer
	}

	if apnObject.Proxy.IsExist() {
		_apnObject.APNObjectProxy = apnObject.Proxy
	}

	if apnObject.Mms.IsExist() {
		_apnObject.APNObjectMms = apnObject.Mms
	}

	if apnObject.Mvno.IsExist() {
		_apnObject.APNObjectMvno = apnObject.Mvno
	}

	if apnObject.Limit.IsExist() {
		_apnObject.APNObjectLimit = apnObject.Limit
	}

	if apnObject.Other.IsExist() {
		_apnObject.APNObjectOther = apnObject.Other
	}

	err := xmlEncoder.EncodeElement(_apnObject, xmlStart)
	if err != nil {
		return err
	}

	return nil
}

func (apnPointer *APNObject) UnmarshalXML(xmlDecoder *xml.Decoder, xmlStart xml.StartElement) error {
	var _apnObject helperAPNObject

	err := xmlDecoder.DecodeElement(&_apnObject, &xmlStart)
	if err != nil {
		return err
	}

	if _apnObject.APNObjectRoot.IsExist() {
		apnPointer.APNObjectRoot = _apnObject.APNObjectRoot
	}

	if _apnObject.APNObjectBase.IsExist() {
		apnPointer.Base = _apnObject.APNObjectBase
	}

	if _apnObject.APNObjectAuth.IsExist() {
		apnPointer.Auth = _apnObject.APNObjectAuth
	}

	if _apnObject.APNObjectBearer.IsExist() {
		apnPointer.Bearer = _apnObject.APNObjectBearer
	}

	if _apnObject.APNObjectProxy.IsExist() {
		apnPointer.Proxy = _apnObject.APNObjectProxy
	}

	if _apnObject.APNObjectMms.IsExist() {
		apnPointer.Mms = _apnObject.APNObjectMms
	}

	if _apnObject.APNObjectMvno.IsExist() {
		apnPointer.Mvno = _apnObject.APNObjectMvno
	}

	if _apnObject.APNObjectLimit.IsExist() {
		apnPointer.Limit = _apnObject.APNObjectLimit
	}

	if _apnObject.APNObjectOther.IsExist() {
		apnPointer.Other = _apnObject.APNObjectOther
	}

	return nil
}

func (apnObject APNObject) Clone() (_apnObject APNObject) {
	jsonData, err := json.Marshal(apnObject)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(jsonData, &_apnObject)
	if err != nil {
		panic(err)
	}

	return
}

//--------------------------------------------------------------------------------//
// APN Unit Root
//--------------------------------------------------------------------------------//

type APNObjectRoot struct {
	APNObjectInterface `json:"-" xml:"-"`

	Carrier   string `json:"carrierName" xml:"carrier,attr,omitempty"`
	CarrierID *int   `json:"carrierID,omitempty"   xml:"carrier_id,attr,omitempty"`
	Mcc       *int   `json:"mcc,omitempty"         xml:"mcc,attr,omitempty"`
	Mnc       *int   `json:"mnc,omitempty"         xml:"mnc,attr,omitempty"`
}

func (apnPointerRoot *APNObjectRoot) IsExist() bool {
	if apnPointerRoot != nil && (apnPointerRoot.Mcc != nil && apnPointerRoot.Mnc != nil) {
		return true
	}

	return false
}

var _APNObjectRootCarrierWordMask = map[string]bool{
	"internet": false,
	"5g":       false,
	"4g":       false,
	"3g":       false,
	"2g":       false,
	"nsa":      false,
	"sa":       false,
	"lte":      false,
	"wap":      false,
	"gprs":     false,
	"web":      false,
}

func init() {
	for _, value := range _APNTypeBaseTypeMapByIndex {
		_APNObjectRootCarrierWordMask[value] = false
	}
}

func (apnRoot APNObjectRoot) GetCarrier() string {
	var (
		apnRootCarrierString     = apnRoot.Carrier
		apnRootCarrierWordArray  = strings.Split(strings.TrimSpace(apnRootCarrierString), " ")
		_apnRootCarrierWordArray = []string{}
	)

	if strings.Contains(apnRootCarrierString, " ") {
		apnRootCarrierWordArray = strings.Split(apnRootCarrierString, " ")
	} else if strings.Contains(apnRootCarrierString, ":") {
		apnRootCarrierWordArray = strings.Split(apnRootCarrierString, ":")
	} else if strings.Contains(apnRootCarrierString, ".") {
		apnRootCarrierWordArray = strings.Split(apnRootCarrierString, ".")
	} else if strings.Contains(apnRootCarrierString, "_") {
		apnRootCarrierWordArray = strings.Split(apnRootCarrierString, "_")
	} else if strings.Contains(apnRootCarrierString, "-") {
		apnRootCarrierWordArray = strings.Split(apnRootCarrierString, "-")
	} else {
		return apnRootCarrierString
	}

	for _, apnRootCarrierWord := range apnRootCarrierWordArray {
		apnRootCarrierWord = strings.Trim(apnRootCarrierWord, "-_.,:; 0123456789")

		_, ok := _APNObjectRootCarrierWordMask[strings.ToLower(apnRootCarrierWord)]
		if len(apnRootCarrierWord) > 0 && !ok {
			_apnRootCarrierWordArray = append(_apnRootCarrierWordArray, apnRootCarrierWord)
		}
	}

	if len(_apnRootCarrierWordArray) == 0 {
		return apnRootCarrierString
	}

	return strings.Join(_apnRootCarrierWordArray, " ")
}

func (apnRoot APNObjectRoot) GetID() string {
	if apnRoot.CarrierID != nil {
		return fmt.Sprintf("CID:%d", *apnRoot.CarrierID)
	}

	return fmt.Sprintf("PLMN:%s", apnRoot.GetPLMN())
}

func (apnRoot APNObjectRoot) GetPLMN() string {
	return fmt.Sprintf("%03d%02d", *apnRoot.Mcc, *apnRoot.Mnc)
}

//--------------------------------------------------------------------------------//
// APN Unit Base
//--------------------------------------------------------------------------------//

type APNObjectBase struct {
	APNObjectInterface `json:"-" xml:"-"`

	Apn       *string          `json:"apn,omitempty"       xml:"apn,attr,omitempty"`
	Type      *APNTypeBaseType `json:"type,omitempty" xml:"type,attr,omitempty"`
	ProfileID *string          `json:"profileID,omitempty" xml:"profile_id,attr,omitempty"`
}

func (apnPointerBase *APNObjectBase) IsExist() bool {
	if apnPointerBase != nil && (apnPointerBase.Apn != nil || apnPointerBase.Type != nil || apnPointerBase.ProfileID != nil) {
		return true
	}

	return false
}

//--------------------------------------------------------------------------------//
// APN Unit Auth
//--------------------------------------------------------------------------------//

type APNObjectAuth struct {
	APNObjectInterface `json:"-" xml:"-"`

	Type     *APNTypeAuthType `json:"type,omitempty" xml:"authtype,attr,omitempty"`
	Username *string          `json:"username,omitempty" xml:"user,attr,omitempty"`
	Password *string          `json:"password,omitempty" xml:"password,attr,omitempty"`
}

func (apnPointerAuth *APNObjectAuth) IsExist() bool {
	if apnPointerAuth != nil {
		if apnPointerAuth.Type != nil {
			if apnPointerAuth.Username != nil || apnPointerAuth.Password != nil {
				var value = ""

				if apnPointerAuth.Username == nil {
					apnPointerAuth.Username = &value
				}

				if apnPointerAuth.Password == nil {
					apnPointerAuth.Password = &value
				}

				return true
			}
		}
	}

	return false
}

//--------------------------------------------------------------------------------//
// APN Unit Bearer
//--------------------------------------------------------------------------------//

type APNObjectBearer struct {
	APNObjectInterface `json:"-" xml:"-"`

	Type        *APNTypeBearerProtocol `json:"type,omitempty"         xml:"protocol,attr,omitempty"`
	TypeRoaming *APNTypeBearerProtocol `json:"typeRoaming,omitempty"  xml:"roaming_protocol,attr,omitempty"`
	Mtu         *int                   `json:"mtu,omitempty"          xml:"mtu,attr,omitempty"`
	Server      *string                `json:"server,omitempty"       xml:"server,attr,omitempty"`
}

func (apnPointerBearer *APNObjectBearer) IsExist() bool {
	if apnPointerBearer != nil && (apnPointerBearer.Type != nil || apnPointerBearer.TypeRoaming != nil) {
		return true
	}

	return false
}

//--------------------------------------------------------------------------------//
// APN Unit Proxy
//--------------------------------------------------------------------------------//

type APNObjectProxy struct {
	APNObjectInterface `json:"-" xml:"-"`

	Server *string `json:"server,omitempty" xml:"proxy,attr,omitempty"`
	Port   *int    `json:"port,omitempty"   xml:"port,attr,omitempty"`
}

func (apnPointerProxy *APNObjectProxy) IsExist() bool {
	if apnPointerProxy != nil && ((apnPointerProxy.Server != nil && *apnPointerProxy.Server != "") && apnPointerProxy.Port != nil) {
		return true
	}

	return false
}

//--------------------------------------------------------------------------------//
// APN Unit MMS
//--------------------------------------------------------------------------------//

type APNObjectMms struct {
	APNObjectInterface `json:"-" xml:"-"`

	Center *string `json:"center,omitempty" xml:"mmsc,attr,omitempty"`
	Server *string `json:"server,omitempty" xml:"mmsproxy,attr,omitempty"`
	Port   *int    `json:"port,omitempty"   xml:"mmsport,attr,omitempty"`
}

func (apnPointerMms *APNObjectMms) IsExist() bool {
	if apnPointerMms != nil && ((apnPointerMms.Center != nil && *apnPointerMms.Center != "") || (apnPointerMms.Server != nil && *apnPointerMms.Server != "") || apnPointerMms.Port != nil) {
		return true
	}

	return false
}

//--------------------------------------------------------------------------------//
// APN Unit MVNO
//--------------------------------------------------------------------------------//

type APNObjectMvno struct {
	APNObjectInterface `json:"-" xml:"-"`

	Type *string `json:"type,omitempty" xml:"mvno_type,attr,omitempty"`
	Data *string `json:"data,omitempty" xml:"mvno_match_data,attr,omitempty"`
}

func (apnPointerMvno *APNObjectMvno) IsExist() bool {
	if apnPointerMvno != nil && (apnPointerMvno.Type != nil || apnPointerMvno.Data != nil) {
		return true
	}

	return false
}

//--------------------------------------------------------------------------------//
// APN Unit Limit
//--------------------------------------------------------------------------------//

type APNObjectLimit struct {
	APNObjectInterface `json:"-" xml:"-"`

	MaxConn     *int `json:"maxConn,omitempty"      xml:"max_conns,attr,omitempty"`
	MaxConnTime *int `json:"maxConnTime,omitempty"  xml:"max_conns_time,attr,omitempty"`
}

func (apnPointerLimit *APNObjectLimit) IsExist() bool {
	if apnPointerLimit != nil && (apnPointerLimit.MaxConn != nil || apnPointerLimit.MaxConnTime != nil) {
		return true
	}

	return false
}

//--------------------------------------------------------------------------------//
// APN Unit Other
//--------------------------------------------------------------------------------//

type APNObjectOther struct {
	APNObjectInterface `json:"-" xml:"-"`

	// Network restrictions
	NetworkTypeBitmask *string `json:"networkTypeBitmask,omitempty" xml:"network_type_bitmask,attr,omitempty"`

	// Modem settings
	ModemCognitive *bool `json:"modemCognitive,omitempty" xml:"modem_cognitive,attr,omitempty"`

	// Carrier control flags
	CarrierEnabled *bool `json:"IsEnabled,omitempty" xml:"carrier_enabled,attr,omitempty"`
	UserVisible    *bool `json:"IsVisible,omitempty" xml:"user_visible,attr,omitempty"`
	UserEditable   *bool `json:"IsEditable,omitempty" xml:"user_editable,attr,omitempty"`
}

func (apnPointerOther *APNObjectOther) IsExist() bool {
	if apnPointerOther != nil && (apnPointerOther.NetworkTypeBitmask != nil || apnPointerOther.ModemCognitive != nil || apnPointerOther.CarrierEnabled != nil || apnPointerOther.UserVisible != nil || apnPointerOther.UserEditable != nil) {
		return true
	}

	return false
}

//--------------------------------------------------------------------------------//
