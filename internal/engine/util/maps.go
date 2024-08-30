package util

// Keys returns the keys of the map m. The keys will be an indeterminate order.
func Keys[M ~map[K]V, K comparable, V any](input M) []K {
	output := make([]K, 0, len(input))
	for k := range input {
		output = append(output, k)
	}
	return output
}

// Inverse switches the values to keys and the keys to values.
func Inverse(input map[string]string) map[string]string {
	output := make(map[string]string)
	for k, v := range input {
		output[v] = k
	}
	return output
}

// Cast turns a map[K]V to a map[K]any, so values will downcast to 'any' type.
func Cast[M ~map[K]V, K comparable, V any](input M) map[K]any {
	output := make(map[K]any, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}
