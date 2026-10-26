// Package slice содержит преобразования коллекций с сохранением nil.
package slice

func Map[T, U any](values []T, convert func(T) U) []U {
	if values == nil {
		return nil
	}
	result := make([]U, len(values))
	for i, value := range values {
		result[i] = convert(value)
	}
	return result
}
