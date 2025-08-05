using System.Collections;
using System.Collections.Generic;
using MerchantTails.Core;
using MerchantTails.Data;
using MerchantTails.Inventory;
using MerchantTails.Market;
using UnityEngine;

namespace MerchantTails.Testing
{
    /// <summary>
    /// 安定性重視のテスト実行クラス
    /// 長時間実行とストレステストで堅牢性を確認
    /// </summary>
    public class StabilityTest : MonoBehaviour
    {
        [Header("Stability Test Settings")]
        [SerializeField]
        private bool runStabilityTests = false;

        [SerializeField]
        private int stressTestCycles = 100;

        [SerializeField]
        private float memoryCheckInterval = 10f;

        private bool stabilityTestRunning = false;
        private List<string> errorLog = new List<string>();
        private int totalOperations = 0;
        private int failedOperations = 0;

        public struct StabilityReport
        {
            public int totalOperations;
            public int failedOperations;
            public float successRate;
            public long averageMemoryUsage;
            public List<string> criticalErrors;
        }

        private void Start()
        {
            if (runStabilityTests)
            {
                StartCoroutine(RunStabilityTests());
            }
        }

        public void StartStabilityTest()
        {
            if (!stabilityTestRunning)
            {
                StartCoroutine(RunStabilityTests());
            }
        }

        private IEnumerator RunStabilityTests()
        {
            stabilityTestRunning = true;
            errorLog.Clear();
            totalOperations = 0;
            failedOperations = 0;

            ErrorHandler.LogInfo("Starting comprehensive stability tests", "StabilityTest");

            // Wait for all systems to be ready
            yield return StartCoroutine(WaitForSystemInitialization());

            // Run stress tests
            yield return StartCoroutine(RunStressTests());

            // Run memory leak detection
            yield return StartCoroutine(RunMemoryLeakTest());

            // Run event system stress test
            yield return StartCoroutine(RunEventStressTest());

            // Run long-duration test
            yield return StartCoroutine(RunLongDurationTest());

            // Generate stability report
            GenerateStabilityReport();

            stabilityTestRunning = false;
            ErrorHandler.LogInfo("Stability tests completed", "StabilityTest");
        }

        private IEnumerator WaitForSystemInitialization()
        {
            float timeout = 30f;
            float elapsed = 0f;

            while (elapsed < timeout)
            {
                if (
                    GameManager.Instance != null
                    && TimeManager.Instance != null
                    && MarketSystem.Instance != null
                    && InventorySystem.Instance != null
                )
                {
                    ErrorHandler.LogInfo("All systems initialized for stability testing", "StabilityTest");
                    yield break;
                }

                yield return new WaitForSeconds(0.1f);
                elapsed += 0.1f;
            }

            ErrorHandler.LogError("System initialization timeout", null, "StabilityTest");
        }

        private IEnumerator RunStressTests()
        {
            ErrorHandler.LogInfo($"Running stress tests: {stressTestCycles} cycles", "StabilityTest");

            for (int cycle = 0; cycle < stressTestCycles; cycle++)
            {
                // Market system stress test
                yield return StartCoroutine(StressTestMarketSystem());

                // Inventory system stress test
                yield return StartCoroutine(StressTestInventorySystem());

                // Time system stress test
                yield return StartCoroutine(StressTestTimeSystem());

                // Check system health every 10 cycles
                if (cycle % 10 == 0)
                {
                    bool healthCheck = ErrorHandler.CheckSystemHealth();
                    if (!healthCheck)
                    {
                        errorLog.Add($"Health check failed at cycle {cycle}");
                    }

                    // Force garbage collection
                    System.GC.Collect();
                    yield return new WaitForSeconds(0.1f);
                }

                yield return null; // Yield frame
            }

            ErrorHandler.LogInfo("Stress tests completed", "StabilityTest");
        }

        private IEnumerator StressTestMarketSystem()
        {
            totalOperations++;

            bool success = ErrorHandler.SafeExecute(
                () =>
                {
                    // Test rapid price queries
                    for (int i = 0; i < 100; i++)
                    {
                        foreach (ItemType itemType in System.Enum.GetValues(typeof(ItemType)))
                        {
                            var price = MarketSystem.Instance.GetCurrentPrice(itemType);
                            var marketData = MarketSystem.Instance.GetMarketData(itemType);
                        }
                    }

                    // Test event generation
                    var testEvent = new GameEventTriggeredEvent(
                        "Stress Test Event",
                        "Testing market stability",
                        new ItemType[] { ItemType.Fruit },
                        new float[] { 1.1f },
                        1
                    );
                    EventBus.Publish(testEvent);
                },
                "StressTestMarket"
            );

            if (!success)
            {
                failedOperations++;
                errorLog.Add("Market system stress test failed");
            }

            yield return null;
        }

