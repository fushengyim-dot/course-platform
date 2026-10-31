# 后端开发
handler处理HTTP与参数校验→service处理状态→repository定义探测接口；adapter处理MySQL/Qdrant；cmd组装。service不得导入net/http、database/sql或adapter，handler不得导入驱动/adapter。scripts/check_architecture.py执行检查。
业务路由/api/v1/小写复数资源；探针例外为/health与/live。读GET；写POST/PUT/PATCH；每个新增写接口必须声明幂等策略与重复提交测试。基础工程无业务写接口。
统一响应{code,message,data,request_id}。OK=200，INVALID_ARGUMENT=400，UNAUTHORIZED=401，FORBIDDEN=403，NOT_FOUND=404，METHOD_NOT_ALLOWED=405，INTERNAL_ERROR=500，DEPENDENCY_UNAVAILABLE=503。新码先登记。
所有配置从环境变量读取，缺少/非法立即非零退出；秘密和配置值不输出。参数白名单在handler验证；SQL仅在adapter/repository实现中，以参数绑定，不拼用户输入。
未来事务由service定义边界，事务内不执行外部HTTP；列表稳定排序加唯一键，非法筛选拒绝，不回退全量。发现违规阻断提交并修复。

隔离记录示例不挂载正式服务：创建及幂等重放均返回201/CREATED，冲突返回409/CONFLICT，读取成功200/OK；其余沿用统一错误码与Envelope。身份由测试注入解析器，空身份拒绝。title按Unicode码点计数，TrimSpace后1—80；键限1—128个ASCII字母、数字、横线或下划线。示例拒绝查询参数；已登记路径不支持的方法405，未登记路径404。每次响应生成请求ID，不回显内部错误。
