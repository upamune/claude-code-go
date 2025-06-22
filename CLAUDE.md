# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

claude-code-goは、Claude Code CLIのGo SDKです。GoアプリケーションからClaude AIの機能を簡単に統合できるように設計されています。

## 開発コマンド

### ツール管理
```bash
# aquaで管理されているツールのインストール
aqua i

# コードフォーマット
goimports -w .

# Linting
revive ./...
```

### ビルドとテスト
```bash
# 依存関係の取得
go mod tidy

# ビルド
go build ./...

# テスト実行（テストが実装されたら）
go test ./...

# 使用例の実行
cd examples/basic
go run main.go
```

## アーキテクチャ概要

### コア設計原則
1. **シンプルなラッパー**: Claude Code CLIを`exec.Command`で実行し、出力をパース
2. **型安全**: 各メッセージタイプに対応する構造体とインターフェースを提供
3. **最小限の依存**: 標準ライブラリのみを使用

### 主要コンポーネント

#### claude.go - メインAPI
- `Query()`: 同期的にクエリを実行し、最終結果を返す
- `QueryStream()`: ストリーミング出力でクエリを実行、各メッセージでコールバック
- `Exec()`: 生のCLIコマンドを実行

#### types.go - 型定義
- `Options`: SDK設定（モデル、ツール制限、権限など）
- `Message`インターフェース: 全メッセージタイプの共通インターフェース
- MCPサーバー設定: Stdio、SSE、HTTPタイプをサポート

#### message.go - メッセージパーシング
- CLIのJSON出力を適切な型にパース
- 各メッセージタイプ（User、Assistant、Result、System、PermissionRequest）の処理

#### errors.go - エラー型
- `AbortError`: ユーザー中断
- `ProcessError`: CLIプロセスエラー
- `ParseError`: JSONパースエラー
- `ConfigError`: 設定エラー

### メッセージフロー
1. ユーザーが`Query()`または`QueryStream()`を呼び出し
2. SDKがClaude CLIプロセスを起動
3. CLIからのJSON出力を`parseMessage()`でパース
4. 適切なメッセージ型として返却またはコールバック実行

## 開発時の注意点

### CLIとの統合
- Claude Code CLIがインストール・設定済みであることが前提
- `exec.Command`の出力はバッファリングされるため、大量出力時は注意
- プロセスの適切なクリーンアップを保証（defer cmd.Wait()）

### エラーハンドリング
- CLIの終了コードをチェックし、適切なエラー型を返す
- パースエラーは元のメッセージを含めて返す
- ユーザー中断（Ctrl+C）は`AbortError`として処理

### 今後の実装予定
- テストの追加（単体テスト、統合テスト）
- より高度なストリーミング制御
- セッション管理機能の強化
