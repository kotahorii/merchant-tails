using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using MerchantTails.Core;
using MerchantTails.Data;
using MerchantTails.Events;
using MerchantTails.Inventory;
using MerchantTails.Market;

namespace MerchantTails.UI
{
    /// <summary>
    /// インベントリ画面のUI制御
    /// 在庫管理、アイテム移動、仕入れ機能を提供
    /// </summary>
    public class InventoryPanel : UIPanel
    {
        [Header("Inventory Display")]
        [SerializeField] private Text inventoryTitleText;
        [SerializeField] private Text totalMoneyText;
        
        [Header("Storage Sections")]
        [SerializeField] private Transform storefrontContainer;
        [SerializeField] private Transform tradingContainer;
        [SerializeField] private GameObject inventoryItemPrefab;
        
        [Header("Capacity Display")]
        [SerializeField] private Text storefrontCapacityText;
        [SerializeField] private Text tradingCapacityText;
        [SerializeField] private Slider storefrontCapacityBar;
        [SerializeField] private Slider tradingCapacityBar;
        
        [Header("Item Transfer")]
        [SerializeField] private Button transferToStorefrontButton;
        [SerializeField] private Button transferToTradingButton;
        [SerializeField] private InputField transferQuantityInput;
        [SerializeField] private Text selectedItemInfoText;
        
        [Header("Purchase Section")]
        [SerializeField] private GameObject purchasePanel;
        [SerializeField] private Transform purchaseItemsContainer;
        [SerializeField] private Button confirmPurchaseButton;
        [SerializeField] private Text totalPurchaseCostText;
        
        [Header("Quick Actions")]
        [SerializeField] private Button shopButton;
        [SerializeField] private Button marketButton;
        [SerializeField] private Button sortButton;
        [SerializeField] private Button filterButton;
        
        [Header("Filter Options")]
        [SerializeField] private GameObject filterPanel;
        [SerializeField] private Toggle[] itemTypeToggles;
        [SerializeField] private Toggle showExpiredToggle;
        
        private Dictionary<ItemType, InventoryItemUI> storefrontItems = new Dictionary<ItemType, InventoryItemUI>();
        private Dictionary<ItemType, InventoryItemUI> tradingItems = new Dictionary<ItemType, InventoryItemUI>();
        private Dictionary<ItemType, int> purchaseCart = new Dictionary<ItemType, int>();
        
        private ItemType? selectedItemType = null;
        private InventoryLocation selectedLocation = InventoryLocation.Storefront;
        private List<ItemType> activeFilters = new List<ItemType>();
        
        protected override void OnInitialize()
        {
            SetupButtons();
            SetupEventListeners();
            InitializeFilters();
        }
        
        protected override void OnShow()
        {
            LogUIAction("Inventory panel shown");
            RefreshInventoryDisplay();
        }
        
        protected override void OnHide()
        {
            LogUIAction("Inventory panel hidden");
            ClearPurchaseCart();
        }
        
        private void SetupButtons()
        {
            if (transferToStorefrontButton != null)
                transferToStorefrontButton.onClick.AddListener(() => OnTransferPressed(InventoryLocation.Storefront));
            
            if (transferToTradingButton != null)
                transferToTradingButton.onClick.AddListener(() => OnTransferPressed(InventoryLocation.Trading));
            
            if (confirmPurchaseButton != null)
                confirmPurchaseButton.onClick.AddListener(OnConfirmPurchasePressed);
            
            if (shopButton != null)
                shopButton.onClick.AddListener(OnShopPressed);
            
            if (marketButton != null)
                marketButton.onClick.AddListener(OnMarketPressed);
            
            if (sortButton != null)
                sortButton.onClick.AddListener(OnSortPressed);
            
            if (filterButton != null)
                filterButton.onClick.AddListener(OnFilterPressed);
        }
        
        private void SetupEventListeners()
        {
            EventBus.Subscribe<InventoryChangedEvent>(OnInventoryChanged);
            EventBus.Subscribe<ItemDecayedEvent>(OnItemDecayed);
            EventBus.Subscribe<TransactionCompletedEvent>(OnTransactionCompleted);
        }
        
        private void InitializeFilters()
        {
            // 全アイテムタイプをデフォルトで表示
            foreach (ItemType itemType in System.Enum.GetValues(typeof(ItemType)))
            {
                activeFilters.Add(itemType);
            }
            
            // フィルタートグルの設定
            if (itemTypeToggles != null)
            {
                for (int i = 0; i < itemTypeToggles.Length && i < 6; i++)
                {
                    int index = i;
                    itemTypeToggles[i].onValueChanged.AddListener((isOn) => OnFilterToggleChanged((ItemType)index, isOn));
                }
            }
        }
        
