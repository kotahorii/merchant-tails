using System;
using System.Collections.Generic;
using System.Linq;
using MerchantTails.Core;
using MerchantTails.Data;
using UnityEngine;

namespace MerchantTails.Inventory
{
    /// <summary>
    /// 在庫管理システム
    /// 店頭販売用と相場取引用の二重在庫を管理し、商品の劣化・移動を処理
    /// </summary>
    public class InventorySystem : MonoBehaviour
    {
        [Header("Inventory Configuration")]
        [SerializeField]
        private int maxStorefrontCapacity = 50;

        [SerializeField]
        private int maxTradingCapacity = 100;

        [SerializeField]
        private bool enableItemDecay = true;

        [Header("Decay Settings")]
        [SerializeField]
        private int fruitDecayDays = 3;

        [SerializeField]
        private int potionDecayDays = 30;

        [SerializeField]
        private int weaponDecayDays = -1; // No decay

        [SerializeField]
        private int accessoryDecayDays = -1; // No decay

        [SerializeField]
        private int magicBookDecayDays = -1; // No decay

        [SerializeField]
        private int gemDecayDays = -1; // No decay

        // Inventory storage
        private Dictionary<ItemType, List<InventoryItem>> storefrontInventory;
        private Dictionary<ItemType, List<InventoryItem>> tradingInventory;

        public static InventorySystem Instance { get; private set; }

        // Events
        public event Action<ItemType, int, InventoryLocation> OnInventoryChanged;
        public event Action<InventoryItem> OnItemDecayed;
        public event Action<InventoryItem, InventoryLocation, InventoryLocation> OnItemMoved;

        // Properties
        public int StorefrontCapacityUsed => storefrontInventory.Values.Sum(list => list.Count);
        public int TradingCapacityUsed => tradingInventory.Values.Sum(list => list.Count);
        public int StorefrontCapacityRemaining => maxStorefrontCapacity - StorefrontCapacityUsed;
        public int TradingCapacityRemaining => maxTradingCapacity - TradingCapacityUsed;
        public int StorefrontCapacityTotal => maxStorefrontCapacity;
        public int TradingCapacityTotal => maxTradingCapacity;

        private void Awake()
        {
            if (Instance == null)
            {
                Instance = this;
                DontDestroyOnLoad(gameObject);
                InitializeInventorySystem();
            }
            else
            {
                Destroy(gameObject);
            }
        }

        private void Start()
        {
            SubscribeToEvents();
        }

        private void InitializeInventorySystem()
        {
            Debug.Log("[InventorySystem] Initializing inventory system...");

            storefrontInventory = new Dictionary<ItemType, List<InventoryItem>>();
            tradingInventory = new Dictionary<ItemType, List<InventoryItem>>();

            // Initialize storage for all item types
            foreach (ItemType itemType in Enum.GetValues(typeof(ItemType)))
            {
                if (itemType == ItemType.None)
                    continue;

                storefrontInventory[itemType] = new List<InventoryItem>();
                tradingInventory[itemType] = new List<InventoryItem>();
            }

            Debug.Log("[InventorySystem] Inventory system initialized");
        }

        private void SubscribeToEvents()
        {
            EventBus.Subscribe<DayChangedEvent>(OnDayChanged);
            EventBus.Subscribe<TransactionCompletedEvent>(OnTransactionCompleted);
        }

        private void OnDestroy()
        {
            EventBus.Unsubscribe<DayChangedEvent>(OnDayChanged);
            EventBus.Unsubscribe<TransactionCompletedEvent>(OnTransactionCompleted);
        }

        // Event handlers
        private void OnDayChanged(DayChangedEvent evt)
        {
            if (enableItemDecay)
            {
                ProcessItemDecay(evt.NewDay);
            }
        }

        private void OnTransactionCompleted(TransactionCompletedEvent evt)
        {
            if (evt.IsPurchase)
            {
                // Add purchased items to trading inventory by default
                AddItem(evt.ItemType, evt.Quantity, InventoryLocation.Trading);
            }
            else
            {
                // Remove sold items (should already be handled by transaction system)
                Debug.Log($"[InventorySystem] Items sold: {evt.Quantity}x {evt.ItemType}");
            }
        }

