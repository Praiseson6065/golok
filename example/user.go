package example

//go:generate go run github.com/praiseson6065/golok/cmd/golok -file=$GOFILE

// +golok:all
type User struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required"`
	Age   int    `json:"age"`
}