        private void RefreshInventoryDisplay()
        {
            UpdateMoneyDisplay();
            UpdateCapacityDisplay();
            UpdateInventoryItems();
            UpdatePurchaseSection();
        }
        
        private void UpdateMoneyDisplay()
        {
            var playerData = GameManager.Instance?.GetPlayerData();
            if (playerData != null && totalMoneyText != null)
            {
                totalMoneyText.text = $"所持金: {playerData.CurrentMoney:N0}G";
            }
        }
        
        private void UpdateCapacityDisplay()
        {
            if (InventorySystem.Instance == null) return;
            
            // 店頭容量
            int storefrontUsed = InventorySystem.Instance.StorefrontCapacityUsed;
            int storefrontTotal = InventorySystem.Instance.StorefrontCapacityTotal;
            
            if (storefrontCapacityText != null)
                storefrontCapacityText.text = $"店頭: {storefrontUsed}/{storefrontTotal}";
            
            if (storefrontCapacityBar != null)
                storefrontCapacityBar.value = (float)storefrontUsed / storefrontTotal;
            
            // 取引用容量
            int tradingUsed = InventorySystem.Instance.TradingCapacityUsed;
            int tradingTotal = InventorySystem.Instance.TradingCapacityTotal;
            
            if (tradingCapacityText != null)
                tradingCapacityText.text = $"倉庫: {tradingUsed}/{tradingTotal}";
            
            if (tradingCapacityBar != null)
                tradingCapacityBar.value = (float)tradingUsed / tradingTotal;
        }
        
        private void UpdateInventoryItems()
        {
            ClearInventoryUI();
            
            if (InventorySystem.Instance == null) return;
            
            // 店頭在庫の表示
            var storefrontData = InventorySystem.Instance.GetStorefrontItems();
            foreach (var item in storefrontData)
            {
                if (IsItemFiltered(item.Key)) continue;
                CreateInventoryItemUI(item.Key, item.Value, InventoryLocation.Storefront);
            }
            
            // 取引用在庫の表示
            var tradingData = InventorySystem.Instance.GetTradingItems();
            foreach (var item in tradingData)
            {
                if (IsItemFiltered(item.Key)) continue;
                CreateInventoryItemUI(item.Key, item.Value, InventoryLocation.Trading);
            }
        }
        
        private void CreateInventoryItemUI(ItemType itemType, int quantity, InventoryLocation location)
        {
            if (inventoryItemPrefab == null) return;
            
            Transform container = location == InventoryLocation.Storefront ? storefrontContainer : tradingContainer;
            if (container == null) return;
            
            var itemGO = Instantiate(inventoryItemPrefab, container);
            var itemUI = itemGO.GetComponent<InventoryItemUI>();
            
            if (itemUI == null)
                itemUI = itemGO.AddComponent<InventoryItemUI>();
            
            // アイテム情報を設定
            float currentPrice = MarketSystem.Instance?.GetCurrentPrice(itemType) ?? 0f;
            int? expiryDays = InventorySystem.Instance.GetItemExpiryDays(itemType, location);
            
            itemUI.Setup(itemType, quantity, currentPrice, location, expiryDays);
            itemUI.OnItemSelected += OnInventoryItemSelected;
            
            // 辞書に追加
            if (location == InventoryLocation.Storefront)
                storefrontItems[itemType] = itemUI;
            else
                tradingItems[itemType] = itemUI;
        }
        
        private void ClearInventoryUI()
        {
            // 店頭アイテムUIをクリア
            foreach (var itemUI in storefrontItems.Values)
            {
                if (itemUI != null)
                {
                    itemUI.OnItemSelected -= OnInventoryItemSelected;
                    Destroy(itemUI.gameObject);
                }
            }
            storefrontItems.Clear();
            
            // 取引用アイテムUIをクリア
            foreach (var itemUI in tradingItems.Values)
            {
                if (itemUI != null)
                {
                    itemUI.OnItemSelected -= OnInventoryItemSelected;
                    Destroy(itemUI.gameObject);
                }
            }
            tradingItems.Clear();
        }
        
        private void UpdatePurchaseSection()
        {
            if (purchasePanel == null) return;
            
            // 仕入れ可能なアイテムを表示
            UpdatePurchaseItems();
            UpdatePurchaseCost();
        }
        
