using System;
using System.Collections.Generic;
using MerchantTails.Core;
using MerchantTails.Data;
using MerchantTails.Core;
using MerchantTails.UI;
using UnityEngine;

namespace MerchantTails.Tutorial
{
    /// <summary>
    /// チュートリアル後のヒントシステム
    /// プレイヤーの行動に応じて適切なヒントを表示
    /// </summary>
    public class TutorialHintSystem : MonoBehaviour
    {
        private static TutorialHintSystem instance;
        public static TutorialHintSystem Instance => instance;

        [SerializeField]
        private float hintCooldown = 60f; // ヒント表示の最小間隔

        [SerializeField]
        private int maxHintsPerSession = 5; // 1セッションあたりの最大ヒント数

        private Dictionary<string, HintData> hints = new Dictionary<string, HintData>();
        private Queue<HintData> pendingHints = new Queue<HintData>();
        private HashSet<string> shownHints = new HashSet<string>();
        private float lastHintTime = -999f;
        private int hintsShownThisSession = 0;
        private bool isShowingHint = false;

        [Serializable]
        public class HintData
        {
            public string id;
            public string title;
            public string message;
            public HintPriority priority;
            public HintTrigger trigger;
            public int minDayToShow = 1;
            public int maxShowCount = 3;
            public float displayDuration = 5f;
            public bool requiresTutorialComplete = true;
        }

        public enum HintPriority
        {
            Low,
            Medium,
            High,
            Critical,
        }

        public enum HintTrigger
        {
            LowMoney,
            HighInventory,
            ItemDecay,
            PriceSpike,
            PriceDrop,
            NewSeason,
            FirstRankUp,
            LongIdle,
            BadTrade,
            GoodTrade,
            EventStart,
        }

        private void Awake()
        {
            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;

            InitializeHints();
            SubscribeToEvents();
        }

        private void OnDestroy()
        {
            if (instance == this)
            {
                instance = null;
            }
            UnsubscribeFromEvents();
        }

        private void InitializeHints()
        {
            // 基本的なヒントを定義
            AddHint(
                new HintData
                {
                    id = "low_money_warning",
                    title = "資金不足の警告",
                    message = "所持金が少なくなっています。安い商品を仕入れて、利益を確保しましょう。",
                    priority = HintPriority.High,
                    trigger = HintTrigger.LowMoney,
                    minDayToShow = 3,
                }
            );

            AddHint(
                new HintData
                {
                    id = "fruit_decay_warning",
                    title = "商品の劣化に注意",
                    message = "くだものは時間が経つと腐ってしまいます。早めに売りましょう！",
                    priority = HintPriority.Medium,
                    trigger = HintTrigger.ItemDecay,
                    minDayToShow = 2,
                }
            );

            AddHint(
                new HintData
                {
                    id = "price_spike_opportunity",
                    title = "価格高騰のチャンス",
                    message = "商品の価格が高騰しています！在庫があれば今が売り時です。",
                    priority = HintPriority.High,
                    trigger = HintTrigger.PriceSpike,
                    minDayToShow = 5,
                }
            );

            AddHint(
                new HintData
                {
                    id = "price_drop_buying",
                    title = "仕入れのチャンス",
                    message = "価格が下がっています。今のうちに仕入れておくと良いでしょう。",
                    priority = HintPriority.Medium,
                    trigger = HintTrigger.PriceDrop,
                    minDayToShow = 4,
                }
            );

            AddHint(
                new HintData
                {
                    id = "seasonal_change",
                    title = "季節の変わり目",
                    message = "新しい季節になりました。商品の需要が変化するので、価格に注目しましょう。",
                    priority = HintPriority.Medium,
                    trigger = HintTrigger.NewSeason,
                    minDayToShow = 1,
                }
            );

            AddHint(
                new HintData
                {
                    id = "rank_up_congrats",
                    title = "ランクアップおめでとう！",
                    message = "商人ランクが上がりました！新しい機能が解放されているか確認しましょう。",
                    priority = HintPriority.High,
                    trigger = HintTrigger.FirstRankUp,
                    minDayToShow = 1,
                    maxShowCount = 1,
                }
            );

            AddHint(
                new HintData
                {
                    id = "idle_reminder",
                    title = "商売を続けましょう",
                    message = "しばらく取引がありません。市場をチェックして、商機を見つけましょう。",
                    priority = HintPriority.Low,
                    trigger = HintTrigger.LongIdle,
                    minDayToShow = 7,
                }
            );

            AddHint(
                new HintData
                {
                    id = "bad_trade_advice",
                    title = "取引の見直し",
                    message = "損失が出ました。価格をよく確認してから取引しましょう。",
                    priority = HintPriority.Medium,
                    trigger = HintTrigger.BadTrade,
                    minDayToShow = 3,
                }
            );

            AddHint(
                new HintData
                {
                    id = "good_trade_praise",
                    title = "素晴らしい取引！",
                    message = "大きな利益を上げました！この調子で商売を続けましょう。",
                    priority = HintPriority.Low,
                    trigger = HintTrigger.GoodTrade,
                    minDayToShow = 2,
                }
            );

            AddHint(
                new HintData
                {
                    id = "event_notification",
                    title = "イベント発生",
                    message = "特別なイベントが発生しました。価格変動に注目しましょう！",
                    priority = HintPriority.High,
                    trigger = HintTrigger.EventStart,
                    minDayToShow = 1,
                }
            );
        }

