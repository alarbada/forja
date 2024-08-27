package main

import (
	"github.com/alarbada/gotsclient"
	"github.com/alarbada/gotsclient/cmd/nested/pkg"

	"github.com/gookit/goutil/dump"
	"github.com/labstack/echo/v4"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type ExampleParams struct {
	Name  string `json:"name"`
	Users []User `json:"users"`
}

type ExampleResponse struct {
	Greeting string `json:"greeting"`
}

func ExampleHandler1(c echo.Context, params ExampleParams) (*ExampleResponse, error) {
	dump.P(params)

	return &ExampleResponse{Greeting: "Hello, " + params.Name}, nil
}

func ExampleHandler2(c echo.Context, params ExampleParams) (*ExampleResponse, error) {
	dump.P(params)

	return &ExampleResponse{Greeting: "Hello, " + params.Name}, nil
}

func HelloWorld(c echo.Context, params struct{}) (string, error) {
	return "hello world", nil
}

type Playlist struct {
	ID          string `json:"id,omitempty"`
	PlaylistID  string `json:"playlistId,omitempty"`
	Title       string `json:"title,omitempty"`
	Pinned      bool   `json:"pinned,omitempty"`
	Description string `json:"description,omitempty"`
}

func getPlaylists(
	c echo.Context, i *struct{},
) ([]Playlist, error) {
	return nil, nil
}

func main() {
	e := echo.New()
	th := gotsclient.NewTypedHandlers(e)

	gotsclient.AddHandler(th, ExampleHandler1)
	gotsclient.AddHandler(th, ExampleHandler2)
	gotsclient.AddHandler(th, HelloWorld)
	gotsclient.AddHandler(th, pkg.SomeHandler)
	gotsclient.AddHandler(th, getPlaylists)

	gotsclient.WriteToFile(th, "scripts/apiclient.ts")

	e.Logger.Fatal(e.Start(":8080"))
}
