# 装修预算与支出管理 API 服务

纯后端 RESTful API 服务，面向装修公司和业主提供预算编制、支出审批、供应商管理和对账结算能力。

## 快速启动（Docker）

```bash
cp .env.example .env
docker compose up -d
```

服务启动后：

- 健康检查：http://localhost:19306/healthz
- Swagger 文档：http://localhost:19306/swagger/index.html

预置账号（角色）：

| 账号 | 密码 | 角色 |
| --- | --- | --- |
| admin | Admin123! | Admin |
| finance | Finance123! | FinanceManager |
| project | Project123! | ProjectManager |
| accountant | Accountant123! | Accountant |
| owner | Owner123! | Owner |

## 技术栈

| 层次 | 技术 |
| --- | --- |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | PostgreSQL 15 |
| 缓存 | Redis 7 |
| 认证 | JWT + RBAC |
| API 文档 | Swagger (swaggo) |

## 目录结构

```
backend/
├── cmd/server/main.go            # 入口：加载配置、初始化依赖、启动服务
├── internal/
│   ├── config/                   # 环境变量配置
│   ├── model/                    # GORM 模型
│   ├── repository/               # 数据访问层
│   ├── service/                  # 业务逻辑层
│   ├── handler/                  # HTTP 处理层
│   ├── router/                   # 路由与中间件挂载
│   ├── middleware/               # 认证/鉴权/审计/限流/日志/校验
│   ├── dto/                      # 请求与响应结构体
│   └── constants/                # 枚举与错误码
├── migrations/                   # SQL 迁移脚本
├── api/openapi.yaml              # OpenAPI 描述
├── deploy/k8s.yaml               # Kubernetes 部署样例
└── Dockerfile                    # Go 多阶段构建
docker-compose.yml
.env.example
```

## 枚举位置

所有共享枚举定义在 `backend/internal/constants/enums.go`：

- `BudgetStatus`: Draft / Active / Locked / Archived
- `ExpenseStatus`: Draft / Submitted / Approved / Rejected / Paid
- `PaymentMethod`: Cash / BankTransfer / Credit / Company
- `SupplierStatus`: Active / Suspended / Blacklisted
- `SupplierCategory`: Material / Furniture / Appliance / Labor / Design / Other
- `BudgetCategory`: Design / Material / Labor / Furniture / Appliance / Contingency / Other
- `ReconciliationStatus`: Pending / Confirmed / Disputed / Resolved

## 本地开发

```bash
cd backend
go run ./cmd/server
# 本地默认监听 3000
```

## License

MIT
