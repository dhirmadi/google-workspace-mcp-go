// Package ptr provides generic pointer-helper functions used across tool packages.
package ptr

// Bool returns a pointer to the given bool value.
func Bool(b bool) *bool { return &b }
