// # APN Object
//
// File defines structures and utilities for handling APN (Access Point Name)
// configuration data commonly used in mobile network settings. It supports XML marshaling
// and unmarshaling, validation, and cloning of APN configurations.
//
// The package defines a modular APN object composed of several objects:
//   - Root: Carrier, MCC, MNC identification
//   - Base: APN name, type, profile ID
//   - Auth: Authentication type, username, password
//   - Bearer: Protocol, roaming protocol, MTU, server
//   - Proxy: Proxy server and port
//   - MMS: MMSC, proxy, and port
//   - MVNO: MVNO type and match data
//   - Limit: Connection limits
//   - Other: Miscellaneous flags and settings
//
// Each object implements the APNObjectInterface, supporting Clone and Validate methods.
// The main APNObject type aggregates all objects and handles XML serialization via
// a helper struct to manage optional fields correctly.
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

// APNObjectInterface defines the contract for APN configuration objects.
// Each object must implement Clone to produce a deep copy and Validate to check
// if the object contains sufficient or valid data.
type APNObjectInterface[Type any] interface {
	Clone() Type
	Validate() bool
	IsLike(Type) bool
}

//--------------------------------------------------------------------------------//
// APNObject Core & Helper
//--------------------------------------------------------------------------------//

// APNObject represents a complete APN configuration, composed of optional sub-objects.
// It embeds APNObjectRoot and may include Base, Auth, Bearer, Proxy, Mms, Mvno,
// Limit, and Other configuration blocks. It also maintains a map for grouping by type.
//
// Implements APNObjectInterface[APNObject] and custom XML marshaling/unmarshaling.
type APNObject struct {
	APNObjectInterface[APNObject] `json:"-" xml:"-"`

	*APNObjectRoot
	Base   *APNObjectBase   `json:"base,omitempty"`
	Auth   *APNObjectAuth   `json:"auth,omitempty"`
	Bearer *APNObjectBearer `json:"bearer,omitempty"`
	Proxy  *APNObjectProxy  `json:"proxy,omitempty"`
	Mms    *APNObjectMms    `json:"mms,omitempty"`
	Mvno   *APNObjectMvno   `json:"mvno,omitempty"`
	Limit  *APNObjectLimit  `json:"limit,omitempty"`
	Other  *APNObjectOther  `json:"other,omitempty"`

	GroupMapByType map[APNTypeBaseType]*APNObject `json:"groupMap,omitempty"`
}

// helperAPNObject is an internal helper struct used to correctly marshal/unmarshal
// optional embedded fields in APNObject via xml.Encoder/Decoder.
// All fields are marked with `xml:",omitempty"` to omit empty values in XML output.
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

// Clone creates a deep copy of the APNObject, cloning all its sub-objects.
func (apnPointerCore *APNObject) Clone() *APNObject {
	apnObject := APNObject{
		APNObjectRoot: helperApnPointerClone(apnPointerCore.APNObjectRoot),
		Base:          helperApnPointerClone(apnPointerCore.Base),
		Auth:          helperApnPointerClone(apnPointerCore.Auth),
		Bearer:        helperApnPointerClone(apnPointerCore.Bearer),
		Proxy:         helperApnPointerClone(apnPointerCore.Proxy),
		Mms:           helperApnPointerClone(apnPointerCore.Mms),
		Mvno:          helperApnPointerClone(apnPointerCore.Mvno),
		Limit:         helperApnPointerClone(apnPointerCore.Limit),
		Other:         helperApnPointerClone(apnPointerCore.Other),
	}

	if apnPointerCore.GroupMapByType != nil {
		apnObject.GroupMapByType = map[APNTypeBaseType]*APNObject{}

		for apnPointerBaseTypeString, apnPointer := range apnPointerCore.GroupMapByType {
			apnObject.GroupMapByType[apnPointerBaseTypeString] = apnPointer.Clone()
		}
	}

	return &apnObject
}

