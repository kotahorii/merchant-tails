using System;
using System.Collections.Generic;
using UnityEngine;
using MerchantTails.Data;

namespace MerchantTails.Inventory
{
    /// <summary>
    /// 個別アイテムの在庫データ
    /// 購入情報、品質、劣化状況を管理
    /// </summary>
    [System.Serializable]
    public class InventoryItem
    {
        [Header("Basic Information")]
        public string uniqueId;
        public ItemType itemType;
        public ItemQuality quality;
        
        [Header("Purchase Information")]
        public int purchaseDay;
        public float purchasePrice;
        public InventoryLocation location;
        
        [Header("Condition")]
        public int expiryDay = -1; // -1 means no expiry
        
        /// <summary>
        /// アイテムが期限切れかどうかチェック
        /// </summary>
        public bool IsExpired(int currentDay)
        {
            return expiryDay > 0 && currentDay >= expiryDay;
        }
        
        /// <summary>
        /// 期限切れまでの残り日数を取得
        /// </summary>
        public int GetDaysUntilExpiry(int currentDay)
        {
            if (expiryDay <= 0) return -1; // No expiry
            return Math.Max(0, expiryDay - currentDay);
        }
        
        /// <summary>
        /// アイテムの新鮮度を取得（0.0 - 1.0）
        /// </summary>
        public float GetFreshness(int currentDay)
        {
            if (expiryDay <= 0) return 1.0f; // No decay
            
            int totalLifespan = expiryDay - purchaseDay;
            int remainingDays = expiryDay - currentDay;
            
            if (totalLifespan <= 0) return 0.0f;
            if (remainingDays <= 0) return 0.0f;
            
            return (float)remainingDays / totalLifespan;
        }
        
        /// <summary>
        /// 品質による価格修正係数を取得
        /// </summary>
        public float GetQualityMultiplier()
        {
            return quality switch
            {
                ItemQuality.Poor => 0.7f,
                ItemQuality.Common => 1.0f,
                ItemQuality.Good => 1.3f,
                ItemQuality.Excellent => 1.6f,
                _ => 1.0f
            };
        }
        
        /// <summary>
        /// 新鮮度による価格修正係数を取得
        /// </summary>
        public float GetFreshnessMultiplier(int currentDay)
        {
            float freshness = GetFreshness(currentDay);
            
            // 新鮮度が下がると価格も下がる
            return Mathf.Lerp(0.5f, 1.0f, freshness);
        }
        
        /// <summary>
        /// 現在の実効価格を計算
        /// </summary>
        public float GetEffectivePrice(int currentDay, float basePrice)
        {
            return basePrice * GetQualityMultiplier() * GetFreshnessMultiplier(currentDay);
        }
        
        /// <summary>
        /// アイテムの状態説明を取得
        /// </summary>
        public string GetConditionDescription(int currentDay)
        {
            if (IsExpired(currentDay))
                return "期限切れ";
            
            int daysLeft = GetDaysUntilExpiry(currentDay);
            if (daysLeft <= 0)
                return "保存期限なし";
            
            if (daysLeft == 1)
                return "明日期限切れ";
            
            return $"あと{daysLeft}日";
        }
    }
    
    /// <summary>
    /// 在庫の保存場所を示すenum
    /// </summary>
    public enum InventoryLocation
    {
        /// <summary>店頭販売用在庫</summary>
        Storefront,
        
        /// <summary>相場取引用在庫</summary>
        Trading
    }
    
    /// <summary>
    /// 在庫データの保存用クラス
    /// </summary>
    [System.Serializable]
    public class InventoryData
    {
        public List<InventoryItem> storefrontItems = new List<InventoryItem>();
        public List<InventoryItem> tradingItems = new List<InventoryItem>();
        
        /// <summary>
        /// 総アイテム数を取得
        /// </summary>
        public int GetTotalItemCount()
        {
            return storefrontItems.Count + tradingItems.Count;
        }
        
        /// <summary>
        /// 指定したアイテムタイプの総数を取得
        /// </summary>
        public int GetItemTypeCount(ItemType itemType)
        {
            int count = 0;
            count += storefrontItems.FindAll(item => item.itemType == itemType).Count;
            count += tradingItems.FindAll(item => item.itemType == itemType).Count;
            return count;
        }
        
        /// <summary>
        /// 期限切れアイテムの数を取得
        /// </summary>
        public int GetExpiredItemCount(int currentDay)
        {
            int count = 0;
            count += storefrontItems.FindAll(item => item.IsExpired(currentDay)).Count;
            count += tradingItems.FindAll(item => item.IsExpired(currentDay)).Count;
            return count;
        }
    }
    
    /// <summary>
    /// 在庫統計情報
    /// </summary>
    [System.Serializable]
    public class InventoryStats
    {
        public int totalItems;
        public int storefrontItems;
        public int tradingItems;
        public int expiredItems;
        public float totalValue;
        public Dictionary<ItemType, int> itemTypeCounts = new Dictionary<ItemType, int>();
        public Dictionary<ItemQuality, int> qualityCounts = new Dictionary<ItemQuality, int>();
        
        public InventoryStats()
        {
            // Initialize dictionaries
            foreach (ItemType itemType in Enum.GetValues(typeof(ItemType)))
            {
                itemTypeCounts[itemType] = 0;
            }
            
            foreach (ItemQuality quality in Enum.GetValues(typeof(ItemQuality)))
            {
                qualityCounts[quality] = 0;
            }
        }
    }
    
    /// <summary>
    /// 在庫移動操作の結果
    /// </summary>
    public class InventoryMoveResult
    {
        public bool Success { get; set; }
        public string ErrorMessage { get; set; }
        public int ItemsMoved { get; set; }
        
        public static InventoryMoveResult CreateSuccess(int itemsMoved)
        {
            return new InventoryMoveResult
            {
                Success = true,
                ItemsMoved = itemsMoved,
                ErrorMessage = ""
            };
        }
        
        public static InventoryMoveResult CreateFailure(string errorMessage)
        {
            return new InventoryMoveResult
            {
                Success = false,
                ItemsMoved = 0,
                ErrorMessage = errorMessage
            };
        }
    }
    
    /// <summary>
    /// 在庫検索フィルター
    /// </summary>
    [System.Serializable]
    public class InventoryFilter
    {
        public ItemType? itemType = null;
        public ItemQuality? quality = null;
        public InventoryLocation? location = null;
        public bool onlyExpired = false;
        public bool onlyExpiringSoon = false;
        public int expiringSoonDays = 2;
        
        /// <summary>
        /// アイテムがフィルター条件に合致するかチェック
        /// </summary>
        public bool Matches(InventoryItem item, int currentDay)
        {
            if (itemType.HasValue && item.itemType != itemType.Value)
                return false;
            
            if (quality.HasValue && item.quality != quality.Value)
                return false;
            
            if (location.HasValue && item.location != location.Value)
                return false;
            
            if (onlyExpired && !item.IsExpired(currentDay))
                return false;
            
            if (onlyExpiringSoon)
            {
                int daysLeft = item.GetDaysUntilExpiry(currentDay);
                if (daysLeft < 0 || daysLeft > expiringSoonDays)
                    return false;
            }
            
            return true;
        }
    }
}