package main

import (
	"time"

	"github.com/alarbada/forja"
	"github.com/alarbada/forja/cmd/nested/pkg"

	"github.com/gookit/goutil/dump"
	"github.com/labstack/echo/v4"
)

type User struct {
	Name    string    `json:"name"`
	Age     int       `json:"age"`
	Created time.Time `json:"created"`
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
	fj := forja.NewForja(e)

	forja.AddHandler(fj, ExampleHandler1)
	forja.AddHandler(fj, ExampleHandler2)
	forja.AddHandler(fj, HelloWorld)
	forja.AddHandler(fj, pkg.SomeHandler)
	forja.AddHandler(fj, getPlaylists)
	forja.AddHandler(fj, ExampleWithExternalTypes)

	server := Server{}
	forja.AddHandler(fj, server.theHandler)
	forja.AddHandler(fj, server.theHandlerPtr)
	forja.AddHandler(fj, circular)
	forja.AddHandler(fj, weHandleInputPointers)
	forja.AddHandler(fj, weAlsoHandleEnums)

	// Add custom variables to be exported in the TypeScript client

	// Simple primitive values
	fj.AddVariable("API_VERSION", "1.0.0")
	fj.AddVariable("MAX_RETRIES", 3)
	fj.AddVariable("DEBUG_MODE", true)

	// Complex objects
	fj.AddVariable("DEFAULT_USER", User{
		Name:    "John Doe",
		Age:     30,
		Created: time.Now(),
	})

	// Arrays/slices
	fj.AddVariable("SUPPORTED_FORMATS", []string{"json", "xml", "yaml"})

	// Maps
	fj.AddConstVariable("ERROR_CODES", map[string]int{
		"NOT_FOUND":    404,
		"UNAUTHORIZED": 401,
		"SERVER_ERROR": 500,
		"BAD_REQUEST":  400,
		"RATE_LIMITED": 429,
	})

	// Nested complex structures
	fj.AddVariable("SAMPLE_PLAYLISTS", []Playlist{
		{
			ID:          "pl1",
			PlaylistID:  "playlist1",
			Title:       "My Favorites",
			Pinned:      true,
			Description: "A collection of my favorite songs",
		},
		{
			ID:          "pl2",
			PlaylistID:  "playlist2",
			Title:       "Workout Mix",
			Pinned:      false,
			Description: "Songs for the gym",
		},
	})

	forja.WriteToFile(fj, "scripts/apiclient.ts")

	e.Logger.Fatal(e.Start(":8080"))
}
