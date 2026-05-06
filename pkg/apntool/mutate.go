package apntool

import "github.com/GlshchnkLx/go-aospapn/pkg/apnxml"

type SetPolicy int

const (
	SetIfEmpty SetPolicy = iota
	SetIfExists
	SetAlways
)

func shouldSetPointer[Type comparable](target *Type, policy SetPolicy) bool {
	switch policy {
	case SetIfEmpty:
		var zero Type
		return target == nil || *target == zero
	case SetIfExists:
		return target != nil
	case SetAlways:
		return true
	default:
		return false
	}
}

func Set[Type comparable](target **Type, value Type, policy SetPolicy) bool {
	if target == nil || !shouldSetPointer(*target, policy) {
		return false
	}

	*target = &value
	return true
}

func EnsureRoot(record *apnxml.Object) *apnxml.ObjectRoot {
	if record == nil {
		return nil
	}

	if record.ObjectRoot == nil {
		record.ObjectRoot = &apnxml.ObjectRoot{}
	}

	return record.ObjectRoot
}

func EnsureBase(record *apnxml.Object) *apnxml.ObjectBase {
	if record == nil {
		return nil
	}

	if record.Base == nil {
		record.Base = &apnxml.ObjectBase{}
	}

	return record.Base
}

func EnsureAuth(record *apnxml.Object) *apnxml.ObjectAuth {
	if record == nil {
		return nil
	}

	if record.Auth == nil {
		record.Auth = &apnxml.ObjectAuth{}
	}

	return record.Auth
}

func EnsureBearer(record *apnxml.Object) *apnxml.ObjectBearer {
	if record == nil {
		return nil
	}

	if record.Bearer == nil {
		record.Bearer = &apnxml.ObjectBearer{}
	}

	return record.Bearer
}

func EnsureProxy(record *apnxml.Object) *apnxml.ObjectProxy {
	if record == nil {
		return nil
	}

	if record.Proxy == nil {
		record.Proxy = &apnxml.ObjectProxy{}
	}

	return record.Proxy
}

func EnsureMMS(record *apnxml.Object) *apnxml.ObjectMMS {
	if record == nil {
		return nil
	}

	if record.Mms == nil {
		record.Mms = &apnxml.ObjectMMS{}
	}

	return record.Mms
}

func EnsureMVNO(record *apnxml.Object) *apnxml.ObjectMVNO {
	if record == nil {
		return nil
	}

	if record.Mvno == nil {
		record.Mvno = &apnxml.ObjectMVNO{}
	}

	return record.Mvno
}

func EnsureLimit(record *apnxml.Object) *apnxml.ObjectLimit {
	if record == nil {
		return nil
	}

	if record.Limit == nil {
		record.Limit = &apnxml.ObjectLimit{}
	}

	return record.Limit
}

func EnsureOther(record *apnxml.Object) *apnxml.ObjectOther {
	if record == nil {
		return nil
	}

	if record.Other == nil {
		record.Other = &apnxml.ObjectOther{}
	}

	return record.Other
}
