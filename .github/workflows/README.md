# GitHub Actions Workflows

このディレクトリには、Merchant Tailsプロジェクトの継続的インテグレーション/継続的デプロイメント（CI/CD）ワークフローが含まれています。

## ⚠️ 現在の状況

**Unity 6 互換性の問題**: Unity 6 (6000.x.x) のバージョン形式が game-ci アクションでまだサポートされていないため、Unity テストとビルドは一時的に無効化されています。

## ワークフロー

### Unity CI/CD Pipeline (`unity-ci.yml`)

**目的**: Unityプロジェクトの自動テストとビルド

**トリガー**:
- `main`, `develop`ブランチへのプッシュ
- `main`ブランチへのプルリクエスト
- 手動実行（workflow_dispatch）

**ジョブ**:

#### 1. Test (Unity Tests)
- **EditMode Tests**: エディター環境でのユニットテスト
- **PlayMode Tests**: ゲーム実行環境での統合テスト
- テスト結果をJUnit形式でレポート
- テスト結果をアーティファクトとして保存

#### 2. Build
- 対象プラットフォーム: Windows, Linux, macOS
- ビルド成果物をアーティファクトとして保存

#### 3. Code Quality
- CSharpierによるコードフォーマットチェック
- コード品質の維持

#### 4. Performance Test
- パフォーマンステストの実行
- パフォーマンス結果をアーティファクトとして保存

**必要なSecrets**:
- `UNITY_LICENSE`: Unityライセンスファイル（.ulf）の内容
- `UNITY_EMAIL`: Unity IDのメールアドレス（オプション）
- `UNITY_PASSWORD`: Unity IDのパスワード（オプション）

## セットアップ手順

### Unity ライセンスの設定

#### Unity Personal版を使用する場合

**注意**: 2023年12月以降、Unity PersonalライセンスはManual Activationでの.ulfファイル取得ができません。

**推奨方法**:
1. Unity Hubでローカルにライセンスをアクティベート
2. 既存のライセンスファイルの場所を確認:
   ```bash
   # macOS
   cat ~/Library/Unity/licenses/UnityEntitlementLicense.xml
   
   # Windows
   type %PROGRAMDATA%\Unity\Unity_lic.ulf
   
   # Linux
   cat ~/.local/share/unity3d/Unity/Unity_lic.ulf
   ```
3. GitHubリポジトリの Settings > Secrets and variables > Actions で設定:
   - `UNITY_LICENSE`: ライセンスファイルの内容全体をコピー&ペースト

**代替方法**: Self-hosted Runnerを使用（ローカルマシンでCIを実行）

#### Unity Pro/Plus版を使用する場合

標準的な方法でライセンスファイルを取得可能です。

### ローカルでのテスト実行

CIで実行される前にローカルでテストを確認:

```bash
# Unity 6.1 LTS (6000.1.14f1) を使用

# EditMode テスト実行
/Applications/Unity/Hub/Editor/6000.1.14f1/Unity.app/Contents/MacOS/Unity \
  -runTests -projectPath . -testPlatform EditMode \
  -testResults test-results/EditMode-results.xml -logFile -

# PlayMode テスト実行
/Applications/Unity/Hub/Editor/6000.1.14f1/Unity.app/Contents/MacOS/Unity \
  -runTests -projectPath . -testPlatform PlayMode \
  -testResults test-results/PlayMode-results.xml -logFile -

# CSharpierフォーマットチェック
dotnet tool restore
dotnet csharpier Assets/Scripts --check
```

### ローカルビルドスクリプト

プロジェクトに含まれる`scripts/local-build.sh`を使用:
```bash
./scripts/local-build.sh
```

## トラブルシューティング

### Unity ライセンスエラー

エラー: `License activation failed`

解決方法:
1. Unity Personal版の場合、ローカルでアクティベートしたライセンスを使用
2. GitHub Secretsの`UNITY_LICENSE`が正しく設定されているか確認
3. ライセンスが期限切れの場合は、Unity Hubで再アクティベート

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
6. **Unity 6.1 LTS (6000.1.14f1) を使用**
7. **テストは EditMode と PlayMode の両方を作成**

## 参考リンク

- [Unity Test Framework Documentation](https://docs.unity3d.com/Packages/com.unity.test-framework@latest)
- [Game-CI Documentation](https://game-ci.com/docs)
- [CSharpier Documentation](https://csharpier.com/)
- [GitHub Actions Documentation](https://docs.github.com/actions)