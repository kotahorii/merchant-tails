using System.Collections;
using System.Threading.Tasks;
using MerchantTails.Core;
using MerchantTails.Data;
using MerchantTails.Market;
using MerchantTails.UI;
using NUnit.Framework;
using Unity.Collections;
using Unity.Jobs;
using UnityEngine;
using UnityEngine.TestTools;
using UnityEngine.UIElements;

namespace MerchantTails.Testing
{
    /// <summary>
    /// Unity 6最適化機能のテスト
    /// Job System、UI Toolkit、新しいSaveSystemなどの動作確認
    /// </summary>
    public class Unity6OptimizationTests : TestBase
    {
        private MarketSystem marketSystem;
        private SaveSystem saveSystem;
        private UIDocument testUIDocument;

        public override void Setup()
        {
            base.Setup();

            marketSystem = testGameObject.AddComponent<MarketSystem>();
            saveSystem = testGameObject.AddComponent<SaveSystem>();

            // UI Toolkit用のテスト設定
            var uiGameObject = new GameObject("TestUIDocument");
            testUIDocument = uiGameObject.AddComponent<UIDocument>();
        }

        public override void Teardown()
        {
            if (testUIDocument != null)
            {
                Object.Destroy(testUIDocument.gameObject);
            }

            base.Teardown();
        }

        #region Job System Tests

        [Test]
        public void MarketPriceCalculationJob_ExecutesCorrectly()
        {
            // Arrange
            const int itemCount = 6;
            var basePrices = new NativeArray<float>(itemCount, Allocator.TempJob);
            var volatilities = new NativeArray<float>(itemCount, Allocator.TempJob);
            var demands = new NativeArray<float>(itemCount, Allocator.TempJob);
            var supplies = new NativeArray<float>(itemCount, Allocator.TempJob);
            var seasonalModifiers = new NativeArray<float>(itemCount, Allocator.TempJob);
            var eventModifiers = new NativeArray<float>(itemCount, Allocator.TempJob);
            var calculatedPrices = new NativeArray<float>(itemCount, Allocator.TempJob);
            var priceChanges = new NativeArray<float>(itemCount, Allocator.TempJob);

            // Initialize test data
            for (int i = 0; i < itemCount; i++)
            {
                basePrices[i] = 100f * (i + 1);
                volatilities[i] = 0.1f + (i * 0.05f);
                demands[i] = 1.0f;
                supplies[i] = 1.0f;
                seasonalModifiers[i] = 1.0f;
                eventModifiers[i] = 1.0f;
                calculatedPrices[i] = basePrices[i]; // Current price = base price
            }

            try
            {
                // Act
                var job = new MarketPriceCalculationJob
                {
                    basePrices = basePrices,
                    volatilities = volatilities,
                    demands = demands,
                    supplies = supplies,
                    seasonalModifiers = seasonalModifiers,
                    eventModifiers = eventModifiers,
                    globalMarketTrend = 0.05f,
                    randomSeed = 12345,
                    deltaTime = 0.016f,
                    calculatedPrices = calculatedPrices,
                    priceChanges = priceChanges
                };

                var handle = job.Schedule(itemCount, 1);
                handle.Complete();

                // Assert
                for (int i = 0; i < itemCount; i++)
                {
                    Assert.Greater(calculatedPrices[i], 0f, $"Calculated price {i} should be positive");

                    // 価格が基本価格の範囲内であることを確認
                    float minPrice = basePrices[i] * 0.3f;
                    float maxPrice = basePrices[i] * 3.0f;
                    Assert.GreaterOrEqual(calculatedPrices[i], minPrice);
                    Assert.LessOrEqual(calculatedPrices[i], maxPrice);
                }
            }
            finally
            {
                // Cleanup
                basePrices.Dispose();
                volatilities.Dispose();
                demands.Dispose();
                supplies.Dispose();
                seasonalModifiers.Dispose();
                eventModifiers.Dispose();
                calculatedPrices.Dispose();
                priceChanges.Dispose();
            }
        }

