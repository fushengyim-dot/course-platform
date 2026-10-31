# Why
版本需要可自动检查的合并条件，以及从指定版本重建和回退的可执行步骤。

## What Changes
- 为分支和 Pull request 配置测试及独立环境启动检查。
- 主干合并要求检查通过，禁止直接推送、强制推送和删除。
- 增加隔离版本验证脚本，记录当前版本、上一版本和回退后的健康检查及耗时。
- 选择独立本机入口验证，替代公网部署；不触及既有服务与数据卷。

## Capabilities
### New Capabilities
- `release-verification`: 自动合并检查及可重现的版本启动、回退验证。

## Impact
新增 .github/workflows、scripts 与验证文档，配置仓库主干规则；不修改业务数据结构。
