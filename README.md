# LLMを使って複数ファイル間の不一致を検出する実験

複数のファイルをLLMに渡して、それらの間の矛盾を指摘してもらう実験。

## フォルダ構成
- correct/
  - 矛盾のないファイル群が格納されている
- incorrect/
  - 矛盾のあるファイル群が格納されている
- メインプログラム
  - 起動時に、引数で2つ以上のチェック対象ファイルを受け取る
  - 矛盾点があれば、`ファイル1,ファイル2:矛盾内容` というフォーマットで、1行1矛盾点として標準出力に書き出す
  - 矛盾がなければexit 0、矛盾があればexit 1

## 使用方法

```bash
# ビルドする
go build -o check_differences

# 実行する（最低2つのファイルが必要）
./check_differences <ファイル1> <ファイル2> [<ファイル3>...]
```

環境変数:
- `LLM_MODE=mock`: モックモードで実行（デフォルト）
- `LLM_MODE=openai`: OpenAI APIを使用してLLM分析を行う
- `OPENAI_API_KEY`: OpenAI APIを使用する際に必要なAPIキー（`LLM_MODE=openai`の場合）

## メタテスト

プログラムが正しく動作するかを検証するためのメタテストも提供しています。

```bash
# メタテストを実行
node run_meta_test.js
```

メタテストでは以下のテストを行います：

1. **mockモードでのテスト**
   - correct/フォルダのファイル間で矛盾なしを検証
   - incorrect/フォルダのファイル間で矛盾ありを検証

2. **openaiモードでのテスト** (OPENAI_API_KEYが設定されている場合のみ)
   - correct/フォルダのファイル間で矛盾なしを検証
   - incorrect/フォルダのファイル間で矛盾ありを検証

各テストケースで期待される結果と実際の結果が一致した場合のみ、メタテストは成功（exit 0）となります。
テスト結果のサマリーも表示されるため、どのテストケースが成功・失敗したかが一目でわかります。

## 実行例

```
$ LLM_MODE=openai ./check_differences incorrect/*
OpenAI APIを使ってファイル間の矛盾を分析しています...
2025/05/19 12:19:49 OpenAI APIからの応答: incorrect/calculator.js,incorrect/calculator.md: 矛盾ポイント1: add関数は数値型に変換していないが、ドキュメントでは数値型に変換されると記載されている
incorrect/calculator.js,incorrect/calculator.md: 矛盾ポイント2: multiply関数の記載がドキュメントにない
incorrect/calculator.js,incorrect/calculator.md: 矛盾ポイント3: コマンドライン引数の受け付け数がドキュメントでは2つだが、実際は3つの引数を受け付けている
incorrect/calculator.js,incorrect/calculator.md: 矛盾ポイント4: モジュールエクスポートでmultiply関数も公開されているが、ドキュメントに記載されていない
incorrect/calculator.js,incorrect/calculator.md: 矛盾ポイント1: add関数は数値型に変換していないが、ドキュメントでは数値型に変換されると記載されている
incorrect/calculator.js,incorrect/calculator.md: 矛盾ポイント2: multiply関数の記載がドキュメントにない
incorrect/calculator.js,incorrect/calculator.md: 矛盾ポイント3: コマンドライン引数の受け付け数がドキュメントでは2つだが、実際は3つの引数を受け付けている
incorrect/calculator.js,incorrect/calculator.md: 矛盾ポイント4: モジュールエクスポートでmultiply関数も公開されているが、ドキュメントに記載されていない
```

```
$ LLM_MODE=openai ./check_differences correct/*
OpenAI APIを使ってファイル間の矛盾を分析しています...
2025/05/19 12:23:04 OpenAI APIからの応答: 問題なし
矛盾は見つかりませんでした
```