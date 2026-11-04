# 上游同步说明

本项目基于 [SnlperStripes/claude-account-switcher](https://github.com/SnlperStripes/claude-account-switcher)。2026 年 11 月合并了上游 v0.3.6 的会话续接功能，以及当时上游主线的其他提交；具体内容以 Git 历史为准。

本地仓库的 `origin` 指向中文版，`upstream` 指向原项目。后续借鉴上游功能时，先比较代码并保留中文版的界面、中文文本、登录保护和窗口复用，再运行 Windows 测试与构建。修改说明见 [MODIFICATIONS.md](MODIFICATIONS.md)。

中文版的更新检查和下载地址在 `update.go` 中，指向中文版自己的 GitHub Releases。合并上游后应检查此地址，并运行 `TestUpdateSourceIsChineseEdition`，避免更新时安装成上游原版。
