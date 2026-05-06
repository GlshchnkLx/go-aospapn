package apnxml

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

//--------------------------------------------------------------------------------//
// Object Update
//--------------------------------------------------------------------------------//

type ObjectUpdateMode int

const (
	ObjectUpdateMerge ObjectUpdateMode = iota
	ObjectUpdatePatch
	ObjectUpdateApply
)

func cloneObjectFields[Type any](source *Type) *Type {
	if source == nil {
		return nil
	}

	target := new(Type)
	updateObjectFields(target, source, ObjectUpdateApply)
	return target
}

func hasObjectFields[Type any](source *Type) bool {
	if source == nil {
		return false
	}

	sourceValue := reflect.ValueOf(source)
	if sourceValue.Kind() != reflect.Ptr || sourceValue.IsNil() {
		return false
	}

	sourceValue = sourceValue.Elem()
	sourceType := sourceValue.Type()

	for fieldIndex := 0; fieldIndex < sourceValue.NumField(); fieldIndex++ {
		if sourceType.Field(fieldIndex).IsExported() && !sourceValue.Field(fieldIndex).IsZero() {
			return true
		}
	}

	return false
}

func updateObjectFields[Type any](target *Type, source *Type, mode ObjectUpdateMode) bool {
	if target == nil || source == nil {
		return false
	}

	targetValue := reflect.ValueOf(target)
	sourceValue := reflect.ValueOf(source)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() || sourceValue.Kind() != reflect.Ptr || sourceValue.IsNil() {
		return false
	}

	targetValue = targetValue.Elem()
	sourceValue = sourceValue.Elem()
	if targetValue.Type() != sourceValue.Type() {
		return false
	}

	targetType := targetValue.Type()
	for fieldIndex := 0; fieldIndex < targetValue.NumField(); fieldIndex++ {
		if !targetType.Field(fieldIndex).IsExported() {
			continue
		}

		targetField := targetValue.Field(fieldIndex)
		sourceField := sourceValue.Field(fieldIndex)
		if !targetField.CanSet() {
			continue
		}

		switch mode {
		case ObjectUpdateMerge:
			if targetField.IsZero() && !sourceField.IsZero() {
				setClonedField(targetField, sourceField)
			}
		case ObjectUpdatePatch:
			if !sourceField.IsZero() {
				setClonedField(targetField, sourceField)
			}
		case ObjectUpdateApply:
			setClonedField(targetField, sourceField)
		}
	}

	return true
}

func updateObjectPointer[Type any](target **Type, source *Type, mode ObjectUpdateMode) {
	if target == nil {
		return
	}

	switch mode {
	case ObjectUpdateMerge:
		if *target == nil {
			*target = cloneObjectFields(source)
		} else {
			updateObjectFields(*target, source, mode)
		}
	case ObjectUpdatePatch:
		if source == nil {
			return
		}
		if *target == nil {
			*target = cloneObjectFields(source)
		} else {
			updateObjectFields(*target, source, mode)
		}
	case ObjectUpdateApply:
		*target = cloneObjectFields(source)
	}
}

func setClonedField(target reflect.Value, source reflect.Value) {
	if !source.IsValid() {
		target.Set(reflect.Zero(target.Type()))
		return
	}

	if source.Kind() != reflect.Ptr {
		target.Set(source)
		return
	}

	if source.IsNil() {
		target.Set(reflect.Zero(target.Type()))
		return
	}

	targetPointer := reflect.New(source.Type().Elem())
	targetPointer.Elem().Set(source.Elem())
	target.Set(targetPointer)
}

//--------------------------------------------------------------------------------//
// Object Core & Helper
//--------------------------------------------------------------------------------//

type Object struct {
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
	if apnPointerClone == nil {
		return nil
	}

	apnPointerClone.Normalize()
	return apnPointerClone
}

