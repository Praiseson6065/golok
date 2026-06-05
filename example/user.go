package example

//go:generate go run github.com/praiseson6065/golok/cmd/golok -file=$GOFILE

// +golok:all
type User struct {
	Name  string
	Email string
	Age   int
}


