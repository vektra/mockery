package embeddedtypes

import "context"

type GreeterServer interface {
	Greet(context.Context, string) (string, error)
	mustEmbedUnimplementedGreeterServer()
}

type UnimplementedGreeterServer struct{}

func (UnimplementedGreeterServer) Greet(context.Context, string) (string, error) {
	return "unimplemented", nil
}

func (UnimplementedGreeterServer) mustEmbedUnimplementedGreeterServer() {}

type OtherServer interface {
	Version() string
	mustEmbedUnimplementedOtherServer()
}

type UnimplementedOtherServer struct{}

func (UnimplementedOtherServer) Version() string { return "unimplemented" }

func (UnimplementedOtherServer) mustEmbedUnimplementedOtherServer() {}

type Store[T any] interface {
	Load() T
	mustEmbedUnimplementedStore()
}

type UnimplementedStore[T any] struct{}

func (UnimplementedStore[T]) Load() T {
	var value T
	return value
}

func (UnimplementedStore[T]) mustEmbedUnimplementedStore() {}
