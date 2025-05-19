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
2025/05/19 21:42:58 OpenAI APIからの応答を受信しました
分析結果: JavaScriptファイルとドキュメントの間にいくつかの矛盾があります。特に、関数の動作とドキュメントの記載に関する差異が見られます。
incorrect/calculator.js,incorrect/calculator.md:add関数はドキュメントでは数値型に変換されると記載されていますが、コードでは変換されていません。そのため、文字列として連結される可能性があります。
incorrect/calculator.js,incorrect/calculator.md:multiply関数がcalculator.jsに存在しますが、calculator.mdには記載されていません。このため、ドキュメントからはこの関数の存在がわかりません。
incorrect/calculator.js,incorrect/calculator.md:使用例の項目で、コマンドライン引数がJavaScriptファイルでは2または3個必要とされていますが、ドキュメントでは2個のみが受け入れられると記載されています。
incorrect/calculator.js,incorrect/calculator.md:モジュールとしてエクスポートする際に、calculator.jsにはmultiply関数もエクスポートされていますが、calculator.mdにはそのような記載がありません。
```

```
$ LLM_MODE=openai ./check_differences correct/*
OpenAI APIを使ってファイル間の矛盾を分析しています...
2025/05/19 21:43:18 OpenAI APIからの応答を受信しました
分析結果: 提供された計算機モジュールに関して、コードとドキュメントの間で顕著な矛盾は見られません。コードとドキュメントは、一貫して.add関数の使用方法と動作について説明しています。注意事項に関して言及されている内容も、コードでの実装と一致しています。
矛盾は見つかりませんでした
```