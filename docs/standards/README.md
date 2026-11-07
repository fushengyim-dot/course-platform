# 工程规范索引

本目录是工程约束的权威定义。根目录[AGENTS.md](../../AGENTS.md)为开发入口；能力与场景定义在[正式规约](../../openspec/specs/)，操作顺序在[实现检查](../workflows/verification.md)和[版本检查](../workflows/release.md)。新增或修改规则时同步本表，避免重复定义。

|类别|权威规范|主要判据与检查入口|
|---|---|---|
|产品|[product.md](product.md)|页面四状态、统一组件、键盘/窄屏；前端状态测试及浏览器检查，关联runtime-foundation S8与extension-contract页面场景|
|后端|[backend.md](backend.md)|依赖方向、响应码、参数边界；scripts/check_architecture.py、后端测试与extension-contract读写场景|
|安全|[security.md](security.md)|默认拒绝、资源归属、秘密边界、依赖风险；scripts/check_secrets.py、handler测试、依赖变更/发布扫描|
|测试|[testing.md](testing.md)|正常/失败/边界、超时、断言有效性；README一键测试、scripts/check_mutations.py与场景映射|
|运维|[ops.md](ops.md)|端口隔离、健康与日志、启动/回退；Compose配置、scripts/verify_runtime.py、scripts/verify_release.py与release-verification|
|AI|[ai.md](ai.md)|AI-01—04：输入范围、输出验收、外部调用启用条件、公开记录；输入/差异人工审查与同一套实现检查|
|数据|[data.md](data.md)|DATA-01—04：合同、隔离、迁移、生命周期；接口边界/跨主体/并发测试，未来迁移与删除需独立场景验证|
|评审|[review.md](review.md)|REVIEW-01—05：变更依据、同提交检查、例外与关闭；七类代码核查及PR评审记录见[代码核查清单](../code-review/checklist.md)|

每次变更的验证记录应列出：适用规则/规约场景→测试文件或人工步骤→命令与提交→实际结果。未触及的条件性规则可标“不适用”并说明原因，不能标为“通过”。手工检查需写明检查对象和结论，不能冒充自动强制检查。

规则冲突先定位权威定义：能力行为修正规约与实现，工程约束修正对应规范，操作顺序修正工作流。例外按review.md处理。当前未接入模型服务、无生产业务数据；相关启用条件已有规范不代表功能已实现。
