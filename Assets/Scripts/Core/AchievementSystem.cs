using System;
using System.Collections.Generic;
using System.Linq;
using MerchantTails.Data;
using MerchantTails.Inventory;
using MerchantTails.UI;
using UnityEngine;

namespace MerchantTails.Core
{
    /// <summary>
    /// 実績システムを管理するクラス
    /// プレイヤーの行動や成果に応じて実績を解除
    /// </summary>
    public class AchievementSystem : MonoBehaviour
    {
        private static AchievementSystem instance;
        public static AchievementSystem Instance => instance;

        [Header("Achievement Settings")]
        [SerializeField]
        private List<Achievement> achievements = new List<Achievement>();

        [SerializeField]
        private bool saveAchievementsGlobally = true; // アカウント単位で保存

        private Dictionary<string, AchievementProgress> progressData = new Dictionary<string, AchievementProgress>();
        private PlayerData playerData;

        public event Action<Achievement> OnAchievementUnlocked;
        public event Action<Achievement, float> OnAchievementProgress;

        private void Awake()
        {
            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;

            InitializeAchievements();
            LoadAchievementData();
        }

        private void OnDestroy()
        {
            if (instance == this)
            {
                instance = null;
            }
            UnsubscribeFromEvents();
            SaveAchievementData();
        }

        private void Start()
        {
            playerData = GameManager.Instance?.PlayerData;
            SubscribeToEvents();
        }

        private void InitializeAchievements()
        {
            if (achievements.Count == 0)
            {
                achievements = new List<Achievement>
                {
                    // 取引関連の実績
                    new Achievement
                    {
                        id = "first_trade",
                        name = "初めての取引",
                        description = "最初の商品を売買する",
                        points = 10,
                        category = AchievementCategory.Trading,
                        maxProgress = 1,
                        hidden = false,
                    },
                    new Achievement
                    {
                        id = "trader_100",
                        name = "百人力の商人",
                        description = "100回の取引を完了する",
                        points = 50,
                        category = AchievementCategory.Trading,
                        maxProgress = 100,
                    },
                    new Achievement
                    {
                        id = "profit_master",
                        name = "利益の達人",
                        description = "1回の取引で1,000G以上の利益を出す",
                        points = 30,
                        category = AchievementCategory.Trading,
                        maxProgress = 1,
                    },
                    // 資産関連の実績
                    new Achievement
                    {
                        id = "millionaire",
                        name = "百万長者",
                        description = "総資産が100,000Gを超える",
                        points = 100,
                        category = AchievementCategory.Wealth,
                        maxProgress = 100000,
                        showProgressBar = true,
                    },
                    new Achievement
                    {
                        id = "first_5000",
                        name = "一人前の証",
                        description = "総資産が5,000Gに到達",
                        points = 20,
                        category = AchievementCategory.Wealth,
                        maxProgress = 5000,
                        showProgressBar = true,
                    },
                    // ランク関連の実績
                    new Achievement
                    {
                        id = "rank_skilled",
                        name = "一人前商人",
                        description = "商人ランクが「一人前」に到達",
                        points = 30,
                        category = AchievementCategory.Progression,
                        maxProgress = 1,
                    },
                    new Achievement
                    {
                        id = "rank_master",
                        name = "商業王",
                        description = "商人ランクが「マスター」に到達",
                        points = 100,
                        category = AchievementCategory.Progression,
                        maxProgress = 1,
                    },
                    // 季節関連の実績
                    new Achievement
                    {
                        id = "four_seasons",
                        name = "四季の商人",
                        description = "4つの季節すべてで商売を経験",
                        points = 40,
                        category = AchievementCategory.Seasonal,
                        maxProgress = 4,
                        showProgressBar = true,
                    },
                    new Achievement
                    {
                        id = "winter_survivor",
                        name = "冬を越えて",
                        description = "冬の間に黒字を維持",
                        points = 50,
                        category = AchievementCategory.Seasonal,
                        maxProgress = 1,
                    },
                    // アイテム関連の実績
                    new Achievement
                    {
                        id = "fruit_expert",
                        name = "果物商人",
                        description = "果物を100個売る",
                        points = 20,
                        category = AchievementCategory.Specialist,
                        maxProgress = 100,
                        showProgressBar = true,
                    },
                    new Achievement
                    {
                        id = "gem_collector",
                        name = "宝石収集家",
                        description = "宝石を10個以上所持",
                        points = 30,
                        category = AchievementCategory.Specialist,
                        maxProgress = 10,
                        showProgressBar = true,
                    },
                    new Achievement
                    {
                        id = "all_items",
                        name = "万物の商人",
                        description = "6種類すべての商品を扱う",
                        points = 40,
                        category = AchievementCategory.Specialist,
                        maxProgress = 6,
                        showProgressBar = true,
                    },
                    // イベント関連の実績
                    new Achievement
                    {
                        id = "event_master",
                        name = "イベントマスター",
                        description = "10個の市場イベントを経験",
                        points = 50,
                        category = AchievementCategory.Events,
                        maxProgress = 10,
                        showProgressBar = true,
                    },
                    new Achievement
                    {
                        id = "dragon_profit",
                        name = "竜の恩恵",
                        description = "ドラゴン討伐イベント中に5,000G以上の利益",
                        points = 60,
                        category = AchievementCategory.Events,
                        maxProgress = 1,
                        hidden = true, // 隠し実績
                    },
                    // その他の実績
                    new Achievement
                    {
                        id = "early_bird",
                        name = "早起きは三文の徳",
                        description = "30日連続で朝一番に店を開く",
                        points = 40,
                        category = AchievementCategory.Special,
                        maxProgress = 30,
                        showProgressBar = true,
                    },
                    new Achievement
                    {
                        id = "no_loss_week",
                        name = "完璧な一週間",
                        description = "1週間赤字なしで経営",
                        points = 50,
                        category = AchievementCategory.Special,
                        maxProgress = 7,
                        showProgressBar = true,
                    },
                };
            }

            // 進捗データを初期化
            foreach (var achievement in achievements)
            {
                if (!progressData.ContainsKey(achievement.id))
                {
                    progressData[achievement.id] = new AchievementProgress
                    {
                        currentProgress = 0,
                        unlocked = false,
                        unlockedDate = null,
                    };
                }
            }
        }

