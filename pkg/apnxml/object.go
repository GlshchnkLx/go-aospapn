package apnxml

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

//--------------------------------------------------------------------------------//
// Object Interface
//--------------------------------------------------------------------------------//

type ObjectInterface[Type any] interface {
	Clone() Type
	Validate() bool
	IsLike(Type) bool
}

//--------------------------------------------------------------------------------//
// Object Core & Helper
//--------------------------------------------------------------------------------//

type Object struct {
	ObjectInterface[Object] `json:"-" xml:"-"`

	*ObjectRoot
	Base   *ObjectBase   `json:"base,omitempty"`
	Auth   *ObjectAuth   `json:"auth,omitempty"`
	Bearer *ObjectBearer `json:"bearer,omitempty"`
	Proxy  *ObjectProxy  `json:"proxy,omitempty"`
	Mms    *ObjectMMS    `json:"mms,omitempty"`
	Mvno   *ObjectMVNO   `json:"mvno,omitempty"`
	Limit  *ObjectLimit  `json:"limit,omitempty"`
	Other  *ObjectOther  `json:"other,omitempty"`

	GroupMapByType map[ObjectBaseType]*Object `json:"groupMap,omitempty"`
}

type helperObject struct {
	*ObjectRoot   `xml:",omitempty"`
	*ObjectBase   `xml:",omitempty"`
	*ObjectAuth   `xml:",omitempty"`
	*ObjectBearer `xml:",omitempty"`
	*ObjectProxy  `xml:",omitempty"`
	*ObjectMMS    `xml:",omitempty"`
	*ObjectMVNO   `xml:",omitempty"`
	*ObjectLimit  `xml:",omitempty"`
	*ObjectOther  `xml:",omitempty"`
}

func (apnPointerCore *Object) Clone() *Object {
	if apnPointerCore == nil {
		return nil
	}

	apnObject := Object{
		ObjectRoot: apnPointerCore.ObjectRoot.Clone(),
		Base:       apnPointerCore.Base.Clone(),
		Auth:       apnPointerCore.Auth.Clone(),
		Bearer:     apnPointerCore.Bearer.Clone(),
		Proxy:      apnPointerCore.Proxy.Clone(),
		Mms:        apnPointerCore.Mms.Clone(),
		Mvno:       apnPointerCore.Mvno.Clone(),
		Limit:      apnPointerCore.Limit.Clone(),
		Other:      apnPointerCore.Other.Clone(),
	}

	if apnPointerCore.GroupMapByType != nil {
		apnObject.GroupMapByType = map[ObjectBaseType]*Object{}

		for apnPointerBaseTypeString, apnPointer := range apnPointerCore.GroupMapByType {
			apnObject.GroupMapByType[apnPointerBaseTypeString] = apnPointer.Clone()
		}
	}

	return &apnObject
}

func (apnPointerCore *Object) HasGroup() bool {
	return apnPointerCore != nil && len(apnPointerCore.GroupMapByType) > 0
}

func (apnPointerCore *Object) CountRecords() int {
	if apnPointerCore == nil {
		return 0
	}

	if len(apnPointerCore.GroupMapByType) == 0 {
		return 1
	}

	return len(apnPointerCore.GroupMapByType)
}

func (apnPointerCore *Object) GroupTypes() []ObjectBaseType {
	if apnPointerCore == nil || len(apnPointerCore.GroupMapByType) == 0 {
		return nil
	}

	apnTypeArray := make([]ObjectBaseType, 0, len(apnPointerCore.GroupMapByType))
	for apnType := range apnPointerCore.GroupMapByType {
		apnTypeArray = append(apnTypeArray, apnType)
	}

	sort.Slice(apnTypeArray, func(i, j int) bool {
		return apnTypeArray[i] < apnTypeArray[j]
	})

	return apnTypeArray
}

