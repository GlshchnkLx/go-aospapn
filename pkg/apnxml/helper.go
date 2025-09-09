package apnxml

//--------------------------------------------------------------------------------//
// Helper Method
//--------------------------------------------------------------------------------//

func helperClonePointer[Type any](pointer *Type) *Type {
	if pointer == nil {
		return nil
	}

	object := *pointer
	return &object
}

func helperApnPointerClone[Type APNObjectInterface[Type]](apnPointer Type) Type {
	if apnPointer.Validate() {
		return apnPointer.Clone()
	}

	var apnPointerIsNil Type
	return apnPointerIsNil
}

//--------------------------------------------------------------------------------//
