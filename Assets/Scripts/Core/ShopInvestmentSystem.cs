using System;
using System.Collections.Generic;
using MerchantTails.Data;
using MerchantTails.Events;
using UnityEngine;

namespace MerchantTails.Core
{
    /// <summary>
    /// 店舗投資システム
    /// 店舗の設備改善による効率向上を管理
    /// </summary>
    public class ShopInvestmentSystem : MonoBehaviour
    {
        private static ShopInvestmentSystem instance;
        public static ShopInvestmentSystem Instance => instance;

        [Header("Investment Categories")]
        [SerializeField]
        private List<ShopUpgrade> availableUpgrades = new List<ShopUpgrade>();

        private Dictionary<string, ShopUpgradeProgress> upgradeProgress = new Dictionary<string, ShopUpgradeProgress>();
        private PlayerData playerData;
        private FeatureUnlockSystem featureUnlockSystem;

        // 現在の効果
        private float totalStorageBonus = 0f;
        private float totalEfficiencyBonus = 0f;
        private float totalCustomerBonus = 0f;
        private float totalQualityBonus = 0f;

        // プロパティ
        public float StorageCapacityMultiplier => 1f + totalStorageBonus;
        public float TransactionEfficiencyMultiplier => 1f + totalEfficiencyBonus;
        public float CustomerFlowMultiplier => 1f + totalCustomerBonus;
        public float ItemQualityMultiplier => 1f + totalQualityBonus;
        public bool IsUnlocked =>
            featureUnlockSystem != null && featureUnlockSystem.IsFeatureUnlocked(GameFeature.ShopInvestment);

        // イベント
        public event Action<ShopUpgrade, int> OnUpgradePurchased;
        public event Action<ShopUpgrade> OnUpgradeMaxed;
        public event Action OnBonusesUpdated;

        private void Awake()
        {
            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;

            InitializeUpgrades();
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
            featureUnlockSystem = FeatureUnlockSystem.Instance;

            SubscribeToEvents();
            LoadUpgradeData();
            CalculateTotalBonuses();
        }

