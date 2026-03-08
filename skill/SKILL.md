---
name: picoly
description: Skill for using `picoly` Kanban Board manager CLI to manage tasks
---

# Picolyを使ったタスク管理

## 概要

このSkillは `picoly` CLIを使ってKanbanボードでタスクを管理するための行動規範を定めます。

## セッション開始時の確認

セッション開始時、カレントディレクトリに `picoly.db` が存在する場合はPicolyでタスク管理を行います。

```bash
ls picoly.db 2>/dev/null && echo "picoly available"
```

存在する場合、以降のすべての作業をPicolyで追跡してください。

## 呼び出し元（Orchestrator）の行動規範

### タスクをSubagentに委譲するとき

1. **チケットを作成する**

   ```bash
   picoly add --title "タスクの概要" --desc "詳細な指示内容"
   ```

   出力例:
   ```json
   {"id": 3, "title": "...", "status": "todo", ...}
   ```

2. **Subagentを起動し、チケット番号を通知する**

   Subagentへの指示には必ず以下を含めてください:
   - Picolyチケット番号（例: `チケット番号: 3`）
   - `PICOLY_USER_ID` に設定すべきユーザー名（Subagentのエージェント名）
   - タスクの詳細（チケットから `picoly read <id>` で取得するよう指示）

3. **Subagent完了後、チケットの内容を確認する**

   ```bash
   picoly read <id>
   ```

   チケットのステータスとコメントを見て、次のアクションを決定します。
   - `done`: 成功。次のタスクへ
   - `cancelled`: 失敗（続行不可）。コメントを読んでリカバリーを検討
   - `problem`: 問題発生（作業は継続可能）。コメントを読んで対応を判断

## Subagentの行動規範

Subagentとして起動されたとき、チケット番号が通知された場合は以下の手順に従ってください。

### 1. チケットを読む

```bash
picoly read <チケット番号>
```

チケットの `description` に詳細な指示が含まれています。

### 2. ステータスをin_progressに変更する

作業開始前に必ずステータスを更新します。`--since` にチケットの `updated_at` を指定します。

```bash
picoly work <id> --status in_progress --since "<updated_at>"
```

### 3. 途中経過を報告する

長いタスクでは定期的にコメントで進捗を報告します。`--since` には直前の `picoly read` または `picoly work` の結果の `updated_at` を使います。

```bash
picoly work <id> --comment "○○まで完了。次は△△に取り組みます。" --since "<updated_at>"
```

### 4. タスク完了時

成功した場合:

```bash
picoly work <id> --status done --comment "完了。<結果の要約>" --since "<updated_at>"
```

問題が発生したが作業を継続できる場合:

```bash
picoly work <id> --status problem --comment "問題: <詳細>。対処中: <内容>" --since "<updated_at>"
```

問題が発生して続行不可能な場合:

```bash
picoly work <id> --status cancelled --comment "問題: <詳細>。試みた対処: <内容>" --since "<updated_at>"
```

その後、呼び出し元に結果を報告してください。

## `--since` の扱い

`--since` には直前に読んだチケットの `updated_at` 値をそのまま渡します。

```bash
# チケットを読んでupdated_atを取得
OUT=$(picoly read <id>)
TS=$(echo "$OUT" | python3 -c "import sys,json; print(json.load(sys.stdin)['updated_at'])")

# --sinceに渡す
picoly work <id> --status in_progress --since "$TS"
```

または `picoly work` の出力からも取得できます:

```bash
OUT=$(picoly work <id> --status in_progress --since "$TS")
TS=$(echo "$OUT" | python3 -c "import sys,json; print(json.load(sys.stdin)['updated_at'])")
# 以降のworkコマンドには更新後のTSを使う
```

## 環境変数

| 変数 | 説明 |
|---|---|
| `PICOLY_USER_ID` | 必須。操作するユーザー名（エージェント名を設定する） |
| `PICOLY_DB` | DBファイルパス（デフォルト: `./picoly.db`） |
| `PICOLY_EXPORT` | Markdown出力先（デフォルト: `./picoly/`） |

Subagentを起動するときは `PICOLY_USER_ID` にそのSubagentのエージェント名を設定するよう指示してください。

## コマンドリファレンス

```bash
# チケット追加（管理者以上）
picoly add --title "タイトル" --desc "説明"
picoly add --title "タイトル" --desc "説明" --assignee <user_id>
picoly add --json '{"title":"...","description":"..."}'

# チケット編集（管理者以上）
picoly edit <id> --title "新タイトル" --since "<updated_at>"

# ステータス変更・コメント追加（全ロール、god以外は--since必須）
picoly work <id> --status <status> --comment "コメント" --since "<updated_at>"
picoly work <id> --assignee <user_id> --since "<updated_at>"
# status: todo | in_progress | done | cancelled | problem
# in_progressチケットのステータス変更はassignee本人 / god / adminのみ可能

# チケット詳細取得
picoly read <id>

# チケット一覧
picoly list
picoly list --status in_progress
picoly list --sort updated_at --asc
```