func (apnPointerCore *Object) Update(source *Object, mode ObjectUpdateMode) bool {
	if apnPointerCore == nil || source == nil {
		return false
	}

	updateObjectPointer(&apnPointerCore.ObjectRoot, source.ObjectRoot, mode)
	updateObjectPointer(&apnPointerCore.Base, source.Base, mode)
	updateObjectPointer(&apnPointerCore.Auth, source.Auth, mode)
	updateObjectPointer(&apnPointerCore.Bearer, source.Bearer, mode)
	updateObjectPointer(&apnPointerCore.Proxy, source.Proxy, mode)
	updateObjectPointer(&apnPointerCore.Mms, source.Mms, mode)
	updateObjectPointer(&apnPointerCore.Mvno, source.Mvno, mode)
	updateObjectPointer(&apnPointerCore.Limit, source.Limit, mode)
	updateObjectPointer(&apnPointerCore.Other, source.Other, mode)

	if mode == ObjectUpdateApply {
		apnPointerCore.GroupMapByType = nil
		if source.GroupMapByType != nil {
			apnPointerCore.GroupMapByType = map[ObjectBaseType]*Object{}
			for apnType, apnPointer := range source.GroupMapByType {
				apnPointerCore.GroupMapByType[apnType] = apnPointer.Clone()
			}
		}
	} else if source.GroupMapByType != nil {
		if apnPointerCore.GroupMapByType == nil {
			apnPointerCore.GroupMapByType = map[ObjectBaseType]*Object{}
		}
		for apnType, apnPointer := range source.GroupMapByType {
			if apnPointerCore.GroupMapByType[apnType] == nil {
				apnPointerCore.GroupMapByType[apnType] = apnPointer.Clone()
			} else {
				apnPointerCore.GroupMapByType[apnType].Update(apnPointer, mode)
			}
		}
	}

	return true
}

func (apnPointerCore *Object) Merge(source *Object) bool {
	return apnPointerCore.Update(source, ObjectUpdateMerge)
}

func (apnPointerCore *Object) Patch(source *Object) bool {
	return apnPointerCore.Update(source, ObjectUpdatePatch)
}

func (apnPointerCore *Object) Apply(source *Object) bool {
	return apnPointerCore.Update(source, ObjectUpdateApply)
}

func (apnPointerCore *Object) GetMatchPointer(apnPointerQuery *Object) *Object {
	if apnPointerCore == nil || apnPointerQuery == nil {
		if apnPointerQuery == nil {
			return apnPointerCore
		} else {
			return nil
		}
	}

	if apnPointerCore.ObjectRoot != nil {
		if !apnPointerCore.ObjectRoot.Match(apnPointerQuery.ObjectRoot) {
			return nil
		}
	}

	if apnPointerCore.GroupMapByType != nil {
		for _, apnPointer := range apnPointerCore.GroupMapByType {
			apnPointer = apnPointer.GetMatchPointer(apnPointerQuery)

			if apnPointer != nil {
				return apnPointer
			}
		}

		return nil
	}

	if apnPointerCore.Base.Match(apnPointerQuery.Base) &&
		apnPointerCore.Auth.Match(apnPointerQuery.Auth) &&
		apnPointerCore.Bearer.Match(apnPointerQuery.Bearer) &&
		apnPointerCore.Proxy.Match(apnPointerQuery.Proxy) &&
		apnPointerCore.Mms.Match(apnPointerQuery.Mms) &&
		apnPointerCore.Mvno.Match(apnPointerQuery.Mvno) &&
		apnPointerCore.Limit.Match(apnPointerQuery.Limit) &&
		apnPointerCore.Other.Match(apnPointerQuery.Other) {

		return apnPointerCore
	}

	return nil
}

func (apnPointerCore *Object) Match(apnPointerQuery *Object) bool {
	return apnPointerCore.GetMatchPointer(apnPointerQuery) != nil
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
	Carrier   string `json:"carrierName" xml:"carrier,attr,omitempty"`
	CarrierID *int   `json:"carrierID,omitempty"   xml:"carrier_id,attr,omitempty"`
	Mcc       *int   `json:"mcc,omitempty"         xml:"mcc,attr,omitempty"`
	Mnc       *int   `json:"mnc,omitempty"         xml:"mnc,attr,omitempty"`
}

func (apnPointerRoot *ObjectRoot) Clone() *ObjectRoot {
	return cloneObjectFields(apnPointerRoot)
}

func (apnPointerRoot *ObjectRoot) Update(source *ObjectRoot, mode ObjectUpdateMode) bool {
	return updateObjectFields(apnPointerRoot, source, mode)
}

func (apnPointerRoot *ObjectRoot) Merge(source *ObjectRoot) bool {
	return apnPointerRoot.Update(source, ObjectUpdateMerge)
}

func (apnPointerRoot *ObjectRoot) Patch(source *ObjectRoot) bool {
	return apnPointerRoot.Update(source, ObjectUpdatePatch)
}

func (apnPointerRoot *ObjectRoot) Apply(source *ObjectRoot) bool {
	return apnPointerRoot.Update(source, ObjectUpdateApply)
}