        private void InitializeUpgrades()
        {
            if (availableUpgrades.Count == 0)
            {
                availableUpgrades = new List<ShopUpgrade>
                {
                    // 保管設備
                    new ShopUpgrade
                    {
                        id = "storage_basic",
                        name = "基本倉庫拡張",
                        description = "商品の保管容量を増やす",
                        category = UpgradeCategory.Storage,
                        maxLevel = 5,
                        baseCost = 1000,
                        costMultiplier = 1.5f,
                        effectType = UpgradeEffectType.StorageCapacity,
                        effectPerLevel = 0.1f, // +10%/レベル
                        iconName = "icon_storage",
                    },
                    new ShopUpgrade
                    {
                        id = "storage_cooling",
                        name = "冷蔵設備",
                        description = "果物の劣化速度を遅くする",
                        category = UpgradeCategory.Storage,
                        maxLevel = 3,
                        baseCost = 2000,
                        costMultiplier = 2f,
                        effectType = UpgradeEffectType.QualityPreservation,
                        effectPerLevel = 0.15f,
                        requiredRank = MerchantRank.Skilled,
                    },
                    // 効率化設備
                    new ShopUpgrade
                    {
                        id = "efficiency_counter",
                        name = "高速レジカウンター",
                        description = "取引速度を向上させる",
                        category = UpgradeCategory.Efficiency,
                        maxLevel = 4,
                        baseCost = 1500,
                        costMultiplier = 1.8f,
                        effectType = UpgradeEffectType.TransactionSpeed,
                        effectPerLevel = 0.12f,
                    },
                    new ShopUpgrade
                    {
                        id = "efficiency_display",
                        name = "商品陳列棚",
                        description = "商品の魅力を高める",
                        category = UpgradeCategory.Efficiency,
                        maxLevel = 5,
                        baseCost = 800,
                        costMultiplier = 1.4f,
                        effectType = UpgradeEffectType.CustomerAttraction,
                        effectPerLevel = 0.08f,
                    },
                    // 集客設備
                    new ShopUpgrade
                    {
                        id = "customer_sign",
                        name = "看板改良",
                        description = "来客数を増やす",
                        category = UpgradeCategory.Customer,
                        maxLevel = 3,
                        baseCost = 500,
                        costMultiplier = 2.5f,
                        effectType = UpgradeEffectType.CustomerFlow,
                        effectPerLevel = 0.15f,
                    },
                    new ShopUpgrade
                    {
                        id = "customer_comfort",
                        name = "店内環境改善",
                        description = "顧客満足度を高める",
                        category = UpgradeCategory.Customer,
                        maxLevel = 4,
                        baseCost = 1200,
                        costMultiplier = 1.6f,
                        effectType = UpgradeEffectType.CustomerSatisfaction,
                        effectPerLevel = 0.1f,
                        requiredRank = MerchantRank.Skilled,
                    },
                    // 特殊設備（ベテラン以上）
                    new ShopUpgrade
                    {
                        id = "special_security",
                        name = "セキュリティシステム",
                        description = "高額商品の取り扱いが可能に",
                        category = UpgradeCategory.Special,
                        maxLevel = 2,
                        baseCost = 5000,
                        costMultiplier = 3f,
                        effectType = UpgradeEffectType.PremiumItems,
                        effectPerLevel = 0.25f,
                        requiredRank = MerchantRank.Veteran,
                    },
                    new ShopUpgrade
                    {
                        id = "special_automation",
                        name = "自動化システム",
                        description = "店舗運営を部分的に自動化",
                        category = UpgradeCategory.Special,
                        maxLevel = 3,
                        baseCost = 8000,
                        costMultiplier = 2.5f,
                        effectType = UpgradeEffectType.Automation,
                        effectPerLevel = 0.2f,
                        requiredRank = MerchantRank.Veteran,
                    },
                    // マスター専用設備
                    new ShopUpgrade
                    {
                        id = "master_prestige",
                        name = "威信の店構え",
                        description = "すべての効果を向上させる",
                        category = UpgradeCategory.Master,
                        maxLevel = 1,
                        baseCost = 20000,
                        costMultiplier = 1f,
                        effectType = UpgradeEffectType.AllBonuses,
                        effectPerLevel = 0.5f,
                        requiredRank = MerchantRank.Master,
                    },
                };
            }

            // 進捗データを初期化
            foreach (var upgrade in availableUpgrades)
            {
                if (!upgradeProgress.ContainsKey(upgrade.id))
                {
                    upgradeProgress[upgrade.id] = new ShopUpgradeProgress { currentLevel = 0, totalInvested = 0 };
                }
            }
        }

        private void SubscribeToEvents()
        {
            EventBus.Subscribe<FeatureUnlockedEvent>(OnFeatureUnlocked);
            EventBus.Subscribe<RankChangedEvent>(OnRankChanged);
        }

        private void UnsubscribeFromEvents()
        {
            EventBus.Unsubscribe<FeatureUnlockedEvent>(OnFeatureUnlocked);
            EventBus.Unsubscribe<RankChangedEvent>(OnRankChanged);
        }

        private void OnFeatureUnlocked(FeatureUnlockedEvent e)
        {
            if (e.Feature == GameFeature.ShopInvestment)
            {
                ErrorHandler.LogInfo("Shop investment system unlocked!", "ShopInvestment");
                EventBus.Publish(new ShopInvestmentUnlockedEvent());
            }
        }

        private void OnRankChanged(RankChangedEvent e)
        {
            // ランクアップで新しいアップグレードが解放される可能性がある
            OnBonusesUpdated?.Invoke();
        }

