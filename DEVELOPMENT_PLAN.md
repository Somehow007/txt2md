# txt2md — 原始文本自动转 Markdown 工具 · 开发计划

## 一、项目概述

开发一款命令行工具，能够自动将原始文本文件转换为格式合理、排版美观的 Markdown 文件。工具通过智能识别文本内容（标题、段落、列表、引用、代码块、表格等），自动应用 Markdown 语法，输出排版优化后的 `.md` 文件。

### 核心需求（用户确认）

1. **语言**：Go
2. **运行方式**：命令行工具，用户指定输入文件路径
3. **输出行为**：在当前工作目录生成 `.md` 文件（无需用户指定输出路径）
4. **典型用法**：
   ```bash
   txt2md convert input.txt          # → 在当前目录生成 input.md
   txt2md convert input.txt -o out.md  # 可选：指定输出文件名
   ```

---

## 二、技术可行性分析

### 2.1 文本解析算法选择

文本到 Markdown 的自动转换核心挑战是**从无格式文本推断结构**。推荐采用**分层混合架构**：

```
输入文本
  → 第一层: 快速规则引擎 (正则+状态机) — 处理 80% 的结构化 case
  → 第二层: NLP 模型 (spaCy) — 处理规则覆盖不到的段落级语义（后续版本）
  → 第三层: LLM (可选，按需调用) — 处理高度模糊的 case（后续版本）
```

纯规则引擎在第一版足以覆盖：标题检测、列表检测、代码块检测、引用检测、表格检测。

### 2.2 跨平台实现方案

选择 **Go 编译为原生静态二进制**：

| 方案                                   | 优势                  |
| -------------------------------------- | --------------------- |
| `GOOS=darwin GOARCH=arm64 go build`  | macOS (Apple Silicon) |
| `GOOS=darwin GOARCH=amd64 go build`  | macOS (Intel)         |
| `GOOS=linux GOARCH=amd64 go build`   | Linux                 |
| `GOOS=windows GOARCH=amd64 go build` | Windows               |

一次编译、三平台分发、零运行时依赖。配合 GoReleaser 可实现自动多平台构建与发布。

### 2.3 潜在技术难点及解决方案

| 难点                     | 具体表现                             | 解决方案                                                                  |
| ------------------------ | ------------------------------------ | ------------------------------------------------------------------------- |
| **标题等级推断**   | 如何区分 h1 / h2 / h3？              | 基于行长度 + 前后文密度 + 关键词位置，提供 `--heading-style` 参数       |
| **列表嵌套深度**   | 缩进是二级列表还是代码块？           | 两阶段：先扫描全文统计缩进模式，再确定深度映射；提供 `--tab-width` 参数 |
| **代码块边界**     | 无 fences 标记，如何推断代码块起止？ | 连续 N 行满足"代码特征"（操作符、特殊字符密度、缩进突变）则视为代码块     |
| **表格检测**       | 对齐表格 vs 碰巧对齐的普通文本       | 检测列间至少 2 个连续空格 + 同行多列 + 至少 2 行模式一致                  |
| **混合中英文排版** | 中文两端未留空格、全角半角混用       | 内置后处理规则层：中英文间自动插入空格（可配置开关）                      |
| **歧义消解**       | 同一行可能被多条规则匹配             | 每条规则输出置信度，多规则冲突时选高分者；低于阈值保留原文                |

---

## 三、技术栈选型

```
语言:          Go 1.22+
CLI 框架:      cobra + viper
Markdown 生成: gomarkdown/markdown
测试:          go test + testify + golden files 模式
发布:          GoReleaser (多平台二进制 + Homebrew/scoop 公式)
CI/CD:         GitHub Actions (多 OS 矩阵测试 + 自动发布)
```

---

## 四、核心模块设计

```
cmd/txt2md/
├── main.go                 # CLI 入口
├── commands/
│   ├── convert.go          # convert 子命令（核心）
│   └── config.go           # config 子命令（配置管理）
├── internal/
│   ├── scanner/            # 词法扫描层：将原始文本拆分为行级 token 流
│   │   ├── scanner.go      #   行级扫描，输出 []Line
│   │   └── models.go       #   Line, Token 类型定义
│   ├── classifier/         # 分类层：为每行/块打标签
│   │   ├── engine.go       #   分类引擎（规则链）
│   │   ├── rules/          #   规则集（每条独立一个文件）
│   │   │   ├── heading.go
│   │   │   ├── list.go
│   │   │   ├── codeblock.go
│   │   │   ├── blockquote.go
│   │   │   ├── table.go
│   │   │   ├── paragraph.go
│   │   │   └── horizontal.go
│   │   └── confidence.go   #   自信度计算 & 冲突消解
│   ├── renderer/           # 渲染层：将分类后的块渲染为 Markdown
│   │   ├── renderer.go
│   │   └── formatter.go    #   排版美化（中英文空格等后处理）
│   ├── nlp/                # NLP 辅助模块（Phase 2+）
│   │   └── segment.go
│   └── pipeline/           # 流水线编排
│       └── pipeline.go     #   scanner → classifier → renderer
├── testdata/               # golden files 测试数据
└── go.mod
```