        private void SubscribeToEvents()
        {
            // 取引関連
            EventBus.Subscribe<TransactionCompletedEvent>(OnTransactionCompleted);
            EventBus.Subscribe<DailyAssetReportEvent>(OnDailyAssetReport);

            // ランク関連
            EventBus.Subscribe<RankChangedEvent>(OnRankChanged);

            // 季節関連
            EventBus.Subscribe<SeasonChangedEvent>(OnSeasonChanged);
            EventBus.Subscribe<DayChangedEvent>(OnDayChanged);

            // イベント関連
            EventBus.Subscribe<MarketEventTriggeredEvent>(OnMarketEvent);

            // アイテム関連
            EventBus.Subscribe<PurchaseCompletedEvent>(OnPurchaseCompleted);
        }

        private void UnsubscribeFromEvents()
        {
            EventBus.Unsubscribe<TransactionCompletedEvent>(OnTransactionCompleted);
            EventBus.Unsubscribe<DailyAssetReportEvent>(OnDailyAssetReport);
            EventBus.Unsubscribe<RankChangedEvent>(OnRankChanged);
            EventBus.Unsubscribe<SeasonChangedEvent>(OnSeasonChanged);
            EventBus.Unsubscribe<DayChangedEvent>(OnDayChanged);
            EventBus.Unsubscribe<MarketEventTriggeredEvent>(OnMarketEvent);
            EventBus.Unsubscribe<PurchaseCompletedEvent>(OnPurchaseCompleted);
        }

        private void OnTransactionCompleted(TransactionCompletedEvent e)
        {
            // 初めての取引
            UpdateProgress("first_trade", 1);

            // 取引回数
            IncrementProgress("trader_100", 1);

            // 利益マスター
            if (!e.IsPurchase && e.Profit >= 1000)
            {
                UpdateProgress("profit_master", 1);
            }

            // アイテム別の実績
            if (!e.IsPurchase)
            {
                switch (e.ItemType)
                {
                    case ItemType.Fruit:
                        IncrementProgress("fruit_expert", e.Quantity);
                        break;
                }
            }

            // 全アイテム取り扱い
            CheckAllItemsAchievement();
        }

        private void OnDailyAssetReport(DailyAssetReportEvent e)
        {
            // 資産関連の実績
            UpdateProgress("millionaire", e.Report.totalAssets);
            UpdateProgress("first_5000", e.Report.totalAssets);

            // 連続黒字の確認
            if (e.Report.dailyProfit >= 0)
            {
                IncrementProgress("no_loss_week", 1);
            }
            else
            {
                // 赤字が出たらリセット
                UpdateProgress("no_loss_week", 0);
            }
        }

