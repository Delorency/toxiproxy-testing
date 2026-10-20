package client

import "context"

type Handler func(ctx context.Context, path string) error

type Middleware func(next Handler) Handler
