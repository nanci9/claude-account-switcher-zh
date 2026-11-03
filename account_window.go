// 中文社区版修改文件；变更概述见 MODIFICATIONS.md。
package main

import (
	"crypto/rand"
	"crypto/subtle"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type accountWindow struct {
	s       *switcher
	token   string
	address string
	mu      sync.Mutex
	handle  uintptr
	opening bool
}

func newAccountWindow(s *switcher) (*accountWindow, error) {
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return nil, err
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	w := &accountWindow{s: s, token: hex.EncodeToString(secret[:]), address: l.Addr().String()}
	server := &http.Server{Handler: w, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := server.Serve(l); err != nil && err != http.ErrServerClosed {
			log.Printf("account window: %v", err)
		}
	}()
	return w, nil
}

func (w *accountWindow) open() {
	if !w.beginOpen(activateAccountWindow) {
		return
	}
	url := "http://" + w.address + "/#" + w.token
	if runtime.GOOS == "windows" {
		edge := filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft", "Edge", "Application", "msedge.exe")
		if _, err := os.Stat(edge); err == nil {
			previous := accountWindowHandles()
			if err := exec.Command(edge, "--app="+url, "--window-size=760,620").Start(); err == nil {
				go func() { w.finishOpen(resizeAccountWindow(previous)) }()
				return
			}
		}
	}
	shellOpen(url)
	w.finishOpen(0)
}

// beginOpen serializes window creation and reuses the current live window.
func (w *accountWindow) beginOpen(activate func(uintptr) bool) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.opening || (w.handle != 0 && activate(w.handle)) {
		return false
	}
	w.handle, w.opening = 0, true
	return true
}

func (w *accountWindow) finishOpen(handle uintptr) {
	w.mu.Lock()
	w.handle, w.opening = handle, false
	w.mu.Unlock()
}

type windowAccount struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	Current      bool     `json:"current"`
	Saved        bool     `json:"saved"`
	Problem      string   `json:"problem"`
	Usage        string   `json:"usage"`
	Session      *float64 `json:"session"`
	SessionReset string   `json:"sessionReset"`
	Weekly       *float64 `json:"weekly"`
	WeeklyReset  string   `json:"weeklyReset"`
	Updated      string   `json:"updated"`
	resetAt      time.Time
}

func orderedWindowAccounts(s *switcher, snap snapshot, now time.Time) []windowAccount {
	accounts := make([]windowAccount, 0, len(snap.Accounts))
	for _, a := range snap.Accounts {
		_, err := os.Stat(filepath.Join(s.profileDir(a.UUID), "config.json"))
		item := windowAccount{ID: a.UUID, Name: a.label(), Email: a.Email, Current: a.UUID == snap.Active, Saved: err == nil, Problem: localizeStatusZH(a.Problem), Usage: usageTextZH(a, snap.Threshold, now)}
		if a.Usage != nil {
			u := a.Usage.current(now)
			if !math.IsNaN(u.Session) && !math.IsInf(u.Session, 0) {
				value := max(0.0, min(100.0, u.Session))
				item.Session = &value
			}
			if !math.IsNaN(u.Weekly) && !math.IsInf(u.Weekly, 0) {
				value := max(0.0, min(100.0, u.Weekly))
				item.Weekly = &value
			}
			if u.SessionReset.After(now) {
				item.SessionReset = resetTextZH(u.SessionReset, now)
				item.resetAt = u.SessionReset
			}
			if u.WeeklyReset.After(now) {
				item.WeeklyReset = resetTextZH(u.WeeklyReset, now)
			}
			if !u.CheckedAt.IsZero() {
				item.Updated = "更新于 " + u.CheckedAt.Local().Format("01-02 15:04")
				if now.Sub(u.CheckedAt) > 15*time.Minute {
					item.Updated += " · 可能过期"
				}
			}
		}
		accounts = append(accounts, item)
	}
	sort.SliceStable(accounts, func(i, j int) bool {
		a, b := accounts[i], accounts[j]
		if a.Current != b.Current {
			return a.Current
		}
		if (a.Session == nil) != (b.Session == nil) {
			return a.Session != nil
		}
		if a.Session == nil {
			return false
		}
		if *a.Session != *b.Session {
			return *a.Session < *b.Session
		}
		if *a.Session >= 100 {
			if a.resetAt.IsZero() != b.resetAt.IsZero() {
				return !a.resetAt.IsZero()
			}
			if !a.resetAt.IsZero() && !a.resetAt.Equal(b.resetAt) {
				return a.resetAt.Before(b.resetAt)
			}
		}
		return false
	})
	return accounts
}

func (w *accountWindow) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Host != w.address {
		http.Error(rw, "invalid host", http.StatusForbidden)
		return
	}
	rw.Header().Set("Cache-Control", "no-store")
	rw.Header().Set("X-Content-Type-Options", "nosniff")
	rw.Header().Set("Referrer-Policy", "no-referrer")
	rw.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'unsafe-inline'; script-src 'unsafe-inline'; frame-ancestors 'none'; base-uri 'none'")
	if r.URL.Path == "/" && r.Method == http.MethodGet {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(rw, accountWindowHTML)
		return
	}
	if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Window-Token")), []byte(w.token)) != 1 || (r.Method != http.MethodGet && r.Header.Get("Origin") != "http://"+w.address) {
		http.Error(rw, "forbidden", http.StatusForbidden)
		return
	}
	if r.URL.Path == "/api/state" && r.Method == http.MethodGet {
		snap := w.s.snapshot()
		accounts := orderedWindowAccounts(w.s, snap, time.Now())
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(map[string]any{"accounts": accounts, "busy": snap.Busy, "status": localizeStatusZH(snap.Status)})
		return
	}
	if r.URL.Path == "/api/action" && r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(rw, r.Body, 4096)
		var body struct {
			Action string `json:"action"`
			ID     string `json:"id"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			http.Error(rw, "invalid request", 400)
			return
		}
		if w.s.busy.Load() {
			http.Error(rw, "正在处理其他操作", 409)
			return
		}
		switch body.Action {
		case "switch":
			found := false
			for _, a := range w.s.snapshot().Accounts {
				if a.UUID == body.ID {
					found = true
					break
				}
			}
			if !found || strings.TrimSpace(body.ID) == "" {
				http.Error(rw, "账号不存在", 400)
				return
			}
			go w.s.run("Switch", func() error { return w.s.switchTo(body.ID, byHand) })
		case "add":
			go w.s.run("Add account", w.s.addAccount)
		case "sync":
			go w.s.run("Import", w.s.importChats)
		default:
			http.Error(rw, "操作不存在", 400)
			return
		}
		rw.WriteHeader(http.StatusAccepted)
		return
	}
	http.NotFound(rw, r)
}

//go:embed account_window.html
var accountWindowHTML string
