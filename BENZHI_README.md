基于 Go 实现的天文光谱线归属复核全栈 Web 项目，一款科研分析服务，处理光谱峰匹配、伪线复核与候选结论发布。

# specline 评测说明

天文光谱线归属复核服务（全栈 Web）。

## 运行契约

- **启动服务**：`/app/specline --addr :8080 --db specline.db`
- **端到端自检**：`/app/specline --smoke-test --db smoke.db`
  - 真实导入光谱峰（含一条宇宙线伪峰与一条互斥双线区峰）、估计波长漂移校准、
    生成归属候选并识别互斥、标记并排除伪线、登记反证并确认一条归属、
    冻结归属版本，关闭并重开同一数据库验证持久化与重启恢复，最终以退出码 0 结束。
  - 这是 Docker `CMD` 与双架构验证的唯一判据，**只传 flag，不传路径位置参数**。

## Docker 双架构验证

```bash
# amd64
docker buildx build --platform linux/amd64 --load -t specline:amd64 .
docker run --rm specline:amd64 --smoke-test

# arm64
docker buildx build --platform linux/arm64 --load -t specline:arm64 .
docker run --rm specline:arm64 --smoke-test
```

两项 `docker run` 均须退出码 0。

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
