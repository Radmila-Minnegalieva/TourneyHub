// Package authctx переносит проверенную учётную запись в контексте операции.
package authctx

import (
	"context"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/google/uuid"
)

type key struct{}
type Principal = model.Principal

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, key{}, p)
}
func FromContext(ctx context.Context) Principal { p, _ := ctx.Value(key{}).(Principal); return p }
func CurrentUser(ctx context.Context) uuid.UUID { return FromContext(ctx).ID }
func HasRole(ctx context.Context, roles ...model.Role) bool {
	for _, actual := range FromContext(ctx).Roles {
		if actual == model.Admin {
			return true
		}
		for _, role := range roles {
			if actual == role {
				return true
			}
		}
	}
	return false
}
