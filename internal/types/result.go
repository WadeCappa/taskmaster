package types

type Result[T any] struct {
	v   T
	err error
}

func Success[T any](v T) Result[T] {
	return Result[T]{
		v: v,
	}
}

func Failure[T any](err error) Result[T] {
	return Result[T]{
		err: err,
	}
}

func (o Result[T]) Unwrap() (T, error) {
	return o.v, o.err
}
