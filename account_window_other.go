// 中文社区版修改文件；变更概述见 docs/MODIFICATIONS.md。
//go:build !windows

package main

func accountWindowHandles() map[uintptr]bool     { return nil }
func findAccountWindow(map[uintptr]bool) uintptr { return 0 }
func activateAccountWindow(uintptr) bool         { return false }
func runAccountWindow(string) int                { return 1 }