        private IEnumerator StressTestInventorySystem()
        {
            totalOperations++;

            bool success = ErrorHandler.SafeExecute(
                () =>
                {
                    // Test rapid inventory operations
                    foreach (ItemType itemType in System.Enum.GetValues(typeof(ItemType)))
                    {
                        // Add items
                        InventorySystem.Instance.AddItem(itemType, 10, InventoryLocation.Trading);

                        // Move items
                        InventorySystem.Instance.MoveItem(
                            itemType,
                            5,
                            InventoryLocation.Trading,
                            InventoryLocation.Storefront
                        );

                        // Check counts
                        var tradingCount = InventorySystem.Instance.GetItemCount(itemType, InventoryLocation.Trading);
                        var storefrontCount = InventorySystem.Instance.GetItemCount(
                            itemType,
                            InventoryLocation.Storefront
                        );

                        // Remove items
                        InventorySystem.Instance.RemoveItem(itemType, 3, InventoryLocation.Trading);
                        InventorySystem.Instance.RemoveItem(itemType, 2, InventoryLocation.Storefront);
                    }

                    // Test data persistence
                    var inventoryData = InventorySystem.Instance.GetInventoryData();
                },
                "StressTestInventory"
            );

            if (!success)
            {
                failedOperations++;
                errorLog.Add("Inventory system stress test failed");
            }

            yield return null;
        }

        private IEnumerator StressTestTimeSystem()
        {
            totalOperations++;

            bool success = ErrorHandler.SafeExecute(
                () =>
                {
                    // Test rapid time queries
                    for (int i = 0; i < 50; i++)
                    {
                        var currentTime = TimeManager.Instance.GetFormattedTime();
                        var phaseProgress = TimeManager.Instance.GetPhaseProgress();
                        var timeData = TimeManager.Instance.GetTimeData();
                    }

                    // Test time advancement
                    TimeManager.Instance.SkipToNextPhase();
                },
                "StressTestTime"
            );

            if (!success)
            {
                failedOperations++;
                errorLog.Add("Time system stress test failed");
            }

            yield return null;
        }

        private IEnumerator RunMemoryLeakTest()
        {
            ErrorHandler.LogInfo("Running memory leak detection test", "StabilityTest");

            long initialMemory = System.GC.GetTotalMemory(true);

            // Run intensive operations
            for (int i = 0; i < 1000; i++)
            {
                // Create and destroy temporary objects
                var tempData = new PlayerData();
                var tempMarketData = MarketSystem.Instance.GetMarketData(ItemType.Fruit);

                // Trigger events
                var tempEvent = new PriceChangedEvent(ItemType.Potion, 100f, 110f);
                EventBus.Publish(tempEvent);

                if (i % 100 == 0)
                {
                    yield return null;
                }
            }

            // Force garbage collection
            System.GC.Collect();
            System.GC.WaitForPendingFinalizers();
            System.GC.Collect();

            yield return new WaitForSeconds(2f);

            long finalMemory = System.GC.GetTotalMemory(true);
            long memoryDifference = finalMemory - initialMemory;

            ErrorHandler.LogInfo(
                $"Memory test: Initial={initialMemory / 1024}KB, Final={finalMemory / 1024}KB, Diff={memoryDifference / 1024}KB",
                "StabilityTest"
            );

            if (memoryDifference > 10 * 1024 * 1024) // 10MB threshold
            {
                errorLog.Add($"Potential memory leak detected: {memoryDifference / 1024 / 1024}MB increase");
            }
        }

