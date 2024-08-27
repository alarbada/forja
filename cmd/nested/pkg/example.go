package pkg

import "github.com/labstack/echo/v4"

type SomeHandlerReq struct{}
type SomeHandlerRes struct{}

func SomeHandler(e echo.Context, params SomeHandlerReq) (string, error) {
	return "some handler response", nil
}