        /// <summary>
        /// アップグレードを購入
        /// </summary>
        public bool PurchaseUpgrade(string upgradeId)
        {
            if (!IsUnlocked)
            {
                ErrorHandler.LogWarning("Shop investment feature is not unlocked", "ShopInvestment");
                return false;
            }

            var upgrade = GetUpgrade(upgradeId);
            if (upgrade == null)
            {
                ErrorHandler.LogError($"Upgrade not found: {upgradeId}", "ShopInvestment");
                return false;
            }

            var progress = GetProgress(upgradeId);
            if (progress == null || progress.currentLevel >= upgrade.maxLevel)
            {
                ErrorHandler.LogWarning($"Upgrade already maxed: {upgradeId}", "ShopInvestment");
                return false;
            }

            // ランク要件チェック
            if (upgrade.requiredRank > MerchantRank.Apprentice && playerData.CurrentRank < upgrade.requiredRank)
            {
                ErrorHandler.LogWarning($"Rank requirement not met for {upgradeId}", "ShopInvestment");
                return false;
            }

            // コスト計算
            int cost = CalculateUpgradeCost(upgrade, progress.currentLevel);

            if (playerData.CurrentMoney < cost)
            {
                ErrorHandler.LogWarning(
                    $"Insufficient funds for upgrade. Required: {cost}, Available: {playerData.CurrentMoney}",
                    "ShopInvestment"
                );
                return false;
            }

            // 購入処理
            if (playerData.ChangeMoney(-cost))
            {
                progress.currentLevel++;
                progress.totalInvested += cost;

                // ボーナス再計算
                CalculateTotalBonuses();

                OnUpgradePurchased?.Invoke(upgrade, progress.currentLevel);

                if (progress.currentLevel >= upgrade.maxLevel)
                {
                    OnUpgradeMaxed?.Invoke(upgrade);
                }

                EventBus.Publish(new ShopUpgradePurchasedEvent(upgrade, progress.currentLevel, cost));
                ErrorHandler.LogInfo(
                    $"Purchased {upgrade.name} Level {progress.currentLevel} for {cost}G",
                    "ShopInvestment"
                );

                SaveUpgradeData();
                return true;
            }

            return false;
        }

        /// <summary>
        /// アップグレードのコストを計算
        /// </summary>
        public int CalculateUpgradeCost(ShopUpgrade upgrade, int currentLevel)
        {
            return Mathf.RoundToInt(upgrade.baseCost * Mathf.Pow(upgrade.costMultiplier, currentLevel));
        }

        /// <summary>
        /// 特定のアップグレードを取得
        /// </summary>
        public ShopUpgrade GetUpgrade(string upgradeId)
        {
            return availableUpgrades.Find(u => u.id == upgradeId);
        }

        /// <summary>
        /// アップグレードの進捗を取得
        /// </summary>
        public ShopUpgradeProgress GetProgress(string upgradeId)
        {
            return upgradeProgress.TryGetValue(upgradeId, out var progress) ? progress : null;
        }

        /// <summary>
        /// 利用可能なアップグレードのリストを取得
        /// </summary>
        public List<ShopUpgrade> GetAvailableUpgrades()
        {
            if (!IsUnlocked)
                return new List<ShopUpgrade>();

            return availableUpgrades.FindAll(u =>
                u.requiredRank <= playerData.CurrentRank && GetProgress(u.id).currentLevel < u.maxLevel
            );
        }

        /// <summary>
        /// カテゴリ別のアップグレードを取得
        /// </summary>
        public List<ShopUpgrade> GetUpgradesByCategory(UpgradeCategory category)
        {
            return availableUpgrades.FindAll(u => u.category == category);
        }

        /// <summary>
        /// 総投資額を取得
        /// </summary>
        public int GetTotalInvestment()
        {
            int total = 0;
            foreach (var progress in upgradeProgress.Values)
            {
                total += progress.totalInvested;
            }
            return total;
        }