func (apnPointerCore *Object) Records() []*Object {
	if apnPointerCore == nil {
		return nil
	}

	if len(apnPointerCore.GroupMapByType) == 0 {
		return []*Object{apnPointerCore}
	}

	apnPointerArray := make([]*Object, 0, len(apnPointerCore.GroupMapByType))
	for _, apnType := range apnPointerCore.GroupTypes() {
		apnPointerArray = append(apnPointerArray, apnPointerCore.GroupMapByType[apnType])
	}

	return apnPointerArray
}

func (apnObjectCore Object) Validate() bool {
	return apnObjectCore.ObjectRoot.Validate()
}

func (apnPointerCore *Object) Normalize() {
	if apnPointerCore == nil {
		return
	}

	apnPointerCore.Base.Normalize()
	apnPointerCore.Auth.Normalize()
	apnPointerCore.Bearer.Normalize()
	apnPointerCore.Proxy.Normalize()
	apnPointerCore.Mms.Normalize()
	apnPointerCore.Mvno.Normalize()
	apnPointerCore.Limit.Normalize()
	apnPointerCore.Other.Normalize()

	for _, apnPointer := range apnPointerCore.GroupMapByType {
		apnPointer.Normalize()
	}
}

func (apnPointerCore *Object) NormalizedClone() *Object {
	apnPointerClone := apnPointerCore.Clone()
	apnPointerClone.Normalize()
	return apnPointerClone
}

func (apnPointerCore *Object) GetIsLikePointer(apnPointerQuery *Object) *Object {
	if apnPointerCore == nil || apnPointerQuery == nil {
		if apnPointerQuery == nil {
			return apnPointerCore
		} else {
			return nil
		}
	}

	if apnPointerCore.ObjectRoot != nil {
		if !apnPointerCore.ObjectRoot.IsLike(apnPointerQuery.ObjectRoot) {
			return nil
		}
	}

	if apnPointerCore.GroupMapByType != nil {
		for _, apnPointer := range apnPointerCore.GroupMapByType {
			apnPointer = apnPointer.GetIsLikePointer(apnPointerQuery)

			if apnPointer != nil {
				return apnPointer
			}
		}

		return nil
	}

	if apnPointerCore.Base.IsLike(apnPointerQuery.Base) &&
		apnPointerCore.Auth.IsLike(apnPointerQuery.Auth) &&
		apnPointerCore.Bearer.IsLike(apnPointerQuery.Bearer) &&
		apnPointerCore.Proxy.IsLike(apnPointerQuery.Proxy) &&
		apnPointerCore.Mms.IsLike(apnPointerQuery.Mms) &&
		apnPointerCore.Mvno.IsLike(apnPointerQuery.Mvno) &&
		apnPointerCore.Limit.IsLike(apnPointerQuery.Limit) &&
		apnPointerCore.Other.IsLike(apnPointerQuery.Other) {

		return apnPointerCore
	}

	return nil
}

func (apnPointerCore *Object) IsLike(apnPointerQuery *Object) bool {
	return apnPointerCore.GetIsLikePointer(apnPointerQuery) != nil
}

func (apnObjectCore Object) String() string {
	jsonData, err := json.MarshalIndent(apnObjectCore, "", "\t")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return string(jsonData)
}

func (apnObjectCore Object) MarshalXML(xmlEncoder *xml.Encoder, xmlStart xml.StartElement) error {
	apnPointerCore := apnObjectCore.NormalizedClone()
	apnObjectHelper := helperObject{
		ObjectRoot:   apnPointerCore.ObjectRoot,
		ObjectBase:   apnPointerCore.Base,
		ObjectAuth:   apnPointerCore.Auth,
		ObjectBearer: apnPointerCore.Bearer,
		ObjectProxy:  apnPointerCore.Proxy,
		ObjectMMS:    apnPointerCore.Mms,
		ObjectMVNO:   apnPointerCore.Mvno,
		ObjectLimit:  apnPointerCore.Limit,
		ObjectOther:  apnPointerCore.Other,
	}

	err := xmlEncoder.EncodeElement(apnObjectHelper, xmlStart)
	if err != nil {
		return err
	}

	return nil
}

