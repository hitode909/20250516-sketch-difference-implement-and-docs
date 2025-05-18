#!/usr/bin/env node
// filepath: /Users/hitode909/co/github.com/hitode909/20250516-sketch-difference-implement-and-docs/run_meta_test.js

/**
 * このプログラムはcheck_differencesの動作をテストします。
 * - correct/ : 矛盾がない（チェックプログラムはexit 0を返す）
 * - incorrect/: 矛盾がある（チェックプログラムはexit 1を返す）
 *
 * 上記の条件を満たしている場合のみ、このメタテストプログラムはexit 0を返します。
 */

const { execSync, spawnSync } = require('child_process');
const path = require('path');
const fs = require('fs');

// 作業ディレクトリのパス
const workDir = __dirname;

// check_differencesのパス
const checkDiffProgram = path.join(workDir, 'check_differences');

// テスト結果を格納する変数
let allTestsPassed = true;
let testResults = [];

// 実行ファイルの存在確認
if (!fs.existsSync(checkDiffProgram)) {
  console.error(`エラー: check_differences実行ファイルが見つかりません: ${checkDiffProgram}`);
  console.log('先に「go build -o check_differences」を実行してください');
  process.exit(1);
}

/**
 * check_differences実行ファイルを実行してテスト
 * @param {string[]} files テスト対象のファイルパスの配列
 * @param {boolean} expectedContradiction 矛盾があることを期待するか
 * @param {string} llmMode 使用するLLMモード (省略時はmock)
 * @returns {object} テスト結果
 */
function runContradictionTest(files, expectedContradiction, llmMode = "mock") {
  console.log(`テスト実行: ${files.join(' と ')} の矛盾チェック (モード: ${llmMode})`);
  console.log(`期待される結果: 矛盾${expectedContradiction ? 'あり' : 'なし'}`);

  try {
    // 環境変数の設定
    const env = { ...process.env, LLM_MODE: llmMode };

    // check_differencesを実行
    const result = spawnSync(checkDiffProgram, files, {
      encoding: 'utf8',
      stdio: ['pipe', 'pipe', 'pipe'],
      env: env
    });

    const exitCode = result.status;
    const stdout = result.stdout.trim();
    const stderr = result.stderr.trim();

    // 期待される終了コード
    const expectedExitCode = expectedContradiction ? 1 : 0;

    // 結果を評価
    const passed = exitCode === expectedExitCode;

    if (passed) {
      console.log('✅ テスト成功');
    } else {
      console.log('❌ テスト失敗');
      console.log(`  期待された終了コード: ${expectedExitCode}, 実際: ${exitCode}`);
      allTestsPassed = false;
    }

    if (stdout) {
      console.log('出力:');
      console.log(stdout);
    }

    if (stderr) {
      console.log('エラー:');
      console.log(stderr);
    }

    console.log('-'.repeat(40));

    return {
      files: files.join(' と '),
      expectedContradiction,
      actualExitCode: exitCode,
      passed,
      output: stdout,
      error: stderr
    };
  } catch (err) {
    console.error(`テスト実行中にエラーが発生しました: ${err.message}`);
    allTestsPassed = false;
    return {
      files: files.join(' と '),
      expectedContradiction,
      passed: false,
      error: err.message
    };
  }
}

// テストケースを実行
console.log('=== 矛盾チェックメタテスト開始 ===');

// ファイルパスの設定
const correctImplFile = path.join(workDir, 'correct', 'calculator.js');
const correctDocFile = path.join(workDir, 'correct', 'calculator.md');
const incorrectImplFile = path.join(workDir, 'incorrect', 'calculator.js');
const incorrectDocFile = path.join(workDir, 'incorrect', 'calculator.md');

// テスト対象ファイルをグループ化
const correctFiles = [correctImplFile, correctDocFile];
const incorrectFiles = [incorrectImplFile, incorrectDocFile];

// mockモードでのテスト
console.log('=== mockモードでのテスト ===');
// correct/のテスト - 矛盾なしを期待
testResults.push(runContradictionTest(correctFiles, false, "mock"));
// incorrect/のテスト - 矛盾ありを期待
testResults.push(runContradictionTest(incorrectFiles, true, "mock"));

// OpenAI APIキーが設定されている場合はopenaiモードでもテスト
if (process.env.OPENAI_API_KEY) {
  console.log('=== openaiモードでのテスト ===');
  // correct/のテスト - 矛盾なしを期待
  testResults.push(runContradictionTest(correctFiles, false, "openai"));
  // incorrect/のテスト - 矛盾ありを期待
  testResults.push(runContradictionTest(incorrectFiles, true, "openai"));
} else {
  console.log('OPENAI_API_KEYが設定されていないため、openaiモードのテストはスキップします');
}

// 結果のサマリーを表示
console.log('=== テスト結果サマリー ===');
testResults.forEach((result, index) => {
  console.log(`${index + 1}. ${result.files}: ${result.passed ? '成功' : '失敗'}`);
});

console.log(`\n全体結果: ${allTestsPassed ? '成功 ✅' : '失敗 ❌'}`);

// 最終的なテスト結果に基づいて終了コードを設定
process.exit(allTestsPassed ? 0 : 1);