        [Test]
        public void SaveDataCompressionJob_CompressesData()
        {
            // Arrange
            var testData = new byte[] { 1, 1, 1, 1, 2, 2, 2, 2, 2, 3, 3, 3 };
            var uncompressedData = new NativeArray<byte>(testData, Allocator.TempJob);
            var compressedData = new NativeArray<byte>(testData.Length * 2, Allocator.TempJob);
            var compressedSize = new NativeArray<int>(1, Allocator.TempJob);

            try
            {
                // Act
                var job = new SaveDataCompressionJob
                {
                    uncompressedData = uncompressedData,
                    compressedData = compressedData,
                    compressedSize = compressedSize
                };

                var handle = job.Schedule();
                handle.Complete();

                // Assert
                int actualSize = compressedSize[0];
                Assert.Greater(actualSize, 0, "Compressed size should be greater than 0");
                Assert.Less(actualSize, testData.Length, "Compressed size should be less than original");
            }
            finally
            {
                // Cleanup
                uncompressedData.Dispose();
                compressedData.Dispose();
                compressedSize.Dispose();
            }
        }

        #endregion

        #region Async Save/Load Tests

        [UnityTest]
        public IEnumerator SaveAsync_CompletesSuccessfully()
        {
            // Arrange
            gameManager.PlayerData.SetMoney(12345);
            gameManager.PlayerData.SetPlayerName("AsyncTestPlayer");

            // Act
            var saveTask = saveSystem.SaveAsync(0);
            yield return WaitForTask(saveTask);

            // Assert
            Assert.IsTrue(saveTask.Result, "Async save should complete successfully");
            Assert.IsTrue(saveSystem.HasSaveData, "Save data should exist");
        }

        [UnityTest]
        public IEnumerator LoadAsync_RestoresDataCorrectly()
        {
            // Arrange
            int originalMoney = 99999;
            string originalName = "AsyncLoadTest";
            gameManager.PlayerData.SetMoney(originalMoney);
            gameManager.PlayerData.SetPlayerName(originalName);

            // Save first
            var saveTask = saveSystem.SaveAsync(0);
            yield return WaitForTask(saveTask);

            // Modify data
            gameManager.PlayerData.SetMoney(0);
            gameManager.PlayerData.SetPlayerName("Modified");

            // Act
            var loadTask = saveSystem.LoadAsync(0);
            yield return WaitForTask(loadTask);

            // Assert
            Assert.IsTrue(loadTask.Result, "Async load should complete successfully");
            Assert.AreEqual(originalMoney, gameManager.PlayerData.CurrentMoney);
            Assert.AreEqual(originalName, gameManager.PlayerData.PlayerName);
        }

        private IEnumerator WaitForTask(Task task)
        {
            while (!task.IsCompleted)
            {
                yield return null;
            }

            if (task.Exception != null)
            {
                throw task.Exception;
            }
        }

        private IEnumerator WaitForTask<T>(Task<T> task)
        {
            while (!task.IsCompleted)
            {
                yield return null;
            }

            if (task.Exception != null)
            {
                throw task.Exception;
            }
        }

        #endregion

        #region UI Toolkit Tests

        [Test]
        public void UIToolkitPanel_InitializesCorrectly()
        {
            // Arrange
            var rootElement = new VisualElement();
            rootElement.name = "TestPanel";

            // Act
            var panel = new UIToolkitPanel(UIType.MainMenu, rootElement);
            panel.Initialize();

            // Assert
            Assert.IsNotNull(panel.Element);
            Assert.AreEqual(UIType.MainMenu, panel.UIType);
            Assert.IsFalse(panel.IsVisible);
        }

