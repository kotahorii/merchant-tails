using System;
using MerchantTails.Data;
using UnityEngine;

namespace MerchantTails.Market
{
    /// <summary>
    /// 個別商品の市場データを格納するクラス
    /// 価格、需給、変動率などの情報を管理
    /// </summary>
    [System.Serializable]
    public class MarketData
    {
        [Header("Basic Information")]
        public ItemType itemType;
        public float basePrice;
        public float currentPrice;

        [Header("Market Dynamics")]
        public float volatility; // 価格変動の激しさ (0.0 - 1.0)
        public float demand; // 需要係数 (1.0 = 通常)
        public float supply; // 供給係数 (1.0 = 通常)

        [Header("Update Tracking")]
        public int lastUpdateDay;

        /// <summary>
        /// 基準価格からの変動率を計算
        /// </summary>
        public float GetPriceChangePercentage()
        {
            return basePrice > 0 ? ((currentPrice - basePrice) / basePrice) * 100f : 0f;
        }

        /// <summary>
        /// 現在の市場状況（需給バランス）を取得
        /// </summary>
        public MarketCondition GetMarketCondition()
        {
            float ratio = demand / supply;

            if (ratio >= 1.3f)
                return MarketCondition.HighDemand;
            if (ratio >= 1.1f)
                return MarketCondition.ModeratelyHigh;
            if (ratio <= 0.7f)
                return MarketCondition.LowDemand;
            if (ratio <= 0.9f)
                return MarketCondition.ModeratelyLow;
            return MarketCondition.Balanced;
        }

        /// <summary>
        /// 価格の安定性レベルを取得
        /// </summary>
        public PriceStability GetPriceStability()
        {
            return volatility switch
            {
                >= 0.3f => PriceStability.VeryVolatile,
                >= 0.2f => PriceStability.Volatile,
                >= 0.1f => PriceStability.Moderate,
                >= 0.05f => PriceStability.Stable,
                _ => PriceStability.VeryStable,
            };
        }
    }

    /// <summary>
    /// 価格履歴を記録するデータクラス
    /// </summary>
    [System.Serializable]
    public class PriceHistory
    {
        public int day;
        public float price;
        public DateTime timestamp;

        public PriceHistory()
        {
            timestamp = DateTime.Now;
        }

        public PriceHistory(int day, float price)
        {
            this.day = day;
            this.price = price;
            this.timestamp = DateTime.Now;
        }
    }

    /// <summary>
    /// 市場イベントを管理するデータクラス
    /// </summary>
    [System.Serializable]
    public class MarketEvent
    {
        public string eventName;
        public ItemType[] affectedItems;
        public float[] modifiers;
        public int remainingDuration;
        public string description;

        public bool IsActive => remainingDuration > 0;

        public void DecreaseDuration()
        {
            remainingDuration = Mathf.Max(0, remainingDuration - 1);
        }
    }

    /// <summary>
    /// 市場設定を管理するScriptableObject
    /// </summary>
    [CreateAssetMenu(fileName = "MarketConfiguration", menuName = "MerchantTails/Market Configuration")]
    public class MarketConfiguration : ScriptableObject
    {
        [Header("Base Prices")]
        public float fruitBasePrice = 10f;
        public float potionBasePrice = 50f;
        public float weaponBasePrice = 200f;
        public float accessoryBasePrice = 75f;
        public float magicBookBasePrice = 300f;
        public float gemBasePrice = 150f;

        [Header("Volatility Settings")]
        public float fruitVolatility = 0.3f;
        public float potionVolatility = 0.2f;
        public float weaponVolatility = 0.1f;
        public float accessoryVolatility = 0.25f;
        public float magicBookVolatility = 0.05f;
        public float gemVolatility = 0.4f;

        [Header("Market Settings")]
        public float maxPriceMultiplier = 3.0f;
        public float minPriceMultiplier = 0.3f;
        public float fluctuationIntensity = 1.0f;
        public int priceHistoryDays = 90;

        public float GetBasePrice(ItemType itemType)
        {
            return itemType switch
            {
                ItemType.Fruit => fruitBasePrice,
                ItemType.Potion => potionBasePrice,
                ItemType.Weapon => weaponBasePrice,
                ItemType.Accessory => accessoryBasePrice,
                ItemType.MagicBook => magicBookBasePrice,
                ItemType.Gem => gemBasePrice,
                _ => 100f,
            };
        }

        public float GetVolatility(ItemType itemType)
        {
            return itemType switch
            {
                ItemType.Fruit => fruitVolatility,
                ItemType.Potion => potionVolatility,
                ItemType.Weapon => weaponVolatility,
                ItemType.Accessory => accessoryVolatility,
                ItemType.MagicBook => magicBookVolatility,
                ItemType.Gem => gemVolatility,
                _ => 0.2f,
            };
        }
    }

    /// <summary>
    /// 市場の需給状況を表すenum
    /// </summary>
    public enum MarketCondition
    {
        LowDemand, // 需要不足
        ModeratelyLow, // やや需要不足
        Balanced, // バランス良好
        ModeratelyHigh, // やや需要過多
        HighDemand, // 需要過多
    }

    /// <summary>
    /// 価格の安定性を表すenum
    /// </summary>
    public enum PriceStability
    {
        VeryStable, // 非常に安定
        Stable, // 安定
        Moderate, // 普通
        Volatile, // 不安定
        VeryVolatile, // 非常に不安定
    }
}
