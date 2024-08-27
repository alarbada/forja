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

# TODO

- [ ] Add adapters support, not `echo` only.
- [ ] Generate type definition for each handler instead of inlining so that
      user can import individual type definitions.
- [ ] Better docs. Maybe with website?
