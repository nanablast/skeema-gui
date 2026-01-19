# SyncForge

**免费开源的数据库结构对比与数据同步工具**

一款强大的跨平台图形化工具，用于数据库结构比对和数据同步。**Navicat 同步功能的免费替代品**，适用于数据库迁移、结构对比、数据同步等场景。

[English](README.md)

![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-blue)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vue.js)
![License](https://img.shields.io/badge/license-MIT-green)

## 为什么选择 SyncForge？

- **免费开源** - 无订阅费用，无功能限制
- **跨平台** - 原生支持 macOS、Windows、Linux
- **多数据库支持** - MySQL、PostgreSQL、SQLite、SQL Server
- **现代化界面** - 简洁直观的深色主题 UI
- **中英双语** - 支持中文和英文界面切换

## 功能特性

### 结构比对 (Schema Compare)
- 比较两个数据库之间的表结构差异
- 检测新增、删除、修改的表和列
- 自动生成 ALTER TABLE、CREATE TABLE、DROP TABLE 语句
- 支持一键执行或选择性应用

### 数据同步 (Data Sync)
- 基于主键的行级数据差异对比
- 选择性同步：可单独选择 INSERT、UPDATE、DELETE 操作
- 批量处理，实时进度显示
- 执行前预览 SQL

### 表浏览器 (Table Browser)
- 浏览表结构（列、索引、键）
- 分页浏览表数据
- 查看建表语句
- 源库/目标库快速切换

### 连接管理
- 保存和管理多个数据库连接
- 快速连接常用数据库
- 支持不同数据库类型

## 截图

<!-- 在此添加截图 -->

## 安装

### 下载预编译版本

从 [Releases](https://github.com/nanablast/syncforge/releases) 下载适合你平台的版本。

| 平台 | 下载文件 |
|------|----------|
| macOS (Intel + Apple Silicon) | `syncforge-darwin-universal.zip` |
| Windows 64位 | `syncforge-windows-amd64.exe` |
| Linux 64位 | `syncforge-linux-amd64` |

### 从源码构建

**前置要求：**
- Go 1.21+
- Node.js 18+
- Wails CLI

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 克隆仓库
git clone https://github.com/nanablast/syncforge.git
cd syncforge

# 构建当前平台版本
wails build

# 或以开发模式运行
wails dev
```

**跨平台构建：**

```bash
wails build -platform darwin/universal    # macOS 通用版
wails build -platform windows/amd64       # Windows 64位
wails build -platform linux/amd64         # Linux 64位
```

## 快速开始

### 1. 连接数据库

1. 点击顶部状态栏的连接状态
2. 输入源数据库连接信息（要同步的数据来源）
3. 输入目标数据库连接信息（要同步到的目标）
4. 点击连接并选择数据库
5. 使用 💾 按钮保存常用连接

### 2. 结构比对

1. 切换到 **结构对比** 标签
2. 点击 **对比结构**
3. 查看差异（绿色=新增，红色=删除，黄色=修改）
4. 单独执行或批量执行 SQL

### 3. 数据同步

1. 切换到 **数据同步** 标签
2. 点击 **刷新** 加载表
3. 选择要对比的表（Shift+点击可范围选择）
4. 点击 **对比** 查找差异
5. 选择操作类型（INSERT/UPDATE/DELETE）
6. 执行同步

## 支持的数据库

| 数据库 | 结构比对 | 数据同步 | 表浏览器 |
|--------|---------|---------|---------|
| MySQL | ✅ | ✅ | ✅ |
| PostgreSQL | ✅ | ✅ | ✅ |
| SQLite | ✅ | ✅ | ✅ |
| SQL Server | ✅ | ✅ | ✅ |

## 使用场景

- **数据库迁移** - 在不同环境间迁移结构和数据
- **开发同步** - 保持 开发/测试/生产 环境数据库同步
- **结构版本控制** - 部署前对比结构变更
- **数据备份** - 选择性数据同步
- **差异对比** - 快速发现两个数据库的差异

## 技术栈

- **后端：** Go + [Wails](https://wails.io/)
- **前端：** Vue 3 + TypeScript + Vite
- **数据库驱动：**
  - MySQL (go-sql-driver/mysql)
  - PostgreSQL (lib/pq)
  - SQLite (mattn/go-sqlite3)
  - SQL Server (denisenkom/go-mssqldb)

## 配置文件

保存的连接存储在：
- **macOS/Linux：** `~/.syncforge/connections.json`
- **Windows：** `C:\Users\{用户名}\.syncforge\connections.json`

## 与其他工具对比

| 功能 | SyncForge | Navicat | DBeaver |
|------|-----------|---------|---------|
| 价格 | **免费** | ¥1000+ | 免费/付费 |
| 结构比对 | ✅ | ✅ | 需插件 |
| 数据同步 | ✅ | ✅ | 有限支持 |
| 跨平台 | ✅ | ✅ | ✅ |
| 开源 | ✅ | ❌ | ✅ |
| 原生应用 | ✅ | ✅ | Java |

## 贡献

欢迎贡献代码！请随时提交 Pull Request。

## 关键词

数据库同步, 结构对比, 数据同步工具, 数据库迁移, MySQL同步, PostgreSQL同步, SQLite同步, SQL Server同步, Navicat替代, Navicat免费替代品, 数据库对比, 表结构对比, 数据库工具, 跨平台数据库工具, 免费数据库工具, 开源数据库同步, 数据库差异对比

## 许可证

MIT License - 可免费用于个人和商业用途。

---

**如果觉得有用，请给个 Star ⭐！**
