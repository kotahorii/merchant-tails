using System;
using System.Collections.Generic;
using MerchantTails.Data;

namespace MerchantTails.Core
{
    /// <summary>
    /// ゲーム状態が変更されたときのイベント
    /// </summary>
    public class GameStateChangedEvent : BaseGameEvent
    {
        public GameState PreviousState { get; }
        public GameState NewState { get; }

        public GameStateChangedEvent(GameState previousState, GameState newState)
        {
            PreviousState = previousState;
            NewState = newState;
        }
    }

    /// <summary>
    /// 時間フェーズが変更されたときのイベント
    /// </summary>
    public class PhaseChangedEvent : BaseGameEvent
    {
        public DayPhase PreviousPhase { get; }
        public DayPhase NewPhase { get; }
        public int CurrentDay { get; }

        public PhaseChangedEvent(DayPhase previousPhase, DayPhase newPhase, int currentDay)
        {
            PreviousPhase = previousPhase;
            NewPhase = newPhase;
            CurrentDay = currentDay;
        }
    }

    /// <summary>
    /// 季節が変更されたときのイベント
    /// </summary>
    public class SeasonChangedEvent : BaseGameEvent
    {
        public Season PreviousSeason { get; }
        public Season NewSeason { get; }
        public int Year { get; }

        public SeasonChangedEvent(Season previousSeason, Season newSeason, int year)
        {
            PreviousSeason = previousSeason;
            NewSeason = newSeason;
            Year = year;
        }
    }

    /// <summary>
    /// 価格が変更されたときのイベント
    /// </summary>
    public class PriceChangedEvent : BaseGameEvent
    {
        public ItemType ItemType { get; }
        public float PreviousPrice { get; }
        public float NewPrice { get; }
        public float ChangePercentage { get; }

        public PriceChangedEvent(ItemType itemType, float previousPrice, float newPrice)
        {
            ItemType = itemType;
            PreviousPrice = previousPrice;
            NewPrice = newPrice;
            ChangePercentage = previousPrice > 0 ? ((newPrice - previousPrice) / previousPrice) * 100f : 0f;
        }
    }

    /// <summary>
    /// プレイヤーのお金が変更されたときのイベント
    /// </summary>
    public class MoneyChangedEvent : BaseGameEvent
    {
        public int PreviousAmount { get; }
        public int NewAmount { get; }
        public int ChangeAmount { get; }
        public string Reason { get; }

        public MoneyChangedEvent(int previousAmount, int newAmount, string reason = "")
        {
            PreviousAmount = previousAmount;
            NewAmount = newAmount;
            ChangeAmount = newAmount - previousAmount;
            Reason = reason;
        }
    }

    /// <summary>
    /// プレイヤーのランクが変更されたときのイベント
    /// </summary>
    public class RankChangedEvent : BaseGameEvent
    {
        public MerchantRank PreviousRank { get; }
        public MerchantRank NewRank { get; }
        public bool IsRankUp { get; }

        public RankChangedEvent(MerchantRank previousRank, MerchantRank newRank)
        {
            PreviousRank = previousRank;
            NewRank = newRank;
            IsRankUp = newRank > previousRank;
        }
    }

    /// <summary>
    /// 取引が完了したときのイベント
    /// </summary>
    public class TransactionCompletedEvent : BaseGameEvent
    {
        public ItemType ItemType { get; }
        public int Quantity { get; }
        public float UnitPrice { get; }
        public float TotalPrice { get; }
        public bool IsPurchase { get; }
        public float Profit { get; }

        public TransactionCompletedEvent(
            ItemType itemType,
            int quantity,
            float unitPrice,
            bool isPurchase,
            float profit = 0f
        )
        {
            ItemType = itemType;
            Quantity = quantity;
            UnitPrice = unitPrice;
            TotalPrice = unitPrice * quantity;
            IsPurchase = isPurchase;
            Profit = profit;
        }
    }

    /// <summary>
    /// イベント（収穫祭、ドラゴン討伐など）が発生したときのイベント
    /// </summary>
    public class GameEventTriggeredEvent : BaseGameEvent
    {
        public string EventName { get; }
        public string Description { get; }
        public ItemType[] AffectedItems { get; }
        public float[] PriceModifiers { get; }
        public int Duration { get; }

        public GameEventTriggeredEvent(
            string eventName,
            string description,
            ItemType[] affectedItems,
            float[] priceModifiers,
            int duration
        )
        {
            EventName = eventName;
            Description = description;
            AffectedItems = affectedItems;
            PriceModifiers = priceModifiers;
            Duration = duration;
        }
    }

    /// <summary>
    /// チュートリアルステップが完了したときのイベント
    /// </summary>
    public class TutorialStepCompletedEvent : BaseGameEvent
    {
        public int StepNumber { get; }
        public string StepName { get; }
        public bool IsLastStep { get; }

