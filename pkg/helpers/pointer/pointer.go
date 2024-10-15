package pointer

// GetPointer getting pointer by value and type
func GetPointer[T any](value T) *T {
	return &value
}

// GetValue save getting value of pointer by pointer and type T. If function get nil-pointer returns zero value for T
func GetValue[T any](value *T) T {
	if value == nil {
		var exemplar T
		return exemplar
	}
	return *value
}
