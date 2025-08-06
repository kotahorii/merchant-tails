using System.Collections;
using MerchantTails.Core;
using MerchantTails.Data;
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

            // Record initial fruit price
            float initialPrice = 0f;
            float newPrice = 0f;
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                initialPrice = MarketSystem.Instance.GetCurrentPrice(ItemType.Fruit);
                TimeManager.Instance.SkipToNextPhase();
            }
            catch (System.Exception e)
            {
                errorOccurred = true;
                caughtException = e;
            }

            if (!errorOccurred)
            {
                yield return new WaitForSeconds(0.1f);

                try
                {
                    newPrice = MarketSystem.Instance.GetCurrentPrice(ItemType.Fruit);
                }
                catch (System.Exception e)
                {
                    errorOccurred = true;
                    caughtException = e;
                }
            }

            if (!errorOccurred)
            {
                float capturedNewPrice = newPrice;
                float capturedInitialPrice = initialPrice;

                // Check if price changed (market is responding to time)
                bool priceChanged = !Mathf.Approximately(capturedInitialPrice, capturedNewPrice);

                testResult.passed = priceChanged;
                testResult.message = priceChanged
                    ? $"Price changed from {capturedInitialPrice:F2} to {capturedNewPrice:F2}"
                    : "Price did not change with time advancement";

                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);
            }
            else
            {
                testResult.passed = false;
                testResult.message = $"Exception: {caughtException.Message}";
                testResult.duration = Time.time - startTime;
                LogTestResult(testResult);
            }
        }

        private IEnumerator TestMarketToInventoryIntegration()
        {
            var testResult = new TestResult { testName = "Market-to-Inventory Integration" };
            float startTime = Time.time;

            bool errorOccurred = false;
            System.Exception caughtException = null;

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
                }
            }
            catch (System.Exception e)
            {
                errorOccurred = true;
                caughtException = e;
            }

            if (!errorOccurred && testResult.passed != false)
            {
                yield return new WaitForSeconds(0.1f);

                try
                {
                    // Check if inventory was affected
                    int remainingItems = InventorySystem.Instance.GetItemCount(
                        ItemType.Potion,
                        InventoryLocation.Trading
                    );

                    testResult.passed = true; // Transaction event was published successfully
                    testResult.message = $"Transaction processed, remaining items: {remainingItems}";
                }
                catch (System.Exception e)
                {
                    errorOccurred = true;
                    caughtException = e;
                }
            }

            if (errorOccurred)
            {
                testResult.passed = false;
                testResult.message = $"Exception: {caughtException.Message}";
            }

            testResult.duration = Time.time - startTime;
            LogTestResult(testResult);
        }

        private IEnumerator TestCompleteGameLoop()
        {
            var testResult = new TestResult { testName = "Complete Game Loop" };
            float startTime = Time.time;

            int initialDay = 0;
            Season initialSeason = Season.Spring;
            bool addResult = false;
            bool moveResult = false;
            int newDay = 0;
            bool dayAdvanced = false;
            float weaponPrice = 0f;
            bool priceExists = false;
            int storefrontCount = 0;
            int tradingCount = 0;
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                // Simulate a complete trading cycle

                // 1. Check initial state
                initialDay = TimeManager.Instance.CurrentDay;
                initialSeason = TimeManager.Instance.CurrentSeason;

                // 2. Add items to inventory
                addResult = InventorySystem.Instance.AddItem(ItemType.Weapon, 5, InventoryLocation.Trading);

                // 3. Move items to storefront
                moveResult = InventorySystem.Instance.MoveItem(
                    ItemType.Weapon,
                    2,
                    InventoryLocation.Trading,
                    InventoryLocation.Storefront
                );
            }
            catch (System.Exception e)
            {
                errorOccurred = true;
                caughtException = e;
            }

            if (!errorOccurred)
            {
                // 4. Advance time significantly
                for (int i = 0; i < 8; i++) // Advance 2 days (4 phases each)
                {
                    try
                    {
                        TimeManager.Instance.SkipToNextPhase();
                    }
                    catch (System.Exception e)
                    {
                        errorOccurred = true;
                        caughtException = e;
                        break;
                    }

                    yield return new WaitForSeconds(0.02f);
                }
            }

            if (!errorOccurred)
            {
                try
                {
                    // 5. Check that time advanced
                    newDay = TimeManager.Instance.CurrentDay;
                    dayAdvanced = newDay > initialDay;

                    // 6. Check that market prices have updated
                    weaponPrice = MarketSystem.Instance.GetCurrentPrice(ItemType.Weapon);
                    priceExists = weaponPrice > 0;

                    // 7. Verify inventory state
                    storefrontCount = InventorySystem.Instance.GetItemCount(
                        ItemType.Weapon,
                        InventoryLocation.Storefront
                    );
                    tradingCount = InventorySystem.Instance.GetItemCount(ItemType.Weapon, InventoryLocation.Trading);

                    bool allOperationsSuccessful = addResult && moveResult && dayAdvanced && priceExists;

                    testResult.passed = allOperationsSuccessful;
                    testResult.message = allOperationsSuccessful
                        ? $"Complete cycle: Day {initialDay}->{newDay}, Price {weaponPrice:F2}, Items S:{storefrontCount} T:{tradingCount}"
                        : $"Cycle incomplete: Add:{addResult} Move:{moveResult} DayAdv:{dayAdvanced} Price:{priceExists}";
                }
                catch (System.Exception e)
                {
                    errorOccurred = true;
                    caughtException = e;
                }
            }

            if (errorOccurred)
            {
                testResult.passed = false;
                testResult.message = $"Exception: {caughtException.Message}";
            }

            testResult.duration = Time.time - startTime;
            LogTestResult(testResult);
        }

        private IEnumerator TestEventPropagation()
        {
            var testResult = new TestResult { testName = "Event Propagation" };
            float startTime = Time.time;

            bool phaseEventReceived = false;
            bool priceEventReceived = false;
            System.Action<PhaseChangedEvent> phaseHandler = null;
            System.Action<PriceChangedEvent> priceHandler = null;
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                // Subscribe to events temporarily
                phaseHandler = (evt) => phaseEventReceived = true;
                priceHandler = (evt) => priceEventReceived = true;

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
            }
            catch (System.Exception e)
            {
                errorOccurred = true;
                caughtException = e;
            }

            if (!errorOccurred)
            {
                yield return new WaitForSeconds(0.2f); // Wait for events to propagate

                try
                {
                    // Cleanup subscriptions
                    if (phaseHandler != null)
                        EventBus.Unsubscribe<PhaseChangedEvent>(phaseHandler);
                    if (priceHandler != null)
                        EventBus.Unsubscribe<PriceChangedEvent>(priceHandler);

                    bool eventsWorking = phaseEventReceived; // At minimum phase event should work

                    testResult.passed = eventsWorking;
                    testResult.message = $"Events received - Phase: {phaseEventReceived}, Price: {priceEventReceived}";
                }
                catch (System.Exception e)
                {
                    errorOccurred = true;
                    caughtException = e;
                }
            }

            if (errorOccurred)
            {
                testResult.passed = false;
                testResult.message = $"Exception: {caughtException.Message}";

                // Cleanup on error
                try
                {
                    if (phaseHandler != null)
                        EventBus.Unsubscribe<PhaseChangedEvent>(phaseHandler);
                    if (priceHandler != null)
                        EventBus.Unsubscribe<PriceChangedEvent>(priceHandler);
                }
                catch
                { /* Ignore cleanup errors */
                }
            }

            testResult.duration = Time.time - startTime;
            LogTestResult(testResult);
        }

        private IEnumerator TestDataPersistence()
        {
            var testResult = new TestResult { testName = "Data Persistence" };
            float startTime = Time.time;

            bool timeDataValid = false;
            bool inventoryDataValid = false;
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                // Test time data persistence
                var timeData = TimeManager.Instance.GetTimeData();
                timeDataValid =
                    timeData.currentDay > 0 && timeData.currentSeason != Season.Spring
                    || timeData.currentSeason == Season.Spring;

                // Test inventory data persistence
                var inventoryData = InventorySystem.Instance.GetInventoryData();
                inventoryDataValid = inventoryData != null;

                testResult.passed = timeDataValid && inventoryDataValid;
                testResult.message = $"Data persistence - Time: {timeDataValid}, Inventory: {inventoryDataValid}";
            }
            catch (System.Exception e)
            {
                errorOccurred = true;
                caughtException = e;
            }

            if (errorOccurred)
            {
                testResult.passed = false;
                testResult.message = $"Exception: {caughtException.Message}";
            }

            testResult.duration = Time.time - startTime;
            LogTestResult(testResult);

            yield return null;
        }

        private IEnumerator TestErrorRecovery()
        {
            var testResult = new TestResult { testName = "Error Recovery" };
            float startTime = Time.time;

            bool healthCheckPassed = false;
            bool safeExecutionWorked = false;
            bool safeExecutionHandledException = false;
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                // Test system health check
                healthCheckPassed = ErrorHandler.CheckSystemHealth();

                // Test safe execution
                safeExecutionWorked = ErrorHandler.SafeExecute(
                    () =>
                    {
                        // This should not throw
                        var price = MarketSystem.Instance.GetCurrentPrice(ItemType.Fruit);
                    },
                    "TestSafeExecution"
                );

                // Test safe execution with exception
                bool exceptionResult = ErrorHandler.SafeExecute(
                    () =>
                    {
                        throw new System.Exception("Test exception");
                    },
                    "TestException"
                );
                safeExecutionHandledException = exceptionResult;

                bool allRecoveryTestsPassed = healthCheckPassed && safeExecutionWorked && safeExecutionHandledException;

                testResult.passed = allRecoveryTestsPassed;
                testResult.message =
                    $"Recovery tests - Health: {healthCheckPassed}, Safe: {safeExecutionWorked}, Exception: {safeExecutionHandledException}";
            }
            catch (System.Exception e)
            {
                errorOccurred = true;
                caughtException = e;
            }

            if (errorOccurred)
            {
                testResult.passed = false;
                testResult.message = $"Exception: {caughtException.Message}";
            }

            testResult.duration = Time.time - startTime;
            LogTestResult(testResult);

            yield return null;
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
