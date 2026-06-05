package example

// +golok:constructor,builder,stringer
type Product struct {
	ID    int
	Title string
	Price float64
	Tags  []string // +golok:skip
}