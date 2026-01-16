package daggerx

import (
	"context"
	"io"

	"dagger.io/dagger"
)

// Session wraps a Dagger client and context for shared use.
type Session struct {
	Ctx    context.Context
	Client *dagger.Client
}

// NewSession creates a Dagger session with log output routed to the provided writer.
// If logOutput is nil, Dagger internals are silenced by default.
func NewSession(logOutput io.Writer) (*Session, error) {
	if logOutput == nil {
		logOutput = io.Discard
	}
	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(logOutput))
	if err != nil {
		return nil, err
	}
	return &Session{Ctx: ctx, Client: client}, nil
}

// Close releases the Dagger engine connection.
func (s *Session) Close() error {
	if s == nil || s.Client == nil {
		return nil
	}
	return s.Client.Close()
}
