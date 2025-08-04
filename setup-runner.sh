#!/bin/bash

# GitHub Actions Self-hosted Runner セットアップスクリプト
# 注意: GitHubのSettings > Actions > Runnersから最新のトークンを取得してください

echo "GitHub Actions Self-hosted Runner セットアップ"
echo "=============================================="
echo ""
echo "1. https://github.com/kotahorii/merchant-tails/settings/actions/runners にアクセス"
echo "2. 'New self-hosted runner' をクリック"
echo "3. macOSを選択"
echo "4. 表示されるトークンをコピー"
echo ""
read -p "トークンを入力してください: " RUNNER_TOKEN

if [ -z "$RUNNER_TOKEN" ]; then
    echo "エラー: トークンが入力されていません"
    exit 1
fi

# Runnerディレクトリの作成
mkdir -p ~/actions-runner
cd ~/actions-runner

# 最新のrunnerをダウンロード
echo "Runnerをダウンロード中..."
curl -o actions-runner-osx-arm64-2.319.1.tar.gz -L https://github.com/actions/runner/releases/download/v2.319.1/actions-runner-osx-arm64-2.319.1.tar.gz

# 展開
echo "展開中..."
tar xzf ./actions-runner-osx-arm64-2.319.1.tar.gz

# 設定
echo "Runner設定中..."
./config.sh --url https://github.com/kotahorii/merchant-tails --token $RUNNER_TOKEN --name "unity-build-mac" --labels "self-hosted,macOS,ARM64,unity" --work "_work"

# サービスとして実行
echo ""
echo "Runnerを起動するには以下のコマンドを実行してください:"
echo "cd ~/actions-runner && ./run.sh"
echo ""
echo "バックグラウンドで実行する場合:"
echo "./svc.sh install"
echo "./svc.sh start"