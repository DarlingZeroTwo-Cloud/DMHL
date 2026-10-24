# DMHL — Darling Minecraft Hello Launcher

一个 Flutter UI + Go 后端的跨平台 Minecraft Java 版启动器。
目标平台：Windows / Linux（后续可扩 macOS）。

## 架构

```
dmhl/
├── cmd/server/         Go 后端入口（HTTP localhost:18923）
├── internal/
│   ├── api/             HTTP handlers
│   ├── mojang/          Mojang 版本清单与版本详情
│   ├── mirror/          BMCLAPI 国内镜像 / 华为云 JRE 镜像
│   ├── instance/        多实例管理
│   ├── launcher/         启动命令拼装
│   ├── java/            Java 探测 + Temurin JRE 自动下载
│   ├── auth/            离线 / 微软 / LittleSkin 登录
│   └── skin/            PNG 皮肤导入与管理
├── client/              Flutter UI
│   └── lib/
│       ├── main.dart
│       ├── api.dart
│       ├── theme.dart
│       └── pages/       home / versions / account / settings
└── assets/
    ├── icon.png         DMHL 白底黑字图标
    └── linux/           Linux hicolor 图标
```

## 已实现

- Mojang 全版本清单（release / snapshot / old_beta / old_alpha）
- BMCLAPI 国内镜像一键切换
- 自动探测本机 Java；缺失自动从华为云镜像下载 Temurin JRE
- 多实例管理
- 离线登录
- 主题色自定义、跟随系统配色
- Flutter 液态玻璃 UI（NavigationRail + 渐变背景）

## 构建

### 后端
```bash
go build -o dmhl-server ./cmd/server
```

### Flutter UI
```bash
cd client
flutter pub get
flutter run -d linux   # 或 -d windows
```

## 运行
```bash
./dmhl-server --port 18923
# 另开一个终端
cd client && flutter run
```

## License
MIT