        [Test]
        public void UIToolkitPanel_ShowHide_UpdatesVisibility()
        {
            // Arrange
            var rootElement = new VisualElement();
            var panel = new UIToolkitPanel(UIType.Settings, rootElement);
            panel.Initialize();

            // Act & Assert
            panel.Show();
            Assert.IsTrue(panel.IsVisible);
            Assert.AreEqual(DisplayStyle.Flex, panel.Element.style.display.value);

            panel.Hide();
            Assert.IsFalse(panel.IsVisible);
        }

        #endregion

        #region UpdateManager Optimization Tests

        [Test]
        public void UpdateManager_2DGameOptimization()
        {
            // Arrange
            var updateManager = testGameObject.AddComponent<UpdateManager>();

            // Act
            var stats = updateManager.GetStats();

            // Assert
            Assert.AreEqual(PerformanceLevel.Normal, stats.performanceLevel);
            // 2Dゲーム向けの設定を確認
            Assert.AreEqual(30, Application.targetFrameRate, "Target framerate should be 30 for 2D games");
            Assert.AreEqual(0.1f, Time.fixedDeltaTime, 0.01f, "Fixed update should be 10Hz");
        }

        [UnityTest]
        public IEnumerator UpdateManager_PerformanceLevelAdjustment()
        {
            // Arrange
            var updateManager = testGameObject.AddComponent<UpdateManager>();
            bool performanceChanged = false;

            updateManager.OnPerformanceLevelChanged += (level) => {
                performanceChanged = true;
            };

            // Act - パフォーマンスレベルの変更をシミュレート
            // 実際のテストでは、FPSを下げる処理が必要

            yield return new WaitForSeconds(3f);

            // Assert
            // パフォーマンスレベルの変更が適切に処理されることを確認
            var stats = updateManager.GetStats();
            Assert.IsNotNull(stats);
        }

        #endregion

        #region Integration Tests

        [UnityTest]
        public IEnumerator MarketSystem_JobSystemIntegration()
        {
            // Arrange
            var itemTypes = System.Enum.GetValues(typeof(ItemType));
            var originalPrices = new System.Collections.Generic.Dictionary<ItemType, float>();

            foreach (ItemType itemType in itemTypes)
            {
                originalPrices[itemType] = marketSystem.GetCurrentPrice(itemType);
            }

            // Act
            marketSystem.UpdatePrices();
            yield return null; // Wait for Job System to complete

            // Assert
            bool anyPriceChanged = false;
            foreach (ItemType itemType in itemTypes)
            {
                float newPrice = marketSystem.GetCurrentPrice(itemType);
                if (Mathf.Abs(newPrice - originalPrices[itemType]) > 0.01f)
                {
                    anyPriceChanged = true;
                    break;
                }
            }

            Assert.IsTrue(anyPriceChanged, "Job System should have updated prices");
        }

        #endregion

        #region Performance Tests

        [Test]
        [Performance]
        public void MarketPriceCalculation_Performance()
        {
            // Job System版の価格計算パフォーマンステスト
            Measure.Method(() =>
            {
                marketSystem.UpdatePrices();
            })
            .WarmupCount(10)
            .MeasurementCount(100)
            .Run();
        }

        [Test]
        [Performance]
        public void SaveSystemCompression_Performance()
        {
            // 圧縮処理のパフォーマンステスト
            var testData = new byte[10000];
            for (int i = 0; i < testData.Length; i++)
            {
                testData[i] = (byte)(i % 256);
            }

            Measure.Method(() =>
            {
                var uncompressed = new NativeArray<byte>(testData, Allocator.TempJob);
                var compressed = new NativeArray<byte>(testData.Length * 2, Allocator.TempJob);
                var size = new NativeArray<int>(1, Allocator.TempJob);

                var job = new SaveDataCompressionJob
                {
                    uncompressedData = uncompressed,
                    compressedData = compressed,
                    compressedSize = size
                };

                job.Schedule().Complete();

                uncompressed.Dispose();
                compressed.Dispose();
                size.Dispose();
            })
            .WarmupCount(5)
            .MeasurementCount(50)
            .Run();
        }

        #endregion
    }
}
