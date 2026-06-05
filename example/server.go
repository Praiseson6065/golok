package example

//go:generate go run github.com/praiseson6065/golok/cmd/golok -file=$GOFILE

// +golok:functional_options,stringer,validate
type Server struct {
	Host string `validate:"required"`
	Port int
	TLS  bool
}