        // Public API methods
        public bool AddItem(ItemType itemType, int quantity, InventoryLocation location)
        {
            if (quantity <= 0)
                return false;

            var targetInventory = GetInventoryByLocation(location);
            int capacity = GetCapacityByLocation(location);
            int currentUsed = GetUsedCapacityByLocation(location);

            if (currentUsed + quantity > capacity)
            {
                Debug.LogWarning(
                    $"[InventorySystem] Not enough capacity in {location}. "
                        + $"Required: {quantity}, Available: {capacity - currentUsed}"
                );
                return false;
            }

            // Create inventory items
            int currentDay = TimeManager.Instance?.CurrentDay ?? 1;
            for (int i = 0; i < quantity; i++)
            {
                var item = new InventoryItem
                {
                    itemType = itemType,
                    quality = ItemQuality.Common, // Default quality
                    purchaseDay = currentDay,
                    expiryDay = CalculateExpiryDay(itemType, currentDay),
                    purchasePrice = 0f, // Will be set by transaction system
                    location = location,
                    uniqueId = System.Guid.NewGuid().ToString(),
                };

                targetInventory[itemType].Add(item);
            }

            TriggerInventoryChangedEvent(itemType, quantity, location);
            Debug.Log($"[InventorySystem] Added {quantity}x {itemType} to {location}");
            return true;
        }

        public bool RemoveItem(ItemType itemType, int quantity, InventoryLocation location)
        {
            if (quantity <= 0)
                return false;

            var targetInventory = GetInventoryByLocation(location);
            var items = targetInventory[itemType];

            if (items.Count < quantity)
            {
                Debug.LogWarning(
                    $"[InventorySystem] Not enough {itemType} in {location}. "
                        + $"Requested: {quantity}, Available: {items.Count}"
                );
                return false;
            }

            // Remove items (FIFO - First In, First Out)
            for (int i = 0; i < quantity; i++)
            {
                if (items.Count > 0)
                {
                    items.RemoveAt(0);
                }
            }

            TriggerInventoryChangedEvent(itemType, -quantity, location);
            Debug.Log($"[InventorySystem] Removed {quantity}x {itemType} from {location}");
            return true;
        }

        public bool MoveItem(ItemType itemType, int quantity, InventoryLocation from, InventoryLocation to)
        {
            if (quantity <= 0)
                return false;

            var fromInventory = GetInventoryByLocation(from);
            var toInventory = GetInventoryByLocation(to);

            // Check availability in source
            if (GetItemCount(itemType, from) < quantity)
            {
                Debug.LogWarning(
                    $"[InventorySystem] Not enough {itemType} in {from} to move. "
                        + $"Requested: {quantity}, Available: {GetItemCount(itemType, from)}"
                );
                return false;
            }

            // Check capacity in destination
            int toCapacity = GetCapacityByLocation(to);
            int toUsed = GetUsedCapacityByLocation(to);

            if (toUsed + quantity > toCapacity)
            {
                Debug.LogWarning(
                    $"[InventorySystem] Not enough capacity in {to}. "
                        + $"Required: {quantity}, Available: {toCapacity - toUsed}"
                );
                return false;
            }

            // Move items
            var itemsToMove = fromInventory[itemType].Take(quantity).ToList();
            foreach (var item in itemsToMove)
            {
                fromInventory[itemType].Remove(item);
                item.location = to;
                toInventory[itemType].Add(item);

                OnItemMoved?.Invoke(item, from, to);
            }

            TriggerInventoryChangedEvent(itemType, -quantity, from);
            TriggerInventoryChangedEvent(itemType, quantity, to);

            Debug.Log($"[InventorySystem] Moved {quantity}x {itemType} from {from} to {to}");
            return true;
        }

        public int GetItemCount(ItemType itemType, InventoryLocation location)
        {
            var inventory = GetInventoryByLocation(location);
            return inventory[itemType].Count;
        }

        public int GetTotalItemCount(ItemType itemType)
        {
            return GetItemCount(itemType, InventoryLocation.Storefront)
                + GetItemCount(itemType, InventoryLocation.Trading);
        }

        public List<InventoryItem> GetItems(ItemType itemType, InventoryLocation location)
        {
            var inventory = GetInventoryByLocation(location);
            return new List<InventoryItem>(inventory[itemType]);
        }

        public List<InventoryItem> GetAllItems(InventoryLocation location)
        {
            var inventory = GetInventoryByLocation(location);
            var allItems = new List<InventoryItem>();

            foreach (var itemList in inventory.Values)
            {
                allItems.AddRange(itemList);
            }

            return allItems;
        }

