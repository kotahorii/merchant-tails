using System.Collections;
using UnityEngine;
using UnityEngine.UI;
using TMPro;
using MerchantTails.Core;

namespace MerchantTails.Testing
{
    /// <summary>
    /// 統合テスト実行管理クラス
    /// SystemTestControllerとIntegrationTestを連携管理
    /// </summary>
    public class TestRunner : MonoBehaviour
    {
        [Header("UI References")]
        [SerializeField] private Canvas testCanvas;
        [SerializeField] private TextMeshProUGUI statusDisplay;
        [SerializeField] private Button runIntegrationTestsButton;
        [SerializeField] private Button runSystemTestsButton;
        [SerializeField] private Button runStabilityTestsButton;
        [SerializeField] private Button runErrorRecoveryTestsButton;
        [SerializeField] private Button healthCheckButton;

        [Header("Components")]
        [SerializeField] private SystemTestController systemTestController;
        [SerializeField] private IntegrationTest integrationTest;
        [SerializeField] private StabilityTest stabilityTest;
        [SerializeField] private ErrorRecoveryTest errorRecoveryTest;

        private bool testsRunning = false;

        private void Start()
        {
            InitializeTestRunner();
        }

        private void InitializeTestRunner()
        {
            // Initialize ErrorHandler
            ErrorHandler.Initialize();

            // Set up UI
            if (runIntegrationTestsButton != null)
                runIntegrationTestsButton.onClick.AddListener(RunIntegrationTests);

            if (runSystemTestsButton != null)
                runSystemTestsButton.onClick.AddListener(RunSystemTests);

            if (runStabilityTestsButton != null)
                runStabilityTestsButton.onClick.AddListener(RunStabilityTests);

            if (runErrorRecoveryTestsButton != null)
                runErrorRecoveryTestsButton.onClick.AddListener(RunErrorRecoveryTests);

            if (healthCheckButton != null)
                healthCheckButton.onClick.AddListener(RunHealthCheck);

            // Get component references if not assigned
            if (systemTestController == null)
                systemTestController = GetComponent<SystemTestController>();

            if (integrationTest == null)
                integrationTest = GetComponent<IntegrationTest>();

            if (stabilityTest == null)
                stabilityTest = GetComponent<StabilityTest>();

            if (errorRecoveryTest == null)
                errorRecoveryTest = GetComponent<ErrorRecoveryTest>();

            UpdateStatus("Test Runner initialized. Ready to run tests.");
        }

        public void RunIntegrationTests()
        {
            if (testsRunning) return;

            if (integrationTest != null)
            {
                testsRunning = true;
                UpdateStatus("Starting integration tests...");
                integrationTest.StartIntegrationTests();
                StartCoroutine(WaitForIntegrationTestsComplete());
            }
            else
            {
                UpdateStatus("ERROR: IntegrationTest component not found!");
            }
        }

        public void RunSystemTests()
        {
            if (testsRunning) return;

            if (systemTestController != null)
            {
                testsRunning = true;
                UpdateStatus("Running system tests...");

                // Run comprehensive system tests
                systemTestController.TestMarketSystem();
                systemTestController.TestInventorySystem();
                systemTestController.LogAllSystemStates();

                StartCoroutine(CompleteSystemTests());
            }
            else
            {
                UpdateStatus("ERROR: SystemTestController component not found!");
            }
        }

        public void RunStabilityTests()
        {
            if (testsRunning) return;

            if (stabilityTest != null)
            {
                testsRunning = true;
                UpdateStatus("Starting stability tests...");
                stabilityTest.StartStabilityTest();
                StartCoroutine(WaitForStabilityTestsComplete());
            }
            else
            {
                UpdateStatus("ERROR: StabilityTest component not found!");
            }
        }

        public void RunErrorRecoveryTests()
        {
            if (testsRunning) return;

            if (errorRecoveryTest != null)
            {
                testsRunning = true;
                UpdateStatus("Starting error recovery tests...");
                errorRecoveryTest.StartErrorRecoveryTest();
                StartCoroutine(WaitForErrorRecoveryTestsComplete());
            }
            else
            {
                UpdateStatus("ERROR: ErrorRecoveryTest component not found!");
            }
        }

        public void RunHealthCheck()
        {
            UpdateStatus("Running system health check...");

            bool healthResult = ErrorHandler.CheckSystemHealth();
            string healthStatus = healthResult ? "HEALTHY" : "ISSUES DETECTED";

            UpdateStatus($"Health Check: {healthStatus}");
        }

        private IEnumerator WaitForIntegrationTestsComplete()
        {
            while (integrationTest.IsTestInProgress())
            {
                yield return new WaitForSeconds(0.5f);
            }

            integrationTest.GetTestSummary(out int passed, out int failed);
            UpdateStatus($"Integration tests complete: {passed} passed, {failed} failed");
            testsRunning = false;
        }

        private IEnumerator WaitForStabilityTestsComplete()
        {
            while (stabilityTest.IsStabilityTestRunning())
            {
                yield return new WaitForSeconds(1f);
            }

            var report = stabilityTest.GetLastReport();
            UpdateStatus($"Stability tests complete: {report.successRate:F1}% success rate, {report.criticalErrors.Count} errors");
            testsRunning = false;
        }

        private IEnumerator WaitForErrorRecoveryTestsComplete()
        {
            while (errorRecoveryTest.IsRecoveryTestRunning())
            {
                yield return new WaitForSeconds(0.5f);
            }

            float successRate = errorRecoveryTest.GetRecoverySuccessRate();
            UpdateStatus($"Error recovery tests complete: {successRate:F1}% success rate");
            testsRunning = false;
        }

        private IEnumerator CompleteSystemTests()
        {
            yield return new WaitForSeconds(2f);
            UpdateStatus("System tests completed");
            testsRunning = false;
        }

        private void UpdateStatus(string message)
        {
            ErrorHandler.LogInfo(message, "TestRunner");

            if (statusDisplay != null)
            {
                statusDisplay.text = $"[{System.DateTime.Now:HH:mm:ss}] {message}";
            }
        }

        private void OnDestroy()
        {
            ErrorHandler.Cleanup();
        }
    }
}