        public TutorialStepCompletedEvent(int stepNumber, string stepName, bool isLastStep)
        {
            StepNumber = stepNumber;
            StepName = stepName;
            IsLastStep = isLastStep;
        }
    }

    /// <summary>
    /// セーブ/ロード操作が完了したときのイベント
    /// </summary>
    public class SaveLoadEvent : BaseGameEvent
    {
        public bool IsLoadOperation { get; }
        public bool Success { get; }
        public string ErrorMessage { get; }

        public SaveLoadEvent(bool isLoadOperation, bool success, string errorMessage = "")
        {
            IsLoadOperation = isLoadOperation;
            Success = success;
            ErrorMessage = errorMessage;
        }
    }

    /// <summary>
    /// 日が変更されたときのイベント
    /// </summary>
    public class DayChangedEvent : BaseGameEvent
    {
        public int PreviousDay { get; }
        public int NewDay { get; }
        public Season CurrentSeason { get; }
        public int CurrentYear { get; }

        public DayChangedEvent(int previousDay, int newDay, Season currentSeason, int currentYear)
        {
            PreviousDay = previousDay;
            NewDay = newDay;
            CurrentSeason = currentSeason;
            CurrentYear = currentYear;
        }
    }

    /// <summary>
    /// 年が変更されたときのイベント
    /// </summary>
    public class YearChangedEvent : BaseGameEvent
    {
        public int PreviousYear { get; }
        public int NewYear { get; }

        public YearChangedEvent(int previousYear, int newYear)
        {
            PreviousYear = previousYear;
            NewYear = newYear;
        }
    }

    /// <summary>購入完了を通知</summary>
    public class PurchaseCompletedEvent : BaseGameEvent
    {
        public List<(ItemType ItemType, int Quantity)> PurchasedItems { get; }
        public float TotalCost { get; }

        public PurchaseCompletedEvent(List<(ItemType, int)> purchasedItems, float totalCost)
        {
            PurchasedItems = new List<(ItemType, int)>(purchasedItems);
            TotalCost = totalCost;
        }
    }

    /// <summary>アイテムの腐敗を通知</summary>
    public class ItemDecayedEvent : BaseGameEvent
    {
        public ItemType ItemType { get; }
        public int Quantity { get; }
        public InventoryLocation Location { get; }

        public ItemDecayedEvent(ItemType itemType, int quantity, InventoryLocation location)
        {
            ItemType = itemType;
            Quantity = quantity;
            Location = location;
        }
    }

    /// <summary>市場イベントの発生を通知</summary>
    public class MarketEventTriggeredEvent : BaseGameEvent
    {
        public string EventName { get; }
        public string Description { get; }
        public Dictionary<ItemType, float> PriceModifiers { get; }

        public MarketEventTriggeredEvent(
            string eventName,
            string description,
            Dictionary<ItemType, float> priceModifiers
        )
        {
            EventName = eventName;
            Description = description;
            PriceModifiers = new Dictionary<ItemType, float>(priceModifiers);
        }
    }

    /// <summary>
    /// 機能が解放されたときのイベント
    /// </summary>
    public class FeatureUnlockedEvent : BaseGameEvent
    {
        public GameFeature Feature { get; }
        public string FeatureName { get; }

        public FeatureUnlockedEvent(GameFeature feature, string featureName)
        {
            Feature = feature;
            FeatureName = featureName;
        }
    }

    /// <summary>
    /// 資産変動イベント
    /// </summary>
    public class AssetChangedEvent : BaseGameEvent
    {
        public float TotalAssets { get; }
        public AssetBreakdown Breakdown { get; }

        public AssetChangedEvent(float totalAssets, AssetBreakdown breakdown)
        {
            TotalAssets = totalAssets;
            Breakdown = breakdown;
        }
    }

    /// <summary>
    /// 日次資産レポートイベント
    /// </summary>
    public class DailyAssetReportEvent : BaseGameEvent
    {
        public DailyAssetReport Report { get; }

        public DailyAssetReportEvent(DailyAssetReport report)
        {
            Report = report;
        }
    }

    /// <summary>
    /// 在庫が変更されたときのイベント
    /// </summary>
    public class InventoryChangedEvent : BaseGameEvent
    {
        public ItemType ItemType { get; }
        public int QuantityChange { get; }
        public InventoryLocation Location { get; }

        public InventoryChangedEvent(ItemType itemType, int quantityChange, InventoryLocation location)
        {
            ItemType = itemType;
            QuantityChange = quantityChange;
            Location = location;
        }
    }

    /// <summary>
    /// 時間が進んだときのイベント
    /// </summary>
    public class TimeAdvancedEvent : BaseGameEvent
    {
        public DayPhase CurrentPhase { get; }
        public int CurrentDay { get; }
        public Season CurrentSeason { get; }

        public TimeAdvancedEvent(DayPhase currentPhase, int currentDay, Season currentSeason)
        {
            CurrentPhase = currentPhase;
            CurrentDay = currentDay;
            CurrentSeason = currentSeason;
        }
    }
}