        public List<InventoryItem> GetExpiringItems(int daysFromNow = 1)
        {
            int currentDay = TimeManager.Instance?.CurrentDay ?? 1;
            int checkDay = currentDay + daysFromNow;
            var expiringItems = new List<InventoryItem>();

            foreach (var inventory in new[] { storefrontInventory, tradingInventory })
            {
                foreach (var itemList in inventory.Values)
                {
                    expiringItems.AddRange(itemList.Where(item => item.expiryDay > 0 && item.expiryDay <= checkDay));
                }
            }

            return expiringItems;
        }

        // Helper methods for UI
        public Dictionary<ItemType, int> GetStorefrontItems()
        {
            var result = new Dictionary<ItemType, int>();
            foreach (var kvp in storefrontInventory)
            {
                result[kvp.Key] = kvp.Value.Count;
            }
            return result;
        }

        public Dictionary<ItemType, int> GetTradingItems()
        {
            var result = new Dictionary<ItemType, int>();
            foreach (var kvp in tradingInventory)
            {
                result[kvp.Key] = kvp.Value.Count;
            }
            return result;
        }

        public int GetItemExpiryDays(ItemType itemType)
        {
            return GetDecayDays(itemType);
        }

        // Private helper methods
        private Dictionary<ItemType, List<InventoryItem>> GetInventoryByLocation(InventoryLocation location)
        {
            return location switch
            {
                InventoryLocation.Storefront => storefrontInventory,
                InventoryLocation.Trading => tradingInventory,
                _ => storefrontInventory,
            };
        }

        private int GetCapacityByLocation(InventoryLocation location)
        {
            return location switch
            {
                InventoryLocation.Storefront => maxStorefrontCapacity,
                InventoryLocation.Trading => maxTradingCapacity,
                _ => maxStorefrontCapacity,
            };
        }

        private int GetUsedCapacityByLocation(InventoryLocation location)
        {
            return location switch
            {
                InventoryLocation.Storefront => StorefrontCapacityUsed,
                InventoryLocation.Trading => TradingCapacityUsed,
                _ => StorefrontCapacityUsed,
            };
        }

        private int CalculateExpiryDay(ItemType itemType, int purchaseDay)
        {
            int decayDays = GetDecayDays(itemType);
            return decayDays > 0 ? purchaseDay + decayDays : -1; // -1 means no expiry
        }

        private int GetDecayDays(ItemType itemType)
        {
            return itemType switch
            {
                ItemType.Fruit => fruitDecayDays,
                ItemType.Potion => potionDecayDays,
                ItemType.Weapon => weaponDecayDays,
                ItemType.Accessory => accessoryDecayDays,
                ItemType.MagicBook => magicBookDecayDays,
                ItemType.Gem => gemDecayDays,
                _ => -1,
            };
        }

        private void ProcessItemDecay(int currentDay)
        {
            var expiredItems = new List<(InventoryItem item, InventoryLocation location)>();

            // Check storefront inventory
            CheckInventoryForDecay(storefrontInventory, currentDay, InventoryLocation.Storefront, expiredItems);

            // Check trading inventory
            CheckInventoryForDecay(tradingInventory, currentDay, InventoryLocation.Trading, expiredItems);

            // Remove expired items
            foreach (var (item, location) in expiredItems)
            {
                var inventory = GetInventoryByLocation(location);
                inventory[item.itemType].Remove(item);

                OnItemDecayed?.Invoke(item);
                TriggerInventoryChangedEvent(item.itemType, -1, location);

                Debug.Log($"[InventorySystem] Item expired: {item.itemType} (Day {item.expiryDay})");
            }

            if (expiredItems.Count > 0)
            {
                Debug.Log($"[InventorySystem] {expiredItems.Count} items expired on day {currentDay}");
            }
        }

        private void CheckInventoryForDecay(
            Dictionary<ItemType, List<InventoryItem>> inventory,
            int currentDay,
            InventoryLocation location,
            List<(InventoryItem, InventoryLocation)> expiredItems
        )
        {
            foreach (var itemList in inventory.Values)
            {
                var expired = itemList.Where(item => item.expiryDay > 0 && item.expiryDay <= currentDay).ToList();
                foreach (var item in expired)
                {
                    expiredItems.Add((item, location));
                }
            }
        }

