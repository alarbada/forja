# forja

The most straight forward way to create typescript clients for go APIs. Think of
this as the tRPC equivalent, but in go.

# Why not openapi?

Basically, I wanted something much more simple and straight forward to use. This
tool only serves one purpose: to generate a typescript api client from go code
at runtime, and nothing else.

Forget about specs, forget about correct REST principles. You can run the
example program at `cmd/` with `air`, and change the req / res json tags. You'll
see that the scripts at `scripts/` typecheck on save.

An [example app](https://github.com/alarbada/forja-solidjs-example)

# TODO

- [ ] Avoid repeating the same input / output type to make generated code slimmer
- [ ] Add adapters support, not `echo` only.
- [ ] Better docs. Maybe with website?