func (apnPointerRoot *ObjectRoot) Validate() bool {
	if apnPointerRoot != nil && (apnPointerRoot.Mcc != nil && apnPointerRoot.Mnc != nil) {
		return true
	}

	return false
}

func (apnPointerRoot *ObjectRoot) Match(apnPointer *ObjectRoot) bool {
	if apnPointerRoot == nil || apnPointer == nil {
		return apnPointer == nil
	}

	var (
		isMatchCarrierID = true
		isMatchPlmn      = true
	)

	if apnPointer.CarrierID != nil {
		if apnPointerRoot.CarrierID == nil {
			return false
		}

		isMatchCarrierID = *apnPointerRoot.CarrierID == *apnPointer.CarrierID
	}

	if apnPointer.Mcc != nil && apnPointer.Mnc != nil {
		if apnPointerRoot.Mcc == nil || apnPointerRoot.Mnc == nil {
			return false
		}

		isMatchPlmn = (*apnPointerRoot.Mcc == *apnPointer.Mcc) && (*apnPointerRoot.Mnc == *apnPointer.Mnc)
	}

	return matchString(apnPointerRoot.Carrier, apnPointer.Carrier) && isMatchCarrierID && isMatchPlmn
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
	Apn       *string         `json:"apn,omitempty"       xml:"apn,attr,omitempty"`
	Type      *ObjectBaseType `json:"type,omitempty" xml:"type,attr,omitempty"`
	ProfileID *int            `json:"profileID,omitempty" xml:"profile_id,attr,omitempty"`
}

func (apnPointerBase *ObjectBase) Clone() *ObjectBase {
	return cloneObjectFields(apnPointerBase)
}

func (apnPointerBase *ObjectBase) Update(source *ObjectBase, mode ObjectUpdateMode) bool {
	return updateObjectFields(apnPointerBase, source, mode)
}

func (apnPointerBase *ObjectBase) Merge(source *ObjectBase) bool {
	return apnPointerBase.Update(source, ObjectUpdateMerge)
}

func (apnPointerBase *ObjectBase) Patch(source *ObjectBase) bool {
	return apnPointerBase.Update(source, ObjectUpdatePatch)
}

func (apnPointerBase *ObjectBase) Apply(source *ObjectBase) bool {
	return apnPointerBase.Update(source, ObjectUpdateApply)
}

