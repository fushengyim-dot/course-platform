# 后端开发
handler处理HTTP与参数校验→service处理状态→repository定义探测接口；adapter处理MySQL/Qdrant；cmd组装。service不得导入net/http、database/sql或adapter，handler不得导入驱动/adapter。scripts/check_architecture.py执行检查。
业务路由/api/v1/小写复数资源；探针例外为/health与/live。读GET；写POST/PUT/PATCH；每个新增写接口必须声明幂等策略与重复提交测试。基础工程无业务写接口。
统一响应{code,message,data,request_id}。OK=200，INVALID_ARGUMENT=400，UNAUTHORIZED=401，FORBIDDEN=403，NOT_FOUND=404，METHOD_NOT_ALLOWED=405，INTERNAL_ERROR=500，DEPENDENCY_UNAVAILABLE=503。新码先登记。
所有配置从环境变量读取，缺少/非法立即非零退出；秘密和配置值不输出。参数白名单在handler验证；SQL仅在adapter/repository实现中，以参数绑定，不拼用户输入。
未来事务由service定义边界，事务内不执行外部HTTP；列表稳定排序加唯一键，非法筛选拒绝，不回退全量。发现违规阻断提交并修复。
