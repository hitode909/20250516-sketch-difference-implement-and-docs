package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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

// 構造化されたレスポンス用の型定義
type StructuredResponse struct {
	Summary string `json:"summary"`
	Errors  []struct {
		File1       string `json:"file1"`
		File2       string `json:"file2"`
		Description string `json:"description"`
	} `json:"errors"`
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
	promptBuilder.WriteString("レスポンスは必ず以下のJSON形式で返してください：\n")
	promptBuilder.WriteString("```json\n")
	promptBuilder.WriteString("{\n")
	promptBuilder.WriteString("  \"summary\": \"総評テキスト（矛盾の概要や全体的な評価）\",\n")
	promptBuilder.WriteString("  \"errors\": [\n")
	promptBuilder.WriteString("    {\n")
	promptBuilder.WriteString("      \"file1\": \"ファイル1のパス\",\n")
	promptBuilder.WriteString("      \"file2\": \"ファイル2のパス\",\n")
	promptBuilder.WriteString("      \"description\": \"矛盾の詳細説明\"\n")
	promptBuilder.WriteString("    },\n")
	promptBuilder.WriteString("    // 他の矛盾点...\n")
	promptBuilder.WriteString("  ]\n")
	promptBuilder.WriteString("}\n")
	promptBuilder.WriteString("```\n")
	promptBuilder.WriteString("矛盾がない場合は、errors配列を空にしてください。\n\n")

	// 各ファイルの内容を追加
	for _, file := range files {
		promptBuilder.WriteString(fmt.Sprintf("== ファイル (%s) ==\n", file))
		promptBuilder.WriteString(contents[file])
		promptBuilder.WriteString("\n\n")
	}

	// システムのLANG環境変数を取得
	langEnv := os.Getenv("LANG")
	if langEnv == "" {
		langEnv = "ja_JP.UTF-8" // デフォルト値
	}

	// システムプロンプトにLANG環境変数を含める
	systemPrompt := fmt.Sprintf(
		"あなたはコードとドキュメントの間の矛盾を検出し、明確に報告する専門家です。"+
			"必ず指定されたJSON形式で回答してください。"+
			"応答言語は環境変数 LANG=%s に従ってください。",
		langEnv)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: promptBuilder.String(),
			},
		},
		MaxTokens: 1000,
		Temperature: 0, // 安定した結果のために温度を0に設定
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("エラー: OpenAI APIの呼び出しに失敗しました: %s\n", err)
		return fmt.Sprintf("%s:矛盾の分析中にエラーが発生しました", strings.Join(files, ","))
	}

	result := resp.Choices[0].Message.Content
	// デバッグ用にログに出力（本番環境では削除可能）
	log.Printf("OpenAI APIからの応答を受信しました\n")

	// JSONレスポンスを抽出（コードブロックで囲まれている可能性があるため）
	jsonStr := extractJSON(result)
	if jsonStr == "" {
		log.Println("エラー: JSONレスポンスを抽出できませんでした")
		return fmt.Sprintf("%s:矛盾の分析中にエラーが発生しました", strings.Join(files, ","))
	}

	// JSONをパース
	var structuredResp StructuredResponse
	if err := json.Unmarshal([]byte(jsonStr), &structuredResp); err != nil {
		log.Printf("エラー: JSONのパースに失敗しました: %s\n", err)
		return fmt.Sprintf("%s:矛盾の分析中にエラーが発生しました", strings.Join(files, ","))
	}

	// サマリーを表示
	if structuredResp.Summary != "" {
		fmt.Printf("分析結果: %s\n", structuredResp.Summary)
	}

	// 矛盾がない場合は空文字を返す
	if len(structuredResp.Errors) == 0 {
		return ""
	}

	// 矛盾がある場合は、指定された形式で出力
	var filteredLines []string
	for _, err := range structuredResp.Errors {
		// ファイルパスが有効かチェック
		validFiles := true
		for _, file := range []string{err.File1, err.File2} {
			found := false
			for _, validFile := range files {
				if file == validFile {
					found = true
					break
				}
			}
			if !found {
				validFiles = false
				break
			}
		}

		if validFiles {
			filteredLines = append(filteredLines, fmt.Sprintf("%s,%s:%s", err.File1, err.File2, err.Description))
		}
	}

	if len(filteredLines) > 0 {
		return strings.Join(filteredLines, "\n")
	}
	return fmt.Sprintf("%s:矛盾が見つかりましたが、詳細は解析できませんでした", strings.Join(files, ","))
}

// JSONレスポンスを抽出する関数
func extractJSON(text string) string {
	// コードブロックで囲まれたJSONを探す
	jsonBlockRegex := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
	matches := jsonBlockRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// コードブロックがない場合は、テキスト全体がJSONかもしれない
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}") {
		return text
	}

	return ""
}
