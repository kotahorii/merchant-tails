# Testing Guide - Merchant Tales

## Overview

Merchant Tales プロジェクトでは、安定性重視のアプローチに基づいた包括的なテストシステムを構築しています。このガイドでは、テストの実行方法と CI/CD 設定について説明します。

## Test Architecture

### Core Test Components

1. **SystemTestController** - 基本機能と UI 統合テスト
2. **IntegrationTest** - システム間連携テスト（6 つのシナリオ）
3. **StabilityTest** - 長時間実行・ストレステスト
4. **ErrorRecoveryTest** - 例外処理・システム回復テスト
5. **AutomatedTestRunner** - CI 環境での自動テスト実行
6. **CITestValidator** - テスト前の環境検証

### Test Categories

-   **Integration**: システム間連携テスト
-   **Performance**: パフォーマンス・安定性テスト
-   **ErrorHandling**: エラー処理・回復テスト
-   **Functional**: 基本機能テスト
-   **Configuration**: 設定・構成テスト
-   **Smoke**: 最小限の動作確認テスト

## Local Testing

### Unity Editor での実行

1. **Test Runner Window**を開く (Window > General > Test Runner)
2. **EditMode**タブで設定テストを実行
3. **PlayMode**タブで統合テストを実行

### 手動テスト実行

1. MainGame シーンを開く
2. Play ボタンを押す
3. OnGUI のテストボタンで個別テスト実行
4. Console Window で結果を確認

### テストコマンド

```bash
# Unity Test Runner (エディターモード)
Unity -projectPath . -runTests -testPlatform editmode -testResults test-results-edit.xml

# Unity Test Runner (プレイモード)
Unity -projectPath . -runTests -testPlatform playmode -testResults test-results-play.xml

# 特定カテゴリのテスト
Unity -projectPath . -runTests -testPlatform playmode -testCategory "Integration"
```

## CI/CD Pipeline

### GitHub Actions Workflow

`.github/workflows/unity-ci.yml`に CI/CD パイプラインを設定済みです。

#### Workflow Stages

1. **Test Stage**

    - EditMode テスト実行
    - PlayMode テスト実行
    - 複数 Unity バージョンでのテスト

2. **Build Stage**

    - Windows、Linux、macOS 向けビルド
    - 並列ビルド実行

3. **Code Quality Stage**

    - コードフォーマットチェック
    - セキュリティ解析

4. **Performance Test Stage**

    - パフォーマンステスト実行
    - メモリ使用量チェック

5. **Deploy Stage**
    - Steam 配信準備（本番環境のみ）

### Required Secrets

GitHub Repository に以下の Secrets を設定してください：

```
UNITY_LICENSE: Unity Personalライセンス文字列
UNITY_EMAIL: Unityアカウントのメールアドレス
UNITY_PASSWORD: Unityアカウントのパスワード
```

### Unity License 取得

```bash
# Unity License取得スクリプト
Unity -quit -batchmode -serial YOUR_SERIAL_KEY -username YOUR_EMAIL -password YOUR_PASSWORD
```

## Test Execution Results

### Expected Test Results

-   **EditMode Tests**: 8 tests passing

    -   Project structure validation
    -   Script compilation validation
    -   Player settings validation
    -   Build settings validation
    -   ScriptableObject validation
    -   Event system validation
    -   Enum definitions validation
    -   Testing infrastructure validation

-   **PlayMode Tests**: 8 tests passing
    -   System health check
    -   Integration test suite (6 scenarios)
    -   Stability test suite
    -   Error recovery test suite
    -   System functionality verification
    -   Memory usage verification
    -   Configuration verification
    -   Smoke test

### Success Criteria

-   **Integration Tests**: 100% pass rate (0 failures)
-   **Stability Tests**: ≥95% success rate, 0 critical errors
-   **Error Recovery Tests**: ≥80% success rate
-   **Memory Usage**: <200MB in CI environment
-   **Performance**: All tests complete within timeout

## Troubleshooting

### Common Issues

1. **Tests Time Out**

    - CI 環境では短縮設定を使用
    - ネットワーク接続を確認

2. **Memory Issues**

    - GC を強制実行
    - オブジェクトの適切な破棄

3. **Platform Issues**
    - Unity version compatibility 確認
    - Assembly definitions の設定確認

### Debug Commands

```bash
# テスト詳細ログ
Unity -projectPath . -runTests -testPlatform playmode -logFile test.log

# メモリプロファイリング
Unity -projectPath . -runTests -testPlatform playmode -enableCodeCoverage

# エラー詳細出力
Unity -projectPath . -runTests -testPlatform playmode -stackTraceLogType Full
```

## Performance Benchmarks

### Target Performance

-   **Initialization**: <2 seconds
-   **Integration Tests**: <30 seconds
-   **Stability Tests**: <60 seconds (CI), <120 seconds (full)
-   **Memory Usage**: <100MB baseline, <200MB peak
-   **Frame Rate**: >30 FPS during testing

### Monitoring

-   Memory usage tracking
-   Frame rate monitoring
-   Test execution time measurement
-   Error frequency analysis

## Continuous Integration Best Practices

1. **Fast Feedback Loop**: Critical tests run first
2. **Parallel Execution**: Independent tests run concurrently
3. **Environment Isolation**: Each test run in clean environment
4. **Comprehensive Coverage**: All major systems tested
5. **Clear Reporting**: Detailed test results and artifacts

## Future Improvements

-   [ ] Add visual regression testing
-   [ ] Implement automated performance benchmarking
-   [ ] Add cross-platform compatibility tests
-   [ ] Integrate with Steam backend testing
-   [ ] Add localization testing
-   [ ] Implement stress testing for longer durations

## Contact

For testing-related questions or issues, please check the test logs and error reports in the CI artifacts.
