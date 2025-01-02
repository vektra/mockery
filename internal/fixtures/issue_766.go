package test

type Issue766 interface {
	FetchData(
		fetchFunc func(x ...int) ([]int, error),
	) ([]int, error)
}
