package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/apperror"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/security"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

// resetRepository оставляет сервисному тесту только типизированные операции хранения.
type resetRepository struct {
	repository.Store
	userID       uuid.UUID
	tokenError   error
	calls        []string
	passwordHash string
}

func (r *resetRepository) WithinTx(ctx context.Context, fn func(repository.Store) error) error {
	return fn(r)
}
func (r *resetRepository) ConsumeAccountToken(_ context.Context, _, purpose string) (uuid.UUID, error) {
	r.calls = append(r.calls, "consume:"+purpose)
	return r.userID, r.tokenError
}
func (r *resetRepository) SetPasswordHash(_ context.Context, id uuid.UUID, hash string) error {
	if id != r.userID {
		return errors.New("unexpected user")
	}
	r.calls = append(r.calls, "password")
	r.passwordHash = hash
	return nil
}
func (r *resetRepository) RevokeUserSessions(_ context.Context, id uuid.UUID) error {
	if id != r.userID {
		return errors.New("unexpected user")
	}
	r.calls = append(r.calls, "sessions")
	return nil
}
func (r *resetRepository) InvalidateAccountTokens(_ context.Context, id uuid.UUID, purpose string) error {
	if id != r.userID {
		return errors.New("unexpected user")
	}
	r.calls = append(r.calls, "invalidate:"+purpose)
	return nil
}

func TestResetPasswordUsesTypedRepository(t *testing.T) {
	r := &resetRepository{userID: uuid.New()}
	s := New(r, config.Config{}, nil)
	password := "new-password-123"
	if err := s.ResetPassword(context.Background(), model.ResetPasswordRequest{Password: password, Token: "one-time"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"consume:reset_password", "password", "sessions", "invalidate:reset_password"}
	if !reflect.DeepEqual(r.calls, want) {
		t.Fatalf("operations=%v, want=%v", r.calls, want)
	}
	if !security.CheckPassword(r.passwordHash, password) || r.passwordHash == password {
		t.Fatal("password was not hashed before persistence")
	}
}

func TestInvalidResetTokenDoesNotChangeAccount(t *testing.T) {
	r := &resetRepository{userID: uuid.New(), tokenError: repository.ErrNotFound}
	s := New(r, config.Config{}, nil)
	err := s.ResetPassword(context.Background(), model.ResetPasswordRequest{Password: "new-password-123", Token: "expired"})
	var applicationError *apperror.Error
	if !errors.As(err, &applicationError) || applicationError.Kind != apperror.Validation {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(r.calls, []string{"consume:reset_password"}) {
		t.Fatalf("invalid token changed account: %v", r.calls)
	}
}
