using System.Collections;
using MerchantTails.Core;
using MerchantTails.Inventory;
using MerchantTails.Market;
using UnityEngine;

namespace MerchantTails.Testing
{
    /// <summary>
    /// CI環境でのテスト検証クラス
    /// 実際のテスト実行前に基本検証を行う
    /// </summary>
    public class CITestValidator : MonoBehaviour
    {
        [Header("CI Test Settings")]
        [SerializeField]
        private bool autoValidateOnStart = true;

        [SerializeField]
        private float validationTimeout = 10f;

        private bool validationComplete = false;
        private bool validationSuccess = false;

        private void Start()
        {
            if (autoValidateOnStart)
            {
                StartCoroutine(RunValidation());
            }
        }

        public IEnumerator RunValidation()
        {
            ErrorHandler.LogInfo("Starting CI validation", "CITestValidator");

            validationComplete = false;
            validationSuccess = true;

            // Test 1: Basic system availability
            yield return StartCoroutine(ValidateSystemAvailability());

            // Test 2: Script compilation
            yield return StartCoroutine(ValidateScriptCompilation());

            // Test 3: Component instantiation
            yield return StartCoroutine(ValidateComponentInstantiation());

            // Test 4: Basic functionality
            yield return StartCoroutine(ValidateBasicFunctionality());

            validationComplete = true;

            string result = validationSuccess ? "PASSED" : "FAILED";
            ErrorHandler.LogInfo($"CI validation {result}", "CITestValidator");

            if (!validationSuccess)
            {
                ErrorHandler.LogError("CI validation failed - tests may not run correctly", null, "CITestValidator");
            }
        }

        private IEnumerator ValidateSystemAvailability()
        {
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                // Check if Unity is running properly
                bool unityRunning = Application.isPlaying;
                if (!unityRunning)
                {
                    ErrorHandler.LogError("Unity not in play mode", null, "CITestValidator");
                    validationSuccess = false;
                }

                // Check platform
                string platform = Application.platform.ToString();
                ErrorHandler.LogInfo($"Running on platform: {platform}", "CITestValidator");

                // Check memory availability
                long totalMemory = System.GC.GetTotalMemory(false);
                ErrorHandler.LogInfo($"Initial memory usage: {totalMemory / 1024 / 1024}MB", "CITestValidator");
            }
            catch (System.Exception e)
            {
                errorOccurred = true;
                caughtException = e;
            }

            if (errorOccurred)
            {
                ErrorHandler.LogError(
                    $"System availability validation failed: {caughtException.Message}",
                    caughtException,
                    "CITestValidator"
                );
                validationSuccess = false;
            }

            yield return new WaitForSeconds(0.1f);
        }

        private IEnumerator ValidateScriptCompilation()
        {
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                // Check if core types exist
                var gameManagerType = System.Type.GetType("MerchantTails.Core.GameManager");
                var timeManagerType = System.Type.GetType("MerchantTails.Core.TimeManager");
                var marketSystemType = System.Type.GetType("MerchantTails.Market.MarketSystem");
                var inventorySystemType = System.Type.GetType("MerchantTails.Inventory.InventorySystem");

                if (gameManagerType == null)
                {
                    ErrorHandler.LogError("GameManager type not found", null, "CITestValidator");
                    validationSuccess = false;
                }

                if (timeManagerType == null)
                {
                    ErrorHandler.LogError("TimeManager type not found", null, "CITestValidator");
                    validationSuccess = false;
                }

                if (marketSystemType == null)
                {
                    ErrorHandler.LogError("MarketSystem type not found", null, "CITestValidator");
                    validationSuccess = false;
                }

                if (inventorySystemType == null)
                {
                    ErrorHandler.LogError("InventorySystem type not found", null, "CITestValidator");
                    validationSuccess = false;
                }

                // Check test types
                var testRunnerType = System.Type.GetType("MerchantTails.Testing.TestRunner");
                var integrationTestType = System.Type.GetType("MerchantTails.Testing.IntegrationTest");

                if (testRunnerType == null)
                {
                    ErrorHandler.LogError("TestRunner type not found", null, "CITestValidator");
                    validationSuccess = false;
                }

                if (integrationTestType == null)
                {
                    ErrorHandler.LogError("IntegrationTest type not found", null, "CITestValidator");
                    validationSuccess = false;
                }

                ErrorHandler.LogInfo("Script compilation validation completed", "CITestValidator");
            }
            catch (System.Exception e)
            {
                errorOccurred = true;
                caughtException = e;
            }

            if (errorOccurred)
            {
                ErrorHandler.LogError(
                    $"Script compilation validation failed: {caughtException.Message}",
                    caughtException,
                    "CITestValidator"
                );
                validationSuccess = false;
            }

