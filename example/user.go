package example

//go:generate go run github.com/praiseson6065/golok/cmd/golok -file=$GOFILE

// +golok:getter,stringer,interface,mapper=UserDTO
type User struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required"`
	Age   int    `json:"age"`
}

// UserDTO is a data transfer object — a subset of User for API responses.
// No golok directive needed on the target; it just needs to exist in the same file.
type UserDTO struct {
	Name  string
	Email string
}
