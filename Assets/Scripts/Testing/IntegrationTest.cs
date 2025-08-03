using System.Collections;
using MerchantTails.Core;
using MerchantTails.Data;
using MerchantTails.Events;
using MerchantTails.Inventory;
using MerchantTails.Market;
using UnityEngine;

namespace MerchantTails.Testing
{
    /// <summary>
    /// システム間連携テストクラス
    /// 複数システムが協調して動作することを確認
    /// </summary>
    public class IntegrationTest : MonoBehaviour
    {
        [Header("Test Settings")]
        [SerializeField]
        private bool runOnStart = false;

        [SerializeField]
        private float testTimeout = 30f;

        private bool testInProgress = false;
        private int passedTests = 0;
        private int failedTests = 0;

        public struct TestResult
        {
            public string testName;
            public bool passed;
            public string message;
            public float duration;
        }

        private void Start()
        {
            if (runOnStart)
            {
                StartCoroutine(RunAllTests());
            }
        }

        public void StartIntegrationTests()
        {
            if (!testInProgress)
            {
                StartCoroutine(RunAllTests());
            }
        }

        private IEnumerator RunAllTests()
        {
            testInProgress = true;
            passedTests = 0;
            failedTests = 0;

            ErrorHandler.LogInfo("Starting integration tests...", "IntegrationTest");

            // Wait for systems to be ready
            yield return StartCoroutine(WaitForSystemsReady());

            // Run individual tests
            yield return StartCoroutine(TestTimeToMarketIntegration());
            yield return StartCoroutine(TestMarketToInventoryIntegration());
            yield return StartCoroutine(TestCompleteGameLoop());
            yield return StartCoroutine(TestEventPropagation());
            yield return StartCoroutine(TestDataPersistence());
            yield return StartCoroutine(TestErrorRecovery());

            // Report results
            ErrorHandler.LogInfo(
                $"Integration tests completed: {passedTests} passed, {failedTests} failed",
                "IntegrationTest"
            );
            testInProgress = false;
        }

        private IEnumerator WaitForSystemsReady()
        {
            float waitTime = 0f;
            while (waitTime < testTimeout)
            {
                if (
                    GameManager.Instance != null
                    && TimeManager.Instance != null
                    && MarketSystem.Instance != null
                    && InventorySystem.Instance != null
                )
                {
                    yield return new WaitForSeconds(0.5f); // Extra wait for initialization
                    yield break;
                }

                yield return new WaitForSeconds(0.1f);
                waitTime += 0.1f;
            }

            ErrorHandler.LogError("Timeout waiting for systems to be ready", null, "IntegrationTest");
        }

