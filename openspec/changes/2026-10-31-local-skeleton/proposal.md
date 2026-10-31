# Why
当前仅有选题资料，没有独立、可重复运行的工程。先建立能启动、能诊断的共同基线，后续业务增量才有比较依据。

## What Changes
- 建立本地独立仓库，五类项目规范及最小运行骨架。
- 实现系统状态页、配置校验、健康和存活接口、结构化日志、容器编排及测试。
- 放弃直接迁入个人项目：其数据、凭据和业务范围不属于小组工程。
- 放弃一次生成全部业务：名称与PDF功能差异未确认。本轮不实现业务、额外外部服务、远端发布。

## Capabilities
### New Capabilities
- `runtime-foundation`: 可独立启动、可观测且默认拒绝未登记API的基础工程。

## Impact
新建 frontend/backend/deploy/scripts/docs/openspec，旧PDF/HTML原地保留，不改源模板和全局规则。
