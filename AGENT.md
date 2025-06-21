# Forja Project - Agent Guide

## Project Overview
Forja is a TypeScript client generation tool for Go APIs. It generates TypeScript API clients from Go handler functions at runtime, similar to tRPC but for Go.

## Key Commands

### Development
- `air` - Run development server with hot reload (uses .air.toml config)
- `go run ./cmd` - Generate TypeScript client (`scripts/apiclient.ts`)
- `go build -o ./tmp/main ./cmd` - Build the example application

### Testing
- `go test ./...` - Run all tests
- Check `scripts/apiclient.ts` - Verify generated TypeScript client
- TypeScript files in `scripts/` typecheck on save

### Build & Check
- `go build -o tmp/main .` - build package and check for errors
- `go fmt ./...` - Format Go code

## Key Components

### Features
- Automatic TypeScript client generation from Go handlers
- Support for custom types and variables
- Optional values with `Option[T]` type
- Echo framework integration
- Error handling with `ApiError` type

## Coding Conventions

### Go
- Follow standard Go naming conventions
- Use reflection for runtime type information
- Error handling returns `(result, error)` pattern

### TypeScript Generation
- Generates type-safe API clients
- Uses `ApiResponse<T>` wrapper type
- Supports optional parameters
- Exports custom variables and constants

## Dependencies
- `github.com/labstack/echo/v4` - Web framework
- Go 1.23.0+ required

## Testing Strategy
- Integration tests by running example and checking generated TypeScript
- TypeScript type checking in `scripts/` directory
- Manual verification of generated client functionality
