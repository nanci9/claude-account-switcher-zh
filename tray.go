// Modified 2026-10-29: local Simplified Chinese UI; account/session logic unchanged.
package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"fyne.io/systray"
)

const maxAccounts = 8

var thresholds = []float64{90, 94, 96, 98}

type tray struct {
	s          *switcher
	window     *accountWindow
	status     *systray.MenuItem
	update     *systray.MenuItem
	accounts   [maxAccounts]*systray.MenuItem
	auto       *systray.MenuItem
	spend      *systray.MenuItem
	thresholds []*systray.MenuItem
	add        *systray.MenuItem
	importNow  *systray.MenuItem
	autostart  *systray.MenuItem
	lastIcon   string
	updates    chan struct{}
}

func (t *tray) onReady() {
	systray.SetTooltip("Claude 账号切换器（简体中文）")
	t.setIcon(-1, defaultThreshold)
	title := systray.AddMenuItem("Claude 账号切换器 · 登录保护版", "基于 v"+version+" 的本地中文修复构建（2026-11-02）")
	title.Disable()
	systray.AddSeparator()
	openWindow := systray.AddMenuItem("打开账号管理窗口", "查看账号状态并切换")

	t.status = systray.AddMenuItem("", "")
	t.status.Disable()
	t.status.Hide()
	t.update = systray.AddMenuItem("", "下载官方版本并重启；更新会覆盖中文界面和本地登录保护修复")
	t.update.Hide()
	for i := range t.accounts {
		t.accounts[i] = systray.AddMenuItemCheckbox("", "将 Claude 切换到此账号", false)
		t.accounts[i].Hide()
	}
	systray.AddSeparator()
	t.auto = systray.AddMenuItemCheckbox("额度将用尽时自动切换", "切换到剩余额度最多的账号", false)
	limit := systray.AddMenuItem("自动切换阈值", "当前额度使用达到此比例时触发")
	for _, v := range thresholds {
		t.thresholds = append(t.thresholds, limit.AddSubMenuItemCheckbox(fmt.Sprintf("%.0f%%", v), "", false))
	}
	t.spend = systray.AddMenuItemCheckbox("自动切换至周额度即将重置的账号", "切换到周额度将在一天内重置、且剩余额度至少一半的账号", false)
	t.add = systray.AddMenuItem("添加账号…", "保存当前账号登录信息，并重启 Claude 进入登录界面")
	t.importNow = systray.AddMenuItem("同步所有账号的开发会话", "确认已进入聊天主页后使用：正常退出 Claude，保存登录并同步本地 Code 会话")
	systray.AddSeparator()
	// Rewriting an enabled entry keeps its path current and adds the Task
	// Manager record that entries from before v0.1.2 are missing.
	if autostartEnabled() {
		if err := setAutostart(true); err != nil {
			log.Printf("refresh start at login: %v", err)
		}
	}
	t.autostart = systray.AddMenuItemCheckbox("登录 Windows 时启动", "", autostartEnabled())
	folder := systray.AddMenuItem("打开数据文件夹", "")
	quit := systray.AddMenuItem("退出切换器", "仅退出切换器，不关闭 Claude")

	for i, item := range t.accounts {
		go func() {
			for range item.ClickedCh {
				snap := t.s.snapshot()
				if i < len(snap.Accounts) && snap.Accounts[i].UUID != snap.Active {
					go t.s.run("Switch", func() error { return t.s.switchTo(snap.Accounts[i].UUID, byHand) })
				}
				t.s.onChange()
			}
		}()
	}
	for i, item := range t.thresholds {
		go func() {
			for range item.ClickedCh {
				t.s.setThreshold(thresholds[i])
			}
		}()
	}
	go func() {
		for {
			select {
			case <-openWindow.ClickedCh:
				if t.window != nil {
					t.window.open()
				}
			case <-t.auto.ClickedCh:
				t.s.setAuto(!t.auto.Checked())
			case <-t.spend.ClickedCh:
				t.s.setSpendWeekly(!t.spend.Checked())
			case <-t.add.ClickedCh:
				go t.s.run("Add account", t.s.addAccount)
			case <-t.importNow.ClickedCh:
				go t.s.run("Import", t.s.importChats)
			case <-t.autostart.ClickedCh:
				if err := setAutostart(!t.autostart.Checked()); err != nil {
					t.s.setStatus("Start at login: " + err.Error())
				}
				setChecked(t.autostart, autostartEnabled())
			case <-t.update.ClickedCh:
				if v := t.s.snapshot().Update; v != "" {
					go t.s.installUpdate(context.Background(), v)
				}
			case <-folder.ClickedCh:
				shellOpen(t.s.dir)
			case <-quit.ClickedCh:
				systray.Quit()
			}
		}
	}()

	// Menu updates arrive from worker goroutines; one goroutine applies them.
	t.s.onChange = func() {
		select {
		case t.updates <- struct{}{}:
		default:
		}
	}
	go func() {
		for range t.updates {
			t.render()
		}
	}()
	t.s.onChange()
	go t.loop()
	go t.updateLoop()
}

