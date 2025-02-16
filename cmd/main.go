package main

import (
	"github.com/alarbada/forja"
	"github.com/alarbada/forja/cmd/nested/pkg"

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

func ExampleWithExternalTypes(c echo.Context, params pkg.Type2) (*pkg.Type1, error) {
	return nil, nil
}

type HelloWorldOutput struct {
	Result string `json:"result"`
}

func HelloWorld(c echo.Context, params struct{}) (*HelloWorldOutput, error) {
	return &HelloWorldOutput{
		Result: "hello world",
	}, nil
}

type Playlist struct {
	ID          string `json:"id,omitempty"`
	PlaylistID  string `json:"playlistId,omitempty"`
	Title       string `json:"title,omitempty"`
	Pinned      bool   `json:"pinned,omitempty"`
	Description string `json:"description,omitempty"`
}

func getPlaylists(c echo.Context, _ struct{}) ([]Playlist, error) {
	return nil, nil
}

type Server struct{}

func (s Server) theHandler(c echo.Context, input struct{}) (*struct{}, error) {
	return nil, nil
}

func (s *Server) theHandlerPtr(c echo.Context, input struct{}) (*struct{}, error) {
	return nil, nil
}

type Node struct {
	Children []Node
}

func circular(c echo.Context, input *struct{}) (*Node, error) {
	return nil, nil
}

// PointersAreUndefined, so that we don't need to fill them in our typescript definitions
type PointersAreUndefined struct {
	APtr       *string
	AnotherPtr *struct{ Name string }
}

type weHandleInputPointersOutput struct {
	APtrIsUndefined       bool
	AnotherPtrIsUndefined bool
}

func weHandleInputPointers(
	c echo.Context, input PointersAreUndefined,
) (*weHandleInputPointersOutput, error) {
	var res weHandleInputPointersOutput
	if input.APtr == nil {
		res.APtrIsUndefined = true
	}
	if input.AnotherPtr == nil {
		res.AnotherPtrIsUndefined = true
	}

	return &res, nil
}

// Yeah, ik, golang does not have enums, so this is the best I can think of to
// encode enums. I cannot think of a way to make the tag approach (have an int
// or string indicating which option is valid) work, as we cannot obtain all possible
// tag values by reflection.
type EnumLike struct {
	Opt1 forja.Option[string]
	Opt2 forja.Option[struct {
		Name string
		Age  int
	}]
}

type weAlsoHandleEnumsResult struct {
	Opt1WasFilled bool
	Opt2WasFilled bool
}

func weAlsoHandleEnums(c echo.Context, input EnumLike) (_ *weAlsoHandleEnumsResult, err error) {
	var res weAlsoHandleEnumsResult
	if input.Opt1.Valid() {
		res.Opt1WasFilled = true
	}
	if input.Opt2.Valid() {
		res.Opt2WasFilled = true
	}

	return &res, nil
}

func main() {
	e := echo.New()
	th := forja.NewForja(e)

	forja.AddHandler(th, ExampleHandler1)
	forja.AddHandler(th, ExampleHandler2)
	forja.AddHandler(th, HelloWorld)
	forja.AddHandler(th, pkg.SomeHandler)
	forja.AddHandler(th, getPlaylists)
	forja.AddHandler(th, ExampleWithExternalTypes)

	server := Server{}
	forja.AddHandler(th, server.theHandler)
	forja.AddHandler(th, server.theHandlerPtr)
	forja.AddHandler(th, circular)
	forja.AddHandler(th, weHandleInputPointers)
	forja.AddHandler(th, weAlsoHandleEnums)

	forja.WriteToFile(th, "scripts/apiclient.ts")

	e.Logger.Fatal(e.Start(":8080"))
}