        private void TriggerInventoryChangedEvent(ItemType itemType, int quantityChange, InventoryLocation location)
        {
            OnInventoryChanged?.Invoke(itemType, quantityChange, location);
        }

        // Utility methods
        public InventoryData GetInventoryData()
        {
            var data = new InventoryData
            {
                storefrontItems = new List<InventoryItem>(),
                tradingItems = new List<InventoryItem>(),
            };

            // Collect all items from both inventories
            foreach (var itemList in storefrontInventory.Values)
            {
                data.storefrontItems.AddRange(itemList);
            }

            foreach (var itemList in tradingInventory.Values)
            {
                data.tradingItems.AddRange(itemList);
            }

            return data;
        }

        public void LoadInventoryData(InventoryData data)
        {
            // Clear existing inventories
            foreach (var itemType in storefrontInventory.Keys.ToList())
            {
                storefrontInventory[itemType].Clear();
                tradingInventory[itemType].Clear();
            }

            // Load storefront items
            foreach (var item in data.storefrontItems)
            {
                storefrontInventory[item.itemType].Add(item);
            }

            // Load trading items
            foreach (var item in data.tradingItems)
            {
                tradingInventory[item.itemType].Add(item);
            }

            Debug.Log(
                $"[InventorySystem] Loaded {data.storefrontItems.Count} storefront items and {data.tradingItems.Count} trading items"
            );
        }

        public void LogInventoryState()
        {
            Debug.Log("[InventorySystem] Current Inventory State:");
            Debug.Log($"  Storefront Capacity: {StorefrontCapacityUsed}/{maxStorefrontCapacity}");
            Debug.Log($"  Trading Capacity: {TradingCapacityUsed}/{maxTradingCapacity}");

            foreach (ItemType itemType in Enum.GetValues(typeof(ItemType)))
            {
                int storefrontCount = GetItemCount(itemType, InventoryLocation.Storefront);
                int tradingCount = GetItemCount(itemType, InventoryLocation.Trading);

                if (storefrontCount > 0 || tradingCount > 0)
                {
                    Debug.Log($"  {itemType}: Storefront={storefrontCount}, Trading={tradingCount}");
                }
            }
        }

        // Additional methods for test compatibility

        /// <summary>
        /// AddItem overload that accepts a single InventoryItem (for test compatibility)
        /// </summary>
        public bool AddItem(InventoryItem item)
        {
            if (item == null)
                return false;

            var targetInventory = GetInventoryByLocation(item.location);
            int capacity = GetCapacityByLocation(item.location);
            int currentUsed = GetUsedCapacityByLocation(item.location);

            if (currentUsed + 1 > capacity)
            {
                Debug.LogWarning(
                    $"[InventorySystem] Not enough capacity in {item.location}. "
                        + $"Available: {capacity - currentUsed}"
                );
                return false;
            }

            // Set unique ID if not already set
            if (string.IsNullOrEmpty(item.uniqueId))
            {
                item.uniqueId = System.Guid.NewGuid().ToString();
            }

            // Set purchase day if not already set
            if (item.purchaseDay == 0)
            {
                item.purchaseDay = TimeManager.Instance?.CurrentDay ?? 1;
            }

            // Calculate expiry if not set
            if (item.expiryDay == 0)
            {
                item.expiryDay = CalculateExpiryDay(item.itemType, item.purchaseDay);
            }

            targetInventory[item.itemType].Add(item);
            TriggerInventoryChangedEvent(item.itemType, 1, item.location);
            Debug.Log($"[InventorySystem] Added {item.itemType} to {item.location}");
            return true;
        }

        /// <summary>
        /// RemoveItem overload that accepts item ID (for test compatibility)
        /// </summary>
        public bool RemoveItem(string itemId, int quantity = 1)
        {
            // Find the item across both inventories
            foreach (var inventory in new[] { storefrontInventory, tradingInventory })
            {
                foreach (var itemList in inventory.Values)
                {
                    for (int i = itemList.Count - 1; i >= 0 && quantity > 0; i--)
                    {
                        if (itemList[i].uniqueId == itemId)
                        {
                            var item = itemList[i];
                            itemList.RemoveAt(i);
                            quantity--;
                            TriggerInventoryChangedEvent(item.itemType, -1, item.location);

                            if (quantity == 0)
                                return true;
                        }
                    }
                }
            }

            return quantity == 0;
        }

