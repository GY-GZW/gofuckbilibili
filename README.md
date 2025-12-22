# Bilibili 视频下载器（Go 版）

Copyright (C) 2025 GY-GZW

This project is licensed under the GNU Affero General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

> 一个用 Go 语言编写的命令行工具，可下载 Bilibili 视频并自动合并音视频。

本项目受 [AlqhaGo呐](https://space.bilibili.com/476956332) 的 Python 脚本启发，参考视频：[《爬取B站的视频 Bilibili》](https://www.bilibili.com/video/BV1T84y197FV/)。  
**但本工具完全用 Go 重写，支持命令行交互、流式下载、自动调用 `ffmpeg` 合并。**

---

## ✨ 特性

- 🚀 **纯 Go 实现**：无外部依赖（除 `ffmpeg` 外）
- 📥 **流式下载**：大文件不占内存
- 🔒 **自动处理防盗链**：模拟浏览器请求头
- 🧩 **自动合并音视频**：调用 `ffmpeg` 生成完整 MP4
- 🖥️ **命令行友好**：支持直接传入 B 站 URL

---

## 🛠️ 使用要求

- Go 1.16+
- [ffmpeg](https://ffmpeg.org)（必须安装并加入系统 PATH）

---

## 📦 构建

```bash
# 克隆项目
git clone https://github.com/GY-GZW/gofuckbilibili.git
cd gofuckbilibili

# 构建（可选）
go build -o gofuckbilibili .
```

---

## 使用

```bash
./gofuckbilibili -url B站视频网址
```