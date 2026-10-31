# Foundation verification

验证日期：2026-10-31。环境：Windows、Docker Desktop Linux 容器、Docker Compose v5.5.1；测试镜像使用 Go 1.27.3、Node 22.21.1。运行项目为 `foundation-independent-runtime`，检查项目为 `foundation-independent-check`，唯一宿主入口为 `127.0.0.1:19871`。

## 结果

隔离读写示例、共享页面自动化检查及容器运行检查通过。首页与详情页已完成实际浏览器导航、375像素窄屏、键盘操作、故障与重试恢复检查。

|要求|实现和证据|结果|
|---|---|---|
|两页复用|`StatusPage` 的首页与详情页使用同一 `StatusPanel`；渲染测试检查页面标题、导航、四状态、版本、时间、依赖与加载禁用|自动化通过|
|页面路由|NGINX 明确登记 `/status/details`；实际 GET `/`、`/status/details` 均200，未知页面404|通过|
|创建和读取|`TestNormalCreateRead` 核对201/CREATED、200/OK、字段、请求ID与标题去空白|通过|
|输入边界|`TestBoundaryInvalidInput`、`TestBoundaryValidInput`：空值、非法JSON、未知字段、尾随值、超长标题/请求体；80码点中文和表情、1024字节、128字符键|通过|
|身份与归属|`TestFailureUnauthorized` 验证无效身份不访问仓库；`TestFailureOwnership` 验证他人及不存在记录404|通过|
|存储失败|`TestFailureStorage` 核对503及不暴露原始错误|通过|
|并发幂等|`TestNormalConcurrentIdempotency` 同时发起32次请求，核对相同ID与仅一条记录；同键不同内容409；不同主体独立|通过|
|生产隔离|`TestFailureProductionIsolation` 验证七种HTTP方法保持401；镜像只复制cmd/internal，不包含examples|通过|
|分层与违规检查|六项违规注入：生产反向依赖、行内样式、生产拒绝判断、示例反向依赖、示例身份判断、示例归属判断；均被拒绝并恢复。授权注入要求测试断言失败，编译错误不算通过|通过|
|构建|Go测试（随机顺序）、go vet、5项前端测试、TypeScript类型检查及Vite构建，最终命令退出0|通过|
|规约|两项变更的OpenSpec严格校验|通过|
|浏览器|首页与详情页显示正常；返回/详情链接可用；375像素布局无横向截断；Tab/Enter可触发重试；依赖故障后显示错误，恢复后重试正常|通过|

## 骨架一致性审查

- S1、S3—S7：现有Go测试覆盖健康、故障、默认拒绝、未知路径、输入与日志、超时；实际容器健康为200，依赖断开后为503，存活探针仍200，依赖恢复后健康重新200。
- S2：发现Qdrant URL端口没有范围验证，现已补充0、65536及空端口拒绝测试；缺少APP_VERSION的进程实际非零退出。
- S8：原分类函数会接受缺少版本、时间或完整依赖的数据。现已要求完整健康结构，畸形数据显示错误，空数据保留空状态；两页共用同一套逻辑。浏览器已补充导航及重试检查。
- S9：本次从已有独立克隆生成全新配置，启动独立网络及两个独立数据卷，没有复制其他项目数据。仅证明同一机器上的独立启动，不等于跨机器复现。
- S10：`TestNormalShutdownDrainsInflight` 验证请求完成后关闭；本次未额外测量容器收到SIGTERM后的完整5秒上限。
- S11：实际替换后端容器后，网关未重启，健康恢复200。本次没有单独测量“后端就绪到DNS缓存刷新”的精确时间，不能将此记录解释为5秒边界测量。
- 依赖故障返回耗时约2.02秒，超过2秒观察线，低于3秒失败线。正常健康请求在2秒内完成。
- 原 `scripts/verify_runtime.py` 的错误密码步骤只执行一次独立容器的 `probe`，没有先启动该容器内的HTTP服务；非零退出不足以证明MySQL认证失败。新增 `scripts/verify_extension_runtime.py` 启动错误密码实例，再实际请求其 `/health`，确认503、mysql down、qdrant up。原脚本此处不作为认证证据。

## 接口约定

读写示例使用201/CREATED、409/CONFLICT；标题按Unicode码点计数，键限定长度和字符集。方法、查询参数与无效身份分别返回405、400、401；空数据与畸形健康数据使用不同页面状态。详细场景见extension-contract规约。

## 复现

从项目根目录运行：

```powershell
docker compose -p foundation-independent-check -f compose.test.yaml run --build --rm checks
openspec validate 2026-10-31-foundation-verification --strict
openspec validate 2026-10-31-local-skeleton --strict
```

运行验证需要一个专用本地环境文件，字段参照 `.env.example`，使用全新随机数据库密码；设置 `COMPOSE_PROJECT_NAME=foundation-independent-runtime`、`WEB_PORT=19871`、`API_UPSTREAM=api:8080`，其余内部地址与Compose服务名一致。不要复用已有项目的配置和数据。以下命令中的 `local-test.env` 表示该环境文件，应放在Git忽略目录并替换为实际路径：

```powershell
docker compose -p foundation-independent-runtime --env-file local-test.env config --quiet
docker compose -p foundation-independent-runtime --env-file local-test.env up --build -d --wait
python scripts/verify_extension_runtime.py --env-file local-test.env --port 19871
docker compose -p foundation-independent-runtime --env-file local-test.env down
```

运行脚本固定操作 `foundation-independent-runtime`，会短暂停止并恢复其MySQL/Qdrant、替换后端，并创建后删除一个错误密码测试容器。使用前保证该名称仅用于本次专用测试。保留数据卷，不使用 `down -v`。

## 验证范围

接口示例仅在隔离测试环境中运行，不接入正式业务路由。现有检查不包括完整容器操作系统漏洞扫描和Go竞态检测；并发用例验证相同请求仅产生一条记录。自动检查、跨环境启动与版本回退记录见版本说明。
