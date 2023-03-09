package context

import (
	"context"
	"github.com/arkadiont/lenslocked/models"
)

type key int

const (
	userKey key = iota
)

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.User {
	val := ctx.Value(userKey)
	if user, ok := val.(*models.User); ok {
		return user
	}
	return nil
}
