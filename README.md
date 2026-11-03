# Claude Account Switcher 中文版 · Claude 账号切换器

面向 **Claude Desktop Windows 桌面版**的多账号管理工具：在中文界面查看账号和额度，一键切换已保存的登录状态，并尝试同步桌面版 Claude Code 的本地开发会话。

本项目是 [SnlperStripes/claude-account-switcher](https://github.com/SnlperStripes/claude-account-switcher) 的社区二次开发版，保留原项目的 Apache 2.0 许可证和作者署名。**非 Anthropic 官方产品，也不隶属于 Anthropic。**

![账号管理界面：示例账号与模拟用量](assets/account-window-demo.png)

> 截图中的账号、邮箱和用量均为模拟数据。

## 功能

- 中文图形界面与系统托盘：查看账号、登录快照状态、5 小时和每周用量及重置时间。
- 当前账号置顶；其余按 5 小时用量从低到高排序；用量同为 100% 时，重置时间较早的排在前面。
- 一键切换已保存的账号；再次打开程序时唤回已有窗口。
- 同步本地 Claude Code 开发会话与项目列表；切号中断正在执行的任务时，可以选择需要续接的会话。
- 保留原项目的可选自动切号功能；可在托盘菜单中控制。新增账号登录后，程序不会立即强制重启 Claude。
- Windows 下 Claude 未能正常退出时，取消切换而不强制结束进程，以减少登录状态尚未保存时丢失的风险。

“已保存登录”表示本机存在登录快照，**不保证会话仍有效**；过期后需要重新登录。会话同步主要针对桌面版的本地 Claude Code 开发会话，不承诺所有云端普通聊天或所有项目都能跨账号共享；切换需要重启 Claude，正在执行的任务可能中断。

## 下载和使用

1. 前往 [Releases 下载页](https://github.com/nanci9/claude-account-switcher-zh/releases/latest)，下载 `claude-account-switcher-windows-amd64.exe`。需要 Windows 10/11 64 位、Claude Desktop 和 Microsoft Edge WebView2 Runtime。
2. 在 Claude Desktop 登录第一个账号，运行下载的程序，等待账号显示在窗口中。
3. 点击「添加账号」，按提示登录其他账号；确认已进入聊天主页后，可点击「同步会话」保存当前登录并同步本地开发会话。同步会重新打开 Claude，请先保存正在进行的工作。
4. 以后在账号窗口中选择「切换账号」。也可通过系统托盘菜单打开窗口或退出切换器。

切换登录状态前请等待重要任务完成。不要把 Claude 中的“退出登录”当作切号方式：退出登录可能使已保存的登录状态失效。

Release 同时提供 `.sha256` 校验文件。发布程序目前未进行代码签名，Windows 可能显示未知发布者提示；请确认下载来源并核对哈希后再决定是否运行：

```powershell
Get-FileHash .\claude-account-switcher-windows-amd64.exe -Algorithm SHA256
```

## 数据与隐私

- 登录信息和备份保存在本机。数据文件和日志可能包含敏感信息，请勿随意分享。
- 程序访问 Anthropic 服务以读取账号和用量；检查更新时访问 GitHub。请根据自己的网络环境与服务条款使用。
- 程序切换的是 Claude Desktop 的登录状态，不切换终端里独立登录的 Claude Code CLI。
- Claude 的本地文件格式属于其内部实现，桌面版更新后可能影响切号或会话同步；重要数据请自行备份。

## 更新说明

程序从本项目检查和下载更新；也可以前往 [Releases](https://github.com/nanci9/claude-account-switcher-zh/releases/latest) 手动下载最新版。

## 从源码构建

需要 Go 1.26 或更新版本。在 Windows 终端中运行：

```powershell
go test ./...
go vet ./...
go build -trimpath -ldflags "-H=windowsgui -s -w" -o claude-account-switcher-windows-amd64.exe .
```

本分支仅发布 Windows 构建；其他系统未经过本分支的完整验证。

## 来源与许可证

- 原项目：[SnlperStripes/claude-account-switcher](https://github.com/SnlperStripes/claude-account-switcher)，作者 SnlperStripes。
- 中文版在上游基础上增加图形界面、汉化、用量排序、窗口复用和登录保护，并将更新源改为本仓库。上游 v0.3.6 的任务续接功能已合并；具体变更可查看 Git 提交记录。
- 许可证：[Apache License 2.0](LICENSE)；原作者声明见 [NOTICE](NOTICE)。分发源码或程序时请一并保留这两个文件及相应修改说明。