        private void UpdatePurchaseItems()
        {
            if (purchaseItemsContainer == null || inventoryItemPrefab == null) return;
            
            // 既存の購入アイテムUIをクリア
            foreach (Transform child in purchaseItemsContainer)
            {
                Destroy(child.gameObject);
            }
            
            // 各アイテムタイプの購入UIを作成
            foreach (ItemType itemType in System.Enum.GetValues(typeof(ItemType)))
            {
                if (IsItemFiltered(itemType)) continue;
                
                var purchaseItemGO = Instantiate(inventoryItemPrefab, purchaseItemsContainer);
                var purchaseItemUI = purchaseItemGO.GetComponent<PurchaseItemUI>();
                
                if (purchaseItemUI == null)
                    purchaseItemUI = purchaseItemGO.AddComponent<PurchaseItemUI>();
                
                float purchasePrice = MarketSystem.Instance?.GetCurrentPrice(itemType) ?? 100f;
                purchaseItemUI.Setup(itemType, purchasePrice, 0);
                purchaseItemUI.OnQuantityChanged += OnPurchaseQuantityChanged;
            }
        }
        
        private void UpdatePurchaseCost()
        {
            float totalCost = 0f;
            
            foreach (var item in purchaseCart)
            {
                float price = MarketSystem.Instance?.GetCurrentPrice(item.Key) ?? 0f;
                totalCost += price * item.Value;
            }
            
            if (totalPurchaseCostText != null)
                totalPurchaseCostText.text = $"合計: {totalCost:N0}G";
            
            // 購入ボタンの有効/無効
            if (confirmPurchaseButton != null)
            {
                var playerData = GameManager.Instance?.GetPlayerData();
                bool canAfford = playerData != null && playerData.CurrentMoney >= totalCost;
                bool hasItems = purchaseCart.Count > 0 && totalCost > 0;
                
                confirmPurchaseButton.interactable = canAfford && hasItems;
            }
        }
        
        private bool IsItemFiltered(ItemType itemType)
        {
            return !activeFilters.Contains(itemType);
        }
        
        // UIイベントハンドラー
        private void OnInventoryItemSelected(ItemType itemType, InventoryLocation location)
        {
            selectedItemType = itemType;
            selectedLocation = location;
            
            UpdateSelectedItemInfo();
            UpdateTransferButtons();
        }
        
        private void UpdateSelectedItemInfo()
        {
            if (selectedItemInfoText == null || !selectedItemType.HasValue) return;
            
            var itemType = selectedItemType.Value;
            int quantity = InventorySystem.Instance?.GetItemCount(itemType, selectedLocation) ?? 0;
            float price = MarketSystem.Instance?.GetCurrentPrice(itemType) ?? 0f;
            
            selectedItemInfoText.text = $"{GetItemDisplayName(itemType)} - 数量: {quantity} - 価格: {price:F0}G";
        }
        
        private void UpdateTransferButtons()
        {
            bool hasSelection = selectedItemType.HasValue;
            
            if (transferToStorefrontButton != null)
                transferToStorefrontButton.interactable = hasSelection && selectedLocation == InventoryLocation.Trading;
            
            if (transferToTradingButton != null)
                transferToTradingButton.interactable = hasSelection && selectedLocation == InventoryLocation.Storefront;
        }
        
        private void OnTransferPressed(InventoryLocation targetLocation)
        {
            if (!selectedItemType.HasValue) return;
            
            int quantity = 1;
            if (transferQuantityInput != null && int.TryParse(transferQuantityInput.text, out int inputQuantity))
            {
                quantity = Mathf.Max(1, inputQuantity);
            }
            
            // アイテム移動を実行
            bool success = InventorySystem.Instance?.MoveItem(
                selectedItemType.Value,
                quantity,
                selectedLocation,
                targetLocation
            ) ?? false;
            
            if (success)
            {
                LogUIAction($"Transferred {quantity} {selectedItemType.Value} to {targetLocation}");
                RefreshInventoryDisplay();
            }
            else
            {
                ShowTransferError();
            }
        }
        
        private void OnPurchaseQuantityChanged(ItemType itemType, int quantity)
        {
            if (quantity > 0)
                purchaseCart[itemType] = quantity;
            else
                purchaseCart.Remove(itemType);
            
            UpdatePurchaseCost();
        }
        
        private void OnConfirmPurchasePressed()
        {
            LogUIAction("Confirm purchase pressed");
            ExecutePurchase();
        }
        