// Validate checks whether the APNObject is valid by delegating to its Root object.
// A valid APN must have non-nil Mcc and Mnc in its root.
func (apnObjectCore APNObject) Validate() bool {
	return apnObjectCore.APNObjectRoot.Validate()
}

// GetIsLikePointer returns the receiver if it matches the query APNObject,
// or a matching entry from GroupMapByType if present. Returns nil if no match.
func (apnPointerCore *APNObject) GetIsLikePointer(apnPointerQuery *APNObject) *APNObject {
	if apnPointerCore == nil || apnPointerQuery == nil {
		if apnPointerQuery == nil {
			return apnPointerCore
		} else {
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

	if apnPointerCore.APNObjectRoot != nil {
		if !apnPointerCore.APNObjectRoot.IsLike(apnPointerQuery.APNObjectRoot) {
			return nil
		}
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

// IsLike returns true if the receiver matches the query APNObject.
func (apnPointerCore *APNObject) IsLike(apnPointerQuery *APNObject) bool {
	return apnPointerCore.GetIsLikePointer(apnPointerQuery) != nil
}

// String returns the JSON representation of the APNObject, indented for readability.
func (apnObjectCore APNObject) String() string {
	jsonData, err := json.MarshalIndent(apnObjectCore, "", "\t")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return string(jsonData)
}

// MarshalXML implements custom XML marshaling for APNObject.
// It uses helperAPNObject to ensure optional fields are omitted when nil.
func (apnObjectCore APNObject) MarshalXML(xmlEncoder *xml.Encoder, xmlStart xml.StartElement) error {
	_apnObject := helperAPNObject{
		APNObjectRoot:   helperApnPointerClone(apnObjectCore.APNObjectRoot),
		APNObjectBase:   helperApnPointerClone(apnObjectCore.Base),
		APNObjectAuth:   helperApnPointerClone(apnObjectCore.Auth),
		APNObjectBearer: helperApnPointerClone(apnObjectCore.Bearer),
		APNObjectProxy:  helperApnPointerClone(apnObjectCore.Proxy),
		APNObjectMms:    helperApnPointerClone(apnObjectCore.Mms),
		APNObjectMvno:   helperApnPointerClone(apnObjectCore.Mvno),
		APNObjectLimit:  helperApnPointerClone(apnObjectCore.Limit),
		APNObjectOther:  helperApnPointerClone(apnObjectCore.Other),
	}

	err := xmlEncoder.EncodeElement(_apnObject, xmlStart)
	if err != nil {
		return err
	}

	return nil
}

// UnmarshalXML implements custom XML unmarshaling for APNObject.
// It decodes into a helperAPNObject and then assigns cloned pointers to the receiver.
func (apnPointerCore *APNObject) UnmarshalXML(xmlDecoder *xml.Decoder, xmlStart xml.StartElement) error {
	var _apnObject helperAPNObject

	err := xmlDecoder.DecodeElement(&_apnObject, &xmlStart)
	if err != nil {
		return err
	}

	apnPointerCore.APNObjectRoot = helperApnPointerClone(_apnObject.APNObjectRoot)
	apnPointerCore.Base = helperApnPointerClone(_apnObject.APNObjectBase)
	apnPointerCore.Auth = helperApnPointerClone(_apnObject.APNObjectAuth)
	apnPointerCore.Bearer = helperApnPointerClone(_apnObject.APNObjectBearer)
	apnPointerCore.Proxy = helperApnPointerClone(_apnObject.APNObjectProxy)
	apnPointerCore.Mms = helperApnPointerClone(_apnObject.APNObjectMms)
	apnPointerCore.Mvno = helperApnPointerClone(_apnObject.APNObjectMvno)
	apnPointerCore.Limit = helperApnPointerClone(_apnObject.APNObjectLimit)
	apnPointerCore.Other = helperApnPointerClone(_apnObject.APNObjectOther)

	return nil
}

//--------------------------------------------------------------------------------//
// APN Object Root
//--------------------------------------------------------------------------------//

// APNObjectRoot contains the core identifying fields of an APN: carrier name, carrier ID, MCC, and MNC.
// It is the minimal required object for a valid APN configuration.
type APNObjectRoot struct {
	APNObjectInterface[APNObjectRoot] `json:"-" xml:"-"`

	Carrier   string `json:"carrierName" xml:"carrier,attr,omitempty"`
	CarrierID *int   `json:"carrierID,omitempty"   xml:"carrier_id,attr,omitempty"`
	Mcc       *int   `json:"mcc,omitempty"         xml:"mcc,attr,omitempty"`
	Mnc       *int   `json:"mnc,omitempty"         xml:"mnc,attr,omitempty"`
}

// Clone creates a deep copy of APNObjectRoot.
// Returns nil if the root is invalid (missing MCC/MNC).
func (apnPointerRoot *APNObjectRoot) Clone() *APNObjectRoot {
	if apnPointerRoot.Validate() {
		return &APNObjectRoot{
			Carrier:   apnPointerRoot.Carrier,
			CarrierID: helperClonePointer(apnPointerRoot.CarrierID),
			Mcc:       helperClonePointer(apnPointerRoot.Mcc),
			Mnc:       helperClonePointer(apnPointerRoot.Mnc),
		}
	}

	return nil
}

// Validate checks if APNObjectRoot has valid MCC and MNC fields.
// Required for the APN to be considered minimally valid.
func (apnPointerRoot *APNObjectRoot) Validate() bool {
	if apnPointerRoot != nil && (apnPointerRoot.Mcc != nil && apnPointerRoot.Mnc != nil) {
		return true
	}

	return false
}

// IsLike returns true if the receiver matches the given APNObjectRoot.
// Matching includes Carrier (ignoring common keywords), CarrierID, and PLMN (MCC+MNC).
func (apnPointerRoot *APNObjectRoot) IsLike(apnPointer *APNObjectRoot) bool {
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

	return helperIsLikeString(apnPointerRoot.Carrier, apnPointer.Carrier) && isLikeCarrierID && isLikePlmn
}

// apnRootCarrierWordMask is a set of common carrier-related keywords to be filtered
// out when cleaning up the carrier name via GetCarrier().
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

// GetCarrier returns a cleaned-up version of the carrier name by removing common
// technical or generic keywords found in apnRootCarrierWordMask.
// Splits the name by common separators and filters out masked words.
func (apnRoot APNObjectRoot) GetCarrier() string {
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

// GetID returns a unique identifier string for the APN, combining CarrierID (if present)
// and PLMN (Public Land Mobile Network) code in the format "CID:X;PLMN:YYYYY;".
func (apnRoot APNObjectRoot) GetID() string {
	var (
		apnRootID string
	)

	if apnRoot.CarrierID != nil {
		apnRootID += fmt.Sprintf("CID:%d;", *apnRoot.CarrierID)
	}

	apnRootID += fmt.Sprintf("PLMN:%s;", apnRoot.GetPLMN())

	return apnRootID
}

// GetPLMN returns the PLMN (Public Land Mobile Network) code as a 5-digit string
// formatted as "MMCCNN" (3-digit MCC + 2-digit MNC). Returns "00000" if MCC or MNC is nil.
func (apnRoot APNObjectRoot) GetPLMN() string {
	if apnRoot.Mcc == nil || apnRoot.Mnc == nil {
		return "00000"
	}

	return fmt.Sprintf("%03d%02d", *apnRoot.Mcc, *apnRoot.Mnc)
}

//--------------------------------------------------------------------------------//
// APN Object Base
//--------------------------------------------------------------------------------//

// APNObjectBase contains basic APN settings: the APN string, connection type, and profile ID.
type APNObjectBase struct {
	APNObjectInterface[APNObjectBase] `json:"-" xml:"-"`

	Apn       *string          `json:"apn,omitempty"       xml:"apn,attr,omitempty"`
	Type      *APNTypeBaseType `json:"type,omitempty" xml:"type,attr,omitempty"`
	ProfileID *int             `json:"profileID,omitempty" xml:"profile_id,attr,omitempty"`
}

// Clone creates a deep copy of APNObjectBase.
// Returns nil if the base is invalid (all fields nil).
func (apnPointerBase *APNObjectBase) Clone() *APNObjectBase {
	if apnPointerBase.Validate() {
		return &APNObjectBase{
			Apn:       helperClonePointer(apnPointerBase.Apn),
			Type:      helperClonePointer(apnPointerBase.Type),
			ProfileID: helperClonePointer(apnPointerBase.ProfileID),
		}
	}

	return nil
}

// Validate checks if APNObjectBase has at least one non-nil field.
// If Apn is set but Type is nil, it auto-assigns APNTYPE_BASE_TYPE_DEFAULT.
func (apnPointerBase *APNObjectBase) Validate() bool {
	if apnPointerBase != nil {
		if apnPointerBase.Apn != nil && apnPointerBase.Type == nil {
			var value = APNTYPE_BASE_TYPE_DEFAULT
			apnPointerBase.Type = &value
		}

		if apnPointerBase.Apn != nil || apnPointerBase.Type != nil || apnPointerBase.ProfileID != nil {
			return true
		}
	}

	return false
}

// IsLike returns true if the receiver matches the given APNObjectBase.
func (apnPointerBase *APNObjectBase) IsLike(apnPointer *APNObjectBase) bool {
	if apnPointerBase == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return helperIsLikeStringPointer(apnPointerBase.Apn, apnPointer.Apn) &&
		helperIsLikeMaskPointer(apnPointerBase.Type, apnPointer.Type) &&
		helperIsLikeIntPointer(apnPointerBase.ProfileID, apnPointer.ProfileID)
}

//--------------------------------------------------------------------------------//
// APN Object Auth
//--------------------------------------------------------------------------------//

// APNObjectAuth contains authentication settings: type, username, and password.
type APNObjectAuth struct {
	APNObjectInterface[APNObjectAuth] `json:"-" xml:"-"`

	Type     *APNTypeAuthType `json:"type,omitempty" xml:"authtype,attr,omitempty"`
	Username *string          `json:"username,omitempty" xml:"user,attr,omitempty"`
	Password *string          `json:"password,omitempty" xml:"password,attr,omitempty"`
}

// Clone creates a deep copy of APNObjectAuth.
// Returns nil if validation fails.
func (apnPointerAuth *APNObjectAuth) Clone() *APNObjectAuth {
	if apnPointerAuth.Validate() {
		return &APNObjectAuth{
			Type:     helperClonePointer(apnPointerAuth.Type),
			Username: helperClonePointer(apnPointerAuth.Username),
			Password: helperClonePointer(apnPointerAuth.Password),
		}
	}

	return nil
}

// Validate checks if authentication type is set and ensures that if username or password
// is provided, both are non-nil (empty string is allowed).
func (apnPointerAuth *APNObjectAuth) Validate() bool {
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

// IsLike returns true if the receiver matches the given APNObjectAuth.
func (apnPointerAuth *APNObjectAuth) IsLike(apnPointer *APNObjectAuth) bool {
	if apnPointerAuth == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return helperIsLikeMaskPointer(apnPointerAuth.Type, apnPointer.Type) &&
		helperIsLikeStringPointer(apnPointerAuth.Username, apnPointer.Username) &&
		helperIsLikeStringPointer(apnPointerAuth.Password, apnPointer.Password)
}

//--------------------------------------------------------------------------------//
// APN Object Bearer
//--------------------------------------------------------------------------------//

// APNObjectBearer contains bearer-level settings: protocol, roaming protocol, MTU, and server.
type APNObjectBearer struct {
	APNObjectInterface[APNObjectBearer] `json:"-" xml:"-"`

	Type        *APNTypeBearerProtocol `json:"type,omitempty"         xml:"protocol,attr,omitempty"`
	TypeRoaming *APNTypeBearerProtocol `json:"typeRoaming,omitempty"  xml:"roaming_protocol,attr,omitempty"`
	Mtu         *int                   `json:"mtu,omitempty"          xml:"mtu,attr,omitempty"`
	Server      *string                `json:"server,omitempty"       xml:"server,attr,omitempty"`
}

// Clone creates a deep copy of APNObjectBearer.
// Returns nil if validation fails.
func (apnPointerBearer *APNObjectBearer) Clone() *APNObjectBearer {
	if apnPointerBearer.Validate() {
		return &APNObjectBearer{
			Type:        helperClonePointer(apnPointerBearer.Type),
			TypeRoaming: helperClonePointer(apnPointerBearer.TypeRoaming),
			Mtu:         helperClonePointer(apnPointerBearer.Mtu),
			Server:      helperClonePointer(apnPointerBearer.Server),
		}
	}

	return nil
}

// Validate checks if at least one of Type or TypeRoaming is set.
func (apnPointerBearer *APNObjectBearer) Validate() bool {
	if apnPointerBearer != nil && (apnPointerBearer.Type != nil || apnPointerBearer.TypeRoaming != nil) {
		return true
	}

	return false
}

// IsLike returns true if the receiver matches the given APNObjectBearer.
func (apnPointerBearer *APNObjectBearer) IsLike(apnPointer *APNObjectBearer) bool {
	if apnPointerBearer == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return helperIsLikeMaskPointer(apnPointerBearer.Type, apnPointer.Type) &&
		helperIsLikeMaskPointer(apnPointerBearer.TypeRoaming, apnPointer.TypeRoaming) &&
		helperIsLikeIntPointer(apnPointerBearer.Mtu, apnPointer.Mtu) &&
		helperIsLikeStringPointer(apnPointerBearer.Server, apnPointer.Server)
}

//--------------------------------------------------------------------------------//
// APN Object Proxy
//--------------------------------------------------------------------------------//

// APNObjectProxy contains proxy server settings: server address and port.
type APNObjectProxy struct {
	APNObjectInterface[APNObjectProxy] `json:"-" xml:"-"`

	Server *string `json:"server,omitempty" xml:"proxy,attr,omitempty"`
	Port   *int    `json:"port,omitempty"   xml:"port,attr,omitempty"`
}

// Clone creates a deep copy of APNObjectProxy.
// Returns nil if validation fails.
func (apnPointerProxy *APNObjectProxy) Clone() *APNObjectProxy {
	if apnPointerProxy.Validate() {
		return &APNObjectProxy{
			Server: helperClonePointer(apnPointerProxy.Server),
			Port:   helperClonePointer(apnPointerProxy.Port),
		}
	}

	return nil
}

// Validate checks that both Server (non-empty) and Port are set.
func (apnPointerProxy *APNObjectProxy) Validate() bool {
	if apnPointerProxy != nil && ((apnPointerProxy.Server != nil && *apnPointerProxy.Server != "") && apnPointerProxy.Port != nil) {
		return true
	}

	return false
}

// IsLike returns true if the receiver matches the given APNObjectProxy.
func (apnPointerProxy *APNObjectProxy) IsLike(apnPointer *APNObjectProxy) bool {
	if apnPointerProxy == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return helperIsLikeStringPointer(apnPointerProxy.Server, apnPointer.Server) &&
		helperIsLikeIntPointer(apnPointerProxy.Port, apnPointer.Port)
}

//--------------------------------------------------------------------------------//
// APN Object MMS
//--------------------------------------------------------------------------------//

// APNObjectMms contains MMS (Multimedia Messaging Service) settings: MMSC URL, proxy, and port.
type APNObjectMms struct {
	APNObjectInterface[APNObjectMms] `json:"-" xml:"-"`

	Center *string `json:"center,omitempty" xml:"mmsc,attr,omitempty"`
	Server *string `json:"server,omitempty" xml:"mmsproxy,attr,omitempty"`
	Port   *int    `json:"port,omitempty"   xml:"mmsport,attr,omitempty"`
}

// Clone creates a deep copy of APNObjectMms.
// Returns nil if validation fails.
func (apnPointerMms *APNObjectMms) Clone() *APNObjectMms {
	if apnPointerMms.Validate() {
		return &APNObjectMms{
			Center: helperClonePointer(apnPointerMms.Center),
			Server: helperClonePointer(apnPointerMms.Server),
			Port:   helperClonePointer(apnPointerMms.Port),
		}
	}

	return nil
}

// Validate checks that at least one of Center, Server, or Port is set.
func (apnPointerMms *APNObjectMms) Validate() bool {
	if apnPointerMms != nil && ((apnPointerMms.Center != nil && *apnPointerMms.Center != "") || (apnPointerMms.Server != nil && *apnPointerMms.Server != "") || apnPointerMms.Port != nil) {
		return true
	}

	return false
}

// IsLike returns true if the receiver matches the given APNObjectMms.
func (apnPointerMms *APNObjectMms) IsLike(apnPointer *APNObjectMms) bool {
	if apnPointerMms == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return helperIsLikeStringPointer(apnPointerMms.Center, apnPointer.Center) &&
		helperIsLikeStringPointer(apnPointerMms.Server, apnPointer.Server) &&
		helperIsLikeIntPointer(apnPointerMms.Port, apnPointer.Port)
}

//--------------------------------------------------------------------------------//
// APN Object MVNO
//--------------------------------------------------------------------------------//

// APNObjectMvno contains MVNO (Mobile Virtual Network Operator) matching settings: type and data.
type APNObjectMvno struct {
	APNObjectInterface[APNObjectMvno] `json:"-" xml:"-"`

	Type *string `json:"type,omitempty" xml:"mvno_type,attr,omitempty"`
	Data *string `json:"data,omitempty" xml:"mvno_match_data,attr,omitempty"`
}

// Clone creates a deep copy of APNObjectMvno.
// Returns nil if validation fails.
func (apnPointerMvno *APNObjectMvno) Clone() *APNObjectMvno {
	if apnPointerMvno.Validate() {
		return &APNObjectMvno{
			Type: helperClonePointer(apnPointerMvno.Type),
			Data: helperClonePointer(apnPointerMvno.Data),
		}
	}

	return nil
}

// Validate checks that at least one of Type or Data is set.
func (apnPointerMvno *APNObjectMvno) Validate() bool {
	if apnPointerMvno != nil && (apnPointerMvno.Type != nil || apnPointerMvno.Data != nil) {
		return true
	}

	return false
}

// IsLike returns true if the receiver matches the given APNObjectMvno.
func (apnPointerMvno *APNObjectMvno) IsLike(apnPointer *APNObjectMvno) bool {
	if apnPointerMvno == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return helperIsLikeStringPointer(apnPointerMvno.Type, apnPointer.Type) &&
		helperIsLikeStringPointer(apnPointerMvno.Data, apnPointer.Data)
}

//--------------------------------------------------------------------------------//
// APN Object Limit
//--------------------------------------------------------------------------------//

// APNObjectLimit contains connection limiting settings: maximum concurrent connections and duration.
type APNObjectLimit struct {
	APNObjectInterface[APNObjectLimit] `json:"-" xml:"-"`

	MaxConn     *int `json:"maxConn,omitempty"      xml:"max_conns,attr,omitempty"`
	MaxConnTime *int `json:"maxConnTime,omitempty"  xml:"max_conns_time,attr,omitempty"`
}

// Clone creates a deep copy of APNObjectLimit.
// Returns nil if validation fails.
func (apnPointerLimit *APNObjectLimit) Clone() *APNObjectLimit {
	if apnPointerLimit.Validate() {
		return &APNObjectLimit{
			MaxConn:     helperClonePointer(apnPointerLimit.MaxConn),
			MaxConnTime: helperClonePointer(apnPointerLimit.MaxConnTime),
		}
	}

	return nil
}

// Validate checks that at least one of MaxConn or MaxConnTime is set.
func (apnPointerLimit *APNObjectLimit) Validate() bool {
	if apnPointerLimit != nil && (apnPointerLimit.MaxConn != nil || apnPointerLimit.MaxConnTime != nil) {
		return true
	}

	return false
}

// IsLike returns true if the receiver matches the given APNObjectLimit.
func (apnPointerLimit *APNObjectLimit) IsLike(apnPointer *APNObjectLimit) bool {
	if apnPointerLimit == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return helperIsLikeIntPointer(apnPointerLimit.MaxConn, apnPointer.MaxConn) &&
		helperIsLikeIntPointer(apnPointerLimit.MaxConnTime, apnPointer.MaxConnTime)
}

//--------------------------------------------------------------------------------//
// APN Object Other
//--------------------------------------------------------------------------------//

// APNObjectOther contains miscellaneous flags and settings not covered by other objects.
// Includes network restrictions, modem settings, and carrier control flags.
type APNObjectOther struct {
	APNObjectInterface[APNObjectOther] `json:"-" xml:"-"`

	// Network restrictions
	NetworkTypeBitmask *APNTypeNetworkType `json:"networkTypeBitmask,omitempty" xml:"network_type_bitmask,attr,omitempty"`

	// Modem settings
	ModemCognitive *bool `json:"modemCognitive,omitempty" xml:"modem_cognitive,attr,omitempty"`

	// Carrier control flags
	CarrierEnabled *bool `json:"IsEnabled,omitempty" xml:"carrier_enabled,attr,omitempty"`
	UserVisible    *bool `json:"IsVisible,omitempty" xml:"user_visible,attr,omitempty"`
	UserEditable   *bool `json:"IsEditable,omitempty" xml:"user_editable,attr,omitempty"`
}

// Clone creates a deep copy of APNObjectOther.
// Returns nil if validation fails.
func (apnPointerOther *APNObjectOther) Clone() *APNObjectOther {
	if apnPointerOther.Validate() {
		return &APNObjectOther{
			NetworkTypeBitmask: helperClonePointer(apnPointerOther.NetworkTypeBitmask),
			ModemCognitive:     helperClonePointer(apnPointerOther.ModemCognitive),
			CarrierEnabled:     helperClonePointer(apnPointerOther.CarrierEnabled),
			UserVisible:        helperClonePointer(apnPointerOther.UserVisible),
			UserEditable:       helperClonePointer(apnPointerOther.UserEditable),
		}
	}

	return nil
}

// Validate checks that at least one field in APNObjectOther is set.
func (apnPointerOther *APNObjectOther) Validate() bool {
	if apnPointerOther != nil && (apnPointerOther.NetworkTypeBitmask != nil || apnPointerOther.ModemCognitive != nil || apnPointerOther.CarrierEnabled != nil || apnPointerOther.UserVisible != nil || apnPointerOther.UserEditable != nil) {
		return true
	}

	return false
}

// IsLike returns true if the receiver matches the given APNObjectOther.
// Currently only compares NetworkTypeBitmask.
func (apnPointerOther *APNObjectOther) IsLike(apnPointer *APNObjectOther) bool {
	if apnPointerOther == nil || apnPointer == nil {
		return apnPointer == nil
	}

	return helperIsLikeMaskPointer(apnPointerOther.NetworkTypeBitmask, apnPointer.NetworkTypeBitmask)
}

//--------------------------------------------------------------------------------//
