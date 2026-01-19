# SyncForge

**Free, Open-Source Database Schema Comparison & Data Synchronization Tool**

A powerful cross-platform GUI application for comparing and synchronizing database schemas and data. The perfect **Navicat alternative** for database migration, schema diff, and data sync tasks.

[中文文档](README.zh-CN.md)

![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-blue)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vue.js)
![License](https://img.shields.io/badge/license-MIT-green)

## Why SyncForge?

- **Free & Open Source** - No subscription fees, no feature limitations
- **Cross-Platform** - Native apps for macOS, Windows, and Linux
- **Multi-Database Support** - MySQL, PostgreSQL, SQLite, SQL Server
- **Modern UI** - Clean, intuitive interface with dark theme
- **Bilingual** - English and Chinese (中文) interface

## Features

### Schema Comparison
- Compare table structures between two databases
- Detect added, removed, and modified tables/columns
- Generate ALTER TABLE, CREATE TABLE, DROP TABLE statements
- One-click execution or selective application

### Data Synchronization
- Compare row-level data differences using primary keys
- Selective sync: choose INSERT, UPDATE, or DELETE operations
- Batch processing with progress tracking
- Preview SQL before execution

### Table Browser
- Browse table structures (columns, indexes, keys)
- View table data with pagination
- View CREATE TABLE statements
- Switch between source and target databases

### Connection Management
- Save and manage multiple database connections
- Quick connect with saved credentials
- Support for different database types

## Screenshots

<!-- Add screenshots here -->

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from [Releases](https://github.com/nanablast/syncforge/releases).

| Platform | Download |
|----------|----------|
| macOS (Intel + Apple Silicon) | `syncforge-darwin-universal.zip` |
| Windows 64-bit | `syncforge-windows-amd64.exe` |
| Linux 64-bit | `syncforge-linux-amd64` |

### Build from Source

**Prerequisites:**
- Go 1.21+
- Node.js 18+
- Wails CLI

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone the repository
git clone https://github.com/nanablast/syncforge.git
cd syncforge

# Build for current platform
wails build

# Or run in development mode
wails dev
```

**Cross-platform builds:**

```bash
wails build -platform darwin/universal    # macOS (Universal)
wails build -platform windows/amd64       # Windows 64-bit
wails build -platform linux/amd64         # Linux 64-bit
```

## Quick Start

### 1. Connect to Databases

1. Click the connection status in the top bar
2. Enter Source database credentials (the database you want to sync FROM)
3. Enter Target database credentials (the database you want to sync TO)
4. Click Connect and select a database
5. Save connections with 💾 for quick access

### 2. Compare Schemas

1. Go to **Schema Compare** tab
2. Click **Compare Schemas**
3. Review differences (green = add, red = remove, yellow = modify)
4. Execute SQL statements individually or all at once

### 3. Sync Data

1. Go to **Data Sync** tab
2. Click **Refresh** to load tables
3. Select tables to compare (Shift+click for range selection)
4. Click **Compare** to find differences
5. Choose operations (INSERT/UPDATE/DELETE)
6. Execute synchronization

## Supported Databases

| Database | Schema Compare | Data Sync | Table Browser |
|----------|---------------|-----------|---------------|
| MySQL | ✅ | ✅ | ✅ |
| PostgreSQL | ✅ | ✅ | ✅ |
| SQLite | ✅ | ✅ | ✅ |
| SQL Server | ✅ | ✅ | ✅ |

## Use Cases

- **Database Migration** - Migrate schema and data between environments
- **Development Sync** - Keep dev/staging/prod databases in sync
- **Schema Versioning** - Compare schema changes before deployment
- **Data Backup** - Selective data synchronization
- **Database Diff** - Find differences between databases

## Tech Stack

- **Backend:** Go + [Wails](https://wails.io/)
- **Frontend:** Vue 3 + TypeScript + Vite
- **Database Drivers:**
  - MySQL (go-sql-driver/mysql)
  - PostgreSQL (lib/pq)
  - SQLite (mattn/go-sqlite3)
  - SQL Server (denisenkom/go-mssqldb)

## Configuration

Saved connections are stored in:
- **macOS/Linux:** `~/.syncforge/connections.json`
- **Windows:** `C:\Users\{username}\.syncforge\connections.json`

## Comparison with Other Tools

| Feature | SyncForge | Navicat | DBeaver |
|---------|-----------|---------|---------|
| Price | **Free** | $$$$ | Free/Paid |
| Schema Compare | ✅ | ✅ | Plugin |
| Data Sync | ✅ | ✅ | Limited |
| Cross-Platform | ✅ | ✅ | ✅ |
| Open Source | ✅ | ❌ | ✅ |
| Native App | ✅ | ✅ | Java |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Keywords

database sync, schema compare, data synchronization, database migration, MySQL sync, PostgreSQL sync, SQLite sync, SQL Server sync, Navicat alternative, database diff, schema diff, table compare, database tool, cross-platform database, free database tool, open source database sync

## License

MIT License - Free for personal and commercial use.

---

**Star ⭐ this repo if you find it useful!**