        private void OnRankChanged(RankChangedEvent e)
        {
            if (e.IsRankUp)
            {
                switch (e.NewRank)
                {
                    case MerchantRank.Skilled:
                        UpdateProgress("rank_skilled", 1);
                        break;
                    case MerchantRank.Master:
                        UpdateProgress("rank_master", 1);
                        break;
                }
            }
        }

        private void OnSeasonChanged(SeasonChangedEvent e)
        {
            // 四季の商人
            var seasonsMask = GetSeasonsMask();
            UpdateProgress("four_seasons", seasonsMask.Count(s => s));

            // 冬の生存者（冬が終わったときにチェック）
            if (e.PreviousSeason == Season.Winter)
            {
                CheckWinterSurvivorAchievement();
            }
        }

        private void OnDayChanged(DayChangedEvent e)
        {
            // 早起きチェック（店を開いた時間を記録する必要あり）
            // TODO: 店を開いた時刻の記録が必要
        }

        private void OnMarketEvent(MarketEventTriggeredEvent e)
        {
            IncrementProgress("event_master", 1);

            // ドラゴン討伐イベントの処理
            if (e.EventName.Contains("ドラゴン"))
            {
                // TODO: イベント中の利益追跡が必要
            }
        }

        private void OnPurchaseCompleted(PurchaseCompletedEvent e)
        {
            // 宝石収集家
            if (InventorySystem.Instance != null)
            {
                var gemCount = InventorySystem.Instance.GetItemCount(ItemType.Gem);
                UpdateProgress("gem_collector", gemCount);
            }
        }

        /// <summary>
        /// 実績の進捗を更新
        /// </summary>
        public void UpdateProgress(string achievementId, float newProgress)
        {
            if (!progressData.ContainsKey(achievementId))
                return;

            var progress = progressData[achievementId];
            if (progress.unlocked)
                return; // 既に解除済み

            var achievement = GetAchievement(achievementId);
            if (achievement == null)
                return;

            float previousProgress = progress.currentProgress;
            progress.currentProgress = Mathf.Min(newProgress, achievement.maxProgress);

            // 進捗イベントを発行
            if (progress.currentProgress != previousProgress)
            {
                OnAchievementProgress?.Invoke(achievement, progress.currentProgress / achievement.maxProgress);
            }

            // 実績解除チェック
            if (progress.currentProgress >= achievement.maxProgress)
            {
                UnlockAchievement(achievementId);
            }
        }

        /// <summary>
        /// 実績の進捗を増加
        /// </summary>
        public void IncrementProgress(string achievementId, float amount)
        {
            if (!progressData.ContainsKey(achievementId))
                return;

            var currentProgress = progressData[achievementId].currentProgress;
            UpdateProgress(achievementId, currentProgress + amount);
        }

        /// <summary>
        /// 実績を解除
        /// </summary>
        private void UnlockAchievement(string achievementId)
        {
            if (!progressData.ContainsKey(achievementId))
                return;

            var progress = progressData[achievementId];
            if (progress.unlocked)
                return;

            var achievement = GetAchievement(achievementId);
            if (achievement == null)
                return;

            progress.unlocked = true;
            progress.unlockedDate = DateTime.Now;

            // 通知を表示
            if (UIManager.Instance != null)
            {
                UIManager.Instance.ShowNotification(
                    "実績解除！",
                    $"{achievement.name}\n{achievement.description}",
                    5f,
                    UIManager.NotificationType.Success
                );
            }

            // イベントを発行
            OnAchievementUnlocked?.Invoke(achievement);
            EventBus.Publish(new AchievementUnlockedEvent(achievement));

            ErrorHandler.LogInfo($"Achievement unlocked: {achievement.name} ({achievement.points}pts)", "Achievement");

            SaveAchievementData();
        }

        /// <summary>
        /// 特定の実績を取得
        /// </summary>
        public Achievement GetAchievement(string achievementId)
        {
            return achievements.Find(a => a.id == achievementId);
        }

        /// <summary>
        /// 実績の進捗情報を取得
        /// </summary>
        public AchievementProgress GetProgress(string achievementId)
        {
            return progressData.TryGetValue(achievementId, out var progress) ? progress : null;
        }

