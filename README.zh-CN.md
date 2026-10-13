# Behavior Review

按次选择的 OCR 行为审阅插件：从有依据的行为约定出发，沿受影响代码的关系寻找反例，再核对证据并裁决候选。
三阶段流程、报告格式、引用修复、证据校验和完成条件均由本仓库中的独立程序负责。

## 宿主要求

使用支持 `stdio-v1` 审阅插件接口的 OCR，目前对应
[codex/behavior-review-plugin](https://github.com/Qiyuanqiii/open-code-review/tree/codex/behavior-review-plugin)
分支，基于你的 `codex/plugins-tui-latest` 分支。
该接口允许其他插件定义不同的阶段、提交格式和校验逻辑。
OCR 提供模型、工具、文件选择、规则、session、评论处理和过滤能力。

安装 Go 后可以构建对应宿主：

```sh
git clone --branch codex/behavior-review-plugin --single-branch https://github.com/Qiyuanqiii/open-code-review.git ocr-behavior-host
cd ocr-behavior-host
go build -o ocr ./cmd/opencodereview
```

Windows 使用 `-o ocr.exe`。后续命令应调用这个二进制，或将其目录放在 `PATH` 最前面。
此前内置 `behavior-v1` 的 v0.1.0 接入版本需要升级。

## 安装与升级

```sh
ocr plugin market add behavior https://raw.githubusercontent.com/Qiyuanqiii/ocr-behavior-review/main/marketplace.json
ocr plugin install behavior-review@behavior
```

已从该目录安装旧版时，先升级 OCR，再执行：

```sh
ocr plugin market refresh behavior
ocr plugin update behavior-review
```

发布包包含 Windows、Linux、macOS 的 amd64 和 arm64 程序。
安装过程只复制文件，不运行程序或调用模型；明确启动审阅后，插件程序以当前用户权限运行。

## 启动一次

```sh
ocr review --review-plugin behavior-review --preview
ocr review --review-plugin behavior-review
ocr review --from main --to HEAD --review-plugin behavior-review --format json --output behavior-review.json
```

在 TUI 中设置范围和背景后，输入 `/preview behavior-review` 或 `/review behavior-review`。
预览不调用模型。下一次单独输入 `/review` 使用普通审阅；安装和全局启用都不会自动选择插件审阅流程。
本版需要重新启动完整审阅，暂不支持与 `--resume` 合用。

## 审阅过程

1. 找出有依据的行为约定，并追踪调用、数据、状态和执行顺序所形成的前提。
2. 根据当前代码及其关系构造反例，优先挑战后果重大、证据薄弱的前提。初始候选为零也继续。
3. 回应原始依据并核对反例，允许确认、反驳、缩小或保留未知。

插件保留有效条目，恢复已有引用，并指出具体需要修正的内容。必需的未知项会保持未完成状态。
OCR 在压缩上下文时保留阶段输入，将确认的问题交给原有评论处理与过滤流程，并通过已有覆盖记录和非零退出码表达未完成。

JSON 的 `review_plugin` 数组保存每阶段状态，具体分析记录位于各项的 `report`，包括约定、反例、反驳项、未知项和待修正内容。
该字段取代 v0.1.0 的 `behavior_review`。完成状态说明指定工作是否完成，不能证明代码没有缺陷。

`--effort` 不会删减声明的阶段；模型思考强度仍独立设置。
沿用 OCR 已有的工具轮次、超时、上下文及用户指定的 token 限制，没有额外的总 token 上限。
`--max-tokens-budget 0` 表示总 token 不限。

## 开发

- `plugin/behavior-review.md`：通用审阅原则。
- `cmd/behavior-review/`：阶段格式、状态、校验和进程接口。
- `plugin/ocr-plugin.json`：阶段顺序与程序声明。

插件程序只使用 Go 标准库，不依赖 OCR 内部包。安装 Go 和 Python 3 后运行：

```sh
make check test
python scripts/package.py
ocr plugin install ./dist/behavior-review-0.2.0.zip
```

打包会生成六个平台的程序、ZIP 和带 SHA-256 的发布目录。
变更发布包时使用新版本号。其他插件可参考宿主的
[通用协议文档](https://github.com/Qiyuanqiii/open-code-review/blob/codex/behavior-review-plugin/docs/review-workflow-plugins.md)。

审阅原则保持通用，历史漏检作为回归样本；发现能力的验收使用未参与规则调整的 PR，并用修正后的代码检查误报。
工程接入测试验证安装与执行兼容性，发现能力需要独立评测。

Apache-2.0 许可证。原始实现归属见 [NOTICE](NOTICE)。
