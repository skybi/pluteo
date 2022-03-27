package function

// Nest nests several functions to allow rewriting expressions like `res := a(b(c(final)))` as
// `res := function.Nest(final, a, b, c)`.
func Nest[T any](final T, funcs ...func(T) T) T {
	res := final
	for i := len(funcs); i > 0; i-- {
		res = funcs[i-1](res)
	}
	return res
}