func (apnPointerCore *Object) UnmarshalXML(xmlDecoder *xml.Decoder, xmlStart xml.StartElement) error {
	var apnObjectHelper helperObject

	err := xmlDecoder.DecodeElement(&apnObjectHelper, &xmlStart)
	if err != nil {
		return err
	}

	apnPointerCore.ObjectRoot = apnObjectHelper.ObjectRoot.Clone()
	apnPointerCore.Base = apnObjectHelper.ObjectBase.Clone()
	apnPointerCore.Auth = apnObjectHelper.ObjectAuth.Clone()
	apnPointerCore.Bearer = apnObjectHelper.ObjectBearer.Clone()
	apnPointerCore.Proxy = apnObjectHelper.ObjectProxy.Clone()
	apnPointerCore.Mms = apnObjectHelper.ObjectMMS.Clone()
	apnPointerCore.Mvno = apnObjectHelper.ObjectMVNO.Clone()
	apnPointerCore.Limit = apnObjectHelper.ObjectLimit.Clone()
	apnPointerCore.Other = apnObjectHelper.ObjectOther.Clone()
	apnPointerCore.Normalize()

	return nil
}

//--------------------------------------------------------------------------------//
// Object Root
//--------------------------------------------------------------------------------//

type ObjectRoot struct {
	ObjectInterface[ObjectRoot] `json:"-" xml:"-"`

	Carrier   string `json:"carrierName" xml:"carrier,attr,omitempty"`
	CarrierID *int   `json:"carrierID,omitempty"   xml:"carrier_id,attr,omitempty"`
	Mcc       *int   `json:"mcc,omitempty"         xml:"mcc,attr,omitempty"`
	Mnc       *int   `json:"mnc,omitempty"         xml:"mnc,attr,omitempty"`
}

func (apnPointerRoot *ObjectRoot) Clone() *ObjectRoot {
	if apnPointerRoot == nil {
		return nil
	}

	return &ObjectRoot{
		Carrier:   apnPointerRoot.Carrier,
		CarrierID: clonePtr(apnPointerRoot.CarrierID),
		Mcc:       clonePtr(apnPointerRoot.Mcc),
		Mnc:       clonePtr(apnPointerRoot.Mnc),
	}
}

func (apnPointerRoot *ObjectRoot) Validate() bool {
	if apnPointerRoot != nil && (apnPointerRoot.Mcc != nil && apnPointerRoot.Mnc != nil) {
		return true
	}

	return false
}

func (apnPointerRoot *ObjectRoot) IsLike(apnPointer *ObjectRoot) bool {
	if apnPointerRoot == nil || apnPointer == nil {
		return apnPointer == nil
	}

	var (
		isLikeCarrierID = true
		isLikePlmn      = true
	)

	if apnPointer.CarrierID != nil {
		if apnPointerRoot.CarrierID == nil {
			return false
		}

		isLikeCarrierID = *apnPointerRoot.CarrierID == *apnPointer.CarrierID
	}

	if apnPointer.Mcc != nil && apnPointer.Mnc != nil {
		if apnPointerRoot.Mcc == nil || apnPointerRoot.Mnc == nil {
			return false
		}

		isLikePlmn = (*apnPointerRoot.Mcc == *apnPointer.Mcc) && (*apnPointerRoot.Mnc == *apnPointer.Mnc)
	}

	return matchString(apnPointerRoot.Carrier, apnPointer.Carrier) && isLikeCarrierID && isLikePlmn
}

var apnRootCarrierWordMask = map[string]bool{
	"2g":   true,
	"3g":   true,
	"4g":   true,
	"5g":   true,
	"lte":  true,
	"nsa":  true,
	"sa":   true,
	"gprs": true,

	"none":        true,
	"default":     true,
	"mms":         true,
	"supl":        true,
	"dun":         true,
	"hipri":       true,
	"fota":        true,
	"ims":         true,
	"cbs":         true,
	"ia":          true,
	"emergency":   true,
	"mcx":         true,
	"xcap":        true,
	"vsim":        true,
	"bip":         true,
	"enterprise":  true,
	"rcs":         true,
	"oem_paid":    true,
	"oem_private": true,

	"internet": true,
	"data":     true,
	"web":      true,
	"wap":      true,
	"wifi":     true,
	"vowifi":   true,
	"volte":    true,
	"hotspot":  true,
	"tether":   true,
	"ota":      true,
	"admin":    true,
	"ut":       true,

	"-": true,
}

