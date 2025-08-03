using System.Collections;
using System.Diagnostics;
using MerchantTails.Core;
using MerchantTails.Data;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.TestTools;
using Debug = UnityEngine.Debug;

namespace MerchantTails.Testing
{
    /// <summary>
    /// パフォーマンステスト
    /// 各システムの処理速度とメモリ使用量を測定
    /// </summary>
    public class PerformanceTests : TestBase
    {
        private Stopwatch stopwatch;
        private const int PERFORMANCE_TEST_ITERATIONS = 1000;
        private const float MAX_ACCEPTABLE_TIME_MS = 16.67f; // 60 FPS

        public override void Setup()
        {
            base.Setup();
            stopwatch = new Stopwatch();
        }

        [Test]
        public void MarketPriceUpdate_Performance()
        {
            // Warm up
            marketSystem.UpdatePrices();

            // Measure
            stopwatch.Reset();
            stopwatch.Start();

            for (int i = 0; i < PERFORMANCE_TEST_ITERATIONS; i++)
            {
                marketSystem.UpdatePrices();
            }

            stopwatch.Stop();

            float averageTime = (float)stopwatch.ElapsedMilliseconds / PERFORMANCE_TEST_ITERATIONS;
            Debug.Log($"MarketSystem.UpdatePrices average time: {averageTime:F3}ms");

            Assert.Less(averageTime, MAX_ACCEPTABLE_TIME_MS,
                $"Price update should complete within {MAX_ACCEPTABLE_TIME_MS}ms");
        }

        [Test]
        public void InventoryOperations_Performance()
        {
            // Prepare test items
            var testItems = new ItemData[100];
            for (int i = 0; i < testItems.Length; i++)
            {
                testItems[i] = CreateTestItem((ItemType)(i % 6), 10);
            }

            // Measure add operations
            stopwatch.Reset();
            stopwatch.Start();

            foreach (var item in testItems)
            {
                inventorySystem.AddItem(item);
            }

            stopwatch.Stop();
            float addTime = (float)stopwatch.ElapsedMilliseconds / testItems.Length;
            Debug.Log($"InventorySystem.AddItem average time: {addTime:F3}ms");

            // Measure query operations
            stopwatch.Reset();
            stopwatch.Start();

            for (int i = 0; i < PERFORMANCE_TEST_ITERATIONS; i++)
            {
                inventorySystem.GetAllItems();
                inventorySystem.GetTotalValue();
            }

            stopwatch.Stop();
            float queryTime = (float)stopwatch.ElapsedMilliseconds / PERFORMANCE_TEST_ITERATIONS;
            Debug.Log($"Inventory query average time: {queryTime:F3}ms");

            Assert.Less(addTime, 1f, "Add item should complete within 1ms");
            Assert.Less(queryTime, 0.5f, "Query operations should complete within 0.5ms");
        }

        [Test]
        public void TimeAdvancement_Performance()
        {
            // Measure single hour advancement
            stopwatch.Reset();
            stopwatch.Start();

            for (int i = 0; i < PERFORMANCE_TEST_ITERATIONS; i++)
            {
                timeManager.AdvanceTime(0.01f); // Small time increment
            }

            stopwatch.Stop();

            float averageTime = (float)stopwatch.ElapsedMilliseconds / PERFORMANCE_TEST_ITERATIONS;
            Debug.Log($"TimeManager.AdvanceTime average time: {averageTime:F3}ms");

            Assert.Less(averageTime, 0.1f,
                "Time advancement should be very fast (< 0.1ms)");
        }

        [UnityTest]
        public IEnumerator MemoryAllocation_DuringGameplay()
        {
            // Get initial memory
            System.GC.Collect();
            yield return null;
            long initialMemory = System.GC.GetTotalMemory(false);

            // Simulate 1 minute of gameplay
            float simulationTime = 60f;
            float elapsed = 0f;

            while (elapsed < simulationTime)
            {
                // Normal game operations
                timeManager.AdvanceTime(Time.deltaTime * 10); // 10x speed

                if (Random.Range(0f, 1f) < 0.1f) // 10% chance per frame
                {
                    marketSystem.UpdatePrices();
                }

                if (Random.Range(0f, 1f) < 0.05f) // 5% chance per frame
                {
                    var item = CreateTestItem((ItemType)Random.Range(0, 6), Random.Range(1, 10));
                    inventorySystem.AddItem(item);
                }

                elapsed += Time.deltaTime;
                yield return null;
            }

            // Measure memory growth
            long finalMemory = System.GC.GetTotalMemory(false);
            long memoryGrowth = finalMemory - initialMemory;
            float memoryGrowthMB = memoryGrowth / (1024f * 1024f);

            Debug.Log($"Memory growth during 1 minute: {memoryGrowthMB:F2} MB");

            // Allow some memory growth but flag excessive allocation
            Assert.Less(memoryGrowthMB, 10f,
                "Memory growth should be less than 10MB per minute");
        }