// loop runs refresh every few seconds while a new sign-in is expected,
// otherwise every minute; refresh decides which accounts are due.
func (t *tray) loop() {
	var last time.Time
	for ; ; time.Sleep(5 * time.Second) {
		if !t.s.adding() && time.Since(last) < hotEvery {
			continue
		}
		last = time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		t.s.refresh(ctx)
		cancel()
	}
}

// updateLoop looks for a new release a minute after start, then every hour;
// checkUpdate decides whether a day has passed since the last answer.
func (t *tray) updateLoop() {
	time.Sleep(time.Minute)
	for {
		t.s.checkUpdate(context.Background())
		time.Sleep(time.Hour)
	}
}

func (t *tray) render() {
	snap := t.s.snapshot()
	now := time.Now()

	switch {
	case snap.Status != "":
		t.status.SetTitle(localizeStatusZH(snap.Status))
		t.status.Show()
	case snap.Note != "":
		t.status.SetTitle(localizeStatusZH(snap.Note))
		t.status.Show()
	default:
		t.status.Hide()
	}
	if snap.Update != "" {
		t.update.SetTitle("发现官方版本 " + snap.Update + "（更新会覆盖本地修复）")
		t.update.Show()
	} else {
		t.update.Hide()
	}
	if len(snap.Accounts) == 0 && snap.Status == "" {
		t.status.SetTitle("请打开 Claude 并登录账号")
		t.status.Show()
	}

	activePct := -1.0
	tip := "Claude 账号切换器（简体中文）"
	for i, item := range t.accounts {
		if i >= len(snap.Accounts) {
			item.Hide()
			continue
		}
		a := snap.Accounts[i]
		detail := usageTextZH(a, snap.Threshold, now)
		item.SetTitle(a.label() + menuTab + detail)
		setChecked(item, a.UUID == snap.Active)
		if snap.Busy {
			item.Disable()
		} else {
			item.Enable()
		}
		item.Show()
		if a.UUID == snap.Active {
			tip = a.label() + "\n" + detail
			if a.Usage != nil {
				activePct = a.Usage.current(now).Session
			}
		}
	}
	systray.SetTooltip(tip)
	t.setIcon(activePct, snap.Threshold)

	setChecked(t.auto, snap.AutoSwitch)
	setChecked(t.spend, snap.SpendWeekly)
	for i, item := range t.thresholds {
		setChecked(item, thresholds[i] == snap.Threshold)
	}
	for _, item := range []*systray.MenuItem{t.add, t.importNow} {
		if snap.Busy {
			item.Disable()
		} else {
			item.Enable()
		}
	}
}

// menuTab right-aligns the usage on Windows, where a tab in a menu title
// starts a right-aligned column.
var menuTab = map[bool]string{true: "\t", false: "    "}[runtime.GOOS == "windows"]

func usageText(a Account, threshold float64, now time.Time) string {
	if a.Problem != "" {
		return a.Problem
	}
	if a.Usage == nil {
		return "usage unknown"
	}
	u := a.Usage.current(now)
	parts := []string{fmt.Sprintf("5h %.0f%%", u.Session), fmt.Sprintf("week %.0f%%", u.Weekly)}
	if u.Session >= threshold && !u.SessionReset.IsZero() {
		parts[0] += " until " + resetText(u.SessionReset, now)
	}
	// The weekly reset is always shown: it decides how much of the week is left to use.
	switch {
	case u.WeeklyReset.IsZero() || !u.WeeklyReset.After(now):
	case u.Weekly >= threshold:
		parts[1] += " until " + resetText(u.WeeklyReset, now)
	default:
		parts[1] += ", resets " + resetText(u.WeeklyReset, now)
	}
	return strings.Join(parts, " · ")
}

func resetText(at, now time.Time) string {
	at, now = at.Local(), now.Local()
	switch {
	case at.YearDay() == now.YearDay() && at.Year() == now.Year():
		return at.Format("15:04")
	case at.Sub(now) < 6*24*time.Hour:
		return at.Format("Mon 15:04")
	}
	// A weekday alone would read as today for a reset a week out.
	return at.Format("Mon 2 Jan 15:04")
}

func (t *tray) setIcon(pct, threshold float64) {
	key := fmt.Sprintf("%.0f/%.0f", pct, threshold)
	if key == t.lastIcon {
		return
	}
	t.lastIcon = key
	systray.SetIcon(trayIcon(pct, threshold))
}

func setChecked(item *systray.MenuItem, on bool) {
	if on != item.Checked() {
		if on {
			item.Check()
		} else {
			item.Uncheck()
		}
	}
}