        private void ExecutePurchase()
        {
            if (purchaseCart.Count == 0) return;
            
            var playerData = GameManager.Instance?.GetPlayerData();
            if (playerData == null) return;
            
            float totalCost = 0f;
            var purchaseList = new List<(ItemType, int)>();
            
            // 購入リストを作成
            foreach (var item in purchaseCart)
            {
                float price = MarketSystem.Instance?.GetCurrentPrice(item.Key) ?? 0f;
                totalCost += price * item.Value;
                purchaseList.Add((item.Key, item.Value));
            }
            
            // 資金確認
            if (playerData.CurrentMoney < totalCost)
            {
                ShowInsufficientFundsError();
                return;
            }
            
            // 容量確認
            int totalItems = 0;
            foreach (var item in purchaseList)
            {
                totalItems += item.Item2;
            }
            
            if (InventorySystem.Instance.TradingCapacityRemaining < totalItems)
            {
                ShowInsufficientCapacityError();
                return;
            }
            
            // 購入実行
            foreach (var (itemType, quantity) in purchaseList)
            {
                InventorySystem.Instance.AddItem(itemType, quantity, InventoryLocation.Trading);
            }
            
            // 支払い
            playerData.SpendMoney(Mathf.RoundToInt(totalCost));
            
            // 購入完了イベント
            var purchaseEvent = new PurchaseCompletedEvent(purchaseList, totalCost);
            EventBus.Publish(purchaseEvent);
            
            // UIリフレッシュ
            ClearPurchaseCart();
            RefreshInventoryDisplay();
            
            ShowPurchaseSuccess(totalCost);
        }
        
        private void ClearPurchaseCart()
        {
            purchaseCart.Clear();
            UpdatePurchaseCost();
        }
        
        // ボタンイベントハンドラー
        private void OnShopPressed()
        {
            LogUIAction("Shop button pressed");
            UIManager.Instance.ShowPanel(UIType.ShopManagement);
        }
        
        private void OnMarketPressed()
        {
            LogUIAction("Market button pressed");
            UIManager.Instance.ShowPanel(UIType.MarketAnalysis);
        }
        
        private void OnSortPressed()
        {
            LogUIAction("Sort button pressed");
            // ソート機能の実装（将来実装）
        }
        
        private void OnFilterPressed()
        {
            LogUIAction("Filter button pressed");
            if (filterPanel != null)
                filterPanel.SetActive(!filterPanel.activeSelf);
        }
        
        private void OnFilterToggleChanged(ItemType itemType, bool isOn)
        {
            if (isOn && !activeFilters.Contains(itemType))
            {
                activeFilters.Add(itemType);
            }
            else if (!isOn && activeFilters.Contains(itemType))
            {
                activeFilters.Remove(itemType);
            }
            
            UpdateInventoryItems();
            UpdatePurchaseItems();
        }
        
        // エラー/成功メッセージ
        private void ShowTransferError()
        {
            // 移動エラーの表示（簡易実装）
            ErrorHandler.LogWarning("Item transfer failed", "InventoryPanel");
        }
        
        private void ShowInsufficientFundsError()
        {
            // 資金不足エラーの表示
            ErrorHandler.LogWarning("Insufficient funds for purchase", "InventoryPanel");
        }
        
        private void ShowInsufficientCapacityError()
        {
            // 容量不足エラーの表示
            ErrorHandler.LogWarning("Insufficient inventory capacity", "InventoryPanel");
        }
        
        private void ShowPurchaseSuccess(float totalCost)
        {
            // 購入成功メッセージの表示
            ErrorHandler.LogInfo($"Purchase completed: {totalCost:N0}G", "InventoryPanel");
        }
        
