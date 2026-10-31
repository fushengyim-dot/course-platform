# Design
## Context and goals
单人成组；React/TypeScript/Go/MySQL/Qdrant/Docker Compose。中性工程标识 group-platform；不是最终产品名。
## Decisions
1. handler → service → repository接口，adapter实现依赖探测；cmd组合。业务层不导入具体驱动，拒绝handler直连数据库。选择接口注入而非全局连接，故障可在离线测试中重现。
2. MySQL执行认证后的Ping，Qdrant调用readyz；不用单纯TCP可达冒充服务可用。每次请求总探测超时2秒，不触发外部业务请求。
3. 唯一宿主入口127.0.0.1:18770，NGINX托管前端并转发探针。数据库、Qdrant和API只在项目网络中；不用固定container_name或共享卷，避免个人项目冲突。
4. /health 表示就绪，/live表示存活；健康失败返回503但存活仍200，区别依赖故障与进程故障。成功/失败统一code/message/data/request_id。
5. 环境变量显式提供版本、监听地址、MySQL与Qdrant配置；任何缺失/非法值立即非零退出，仅报告字段名，不报告值。使用真实随机本地测试密码，不提供默认密码。
6. 请求ID由服务器生成，不信任传入标识；日志只记允许字段，不记URL查询、头、正文和驱动错误。
## Scope and risks
没有用户登录和业务资源；/api/前缀一律401，其余未知路由404。后续引入资源必须逐条授权。首次下载镜像/依赖需要网络和磁盘；只用免费公开软件。正式跨会话一致性、另一机器复现、远端保护仍需独立证据。
## Verification
按spec场景执行Go离线测试、前端状态测试、架构违规注入、本机Compose启动、断开本项目依赖、干净目录复现。首个版本无上一运行标签，不能声称完成双版本回滚。
