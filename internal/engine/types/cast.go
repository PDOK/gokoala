package types

func ToInterfaceSlice[IN any, OUT any](in []IN) []OUT {
	out := make([]OUT, len(in))
	for i, v := range in {
		out[i] = any(v).(OUT)
	}
	return out
}

func PtrTo[T any](val T) *T {
	return &val
}