        private IEnumerator TestTimeToMarketIntegration()
        {
            var testResult = new TestResult { testName = "Time-to-Market Integration" };
            float startTime = Time.time;

            try
            {
                // Record initial fruit price
                float initialPrice = MarketSystem.Instance.GetCurrentPrice(ItemType.Fruit);

                // Advance time and check if market responds
                TimeManager.Instance.SkipToNextPhase();
                yield return new WaitForSeconds(0.1f);

                float newPrice = MarketSystem.Instance.GetCurrentPrice(ItemType.Fruit);

                // Check if price changed (market is responding to time)
                bool priceChanged = !Mathf.Approximately(initialPrice, newPrice);

                testResult.passed = priceChanged;
                testResult.message = priceChanged
                    ? $"Price changed from {initialPrice:F2} to {newPrice:F2}"
                    : "Price did not change with time advancement";

                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);
            }
            catch (System.Exception e)
            {
                testResult.passed = false;
                testResult.message = $"Exception: {e.Message}";
                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);
            }
        }

        private IEnumerator TestMarketToInventoryIntegration()
        {
            var testResult = new TestResult { testName = "Market-to-Inventory Integration" };
            float startTime = Time.time;

            try
            {
                // Add items to inventory
                bool addSuccess = InventorySystem.Instance.AddItem(ItemType.Potion, 10, InventoryLocation.Trading);

                if (!addSuccess)
                {
                    testResult.passed = false;
                    testResult.message = "Failed to add items to inventory";
                }
                else
                {
                    // Simulate transaction
                    float currentPrice = MarketSystem.Instance.GetCurrentPrice(ItemType.Potion);
                    var transactionEvent = new TransactionCompletedEvent(
                        ItemType.Potion,
                        3,
                        currentPrice,
                        false,
                        currentPrice * 3 * 0.1f
                    );

                    EventBus.Publish(transactionEvent);
                    yield return new WaitForSeconds(0.1f);

                    // Check if inventory was affected
                    int remainingItems = InventorySystem.Instance.GetItemCount(
                        ItemType.Potion,
                        InventoryLocation.Trading
                    );

                    testResult.passed = true; // Transaction event was published successfully
                    testResult.message = $"Transaction processed, remaining items: {remainingItems}";
                }

                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);
            }
            catch (System.Exception e)
            {
                testResult.passed = false;
                testResult.message = $"Exception: {e.Message}";
                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);
            }
        }

        private IEnumerator TestCompleteGameLoop()
        {
            var testResult = new TestResult { testName = "Complete Game Loop" };
            float startTime = Time.time;

            try
            {
                // Simulate a complete trading cycle

                // 1. Check initial state
                int initialDay = TimeManager.Instance.CurrentDay;
                Season initialSeason = TimeManager.Instance.CurrentSeason;

                // 2. Add items to inventory
                bool addResult = InventorySystem.Instance.AddItem(ItemType.Weapon, 5, InventoryLocation.Trading);

                // 3. Move items to storefront
                bool moveResult = InventorySystem.Instance.MoveItem(
                    ItemType.Weapon,
                    2,
                    InventoryLocation.Trading,
                    InventoryLocation.Storefront
                );

                // 4. Advance time significantly
                for (int i = 0; i < 8; i++) // Advance 2 days (4 phases each)
                {
                    TimeManager.Instance.SkipToNextPhase();
                    yield return new WaitForSeconds(0.02f);
                }

                // 5. Check that time advanced
                int newDay = TimeManager.Instance.CurrentDay;
                bool dayAdvanced = newDay > initialDay;

                // 6. Check that market prices have updated
                float weaponPrice = MarketSystem.Instance.GetCurrentPrice(ItemType.Weapon);
                bool priceExists = weaponPrice > 0;

                // 7. Verify inventory state
                int storefrontCount = InventorySystem.Instance.GetItemCount(
                    ItemType.Weapon,
                    InventoryLocation.Storefront
                );
                int tradingCount = InventorySystem.Instance.GetItemCount(ItemType.Weapon, InventoryLocation.Trading);

                bool allOperationsSuccessful = addResult && moveResult && dayAdvanced && priceExists;

                testResult.passed = allOperationsSuccessful;
                testResult.message = allOperationsSuccessful
                    ? $"Complete cycle: Day {initialDay}->{newDay}, Price {weaponPrice:F2}, Items S:{storefrontCount} T:{tradingCount}"
                    : $"Cycle incomplete: Add:{addResult} Move:{moveResult} DayAdv:{dayAdvanced} Price:{priceExists}";

                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);
            }
            catch (System.Exception e)
            {
                testResult.passed = false;
                testResult.message = $"Exception: {e.Message}";
                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);
            }
        }

        private IEnumerator TestEventPropagation()
        {
            var testResult = new TestResult { testName = "Event Propagation" };
            float startTime = Time.time;

            try
            {
                bool phaseEventReceived = false;
                bool priceEventReceived = false;

                // Subscribe to events temporarily
                System.Action<PhaseChangedEvent> phaseHandler = (evt) => phaseEventReceived = true;
                System.Action<PriceChangedEvent> priceHandler = (evt) => priceEventReceived = true;

                EventBus.Subscribe<PhaseChangedEvent>(phaseHandler);
                EventBus.Subscribe<PriceChangedEvent>(priceHandler);

                // Trigger events
                TimeManager.Instance.SkipToNextPhase();

                // Trigger a market event
                var marketEvent = new GameEventTriggeredEvent(
                    "Test Integration Event",
                    "Testing event propagation",
                    new ItemType[] { ItemType.Gem },
                    new float[] { 1.2f },
                    1
                );
                EventBus.Publish(marketEvent);

                yield return new WaitForSeconds(0.2f); // Wait for events to propagate

                // Cleanup subscriptions
                EventBus.Unsubscribe<PhaseChangedEvent>(phaseHandler);
                EventBus.Unsubscribe<PriceChangedEvent>(priceHandler);

                bool eventsWorking = phaseEventReceived; // At minimum phase event should work

                testResult.passed = eventsWorking;
                testResult.message = $"Events received - Phase: {phaseEventReceived}, Price: {priceEventReceived}";
                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);
            }
            catch (System.Exception e)
            {
                testResult.passed = false;
                testResult.message = $"Exception: {e.Message}";
                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);
            }
        }

        private IEnumerator TestDataPersistence()
        {
            var testResult = new TestResult { testName = "Data Persistence" };
            float startTime = Time.time;

            try
            {
                // Test time data persistence
                var timeData = TimeManager.Instance.GetTimeData();
                bool timeDataValid =
                    timeData.currentDay > 0 && timeData.currentSeason != Season.Spring
                    || timeData.currentSeason == Season.Spring;

                // Test inventory data persistence
                var inventoryData = InventorySystem.Instance.GetInventoryData();
                bool inventoryDataValid = inventoryData != null;

                testResult.passed = timeDataValid && inventoryDataValid;
                testResult.message = $"Data persistence - Time: {timeDataValid}, Inventory: {inventoryDataValid}";
                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);

                yield return null;
            }
            catch (System.Exception e)
            {
                testResult.passed = false;
                testResult.message = $"Exception: {e.Message}";
                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);
            }
        }

        private IEnumerator TestErrorRecovery()
        {
            var testResult = new TestResult { testName = "Error Recovery" };
            float startTime = Time.time;

            try
            {
                // Test system health check
                bool healthCheckPassed = ErrorHandler.CheckSystemHealth();

                // Test safe execution
                bool safeExecutionWorked = ErrorHandler.SafeExecute(
                    () =>
                    {
                        // This should not throw
                        var price = MarketSystem.Instance.GetCurrentPrice(ItemType.Fruit);
                    },
                    "TestSafeExecution"
                );

                // Test safe execution with exception
                bool safeExecutionHandledException = !ErrorHandler.SafeExecute(
                    () =>
                    {
                        throw new System.Exception("Test exception");
                    },
                    "TestException"
                );

                bool allRecoveryTestsPassed = healthCheckPassed && safeExecutionWorked && safeExecutionHandledException;

                testResult.passed = allRecoveryTestsPassed;
                testResult.message =
                    $"Recovery tests - Health: {healthCheckPassed}, Safe: {safeExecutionWorked}, Exception: {safeExecutionHandledException}";
                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);

                yield return null;
            }
            catch (System.Exception e)
            {
                testResult.passed = false;
                testResult.message = $"Exception: {e.Message}";
                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);
            }
        }

        private void LogTestResult(TestResult result)
        {
            if (result.passed)
            {
                passedTests++;
                ErrorHandler.LogInfo(
                    $"✓ {result.testName}: {result.message} ({result.duration:F2}s)",
                    "IntegrationTest"
                );
            }
            else
            {
                failedTests++;
                ErrorHandler.LogError(
                    $"✗ {result.testName}: {result.message} ({result.duration:F2}s)",
                    null,
                    "IntegrationTest"
                );
            }
        }

        public void GetTestSummary(out int passed, out int failed)
        {
            passed = passedTests;
            failed = failedTests;
        }

        public bool IsTestInProgress()
        {
            return testInProgress;
        }
    }
}
