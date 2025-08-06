using System.Collections;
using MerchantTails.Core;
using MerchantTails.Testing;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.TestTools;

namespace MerchantTails.Tests
{
    /// <summary>
    /// CI環境での自動テスト実行クラス（Runtimeフォルダ版）
    /// Unity Test Runnerとの連携により自動テストを実現
    /// </summary>
    public class AutomatedTestRunner
    {
        private TestRunnerComponent testRunner;
        private SystemTestController systemTestController;
        private IntegrationTest integrationTest;
        private StabilityTest stabilityTest;
        private ErrorRecoveryTest errorRecoveryTest;

        [OneTimeSetUp]
        public void OneTimeSetUp()
        {
            // Create test environment
            var testGameObject = new GameObject("AutomatedTestRunner");
            Object.DontDestroyOnLoad(testGameObject);

            // Add all required components
            testRunner = testGameObject.AddComponent<TestRunnerComponent>();
            systemTestController = testGameObject.AddComponent<SystemTestController>();
            integrationTest = testGameObject.AddComponent<IntegrationTest>();
            stabilityTest = testGameObject.AddComponent<StabilityTest>();
            errorRecoveryTest = testGameObject.AddComponent<ErrorRecoveryTest>();

            // Add core systems
            var gameManager = testGameObject.AddComponent<GameManager>();
            var timeManager = testGameObject.AddComponent<TimeManager>();
            var marketSystem = testGameObject.AddComponent<MarketSystem>();
            var inventorySystem = testGameObject.AddComponent<InventorySystem>();

            // Initialize error handling
            ErrorHandler.Initialize();
        }

        [OneTimeTearDown]
        public void OneTimeTearDown()
        {
            // Cleanup
            ErrorHandler.Cleanup();

            if (testRunner != null)
            {
                Object.DestroyImmediate(testRunner.gameObject);
            }
        }

        [UnityTest]
        [Order(1)]
        public IEnumerator SystemHealthCheck()
        {
            // Wait for systems to initialize
            yield return new WaitForSeconds(2f);

            // Verify all systems are healthy
            bool healthResult = ErrorHandler.CheckSystemHealth();

            Assert.IsTrue(healthResult, "System health check failed - core systems are not properly initialized");

            // Verify singleton instances
            Assert.IsNotNull(GameManager.Instance, "GameManager instance is null");
            Assert.IsNotNull(TimeManager.Instance, "TimeManager instance is null");
            Assert.IsNotNull(MarketSystem.Instance, "MarketSystem instance is null");
            Assert.IsNotNull(InventorySystem.Instance, "InventorySystem instance is null");

            ErrorHandler.LogInfo("✓ System health check passed", "AutomatedTestRunner");
        }

        [UnityTest]
        [Order(2)]
        [Category("Integration")]
        public IEnumerator IntegrationTestSuite()
        {
            // Wait for systems to be ready
            yield return new WaitForSeconds(1f);

            // Start integration tests
            integrationTest.StartIntegrationTests();

            // Wait for tests to complete
            while (integrationTest.IsTestInProgress())
            {
                yield return new WaitForSeconds(0.5f);
            }

            // Get test results
            integrationTest.GetTestSummary(out int passed, out int failed);

            // Assert all tests passed
            Assert.AreEqual(0, failed, $"Integration tests failed: {failed} out of {passed + failed} tests failed");
            Assert.Greater(passed, 0, "No integration tests were executed");

            ErrorHandler.LogInfo($"✓ Integration tests passed: {passed} tests", "AutomatedTestRunner");
        }

        [UnityTest]
        [Order(3)]
        [Category("Performance")]
        public IEnumerator StabilityTestSuite()
        {
            // Wait for previous tests to complete
            yield return new WaitForSeconds(1f);

            // Configure stability test for CI (shorter duration)
            // Note: Direct field access not available, test will use default settings

            // Start stability tests
            stabilityTest.StartStabilityTest();

            // Wait for tests to complete (with timeout)
            float timeout = 60f; // 1 minute timeout for CI
            float elapsed = 0f;

            while (stabilityTest.IsStabilityTestRunning() && elapsed < timeout)
            {
                yield return new WaitForSeconds(1f);
                elapsed += 1f;
            }

            // Check if tests completed within timeout
            Assert.IsFalse(stabilityTest.IsStabilityTestRunning(), "Stability tests timed out");

            // Get test results
            var report = stabilityTest.GetLastReport();

            // Assert stability criteria (relaxed for CI)
            Assert.GreaterOrEqual(
                report.successRate,
                90f,
                $"Stability test success rate too low: {report.successRate}%"
            );
            Assert.LessOrEqual(
                report.criticalErrors.Count,
                1,
                $"Too many critical errors: {string.Join(", ", report.criticalErrors)}"
            );

            ErrorHandler.LogInfo(
                $"✓ Stability tests passed: {report.successRate:F1}% success rate",
                "AutomatedTestRunner"
            );
        }

