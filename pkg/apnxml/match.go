package apnxml

import (
	"strings"
)

//--------------------------------------------------------------------------------//
// MatchHelper
//--------------------------------------------------------------------------------//

func matchString(left string, right string) bool {
	left = strings.TrimSpace(strings.ToLower(left))
	right = strings.TrimSpace(strings.ToLower(right))

	return right == "" || strings.Contains(left, right)
}

func matchStringPtr(left *string, right *string) bool {
	if left == nil || right == nil {
		return right == nil
	}

	return matchString(*left, *right)
}

func matchIntPtr(left *int, right *int) bool {
	if left == nil || right == nil {
		return right == nil
	}

	return *left == *right
}

func matchMaskPtr[Type ~int](left *Type, right *Type) bool {
	if left == nil || right == nil {
		return right == nil
	}

	return *left&*right == *right
}

//--------------------------------------------------------------------------------//
