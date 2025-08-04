#!/bin/bash

# Unity ローカルビルドスクリプト
set -e

UNITY_PATH="/Applications/Unity/Hub/Editor/6000.1.14f1/Unity.app/Contents/MacOS/Unity"
PROJECT_PATH="$(pwd)"
COMMAND="${1:-all}"

echo "Unity Local Build Script"
echo "========================"
echo "Unity Path: $UNITY_PATH"
echo "Project Path: $PROJECT_PATH"
echo "Command: $COMMAND"
echo ""

# コマンドによる処理分岐
case "$COMMAND" in
  "test")
    # テストの実行
    echo "Running EditMode Tests..."
    "$UNITY_PATH" \
        -runTests \
        -batchmode \
        -projectPath "$PROJECT_PATH" \
        -testPlatform EditMode \
        -testResults "$PROJECT_PATH/test-results/EditMode-results.xml" \
        -logFile "$PROJECT_PATH/test-results/EditMode.log" \
        || echo "EditMode tests completed"

    echo "Running PlayMode Tests..."
    "$UNITY_PATH" \
        -runTests \
        -batchmode \
        -projectPath "$PROJECT_PATH" \
        -testPlatform PlayMode \
        -testResults "$PROJECT_PATH/test-results/PlayMode-results.xml" \
        -logFile "$PROJECT_PATH/test-results/PlayMode.log" \
        || echo "PlayMode tests completed"
    ;;

  "build")
    # ビルドの実行
    echo "Building for macOS..."
    "$UNITY_PATH" \
        -batchmode \
        -quit \
        -projectPath "$PROJECT_PATH" \
        -buildTarget StandaloneOSX \
        -buildPath "$PROJECT_PATH/build/macOS/MerchantTails.app" \
        -logFile "$PROJECT_PATH/build/build.log"

    echo ""
    echo "Build completed!"
    echo "Output: $PROJECT_PATH/build/macOS/MerchantTails.app"
    ;;

  "all"|*)
    # テストとビルドの両方を実行
    echo "Running all tasks..."
    "$0" test
    "$0" build
    ;;
esac