func (apnPointerBase *ObjectBase) Validate() bool {
	return hasObjectFields(apnPointerBase)
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

func (apnPointerBase *ObjectBase) Match(apnPointer *ObjectBase) bool {
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
	Type     *ObjectAuthType `json:"type,omitempty" xml:"authtype,attr,omitempty"`
	Username *string         `json:"username,omitempty" xml:"user,attr,omitempty"`
	Password *string         `json:"password,omitempty" xml:"password,attr,omitempty"`
}

func (apnPointerAuth *ObjectAuth) Clone() *ObjectAuth {
	return cloneObjectFields(apnPointerAuth)
}

func (apnPointerAuth *ObjectAuth) Update(source *ObjectAuth, mode ObjectUpdateMode) bool {
	return updateObjectFields(apnPointerAuth, source, mode)
}

func (apnPointerAuth *ObjectAuth) Merge(source *ObjectAuth) bool {
	return apnPointerAuth.Update(source, ObjectUpdateMerge)
}

func (apnPointerAuth *ObjectAuth) Patch(source *ObjectAuth) bool {
	return apnPointerAuth.Update(source, ObjectUpdatePatch)
}

func (apnPointerAuth *ObjectAuth) Apply(source *ObjectAuth) bool {
	return apnPointerAuth.Update(source, ObjectUpdateApply)
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

func (apnPointerAuth *ObjectAuth) Match(apnPointer *ObjectAuth) bool {
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
	Type        *ObjectBearerProtocol `json:"type,omitempty"         xml:"protocol,attr,omitempty"`
	TypeRoaming *ObjectBearerProtocol `json:"typeRoaming,omitempty"  xml:"roaming_protocol,attr,omitempty"`
	Mtu         *int                  `json:"mtu,omitempty"          xml:"mtu,attr,omitempty"`
	Server      *string               `json:"server,omitempty"       xml:"server,attr,omitempty"`
}

func (apnPointerBearer *ObjectBearer) Clone() *ObjectBearer {
	return cloneObjectFields(apnPointerBearer)
}

func (apnPointerBearer *ObjectBearer) Update(source *ObjectBearer, mode ObjectUpdateMode) bool {
	return updateObjectFields(apnPointerBearer, source, mode)
}

func (apnPointerBearer *ObjectBearer) Merge(source *ObjectBearer) bool {
	return apnPointerBearer.Update(source, ObjectUpdateMerge)
}

func (apnPointerBearer *ObjectBearer) Patch(source *ObjectBearer) bool {
	return apnPointerBearer.Update(source, ObjectUpdatePatch)
}

func (apnPointerBearer *ObjectBearer) Apply(source *ObjectBearer) bool {
	return apnPointerBearer.Update(source, ObjectUpdateApply)
}

func (apnPointerBearer *ObjectBearer) Validate() bool {
	if apnPointerBearer != nil && (apnPointerBearer.Type != nil || apnPointerBearer.TypeRoaming != nil) {
		return true
	}

	return false
}

func (apnPointerBearer *ObjectBearer) Normalize() {}

func (apnPointerBearer *ObjectBearer) Match(apnPointer *ObjectBearer) bool {
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
	Server *string `json:"server,omitempty" xml:"proxy,attr,omitempty"`
	Port   *int    `json:"port,omitempty"   xml:"port,attr,omitempty"`
}

func (apnPointerProxy *ObjectProxy) Clone() *ObjectProxy {
	return cloneObjectFields(apnPointerProxy)
}

func (apnPointerProxy *ObjectProxy) Update(source *ObjectProxy, mode ObjectUpdateMode) bool {
	return updateObjectFields(apnPointerProxy, source, mode)
}

func (apnPointerProxy *ObjectProxy) Merge(source *ObjectProxy) bool {
	return apnPointerProxy.Update(source, ObjectUpdateMerge)
}

func (apnPointerProxy *ObjectProxy) Patch(source *ObjectProxy) bool {
	return apnPointerProxy.Update(source, ObjectUpdatePatch)
}

func (apnPointerProxy *ObjectProxy) Apply(source *ObjectProxy) bool {
	return apnPointerProxy.Update(source, ObjectUpdateApply)
}

func (apnPointerProxy *ObjectProxy) Validate() bool {
	if apnPointerProxy != nil && ((apnPointerProxy.Server != nil && *apnPointerProxy.Server != "") && apnPointerProxy.Port != nil) {
		return true
	}

	return false
}

func (apnPointerProxy *ObjectProxy) Normalize() {}

func (apnPointerProxy *ObjectProxy) Match(apnPointer *ObjectProxy) bool {
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
	Center *string `json:"center,omitempty" xml:"mmsc,attr,omitempty"`
	Server *string `json:"server,omitempty" xml:"mmsproxy,attr,omitempty"`
	Port   *int    `json:"port,omitempty"   xml:"mmsport,attr,omitempty"`
}

func (apnPointerMms *ObjectMMS) Clone() *ObjectMMS {
	return cloneObjectFields(apnPointerMms)
}

func (apnPointerMms *ObjectMMS) Update(source *ObjectMMS, mode ObjectUpdateMode) bool {
	return updateObjectFields(apnPointerMms, source, mode)
}

func (apnPointerMms *ObjectMMS) Merge(source *ObjectMMS) bool {
	return apnPointerMms.Update(source, ObjectUpdateMerge)
}

func (apnPointerMms *ObjectMMS) Patch(source *ObjectMMS) bool {
	return apnPointerMms.Update(source, ObjectUpdatePatch)
}

func (apnPointerMms *ObjectMMS) Apply(source *ObjectMMS) bool {
	return apnPointerMms.Update(source, ObjectUpdateApply)
}

func (apnPointerMms *ObjectMMS) Validate() bool {
	if apnPointerMms != nil && ((apnPointerMms.Center != nil && *apnPointerMms.Center != "") || (apnPointerMms.Server != nil && *apnPointerMms.Server != "") || apnPointerMms.Port != nil) {
		return true
	}

	return false
}

func (apnPointerMms *ObjectMMS) Normalize() {}

func (apnPointerMms *ObjectMMS) Match(apnPointer *ObjectMMS) bool {
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
	Type *string `json:"type,omitempty" xml:"mvno_type,attr,omitempty"`
	Data *string `json:"data,omitempty" xml:"mvno_match_data,attr,omitempty"`
}

func (apnPointerMvno *ObjectMVNO) Clone() *ObjectMVNO {
	return cloneObjectFields(apnPointerMvno)
}

func (apnPointerMvno *ObjectMVNO) Update(source *ObjectMVNO, mode ObjectUpdateMode) bool {
	return updateObjectFields(apnPointerMvno, source, mode)
}

func (apnPointerMvno *ObjectMVNO) Merge(source *ObjectMVNO) bool {
	return apnPointerMvno.Update(source, ObjectUpdateMerge)
}

func (apnPointerMvno *ObjectMVNO) Patch(source *ObjectMVNO) bool {
	return apnPointerMvno.Update(source, ObjectUpdatePatch)
}

func (apnPointerMvno *ObjectMVNO) Apply(source *ObjectMVNO) bool {
	return apnPointerMvno.Update(source, ObjectUpdateApply)
}

func (apnPointerMvno *ObjectMVNO) Validate() bool {
	if apnPointerMvno != nil && (apnPointerMvno.Type != nil || apnPointerMvno.Data != nil) {
		return true
	}

	return false
}

func (apnPointerMvno *ObjectMVNO) Normalize() {}

func (apnPointerMvno *ObjectMVNO) Match(apnPointer *ObjectMVNO) bool {
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
	MaxConn     *int `json:"maxConn,omitempty"      xml:"max_conns,attr,omitempty"`
	MaxConnTime *int `json:"maxConnTime,omitempty"  xml:"max_conns_time,attr,omitempty"`
}

func (apnPointerLimit *ObjectLimit) Clone() *ObjectLimit {
	return cloneObjectFields(apnPointerLimit)
}

func (apnPointerLimit *ObjectLimit) Update(source *ObjectLimit, mode ObjectUpdateMode) bool {
	return updateObjectFields(apnPointerLimit, source, mode)
}

func (apnPointerLimit *ObjectLimit) Merge(source *ObjectLimit) bool {
	return apnPointerLimit.Update(source, ObjectUpdateMerge)
}

func (apnPointerLimit *ObjectLimit) Patch(source *ObjectLimit) bool {
	return apnPointerLimit.Update(source, ObjectUpdatePatch)
}

func (apnPointerLimit *ObjectLimit) Apply(source *ObjectLimit) bool {
	return apnPointerLimit.Update(source, ObjectUpdateApply)
}

func (apnPointerLimit *ObjectLimit) Validate() bool {
	if apnPointerLimit != nil && (apnPointerLimit.MaxConn != nil || apnPointerLimit.MaxConnTime != nil) {
		return true
	}

	return false
}

func (apnPointerLimit *ObjectLimit) Normalize() {}

func (apnPointerLimit *ObjectLimit) Match(apnPointer *ObjectLimit) bool {
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
	NetworkTypeBitmask *ObjectNetworkType `json:"networkTypeBitmask,omitempty" xml:"network_type_bitmask,attr,omitempty"`
	ModemCognitive     *bool              `json:"modemCognitive,omitempty" xml:"modem_cognitive,attr,omitempty"`
	CarrierEnabled     *bool              `json:"IsEnabled,omitempty" xml:"carrier_enabled,attr,omitempty"`
	UserVisible        *bool              `json:"IsVisible,omitempty" xml:"user_visible,attr,omitempty"`
	UserEditable       *bool              `json:"IsEditable,omitempty" xml:"user_editable,attr,omitempty"`
}

func (apnPointerOther *ObjectOther) Clone() *ObjectOther {
	return cloneObjectFields(apnPointerOther)
}

func (apnPointerOther *ObjectOther) Update(source *ObjectOther, mode ObjectUpdateMode) bool {
	return updateObjectFields(apnPointerOther, source, mode)
}

func (apnPointerOther *ObjectOther) Merge(source *ObjectOther) bool {
	return apnPointerOther.Update(source, ObjectUpdateMerge)
}

func (apnPointerOther *ObjectOther) Patch(source *ObjectOther) bool {
	return apnPointerOther.Update(source, ObjectUpdatePatch)
}

func (apnPointerOther *ObjectOther) Apply(source *ObjectOther) bool {
	return apnPointerOther.Update(source, ObjectUpdateApply)
}

func (apnPointerOther *ObjectOther) Validate() bool {
	if apnPointerOther != nil && (apnPointerOther.NetworkTypeBitmask != nil || apnPointerOther.ModemCognitive != nil || apnPointerOther.CarrierEnabled != nil || apnPointerOther.UserVisible != nil || apnPointerOther.UserEditable != nil) {
		return true
	}

	return false
}

func (apnPointerOther *ObjectOther) Normalize() {}

func (apnPointerOther *ObjectOther) Match(apnPointer *ObjectOther) bool {
	if apnPointerOther == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return matchMaskPtr(apnPointerOther.NetworkTypeBitmask, apnPointer.NetworkTypeBitmask)
}

//--------------------------------------------------------------------------------//
