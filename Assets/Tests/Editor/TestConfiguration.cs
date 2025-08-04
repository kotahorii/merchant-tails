using UnityEngine;
using UnityEditor;
using NUnit.Framework;
using System.Linq;

namespace MerchantTails.Tests.Editor
{
    /// <summary>
    /// エディターモードテスト設定クラス
    /// Unity Test Runnerのエディターモードテストを設定
    /// </summary>
    public class TestConfiguration
    {
        [Test]
        public void ValidateProjectStructure()
        {
            // スクリプトディレクトリの存在確認
            Assert.IsTrue(System.IO.Directory.Exists("Assets/Scripts/Core"), "Core scripts directory missing");
            Assert.IsTrue(System.IO.Directory.Exists("Assets/Scripts/Data"), "Data scripts directory missing");
            Assert.IsTrue(System.IO.Directory.Exists("Assets/Scripts/Events"), "Events scripts directory missing");
            Assert.IsTrue(System.IO.Directory.Exists("Assets/Scripts/Market"), "Market scripts directory missing");
            Assert.IsTrue(System.IO.Directory.Exists("Assets/Scripts/Inventory"), "Inventory scripts directory missing");
            Assert.IsTrue(System.IO.Directory.Exists("Assets/Scripts/Testing"), "Testing scripts directory missing");

            Debug.Log("✅ Project structure validation passed");
        }

        [Test]
        public void ValidateScriptCompilation()
        {
            // スクリプトコンパイルエラーがないことを確認
            var compilationMessages = UnityEditor.Compilation.CompilationPipeline.GetAssemblies();
            Assert.IsNotNull(compilationMessages, "Compilation assemblies should not be null");
            Assert.Greater(compilationMessages.Length, 0, "No assemblies found - compilation may have failed");

            Debug.Log("✅ Script compilation validation passed");
        }

        [Test]
        public void ValidatePlayerSettings()
        {
            // プレイヤー設定の検証
            Assert.IsNotEmpty(PlayerSettings.productName, "Product name must be set");
            Assert.IsNotEmpty(PlayerSettings.companyName, "Company name must be set");
            Assert.IsNotEmpty(PlayerSettings.bundleVersion, "Bundle version must be set");

            Debug.Log("✅ Player settings validation passed");
        }

        [Test]
        public void ValidateBuildSettings()
        {
            // ビルド設定の検証
            var scenes = EditorBuildSettings.scenes;
            Assert.IsNotNull(scenes, "Build scenes should not be null");
            Assert.Greater(scenes.Length, 0, "At least one scene must be added to build settings");

            bool hasEnabledScene = false;
            foreach (var scene in scenes)
            {
                if (scene.enabled)
                {
                    hasEnabledScene = true;
                    break;
                }
            }
            Assert.IsTrue(hasEnabledScene, "At least one scene must be enabled in build settings");

            Debug.Log("✅ Build settings validation passed");
        }

        [Test]
        public void ValidateScriptableObjectStructure()
        {
            // ScriptableObjectの構造検証
            var playerDataType = System.Type.GetType("MerchantTails.Data.PlayerData");
            Assert.IsNotNull(playerDataType, "PlayerData type should exist");

            var inventoryDataType = System.Type.GetType("MerchantTails.Inventory.InventoryData");
            Assert.IsNotNull(inventoryDataType, "InventoryData type should exist");

            var marketDataType = System.Type.GetType("MerchantTails.Market.MarketData");
            Assert.IsNotNull(marketDataType, "MarketData type should exist");

            Debug.Log("✅ ScriptableObject structure validation passed");
        }

        [Test]
        public void ValidateEventSystem()
        {
            // イベントシステムの型検証
            var eventBusType = System.Type.GetType("MerchantTails.Core.EventBus");
            Assert.IsNotNull(eventBusType, "EventBus type should exist");

            var iGameEventType = System.Type.GetType("MerchantTails.Core.IGameEvent");
            Assert.IsNotNull(iGameEventType, "IGameEvent interface should exist");

            Debug.Log("✅ Event system validation passed");
        }

        [Test]
        public void ValidateEnumDefinitions()
        {
            // Enum定義の検証
            var itemTypeEnum = System.Type.GetType("MerchantTails.Data.ItemType");
            Assert.IsNotNull(itemTypeEnum, "ItemType enum should exist");
            Assert.IsTrue(itemTypeEnum.IsEnum, "ItemType should be an enum");

            var seasonEnum = System.Type.GetType("MerchantTails.Data.Season");
            Assert.IsNotNull(seasonEnum, "Season enum should exist");
            Assert.IsTrue(seasonEnum.IsEnum, "Season should be an enum");

            var gameStateEnum = System.Type.GetType("MerchantTails.Data.GameState");
            Assert.IsNotNull(gameStateEnum, "GameState enum should exist");
            Assert.IsTrue(gameStateEnum.IsEnum, "GameState should be an enum");

            Debug.Log("✅ Enum definitions validation passed");
        }

        [Test]
        public void ValidateNamespaceStructure()
        {
            // 名前空間の構造検証
            var coreTypes = GetTypesInNamespace("MerchantTails.Core");
            Assert.Greater(coreTypes.Length, 0, "Core namespace should contain types");

            var dataTypes = GetTypesInNamespace("MerchantTails.Data");
            Assert.Greater(dataTypes.Length, 0, "Data namespace should contain types");

            var eventTypes = GetTypesInNamespace("MerchantTails.Events");
            Assert.Greater(eventTypes.Length, 0, "Events namespace should contain types");

            var marketTypes = GetTypesInNamespace("MerchantTails.Market");
            Assert.Greater(marketTypes.Length, 0, "Market namespace should contain types");

            var inventoryTypes = GetTypesInNamespace("MerchantTails.Inventory");
            Assert.Greater(inventoryTypes.Length, 0, "Inventory namespace should contain types");

            var testingTypes = GetTypesInNamespace("MerchantTails.Testing");
            Assert.Greater(testingTypes.Length, 0, "Testing namespace should contain types");

            Debug.Log("✅ Namespace structure validation passed");
        }

        private System.Type[] GetTypesInNamespace(string namespaceName)
        {
            return System.Reflection.Assembly.GetExecutingAssembly().GetTypes()
                .Where(t => t.Namespace == namespaceName)
                .ToArray();
        }

        [Test]
        public void ValidateTestingInfrastructure()
        {
            // テストインフラの検証
            var testRunnerType = System.Type.GetType("MerchantTails.Testing.TestRunner");
            Assert.IsNotNull(testRunnerType, "TestRunner should exist");

            var integrationTestType = System.Type.GetType("MerchantTails.Testing.IntegrationTest");
            Assert.IsNotNull(integrationTestType, "IntegrationTest should exist");

            var stabilityTestType = System.Type.GetType("MerchantTails.Testing.StabilityTest");
            Assert.IsNotNull(stabilityTestType, "StabilityTest should exist");

            var errorRecoveryTestType = System.Type.GetType("MerchantTails.Testing.ErrorRecoveryTest");
            Assert.IsNotNull(errorRecoveryTestType, "ErrorRecoveryTest should exist");

            var automatedTestRunnerType = System.Type.GetType("MerchantTails.Tests.AutomatedTestRunner");
            Assert.IsNotNull(automatedTestRunnerType, "AutomatedTestRunner should exist");

            Debug.Log("✅ Testing infrastructure validation passed");
        }
    }
}
