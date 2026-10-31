# 部署运维
多阶段Dockerfile，依赖锁文件入库，镜像版本固定；发布时记录实际digest，不把版本标签当不可变digest。Compose唯一入口绑定127.0.0.1，其他服务不映射端口，项目名/卷与个人项目隔离。
显式环境配置，.env仅本地；开发/测试/未来生产用同一镜像换配置。/health免认证、检查MySQL认证Ping及Qdrant就绪，失败503；/live仅检查存活。探测超时2秒；停止等待5秒，Compose给予10秒。
日志单行JSON：time,level,module,request_id,method,route,status,duration_ms。只记录允许字段，不记录任意URL/请求内容。后续指标为请求量、错误率、时延、依赖连通，本轮不冒称已接告警。
本地首版不宣称生产上线或双版本回滚；后续上一标签可运行后按README演练。数据迁移先扩展再迁移再收缩；恢复演练通过才可称可恢复。
按local-release工作流程检查、记录docs/releases/版本.md、再打标签；任何必需检查失败先修复。本轮无业务数据备份恢复需求，不删除既有文件或他项目容器。
