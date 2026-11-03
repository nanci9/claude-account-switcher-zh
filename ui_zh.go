// 中文社区版修改文件；变更概述见 MODIFICATIONS.md。
// Added 2026-10-29: local Simplified Chinese display translations.
package main

import (
	"fmt"
	"strings"
	"time"
)

// Chinese text is applied only at the display boundary. Stored states, account
// labels, authentication, session data and upstream diagnostic errors stay intact.
func localizeStatusZH(text string) string {
	switch text {
	case "signed out, add it again":
		return "登录已失效，请重新添加账号"
	case "Sign in with the other account in Claude":
		return "请在 Claude 中登录另一个账号"
	case "Account detected. Finish signing in, then choose Import all chats to save the sign-in and sync chats.":
		return "账号已识别；进入聊天主页后，请点“同步所有账号的开发会话”保存登录并同步"
	case "Importing chats", "Importing chats…":
		return "正在同步开发会话…"
	case "unknown account":
		return "未知账号"
	case "no signed-in account":
		return "没有已登录的账号"
	case "Claude did not quit":
		return "Claude 未能正常退出，已取消操作；请自行退出 Claude 后重试"
	case "could not find the Claude app":
		return "未找到 Claude 桌面应用"
	case "sign-in expired":
		return "登录已过期，请重新添加账号"
	}
	for _, prefix := range []struct{ english, chinese string }{
		{"Switching to ", "正在切换至 "},
		{"Downloading version ", "正在下载版本 "},
		{"Starting version ", "正在启动版本 "},
		{"Could not start Claude: ", "无法启动 Claude："},
		{"Start at login: ", "登录时启动设置失败："},
		{"Update failed: ", "更新失败："},
		{"Auto switch failed: ", "自动切换失败："},
		{"Switch failed: ", "账号切换失败："},
		{"Add account failed: ", "添加账号失败："},
		{"Import failed: ", "同步会话失败："},
		{"Importing chats failed: ", "同步会话失败："},
		{"backup: ", "备份失败："},
		{"save current sign-in: ", "保存当前登录失败："},
		{"restore sign-in: ", "恢复登录失败："},
	} {
		if strings.HasPrefix(text, prefix.english) {
			rest := strings.TrimPrefix(text, prefix.english)
			if strings.HasSuffix(prefix.english, ": ") {
				rest = localizeStatusZH(rest)
			}
			return prefix.chinese + rest
		}
	}
	if strings.HasPrefix(text, "no saved sign-in for ") && strings.HasSuffix(text, ", use Add account") {
		account := strings.TrimSuffix(strings.TrimPrefix(text, "no saved sign-in for "), ", use Add account")
		return "未保存 " + account + " 的登录，请使用“添加账号”"
	}
	if strings.HasPrefix(text, "Auto-switched at ") {
		rest := strings.TrimPrefix(text, "Auto-switched at ")
		if at, account, ok := strings.Cut(rest, ", "); ok && strings.HasSuffix(account, " was nearly full") {
			account = strings.TrimSuffix(account, " was nearly full")
			if account == "the last account" {
				account = "上一个账号"
			}
			return fmt.Sprintf("已于 %s 自动切换；%s 的额度接近用尽", localizeDateZH(at), account)
		}
		if at, reset, ok := strings.Cut(rest, " to use up this account's week before it resets "); ok {
			return fmt.Sprintf("已于 %s 自动切换；此账号周额度将在 %s 重置", localizeDateZH(at), localizeDateZH(reset))
		}
	}
	// Preserve unrecognized diagnostics verbatim for troubleshooting.
	return text
}

func usageTextZH(a Account, threshold float64, now time.Time) string {
	if a.Problem != "" {
		return localizeStatusZH(a.Problem)
	}
	if a.Usage == nil {
		return "额度暂不可用"
	}
	u := a.Usage.current(now)
	parts := []string{fmt.Sprintf("5 小时已用 %.0f%%", u.Session), fmt.Sprintf("本周已用 %.0f%%", u.Weekly)}
	if u.Session >= threshold && !u.SessionReset.IsZero() {
		parts[0] += "，" + resetTextZH(u.SessionReset, now) + " 重置"
	}
	if !u.WeeklyReset.IsZero() && u.WeeklyReset.After(now) {
		parts[1] += "，" + resetTextZH(u.WeeklyReset, now) + " 重置"
	}
	return strings.Join(parts, " · ")
}

func resetTextZH(at, now time.Time) string {
	at, now = at.Local(), now.Local()
	if at.YearDay() == now.YearDay() && at.Year() == now.Year() {
		return "今天 " + at.Format("15:04")
	}
	weekday := chineseWeekdays[at.Weekday()]
	if at.Sub(now) < 6*24*time.Hour {
		return weekday + " " + at.Format("15:04")
	}
	if at.Year() != now.Year() {
		return fmt.Sprintf("%d 年 %d 月 %d 日 %s %s", at.Year(), at.Month(), at.Day(), weekday, at.Format("15:04"))
	}
	return fmt.Sprintf("%d 月 %d 日 %s %s", at.Month(), at.Day(), weekday, at.Format("15:04"))
}

var chineseWeekdays = [...]string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}

// autoNote uses the upstream English resetText format; translate only its date
// fields, never an account's display name.
func localizeDateZH(text string) string {
	for _, layout := range []string{"Mon 15:04", "Mon 2 Jan 15:04"} {
		if at, err := time.ParseInLocation(layout, text, time.Local); err == nil {
			// Parsing a weekday without a year does not preserve its number.
			weekdays := [...]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
			weekday := ""
			for i, english := range weekdays {
				if strings.HasPrefix(text, english+" ") {
					weekday = chineseWeekdays[i]
					break
				}
			}
			if layout == "Mon 15:04" {
				return weekday + " " + at.Format("15:04")
			}
			return fmt.Sprintf("%d 月 %d 日 %s %s", at.Month(), at.Day(), weekday, at.Format("15:04"))
		}
	}
	return text
}
