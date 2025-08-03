# GitHub Actions Workflows

このディレクトリには、Merchant Tailsプロジェクトの継続的インテグレーション（CI）ワークフローが含まれています。

## Workflows

### 1. Code Format Check (`format-check.yml`)

**目的**: C#コードのフォーマットをチェック

**トリガー**:
- `main`ブランチへのプッシュ
- プルリクエスト

**動作**:
- CSharpierを使用してコードフォーマットをチェック
- フォーマットエラーがある場合はビルドを失敗させる

### 2. Unity Tests (`unity-tests.yml`)

**目的**: Unity Test Frameworkで作成されたテストを実行

**トリガー**:
- `main`, `develop`ブランチへのプッシュ
- `main`ブランチへのプルリクエスト

**動作**:
- PlayModeテストを実行
- テスト結果をアーティファクトとして保存
- コードカバレッジレポートを生成
- PRにテスト結果とカバレッジをコメント

**必要なSecrets**:
- `UNITY_LICENSE`: Unity ProまたはPlusライセンス
- `UNITY_EMAIL`: Unityアカウントのメールアドレス
- `UNITY_PASSWORD`: Unityアカウントのパスワード

## セットアップ手順

### Unity ライセンスの設定

1. ローカルでUnityプロジェクトを開く
2. Unity Hubでライセンスをアクティベート
3. 以下のコマンドでライセンスファイルを取得:
   ```bash
   # macOS/Linux
   cat ~/Library/Unity/Unity_lic.ulf
   
   # Windows
   type %PROGRAMDATA%\Unity\Unity_lic.ulf
   ```
4. GitHubリポジトリの Settings > Secrets and variables > Actions で以下を設定:
   - `UNITY_LICENSE`: ライセンスファイルの内容
   - `UNITY_EMAIL`: Unityアカウントのメール
   - `UNITY_PASSWORD`: Unityアカウントのパスワード

### ローカルでのテスト実行

CIで実行される前にローカルでテストを確認:

```bash
# Unityコマンドラインでテスト実行
Unity -runTests -projectPath . -testResults results.xml -testPlatform PlayMode

# CSharpierフォーマットチェック
dotnet tool restore
dotnet csharpier . --check
```

## トラブルシューティング

### Unity ライセンスエラー

エラー: `License activation failed`

解決方法:
1. Unity Personal版を使用している場合は、game-ci/unity-request-activation-file アクションを使用
2. ライセンスが期限切れの場合は、新しいライセンスを取得して更新

### テスト失敗

エラー: `Tests failed`

解決方法:
1. ローカルでテストを実行して問題を確認
2. テストログをアーティファクトからダウンロードして詳細を確認
3. Unity Editorでテストをデバッグ

### フォーマットエラー

エラー: `Code formatting check failed`

解決方法:
```bash
# フォーマットを自動修正
dotnet csharpier .
```

## ベストプラクティス

1. **プルリクエスト前にローカルでテスト実行**
2. **コードをコミットする前にフォーマット実行**
3. **テストカバレッジ75%以上を維持**
4. **新機能には必ずテストを追加**
5. **CIが失敗したらすぐに修正**

## 参考リンク

- [Unity Test Framework Documentation](https://docs.unity3d.com/Packages/com.unity.test-framework@latest)
- [Game-CI Documentation](https://game-ci.com/docs)
- [CSharpier Documentation](https://csharpier.com/)
- [GitHub Actions Documentation](https://docs.github.com/actions)