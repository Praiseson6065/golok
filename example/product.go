package example

//go:generate go run github.com/praiseson6065/golok/cmd/golok -file=$GOFILE

// +golok:constructor,builder,stringer
type Product struct {
	ID    int
	Title string
	Price float64
	Tags  []string // +golok:skip
}