        private void AddHint(HintData hint)
        {
            hints[hint.id] = hint;
        }

        #region Event Handling

        private void SubscribeToEvents()
        {
            EventBus.Subscribe<MoneyChangedEvent>(OnMoneyChanged);
            EventBus.Subscribe<ItemDecayedEvent>(OnItemDecayed);
            EventBus.Subscribe<PriceChangedEvent>(OnPriceChanged);
            EventBus.Subscribe<SeasonChangedEvent>(OnSeasonChanged);
            EventBus.Subscribe<RankChangedEvent>(OnRankChanged);
            EventBus.Subscribe<TransactionCompletedEvent>(OnTransactionCompleted);
            EventBus.Subscribe<GameEventTriggeredEvent>(OnGameEventTriggered);
        }

        private void UnsubscribeFromEvents()
        {
            EventBus.Unsubscribe<MoneyChangedEvent>(OnMoneyChanged);
            EventBus.Unsubscribe<ItemDecayedEvent>(OnItemDecayed);
            EventBus.Unsubscribe<PriceChangedEvent>(OnPriceChanged);
            EventBus.Unsubscribe<SeasonChangedEvent>(OnSeasonChanged);
            EventBus.Unsubscribe<RankChangedEvent>(OnRankChanged);
            EventBus.Unsubscribe<TransactionCompletedEvent>(OnTransactionCompleted);
            EventBus.Unsubscribe<GameEventTriggeredEvent>(OnGameEventTriggered);
        }

        private void OnMoneyChanged(MoneyChangedEvent e)
        {
            if (e.NewAmount < 500 && e.NewAmount < e.PreviousAmount)
            {
                TriggerHint(HintTrigger.LowMoney);
            }
        }

        private void OnItemDecayed(ItemDecayedEvent e)
        {
            if (e.ItemType == ItemType.Fruit && e.Quantity > 5)
            {
                TriggerHint(HintTrigger.ItemDecay);
            }
        }

        private void OnPriceChanged(PriceChangedEvent e)
        {
            if (e.ChangePercentage > 20f)
            {
                TriggerHint(HintTrigger.PriceSpike);
            }
            else if (e.ChangePercentage < -20f)
            {
                TriggerHint(HintTrigger.PriceDrop);
            }
        }

        private void OnSeasonChanged(SeasonChangedEvent e)
        {
            TriggerHint(HintTrigger.NewSeason);
        }

        private void OnRankChanged(RankChangedEvent e)
        {
            if (e.IsRankUp && e.PreviousRank == MerchantRank.Apprentice)
            {
                TriggerHint(HintTrigger.FirstRankUp);
            }
        }

        private void OnTransactionCompleted(TransactionCompletedEvent e)
        {
            if (e.Profit > 100f)
            {
                TriggerHint(HintTrigger.GoodTrade);
            }
            else if (e.Profit < -50f)
            {
                TriggerHint(HintTrigger.BadTrade);
            }
        }

