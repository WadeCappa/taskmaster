package types

type Option[T any] struct {
	v      T
	exists bool
}

func Some[T any](v T) Option[T] {
	return Option[T]{
		v:      v,
		exists: true,
	}
}

func None[T any]() Option[T] {
	return Option[T]{
		exists: false,
	}
}

func (o Option[T]) Unwrap() (T, bool) {
	return o.v, o.exists
}
