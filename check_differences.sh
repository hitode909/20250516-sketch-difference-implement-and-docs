#!/bin/zsh
set -e

# 引数チェック
if [ $# -ne 2 ]; then
  echo "使用方法: $0 <実装ファイル> <ドキュメントファイル>"
  exit 1
fi

IMPLEMENTATION_FILE=$1
DOCUMENTATION_FILE=$2

# ファイルの存在チェック
if [ ! -f "$IMPLEMENTATION_FILE" ]; then
  echo "エラー: 実装ファイル '$IMPLEMENTATION_FILE' が見つかりません"
  exit 1
fi

if [ ! -f "$DOCUMENTATION_FILE" ]; then
  echo "エラー: ドキュメントファイル '$DOCUMENTATION_FILE' が見つかりません"
  exit 1
fi

# GitHub CLIを使ってCopilotに質問する関数
gh_copilot_check() {
  local impl_file=$1
  local doc_file=$2
  local impl_content=$3
  local doc_content=$4
  
  # プロンプトの作成
  local prompt="
以下の実装ファイルとドキュメントファイルを比較し、不一致や矛盾点を見つけてください。
各矛盾点は「${impl_file},${doc_file}:矛盾内容」の形式で出力してください。
矛盾がない場合は何も出力しないでください。

== 実装ファイル (${impl_file}) ==
${impl_content}

== ドキュメントファイル (${doc_file}) ==
${doc_content}
"

  # GitHub Copilotに質問
  # gh copilot explain コマンドを使用
  echo "$prompt" | gh copilot explain
}

# 実装ファイルとドキュメントファイルの内容を取得
IMPLEMENTATION_CONTENT=$(cat "$IMPLEMENTATION_FILE")
DOCUMENTATION_CONTENT=$(cat "$DOCUMENTATION_FILE")

# GitHub Copilotを使って差異を分析
echo "GitHub Copilotを使って実装とドキュメントの矛盾を分析しています..."
RESULT=$(gh_copilot_check "$IMPLEMENTATION_FILE" "$DOCUMENTATION_FILE" "$IMPLEMENTATION_CONTENT" "$DOCUMENTATION_CONTENT")

# 結果の処理
if [ -z "$RESULT" ]; then
  echo "矛盾は見つかりませんでした"
  exit 0
else
  echo "$RESULT"
  exit 1
fi