        [UnityTest]
        public IEnumerator FrameRate_UnderLoad()
        {
            var frameRateStabilizer = testGameObject.AddComponent<FrameRateStabilizer>();
            var updateManager = testGameObject.AddComponent<UpdateManager>();

            float testDuration = 5f;
            float elapsed = 0f;
            int frameCount = 0;
            float totalDeltaTime = 0f;

            // Create load
            for (int i = 0; i < 100; i++)
            {
                var item = CreateTestItem((ItemType)(i % 6), 50);
                inventorySystem.AddItem(item);
            }

            while (elapsed < testDuration)
            {
                // Heavy operations each frame
                marketSystem.UpdatePrices();
                inventorySystem.GetTotalValue();

                for (int i = 0; i < 10; i++)
                {
                    marketSystem.GetCurrentPrice((ItemType)(i % 6));
                }

                frameCount++;
                totalDeltaTime += Time.deltaTime;
                elapsed += Time.deltaTime;

                yield return null;
            }

            float averageFPS = frameCount / totalDeltaTime;
            Debug.Log($"Average FPS under load: {averageFPS:F1}");

            // Should maintain at least 30 FPS
            Assert.Greater(averageFPS, 30f,
                "Should maintain at least 30 FPS under load");
        }

        [Test]
        public void SaveSystem_Performance()
        {
            var saveSystem = testGameObject.AddComponent<SaveSystem>();

            // Prepare complex game state
            gameManager.PlayerData.SetMoney(99999);
            gameManager.PlayerData.SetRank(MerchantRank.Master);

            for (int i = 0; i < 50; i++)
            {
                var item = CreateTestItem((ItemType)(i % 6), Random.Range(1, 100));
                inventorySystem.AddItem(item);
            }

            // Measure save performance
            stopwatch.Reset();
            stopwatch.Start();

            saveSystem.Save(0);

            stopwatch.Stop();
            Debug.Log($"SaveSystem.Save time: {stopwatch.ElapsedMilliseconds}ms");

            // Measure load performance
            stopwatch.Reset();
            stopwatch.Start();

            saveSystem.Load(0);

            stopwatch.Stop();
            Debug.Log($"SaveSystem.Load time: {stopwatch.ElapsedMilliseconds}ms");

            Assert.Less(stopwatch.ElapsedMilliseconds, 100f,
                "Save/Load should complete within 100ms");
        }

        [Test]
        public void EventSystem_BroadcastPerformance()
        {
            // Subscribe many listeners
            int listenerCount = 100;
            int eventReceivedCount = 0;

            for (int i = 0; i < listenerCount; i++)
            {
                EventBus.Subscribe<PriceChangedEvent>((e) => { eventReceivedCount++; });
            }

            // Measure broadcast time
            stopwatch.Reset();
            stopwatch.Start();

            for (int i = 0; i < PERFORMANCE_TEST_ITERATIONS; i++)
            {
                EventBus.Publish(new PriceChangedEvent(ItemType.Fruit, 10f, 12f));
            }

            stopwatch.Stop();

            float averageTime = (float)stopwatch.ElapsedMilliseconds / PERFORMANCE_TEST_ITERATIONS;
            Debug.Log($"EventBus broadcast average time: {averageTime:F3}ms with {listenerCount} listeners");

            Assert.Less(averageTime, 1f,
                "Event broadcast should complete within 1ms even with many listeners");
        }

        [Test]
        public void LocalizationLookup_Performance()
        {
            var localizationManager = testGameObject.AddComponent<LocalizationManager>();

            // Warm up
            localizationManager.GetText("test.key");

            // Measure lookup performance
            stopwatch.Reset();
            stopwatch.Start();

            for (int i = 0; i < PERFORMANCE_TEST_ITERATIONS * 10; i++)
            {
                localizationManager.GetText("item.fruit");
                localizationManager.GetText("ui.money");
                localizationManager.GetText("rank.master");
            }

            stopwatch.Stop();

            float totalLookups = PERFORMANCE_TEST_ITERATIONS * 10 * 3;
            float averageTime = (float)stopwatch.ElapsedMilliseconds / totalLookups;
            Debug.Log($"Localization lookup average time: {averageTime:F4}ms");

            Assert.Less(averageTime, 0.01f,
                "Localization lookup should be very fast (< 0.01ms)");
        }

        [UnityTest]
        public IEnumerator UpdateManager_Efficiency()
        {
            var updateManager = testGameObject.AddComponent<UpdateManager>();

            // Create many updatable objects
            int updatableCount = 100;
            var updatables = new TestUpdatable[updatableCount];

            for (int i = 0; i < updatableCount; i++)
            {
                updatables[i] = new TestUpdatable();
                updateManager.RegisterUpdatable(updatables[i]);
            }

            // Measure update loop performance
            float testDuration = 2f;
            float elapsed = 0f;

            while (elapsed < testDuration)
            {
                elapsed += Time.deltaTime;
                yield return null;
            }

            // Check that all objects were updated
            foreach (var updatable in updatables)
            {
                Assert.Greater(updatable.UpdateCount, 0,
                    "All updatables should have been updated");
            }

            // Calculate update efficiency
            float expectedUpdates = testDuration / Time.fixedDeltaTime;
            float averageUpdates = updatables[0].UpdateCount;
            float efficiency = averageUpdates / expectedUpdates;

            Debug.Log($"UpdateManager efficiency: {efficiency:P0} ({averageUpdates} updates in {testDuration}s)");

            Assert.Greater(efficiency, 0.9f,
                "UpdateManager should maintain at least 90% efficiency");
        }

        // Helper class for testing
        private class TestUpdatable : IUpdatable
        {
            public int UpdateCount { get; private set; }
            public bool IsActive => true;

            public void OnUpdate(float deltaTime)
            {
                UpdateCount++;
            }
        }
    }
}
