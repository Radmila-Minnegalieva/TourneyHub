// Package pagination задаёт значения по умолчанию для постраничных выборок.
package pagination

func Values(limit, offset *int) (int, int) {
	l, o := 20, 0
	if limit != nil {
		l = *limit
	}
	if offset != nil {
		o = *offset
	}
	return l, o
}