        /// <summary>
        /// 現在の効果ボーナスを再計算
        /// </summary>
        private void CalculateTotalBonuses()
        {
            totalStorageBonus = 0f;
            totalEfficiencyBonus = 0f;
            totalCustomerBonus = 0f;
            totalQualityBonus = 0f;

            foreach (var upgrade in availableUpgrades)
            {
                var progress = GetProgress(upgrade.id);
                if (progress == null || progress.currentLevel == 0)
                    continue;

                float effect = upgrade.effectPerLevel * progress.currentLevel;

                switch (upgrade.effectType)
                {
                    case UpgradeEffectType.StorageCapacity:
                        totalStorageBonus += effect;
                        break;
                    case UpgradeEffectType.TransactionSpeed:
                    case UpgradeEffectType.Automation:
                        totalEfficiencyBonus += effect;
                        break;
                    case UpgradeEffectType.CustomerFlow:
                    case UpgradeEffectType.CustomerAttraction:
                    case UpgradeEffectType.CustomerSatisfaction:
                        totalCustomerBonus += effect;
                        break;
                    case UpgradeEffectType.QualityPreservation:
                    case UpgradeEffectType.PremiumItems:
                        totalQualityBonus += effect;
                        break;
                    case UpgradeEffectType.AllBonuses:
                        totalStorageBonus += effect * 0.25f;
                        totalEfficiencyBonus += effect * 0.25f;
                        totalCustomerBonus += effect * 0.25f;
                        totalQualityBonus += effect * 0.25f;
                        break;
                }
            }

            OnBonusesUpdated?.Invoke();
            ErrorHandler.LogInfo(
                $"Shop bonuses updated - Storage: +{totalStorageBonus * 100}%, Efficiency: +{totalEfficiencyBonus * 100}%, Customer: +{totalCustomerBonus * 100}%, Quality: +{totalQualityBonus * 100}%",
                "ShopInvestment"
            );
        }

        // セーブ/ロード
        private void SaveUpgradeData()
        {
            var data = new ShopInvestmentSaveData { upgradeProgress = upgradeProgress };

            string json = JsonUtility.ToJson(data);
            PlayerPrefs.SetString($"ShopInvestment_{playerData?.PlayerName}", json);
            PlayerPrefs.Save();
        }

        private void LoadUpgradeData()
        {
            string key = $"ShopInvestment_{playerData?.PlayerName}";
            if (PlayerPrefs.HasKey(key))
            {
                string json = PlayerPrefs.GetString(key);
                var data = JsonUtility.FromJson<ShopInvestmentSaveData>(json);

                if (data.upgradeProgress != null)
                {
                    upgradeProgress = data.upgradeProgress;
                }
            }
        }
    }

    /// <summary>
    /// 店舗アップグレードデータ
    /// </summary>
    [Serializable]
    public class ShopUpgrade
    {
        public string id;
        public string name;
        public string description;
        public UpgradeCategory category;
        public int maxLevel;
        public int baseCost;
        public float costMultiplier;
        public UpgradeEffectType effectType;
        public float effectPerLevel;
        public MerchantRank requiredRank = MerchantRank.Apprentice;
        public string iconName;
    }

    /// <summary>
    /// アップグレードカテゴリ
    /// </summary>
    public enum UpgradeCategory
    {
        Storage, // 保管設備
        Efficiency, // 効率化
        Customer, // 集客
        Special, // 特殊
        Master, // マスター専用
    }

    /// <summary>
    /// アップグレード効果タイプ
    /// </summary>
    public enum UpgradeEffectType
    {
        StorageCapacity, // 保管容量
        TransactionSpeed, // 取引速度
        CustomerFlow, // 来客数
        CustomerAttraction, // 商品魅力
        CustomerSatisfaction, // 顧客満足度
        QualityPreservation, // 品質保持
        PremiumItems, // 高級品取扱
        Automation, // 自動化
        AllBonuses, // 全効果
    }

    /// <summary>
    /// アップグレード進捗データ
    /// </summary>
    [Serializable]
    public class ShopUpgradeProgress
    {
        public int currentLevel;
        public int totalInvested;
    }

    /// <summary>
    /// 保存データ
    /// </summary>
    [Serializable]
    public class ShopInvestmentSaveData
    {
        public Dictionary<string, ShopUpgradeProgress> upgradeProgress;
    }

    // イベント
    public class ShopInvestmentUnlockedEvent : BaseGameEvent { }

    public class ShopUpgradePurchasedEvent : BaseGameEvent
    {
        public ShopUpgrade Upgrade { get; }
        public int NewLevel { get; }
        public int Cost { get; }

        public ShopUpgradePurchasedEvent(ShopUpgrade upgrade, int newLevel, int cost)
        {
            Upgrade = upgrade;
            NewLevel = newLevel;
            Cost = cost;
        }
    }
}
