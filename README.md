# Functypes

Converts interface methods to standalone function types.

Example:

From
```go
type MyInterface interface {
	Foo(val string) error
	Bar() (int, error)
}
```

To
```go
type Foo func(val string) error
type Bar func() (int, error)
```

Usage:

Default args: Scan the current directory (`.`) for a go package and write output to `./functypes/`:
```
functypes
```

Set args:
```
functypes --pkg-path /path/to/go/package/dir --out-dir /path/to/output/dir
```

