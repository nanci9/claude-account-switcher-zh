# 中文版修改说明

本项目基于 SnlperStripes 的 Claude Account Switcher，保留 [LICENSE](LICENSE) 和 [NOTICE](NOTICE)。

相对于上游主线，中文版增加 Windows 中文图形账号管理窗口、汉化托盘与交互、用量排序和窗口复用；调整新增账号及正常退出 Claude 的流程以保护登录状态；调整构建、更新源、测试和文档以独立发布。上游 v0.3.6 的会话续接功能已合并并汉化。

以下文件相对本次合并的上游主线有修改或为新增文件；完整内容和以后变更请查看 Git 历史：

- `.github/ISSUE_TEMPLATE/config.yml`
- `.github/workflows/build.yml`
- `.github/workflows/test.yml`
- `.gitignore`
- `CONTRIBUTING.md`
- `README.md`
- `SECURITY.md`
- `UPSTREAM_SYNC.txt`
- `account_window.go`
- `account_window.html`
- `account_window_other.go`
- `account_window_test.go`
- `account_window_windows.go`
- `dialog_windows.go`
- `docs/.nojekyll`
- `docs/index.html`
- `docs/llms.txt`
- `docs/mark.svg`
- `docs/sitemap.xml`
- `docs/social-preview.png`
- `instance.go`
- `instance_test.go`
- `main.go`
- `packaging/README.md`
- `packaging/winget/manifests/s/SnlperStripes/ClaudeAccountSwitcher/0.3.6/SnlperStripes.ClaudeAccountSwitcher.installer.yaml`
- `packaging/winget/manifests/s/SnlperStripes/ClaudeAccountSwitcher/0.3.6/SnlperStripes.ClaudeAccountSwitcher.locale.en-US.yaml`
- `packaging/winget/manifests/s/SnlperStripes/ClaudeAccountSwitcher/0.3.6/SnlperStripes.ClaudeAccountSwitcher.yaml`
- `platform_windows.go`
- `platform_windows_test.go`
- `resume.html`
- `switcher.go`
- `switcher_test.go`
- `tray.go`
- `ui_zh.go`
- `update.go`
- `update_test.go`
- `webview_windows.go`
- `winres/winres.json`
