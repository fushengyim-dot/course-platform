# 部署运维
多阶段Dockerfile，依赖锁文件入库，镜像版本固定；发布时记录实际digest，不把版本标签当不可变digest。Compose唯一入口绑定127.0.0.1，其他服务不映射端口，项目名/卷与其他项目隔离。
显式环境配置，.env仅本地；开发/测试/未来生产用同一镜像换配置。/health免认证、检查MySQL认证Ping及Qdrant就绪，失败503；/live仅检查存活。探测超时2秒；停止等待5秒，Compose给予10秒。
日志单行JSON：time,level,module,request_id,method,route,status,duration_ms。只记录允许字段，不记录任意URL/请求内容。指标定义包括请求量、错误率、时延与依赖连通状态；当前尚未接入监控告警。
发布前按 README 演练当前与上一版本的切换。数据迁移先扩展再迁移再收缩；恢复演练通过才可称可恢复。
按`docs/workflows/release.md`检查、记录docs/releases/版本.md、再打标签；任何必需检查失败先修复。当前无业务数据；演练资源须与其他项目隔离。
