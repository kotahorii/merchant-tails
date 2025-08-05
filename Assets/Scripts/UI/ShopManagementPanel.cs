using System.Collections.Generic;
using MerchantTails.Core;
using MerchantTails.Data;
using MerchantTails.Core;
using MerchantTails.Inventory;
using MerchantTails.Market;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.UI
{
    /// <summary>
    /// ショップ管理画面のUI制御
    /// 店頭商品の管理、価格設定、販売状況の表示
    /// </summary>
    public class ShopManagementPanel : UIPanel
    {
        [Header("Shop Display")]
        [SerializeField]
        private Text shopNameText;

        [SerializeField]
        private Text merchantRankText;

        [SerializeField]
        private Text totalMoneyText;

        [SerializeField]
        private Text dailyProfitText;

        [Header("Time Display")]
        [SerializeField]
        private Text currentTimeText;

        [SerializeField]
        private Text currentSeasonText;

        [SerializeField]
        private Slider timeProgressSlider;

        [Header("Shop Inventory")]
        [SerializeField]
        private Transform shopItemsContainer;

        [SerializeField]
        private GameObject shopItemPrefab;

        [SerializeField]
        private Button restockButton;

        [SerializeField]
        private Text shopCapacityText;

        [Header("Quick Actions")]
        [SerializeField]
        private Button marketButton;

        [SerializeField]
        private Button inventoryButton;

        [SerializeField]
        private Button journalButton;

        [SerializeField]
        private Button settingsButton;

        [Header("Customer Info")]
        [SerializeField]
        private GameObject customerPanel;

        [SerializeField]
        private Text customerCountText;

        [SerializeField]
        private Text demandInfoText;

        private List<ShopItemUI> shopItemUIs = new List<ShopItemUI>();
        private float updateTimer = 0f;
        private const float UPDATE_INTERVAL = 1f;

        protected override void OnInitialize()
        {
            SetupButtons();
            SetupEventListeners();
        }

        protected override void OnShow()
        {
            LogUIAction("Shop Management panel shown");
            RefreshShopDisplay();
            StartPeriodicUpdate();
        }

        protected override void OnHide()
        {
            LogUIAction("Shop Management panel hidden");
            StopPeriodicUpdate();
        }

        private void Update()
        {
            if (IsVisible)
            {
                updateTimer += Time.deltaTime;
                if (updateTimer >= UPDATE_INTERVAL)
                {
                    UpdateTimeDisplay();
                    UpdateShopStatus();
                    updateTimer = 0f;
                }
            }
        }

        private void SetupButtons()
        {
            if (restockButton != null)
                restockButton.onClick.AddListener(OnRestockPressed);

            if (marketButton != null)
                marketButton.onClick.AddListener(OnMarketPressed);

            if (inventoryButton != null)
                inventoryButton.onClick.AddListener(OnInventoryPressed);

            if (journalButton != null)
                journalButton.onClick.AddListener(OnJournalPressed);

            if (settingsButton != null)
                settingsButton.onClick.AddListener(OnSettingsPressed);
        }

        private void SetupEventListeners()
        {
            EventBus.Subscribe<InventoryChangedEvent>(OnInventoryChanged);
            EventBus.Subscribe<PriceChangedEvent>(OnPriceChanged);
            EventBus.Subscribe<TimeAdvancedEvent>(OnTimeAdvanced);
            EventBus.Subscribe<TransactionCompletedEvent>(OnTransactionCompleted);
        }

        private void RefreshShopDisplay()
        {
            UpdateBasicInfo();
            UpdateTimeDisplay();
            UpdateShopInventory();
            UpdateCustomerInfo();
        }

        private void UpdateBasicInfo()
        {
            var playerData = GameManager.Instance?.GetPlayerData();
            if (playerData == null)
                return;

            // ショップ名とランク
            if (shopNameText != null)
                shopNameText.text = $"{playerData.PlayerName}の商店";

            if (merchantRankText != null)
                merchantRankText.text = GetRankDisplayName(playerData.CurrentRank);

            // 所持金
            if (totalMoneyText != null)
                totalMoneyText.text = $"{playerData.CurrentMoney:N0}G";

            // 日次利益
            if (dailyProfitText != null)
            {
                int dailyProfit = CalculateDailyProfit();
                string profitColor = dailyProfit >= 0 ? "#00FF00" : "#FF0000";
                dailyProfitText.text = $"<color={profitColor}>{dailyProfit:+#;-#;0}G</color>";
            }
        }

        private void UpdateTimeDisplay()
        {
            if (TimeManager.Instance == null)
                return;

            // 現在時刻
            if (currentTimeText != null)
                currentTimeText.text = TimeManager.Instance.GetFormattedTime();

            // 現在の季節
            if (currentSeasonText != null)
                currentSeasonText.text = GetSeasonDisplayName(TimeManager.Instance.CurrentSeason);

            // 時間進行バー
            if (timeProgressSlider != null)
                timeProgressSlider.value = TimeManager.Instance.GetPhaseProgress();
        }

        private void UpdateShopInventory()
        {
            if (InventorySystem.Instance == null)
                return;

            // 既存のUIアイテムをクリア
            ClearShopItemUIs();

            // 店頭在庫のアイテムを表示
            var storefrontItems = InventorySystem.Instance.GetStorefrontItems();

            foreach (var item in storefrontItems)
            {
                CreateShopItemUI(item.Key, item.Value);
            }

            // 店頭容量表示
            if (shopCapacityText != null)
            {
                int used = InventorySystem.Instance.StorefrontCapacityUsed;
                int total = InventorySystem.Instance.StorefrontCapacityTotal;
                shopCapacityText.text = $"店頭: {used}/{total}";

                // 容量に応じて色を変更
                float ratio = (float)used / total;
                Color capacityColor = ratio switch
                {
                    >= 0.9f => Color.red,
                    >= 0.7f => Color.yellow,
                    _ => Color.white,
                };
                shopCapacityText.color = capacityColor;
            }
        }

        private void CreateShopItemUI(ItemType itemType, int quantity)
        {
            if (shopItemPrefab == null || shopItemsContainer == null)
                return;

            var itemUI = Instantiate(shopItemPrefab, shopItemsContainer);
            var shopItemUI = itemUI.GetComponent<ShopItemUI>();

            if (shopItemUI == null)
                shopItemUI = itemUI.AddComponent<ShopItemUI>();

            // アイテム情報を設定
            var currentPrice = MarketSystem.Instance?.GetCurrentPrice(itemType) ?? 0f;
            shopItemUI.Setup(itemType, quantity, currentPrice);

            shopItemUIs.Add(shopItemUI);
        }

        private void ClearShopItemUIs()
        {
            foreach (var itemUI in shopItemUIs)
            {
                if (itemUI != null)
                    Destroy(itemUI.gameObject);
            }
            shopItemUIs.Clear();
        }

        private void UpdateCustomerInfo()
        {
            if (customerPanel == null)
                return;

            // 顧客数の表示（簡易実装）
            if (customerCountText != null)
            {
                int customerCount = CalculateCustomerCount();
                customerCountText.text = $"来店客数: {customerCount}人";
            }

            // 需要情報の表示
            if (demandInfoText != null)
            {
                string demandInfo = GetDemandInfo();
                demandInfoText.text = demandInfo;
            }
        }

        private void UpdateShopStatus()
        {
            // 在庫切れアイテムのチェック
            CheckOutOfStockItems();

            // 腐敗アイテムのチェック
            CheckSpoiledItems();
        }

        // ヘルパーメソッド
        private string GetRankDisplayName(MerchantRank rank)
        {
            return rank switch
            {
                MerchantRank.Apprentice => "見習い",
                MerchantRank.Skilled => "一人前",
                MerchantRank.Veteran => "ベテラン",
                MerchantRank.Master => "マスター",
                _ => "商人",
            };
        }

        private string GetSeasonDisplayName(Season season)
        {
            return season switch
            {
                Season.Spring => "春",
                Season.Summer => "夏",
                Season.Autumn => "秋",
                Season.Winter => "冬",
                _ => "不明",
            };
        }

        private int CalculateDailyProfit()
        {
            // 本日の利益計算（簡易実装）
            return Random.Range(-500, 2000);
        }

        private int CalculateCustomerCount()
        {
            // 時間帯と季節に基づく顧客数計算
            var timeManager = TimeManager.Instance;
            if (timeManager == null)
                return 0;

            int baseCustomers = timeManager.CurrentPhase switch
            {
                DayPhase.Morning => 15,
                DayPhase.Afternoon => 25,
                DayPhase.Evening => 20,
                DayPhase.Night => 5,
                _ => 10,
            };

            // 季節補正
            float seasonMultiplier = timeManager.CurrentSeason switch
            {
                Season.Spring => 1.1f,
                Season.Summer => 1.2f,
                Season.Autumn => 1.0f,
                Season.Winter => 0.8f,
                _ => 1.0f,
            };

            return Mathf.RoundToInt(baseCustomers * seasonMultiplier);
        }

        private string GetDemandInfo()
        {
            var timeManager = TimeManager.Instance;
            if (timeManager == null)
                return "需要情報なし";

            // 季節に応じた需要情報
            return timeManager.CurrentSeason switch
            {
                Season.Spring => "くだものとポーションの需要が高まっています",
                Season.Summer => "ポーションとアクセサリーが人気です",
                Season.Autumn => "武器と魔法書の需要が増加中",
                Season.Winter => "宝石と魔法書が求められています",
                _ => "標準的な需要です",
            };
        }

        private void CheckOutOfStockItems()
        {
            // 在庫切れアイテムの通知（簡易実装）
            var storefrontItems = InventorySystem.Instance?.GetStorefrontItems();
            if (storefrontItems == null)
                return;

            foreach (var item in storefrontItems)
            {
                if (item.Value == 0)
                {
                    // 在庫切れ警告（UI更新）
                    UpdateItemUI(item.Key, true);
                }
            }
        }

        private void CheckSpoiledItems()
        {
            // 腐敗アイテムのチェック（将来実装）
        }

        private void UpdateItemUI(ItemType itemType, bool outOfStock)
        {
            var itemUI = shopItemUIs.Find(ui => ui.ItemType == itemType);
            if (itemUI != null)
            {
                itemUI.SetOutOfStock(outOfStock);
            }
        }

        // ボタンイベントハンドラー
        private void OnRestockPressed()
        {
            LogUIAction("Restock button pressed");
            UIManager.Instance.ShowPanel(UIType.Inventory);
        }

        private void OnMarketPressed()
        {
            LogUIAction("Market button pressed");
            UIManager.Instance.ShowPanel(UIType.MarketAnalysis);
        }

        private void OnInventoryPressed()
        {
            LogUIAction("Inventory button pressed");
            UIManager.Instance.ShowPanel(UIType.Inventory);
        }

        private void OnJournalPressed()
        {
            LogUIAction("Journal button pressed");
            UIManager.Instance.ShowPanel(UIType.MerchantJournal);
        }

        private void OnSettingsPressed()
        {
            LogUIAction("Settings button pressed");
            UIManager.Instance.ShowModal(UIType.Settings);
        }

        // イベントハンドラー
        private void OnInventoryChanged(InventoryChangedEvent evt)
        {
            if (evt.Location == InventoryLocation.Storefront)
            {
                UpdateShopInventory();
            }
        }

        private void OnPriceChanged(PriceChangedEvent evt)
        {
            // 価格変更時のUI更新
            var itemUI = shopItemUIs.Find(ui => ui.ItemType == evt.ItemType);
            if (itemUI != null)
            {
                itemUI.UpdatePrice(evt.NewPrice);
            }
        }

        private void OnTimeAdvanced(TimeAdvancedEvent evt)
        {
            UpdateTimeDisplay();
            UpdateCustomerInfo();
        }

        private void OnTransactionCompleted(TransactionCompletedEvent evt)
        {
            UpdateBasicInfo();
            UpdateShopInventory();
        }

        private void StartPeriodicUpdate()
        {
            updateTimer = 0f;
        }

        private void StopPeriodicUpdate()
        {
            updateTimer = 0f;
        }

        private void OnDestroy()
        {
            // イベント解除
            EventBus.Unsubscribe<InventoryChangedEvent>(OnInventoryChanged);
            EventBus.Unsubscribe<PriceChangedEvent>(OnPriceChanged);
            EventBus.Unsubscribe<TimeAdvancedEvent>(OnTimeAdvanced);
            EventBus.Unsubscribe<TransactionCompletedEvent>(OnTransactionCompleted);

            // ボタンイベント解除
            if (restockButton != null)
                restockButton.onClick.RemoveListener(OnRestockPressed);

            if (marketButton != null)
                marketButton.onClick.RemoveListener(OnMarketPressed);

            if (inventoryButton != null)
                inventoryButton.onClick.RemoveListener(OnInventoryPressed);

            if (journalButton != null)
                journalButton.onClick.RemoveListener(OnJournalPressed);

            if (settingsButton != null)
                settingsButton.onClick.RemoveListener(OnSettingsPressed);
        }
    }

    /// <summary>
    /// ショップアイテム表示用のUIコンポーネント
    /// </summary>
    public class ShopItemUI : MonoBehaviour
    {
        [Header("Item Display")]
        [SerializeField]
        private Image itemIcon;

        [SerializeField]
        private Text itemNameText;

        [SerializeField]
        private Text quantityText;

        [SerializeField]
        private Text priceText;

        [SerializeField]
        private GameObject outOfStockOverlay;

        public ItemType ItemType { get; private set; }

        public void Setup(ItemType itemType, int quantity, float price)
        {
            ItemType = itemType;

            if (itemNameText != null)
                itemNameText.text = GetItemDisplayName(itemType);

            if (quantityText != null)
                quantityText.text = $"×{quantity}";

            if (priceText != null)
                priceText.text = $"{price:F0}G";

            SetOutOfStock(quantity == 0);
        }

        public void UpdatePrice(float newPrice)
        {
            if (priceText != null)
                priceText.text = $"{newPrice:F0}G";
        }

        public void SetOutOfStock(bool outOfStock)
        {
            if (outOfStockOverlay != null)
                outOfStockOverlay.SetActive(outOfStock);
        }

        private string GetItemDisplayName(ItemType itemType)
        {
            return itemType switch
            {
                ItemType.Fruit => "くだもの",
                ItemType.Potion => "ポーション",
                ItemType.Weapon => "武器",
                ItemType.Accessory => "アクセサリー",
                ItemType.MagicBook => "魔法書",
                ItemType.Gem => "宝石",
                _ => "不明",
            };
        }
    }
}