func (apnRoot ObjectRoot) GetCarrier() string {
	var (
		apnRootCarrierString     = apnRoot.Carrier
		apnRootCarrierWordArray  []string
		_apnRootCarrierWordArray []string

		carrierSeparatorArray = []string{" ", "-", "_", ".", ",", ":", ";", "|"}
	)

	for _, carrierSeparator := range carrierSeparatorArray {
		if strings.Contains(apnRootCarrierString, carrierSeparator) {
			apnRootCarrierWordArray = strings.Split(apnRootCarrierString, carrierSeparator)
			break
		}
	}

	if len(apnRootCarrierWordArray) == 0 {
		apnRootCarrierWordArray = append(apnRootCarrierWordArray, apnRootCarrierString)
	}

	for _, apnRootCarrierWord := range apnRootCarrierWordArray {
		_apnRootCarrierWord := strings.TrimSpace(strings.ToLower(apnRootCarrierWord))

		if !apnRootCarrierWordMask[_apnRootCarrierWord] {
			_apnRootCarrierWordArray = append(_apnRootCarrierWordArray, apnRootCarrierWord)
		}
	}

	if len(_apnRootCarrierWordArray) == 0 {
		return apnRootCarrierString
	}

	return strings.Join(_apnRootCarrierWordArray, " ")
}

func (apnRoot ObjectRoot) GetID() string {
	var (
		apnRootID string
	)

	if apnRoot.CarrierID != nil {
		apnRootID += fmt.Sprintf("CID:%d;", *apnRoot.CarrierID)
	}

	apnRootID += fmt.Sprintf("PLMN:%s;", apnRoot.GetPLMN())

	return apnRootID
}

func (apnRoot ObjectRoot) GetPLMN() string {
	if apnRoot.Mcc == nil || apnRoot.Mnc == nil {
		return "00000"
	}

	return fmt.Sprintf("%03d%02d", *apnRoot.Mcc, *apnRoot.Mnc)
}

//--------------------------------------------------------------------------------//
// Object Base
//--------------------------------------------------------------------------------//

type ObjectBase struct {
	ObjectInterface[ObjectBase] `json:"-" xml:"-"`

	Apn       *string         `json:"apn,omitempty"       xml:"apn,attr,omitempty"`
	Type      *ObjectBaseType `json:"type,omitempty" xml:"type,attr,omitempty"`
	ProfileID *int            `json:"profileID,omitempty" xml:"profile_id,attr,omitempty"`
}

func (apnPointerBase *ObjectBase) Clone() *ObjectBase {
	if apnPointerBase == nil {
		return nil
	}

	return &ObjectBase{
		Apn:       clonePtr(apnPointerBase.Apn),
		Type:      clonePtr(apnPointerBase.Type),
		ProfileID: clonePtr(apnPointerBase.ProfileID),
	}
}

func (apnPointerBase *ObjectBase) Validate() bool {
	return apnPointerBase != nil &&
		(apnPointerBase.Apn != nil || apnPointerBase.Type != nil || apnPointerBase.ProfileID != nil)
}

func (apnPointerBase *ObjectBase) Normalize() {
	if apnPointerBase == nil {
		return
	}

	if apnPointerBase.Apn != nil && apnPointerBase.Type == nil {
		value := ObjectBaseTypeDefault
		apnPointerBase.Type = &value
	}
}

