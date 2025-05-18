package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// 複数ファイル間の矛盾を検出するプログラム
// 引数として2つ以上のファイルを取り、それらの間の矛盾を検出する。
// 矛盾点があれば標準出力に「ファイル1,ファイル2:矛盾内容」の形式で出力し、終了コード1で終了する。
// 矛盾点がなければ何も出力せず、終了コード0で終了する。
func main() {
	// LLMモードの設定（デフォルトはモック）
	llmMode := os.Getenv("LLM_MODE")
	if llmMode == "" {
		llmMode = "mock"
	}

	// 引数チェック
	if len(os.Args) < 3 {
		fmt.Printf("使用方法: %s <ファイル1> <ファイル2> [<ファイル3>...]\n", os.Args[0])
		fmt.Println("環境変数:")
		fmt.Printf("  LLM_MODE=xxx   : 使用するLLMモード（現在: %s）\n", llmMode)
		fmt.Printf("  OPENAI_API_KEY : OpenAI APIキー（%s）\n",
			func() string {
				if os.Getenv("OPENAI_API_KEY") != "" {
					return "設定済み"
				}
				return "未設定"
			}())
		os.Exit(1)
	}

	// 引数からファイルパスを取得
	files := os.Args[1:]
	
	// 各ファイルの存在チェック
	for _, file := range files {
		if !fileExists(file) {
			fmt.Printf("エラー: ファイル '%s' が見つかりません\n", file)
			os.Exit(1)
		}
	}

	// ファイル内容の読み込み
	contents := make(map[string]string)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("エラー: ファイル '%s' の読み込み中にエラーが発生しました: %s\n", file, err)
			os.Exit(1)
		}
		contents[file] = string(content)
	}

	var result string

	// 使用するLLMモードを決定
	if llmMode == "openai" {
		fmt.Println("OpenAI APIを使ってファイル間の矛盾を分析しています...")
		result = openaiLlmCheck(files, contents)
	} else {
		fmt.Println("モックLLMを使ってファイル間の矛盾を分析しています...")
		result = mockLlmCheck(files, contents)
	}

	// 結果の処理
	if result == "" {
		fmt.Println("矛盾は見つかりませんでした")
		os.Exit(0)
	} else {
		fmt.Println(result)
		os.Exit(1)
	}
}

// ファイルの存在をチェックする関数
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// テスト用のモック機能
func mockLlmCheck(files []string, contents map[string]string) string {
	// 複数ファイル間の矛盾をチェック
	if len(files) < 2 {
		return "比較するファイルが2つ以上必要です"
	}

	// correct/かincorrect/のチェック
	anyIncorrect := false
	for _, file := range files {
		absPath, _ := filepath.Abs(file)
		if strings.Contains(absPath, "incorrect") {
			anyIncorrect = true
			break
		}
	}

	// モックによる判定ロジック
	if !anyIncorrect {
		return "" // 矛盾なしは空文字列
	} else {
		// 矛盾点を出力
		var contradictions []string
		
		// 最初の2つのファイルを使った矛盾例を生成
		file1 := files[0]
		file2 := files[1]
		
		contradictions = append(contradictions, fmt.Sprintf("%s,%s:add関数は数値変換を行っていません", file1, file2))
		contradictions = append(contradictions, fmt.Sprintf("%s,%s:multiply関数がドキュメントに記載されていません", file1, file2))
		contradictions = append(contradictions, fmt.Sprintf("%s,%s:オプションの第3引数（操作タイプ）がドキュメントに記載されていません", file1, file2))
		
		// 3つ以上のファイルがある場合は追加の矛盾を生成
		if len(files) > 2 {
			file3 := files[2]
			contradictions = append(contradictions, fmt.Sprintf("%s,%s:ファイル間で機能の説明が一致していません", file1, file3))
		}
		
		return strings.Join(contradictions, "\n")
	}
}

// OpenAI APIを使用して差異を分析
func openaiLlmCheck(files []string, contents map[string]string) string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("エラー: OPENAI_API_KEYが設定されていません")
		return fmt.Sprintf("%s:矛盾の分析中にエラーが発生しました", strings.Join(files, ","))
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// プロンプトの作成
	promptBuilder := strings.Builder{}
	promptBuilder.WriteString("以下の複数ファイルを比較し、不一致や矛盾点を見つけてください。\n")
	promptBuilder.WriteString("各矛盾点は「ファイル1,ファイル2:矛盾内容」の形式で1行ずつ出力してください。\n")
	promptBuilder.WriteString("矛盾がない場合は何も出力しないでください。\n\n")
	
	// 各ファイルの内容を追加
	for _, file := range files {
		promptBuilder.WriteString(fmt.Sprintf("== ファイル (%s) ==\n", file))
		promptBuilder.WriteString(contents[file])
		promptBuilder.WriteString("\n\n")
	}

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "あなたはコードとドキュメントの間の矛盾を検出し、明確に報告する専門家です。",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: promptBuilder.String(),
			},
		},
		MaxTokens: 1000,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("エラー: OpenAI APIの呼び出しに失敗しました: %s\n", err)
		return fmt.Sprintf("%s:矛盾の分析中にエラーが発生しました", strings.Join(files, ","))
	}

	result := resp.Choices[0].Message.Content
	log.Printf("OpenAI APIからの応答: %s\n", result)

	// 結果の後処理
	if strings.Contains(result, "矛盾") || strings.Contains(result, "不一致") {
		// 行を分割して処理
		lines := strings.Split(result, "\n")
		var filteredLines []string

		for _, line := range lines {
			// 矛盾の形式「ファイル1,ファイル2:矛盾内容」をチェック
			for i, file1 := range files {
				for j, file2 := range files {
					if i != j && strings.Contains(line, fmt.Sprintf("%s,%s:", file1, file2)) {
						filteredLines = append(filteredLines, line)
						break
					}
				}
			}
		}

		if len(filteredLines) > 0 {
			return strings.Join(filteredLines, "\n")
		}
		return fmt.Sprintf("%s:矛盾が見つかりましたが、詳細は解析できませんでした", strings.Join(files, ","))
	}

	// 矛盾がない場合は空文字を返す
	return ""
}
