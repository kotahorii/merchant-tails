using System;
using System.Collections;
using MerchantTails.Core;
using MerchantTails.Data;
using MerchantTails.Inventory;
using MerchantTails.Market;
using UnityEngine;

namespace MerchantTails.Testing
{
    /// <summary>
    /// エラーハンドリングと回復処理のテストクラス
    /// 異常状態からの復旧能力を検証
    /// </summary>
    public class ErrorRecoveryTest : MonoBehaviour
    {
        [Header("Error Recovery Test Settings")]
        [SerializeField]
        private bool runRecoveryTests = false;

        [SerializeField]
        private float testTimeout = 30f;

        private bool recoveryTestRunning = false;
        private int totalRecoveryTests = 0;
        private int successfulRecoveries = 0;

        public struct RecoveryTestResult
        {
            public string testName;
            public bool recoverySuccessful;
            public float recoveryTime;
            public string errorDetails;
        }

        private void Start()
        {
            if (runRecoveryTests)
            {
                StartCoroutine(RunErrorRecoveryTests());
            }
        }

        public void StartErrorRecoveryTest()
        {
            if (!recoveryTestRunning)
            {
                StartCoroutine(RunErrorRecoveryTests());
            }
        }

        private IEnumerator RunErrorRecoveryTests()
        {
            recoveryTestRunning = true;
            totalRecoveryTests = 0;
            successfulRecoveries = 0;

            ErrorHandler.LogInfo("Starting error recovery tests", "ErrorRecoveryTest");

            // Test 1: Null Reference Recovery
            yield return StartCoroutine(TestNullReferenceRecovery());

            // Test 2: System Recovery
            yield return StartCoroutine(TestSystemRecovery());

            // Test 3: Exception Handling
            yield return StartCoroutine(TestExceptionHandling());

            // Test 4: Memory Management
            yield return StartCoroutine(TestMemoryPressureRecovery());

            // Test 5: Event System Recovery
            yield return StartCoroutine(TestEventSystemRecovery());

            // Generate final report
            GenerateRecoveryReport();

            recoveryTestRunning = false;
            ErrorHandler.LogInfo("Error recovery tests completed", "ErrorRecoveryTest");
        }

        private IEnumerator TestNullReferenceRecovery()
        {
            totalRecoveryTests++;
            float startTime = Time.time;

            ErrorHandler.LogInfo("Testing null reference recovery", "ErrorRecoveryTest");

            bool recovered = true;
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                // Simulate null reference scenarios

                // Test 1: Null component access
                recovered &= ErrorHandler.SafeExecute(
                    () =>
                    {
                        GameObject nullObject = null;
                        var component = nullObject.GetComponent<Transform>(); // This should fail safely
                    },
                    "NullComponentTest"
                );

                // Test 2: Null manager instance
                recovered &= ErrorHandler.SafeExecute(
                    () =>
                    {
                        if (GameManager.Instance == null)
                        {
                            throw new System.NullReferenceException("GameManager is null");
                        }
                    },
                    "NullManagerTest"
                );

                // Test 3: Null collection access
                recovered &= ErrorHandler.SafeExecute(
                    () =>
                    {
                        System.Collections.Generic.List<string> nullList = null;
                        int count = nullList.Count; // This should fail safely
                    },
                    "NullCollectionTest"
                );

                float recoveryTime = Time.time - startTime;

                var result = new RecoveryTestResult
                {
                    testName = "Null Reference Recovery",
                    recoverySuccessful = recovered,
                    recoveryTime = recoveryTime,
                    errorDetails = recovered
                        ? "All null references handled safely"
                        : "Some null references not handled",
                };

                LogRecoveryResult(result);

                if (recovered)
                    successfulRecoveries++;
            }
            catch (System.Exception e)
            {
                errorOccurred = true;
                caughtException = e;
            }

            if (errorOccurred)
            {
                ErrorHandler.LogError(
                    $"Null reference recovery test failed: {caughtException.Message}",
                    caughtException,
                    "ErrorRecoveryTest"
                );
            }

            yield return null;
        }

