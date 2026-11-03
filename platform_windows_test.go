// 中文社区版修改文件；变更概述见 MODIFICATIONS.md。
package main

import (
	"os"
	"slices"
	"testing"
	"time"
)

func TestQuitDesktopProcesses(t *testing.T) {
	for _, tc := range []struct {
		name string
		pids []uint32
		gone bool
	}{
		{"not running", nil, false},
		{"normal quit", []uint32{10, 20}, true},
		{"still running", []uint32{10, 20}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requested, waited := false, false
			err := quitDesktopProcesses(tc.pids, func(pids []uint32) {
				if requested || !slices.Equal(pids, tc.pids) {
					t.Fatalf("unexpected quit request: %v", pids)
				}
				requested = true
			}, func(timeout time.Duration) bool {
				if !requested || waited || timeout != 20*time.Second {
					t.Fatalf("unexpected wait: requested=%v waited=%v timeout=%s", requested, waited, timeout)
				}
				waited = true
				return tc.gone
			})
			if len(tc.pids) == 0 {
				if requested || waited || err != nil {
					t.Fatalf("inactive app must remain untouched: requested=%v waited=%v err=%v", requested, waited, err)
				}
				return
			}
			if !requested || !waited {
				t.Fatal("did not request and wait for a normal quit")
			}
			if tc.gone && err != nil {
				t.Fatalf("normal quit failed: %v", err)
			}
			if !tc.gone && (err == nil || err.Error() != "Claude did not quit") {
				t.Fatalf("must abort while Claude is still running: %v", err)
			}
		})
	}
}

func TestStoreAppID(t *testing.T) {
	pids, exe := desktopProcesses()
	if len(pids) == 0 {
		t.Skip("Claude is not running")
	}
	id := storeAppID(exe)
	t.Logf("exe %s -> %q", exe, id)
	if _, err := os.Stat(exe); err != nil {
		t.Fatal(err)
	}
}