### 设计关键点

- **管道式架构**：`scanner → classifier → renderer`，每层单一职责，替换规则或渲染后端不影响其他层
- **规则独立化**：每条分类规则实现统一接口，通过注册机制加入引擎
- **Golden files 测试**：每个测试用例的输入/期望输出保存为文件，回归测试只需 diff

### CLI 命令设计

```bash
# 核心命令：转换文件
txt2md convert <input-file>

# 选项
txt2md convert input.txt -o output.md    # 指定输出文件名
txt2md convert input.txt --tab-width 4   # 指定缩进宽度
txt2md convert input.txt --pretty        # 开启排版美化
txt2md convert input.txt --style compact # 输出风格 (compact/spacious)

# 管道模式（后续版本）
cat input.txt | txt2md convert -

# 批量转换（后续版本）
txt2md convert ./docs/ --recursive
```

---

## 五、功能优先级划分

### P0 — MVP（v0.1.0，必须实现）✅

- [x] `convert` 命令：读取指定文件，在当前目录生成同名 `.md` 文件
- [x] 5 条基本分类规则：标题 / 段落 / 无序列表 / 有序列表 / 代码块 / 引用
- [x] `-o` 参数：可选指定输出文件名
- [x] `--tab-width` 参数：指定缩进宽度
- [x] 跨平台编译验证（macOS / Linux / Windows）
- [x] 错误处理：文件不存在、无权限、编码问题等

### P1 — 重要（v0.2.0）

- [ ] 表格检测与转换
- [ ] 水平分割线检测
- [ ] 中英文排版美化（`--pretty` 开关）
- [ ] 链接/URL 自动识别
- [ ] `--style` 参数（`compact` / `spacious`）
- [ ] STDIN 管道输入模式
- [ ] 配置文件支持（`.txt2md.yaml`）

### P2 — 增强（v0.3.0）

- [ ] 批量转换模式（`--recursive`）
- [ ] NLP 段落分割增强
- [ ] 自定义规则文件支持（YAML/TOML）
- [ ] 增量转换模式（文件变化 watch）

### P3 — 扩展（后续）

- [ ] LLM 集成（`--ai` 调用 API 做语义美化）
- [ ] Web UI（在线编辑器，左右分屏实时预览）
- [ ] 桌面 GUI（Wails 或 Fyne 框架）

---

## 六、迭代开发计划

```
第 1 周: 项目骨架
  - Go module 初始化
  - cobra CLI 骨架搭建
  - Makefile + 交叉编译脚本
  - scanner 模块完成，Line/Token 模型定义
  - pipeline 串联跑通（输入→扫描→空渲染→输出）
  - GitHub Actions CI 配置

第 2 周: 核心规则 P0
  - 实现 heading / paragraph / list / codeblock / blockquote 五条规则
  - 每条规则配 golden file 测试
  - renderer 完成基本 Markdown 输出
  - 发布 v0.1.0

第 3 周: 增强功能 P1
  - 表格检测规则
  - 中英文排版美化后处理器
  - URL 识别
  - 配置文件支持
  - 管道输入模式
  - 发布 v0.2.0

第 4 周: 质量打磨 + 生态
  - 大量真实文本测试，积累 golden files
  - 错误处理完善
  - help 文档 / man page
  - GoReleaser 配置
  - Homebrew tap 创建

第 5-6 周: P2 功能
  - 批处理、自定义规则、NLP 段落分割
  - 性能优化（并发处理多文件）
  - 发布 v0.3.0

第 7 周+: P3 扩展
  - LLM 集成（feature gate 方式）
  - Web UI 原型
```

---

## 七、输出行为详细说明

用户的核心需求是"指定文件，在当前目录生成 md"：

```
# 示例 1：基本用法
$ txt2md convert ./notes/meeting.txt
✅ 已转换 → ./meeting.md

# 示例 2：指定输出名
$ txt2md convert ./notes/meeting.txt -o summary.md
✅ 已转换 → ./summary.md

# 示例 3：管道输入（生成管道输出到 stdout？用 --stdout）
$ cat meeting.txt | txt2md convert -
```

**默认行为**：

1. 输入文件 `path/to/filename.txt`
2. 取输入文件的基础名（去掉目录和扩展名），如 `filename`
3. 在当前工作目录生成 `filename.md`（不覆盖已有文件，或通过 `--force` 覆盖）

---

## 八、项目初始化检查清单

- [ ] `go mod init github.com/<user>/txt2md`
- [ ] 目录结构创建
- [ ] `main.go` + cobra 骨架
- [ ] `Makefile`：`build` / `test` / `lint` / `cross-build`
- [ ] `.github/workflows/ci.yml`：多 OS 矩阵
- [ ] `.goreleaser.yml`
- [ ] 第一条分类规则 + golden file 测试
- [ ] README.md（使用说明）
