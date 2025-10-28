package misc

func Must[T any](t T, e error) T {
	if e == nil {
		return t
	}
	panic(e)
}

func Throw(e error) {
	if e != nil {
		panic(e)
	}
}
