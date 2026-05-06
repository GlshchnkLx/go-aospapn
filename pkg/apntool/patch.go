package apntool

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

type PatchResult struct {
	Data    Array
	Matched int
	Changed int
}

func (array Array) UpdateByFilter(predicate Predicate, patch *apnxml.Object, mode apnxml.ObjectUpdateMode) (PatchResult, error) {
	if predicate == nil {
		predicate = All
	}

	var result PatchResult
	data, err := array.ApplyEntries(func(group *apnxml.Object, record *apnxml.Object) error {
		materialized := MaterializeRecord(group, record)
		if !predicate(materialized) {
			return nil
		}

		result.Matched++
		if patch != nil && patch.ObjectRoot != nil {
			target := record
			if group != nil && group.ObjectRoot != nil {
				target = group
			}
			if target.Update(&apnxml.Object{ObjectRoot: patch.ObjectRoot}, mode) {
				result.Changed++
			}
		}
		if patch != nil && hasPatchSections(patch) && record.Update(&apnxml.Object{
			Base:   patch.Base,
			Auth:   patch.Auth,
			Bearer: patch.Bearer,
			Proxy:  patch.Proxy,
			Mms:    patch.Mms,
			Mvno:   patch.Mvno,
			Limit:  patch.Limit,
			Other:  patch.Other,
		}, mode) {
			result.Changed++
		}
		return nil
	})
	if err != nil {
		return PatchResult{}, err
	}

	result.Data = data
	return result, nil
}

func hasPatchSections(patch *apnxml.Object) bool {
	return patch != nil &&
		(patch.Base != nil ||
			patch.Auth != nil ||
			patch.Bearer != nil ||
			patch.Proxy != nil ||
			patch.Mms != nil ||
			patch.Mvno != nil ||
			patch.Limit != nil ||
			patch.Other != nil)
}

func SetObjectField(record *apnxml.Object, name string, value string) error {
	if record == nil {
		return fmt.Errorf("set %s: nil APN object", name)
	}

	path := strings.ToLower(strings.TrimSpace(name))
	path = strings.ReplaceAll(path, "_", "")
	path = strings.ReplaceAll(path, "-", "")

	switch path {
	case "root.carrier", "carrier", "carriername":
		EnsureRoot(record).Carrier = value
	case "root.carrierid", "carrierid":
		v, err := parseInt(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureRoot(record).CarrierID = &v
	case "root.mcc", "mcc":
		v, err := parseInt(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureRoot(record).Mcc = &v
	case "root.mnc", "mnc":
		v, err := parseInt(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureRoot(record).Mnc = &v
	case "base.apn", "apn":
		EnsureBase(record).Apn = &value
	case "base.type", "type":
		v, err := apnxml.ParseObjectBaseType(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureBase(record).Type = &v
	case "base.profileid", "profileid":
		v, err := parseInt(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureBase(record).ProfileID = &v
	case "auth.type", "authtype":
		v, err := apnxml.ParseObjectAuthType(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureAuth(record).Type = &v
	case "auth.username", "auth.user", "user", "username":
		EnsureAuth(record).Username = &value
	case "auth.password", "password":
		EnsureAuth(record).Password = &value
	case "bearer.type", "bearer.protocol", "protocol":
		v, err := apnxml.ParseObjectBearerProtocol(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureBearer(record).Type = &v
	case "bearer.typeroaming", "bearer.roamingprotocol", "roamingprotocol":
		v, err := apnxml.ParseObjectBearerProtocol(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureBearer(record).TypeRoaming = &v
	case "bearer.mtu", "mtu":
		v, err := parseInt(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureBearer(record).Mtu = &v
	case "bearer.server":
		EnsureBearer(record).Server = &value
	case "proxy.server", "proxy":
		EnsureProxy(record).Server = &value
	case "proxy.port", "port":
		v, err := parseInt(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureProxy(record).Port = &v
	case "mms.center", "mmsc":
		EnsureMMS(record).Center = &value
	case "mms.server", "mmsproxy":
		EnsureMMS(record).Server = &value
	case "mms.port", "mmsport":
		v, err := parseInt(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureMMS(record).Port = &v
	case "mvno.type", "mvnotype":
		EnsureMVNO(record).Type = &value
	case "mvno.data", "mvnomatchdata":
		EnsureMVNO(record).Data = &value
	case "limit.maxconn", "maxconn":
		v, err := parseInt(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureLimit(record).MaxConn = &v
	case "limit.maxconntime", "maxconntime":
		v, err := parseInt(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureLimit(record).MaxConnTime = &v
	case "other.networktypebitmask", "networktypebitmask", "network":
		v, err := apnxml.ParseObjectNetworkType(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureOther(record).NetworkTypeBitmask = &v
	case "other.modemcognitive", "modemcognitive":
		v, err := parseBool(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureOther(record).ModemCognitive = &v
	case "other.carrierenabled", "carrierenabled", "enabled":
		v, err := parseBool(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureOther(record).CarrierEnabled = &v
	case "other.uservisible", "uservisible", "visible":
		v, err := parseBool(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureOther(record).UserVisible = &v
	case "other.usereditable", "usereditable", "editable":
		v, err := parseBool(value)
		if err != nil {
			return fieldError(name, err)
		}
		EnsureOther(record).UserEditable = &v
	default:
		return fmt.Errorf("unsupported APN field: %s", name)
	}

	return nil
}

func SetObjectFieldExpr(record *apnxml.Object, expr string) error {
	name, value, ok := strings.Cut(expr, "=")
	if !ok {
		return fmt.Errorf("invalid set expression %q, expected field=value", expr)
	}
	return SetObjectField(record, name, value)
}

func parseInt(value string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(value))
}

func parseBool(value string) (bool, error) {
	return strconv.ParseBool(strings.TrimSpace(value))
}

func fieldError(name string, err error) error {
	return fmt.Errorf("set %s: %w", name, err)
}
