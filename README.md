# 云数据库 UDB MySQL MCP

UCloud 云数据库 UDB MySQL 的 MCP 服务。在 Cursor、Claude Desktop 等 MCP 客户端中连接后，可用自然语言管理实例、备份、地域可用区、询价和规格目录。

> 云数据库 UDB MySQL 基于成熟云计算技术，提供高可用、高性能、可弹性扩展的在线数据库服务，完全兼容原生 MySQL / Percona
> 协议，并提供备份、监控、弹性扩容等能力。

[产品详情](https://www.ucloud.cn/site/product/udb.html) · [API 文档](https://docs.ucloud.cn/api/udb-api/README)

## 什么是 MCP

[模型上下文协议 (MCP)](https://modelcontextprotocol.io/introduction) 是一套开放协议，用来把大模型应用接到外部工具和数据上。本服务把 UDB MySQL API 暴露成工具，由客户端在对话中调用。

## Tools

默认 `readonly`。`readwrite` / `admin` 才会开放写操作。

| 分类  | 工具                                 | 说明                    |
|-----|------------------------------------|-----------------------|
| 查询  | `udb_mysql_list_projects`          | 列出项目                  |
|     | `udb_mysql_list_regions`           | 列出地域和可用区              |
|     | `udb_mysql_list_instances`         | 列出实例                  |
|     | `udb_mysql_get_instance`           | 查询实例详情                |
|     | `udb_mysql_get_instance_state`     | 查询实例状态                |
|     | `udb_mysql_list_db_types`          | 列出数据库版本               |
|     | `udb_mysql_list_machine_types`     | 列出机型规格                |
|     | `udb_mysql_list_param_groups`      | 列出参数组                 |
|     | `udb_mysql_describe_price`         | 创建询价                  |
|     | `udb_mysql_describe_upgrade_price` | 升降配询价                 |
|     | `udb_mysql_list_backups`           | 列出备份                  |
|     | `udb_mysql_get_backup_state`       | 查询备份状态                |
|     | `udb_mysql_get_backup_url`         | 获取备份下载地址              |
|     | `udb_mysql_get_backup_strategy`    | 查询备份策略                |
| 写入  | `udb_mysql_start_instance`         | 启动实例                  |
|     | `udb_mysql_stop_instance`          | 停止实例                  |
|     | `udb_mysql_restart_instance`       | 重启实例                  |
|     | `udb_mysql_modify_name`            | 修改实例名称                |
|     | `udb_mysql_create_instance`        | 创建实例                  |
|     | `udb_mysql_create_backup`          | 创建备份                  |
| 高风险 | `udb_mysql_resize_instance`        | 升降配，需 `confirm=true`  |
|     | `udb_mysql_reset_password`         | 重置密码                  |
|     | `udb_mysql_delete_backup`          | 删除备份，需 `confirm=true` |

## 参数获取

**密钥**：在 UCloud 控制台生成 PublicKey / PrivateKey → https://console.ucloud.cn/uapi/apikey

**地域和可用区**：调用 `udb_mysql_list_regions` 获取账号可用的 region / zone。
如北京二对应 `cn-bj2`，可用区形如 `cn-bj2-04`。也可参考：https://docs.ucloud.cn/api/summary/regionlist

## 本地安装

二选一：**Go 源码**（stdio）或 **Docker**（SSE）。不提供预编译二进制下载。

Go 源码默认是 stdio：由 MCP 客户端拉起进程，不要自己在终端里常驻运行。进程读环境变量，**不会**自动加载 `.env`。Docker 镜像默认是 SSE，密钥通过 `--env-file` 注入。

可选环境变量见 `.env.example`（`UCLOUD_PROJECT_ID`、`UCLOUD_REGION`、`UCLOUD_MODE`、`MCP_SERVER_SSE_PORT` 等）。

### 方式一：Go 源码 + stdio

1. 安装 [Go](https://go.dev/dl/) 1.26+（已安装可跳过）。
2. 在仓库根目录编译：

   ```bash
   go build -o bin/udb-mysql-mcp-server .
   ```

3. 将配置填入 MCP 客户端。把 `command` 换成 `bin/udb-mysql-mcp-server` 的绝对路径：

```json
{
  "mcpServers": {
    "udb-mysql-mcp-server": {
      "type": "stdio",
      "command": "/ABSOLUTE/PATH/TO/repo/bin/udb-mysql-mcp-server",
      "args": ["--mode", "readonly"],
      "env": {
        "UCLOUD_PUBLIC_KEY": "{Your PublicKey}",
        "UCLOUD_PRIVATE_KEY": "{Your PrivateKey}"
      }
    }
  }
}
```

### 方式二：Docker + SSE

1. 安装 [Docker](https://www.docker.com/)
2. 在仓库根目录：`cp .env.example .env`，再填入密钥
3. 构建并运行：

   ```bash
   docker build -t udb-mysql-mcp-server:latest .
   docker run --rm -it --env-file .env \
     -p 127.0.0.1:9000:9000 \
     udb-mysql-mcp-server:latest \
     --transport sse --listen 0.0.0.0:9000
   ```

   镜像默认监听 `127.0.0.1:9000`（容器内回环）。使用 `-p` 端口映射时，容器内需显式 `--listen 0.0.0.0:9000`，同时**宿主机端口必须绑定回环**（如上 `-p 127.0.0.1:9000:9000`）。

   **禁止**将服务端口直接映射到非回环地址（如 `-p 9000:9000`）或暴露到公网 / 共享内网；进程内持有云账号 API 凭据，网络传输当前无身份认证。

4. 将配置填入 MCP 客户端：

```json
{
  "mcpServers": {
    "udb-mysql-mcp-server": {
      "type": "sse",
      "url": "http://127.0.0.1:9000/sse"
    }
  }
}
```

## 许可证

Apache-2.0
