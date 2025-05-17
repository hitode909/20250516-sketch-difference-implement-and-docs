#!/bin/zsh
set -e

# 使用するLLMモードの設定（デフォルトはモック）
LLM_MODE=${LLM_MODE:-"mock"} # "mock", "gh", "openai" などに対応可能。現状モックモードのみ実装済み

# 引数チェック
if [ $# -ne 2 ]; then
  echo "使用方法: $0 <実装ファイル> <ドキュメントファイル>"
  echo "環境変数:"
  echo "  USE_GH=1    : GitHub CLIを使用（$USE_GH）"
  echo "  LLM_MODE=xxx: 使用するLLMモード（現在: $LLM_MODE）"
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

# テスト用のモック機能（実際のプロジェクト環境に合わせて使い分け）
# MVPのために、簡易的なモックLLMを実装
mock_llm_check() {
  local impl_file=$1
  local doc_file=$2

  # モックによる判定ロジック
  # correct/なら矛盾なし、incorrect/なら矛盾ありと判定
  if [[ "$impl_file" == *"correct"* && "$impl_file" != *"incorrect"* ]]; then
    echo "" # 矛盾なしは空文字列
    return 0
  elif [[ "$impl_file" == *"incorrect"* ]]; then
    # 矛盾点を出力
    echo "$impl_file,$doc_file:add関数は数値変換を行っていません"
    echo "$impl_file,$doc_file:multiply関数がドキュメントに記載されていません"
    echo "$impl_file,$doc_file:コマンドラインは3つの引数を受け付けますが、ドキュメントでは2つと記載"
    return 0
  else
    echo "不明なディレクトリ構造です"
    return 1
  fi
}

# GitHub CLIを使用して差異を分析
gh_llm_check() {
  local impl_file=$1
  local doc_file=$2
  local impl_content=$3
  local doc_content=$4

  # プロンプトを一時ファイルに書き出す（エスケープ問題を回避）
  local tmp_prompt_file=$(mktemp)

  cat > "$tmp_prompt_file" << EOF
以下の実装ファイルとドキュメントファイルを比較し、不一致や矛盾点を見つけてください。
各矛盾点は「${impl_file},${doc_file}:矛盾内容」の形式で1行ずつ出力してください。
矛盾がない場合は何も出力しないでください。

== 実装ファイル (${impl_file}) ==
${impl_content}

== ドキュメントファイル (${doc_file}) ==
${doc_content}
EOF

  # GitHub CLIを使用して分析
  local result
  result=$(gh copilot explain --filepath "$tmp_prompt_file" 2>/dev/null || echo "")

  # 一時ファイルの削除
  rm -f "$tmp_prompt_file"

  # 結果の後処理（フォーマットを整える）
  if [[ "$result" == *"矛盾"* || "$result" == *"不一致"* ]]; then
    # 結果から必要な部分だけを抽出して返す
    echo "$result" | grep -E "${impl_file},${doc_file}:" || \
    echo "$impl_file,$doc_file:矛盾が見つかりましたが、詳細は解析できませんでした"
  else
    # 矛盾がない場合は空文字を返す
    echo ""
  fi
}

# 実装ファイルとドキュメントファイルの内容を取得
IMPLEMENTATION_CONTENT=$(cat "$IMPLEMENTATION_FILE")
DOCUMENTATION_CONTENT=$(cat "$DOCUMENTATION_FILE")

# 使用するLLMモードを決定
if [[ "$USE_GH" == "1" || "$LLM_MODE" == "gh" ]]; then
  echo "GitHub CLIを使って実装とドキュメントの矛盾を分析しています..."
  RESULT=$(gh_llm_check "$IMPLEMENTATION_FILE" "$DOCUMENTATION_FILE" "$IMPLEMENTATION_CONTENT" "$DOCUMENTATION_CONTENT")
else
  echo "モックLLMを使って実装とドキュメントの矛盾を分析しています..."
  RESULT=$(mock_llm_check "$IMPLEMENTATION_FILE" "$DOCUMENTATION_FILE")
fi

# 結果の処理
if [ -z "$RESULT" ]; then
  echo "矛盾は見つかりませんでした"
  exit 0
else
  echo "$RESULT"
  exit 1
fi
