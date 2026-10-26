// Package ptr содержит преобразования необязательных значений.
package ptr

func Map[T, U any](value *T, convert func(T) U) *U {
	if value == nil {
		return nil
	}
	result := convert(*value)
	return &result
}
func Value[T any](value T) *T { return &value }
