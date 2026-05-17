# txt2md — 原始文本自动转 Markdown 工具

自动将纯文本文件转换为格式美观的 Markdown 文件，智能识别标题、段落、列表、代码块、引用、表格、水平线等。

## 功能特性

- 📝 **智能格式识别** — 自动检测标题、列表、代码块、引用、表格、水平线
- 🐚 **Shell 命令识别** — 自动将 `sudo`/`export`/`wget`/`source` 等命令包裹在代码块中
- 📋 **多列表格类型支持** — 支持 `1.` `2.` `-` `*` `1、` `2、` `3.1` `3.2` 等多种格式
- 📂 **批量转换** — 支持 `--recursive` 递归转换目录
- 🔧 **自定义规则** — 支持 YAML/TOML 自定义规则文件
- 🚀 **并发处理** — `--workers` 提升批量转换效率
- 🖥️ **管道支持** — `cat input.txt | txt2md convert -`
- 🌐 **多平台支持** — macOS (arm64/amd64)、Linux (arm64/amd64)、Windows
- 📚 **排版美化** — 中英文自动空格、紧凑/宽松风格

## 安装

### Homebrew (macOS/Linux)

```bash
brew install Somehow007/tap/txt2md
```

### 从源码构建

```bash
git clone https://github.com/Somehow007/txt2md.git
cd txt2md
go build -o txt2md ./cmd/txt2md
```

### 下载预编译二进制

- [txt2md-linux-amd64](https://github.com/Somehow007/txt2md/releases/latest)
- [txt2md-darwin-arm64](https://github.com/Somehow007/txt2md/releases/latest)
- [txt2md-windows-amd64.exe](https://github.com/Somehow007/txt2md/releases/latest)

## 快速开始

### 基本转换

```bash
# 转换单个文件，当前目录生成 input.md
txt2md convert input.txt
```

### 指定输出文件

```bash
txt2md convert input.txt -o output.md
```

### 管道输入

```bash
cat input.txt | txt2md convert - --stdout
```

### 批量转换

```bash
# 递归转换 ./docs/ 目录下所有 .txt 文件
txt2md convert ./docs/ --recursive --workers 8
```

### 美化输出

```bash
txt2md convert input.txt --pretty --style compact
```

## 完整用法

```
txt2md convert <input-file> [flags]

Flags:
  -o, --output string   输出文件路径 (默认: <input-name>.md 当前目录)
      --tab-width int   缩进检测的 Tab 宽度 (默认: 4)
      --pretty          开启排版美化 (中英文自动空格)
      --style string    输出风格: compact/spacious (默认: spacious)
      --diff            显示输入和输出的差异
      --force           覆盖现有输出文件
  -r, --recursive       递归转换目录下所有 .txt 文件
      --stdout          输出到终端而非文件
      --rules string    自定义规则文件路径 (YAML/TOML)
      --workers int     批量转换的并发 worker 数 (默认: 4)

Other Commands:
  config                管理配置文件 (--init 初始化默认配置)
  help                  帮助
  version               版本信息
```

## 配置文件

初始化配置文件：

```bash
txt2md config --init
```

生成的 `.txt2md.yaml`：

```yaml
tab-width: 4
style: spacious
pretty: false
```

## 识别格式

### 标题

- `# Heading 1` → `# Heading 1`
- `一、标题` → `## 一、标题`
- `2 标题` → `## 2 标题`
- 空行包围的短行 → `### 标题`

### 列表

```
1、第一项
2、第二项
```
→
```
1. 第一项
2. 第二项
```

```
3.1 子项1
3.2 子项2
```
→
```
3.1 子项1
3.2 子项2
```

### 代码块

Shell 命令自动识别：

```
sudo wget https://example.com/file
export GOPATH=$HOME/go
```
→
```
sudo wget https://example.com/file
export GOPATH=$HOME/go
```

### 表格

```
Name    Age    City
-----   ---    -----
Alice   25     Beijing
Bob     30     Shanghai
```
→
```
| Name  | Age | City    |
|-------|-----|---------|
| Alice | 25  | Beijing |
| Bob   | 30  | Shanghai|
```

## 自定义规则

创建 `custom_rules.yaml`：

```yaml
heading:
  min-length: 3
  max-length: 60
  detect-uppercase: true
  detect-numbered: true
  detect-contextual: true

list:
  markers: ["-", "*", "1.", "a."]
  max-indent: 8

code:
  min-indent: 2
  min-lines: 1

table:
  min-columns: 2
  min-rows: 2

pretty:
  cjk-spacing: true
```

使用：

```bash
txt2md convert input.txt --rules custom_rules.yaml
```

## 示例文件

```
项目开发计划

一、项目背景

随着数字化转型的推进，越来越多的企业需要将传统文档转换为结构化格式。

二、技术方案

  1. 采用Go语言开发
  2. 使用cobra框架构建CLI

三、预期成果

完成v1.0版本的发布。
```

→

```
## 项目开发计划

### 一、项目背景

随着数字化转型的推进，越来越多的企业需要将传统文档转换为结构化格式。

### 二、技术方案

1. 采用Go语言开发
2. 使用cobra框架构建CLI

### 三、预期成果

完成v1.0版本的发布。
```

## 项目结构

```
cmd/txt2md/
├── main.go                # 入口
└── commands/
    ├── root.go            # 根命令
    ├── convert.go         # convert 命令
    └── config.go          # config 命令

internal/
├── scanner/               # 词法扫描
├── classifier/            # 分类器
│   └── rules/             # 规则集
├── renderer/              # Markdown 渲染
├── pipeline/              # 流水线
├── config/                # 配置
├── customrules/           # 自定义规则
└── nlp/                   # NLP 增强

testdata/                  # 测试用例
```

## 开发与测试

```bash
# 运行所有测试
go test ./... -v

# 竞态检测
go test -race ./...

# 交叉编译
GOOS=darwin GOARCH=arm64 go build -o txt2md-darwin-arm64 ./cmd/txt2md
GOOS=linux GOARCH=amd64 go build -o txt2md-linux-amd64 ./cmd/txt2md

# 使用 GoReleaser 发布
goreleaser release --snapshot --clean
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License — 详见 [LICENSE](LICENSE)
