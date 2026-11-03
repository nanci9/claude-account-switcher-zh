// 中文社区版修改文件；变更概述见 MODIFICATIONS.md。
package main

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestExistingInstanceOpensWindow(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	opened := make(chan struct{})
	go serveInstance(l, func() {}, func() { close(opened) })
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err = fmt.Fprintln(conn, "open-window"); err != nil {
		t.Fatal(err)
	}
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || response != "opened\n" {
		t.Fatalf("response %q: %v", response, err)
	}
	select {
	case <-opened:
	case <-time.After(time.Second):
		t.Fatal("window was not opened")
	}
}

func TestInstanceHello(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	quit := make(chan struct{})
	go serveInstance(l, func() { close(quit) })
	addr := l.Addr().String()

	for _, v := range []string{version, "0.0.1", "nightly"} {
		if got := hello(addr, v); got != "stay" {
			t.Fatalf("hello from %s = %q, want stay", v, got)
		}
	}
	if got := hello(addr, "99.0.0"); got != "quit" {
		t.Fatalf("hello from a newer version = %q, want quit", got)
	}
	select {
	case <-quit:
	case <-time.After(time.Second):
		t.Fatal("did not quit for a newer version")
	}
}

func TestInstanceHelloUnanswered(t *testing.T) {
	// Versions before 0.3.3 hold the port and never answer.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if got := hello(l.Addr().String(), version); got != "" {
		t.Fatalf("hello = %q from a switcher that does not answer", got)
	}
}
