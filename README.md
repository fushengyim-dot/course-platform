# 课程项目平台基础工程

软件开发课程的可运行基础工程。目前提供系统状态页和健康检查；业务功能尚未实现。

## 用户与场景
后续业务范围由项目规约确定。本版本供开发者检查运行环境。页面仅 `/` 系统状态；未知路径返回 404。技术栈采用 React / TypeScript / Go / MySQL / Qdrant / Docker Compose。

## 从零启动
需要 Git、Docker Desktop（Linux 容器，Compose v2）。首次构建需要网络下载公开依赖，无外部业务请求。

1. `git clone https://github.com/fushengyim-dot/course-platform.git group-project`，进入 `group-project`。
2. 复制 `.env.example` 为 `.env`，填写两个不同的随机数据库密码；其他值按样例设置。不要使用真实业务账号。端口冲突时修改 `WEB_PORT`，并为并行副本设置不同 `COMPOSE_PROJECT_NAME`。
3. `docker compose up --build -d --wait`。
4. 打开 `http://127.0.0.1:18770/`；访问 `/health` 应为 HTTP 200、code 为 OK，包含版本、UTC 时间和 mysql/qdrant 状态。`/live` 只检查进程存活。

开发提交前安装本地检查钩子：`git config core.hooksPath .githooks`（需要 Python 3）；手动全历史检查：`python scripts/check_secrets.py --history`。这会检查暂存边界、已配置的本地秘密和私钥标记，不替代完整安全审计。

停止：`docker compose down`（保留数据卷）。不使用 `down -v` 清除资料。仅入口绑定本机，数据库与后端不映射宿主机端口。

## 测试
`docker compose -f compose.test.yaml run --build --rm checks` 执行后端测试、分层检查、前端类型检查与构建；退出 0 才通过，结果在控制台。测试不需要数据库凭据或额外外部服务，安装依赖后测试本身使用本机替身。`docker compose config --quiet` 检查配置。本地故障演练：`python scripts/verify_runtime.py`，需要 Python 3；该命令只操作当前 Compose 项目，会短暂停止并恢复两个依赖、重建后端，请勿在演示中途执行。正式发布前还需执行独立目录启动检查。

## 协作与版本
项目由一名开发者维护。主干 `main`，开发分支 `feature/<短横线名称>`；变更目录 `openspec/changes/YYYY-MM-DD-name`。每项改动通过分支提交 Pull request，完成检查后合并；提交信息说明改动及规约名称。版本验收后打标签。源码仓库：https://github.com/fushengyim-dot/course-platform 。仓库公开，应用仍仅在本地运行。远端分支保护尚未配置。许可：尚未选择，许可方案待确定。

## 配置与安全
配置字段见 `.env.example`。`HTTP_ADDR` 与 `API_UPSTREAM` 的端口须一致。缺少必需值时启动失败，不自动回落。密码只存本机 `.env`，不得放入日志、镜像、前端或 Git。发现泄露时由项目维护者立即撤销并轮换凭据，检查日志、镜像及提交历史中的影响范围；历史清理应保留恢复备份并协调已有克隆。

## 回退
已有两个经过验证的版本时，在独立目录检出上一标签，使用独立 Compose 项目名和端口，按上述步骤启动并核对 `/health`，再切换入口。目前仅有一个骨架版本标签，双版本回退尚未演练。数据库结构变更必须先设计兼容及恢复方案。

## 目录与验证记录
`backend/` 为分层服务，`frontend/` 为状态页，`openspec/` 为能力规约，`docs/standards/` 为工程规范，`docs/releases/` 为版本验证记录。
