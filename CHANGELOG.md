# Changelog

本项目的所有重要变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)（SemVer）规范。


## [0.2.0] - 2026-10-08

新增 Windows 桌面版（Electron）。

### 新增

- Windows 桌面客户端，双击即用、离线可用，无需浏览器或命令行
- 运行数据存放在用户目录，首次启动自动初始化，与安装目录隔离
- 服务仅本机可访问，端口被占用时自动切换
- 退出应用时后端优雅关闭，进程无残留


## [0.1.0] - 2026-09-11

首个公开发行版本。

### 新增

- 内置 MySQL / TDengine / MinIO 备份模板，cron 自动调度
- 通用 shell 任务，支持 `${VAR}` 环境变量注入与 ssh 跨机执行
- 执行 + 检测双状态，`depends_on` 流水线编排
- Web 管理界面（GoFrame v2 + Vue3），任务导出/导入
- 飞书 Webhook 告警、探活接口
- 纯 Go SQLite，无 CGO 依赖
