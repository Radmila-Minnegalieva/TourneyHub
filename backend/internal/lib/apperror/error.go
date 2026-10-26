// Package apperror определяет ошибки приложения независимо от HTTP и драйвера БД.
package apperror

type Kind string

const (
	BadRequest   Kind = "bad_request"
	Unauthorized Kind = "unauthorized"
	Forbidden    Kind = "forbidden"
	NotFound     Kind = "not_found"
	Conflict     Kind = "conflict"
	Validation   Kind = "validation_failed"
	RateLimited  Kind = "rate_limited"
	Unavailable  Kind = "unavailable"
)

// Error содержит безопасное сообщение и необязательную исходную причину.
type Error struct {
	Kind    Kind
	Message string
	Cause   error
}

func (e *Error) Error() string            { return e.Message }
func (e *Error) Unwrap() error            { return e.Cause }
func New(kind Kind, message string) error { return &Error{Kind: kind, Message: message} }
func Wrap(kind Kind, message string, cause error) error {
	return &Error{Kind: kind, Message: message, Cause: cause}
}
