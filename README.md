# ReferenceAnswerRSS

将新枝（xinzhi.zone）上「参考答案阅览室」的订阅文章自动转为 RSS/Atom feed，方便导入 FreshRSS 等阅读器。

## Features

- **RSS & Atom Feed** — 标准 RSS 2.0 和 Atom 格式，带 token 认证
- **定时同步** — 每 12 小时自动从新枝 API 拉取新文章
- **Editorial 风格前端** — 编辑杂志风的优雅阅读界面
- **用户认证** — JWT 登录保护，feed token 独立管理
- **单二进制部署** — Go embed 前端，SQLite 存储，零依赖运行

## Quick Start

### Docker Compose (推荐)

```bash
# 1. 克隆项目
git clone https://github.com/XimilalaXiang/ReferenceAnswerRSS.git
cd ReferenceAnswerRSS

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env，填入 XINZHI_TOKEN 和其他配置

# 3. 启动
docker compose up -d
```

访问 http://localhost:8080 登录使用。

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `XINZHI_TOKEN` | (必填) | 新枝 CLI Token |
| `JWT_SECRET` | `change-me-in-production` | JWT 签名密钥 |
| `ADMIN_USERNAME` | `admin` | 管理员用户名 |
| `ADMIN_PASSWORD` | `admin` | 管理员密码 |
| `AUTHOR_ID` | `6905098d5f77b11d2fb2b653` | 参考答案阅览室的 Author ID |
| `SYNC_INTERVAL_HOURS` | `12` | 同步间隔（小时） |
| `BASE_URL` | `http://localhost:8080` | 服务外部访问地址 |
| `PORT` | `8080` | 服务端口 |
| `DATABASE_PATH` | `./data/rss.db` | SQLite 数据库路径 |

### RSS 订阅

登录后在 Settings 页面获取 RSS feed URL：

```
https://your-domain.com/feed.xml?token=YOUR_FEED_TOKEN
https://your-domain.com/feed.atom?token=YOUR_FEED_TOKEN
```

将上述 URL 添加到 FreshRSS 即可。

## Tech Stack

- **Backend**: Go 1.24+, SQLite, JWT
- **Frontend**: React, TypeScript, Tailwind CSS v4
- **Style**: Editorial (编辑杂志风) from [StyleKit](https://www.stylekit.top/zh/styles/editorial)
- **Deploy**: Docker, GitHub Actions

## Development

```bash
# Backend
go run ./cmd/server/

# Frontend (dev server)
cd web && npm run dev
```

## License

Apache License 2.0
