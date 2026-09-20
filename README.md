# MedCheckup（体检报告数字化平台）

> 项目类型：全栈 Web 应用（医疗健康）

为体检中心提供体检套餐配置、体检结果录入、报告自动生成与解读、异常指标追踪和团检管理的全流程数字化服务。

## 快速启动（Docker Compose 一键部署，首选）

```bash
# 1. 首次启动前复制环境变量
cp .env.example .env

# 2. 启动全部服务（前端 + 后端 + PostgreSQL）
docker compose up -d

# 3. 查看健康状态
docker compose ps
```

访问地址：

- 前端：http://localhost:18941
- 后端 API：http://localhost:19941
- 健康检查：http://localhost:19941/healthz

演示账号：

| 角色 | 账号 | 密码 |
| --- | --- | --- |
| 管理员 | 13800000001 | admin123 |
| 医生 | 13800000002 | doctor123 |
| 前台 | 13800000003 | front123 |
| 体检人 | 13800000004 | examinee123 |

## 本地开发

```bash
# 前端（React 18 + TS + Vite + Ant Design）
cd frontend
npm install
npm run dev        # http://localhost:18941

# 后端（Go）
cd backend
go mod tidy
go run ./cmd/server
```

## 技术栈

| 层次 | 技术 |
| --- | --- |
| 前端 | React 18 + TypeScript + Vite + Ant Design 5 |
| 图表 | ECharts（echarts-for-react） |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | PostgreSQL 15 |
| 认证 | JWT (github.com/golang-jwt/jwt/v5) + RBAC（管理员/医生/前台/体检人） |
| PDF | github.com/jung-kurt/gofpdf（体检报告生成） |
| 校验 | github.com/go-playground/validator/v10 |

## 项目目录结构

```
lp-341/
├── docker-compose.yml        # 前端/后端/PostgreSQL 编排
├── .env.example              # 环境变量示例
├── database/init.sql         # 建表 + 种子数据（首次启动自动执行）
├── frontend/
│   ├── Dockerfile            # Node 构建 + Nginx 托管
│   ├── nginx.conf            # SPA 路由 + /api 反代
│   └── src/
│       ├── api/              # user/package/examinee/registration/examResult/report/abnormalMetric/enterprise/stats
│       ├── stores/           # authStore/userStore
│       ├── components/common/# StatusBadge/AbnormalTag/ReportStatusBadge/ImageUploader/EmptyState/RoleGuard/ErrorBoundary
│       ├── hooks/            # useReportStats/usePagination
│       ├── pages/            # Dashboard/PackageManage/RegistrationManage/ResultEntry/ReportManage/AbnormalMetricTrack/GroupOrderManage/Profile/Login
│       ├── router/           # index.tsx + guards.tsx
│       ├── utils/            # formatReferenceRange/calcAge/dateFormat/request
│       └── constants/        # report/user/errorCodes
└── backend/
    ├── cmd/server/main.go    # 入口：装配依赖、启动 Gin
    └── internal/
        ├── config/           # 环境变量解析
        ├── model/            # user/package/package_item/examinee/registration/exam_result/report/abnormal_metric/enterprise/group_order/stats
        ├── repository/       # 按实体分文件
        ├── service/          # 按实体分文件 + stats
        ├── handler/          # 按实体分文件
        ├── router/           # router.go + 按实体分文件
        ├── middleware/       # auth/rbac/error_handler/rate_limiter/request_logger/upload
        ├── dto/              # 请求/响应结构体
        ├── constants/        # report/user/error_codes/log_templates/messages
        └── util/             # jwt/logger/formatters/pdf_generator/file_upload/app_error/response
```

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| COMPOSE_PROJECT_NAME | gbcheckup | Docker Compose 项目短名 |
| DB_NAME | gbcheckup_db | 数据库名 |
| DB_USER | gbcheckup_user | 数据库用户 |
| DB_PASSWORD | gbcheckup_pwd | 数据库密码 |
| JWT_SECRET | change_me_to_a_long_random_string | JWT 签名密钥 |
| APP_CORS_ORIGINS | http://localhost:18941 | CORS 来源白名单（逗号分隔，生产请显式配置，不允许 `*`） |
| FRONTEND_PORT | 18941 | 前端对外端口 |
| BACKEND_PORT | 19941 | 后端对外端口 |
| DB_PORT | 5432 | 数据库对外端口 |

## API 接口清单

