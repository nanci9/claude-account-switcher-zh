//go:build !windows

package main

func accountWindowHandles() map[uintptr]bool       { return nil }
func resizeAccountWindow(map[uintptr]bool) uintptr { return 0 }
func activateAccountWindow(uintptr) bool           { return false }
