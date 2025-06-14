module circular-api

go 1.21 // Or your Go version, e.g., 1.22

// If your circular package had external dependencies that were not part of the Go standard library,
// they would be listed here after you run `go get` or `go mod tidy`.
// For the provided library, all crypto and http functions are in the standard library,
// so initially this `require` block might be empty or not present until you build/run.
// For example, if it used an external elliptic curve library, it might look like:
// require (
//     github.com/some/other/library v1.2.3
// )