func (apnPointerBase *ObjectBase) IsLike(apnPointer *ObjectBase) bool {
	if apnPointerBase == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return matchStringPtr(apnPointerBase.Apn, apnPointer.Apn) &&
		matchMaskPtr(apnPointerBase.Type, apnPointer.Type) &&
		matchIntPtr(apnPointerBase.ProfileID, apnPointer.ProfileID)
}

//--------------------------------------------------------------------------------//
// Object Auth
//--------------------------------------------------------------------------------//

type ObjectAuth struct {
	ObjectInterface[ObjectAuth] `json:"-" xml:"-"`

	Type     *ObjectAuthType `json:"type,omitempty" xml:"authtype,attr,omitempty"`
	Username *string         `json:"username,omitempty" xml:"user,attr,omitempty"`
	Password *string         `json:"password,omitempty" xml:"password,attr,omitempty"`
}

func (apnPointerAuth *ObjectAuth) Clone() *ObjectAuth {
	if apnPointerAuth == nil {
		return nil
	}

	return &ObjectAuth{
		Type:     clonePtr(apnPointerAuth.Type),
		Username: clonePtr(apnPointerAuth.Username),
		Password: clonePtr(apnPointerAuth.Password),
	}
}

func (apnPointerAuth *ObjectAuth) Validate() bool {
	return apnPointerAuth != nil &&
		apnPointerAuth.Type != nil &&
		(apnPointerAuth.Username != nil || apnPointerAuth.Password != nil)
}

func (apnPointerAuth *ObjectAuth) Normalize() {
	if apnPointerAuth == nil || apnPointerAuth.Type == nil {
		return
	}

	if apnPointerAuth.Username == nil && apnPointerAuth.Password == nil {
		return
	}

	value := ""
	if apnPointerAuth.Username == nil {
		apnPointerAuth.Username = &value
	}

	if apnPointerAuth.Password == nil {
		apnPointerAuth.Password = &value
	}
}

func (apnPointerAuth *ObjectAuth) IsLike(apnPointer *ObjectAuth) bool {
	if apnPointerAuth == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return matchMaskPtr(apnPointerAuth.Type, apnPointer.Type) &&
		matchStringPtr(apnPointerAuth.Username, apnPointer.Username) &&
		matchStringPtr(apnPointerAuth.Password, apnPointer.Password)
}

//--------------------------------------------------------------------------------//
// Object Bearer
//--------------------------------------------------------------------------------//

type ObjectBearer struct {
	ObjectInterface[ObjectBearer] `json:"-" xml:"-"`

	Type        *ObjectBearerProtocol `json:"type,omitempty"         xml:"protocol,attr,omitempty"`
	TypeRoaming *ObjectBearerProtocol `json:"typeRoaming,omitempty"  xml:"roaming_protocol,attr,omitempty"`
	Mtu         *int                  `json:"mtu,omitempty"          xml:"mtu,attr,omitempty"`
	Server      *string               `json:"server,omitempty"       xml:"server,attr,omitempty"`
}

func (apnPointerBearer *ObjectBearer) Clone() *ObjectBearer {
	if apnPointerBearer == nil {
		return nil
	}

	return &ObjectBearer{
		Type:        clonePtr(apnPointerBearer.Type),
		TypeRoaming: clonePtr(apnPointerBearer.TypeRoaming),
		Mtu:         clonePtr(apnPointerBearer.Mtu),
		Server:      clonePtr(apnPointerBearer.Server),
	}
}

func (apnPointerBearer *ObjectBearer) Validate() bool {
	if apnPointerBearer != nil && (apnPointerBearer.Type != nil || apnPointerBearer.TypeRoaming != nil) {
		return true
	}

	return false
}

func (apnPointerBearer *ObjectBearer) Normalize() {}

func (apnPointerBearer *ObjectBearer) IsLike(apnPointer *ObjectBearer) bool {
	if apnPointerBearer == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return matchMaskPtr(apnPointerBearer.Type, apnPointer.Type) &&
		matchMaskPtr(apnPointerBearer.TypeRoaming, apnPointer.TypeRoaming) &&
		matchIntPtr(apnPointerBearer.Mtu, apnPointer.Mtu) &&
		matchStringPtr(apnPointerBearer.Server, apnPointer.Server)
}