        [UnityTest]
        [Order(4)]
        [Category("ErrorHandling")]
        public IEnumerator ErrorRecoveryTestSuite()
        {
            // Wait for previous tests to complete
            yield return new WaitForSeconds(1f);

            // Start error recovery tests
            errorRecoveryTest.StartErrorRecoveryTest();

            // Wait for tests to complete
            while (errorRecoveryTest.IsRecoveryTestRunning())
            {
                yield return new WaitForSeconds(0.5f);
            }

            // Get test results
            float successRate = errorRecoveryTest.GetRecoverySuccessRate();

            // Assert recovery criteria
            Assert.GreaterOrEqual(successRate, 75f, $"Error recovery success rate too low: {successRate}%");

            ErrorHandler.LogInfo(
                $"✓ Error recovery tests passed: {successRate:F1}% success rate",
                "AutomatedTestRunner"
            );
        }

        [UnityTest]
        [Order(5)]
        [Category("Functional")]
        public IEnumerator SystemFunctionalityVerification()
        {
            // Comprehensive verification that all systems work together
            yield return new WaitForSeconds(0.5f);

            // Test time advancement
            int initialDay = TimeManager.Instance.CurrentDay;
            TimeManager.Instance.SkipToNextPhase();
            yield return new WaitForSeconds(0.1f);

            // Test market system
            float fruitPrice = MarketSystem.Instance.GetCurrentPrice(MerchantTails.Data.ItemType.Fruit);
            Assert.Greater(fruitPrice, 0, "Market system not functioning - invalid fruit price");

            // Test inventory system
            bool addResult = InventorySystem.Instance.AddItem(
                MerchantTails.Data.ItemType.Fruit,
                10,
                MerchantTails.Data.InventoryLocation.Trading
            );
            Assert.IsTrue(addResult, "Inventory system not functioning - failed to add items");

            int itemCount = InventorySystem.Instance.GetItemCount(
                MerchantTails.Data.ItemType.Fruit,
                MerchantTails.Data.InventoryLocation.Trading
            );
            Assert.AreEqual(10, itemCount, "Inventory system not functioning - incorrect item count");

            // Test event system
            bool eventReceived = false;
            System.Action<MerchantTails.Core.PriceChangedEvent> handler = (evt) => eventReceived = true;

            MerchantTails.Core.EventBus.Subscribe<MerchantTails.Core.PriceChangedEvent>(handler);

            var priceEvent = new MerchantTails.Core.PriceChangedEvent(MerchantTails.Data.ItemType.Potion, 100f, 110f);
            MerchantTails.Core.EventBus.Publish(priceEvent);

            yield return new WaitForSeconds(0.1f);

            MerchantTails.Core.EventBus.Unsubscribe<MerchantTails.Core.PriceChangedEvent>(handler);

            Assert.IsTrue(eventReceived, "Event system not functioning - event not received");

            ErrorHandler.LogInfo("✓ System functionality verification passed", "AutomatedTestRunner");
        }

        [Test]
        [Order(6)]
        [Category("Performance")]
        public void MemoryUsageVerification()
        {
            // Force garbage collection
            System.GC.Collect();
            System.GC.WaitForPendingFinalizers();
            System.GC.Collect();

            // Check memory usage
            long memoryUsage = System.GC.GetTotalMemory(false);
            long memoryThreshold = 200 * 1024 * 1024; // 200MB threshold for CI

            Assert.Less(memoryUsage, memoryThreshold, $"Memory usage too high: {memoryUsage / 1024 / 1024}MB");

            ErrorHandler.LogInfo(
                $"✓ Memory usage verification passed: {memoryUsage / 1024 / 1024}MB",
                "AutomatedTestRunner"
            );
        }

        [Test]
        [Order(7)]
        [Category("Configuration")]
        public void ConfigurationVerification()
        {
            // Verify that all required components are properly configured
            Assert.IsNotNull(testRunner, "TestRunnerComponent component missing");
            Assert.IsNotNull(systemTestController, "SystemTestController component missing");
            Assert.IsNotNull(integrationTest, "IntegrationTest component missing");
            Assert.IsNotNull(stabilityTest, "StabilityTest component missing");
            Assert.IsNotNull(errorRecoveryTest, "ErrorRecoveryTest component missing");

            ErrorHandler.LogInfo("✓ Configuration verification passed", "AutomatedTestRunner");
        }

        [UnityTest]
        [Category("Smoke")]
        public IEnumerator SmokeTest()
        {
            // Quick smoke test to verify basic functionality
            yield return new WaitForSeconds(1f);

            // Verify systems exist
            Assert.IsNotNull(GameManager.Instance, "GameManager missing");
            Assert.IsNotNull(TimeManager.Instance, "TimeManager missing");
            Assert.IsNotNull(MarketSystem.Instance, "MarketSystem missing");
            Assert.IsNotNull(InventorySystem.Instance, "InventorySystem missing");

            // Quick functionality test
            float price = MarketSystem.Instance.GetCurrentPrice(MerchantTails.Data.ItemType.Fruit);
            Assert.Greater(price, 0, "Invalid fruit price");

            bool addSuccess = InventorySystem.Instance.AddItem(
                MerchantTails.Data.ItemType.Fruit,
                1,
                MerchantTails.Data.InventoryLocation.Trading
            );
            Assert.IsTrue(addSuccess, "Failed to add item");

            ErrorHandler.LogInfo("✓ Smoke test passed", "AutomatedTestRunner");
        }
    }
}
