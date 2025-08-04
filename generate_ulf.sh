#\!/bin/bash
# Unity Personalライセンスの.ulfファイルを生成

# Unity Hubにログインしている状態で実行
/Applications/Unity/Hub/Editor/6000.1.14f1/Unity.app/Contents/MacOS/Unity \
  -batchmode \
  -quit \
  -manualLicenseFile \
  -logFile unity_license.log

# ライセンスファイルの場所を確認
echo "Checking for license files..."
find ~/Library -name "*.ulf" -type f 2>/dev/null | grep -i unity
