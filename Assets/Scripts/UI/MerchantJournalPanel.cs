using System;
using System.Collections.Generic;
using System.Linq;
using MerchantTails.Core;
using MerchantTails.Data;
using MerchantTails.Core;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.UI
{
    /// <summary>
    /// 商人手帳UI - 取引履歴とビジネス分析
    /// </summary>
    public class MerchantJournalPanel : UIPanel
    {
        [Header("Navigation")]
        [SerializeField]
        private Button dayViewButton;

        [SerializeField]
        private Button weekViewButton;

        [SerializeField]
        private Button monthViewButton;

        [SerializeField]
        private Button allTimeViewButton;

        [SerializeField]
        private GameObject viewButtonContainer;

        [Header("Summary Section")]
        [SerializeField]
        private TextMeshProUGUI periodLabel;

        [SerializeField]
        private TextMeshProUGUI totalRevenueText;

        [SerializeField]
        private TextMeshProUGUI totalCostText;

        [SerializeField]
        private TextMeshProUGUI netProfitText;

        [SerializeField]
        private TextMeshProUGUI profitMarginText;

        [SerializeField]
        private TextMeshProUGUI transactionCountText;

        [Header("Statistics")]
        [SerializeField]
        private Transform statisticsContainer;

        [SerializeField]
        private GameObject statisticsItemPrefab;
        private Dictionary<ItemType, StatisticsItem> itemStatistics = new Dictionary<ItemType, StatisticsItem>();

        [Header("Transaction History")]
        [SerializeField]
        private ScrollRect transactionScrollRect;

        [SerializeField]
        private Transform transactionListContainer;

        [SerializeField]
        private GameObject transactionItemPrefab;

        [SerializeField]
        private int maxDisplayedTransactions = 100;

        [Header("Visual Elements")]
        [SerializeField]
        private Image profitTrendArrow;

        [SerializeField]
        private Color profitColor = Color.green;

        [SerializeField]
        private Color lossColor = Color.red;

        [SerializeField]
        private Color neutralColor = Color.gray;

        private JournalViewMode currentViewMode = JournalViewMode.Day;
        private List<TransactionRecord> allTransactions = new List<TransactionRecord>();
        private List<TransactionRecord> filteredTransactions = new List<TransactionRecord>();
        private List<TransactionHistoryItem> transactionUIItems = new List<TransactionHistoryItem>();

        public enum JournalViewMode
        {
            Day,
            Week,
            Month,
            AllTime,
        }

        protected override void Awake()
        {
            base.Awake();
            SetupButtons();
            LoadTransactionHistory();
        }

        protected override void OnEnable()
        {
            base.OnEnable();
            SubscribeToEvents();
            RefreshDisplay();
        }

        protected override void OnDisable()
        {
            base.OnDisable();
            UnsubscribeFromEvents();
        }

        private void SetupButtons()
        {
            if (dayViewButton != null)
                dayViewButton.onClick.AddListener(() => SetViewMode(JournalViewMode.Day));

            if (weekViewButton != null)
                weekViewButton.onClick.AddListener(() => SetViewMode(JournalViewMode.Week));

            if (monthViewButton != null)
                monthViewButton.onClick.AddListener(() => SetViewMode(JournalViewMode.Month));

            if (allTimeViewButton != null)
                allTimeViewButton.onClick.AddListener(() => SetViewMode(JournalViewMode.AllTime));
        }

        private void SubscribeToEvents()
        {
            EventBus.Subscribe<TransactionCompletedEvent>(OnTransactionCompleted);
            EventBus.Subscribe<DayChangedEvent>(OnDayChanged);
        }

        private void UnsubscribeFromEvents()
        {
            EventBus.Unsubscribe<TransactionCompletedEvent>(OnTransactionCompleted);
            EventBus.Unsubscribe<DayChangedEvent>(OnDayChanged);
        }

        private void OnTransactionCompleted(TransactionCompletedEvent e)
        {
            // 新しい取引を記録
            var record = new TransactionRecord
            {
                timestamp = DateTime.Now,
                itemType = e.ItemType,
                quantity = e.Quantity,
                unitPrice = e.UnitPrice,
                totalPrice = e.TotalPrice,
                isPurchase = e.IsPurchase,
                profit = e.Profit,
                day = TimeManager.Instance.CurrentDay,
                season = TimeManager.Instance.CurrentSeason,
            };

            allTransactions.Add(record);
            SaveTransactionHistory();

            // 表示を更新
            if (isActiveAndEnabled)
            {
                RefreshDisplay();
            }
        }

        private void OnDayChanged(DayChangedEvent e)
        {
            if (isActiveAndEnabled && currentViewMode == JournalViewMode.Day)
            {
                RefreshDisplay();
            }
        }

        public void SetViewMode(JournalViewMode mode)
        {
            currentViewMode = mode;
            UpdateViewButtons();
            RefreshDisplay();
        }

        private void UpdateViewButtons()
        {
            // ボタンの選択状態を更新
            if (dayViewButton != null)
                dayViewButton.interactable = currentViewMode != JournalViewMode.Day;

            if (weekViewButton != null)
                weekViewButton.interactable = currentViewMode != JournalViewMode.Week;

            if (monthViewButton != null)
                monthViewButton.interactable = currentViewMode != JournalViewMode.Month;

            if (allTimeViewButton != null)
                allTimeViewButton.interactable = currentViewMode != JournalViewMode.AllTime;
        }

        private void RefreshDisplay()
        {
            FilterTransactions();
            UpdateSummary();
            UpdateStatistics();
            UpdateTransactionList();
        }

        private void FilterTransactions()
        {
            filteredTransactions.Clear();

            int currentDay = TimeManager.Instance != null ? TimeManager.Instance.CurrentDay : 1;

            switch (currentViewMode)
            {
                case JournalViewMode.Day:
                    filteredTransactions = allTransactions.Where(t => t.day == currentDay).ToList();
                    break;

                case JournalViewMode.Week:
                    int weekStart = Math.Max(1, currentDay - 6);
                    filteredTransactions = allTransactions
                        .Where(t => t.day >= weekStart && t.day <= currentDay)
                        .ToList();
                    break;

                case JournalViewMode.Month:
                    int monthStart = Math.Max(1, currentDay - 29);
                    filteredTransactions = allTransactions
                        .Where(t => t.day >= monthStart && t.day <= currentDay)
                        .ToList();
                    break;

                case JournalViewMode.AllTime:
                    filteredTransactions = new List<TransactionRecord>(allTransactions);
                    break;
            }

            // 最新の取引を先頭に
            filteredTransactions.Sort((a, b) => b.timestamp.CompareTo(a.timestamp));
        }

        private void UpdateSummary()
        {
            // 期間ラベル
            if (periodLabel != null)
            {
                periodLabel.text = GetPeriodLabel();
            }

            // 収益計算
            float totalRevenue = 0f;
            float totalCost = 0f;

            foreach (var transaction in filteredTransactions)
            {
                if (transaction.isPurchase)
                {
                    totalCost += transaction.totalPrice;
                }
                else
                {
                    totalRevenue += transaction.totalPrice;
                }
            }

            float netProfit = totalRevenue - totalCost;
            float profitMargin = totalRevenue > 0 ? (netProfit / totalRevenue) * 100f : 0f;

            // テキスト更新
            if (totalRevenueText != null)
                totalRevenueText.text = $"{totalRevenue:N0}G";

            if (totalCostText != null)
                totalCostText.text = $"{totalCost:N0}G";

            if (netProfitText != null)
            {
                netProfitText.text = $"{netProfit:N0}G";
                netProfitText.color = netProfit >= 0 ? profitColor : lossColor;
            }

            if (profitMarginText != null)
            {
                profitMarginText.text = $"{profitMargin:F1}%";
                profitMarginText.color = profitMargin >= 0 ? profitColor : lossColor;
            }

            if (transactionCountText != null)
                transactionCountText.text = $"{filteredTransactions.Count} 件";

            // トレンド矢印
            UpdateProfitTrend(netProfit);
        }

        private string GetPeriodLabel()
        {
            int currentDay = TimeManager.Instance != null ? TimeManager.Instance.CurrentDay : 1;
            Season currentSeason = TimeManager.Instance != null ? TimeManager.Instance.CurrentSeason : Season.Spring;

            return currentViewMode switch
            {
                JournalViewMode.Day => $"Day {currentDay} ({currentSeason})",
                JournalViewMode.Week => $"Week (Day {Math.Max(1, currentDay - 6)} - {currentDay})",
                JournalViewMode.Month => $"Month (Day {Math.Max(1, currentDay - 29)} - {currentDay})",
                JournalViewMode.AllTime => "All Time",
                _ => "",
            };
        }

        private void UpdateProfitTrend(float netProfit)
        {
            if (profitTrendArrow == null)
                return;

            if (netProfit > 0)
            {
                profitTrendArrow.transform.rotation = Quaternion.Euler(0, 0, 45); // 上向き
                profitTrendArrow.color = profitColor;
            }
            else if (netProfit < 0)
            {
                profitTrendArrow.transform.rotation = Quaternion.Euler(0, 0, -45); // 下向き
                profitTrendArrow.color = lossColor;
            }
            else
            {
                profitTrendArrow.transform.rotation = Quaternion.Euler(0, 0, 0); // 水平
                profitTrendArrow.color = neutralColor;
            }
        }

        private void UpdateStatistics()
        {
            // アイテムタイプ別の統計を計算
            Dictionary<ItemType, ItemStatistics> stats = new Dictionary<ItemType, ItemStatistics>();

            foreach (var transaction in filteredTransactions)
            {
                if (!stats.ContainsKey(transaction.itemType))
                {
                    stats[transaction.itemType] = new ItemStatistics();
                }

                var stat = stats[transaction.itemType];

                if (transaction.isPurchase)
                {
                    stat.totalPurchased += transaction.quantity;
                    stat.totalCost += transaction.totalPrice;
                }
                else
                {
                    stat.totalSold += transaction.quantity;
                    stat.totalRevenue += transaction.totalPrice;
                }
            }

            // UI更新
            if (statisticsContainer != null && statisticsItemPrefab != null)
            {
                // 既存のアイテムをクリア
                foreach (var kvp in itemStatistics)
                {
                    if (kvp.Value != null && kvp.Value.gameObject != null)
                    {
                        Destroy(kvp.Value.gameObject);
                    }
                }
                itemStatistics.Clear();

                // 新しい統計アイテムを作成
                foreach (var kvp in stats)
                {
                    var itemGO = Instantiate(statisticsItemPrefab, statisticsContainer);
                    var statItem = itemGO.GetComponent<StatisticsItem>();

                    if (statItem != null)
                    {
                        statItem.Setup(kvp.Key, kvp.Value);
                        itemStatistics[kvp.Key] = statItem;
                    }
                }
            }
        }

        private void UpdateTransactionList()
        {
            if (transactionListContainer == null || transactionItemPrefab == null)
                return;

            // 既存のアイテムをクリア
            foreach (var item in transactionUIItems)
            {
                if (item != null && item.gameObject != null)
                {
                    Destroy(item.gameObject);
                }
            }
            transactionUIItems.Clear();

            // 表示する取引数を制限
            int displayCount = Math.Min(filteredTransactions.Count, maxDisplayedTransactions);

            // 新しいアイテムを作成
            for (int i = 0; i < displayCount; i++)
            {
                var transaction = filteredTransactions[i];
                var itemGO = Instantiate(transactionItemPrefab, transactionListContainer);
                var historyItem = itemGO.GetComponent<TransactionHistoryItem>();

                if (historyItem != null)
                {
                    historyItem.Setup(transaction);
                    transactionUIItems.Add(historyItem);
                }
            }

            // スクロール位置をリセット
            if (transactionScrollRect != null)
            {
                transactionScrollRect.verticalNormalizedPosition = 1f;
            }
        }

        private void LoadTransactionHistory()
        {
            // TODO: セーブシステムから取引履歴を読み込む
            // 今は仮実装
            allTransactions.Clear();
        }

        private void SaveTransactionHistory()
        {
            // TODO: セーブシステムに取引履歴を保存
            // 今は仮実装
        }

        protected override void OnShow()
        {
            base.OnShow();
            RefreshDisplay();
        }
    }

    /// <summary>
    /// 取引記録データ
    /// </summary>
    [Serializable]
    public class TransactionRecord
    {
        public DateTime timestamp;
        public ItemType itemType;
        public int quantity;
        public float unitPrice;
        public float totalPrice;
        public bool isPurchase;
        public float profit;
        public int day;
        public Season season;
    }

    /// <summary>
    /// アイテム統計データ
    /// </summary>
    public class ItemStatistics
    {
        public int totalPurchased;
        public int totalSold;
        public float totalCost;
        public float totalRevenue;

        public float NetProfit => totalRevenue - totalCost;
        public float AverageBuyPrice => totalPurchased > 0 ? totalCost / totalPurchased : 0f;
        public float AverageSellPrice => totalSold > 0 ? totalRevenue / totalSold : 0f;
    }
}