            yield return new WaitForSeconds(0.1f);
        }

        private IEnumerator ValidateComponentInstantiation()
        {
            GameObject testGO = null;
            GameManager gameManager = null;
            TimeManager timeManager = null;
            MarketSystem marketSystem = null;
            InventorySystem inventorySystem = null;
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                // Create temporary game object for testing
                testGO = new GameObject("ValidationTest");

                // Try to add core components
                gameManager = testGO.AddComponent<GameManager>();
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
                    timeManager = testGO.AddComponent<TimeManager>();
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
                    marketSystem = testGO.AddComponent<MarketSystem>();
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
                    inventorySystem = testGO.AddComponent<InventorySystem>();
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

                // Verify components were added
                if (gameManager == null)
                {
                    ErrorHandler.LogError("Failed to instantiate GameManager", null, "CITestValidator");
                    validationSuccess = false;
                }

                if (timeManager == null)
                {
                    ErrorHandler.LogError("Failed to instantiate TimeManager", null, "CITestValidator");
                    validationSuccess = false;
                }

                if (marketSystem == null)
                {
                    ErrorHandler.LogError("Failed to instantiate MarketSystem", null, "CITestValidator");
                    validationSuccess = false;
                }

                if (inventorySystem == null)
                {
                    ErrorHandler.LogError("Failed to instantiate InventorySystem", null, "CITestValidator");
                    validationSuccess = false;
                }

                ErrorHandler.LogInfo("Component instantiation validation completed", "CITestValidator");
            }
            else
            {
                ErrorHandler.LogError(
                    $"Component instantiation validation failed: {caughtException.Message}",
                    caughtException,
                    "CITestValidator"
                );
                validationSuccess = false;
            }

            // Cleanup
            if (testGO != null)
            {
                DestroyImmediate(testGO);
            }
        }

        private IEnumerator ValidateBasicFunctionality()
        {
            GameObject testGO = null;
            bool errorOccurred = false;
            System.Exception caughtException = null;

            try
            {
                // Create test environment
                testGO = new GameObject("FunctionalityTest");
                var gameManager = testGO.AddComponent<GameManager>();
                var timeManager = testGO.AddComponent<TimeManager>();
                var marketSystem = testGO.AddComponent<MarketSystem>();
                var inventorySystem = testGO.AddComponent<InventorySystem>();
            }
            catch (System.Exception e)
            {
                errorOccurred = true;
                caughtException = e;
            }

            if (!errorOccurred)
            {
                // Wait for initialization
                yield return new WaitForSeconds(1f);

                try
                {
                    // Test basic functionality
                    if (GameManager.Instance != null)
                    {
                        ErrorHandler.LogInfo("GameManager singleton working", "CITestValidator");
                    }
                    else
                    {
                        ErrorHandler.LogError("GameManager singleton not working", null, "CITestValidator");
                        validationSuccess = false;
                    }

                    if (TimeManager.Instance != null)
                    {
                        var currentTime = TimeManager.Instance.GetFormattedTime();
                        ErrorHandler.LogInfo($"TimeManager working: {currentTime}", "CITestValidator");
                    }
                    else
                    {
                        ErrorHandler.LogError("TimeManager not working", null, "CITestValidator");
                        validationSuccess = false;
                    }

                    if (MarketSystem.Instance != null)
                    {
                        var fruitPrice = MarketSystem.Instance.GetCurrentPrice(MerchantTails.Data.ItemType.Fruit);
                        ErrorHandler.LogInfo($"MarketSystem working: fruit price {fruitPrice}", "CITestValidator");
                    }
                    else
                    {
                        ErrorHandler.LogError("MarketSystem not working", null, "CITestValidator");
                        validationSuccess = false;
                    }

                    if (InventorySystem.Instance != null)
                    {
                        var capacity = InventorySystem.Instance.StorefrontCapacityRemaining;
                        ErrorHandler.LogInfo($"InventorySystem working: capacity {capacity}", "CITestValidator");
                    }
                    else
                    {
                        ErrorHandler.LogError("InventorySystem not working", null, "CITestValidator");
                        validationSuccess = false;
                    }

                    ErrorHandler.LogInfo("Basic functionality validation completed", "CITestValidator");
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
                    $"Basic functionality validation failed: {caughtException.Message}",
                    caughtException,
                    "CITestValidator"
                );
                validationSuccess = false;
            }

            // Cleanup
            if (testGO != null)
            {
                DestroyImmediate(testGO);
            }
        }

        public bool IsValidationComplete()
        {
            return validationComplete;
        }

        public bool IsValidationSuccessful()
        {
            return validationSuccess;
        }

        /// <summary>
        /// CI環境でのクイック検証
        /// </summary>
        public static bool QuickValidation()
        {
            try
            {
                // Quick type checks
                var gameManagerType = System.Type.GetType("MerchantTails.Core.GameManager");
                var timeManagerType = System.Type.GetType("MerchantTails.Core.TimeManager");
                var marketSystemType = System.Type.GetType("MerchantTails.Market.MarketSystem");
                var inventorySystemType = System.Type.GetType("MerchantTails.Inventory.InventorySystem");

                bool typesExist =
                    gameManagerType != null
                    && timeManagerType != null
                    && marketSystemType != null
                    && inventorySystemType != null;

                if (typesExist)
                {
                    ErrorHandler.LogInfo("Quick validation passed", "CITestValidator");
                    return true;
                }
                else
                {
                    ErrorHandler.LogError("Quick validation failed - core types missing", null, "CITestValidator");
                    return false;
                }
            }
            catch (System.Exception e)
            {
                ErrorHandler.LogError($"Quick validation failed: {e.Message}", e, "CITestValidator");
                return false;
            }
        }
    }
}
