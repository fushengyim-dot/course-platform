# Project rules
本规则仅作用于本小组项目，不修改全局规则或个人项目模板。
开工先读 TASKS.md（若有）、README.md、当前openspec变更及 docs/standards/ 全部五份规范。
先规约后实现；禁止整套迁移个人项目、读取个人凭据、调用额外外部服务、远端创建/push或公网部署。
接口层→业务层→repository接口；adapter实现接口，cmd组装。规范冲突先修代码，例外必须写明范围与验证依据。
每次实现后使用 docs/workflows/verification.md；版本检查使用 docs/workflows/release.md。
失败检查不得改成通过；内部记录在internal/，产品正文和界面不显示制作备注。
