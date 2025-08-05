using System;
using System.Collections.Generic;
using MerchantTails.Data;
using UnityEngine;

namespace MerchantTails.Core
{
    /// <summary>
    /// ランクに応じた機能解放を管理するシステム
    /// </summary>
    public class FeatureUnlockSystem : MonoBehaviour
    {
        private static FeatureUnlockSystem instance;
        public static FeatureUnlockSystem Instance => instance;

        [Header("Feature Settings")]
        [SerializeField]
        private List<FeatureUnlock> featureUnlocks = new List<FeatureUnlock>();

        private Dictionary<GameFeature, bool> unlockedFeatures = new Dictionary<GameFeature, bool>();
        private PlayerData playerData;

        public event Action<GameFeature> OnFeatureUnlocked;

        private void Awake()
        {
            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;

            InitializeFeatures();
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
            SubscribeToEvents();
            UpdateUnlockedFeatures();
        }

        private void InitializeFeatures()
        {
            // デフォルトの機能解放設定
            if (featureUnlocks.Count == 0)
            {
                featureUnlocks = new List<FeatureUnlock>
                {
                    // 見習い（Apprentice）- 基本機能のみ
                    new FeatureUnlock
                    {
                        feature = GameFeature.BasicTrading,
                        requiredRank = MerchantRank.Apprentice,
                        featureName = "基本取引",
                        description = "商品の売買ができます",
                    },
                    new FeatureUnlock
                    {
                        feature = GameFeature.SimpleInventory,
                        requiredRank = MerchantRank.Apprentice,
                        featureName = "簡易在庫管理",
                        description = "基本的な在庫管理ができます",
                    },
                    // 一人前（Skilled）- 価格予測解放
                    new FeatureUnlock
                    {
                        feature = GameFeature.PricePrediction,
                        requiredRank = MerchantRank.Skilled,
                        featureName = "価格予測",
                        description = "翌日の価格動向を予測できます",
                        unlockMessage = "価格予測機能が解放されました！相場画面で翌日の価格動向を確認できます。",
                    },
                    new FeatureUnlock
                    {
                        feature = GameFeature.MarketTrends,
                        requiredRank = MerchantRank.Skilled,
                        featureName = "市場トレンド分析",
                        description = "市場の傾向を分析できます",
                    },
                    new FeatureUnlock
                    {
                        feature = GameFeature.BankAccount,
                        requiredRank = MerchantRank.Skilled,
                        featureName = "商人銀行口座",
                        description = "銀行に預金して利息を得られます",
                        unlockMessage = "商人銀行が利用可能になりました！お金を預けて複利で増やしましょう。",
                    },
                    // ベテラン（Veteran）- 高度な分析解放
                    new FeatureUnlock
                    {
                        feature = GameFeature.AdvancedAnalytics,
                        requiredRank = MerchantRank.Veteran,
                        featureName = "高度な市場分析",
                        description = "詳細な市場データと分析ツールが使えます",
                        unlockMessage = "高度な分析機能が解放されました！より詳細な市場データを確認できます。",
                    },
                    new FeatureUnlock
                    {
                        feature = GameFeature.EventPrediction,
                        requiredRank = MerchantRank.Veteran,
                        featureName = "イベント予測",
                        description = "今後のイベント発生を予測できます",
                    },
                    new FeatureUnlock
                    {
                        feature = GameFeature.ShopInvestment,
                        requiredRank = MerchantRank.Veteran,
                        featureName = "店舗投資",
                        description = "店舗に投資して設備を改善できます",
                        unlockMessage = "店舗投資が可能になりました！設備を改善して効率を上げましょう。",
                    },
                    new FeatureUnlock
                    {
                        feature = GameFeature.AutoPricing,
                        requiredRank = MerchantRank.Veteran,
                        featureName = "自動価格設定",
                        description = "商品の価格を自動で最適化します",
                    },
                    // マスター（Master）- 全機能解放
                    new FeatureUnlock
                    {
                        feature = GameFeature.MerchantNetwork,
                        requiredRank = MerchantRank.Master,
                        featureName = "商人ネットワーク",
                        description = "他の商人と連携して利益を最大化します",
                        unlockMessage = "商人ネットワークが解放されました！他の商人に出資して配当を得られます。",
                    },
                    new FeatureUnlock
                    {
                        feature = GameFeature.MarketManipulation,
                        requiredRank = MerchantRank.Master,
                        featureName = "市場影響力",
                        description = "大量取引で市場価格に影響を与えます",
                    },
                    new FeatureUnlock
                    {
                        feature = GameFeature.ExclusiveDeals,
                        requiredRank = MerchantRank.Master,
                        featureName = "独占取引",
                        description = "特別な取引機会にアクセスできます",
                    },
                    new FeatureUnlock
                    {
                        feature = GameFeature.FullAutomation,
                        requiredRank = MerchantRank.Master,
                        featureName = "完全自動化",
                        description = "取引を完全に自動化できます",
                    },
                };
            }

            // 全機能の解放状態を初期化
            foreach (GameFeature feature in Enum.GetValues(typeof(GameFeature)))
            {
                unlockedFeatures[feature] = false;
            }
        }