        // ヘルパーメソッド
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
                _ => "不明"
            };
        }
        
        // イベントハンドラー
        private void OnInventoryChanged(InventoryChangedEvent evt)
        {
            RefreshInventoryDisplay();
        }
        
        private void OnItemDecayed(ItemDecayedEvent evt)
        {
            RefreshInventoryDisplay();
            // 腐敗通知の表示
        }
        
        private void OnTransactionCompleted(TransactionCompletedEvent evt)
        {
            UpdateMoneyDisplay();
        }
        
        private void OnDestroy()
        {
            // イベント解除
            EventBus.Unsubscribe<InventoryChangedEvent>(OnInventoryChanged);
            EventBus.Unsubscribe<ItemDecayedEvent>(OnItemDecayed);
            EventBus.Unsubscribe<TransactionCompletedEvent>(OnTransactionCompleted);
            
            // ボタンイベント解除
            if (transferToStorefrontButton != null)
                transferToStorefrontButton.onClick.RemoveAllListeners();
            
            if (transferToTradingButton != null)
                transferToTradingButton.onClick.RemoveAllListeners();
            
            if (confirmPurchaseButton != null)
                confirmPurchaseButton.onClick.RemoveListener(OnConfirmPurchasePressed);
            
            if (shopButton != null)
                shopButton.onClick.RemoveListener(OnShopPressed);
            
            if (marketButton != null)
                marketButton.onClick.RemoveListener(OnMarketPressed);
            
            if (sortButton != null)
                sortButton.onClick.RemoveListener(OnSortPressed);
            
            if (filterButton != null)
                filterButton.onClick.RemoveListener(OnFilterPressed);
            
            ClearInventoryUI();
        }
    }
    
    /// <summary>
    /// インベントリアイテム表示用のUIコンポーネント
    /// </summary>
    public class InventoryItemUI : MonoBehaviour
    {
        [Header("Item Display")]
        [SerializeField] private Image itemIcon;
        [SerializeField] private Text itemNameText;
        [SerializeField] private Text quantityText;
        [SerializeField] private Text priceText;
        [SerializeField] private Text expiryText;
        [SerializeField] private Button selectButton;
        [SerializeField] private Image selectionHighlight;
        
        public ItemType ItemType { get; private set; }
        public InventoryLocation Location { get; private set; }
        
        public event System.Action<ItemType, InventoryLocation> OnItemSelected;
        
        public void Setup(ItemType itemType, int quantity, float price, InventoryLocation location, int? expiryDays)
        {
            ItemType = itemType;
            Location = location;
            
            if (itemNameText != null)
                itemNameText.text = GetItemDisplayName(itemType);
            
            if (quantityText != null)
                quantityText.text = $"×{quantity}";
            
            if (priceText != null)
                priceText.text = $"{price:F0}G";
            
            if (expiryText != null && expiryDays.HasValue)
            {
                expiryText.text = $"期限: {expiryDays}日";
                expiryText.color = expiryDays.Value <= 3 ? Color.red : Color.white;
            }
            
            if (selectButton != null)
                selectButton.onClick.AddListener(OnSelectPressed);
        }
        
        private void OnSelectPressed()
        {
            OnItemSelected?.Invoke(ItemType, Location);
            SetSelected(true);
        }
        
        public void SetSelected(bool selected)
        {
            if (selectionHighlight != null)
                selectionHighlight.gameObject.SetActive(selected);
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
                _ => "不明"
            };
        }
        
        private void OnDestroy()
        {
            if (selectButton != null)
                selectButton.onClick.RemoveListener(OnSelectPressed);
        }
    }
    
    /// <summary>
    /// 購入アイテム表示用のUIコンポーネント
    /// </summary>
    public class PurchaseItemUI : MonoBehaviour
    {
        [Header("Purchase Display")]
        [SerializeField] private Image itemIcon;
        [SerializeField] private Text itemNameText;
        [SerializeField] private Text priceText;
        [SerializeField] private InputField quantityInput;
        [SerializeField] private Text totalCostText;
        
        public ItemType ItemType { get; private set; }
        private float unitPrice;
        
        public event System.Action<ItemType, int> OnQuantityChanged;
        
        public void Setup(ItemType itemType, float price, int initialQuantity)
        {
            ItemType = itemType;
            unitPrice = price;
            
            if (itemNameText != null)
                itemNameText.text = GetItemDisplayName(itemType);
            
            if (priceText != null)
                priceText.text = $"{price:F0}G";
            
            if (quantityInput != null)
            {
                quantityInput.text = initialQuantity.ToString();
                quantityInput.onValueChanged.AddListener(OnQuantityInputChanged);
            }
            
            UpdateTotalCost(initialQuantity);
        }
        
        private void OnQuantityInputChanged(string value)
        {
            if (int.TryParse(value, out int quantity))
            {
                quantity = Mathf.Max(0, quantity);
                UpdateTotalCost(quantity);
                OnQuantityChanged?.Invoke(ItemType, quantity);
            }
        }
        
        private void UpdateTotalCost(int quantity)
        {
            if (totalCostText != null)
            {
                float total = unitPrice * quantity;
                totalCostText.text = $"{total:F0}G";
            }
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
                _ => "不明"
            };
        }
        
        private void OnDestroy()
        {
            if (quantityInput != null)
                quantityInput.onValueChanged.RemoveListener(OnQuantityInputChanged);
        }
    }
}