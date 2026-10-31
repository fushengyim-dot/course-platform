# 依赖用途
React / React DOM（MIT）：状态页面；TypeScript（Apache-2.0）类型检查；Vite（MIT）前端构建。
Go标准库（BSD）：HTTP、日志、配置；go-sql-driver/mysql（MPL-2.0）：真实认证Ping，作为明确允许的数据库驱动许可例外。间接依赖 filippo.io/edwards25519（BSD-3-Clause）锁定1.1.1，已修复扫描发现的GO-2026-4503。
MySQL官方服务镜像（GPL及随附组件许可）、Qdrant（Apache-2.0）、NGINX（BSD及随附组件许可）、Node与Go工具镜像用于本地开发与运行；本轮不对外分发，正式分发前需核对镜像内完整许可证清单。
版本锁定见package-lock.json、go.sum、Dockerfile与compose.yaml。安全扫描结果另记版本检查记录，不将选定版本视作安全保证。
