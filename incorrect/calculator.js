/**
 * シンプルな計算機能を提供するモジュール
 */

/**
 * 2つの数値を足し算する関数
 * @param {number} a - 1つ目の数値
 * @param {number} b - 2つ目の数値
 * @returns {number} 足し算の結果
 */
function add(a, b) {
  // 文字列連結として処理
  return a + b;
}

/**
 * 2つの数値を掛け算する関数
 * @param {number} a - 1つ目の数値
 * @param {number} b - 2つ目の数値
 * @returns {number} 掛け算の結果
 */
function multiply(a, b) {
  return Number(a) * Number(b);
}

/**
 * 使用例
 */
if (require.main === module) {
  const args = process.argv.slice(2);
  if (args.length < 2 || args.length > 3) {
    console.error('使用方法: node calculator.js <数値1> <数値2> [operation]');
    process.exit(1);
  }

  let result;
  const operation = args[2] || 'add';

  if (operation === 'multiply') {
    result = multiply(args[0], args[1]);
    console.log(`${args[0]} * ${args[1]} = ${result}`);
  } else {
    result = add(args[0], args[1]);
    console.log(`${args[0]} + ${args[1]} = ${result}`);
  }
}

// モジュールとしてエクスポート
module.exports = {
  add,
  multiply
};
