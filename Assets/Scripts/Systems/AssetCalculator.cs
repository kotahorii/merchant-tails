using System.Collections.Generic;
using MerchantTails.Data;
using MerchantTails.Events;
using MerchantTails.Inventory;
using MerchantTails.Market;
using UnityEngine;

namespace MerchantTails.Systems
{
    /// <summary>
    /// プレイヤーの総資産を計算するシステム
    /// 現金、在庫、投資などすべての資産を統合
    /// </summary>
    public class AssetCalculator : MonoBehaviour
    {
        private static AssetCalculator instance;
        public static AssetCalculator Instance => instance;

        [Header("Asset Calculation Settings")]
        [SerializeField]
        private float inventoryValueMultiplier = 0.8f; // 在庫は時価の80%で評価

        [SerializeField]
        private bool includeDecayedItems = false; // 劣化した商品を含めるか

        private PlayerData playerData;
        private InventorySystem inventorySystem;
        private MarketSystem marketSystem;

        private float lastCalculatedAssets = 0f;
        private float previousDayAssets = 0f;

        public float TotalAssets => lastCalculatedAssets;
        public float DailyProfit => lastCalculatedAssets - previousDayAssets;
        public float DailyProfitPercentage => previousDayAssets > 0 ? (DailyProfit / previousDayAssets) * 100f : 0f;

        private void Awake()
        {
            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;
        }

        private void OnDestroy()
        {
            if (instance == this)
            {
                instance = null;
            }
            UnsubscribeFromEvents();
        }

        private void Start()
        {
            playerData = GameManager.Instance?.PlayerData;
            inventorySystem = InventorySystem.Instance;
            marketSystem = MarketSystem.Instance;

            SubscribeToEvents();
            CalculateTotalAssets();
        }

        private void SubscribeToEvents()
        {
            // 資産に影響するイベントを監視
            EventBus.Subscribe<MoneyChangedEvent>(OnMoneyChanged);
            EventBus.Subscribe<TransactionCompletedEvent>(OnTransactionCompleted);
            EventBus.Subscribe<DayChangedEvent>(OnDayChanged);
            EventBus.Subscribe<ItemDecayedEvent>(OnItemDecayed);
        }

        private void UnsubscribeFromEvents()
        {
            EventBus.Unsubscribe<MoneyChangedEvent>(OnMoneyChanged);
            EventBus.Unsubscribe<TransactionCompletedEvent>(OnTransactionCompleted);
            EventBus.Unsubscribe<DayChangedEvent>(OnDayChanged);
            EventBus.Unsubscribe<ItemDecayedEvent>(OnItemDecayed);
        }

        private void OnMoneyChanged(MoneyChangedEvent e)
        {
            CalculateTotalAssets();
        }

        private void OnTransactionCompleted(TransactionCompletedEvent e)
        {
            CalculateTotalAssets();
        }

        private void OnDayChanged(DayChangedEvent e)
        {
            // 日が変わったら前日の資産を記録
            previousDayAssets = lastCalculatedAssets;
            CalculateTotalAssets();

            // 日次資産レポートを発行
            PublishDailyAssetReport();
        }

        private void OnItemDecayed(ItemDecayedEvent e)
        {
            CalculateTotalAssets();
        }

        /// <summary>
        /// 総資産を計算
        /// </summary>
        public float CalculateTotalAssets()
        {
            float totalAssets = 0f;

            // 1. 現金
            if (playerData != null)
            {
                totalAssets += playerData.CurrentMoney;
            }

            // 2. 在庫資産
            totalAssets += CalculateInventoryValue();

            // 3. 銀行預金（商人銀行システムが実装されたら追加）
            totalAssets += CalculateBankDeposits();

            // 4. 投資資産（店舗投資・他商人出資が実装されたら追加）
            totalAssets += CalculateInvestmentValue();

            lastCalculatedAssets = totalAssets;

            // 資産変動イベントを発行
            EventBus.Publish(new AssetChangedEvent(totalAssets, GetAssetBreakdown()));

            return totalAssets;
        }

        /// <summary>
        /// 在庫の総価値を計算
        /// </summary>
        private float CalculateInventoryValue()
        {
            if (inventorySystem == null || marketSystem == null)
                return 0f;

            float inventoryValue = 0f;
            var allItems = inventorySystem.GetAllItems();

            foreach (var item in allItems)
            {
                // 劣化したアイテムを除外する場合
                if (!includeDecayedItems && item.Value.condition <= 0f)
                    continue;

                // 現在の市場価格を取得
                float marketPrice = marketSystem.GetCurrentPrice(item.Key);

                // 在庫の総価値 = 数量 × 市場価格 × 評価率 × 品質
                float itemValue = item.Value.count * marketPrice * inventoryValueMultiplier;

                // 品質による価値の調整
                if (item.Value.condition > 0f && item.Value.condition < 1f)
                {
                    itemValue *= item.Value.condition;
                }

                inventoryValue += itemValue;
            }

            return inventoryValue;
        }