        /// <summary>
        /// カテゴリ別の実績リストを取得
        /// </summary>
        public List<Achievement> GetAchievementsByCategory(AchievementCategory category)
        {
            return achievements.FindAll(a => a.category == category);
        }

        /// <summary>
        /// 解除済み実績の総ポイントを取得
        /// </summary>
        public int GetTotalUnlockedPoints()
        {
            int totalPoints = 0;
            foreach (var achievement in achievements)
            {
                if (progressData.TryGetValue(achievement.id, out var progress) && progress.unlocked)
                {
                    totalPoints += achievement.points;
                }
            }
            return totalPoints;
        }

        /// <summary>
        /// 実績の完了率を取得
        /// </summary>
        public float GetCompletionPercentage()
        {
            if (achievements.Count == 0)
                return 0f;

            int unlockedCount = 0;
            foreach (var achievement in achievements)
            {
                if (progressData.TryGetValue(achievement.id, out var progress) && progress.unlocked)
                {
                    unlockedCount++;
                }
            }

            return (float)unlockedCount / achievements.Count * 100f;
        }

        /// <summary>
        /// 解放済みの実績IDリストを取得
        /// </summary>
        public List<string> GetUnlockedAchievements()
        {
            var unlockedList = new List<string>();
            foreach (var achievement in achievements)
            {
                if (progressData.TryGetValue(achievement.id, out var progress) && progress.unlocked)
                {
                    unlockedList.Add(achievement.id);
                }
            }
            return unlockedList;
        }

        /// <summary>
        /// 解放済み実績をロード
        /// </summary>
        public void LoadUnlockedAchievements(List<string> unlockedIds)
        {
            foreach (var id in unlockedIds)
            {
                if (progressData.ContainsKey(id))
                {
                    progressData[id].unlocked = true;
                    progressData[id].unlockedDate = DateTime.Now; // セーブデータに日付も含める場合はそちらを使用
                }
            }
        }

        // ヘルパーメソッド
        private bool[] GetSeasonsMask()
        {
            // TODO: 経験した季節を記録する仕組みが必要
            return new bool[4];
        }

        private void CheckWinterSurvivorAchievement()
        {
            // TODO: 冬の間の利益を追跡する仕組みが必要
        }

        private void CheckAllItemsAchievement()
        {
            // TODO: 取り扱ったアイテムの種類を記録する仕組みが必要
        }

        // セーブ/ロード
        private void SaveAchievementData()
        {
            foreach (var kvp in progressData)
            {
                string key = saveAchievementsGlobally
                    ? $"Achievement_{kvp.Key}"
                    : $"Save_{GameManager.Instance.PlayerData.PlayerName}_Achievement_{kvp.Key}";
                string json = JsonUtility.ToJson(kvp.Value);
                PlayerPrefs.SetString(key, json);
            }
            PlayerPrefs.Save();
        }

        private void LoadAchievementData()
        {
            foreach (var achievement in achievements)
            {
                string key = saveAchievementsGlobally
                    ? $"Achievement_{achievement.id}"
                    : $"Save_{GameManager.Instance?.PlayerData?.PlayerName}_Achievement_{achievement.id}";
                if (PlayerPrefs.HasKey(key))
                {
                    string json = PlayerPrefs.GetString(key);
                    progressData[achievement.id] = JsonUtility.FromJson<AchievementProgress>(json);
                }
            }
        }
    }

    /// <summary>
    /// 実績データ
    /// </summary>
    [Serializable]
    public class Achievement
    {
        public string id;
        public string name;
        public string description;
        public int points;
        public AchievementCategory category;
        public float maxProgress;
        public bool showProgressBar;
        public bool hidden; // 隠し実績
        public Sprite icon;
    }

    /// <summary>
    /// 実績カテゴリ
    /// </summary>
    public enum AchievementCategory
    {
        Trading, // 取引
        Wealth, // 資産
        Progression, // 進行
        Seasonal, // 季節
        Specialist, // 専門
        Events, // イベント
        Special, // 特別
    }

    /// <summary>
    /// 実績進捗データ
    /// </summary>
    [Serializable]
    public class AchievementProgress
    {
        public float currentProgress;
        public bool unlocked;
        public DateTime? unlockedDate;
    }

    /// <summary>
    /// 実績解除イベント
    /// </summary>
    public class AchievementUnlockedEvent : BaseGameEvent
    {
        public Achievement Achievement { get; }

        public AchievementUnlockedEvent(Achievement achievement)
        {
            Achievement = achievement;
        }
    }
}
