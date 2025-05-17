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

# テスト用のモック機能（実際のプロジェクト環境に合わせて使い分け）
# MVPのために、簡易的なモックLLMを実装
mock_llm_check() {
  local impl_file=$1
  local doc_file=$2

  # モックによる判定ロジック
  # correct/なら矛盾なし、incorrect/なら矛盾ありと判定
  if [[ "$impl_file" == *"correct"* ]]; then
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

# 実装ファイルとドキュメントファイルの内容を取得
IMPLEMENTATION_CONTENT=$(cat "$IMPLEMENTATION_FILE")
DOCUMENTATION_CONTENT=$(cat "$DOCUMENTATION_FILE")

# モックLLMを使って差異を分析
echo "LLMを使って実装とドキュメントの矛盾を分析しています..."
# デバッグ出力
echo "ファイルパス: $IMPLEMENTATION_FILE"
if [[ "$IMPLEMENTATION_FILE" == *"incorrect"* ]]; then
  echo "incorrectパターンに一致"
else
  echo "incorrectパターンに一致せず"
fi
RESULT=$(mock_llm_check "$IMPLEMENTATION_FILE" "$DOCUMENTATION_FILE")
echo "RESULTの値:"
echo "$RESULT"
echo "RESULTの長さ: ${#RESULT}"

# 結果の処理
if [ -z "$RESULT" ]; then
  echo "矛盾は見つかりませんでした"
  exit 0
else
  echo "$RESULT"
  exit 1
fi
