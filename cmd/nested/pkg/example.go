package pkg

import "github.com/labstack/echo/v4"

type SomeHandlerReq struct{}
type SomeHandlerRes struct{}

func SomeHandler(e echo.Context, params SomeHandlerReq) (string, error) {
	return "some handler response", nil
}

type Type1 struct {
	Name string
	Age  int
}

type Type2 struct {
	Type Type1
}
