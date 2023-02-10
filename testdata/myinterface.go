package testdata

type MyInterface interface {
	Foo(a string, b int, c ...string)
	Bar(a string) error
	Abc() (string, error)
}