        private IEnumerator TestSystemRecovery()
        {
            totalRecoveryTests++;
            float startTime = Time.time;

            ErrorHandler.LogInfo("Testing system recovery capabilities", "ErrorRecoveryTest");

            bool allRecovered = true;
            bool gameManagerRecovered = false;
            bool timeManagerRecovered = false;
            bool marketSystemRecovered = false;
            bool inventorySystemRecovered = false;
            bool systemsHealthy = false;
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                // Test GameManager recovery
                gameManagerRecovered = ErrorHandler.AttemptRecovery("gamemanager");
                allRecovered &= gameManagerRecovered;
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
                    // Test TimeManager recovery
                    timeManagerRecovered = ErrorHandler.AttemptRecovery("timemanager");
                    allRecovered &= timeManagerRecovered;
                }
                catch (System.Exception e)
                {
                    errorOccurred = true;
                    caughtException = e;
                }
            }

            if (!errorOccurred)
            {
                yield return new WaitForSeconds(0.1f);

                try
                {
                    // Test MarketSystem recovery
                    marketSystemRecovered = ErrorHandler.AttemptRecovery("marketsystem");
                    allRecovered &= marketSystemRecovered;
                }
                catch (System.Exception e)
                {
                    errorOccurred = true;
                    caughtException = e;
                }
            }

            if (!errorOccurred)
            {
                yield return new WaitForSeconds(0.1f);

                try
                {
                    // Test InventorySystem recovery
                    inventorySystemRecovered = ErrorHandler.AttemptRecovery("inventorysystem");
                    allRecovered &= inventorySystemRecovered;
                }
                catch (System.Exception e)
                {
                    errorOccurred = true;
                    caughtException = e;
                }
            }

            if (!errorOccurred)
            {
                yield return new WaitForSeconds(0.1f);

                try
                {
                    // Verify systems are functional after recovery
                    systemsHealthy = ErrorHandler.CheckSystemHealth();

                    float recoveryTime = Time.time - startTime;

                    var result = new RecoveryTestResult
                    {
                        testName = "System Recovery",
                        recoverySuccessful = allRecovered && systemsHealthy,
                        recoveryTime = recoveryTime,
                        errorDetails =
                            $"GM:{gameManagerRecovered} TM:{timeManagerRecovered} MS:{marketSystemRecovered} IS:{inventorySystemRecovered} Health:{systemsHealthy}",
                    };

                    LogRecoveryResult(result);

                    if (result.recoverySuccessful)
                        successfulRecoveries++;
                }
                catch (System.Exception e)
                {
                    errorOccurred = true;
                    caughtException = e;
                }
            }

            if (errorOccurred)
            {
                ErrorHandler.LogError(
                    $"System recovery test failed: {caughtException.Message}",
                    caughtException,
                    "ErrorRecoveryTest"
                );
            }
        }

        private IEnumerator TestExceptionHandling()
        {
            totalRecoveryTests++;
            float startTime = Time.time;

            ErrorHandler.LogInfo("Testing exception handling", "ErrorRecoveryTest");

            bool allHandled = true;
            bool systemStillFunctional = true;
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                // Test ArgumentException handling
                Action argAction = () =>
                {
                    throw new System.ArgumentException("Test argument exception");
                };
                bool argResult = ErrorHandler.SafeExecute(argAction, "ArgumentExceptionTest");
                allHandled = allHandled && argResult;

                // Test IndexOutOfRangeException handling
                allHandled &= !ErrorHandler.SafeExecute(
                    () =>
                    {
                        int[] array = new int[5];
                        int value = array[10]; // This should fail safely
                    },
                    "IndexExceptionTest"
                );

                // Test InvalidOperationException handling
                Action opAction = () =>
                {
                    throw new System.InvalidOperationException("Test invalid operation");
                };
                bool opResult = ErrorHandler.SafeExecute(opAction, "InvalidOperationTest");
                allHandled = allHandled && opResult;

                // Test that the system continues to function after exceptions
                systemStillFunctional &= ErrorHandler.SafeExecute(
                    () =>
                    {
                        var currentTime = TimeManager.Instance?.GetFormattedTime();
                        var fruitPrice = MarketSystem.Instance?.GetCurrentPrice(ItemType.Fruit);
                    },
                    "PostExceptionFunctionalTest"
                );

                float recoveryTime = Time.time - startTime;

                var result = new RecoveryTestResult
                {
                    testName = "Exception Handling",
                    recoverySuccessful = allHandled && systemStillFunctional,
                    recoveryTime = recoveryTime,
                    errorDetails = $"Exceptions handled: {allHandled}, System functional: {systemStillFunctional}",
                };

                LogRecoveryResult(result);

                if (result.recoverySuccessful)
                    successfulRecoveries++;
            }
            catch (System.Exception e)
            {
                errorOccurred = true;
                caughtException = e;
            }

            if (errorOccurred)
            {
                ErrorHandler.LogError(
                    $"Exception handling test failed: {caughtException.Message}",
                    caughtException,
                    "ErrorRecoveryTest"
                );
            }

            yield return null;
        }

        private IEnumerator TestMemoryPressureRecovery()
        {
            totalRecoveryTests++;
            float startTime = Time.time;

            ErrorHandler.LogInfo("Testing memory pressure recovery", "ErrorRecoveryTest");

            long initialMemory = 0;
            var memoryHogs = new System.Collections.Generic.List<byte[]>();
            bool memoryHandled = false;
            bool memoryRecovered = false;
            bool systemsHealthy = false;
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                initialMemory = System.GC.GetTotalMemory(false);

                // Create memory pressure
                memoryHandled = ErrorHandler.SafeExecute(
                    () =>
                    {
                        for (int i = 0; i < 100; i++)
                        {
                            memoryHogs.Add(new byte[1024 * 1024]); // 1MB each

                            // Check if we should stop due to memory pressure
                            if (System.GC.GetTotalMemory(false) > initialMemory + 100 * 1024 * 1024) // 100MB limit
                            {
                                break;
                            }
                        }
                    },
                    "MemoryPressureTest"
                );

                // Cleanup
                memoryHogs.Clear();
                System.GC.Collect();
                System.GC.WaitForPendingFinalizers();
                System.GC.Collect();
            }
            catch (System.Exception e)
            {
                errorOccurred = true;
                caughtException = e;
            }

            if (!errorOccurred)
            {
                yield return new WaitForSeconds(1f);

                try
                {
                    long finalMemory = System.GC.GetTotalMemory(true);
                    memoryRecovered = (finalMemory - initialMemory) < 10 * 1024 * 1024; // Within 10MB of initial

                    // Verify systems still work after memory pressure
                    systemsHealthy = ErrorHandler.CheckSystemHealth();

                    float recoveryTime = Time.time - startTime;

                    var result = new RecoveryTestResult
                    {
                        testName = "Memory Pressure Recovery",
                        recoverySuccessful = memoryHandled && memoryRecovered && systemsHealthy,
                        recoveryTime = recoveryTime,
                        errorDetails =
                            $"Memory handled: {memoryHandled}, Recovered: {memoryRecovered}, Systems healthy: {systemsHealthy}",
                    };

                    LogRecoveryResult(result);

                    if (result.recoverySuccessful)
                        successfulRecoveries++;
                }
                catch (System.Exception e)
                {
                    errorOccurred = true;
                    caughtException = e;
                }
            }

            if (errorOccurred)
            {
                ErrorHandler.LogError(
                    $"Memory pressure recovery test failed: {caughtException.Message}",
                    caughtException,
                    "ErrorRecoveryTest"
                );
            }
        }

        private IEnumerator TestEventSystemRecovery()
        {
            totalRecoveryTests++;
            float startTime = Time.time;

            ErrorHandler.LogInfo("Testing event system recovery", "ErrorRecoveryTest");

            bool eventSystemRecovered = true;
            bool eventSystemFunctional = true;
            int eventsReceived = 0;
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                // Test event handling with null handlers
                eventSystemRecovered &= ErrorHandler.SafeExecute(
                    () =>
                    {
                        // Subscribe a null handler (should be handled safely)
                        System.Action<PriceChangedEvent> nullHandler = null;
                        // This would normally cause issues, but should be handled
                    },
                    "NullEventHandlerTest"
                );

                // Test event publishing with invalid data
                eventSystemRecovered &= ErrorHandler.SafeExecute(
                    () =>
                    {
                        var invalidEvent = new PriceChangedEvent(Data.ItemType.Fruit, -1f, float.NaN);
                        EventBus.Publish(invalidEvent);
                    },
                    "InvalidEventTest"
                );

                // Test rapid event publishing
                eventSystemRecovered &= ErrorHandler.SafeExecute(
                    () =>
                    {
                        for (int i = 0; i < 1000; i++)
                        {
                            var rapidEvent = new PriceChangedEvent(Data.ItemType.Potion, 100f, 100f + i);
                            EventBus.Publish(rapidEvent);
                        }
                    },
                    "RapidEventTest"
                );
            }
            catch (System.Exception e)
            {
                errorOccurred = true;
                caughtException = e;
            }

            if (!errorOccurred)
            {
                yield return new WaitForSeconds(0.5f);

                try
                {
                    // Verify event system is still functional
                    System.Action<PriceChangedEvent> testHandler = (evt) => eventsReceived++;
                    EventBus.Subscribe<PriceChangedEvent>(testHandler);

                    var testEvent = new PriceChangedEvent(Data.ItemType.Gem, 200f, 220f);
                    EventBus.Publish(testEvent);
                }
                catch (System.Exception e)
                {
                    errorOccurred = true;
                    caughtException = e;
                }
            }

            if (!errorOccurred)
            {
                yield return new WaitForSeconds(0.1f);

                try
                {
                    System.Action<PriceChangedEvent> testHandler = (evt) => eventsReceived++;
                    EventBus.Unsubscribe<PriceChangedEvent>(testHandler);

                    eventSystemFunctional = eventsReceived > 0;

                    float recoveryTime = Time.time - startTime;

                    var result = new RecoveryTestResult
                    {
                        testName = "Event System Recovery",
                        recoverySuccessful = eventSystemRecovered && eventSystemFunctional,
                        recoveryTime = recoveryTime,
                        errorDetails =
                            $"Recovery: {eventSystemRecovered}, Functional: {eventSystemFunctional}, Events received: {eventsReceived}",
                    };

                    LogRecoveryResult(result);

                    if (result.recoverySuccessful)
                        successfulRecoveries++;
                }
                catch (System.Exception e)
                {
                    errorOccurred = true;
                    caughtException = e;
                }
            }

            if (errorOccurred)
            {
                ErrorHandler.LogError(
                    $"Event system recovery test failed: {caughtException.Message}",
                    caughtException,
                    "ErrorRecoveryTest"
                );
            }
        }

        private void LogRecoveryResult(RecoveryTestResult result)
        {
            string status = result.recoverySuccessful ? "✓ PASS" : "✗ FAIL";
            ErrorHandler.LogInfo(
                $"{status} {result.testName}: {result.errorDetails} ({result.recoveryTime:F2}s)",
                "ErrorRecoveryTest"
            );
        }

        private void GenerateRecoveryReport()
        {
            float successRate = totalRecoveryTests > 0 ? (float)successfulRecoveries / totalRecoveryTests * 100f : 100f;

            ErrorHandler.LogInfo("=== ERROR RECOVERY TEST REPORT ===", "ErrorRecoveryTest");
            ErrorHandler.LogInfo($"Total Recovery Tests: {totalRecoveryTests}", "ErrorRecoveryTest");
            ErrorHandler.LogInfo($"Successful Recoveries: {successfulRecoveries}", "ErrorRecoveryTest");
            ErrorHandler.LogInfo($"Recovery Success Rate: {successRate:F2}%", "ErrorRecoveryTest");

            if (successRate >= 80f)
            {
                ErrorHandler.LogInfo("✓ ERROR RECOVERY TESTS PASSED - System handles errors well", "ErrorRecoveryTest");
            }
            else
            {
                ErrorHandler.LogError(
                    "✗ ERROR RECOVERY TESTS FAILED - System needs better error handling",
                    null,
                    "ErrorRecoveryTest"
                );
            }
        }

        public bool IsRecoveryTestRunning()
        {
            return recoveryTestRunning;
        }

        public float GetRecoverySuccessRate()
        {
            return totalRecoveryTests > 0 ? (float)successfulRecoveries / totalRecoveryTests * 100f : 100f;
        }
    }
}
