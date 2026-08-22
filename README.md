# lichen-atlas

`lichen-atlas` 是高山地衣野外调查服务。它把采样地点、样本、现场读数、鉴定和封存过程存入 SQLite 文件，并通过 HTTP 接口供调查队使用。

## 启动

```bash
GOTOOLCHAIN=local go run . --addr :8080 --db data/lichen-atlas.db
```

服务启动后提供 `/healthz`、样本与报告 API，以及 `/` 现场查询页。