//--------------------------------------------------------------------------------//
// Object Proxy
//--------------------------------------------------------------------------------//

type ObjectProxy struct {
	ObjectInterface[ObjectProxy] `json:"-" xml:"-"`

	Server *string `json:"server,omitempty" xml:"proxy,attr,omitempty"`
	Port   *int    `json:"port,omitempty"   xml:"port,attr,omitempty"`
}

func (apnPointerProxy *ObjectProxy) Clone() *ObjectProxy {
	if apnPointerProxy == nil {
		return nil
	}

	return &ObjectProxy{
		Server: clonePtr(apnPointerProxy.Server),
		Port:   clonePtr(apnPointerProxy.Port),
	}
}

func (apnPointerProxy *ObjectProxy) Validate() bool {
	if apnPointerProxy != nil && ((apnPointerProxy.Server != nil && *apnPointerProxy.Server != "") && apnPointerProxy.Port != nil) {
		return true
	}

	return false
}

func (apnPointerProxy *ObjectProxy) Normalize() {}

func (apnPointerProxy *ObjectProxy) IsLike(apnPointer *ObjectProxy) bool {
	if apnPointerProxy == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return matchStringPtr(apnPointerProxy.Server, apnPointer.Server) &&
		matchIntPtr(apnPointerProxy.Port, apnPointer.Port)
}

//--------------------------------------------------------------------------------//
// Object MMS
//--------------------------------------------------------------------------------//

type ObjectMMS struct {
	ObjectInterface[ObjectMMS] `json:"-" xml:"-"`

	Center *string `json:"center,omitempty" xml:"mmsc,attr,omitempty"`
	Server *string `json:"server,omitempty" xml:"mmsproxy,attr,omitempty"`
	Port   *int    `json:"port,omitempty"   xml:"mmsport,attr,omitempty"`
}

func (apnPointerMms *ObjectMMS) Clone() *ObjectMMS {
	if apnPointerMms == nil {
		return nil
	}

	return &ObjectMMS{
		Center: clonePtr(apnPointerMms.Center),
		Server: clonePtr(apnPointerMms.Server),
		Port:   clonePtr(apnPointerMms.Port),
	}
}

func (apnPointerMms *ObjectMMS) Validate() bool {
	if apnPointerMms != nil && ((apnPointerMms.Center != nil && *apnPointerMms.Center != "") || (apnPointerMms.Server != nil && *apnPointerMms.Server != "") || apnPointerMms.Port != nil) {
		return true
	}

	return false
}

func (apnPointerMms *ObjectMMS) Normalize() {}

func (apnPointerMms *ObjectMMS) IsLike(apnPointer *ObjectMMS) bool {
	if apnPointerMms == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return matchStringPtr(apnPointerMms.Center, apnPointer.Center) &&
		matchStringPtr(apnPointerMms.Server, apnPointer.Server) &&
		matchIntPtr(apnPointerMms.Port, apnPointer.Port)
}

//--------------------------------------------------------------------------------//
// Object MVNO
//--------------------------------------------------------------------------------//

type ObjectMVNO struct {
	ObjectInterface[ObjectMVNO] `json:"-" xml:"-"`

	Type *string `json:"type,omitempty" xml:"mvno_type,attr,omitempty"`
	Data *string `json:"data,omitempty" xml:"mvno_match_data,attr,omitempty"`
}

func (apnPointerMvno *ObjectMVNO) Clone() *ObjectMVNO {
	if apnPointerMvno == nil {
		return nil
	}

	return &ObjectMVNO{
		Type: clonePtr(apnPointerMvno.Type),
		Data: clonePtr(apnPointerMvno.Data),
	}
}

func (apnPointerMvno *ObjectMVNO) Validate() bool {
	if apnPointerMvno != nil && (apnPointerMvno.Type != nil || apnPointerMvno.Data != nil) {
		return true
	}

	return false
}

