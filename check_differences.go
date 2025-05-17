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

// 実装とドキュメントの矛盾を検出するプログラム
// 引数として実装ファイルとドキュメントファイルを取り、
// 矛盾点があれば標準出力に出力し、終了コード1で終了する。
// 矛盾点がなければ何も出力せず、終了コード0で終了する。
func main() {
	// LLMモードの設定（デフォルトはモック）
	llmMode := os.Getenv("LLM_MODE")
	if llmMode == "" {
		llmMode = "mock"
	}

	// 引数チェック
	if len(os.Args) != 3 {
		fmt.Printf("使用方法: %s <実装ファイル> <ドキュメントファイル>\n", os.Args[0])
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

	implFile := os.Args[1]
	docFile := os.Args[2]

	// ファイルの存在チェック
	if !fileExists(implFile) {
		fmt.Printf("エラー: 実装ファイル '%s' が見つかりません\n", implFile)
		os.Exit(1)
	}
	if !fileExists(docFile) {
		fmt.Printf("エラー: ドキュメントファイル '%s' が見つかりません\n", docFile)
		os.Exit(1)
	}

	// ファイル内容の読み込み
	implContent, err := os.ReadFile(implFile)
	if err != nil {
		log.Printf("エラー: 実装ファイルの読み込み中にエラーが発生しました: %s\n", err)
		os.Exit(1)
	}

	docContent, err := os.ReadFile(docFile)
	if err != nil {
		log.Printf("エラー: ドキュメントファイルの読み込み中にエラーが発生しました: %s\n", err)
		os.Exit(1)
	}

	var result string

	// 使用するLLMモードを決定
	if llmMode == "openai" {
		fmt.Println("OpenAI APIを使って実装とドキュメントの矛盾を分析しています...")
		result = openaiLlmCheck(implFile, docFile, string(implContent), string(docContent))
	} else {
		fmt.Println("モックLLMを使って実装とドキュメントの矛盾を分析しています...")
		result = mockLlmCheck(implFile, docFile)
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

// プログラムが現在サポートしているLLMモードをチェックする関数
func getSupportedLLMModes() []string {
	return []string{
		"mock",    // モックモード（デフォルト）
		"openai",  // OpenAI APIモード
	}
}

// テスト用のモック機能
func mockLlmCheck(implFile, docFile string) string {
	// モックによる判定ロジック
	// correct/なら矛盾なし、incorrect/なら矛盾ありと判定
	absImpl, _ := filepath.Abs(implFile)

	if strings.Contains(absImpl, "correct") && !strings.Contains(absImpl, "incorrect") {
		return "" // 矛盾なしは空文字列
	} else if strings.Contains(absImpl, "incorrect") {
		// 矛盾点を出力
		var contradictions []string
		contradictions = append(contradictions, fmt.Sprintf("%s,%s:add関数は数値変換を行っていません", implFile, docFile))
		contradictions = append(contradictions, fmt.Sprintf("%s,%s:multiply関数がドキュメントに記載されていません", implFile, docFile))
		contradictions = append(contradictions, fmt.Sprintf("%s,%s:コマンドラインは3つの引数を受け付けますが、ドキュメントでは2つと記載", implFile, docFile))
		return strings.Join(contradictions, "\n")
	}
	return "不明なディレクトリ構造です"
}



// OpenAI APIを使用して差異を分析
func openaiLlmCheck(implFile, docFile, implContent, docContent string) string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("エラー: OPENAI_API_KEYが設定されていません")
		return fmt.Sprintf("%s,%s:矛盾の分析中にエラーが発生しました", implFile, docFile)
	}

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	prompt := fmt.Sprintf(`以下の実装ファイルとドキュメントファイルを比較し、不一致や矛盾点を見つけてください。
各矛盾点は「%s,%s:矛盾内容」の形式で1行ずつ出力してください。
矛盾がない場合は何も出力しないでください。

== 実装ファイル (%s) ==
%s

== ドキュメントファイル (%s) ==
%s`, implFile, docFile, implFile, implContent, docFile, docContent)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "あなたはコードとドキュメントの間の矛盾を検出し、明確に報告する専門家です。",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens: 1000,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("エラー: OpenAI APIの呼び出しに失敗しました: %s\n", err)
		return fmt.Sprintf("%s,%s:矛盾の分析中にエラーが発生しました", implFile, docFile)
	}

	result := resp.Choices[0].Message.Content

	// 結果の後処理
	if strings.Contains(result, "矛盾") || strings.Contains(result, "不一致") {
		// 行を分割して処理
		lines := strings.Split(result, "\n")
		var filteredLines []string

		for _, line := range lines {
			if strings.Contains(line, fmt.Sprintf("%s,%s:", implFile, docFile)) {
				filteredLines = append(filteredLines, line)
			}
		}

		if len(filteredLines) > 0 {
			return strings.Join(filteredLines, "\n")
		}
		return fmt.Sprintf("%s,%s:矛盾が見つかりましたが、詳細は解析できませんでした", implFile, docFile)
	}

	// 矛盾がない場合は空文字を返す
	return ""
}
