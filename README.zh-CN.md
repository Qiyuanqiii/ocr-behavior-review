# Behavior Review

按次启用的 OCR 行为审阅插件：从有依据的行为约定出发，沿受影响代码的关系寻找反例，再核对证据并裁决候选。

## 宿主要求

使用 [codex/behavior-review-plugin](https://github.com/Qiyuanqiii/open-code-review/tree/codex/behavior-review-plugin)
分支的 OCR，它基于你的 `codex/plugins-tui-latest` 分支。
`ocr review --help` 应包含 `--review-plugin`。此前的插件接口只支持规则包和 MCP 工具，无法加载此插件的审阅流程。

插件沿用 OCR 已配置的模型、规则、文件选择、工具、session、评论处理和过滤。安装不会调用模型。

## 安装

```sh
git clone https://github.com/Qiyuanqiii/ocr-behavior-review.git
ocr plugin install ./ocr-behavior-review/plugin
```

也可以从发布目录安装：

```sh
ocr plugin market add behavior https://raw.githubusercontent.com/Qiyuanqiii/ocr-behavior-review/main/marketplace.json
ocr plugin install behavior-review@behavior
```

保留默认安装状态即可。每次审阅均须显式指定此插件；即使执行过 `plugin enable`，也不会自动切换审阅流程。

## 启动一次

```sh
# 预览文件，不调用模型
ocr review --review-plugin behavior-review --preview

# 审阅当前工作区
ocr review --review-plugin behavior-review

# 审阅分支并保存完整阶段记录
ocr review --from main --to HEAD --review-plugin behavior-review --format json --output behavior-review.json
```

在 TUI 中设置范围和背景后运行：

```text
/scope range main HEAD
/preview behavior-review
/review behavior-review
```

随后单独运行 `/review` 就使用普通审阅。预览不会替下一次审阅选择插件。
本版需要重新启动完整审阅，暂不支持与 `--resume` 合用。

## 审阅过程

1. 找出有依据的行为约定，并追踪调用、数据、状态和执行顺序所形成的前提。
2. 根据当前代码及其关系构造反例，优先挑战后果重大、证据薄弱的前提。初始候选为零也会继续。
3. 回应原始依据，核对每项反例，允许确认、反驳、缩小或保留未知。

宿主保留有效条目、恢复已有引用、指出需要修正的字段，并在压缩上下文时保留阶段输入。确认的问题经过 OCR 原有的评论处理和过滤。
JSON 中的 `behavior_review` 保存各阶段记录，包括反驳项、未知项、待修正项和完成状态。必需工作未完成时会返回非零退出码，并保留已经得到的结果。

此流程固定执行三个阶段，`--effort` 不会删减阶段；模型思考强度仍是独立设置。
沿用 OCR 已有的工具轮次、超时、上下文及用户指定的 token 限制，没有额外的总 token 上限。`--max-tokens-budget 0` 表示总 token 不限。
完成状态说明指定工作是否完成，不能作为代码没有缺陷的证明。

## 修改与评测

审阅原则位于 `plugin/behavior-review.md`，执行与校验由宿主的 `behavior-v1` 实现负责。
修改原则时保持通用性，历史漏检只作为回归案例。发现能力的验收应使用未参与规则调整的 PR，并用修正后的代码检查误报。

```sh
python scripts/package.py
python -m unittest discover -s tests
```

安装和流程测试只验证工程接入；发现能力需要独立实测。发布包使用 Apache-2.0 许可证。