func (apnPointerMvno *ObjectMVNO) Normalize() {}

func (apnPointerMvno *ObjectMVNO) IsLike(apnPointer *ObjectMVNO) bool {
	if apnPointerMvno == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return matchStringPtr(apnPointerMvno.Type, apnPointer.Type) &&
		matchStringPtr(apnPointerMvno.Data, apnPointer.Data)
}

//--------------------------------------------------------------------------------//
// Object Limit
//--------------------------------------------------------------------------------//

type ObjectLimit struct {
	ObjectInterface[ObjectLimit] `json:"-" xml:"-"`

	MaxConn     *int `json:"maxConn,omitempty"      xml:"max_conns,attr,omitempty"`
	MaxConnTime *int `json:"maxConnTime,omitempty"  xml:"max_conns_time,attr,omitempty"`
}

func (apnPointerLimit *ObjectLimit) Clone() *ObjectLimit {
	if apnPointerLimit == nil {
		return nil
	}

	return &ObjectLimit{
		MaxConn:     clonePtr(apnPointerLimit.MaxConn),
		MaxConnTime: clonePtr(apnPointerLimit.MaxConnTime),
	}
}

func (apnPointerLimit *ObjectLimit) Validate() bool {
	if apnPointerLimit != nil && (apnPointerLimit.MaxConn != nil || apnPointerLimit.MaxConnTime != nil) {
		return true
	}

	return false
}

func (apnPointerLimit *ObjectLimit) Normalize() {}

func (apnPointerLimit *ObjectLimit) IsLike(apnPointer *ObjectLimit) bool {
	if apnPointerLimit == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return matchIntPtr(apnPointerLimit.MaxConn, apnPointer.MaxConn) &&
		matchIntPtr(apnPointerLimit.MaxConnTime, apnPointer.MaxConnTime)
}

//--------------------------------------------------------------------------------//
// Object Other
//--------------------------------------------------------------------------------//

type ObjectOther struct {
	ObjectInterface[ObjectOther] `json:"-" xml:"-"`

	NetworkTypeBitmask *ObjectNetworkType `json:"networkTypeBitmask,omitempty" xml:"network_type_bitmask,attr,omitempty"`
	ModemCognitive     *bool              `json:"modemCognitive,omitempty" xml:"modem_cognitive,attr,omitempty"`
	CarrierEnabled     *bool              `json:"IsEnabled,omitempty" xml:"carrier_enabled,attr,omitempty"`
	UserVisible        *bool              `json:"IsVisible,omitempty" xml:"user_visible,attr,omitempty"`
	UserEditable       *bool              `json:"IsEditable,omitempty" xml:"user_editable,attr,omitempty"`
}

func (apnPointerOther *ObjectOther) Clone() *ObjectOther {
	if apnPointerOther == nil {
		return nil
	}

	return &ObjectOther{
		NetworkTypeBitmask: clonePtr(apnPointerOther.NetworkTypeBitmask),
		ModemCognitive:     clonePtr(apnPointerOther.ModemCognitive),
		CarrierEnabled:     clonePtr(apnPointerOther.CarrierEnabled),
		UserVisible:        clonePtr(apnPointerOther.UserVisible),
		UserEditable:       clonePtr(apnPointerOther.UserEditable),
	}
}

func (apnPointerOther *ObjectOther) Validate() bool {
	if apnPointerOther != nil && (apnPointerOther.NetworkTypeBitmask != nil || apnPointerOther.ModemCognitive != nil || apnPointerOther.CarrierEnabled != nil || apnPointerOther.UserVisible != nil || apnPointerOther.UserEditable != nil) {
		return true
	}

	return false
}

func (apnPointerOther *ObjectOther) Normalize() {}

func (apnPointerOther *ObjectOther) IsLike(apnPointer *ObjectOther) bool {
	if apnPointerOther == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return matchMaskPtr(apnPointerOther.NetworkTypeBitmask, apnPointer.NetworkTypeBitmask)
}

//--------------------------------------------------------------------------------//
