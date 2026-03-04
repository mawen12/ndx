package view

import (
	"context"
	"errors"

	"github.com/mawen12/ndx/internal"
)

func extractApp(ctx context.Context) (*App, error) {
	app, ok := ctx.Value(internal.KeyApp).(*App)
	if !ok {
		return nil, errors.New("no application found in context")
	}
	return app, nil
}