        private IEnumerator RunEventStressTest()
        {
            ErrorHandler.LogInfo("Running event system stress test", "StabilityTest");

            totalOperations++;

            int eventsSent = 0;
            int eventsReceived = 0;

            // Create test event handler
            System.Action<PriceChangedEvent> testHandler = (evt) => eventsReceived++;
            EventBus.Subscribe<PriceChangedEvent>(testHandler);

            bool success = ErrorHandler.SafeExecute(
                () =>
                {
                    // Send many events rapidly
                    for (int i = 0; i < 500; i++)
                    {
                        var priceEvent = new PriceChangedEvent(ItemType.Fruit, 100f, 100f + i);
                        EventBus.Publish(priceEvent);
                        eventsSent++;
                    }
                },
                "EventStressTest"
            );

            yield return new WaitForSeconds(1f); // Wait for event processing

            EventBus.Unsubscribe<PriceChangedEvent>(testHandler);

            if (!success || eventsReceived != eventsSent)
            {
                failedOperations++;
                errorLog.Add($"Event stress test failed: Sent {eventsSent}, Received {eventsReceived}");
            }

            ErrorHandler.LogInfo($"Event stress test: {eventsSent} sent, {eventsReceived} received", "StabilityTest");
        }

        private IEnumerator RunLongDurationTest()
        {
            ErrorHandler.LogInfo("Running long duration test (30 seconds)", "StabilityTest");

            float testDuration = 30f;
            float elapsed = 0f;
            int operationCount = 0;

            while (elapsed < testDuration)
            {
                // Continuous operations
                totalOperations++;
                operationCount++;

                bool success = ErrorHandler.SafeExecute(
                    () =>
                    {
                        // Simulate normal game operations
                        var currentPrice = MarketSystem.Instance.GetCurrentPrice(ItemType.Weapon);
                        InventorySystem.Instance.AddItem(ItemType.Weapon, 1, InventoryLocation.Trading);
                        var timeData = TimeManager.Instance.GetTimeData();

                        // Occasionally advance time
                        if (operationCount % 50 == 0)
                        {
                            TimeManager.Instance.SkipToNextPhase();
                        }
                    },
                    "LongDurationTest"
                );

                if (!success)
                {
                    failedOperations++;
                }

                // Health check every 5 seconds
                if (elapsed % 5f < Time.deltaTime)
                {
                    bool health = ErrorHandler.CheckSystemHealth();
                    if (!health)
                    {
                        errorLog.Add($"Health check failed during long duration test at {elapsed:F1}s");
                    }
                }

                yield return new WaitForSeconds(0.1f);
                elapsed += 0.1f;
            }

            ErrorHandler.LogInfo($"Long duration test completed: {operationCount} operations", "StabilityTest");
        }

        private void GenerateStabilityReport()
        {
            float successRate =
                totalOperations > 0 ? (float)(totalOperations - failedOperations) / totalOperations * 100f : 100f;

            long currentMemory = System.GC.GetTotalMemory(false);

            var report = new StabilityReport
            {
                totalOperations = totalOperations,
                failedOperations = failedOperations,
                successRate = successRate,
                averageMemoryUsage = currentMemory,
                criticalErrors = new List<string>(errorLog),
            };

            ErrorHandler.LogInfo($"=== STABILITY TEST REPORT ===", "StabilityTest");
            ErrorHandler.LogInfo($"Total Operations: {report.totalOperations}", "StabilityTest");
            ErrorHandler.LogInfo($"Failed Operations: {report.failedOperations}", "StabilityTest");
            ErrorHandler.LogInfo($"Success Rate: {report.successRate:F2}%", "StabilityTest");
            ErrorHandler.LogInfo($"Memory Usage: {report.averageMemoryUsage / 1024 / 1024}MB", "StabilityTest");
            ErrorHandler.LogInfo($"Critical Errors: {report.criticalErrors.Count}", "StabilityTest");

            foreach (var error in report.criticalErrors)
            {
                ErrorHandler.LogError($"Critical Error: {error}", null, "StabilityTest");
            }

            if (report.successRate >= 95f && report.criticalErrors.Count == 0)
            {
                ErrorHandler.LogInfo("✓ STABILITY TEST PASSED - System is stable", "StabilityTest");
            }
            else
            {
                ErrorHandler.LogError("✗ STABILITY TEST FAILED - System has stability issues", null, "StabilityTest");
            }
        }

        public bool IsStabilityTestRunning()
        {
            return stabilityTestRunning;
        }

        public StabilityReport GetLastReport()
        {
            return new StabilityReport
            {
                totalOperations = totalOperations,
                failedOperations = failedOperations,
                successRate =
                    totalOperations > 0 ? (float)(totalOperations - failedOperations) / totalOperations * 100f : 100f,
                averageMemoryUsage = System.GC.GetTotalMemory(false),
                criticalErrors = new List<string>(errorLog),
            };
        }
    }
}