        /// <summary>
        /// 銀行預金の総額を計算
        /// </summary>
        private float CalculateBankDeposits()
        {
            // TODO: BankSystem implementation
            // return BankSystem.Instance?.GetTotalDeposits() ?? 0f;
            return 0f;
        }

        /// <summary>
        /// 投資資産の総価値を計算
        /// </summary>
        private float CalculateInvestmentValue()
        {
            float investmentValue = 0f;

            // TODO: ShopInvestmentSystem implementation
            // 店舗投資（設備投資は資産価値として計上）
            // investmentValue += ShopInvestmentSystem.Instance?.GetTotalInvestment() ?? 0f;

            // TODO: MerchantInvestmentSystem implementation
            // 他商人への出資
            // investmentValue += MerchantInvestmentSystem.Instance?.GetTotalInvestmentValue() ?? 0f;

            return investmentValue;
        }

        /// <summary>
        /// 資産の内訳を取得
        /// </summary>
        public AssetBreakdown GetAssetBreakdown()
        {
            var breakdown = new AssetBreakdown
            {
                cash = playerData?.CurrentMoney ?? 0f,
                inventoryValue = CalculateInventoryValue(),
                bankDeposits = CalculateBankDeposits(),
                investments = CalculateInvestmentValue(),
                totalAssets = lastCalculatedAssets,
            };

            return breakdown;
        }

        /// <summary>
        /// ランクアップに必要な資産を取得
        /// </summary>
        public float GetRequiredAssetsForNextRank()
        {
            if (playerData == null)
                return 0f;

            return playerData.CurrentRank switch
            {
                MerchantRank.Apprentice => 5000f, // 一人前になるには5,000G
                MerchantRank.Skilled => 10000f, // ベテランになるには10,000G
                MerchantRank.Veteran => 50000f, // マスターになるには50,000G
                MerchantRank.Master => float.MaxValue, // 既に最高ランク
                _ => 0f,
            };
        }

        /// <summary>
        /// 次のランクまでの進捗率を取得
        /// </summary>
        public float GetRankProgress()
        {
            if (playerData == null)
                return 0f;

            float requiredAssets = GetRequiredAssetsForNextRank();
            if (requiredAssets == float.MaxValue)
                return 1f; // 既に最高ランク

            float previousRankAssets = playerData.CurrentRank switch
            {
                MerchantRank.Apprentice => 0f,
                MerchantRank.Skilled => 5000f,
                MerchantRank.Veteran => 10000f,
                MerchantRank.Master => 50000f,
                _ => 0f,
            };

            float progress = (lastCalculatedAssets - previousRankAssets) / (requiredAssets - previousRankAssets);
            return Mathf.Clamp01(progress);
        }

        /// <summary>
        /// ランクアップ可能かチェック
        /// </summary>
        public bool CanRankUp()
        {
            if (playerData == null)
                return false;
            if (playerData.CurrentRank == MerchantRank.Master)
                return false;

            return lastCalculatedAssets >= GetRequiredAssetsForNextRank();
        }

        /// <summary>
        /// ランクアップを実行
        /// </summary>
        public void TryRankUp()
        {
            if (!CanRankUp())
                return;

            MerchantRank newRank = playerData.CurrentRank switch
            {
                MerchantRank.Apprentice => MerchantRank.Skilled,
                MerchantRank.Skilled => MerchantRank.Veteran,
                MerchantRank.Veteran => MerchantRank.Master,
                _ => playerData.CurrentRank,
            };

            var previousRank = playerData.CurrentRank;
            playerData.SetRank(newRank);

            // ランクアップイベントを発行
            EventBus.Publish(new RankChangedEvent(previousRank, newRank));

            ErrorHandler.LogInfo($"Rank up! {previousRank} -> {newRank}", "AssetCalculator");
        }

        /// <summary>
        /// 日次資産レポートを発行
        /// </summary>
        private void PublishDailyAssetReport()
        {
            var report = new DailyAssetReport
            {
                day = TimeManager.Instance?.CurrentDay ?? 1,
                totalAssets = lastCalculatedAssets,
                dailyProfit = DailyProfit,
                profitPercentage = DailyProfitPercentage,
                breakdown = GetAssetBreakdown(),
            };

            EventBus.Publish(new DailyAssetReportEvent(report));
        }
    }

    /// <summary>
    /// 資産の内訳
    /// </summary>
    [System.Serializable]
    public struct AssetBreakdown
    {
        public float cash;
        public float inventoryValue;
        public float bankDeposits;
        public float investments;
        public float totalAssets;
    }

    /// <summary>
    /// 日次資産レポート
    /// </summary>
    [System.Serializable]
    public struct DailyAssetReport
    {
        public int day;
        public float totalAssets;
        public float dailyProfit;
        public float profitPercentage;
        public AssetBreakdown breakdown;
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
}