| 方法 | 路径 | 角色 | 说明 |
| --- | --- | --- | --- |
| GET | `/healthz` | 无 | 健康检查 |
| POST | `/api/v1/auth/register` | 无 | 注册（JWT） |
| POST | `/api/v1/auth/login` | 无 | 登录（JWT） |
| GET | `/api/v1/users/me` | 登录用户 | 当前用户资料 |
| PUT | `/api/v1/users/me` | 登录用户 | 更新当前用户资料 |
| POST | `/api/v1/packages` | admin/front_desk | 创建套餐 |
| GET | `/api/v1/packages` | admin/front_desk | 套餐列表 |
| GET | `/api/v1/packages/:id` | admin/front_desk | 套餐详情 |
| PUT | `/api/v1/packages/:id` | admin/front_desk | 更新套餐 |
| POST | `/api/v1/packages/:id/items` | admin/front_desk | 添加检查项目 |
| GET | `/api/v1/packages/:id/items` | admin/front_desk | 套餐项目列表 |
| PUT | `/api/v1/packages/:id/items/:itemId` | admin/front_desk | 更新检查项目 |
| DELETE | `/api/v1/packages/:id/items/:itemId` | admin/front_desk | 删除检查项目 |
| POST | `/api/v1/examinees` | admin/front_desk/doctor | 新建体检人 |
| GET | `/api/v1/examinees` | admin/front_desk/doctor | 体检人列表 |
| GET | `/api/v1/examinees/:id` | admin/front_desk/doctor | 体检人详情 |
| POST | `/api/v1/examinees/batch-import` | admin/front_desk/doctor | 团体批量导入体检人 |
| POST | `/api/v1/registrations` | admin/front_desk/doctor | 体检登记/导检 |
| GET | `/api/v1/registrations` | admin/front_desk/doctor | 登记列表 |
| GET | `/api/v1/registrations/:id` | admin/front_desk/doctor | 登记详情 |
| PUT | `/api/v1/registrations/:id/status` | admin/front_desk/doctor | 更新登记状态 |
| GET | `/api/v1/exam-results/pending` | admin/doctor | 待录入/审核工作台 |
| GET | `/api/v1/exam-results` | admin/doctor | 按登记查询结果 |
| POST | `/api/v1/exam-results/:id/enter` | admin/doctor | 录入检查结果 |
| POST | `/api/v1/exam-results/:id/review` | admin/doctor | 审核检查结果 |
| POST | `/api/v1/reports/draft` | admin/doctor | 创建/获取草稿报告 |
| GET | `/api/v1/reports` | admin/doctor | 报告列表 |
| GET | `/api/v1/reports/:id` | admin/doctor | 报告详情 |
| POST | `/api/v1/reports/:id/generate` | admin/doctor | 生成报告与 PDF |
| POST | `/api/v1/reports/:id/review` | admin/doctor | 审核报告 |
| POST | `/api/v1/reports/:id/publish` | admin/doctor | 发布报告 |
| GET | `/api/v1/reports/:id/pdf` | admin/doctor | 下载报告 PDF |
| GET | `/api/v1/abnormal-metrics` | admin/doctor/examinee | 异常指标列表 |
| PUT | `/api/v1/abnormal-metrics/:id/follow-up` | admin/doctor/examinee | 异常指标随访 |
| POST | `/api/v1/enterprises` | admin/front_desk | 创建团检企业 |
| GET | `/api/v1/enterprises` | admin/front_desk | 企业列表 |
| POST | `/api/v1/enterprises/orders` | admin/front_desk | 创建团检订单 |
| GET | `/api/v1/enterprises/orders` | admin/front_desk | 团检订单列表 |
| POST | `/api/v1/enterprises/orders/:id/deliver` | admin/front_desk | 报告批量交付 |
| GET | `/api/v1/stats/dashboard` | admin | 运营统计 |

> 除 `/healthz`、注册/登录外，其余接口需携带 `Authorization: Bearer <JWT>`；所有响应头均携带 `X-Request-ID`，响应体统一为 `{code, message, data}`。

## Docker 部署说明

- 端口映射：前端 `18941:80`、后端 `19941:8080`、数据库 `5432:5432`
- 数据卷：`db_data` 命名卷持久化
- 依赖顺序：backend 等待 db 健康；frontend 等待 backend 健康
- 常见问题：端口冲突修改 `.env`；数据重置 `docker compose down -v` 后重建

## 枚举出现位置清单

### AbnormalLevel（异常等级：mild/moderate/severe）

- 后端：`backend/internal/constants/report.go`（定义）、`backend/internal/model/abnormal_metric.go`（模型）、`backend/internal/service/exam_result_service.go`（比对/猜测等级）、`backend/internal/service/abnormal_metric_service.go`（状态校验）、`backend/internal/util/formatters.go`（中文文案）、`backend/internal/constants/log_templates.go`（日志）、`backend/internal/constants/error_codes.go`（错误码）
- 前端：`frontend/src/constants/report.ts`（定义）、`frontend/src/components/common/AbnormalTag.tsx`（着色）、`frontend/src/pages/ResultEntry.tsx`（结果判定）、`frontend/src/pages/AbnormalMetricTrack.tsx`（列表）、`frontend/src/pages/Dashboard.tsx`（异常检出统计）、`frontend/src/types/index.ts`（类型）

### ReportStatus（报告状态：draft/generated/reviewed/published）

- 后端：`backend/internal/constants/report.go`（定义）、`backend/internal/model/report.go`（模型）、`backend/internal/service/report_service.go`（状态机）、`backend/internal/repository/report_repository.go`（按状态查询）、`backend/internal/util/formatters.go`（中文文案）、`backend/internal/constants/log_templates.go`（日志）、`backend/internal/constants/error_codes.go`（错误码）
- 前端：`frontend/src/constants/report.ts`（定义）、`frontend/src/components/common/ReportStatusBadge.tsx`（徽标）、`frontend/src/pages/ReportManage.tsx`（列表/按钮显隐）、`frontend/src/types/index.ts`（类型）

### UserRole（用户角色：admin/doctor/front_desk/examinee）

- 后端：`backend/internal/constants/user.go`、`backend/internal/model/user.go`、`backend/internal/middleware/rbac.go`、`backend/internal/router/*.go`（各分组权限）、`backend/internal/util/formatters.go`
- 前端：`frontend/src/constants/user.ts`、`frontend/src/components/Shell.tsx`、`frontend/src/stores/authStore.ts`、`frontend/src/components/common/RoleGuard.tsx`、`frontend/src/router/guards.tsx`

## License

MIT
