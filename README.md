# specline — 天文光谱线归属复核台

面向天文学家的光谱线归属复核服务（全栈 Web）。导入校准后的光谱峰与观测元数据，
服务估计局部波长漂移、匹配元素跃迁库并生成互斥候选；复核员标记宇宙线伪线、
调整先验、记录反证并冻结一版可引用的归属版本。

## 业务闭环

导入光谱峰 → 波长漂移校准 → 元素跃迁匹配（互斥候选）→ 证据复核（伪线/反证/先验）
→ 冻结归属版本。

## 标准命令

```bash
# 构建 / 静态检查 / 测试
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...

# 启动服务
go run ./cmd/specline --addr :8080 --db specline.db

# 端到端自检（Docker CMD 判据）
go run ./cmd/specline --smoke-test --db smoke.db
```

## 主要 API（前缀 /api）

- 观测集：`POST /api/observations`、`GET /api/observations`、`GET /api/observations/{id}`、
  `POST /api/observations/{id}/publish`、`POST /api/observations/{id}/archive`
- 光谱峰：`POST /api/observations/{id}/peaks`、`GET /api/observations/{id}/peaks`、
  `POST /api/peaks/{id}/mark-artifact`、`POST /api/peaks/{id}/exclude`、`POST /api/peaks/{id}/reopen`
- 校准：`POST /api/observations/{id}/calibrate`、`GET /api/observations/{id}/calibration`
- 匹配：`POST /api/observations/{id}/match`、`GET /api/observations/{id}/candidates`、
  `GET /api/candidates/{id}`、`POST /api/candidates/{id}/confirm`、`POST /api/candidates/{id}/reject`
- 反证/先验：`POST /api/candidates/{id}/refutations`、`GET /api/candidates/{id}/refutations`、
  `PUT /api/observations/{id}/prior`
- 版本：`POST /api/versions`、`GET /api/versions?observation_id=`、`GET /api/versions/{id}`、
  `POST /api/versions/{id}/share`、`POST /api/versions/{id}/freeze`
- 统计/健康：`GET /api/stats`、`GET /api/health`
- 页面：`GET /`、`GET /observations/{id}`

## 环境

- Go 1.26.3，`CGO_ENABLED=0`，`GOPROXY=https://goproxy.cn,direct`，`GOSUMDB=sum.golang.google.cn`
- SQLite 驱动 modernc.org/sqlite v1.52.0（纯 Go），SQLite 3.46.1
