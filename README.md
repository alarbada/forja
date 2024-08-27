# gotsclient

The easiest way to create typescript clients for go APIs

# Why not openapi?

Because openapi is made to create generic API specifications. This is not the purpose of this project. This is meant to be the easiest way to create a front end typescript client from a defined go API.

Forget about specs, forget about correct REST principles. You could say that this is the tRPC equivalent, but in go. You can run the example program at `cmd/` with `air`, and change the req / res json tags. You'll see that the scripts at `scripts/` typecheck on save.
