package purple

type Yellow interface {
	Color(rgb string) error
	Hue(adjust int)
}