        private void OnGameEventTriggered(GameEventTriggeredEvent e)
        {
            TriggerHint(HintTrigger.EventStart);
        }

        #endregion

        /// <summary>
        /// 特定のトリガーに基づいてヒントを表示
        /// </summary>
        public void TriggerHint(HintTrigger trigger)
        {
            // チュートリアルが完了していない場合はスキップ
            if (!IsTutorialCompleted())
                return;

            // ヒントのクールダウン中はスキップ
            if (Time.time - lastHintTime < hintCooldown)
                return;

            // セッションのヒント上限に達している場合はスキップ
            if (hintsShownThisSession >= maxHintsPerSession)
                return;

            // 該当するヒントを探す
            foreach (var hint in hints.Values)
            {
                if (hint.trigger == trigger && CanShowHint(hint))
                {
                    QueueHint(hint);
                    break;
                }
            }
        }

        private bool CanShowHint(HintData hint)
        {
            // チュートリアル完了が必要な場合
            if (hint.requiresTutorialComplete && !IsTutorialCompleted())
                return false;

            // 最小日数のチェック
            if (TimeManager.Instance != null && TimeManager.Instance.CurrentDay < hint.minDayToShow)
                return false;

            // 表示回数のチェック
            int showCount = PlayerPrefs.GetInt($"Hint_{hint.id}_Count", 0);
            if (showCount >= hint.maxShowCount)
                return false;

            return true;
        }

        private void QueueHint(HintData hint)
        {
            // 優先度に基づいてキューに追加
            pendingHints.Enqueue(hint);

            // ヒント表示処理を開始
            if (!isShowingHint)
            {
                ShowNextHint();
            }
        }

        private void ShowNextHint()
        {
            if (pendingHints.Count == 0)
            {
                isShowingHint = false;
                return;
            }

            isShowingHint = true;
            HintData hint = pendingHints.Dequeue();

            // ヒントを表示
            DisplayHint(hint);

            // 統計を更新
            lastHintTime = Time.time;
            hintsShownThisSession++;

            // 表示回数を記録
            int showCount = PlayerPrefs.GetInt($"Hint_{hint.id}_Count", 0);
            PlayerPrefs.SetInt($"Hint_{hint.id}_Count", showCount + 1);
            PlayerPrefs.Save();

            // 次のヒントを表示
            Invoke(nameof(ShowNextHint), hint.displayDuration);
        }

        private void DisplayHint(HintData hint)
        {
            ErrorHandler.LogInfo($"Displaying hint: {hint.id}", "HintSystem");

            // UI経由でヒントを表示
            if (UIManager.Instance != null)
            {
                UIManager.Instance.ShowNotification(
                    hint.title,
                    hint.message,
                    hint.displayDuration,
                    GetNotificationTypeFromPriority(hint.priority)
                );
            }
        }

        private UIManager.NotificationType GetNotificationTypeFromPriority(HintPriority priority)
        {
            return priority switch
            {
                HintPriority.Critical => UIManager.NotificationType.Error,
                HintPriority.High => UIManager.NotificationType.Warning,
                HintPriority.Medium => UIManager.NotificationType.Info,
                HintPriority.Low => UIManager.NotificationType.Success,
                _ => UIManager.NotificationType.Info,
            };
        }

        private bool IsTutorialCompleted()
        {
            return GameManager.Instance != null && GameManager.Instance.IsTutorialCompleted;
        }

        /// <summary>
        /// セッションをリセット（新しいゲーム開始時など）
        /// </summary>
        public void ResetSession()
        {
            hintsShownThisSession = 0;
            pendingHints.Clear();
            isShowingHint = false;
        }

        /// <summary>
        /// すべてのヒント履歴をクリア
        /// </summary>
        public void ClearHintHistory()
        {
            foreach (var hint in hints.Values)
            {
                PlayerPrefs.DeleteKey($"Hint_{hint.id}_Count");
            }
            PlayerPrefs.Save();
            shownHints.Clear();
        }
    }
}