        /// <summary>
        /// Get all items as a dictionary (for test compatibility)
        /// </summary>
        public Dictionary<ItemType, int> GetAllItems()
        {
            var result = new Dictionary<ItemType, int>();

            foreach (ItemType itemType in Enum.GetValues(typeof(ItemType)))
            {
                if (itemType == ItemType.None)
                    continue;

                int totalCount = GetTotalItemCount(itemType);
                if (totalCount > 0)
                {
                    result[itemType] = totalCount;
                }
            }

            return result;
        }

        /// <summary>
        /// Get total value of all items in inventory
        /// </summary>
        public float GetTotalValue()
        {
            float totalValue = 0f;
            int currentDay = TimeManager.Instance?.CurrentDay ?? 1;

            foreach (var inventory in new[] { storefrontInventory, tradingInventory })
            {
                foreach (var itemList in inventory.Values)
                {
                    foreach (var item in itemList)
                    {
                        totalValue +=
                            item.purchasePrice * item.GetQualityMultiplier() * item.GetFreshnessMultiplier(currentDay);
                    }
                }
            }

            return totalValue;
        }

        /// <summary>
        /// Transfer items from trading inventory to storefront
        /// </summary>
        public bool TransferToShop(ItemType itemType, int quantity)
        {
            return MoveItem(itemType, quantity, InventoryLocation.Trading, InventoryLocation.Storefront);
        }

        /// <summary>
        /// Transfer items from storefront to trading inventory
        /// </summary>
        public bool TransferFromShop(ItemType itemType, int quantity)
        {
            return MoveItem(itemType, quantity, InventoryLocation.Storefront, InventoryLocation.Trading);
        }

        /// <summary>
        /// Get items in shop (storefront) only
        /// </summary>
        public Dictionary<ItemType, int> GetShopItems()
        {
            return GetStorefrontItems();
        }

        /// <summary>
        /// Check if there's space for items
        /// </summary>
        public bool HasSpace(ItemType itemType, int quantity)
        {
            // Default to trading inventory for space check
            return TradingCapacityRemaining >= quantity;
        }

        /// <summary>
        /// Clear all items from both inventories
        /// </summary>
        public void ClearInventory()
        {
            foreach (var itemType in storefrontInventory.Keys.ToList())
            {
                int storefrontCount = storefrontInventory[itemType].Count;
                int tradingCount = tradingInventory[itemType].Count;

                storefrontInventory[itemType].Clear();
                tradingInventory[itemType].Clear();

                if (storefrontCount > 0)
                    TriggerInventoryChangedEvent(itemType, -storefrontCount, InventoryLocation.Storefront);
                if (tradingCount > 0)
                    TriggerInventoryChangedEvent(itemType, -tradingCount, InventoryLocation.Trading);
            }

            Debug.Log("[InventorySystem] Inventory cleared");
        }

        /// <summary>
        /// Get items by quality level
        /// </summary>
        public Dictionary<ItemType, int> GetItemsByQuality(ItemQuality quality)
        {
            var result = new Dictionary<ItemType, int>();

            foreach (var inventory in new[] { storefrontInventory, tradingInventory })
            {
                foreach (var kvp in inventory)
                {
                    int count = kvp.Value.Count(item => item.quality == quality);
                    if (count > 0)
                    {
                        result[kvp.Key] = result.GetValueOrDefault(kvp.Key, 0) + count;
                    }
                }
            }

            return result;
        }

        /// <summary>
        /// Update item conditions (for decay simulation)
        /// </summary>
        public void UpdateItemConditions()
        {
            int currentDay = TimeManager.Instance?.CurrentDay ?? 1;
            ProcessItemDecay(currentDay);
        }

        /// <summary>
        /// Get all items of a specific type
        /// </summary>
        public List<InventoryItem> GetItemsOfType(ItemType itemType)
        {
            var result = new List<InventoryItem>();
            result.AddRange(GetItems(itemType, InventoryLocation.Storefront));
            result.AddRange(GetItems(itemType, InventoryLocation.Trading));
            return result;
        }

        /// <summary>
        /// Get item count with optional location filter (backward compatibility)
        /// </summary>
        public int GetItemCount(ItemType itemType, bool isInShop = false)
        {
            return isInShop
                ? GetItemCount(itemType, InventoryLocation.Storefront)
                : GetItemCount(itemType, InventoryLocation.Trading);
        }
    }
}
