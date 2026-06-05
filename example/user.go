package example

//go:generate go run github.com/praiseson6065/golok/cmd/golok -file=$GOFILE

// +golok:all,interface,mapper=UserDTO
type User struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required"`
	Age   int    `json:"age"`
}

// UserDTO is a data transfer object — a subset of User for API responses.
type UserDTO struct {
	Name  string
	Email string
}
