using System.Collections;
using MerchantTails.Core;
using MerchantTails.Data;
using MerchantTails.Core;
using MerchantTails.Inventory;
using MerchantTails.Market;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.Testing
{
    /// <summary>
    /// システム動作確認用のテストコントローラー
    /// 開発中のデバッグとテストに使用
    /// </summary>
    public class SystemTestController : MonoBehaviour
    {
        [Header("UI References")]
        [SerializeField]
        private Canvas debugCanvas;

        [SerializeField]
        private TextMeshProUGUI statusText;

        [SerializeField]
        private Button timeAdvanceButton;

        [SerializeField]
        private Button marketTestButton;

        [SerializeField]
        private Button inventoryTestButton;

        [SerializeField]
        private Button logAllStatesButton;

        [Header("Test Settings")]
        [SerializeField]
        private bool autoRunTests = true;

        [SerializeField]
        private float testInterval = 2.0f;

        private bool testSystemsReady = false;
        private Coroutine autoTestCoroutine;

        private void Start()
        {
            StartCoroutine(InitializeTestSystems());
        }

        private IEnumerator InitializeTestSystems()
        {
            UpdateStatus("Initializing test systems...");

            // Wait for core systems to initialize
            yield return new WaitForSeconds(0.5f);

            // Check system availability
            bool gameManagerReady = GameManager.Instance != null;
            bool timeManagerReady = TimeManager.Instance != null;
            bool marketSystemReady = MarketSystem.Instance != null;
            bool inventorySystemReady = InventorySystem.Instance != null;

            UpdateStatus(
                $"Systems Ready: GM:{gameManagerReady} TM:{timeManagerReady} MS:{marketSystemReady} IS:{inventorySystemReady}"
            );

            if (gameManagerReady && timeManagerReady && marketSystemReady && inventorySystemReady)
            {
                testSystemsReady = true;
                SetupUI();
                SubscribeToEvents();

                if (autoRunTests)
                {
                    autoTestCoroutine = StartCoroutine(AutoTestRoutine());
                }

                UpdateStatus("All systems ready! Test controller active.");
            }
            else
            {
                UpdateStatus("ERROR: Some systems failed to initialize!");
                yield return new WaitForSeconds(2f);
                StartCoroutine(InitializeTestSystems()); // Retry
            }
        }

        private void SetupUI()
        {
            if (timeAdvanceButton != null)
                timeAdvanceButton.onClick.AddListener(TestTimeAdvancement);

            if (marketTestButton != null)
                marketTestButton.onClick.AddListener(TestMarketSystem);

            if (inventoryTestButton != null)
                inventoryTestButton.onClick.AddListener(TestInventorySystem);

            if (logAllStatesButton != null)
                logAllStatesButton.onClick.AddListener(LogAllSystemStates);
        }

        private void SubscribeToEvents()
        {
            EventBus.Subscribe<PhaseChangedEvent>(OnPhaseChanged);
            EventBus.Subscribe<SeasonChangedEvent>(OnSeasonChanged);
            EventBus.Subscribe<PriceChangedEvent>(OnPriceChanged);
            EventBus.Subscribe<DayChangedEvent>(OnDayChanged);
        }

        private void OnDestroy()
        {
            if (autoTestCoroutine != null)
            {
                StopCoroutine(autoTestCoroutine);
            }

            EventBus.Unsubscribe<PhaseChangedEvent>(OnPhaseChanged);
            EventBus.Unsubscribe<SeasonChangedEvent>(OnSeasonChanged);
            EventBus.Unsubscribe<PriceChangedEvent>(OnPriceChanged);
            EventBus.Unsubscribe<DayChangedEvent>(OnDayChanged);
        }

        private IEnumerator AutoTestRoutine()
        {
            while (testSystemsReady)
            {
                yield return new WaitForSeconds(testInterval);

                // Perform periodic tests
                TestBasicFunctionality();

                // Occasionally run more comprehensive tests
                if (Time.time % 10f < testInterval)
                {
                    TestMarketSystem();
                    TestInventorySystem();
                }
            }
        }

        private void TestBasicFunctionality()
        {
            if (!testSystemsReady)
                return;

            // Test basic system availability
            bool allSystemsOk = true;

            try
            {
                var currentTime = TimeManager.Instance.GetFormattedTime();
                var fruitPrice = MarketSystem.Instance.GetCurrentPrice(ItemType.Fruit);
                var storefrontCapacity = InventorySystem.Instance.StorefrontCapacityRemaining;

                UpdateStatus($"Time: {currentTime} | Fruit: {fruitPrice:F1}G | Capacity: {storefrontCapacity}");
            }
            catch (System.Exception e)
            {
                UpdateStatus($"ERROR in basic test: {e.Message}");
                allSystemsOk = false;
            }

            if (!allSystemsOk)
            {
                Debug.LogError("[SystemTestController] Basic functionality test failed!");
            }
        }

        public void TestTimeAdvancement()
        {
            if (!testSystemsReady)
                return;

            Debug.Log("[SystemTestController] Testing time advancement...");

            TimeManager.Instance.SkipToNextPhase();
            UpdateStatus("Time advanced to next phase");
        }

        public void TestMarketSystem()
        {
            if (!testSystemsReady)
                return;

            Debug.Log("[SystemTestController] Testing market system...");

            // Test price retrieval for all items
            foreach (ItemType itemType in System.Enum.GetValues(typeof(ItemType)))
            {
                float currentPrice = MarketSystem.Instance.GetCurrentPrice(itemType);
                float basePrice = MarketSystem.Instance.GetBasePrice(itemType);
                var marketData = MarketSystem.Instance.GetMarketData(itemType);

                Debug.Log(
                    $"[MarketTest] {itemType}: Current={currentPrice:F2}G, Base={basePrice:F2}G, "
                        + $"Demand={marketData.demand:F2}, Supply={marketData.supply:F2}"
                );
            }

            // Test market event simulation
            var testEvent = new GameEventTriggeredEvent(
                "Test Harvest Festival",
                "Testing event effects",
                new ItemType[] { ItemType.Fruit, ItemType.Potion },
                new float[] { 1.5f, 0.8f },
                5
            );

            EventBus.Publish(testEvent);
            UpdateStatus("Market system test completed");
        }

        public void TestInventorySystem()
        {
            if (!testSystemsReady)
                return;

            Debug.Log("[SystemTestController] Testing inventory system...");

            // Test adding items
            bool addResult = InventorySystem.Instance.AddItem(ItemType.Fruit, 5, InventoryLocation.Trading);
            Debug.Log($"[InventoryTest] Add 5 fruits to trading: {addResult}");

            // Test moving items
            bool moveResult = InventorySystem.Instance.MoveItem(
                ItemType.Fruit,
                2,
                InventoryLocation.Trading,
                InventoryLocation.Storefront
            );
            Debug.Log($"[InventoryTest] Move 2 fruits to storefront: {moveResult}");

            // Test inventory counts
            int tradingCount = InventorySystem.Instance.GetItemCount(ItemType.Fruit, InventoryLocation.Trading);
            int storefrontCount = InventorySystem.Instance.GetItemCount(ItemType.Fruit, InventoryLocation.Storefront);

            Debug.Log($"[InventoryTest] Fruit count - Trading: {tradingCount}, Storefront: {storefrontCount}");

            // Test expiring items
            var expiringItems = InventorySystem.Instance.GetExpiringItems(1);
            Debug.Log($"[InventoryTest] Items expiring soon: {expiringItems.Count}");

            UpdateStatus($"Inventory test completed - T:{tradingCount} S:{storefrontCount}");
        }

        public void LogAllSystemStates()
        {
            if (!testSystemsReady)
                return;

            Debug.Log("[SystemTestController] Logging all system states...");

            // Log TimeManager state
            TimeManager.Instance.LogCurrentTimeState();

            // Log MarketSystem state
            MarketSystem.Instance.LogMarketState();

            // Log InventorySystem state
            InventorySystem.Instance.LogInventoryState();

            // Log EventBus state
            EventBus.LogCurrentState();

            UpdateStatus("All system states logged to console");
        }

        // Event handlers
        private void OnPhaseChanged(PhaseChangedEvent evt)
        {
            Debug.Log($"[SystemTestController] Phase changed: {evt.PreviousPhase} -> {evt.NewPhase}");
        }

        private void OnSeasonChanged(SeasonChangedEvent evt)
        {
            Debug.Log($"[SystemTestController] Season changed: {evt.PreviousSeason} -> {evt.NewSeason}");
            UpdateStatus($"Season changed to {evt.NewSeason}!");
        }

        private void OnPriceChanged(PriceChangedEvent evt)
        {
            if (Mathf.Abs(evt.ChangePercentage) > 5f) // Only log significant changes
            {
                Debug.Log(
                    $"[SystemTestController] Price change: {evt.ItemType} "
                        + $"{evt.PreviousPrice:F2}G -> {evt.NewPrice:F2}G ({evt.ChangePercentage:+F1}%)"
                );
            }
        }

        private void OnDayChanged(DayChangedEvent evt)
        {
            Debug.Log(
                $"[SystemTestController] Day changed: {evt.PreviousDay} -> {evt.NewDay} "
                    + $"({evt.CurrentSeason}, Year {evt.CurrentYear})"
            );
        }

        private void UpdateStatus(string message)
        {
            Debug.Log($"[SystemTestController] {message}");

            if (statusText != null)
            {
                statusText.text = $"[{System.DateTime.Now:HH:mm:ss}] {message}";
            }
        }

        // Unity Editor GUI for testing
        private void OnGUI()
        {
            if (!testSystemsReady)
                return;

            GUILayout.BeginArea(new Rect(10, 10, 300, 200));
            GUILayout.Label("System Test Controller", GUI.skin.box);

            if (GUILayout.Button("Advance Time"))
                TestTimeAdvancement();

            if (GUILayout.Button("Test Market"))
                TestMarketSystem();

            if (GUILayout.Button("Test Inventory"))
                TestInventorySystem();

            if (GUILayout.Button("Log All States"))
                LogAllSystemStates();

            GUILayout.Space(10);

            if (TimeManager.Instance != null)
            {
                GUILayout.Label($"Time: {TimeManager.Instance.GetFormattedTime()}");
                GUILayout.Label($"Phase Progress: {TimeManager.Instance.GetPhaseProgress():P1}");
            }

            if (MarketSystem.Instance != null)
            {
                GUILayout.Label($"Fruit Price: {MarketSystem.Instance.GetCurrentPrice(ItemType.Fruit):F1}G");
            }

            GUILayout.EndArea();
        }
    }
}
