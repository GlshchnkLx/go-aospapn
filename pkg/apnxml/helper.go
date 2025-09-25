package apnxml

import (
	"strings"
)

//--------------------------------------------------------------------------------//
// Helper IsLike
//--------------------------------------------------------------------------------//

func helperIsLikeString(left string, right string) bool {
	left = strings.TrimSpace(strings.ToLower(left))
	right = strings.TrimSpace(strings.ToLower(right))

	return right == "" || strings.Contains(left, right)
}

func helperIsLikeStringPointer(left *string, right *string) bool {
	if left == nil || right == nil {
		return right == nil
	}

	return helperIsLikeString(*left, *right)
}

func helperIsLikeIntPointer(left *int, right *int) bool {
	if left == nil || right == nil {
		return right == nil
	}

	return *left == *right
}

func helperIsLikeMaskPointer[Type ~int](left *Type, right *Type) bool {
	if left == nil || right == nil {
		return right == nil
	}

	return *left&*right == *right
}

//--------------------------------------------------------------------------------//
// Helper Clone
//--------------------------------------------------------------------------------//

func helperPointerClone[Type any](pointer *Type) *Type {
	if pointer == nil {
		return nil
	}

	object := *pointer
	return &object
}

//--------------------------------------------------------------------------------//
