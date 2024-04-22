package util

// Keys returns the keys of the map m.
// The keys will be an indeterminate order.
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}

func ReverseMap(input map[string]string) map[string]string {
	output := make(map[string]string)
	for k, v := range input {
		output[v] = k
	}
	return output
}