        private void SubscribeToEvents()
        {
            EventBus.Subscribe<RankChangedEvent>(OnRankChanged);
        }

        private void UnsubscribeFromEvents()
        {
            EventBus.Unsubscribe<RankChangedEvent>(OnRankChanged);
        }

        private void OnRankChanged(RankChangedEvent e)
        {
            if (e.IsRankUp)
            {
                UpdateUnlockedFeatures();
                ShowUnlockNotifications(e.PreviousRank, e.NewRank);
            }
        }

        private void UpdateUnlockedFeatures()
        {
            if (playerData == null)
                return;

            MerchantRank currentRank = playerData.CurrentRank;

            foreach (var unlock in featureUnlocks)
            {
                bool wasUnlocked = unlockedFeatures[unlock.feature];
                bool shouldBeUnlocked = currentRank >= unlock.requiredRank;

                if (!wasUnlocked && shouldBeUnlocked)
                {
                    unlockedFeatures[unlock.feature] = true;
                    OnFeatureUnlocked?.Invoke(unlock.feature);

                    ErrorHandler.LogInfo($"Feature unlocked: {unlock.featureName}", "FeatureUnlock");
                }
                else
                {
                    unlockedFeatures[unlock.feature] = shouldBeUnlocked;
                }
            }
        }

        private void ShowUnlockNotifications(MerchantRank previousRank, MerchantRank newRank)
        {
            var newlyUnlocked = featureUnlocks.FindAll(f => f.requiredRank > previousRank && f.requiredRank <= newRank);

            foreach (var unlock in newlyUnlocked)
            {
                if (!string.IsNullOrEmpty(unlock.unlockMessage))
                {
                    // UIManagerを通じて通知を表示
                    if (UIManager.Instance != null)
                    {
                        UIManager.Instance.ShowNotification(
                            "新機能解放！",
                            unlock.unlockMessage,
                            5f,
                            UIManager.NotificationType.Success
                        );
                    }

                    // イベントを発行
                    EventBus.Publish(new FeatureUnlockedEvent(unlock.feature, unlock.featureName));
                }
            }
        }

        /// <summary>
        /// 特定の機能が解放されているか確認
        /// </summary>
        public bool IsFeatureUnlocked(GameFeature feature)
        {
            return unlockedFeatures.TryGetValue(feature, out bool unlocked) && unlocked;
        }

        /// <summary>
        /// 機能の情報を取得
        /// </summary>
        public FeatureUnlock GetFeatureInfo(GameFeature feature)
        {
            return featureUnlocks.Find(f => f.feature == feature);
        }

        /// <summary>
        /// 現在のランクで解放されている機能のリストを取得
        /// </summary>
        public List<FeatureUnlock> GetUnlockedFeatures()
        {
            if (playerData == null)
                return new List<FeatureUnlock>();

            return featureUnlocks.FindAll(f => f.requiredRank <= playerData.CurrentRank);
        }

        /// <summary>
        /// 次のランクで解放される機能のリストを取得
        /// </summary>
        public List<FeatureUnlock> GetNextRankFeatures()
        {
            if (playerData == null)
                return new List<FeatureUnlock>();

            MerchantRank nextRank = GetNextRank(playerData.CurrentRank);
            if (nextRank == playerData.CurrentRank)
                return new List<FeatureUnlock>(); // 既に最高ランク

            return featureUnlocks.FindAll(f => f.requiredRank == nextRank);
        }

        private MerchantRank GetNextRank(MerchantRank currentRank)
        {
            switch (currentRank)
            {
                case MerchantRank.Apprentice:
                    return MerchantRank.Skilled;
                case MerchantRank.Skilled:
                    return MerchantRank.Veteran;
                case MerchantRank.Veteran:
                    return MerchantRank.Master;
                default:
                    return MerchantRank.Master;
            }
        }

        /// <summary>
        /// デバッグ用：特定の機能を強制的に解放
        /// </summary>
        [ContextMenu("Unlock All Features")]
        public void UnlockAllFeatures()
        {
            foreach (GameFeature feature in Enum.GetValues(typeof(GameFeature)))
            {
                unlockedFeatures[feature] = true;
            }
            ErrorHandler.LogInfo("All features unlocked (Debug)", "FeatureUnlock");
        }
    }

    /// <summary>
    /// 機能解放の設定データ
    /// </summary>
    [Serializable]
    public class FeatureUnlock
    {
        public GameFeature feature;
        public MerchantRank requiredRank;
        public string featureName;
        public string description;
        public string unlockMessage;
        public Sprite icon;
    }

}
