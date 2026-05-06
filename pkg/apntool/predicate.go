package apntool

import (
	"strings"

	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

type Predicate func(record apnxml.Object) bool

func All(record apnxml.Object) bool {
	return true
}

func Not(predicate Predicate) Predicate {
	if predicate == nil {
		return func(record apnxml.Object) bool {
			return false
		}
	}

	return func(record apnxml.Object) bool {
		return !predicate(record)
	}
}

func And(predicates ...Predicate) Predicate {
	return func(record apnxml.Object) bool {
		for _, predicate := range predicates {
			if predicate != nil && !predicate(record) {
				return false
			}
		}

		return true
	}
}

func Or(predicates ...Predicate) Predicate {
	return func(record apnxml.Object) bool {
		for _, predicate := range predicates {
			if predicate != nil && predicate(record) {
				return true
			}
		}

		return false
	}
}

func ByPLMN(mcc int, mnc int) Predicate {
	return func(record apnxml.Object) bool {
		return record.Mcc != nil && record.Mnc != nil && *record.Mcc == mcc && *record.Mnc == mnc
	}
}

func ByMCC(mcc int) Predicate {
	return func(record apnxml.Object) bool {
		return record.Mcc != nil && *record.Mcc == mcc
	}
}

func ByMNC(mnc int) Predicate {
	return func(record apnxml.Object) bool {
		return record.Mnc != nil && *record.Mnc == mnc
	}
}

func ByCarrierID(carrierID int) Predicate {
	return func(record apnxml.Object) bool {
		return record.CarrierID != nil && *record.CarrierID == carrierID
	}
}

func ByType(apnType apnxml.ObjectBaseType) Predicate {
	return func(record apnxml.Object) bool {
		return record.Base != nil &&
			record.Base.Type != nil &&
			*record.Base.Type&apnType == apnType
	}
}

func ByProtocol(protocol apnxml.ObjectBearerProtocol) Predicate {
	return func(record apnxml.Object) bool {
		return record.Bearer != nil &&
			record.Bearer.Type != nil &&
			*record.Bearer.Type&protocol == protocol
	}
}

func ByNetwork(network apnxml.ObjectNetworkType) Predicate {
	return func(record apnxml.Object) bool {
		return record.Other != nil &&
			record.Other.NetworkTypeBitmask != nil &&
			*record.Other.NetworkTypeBitmask&network == network
	}
}

func ByAPN(query string) Predicate {
	query = strings.TrimSpace(strings.ToLower(query))

	return func(record apnxml.Object) bool {
		if query == "" {
			return true
		}
		if record.Base == nil || record.Base.Apn == nil {
			return false
		}

		return strings.EqualFold(strings.TrimSpace(*record.Base.Apn), query)
	}
}

func ByAPNContains(query string) Predicate {
	query = strings.TrimSpace(strings.ToLower(query))

	return func(record apnxml.Object) bool {
		if query == "" {
			return true
		}
		if record.Base == nil || record.Base.Apn == nil {
			return false
		}

		return strings.Contains(strings.ToLower(*record.Base.Apn), query)
	}
}

func ByCarrierName(query string) Predicate {
	query = strings.TrimSpace(strings.ToLower(query))

	return func(record apnxml.Object) bool {
		if query == "" {
			return true
		}

		return strings.Contains(strings.ToLower(record.Carrier), query)
	}
}

func HasRoot(record apnxml.Object) bool {
	return record.ObjectRoot != nil
}

func HasValidRoot(record apnxml.Object) bool {
	return record.ObjectRoot != nil && record.ObjectRoot.Validate()
}

func HasBase(record apnxml.Object) bool {
	return record.Base != nil && record.Base.Validate()
}

func HasAuth(record apnxml.Object) bool {
	return record.Auth != nil && record.Auth.Validate()
}

func HasBearer(record apnxml.Object) bool {
	return record.Bearer != nil && record.Bearer.Validate()
}

func HasProxy(record apnxml.Object) bool {
	return record.Proxy != nil && record.Proxy.Validate()
}

func HasMMS(record apnxml.Object) bool {
	return record.Mms != nil && record.Mms.Validate()
}

func HasMVNO(record apnxml.Object) bool {
	return record.Mvno != nil && record.Mvno.Validate()
}

func IsValid(record apnxml.Object) bool {
	return record.Validate()
}

func Match(query *apnxml.Object) Predicate {
	return func(record apnxml.Object) bool {
		return (&record).Match(query)
	}
}
