using System.Collections.Generic;
using MerchantTails.Core;
using MerchantTails.Data;
using MerchantTails.Core;
using MerchantTails.Market;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.UI
{
    /// <summary>
    /// マーケット分析画面のUI制御
    /// 価格チャート、相場動向、予測情報を表示
    /// </summary>
    public class MarketAnalysisPanel : UIPanel
    {
        [Header("Market Overview")]
        [SerializeField]
        private Text marketTitleText;

        [SerializeField]
        private Text currentSeasonText;

        [SerializeField]
        private Text marketConditionText;

        [Header("Item Selection")]
        [SerializeField]
        private Transform itemTabContainer;

        [SerializeField]
        private GameObject itemTabPrefab;

        [SerializeField]
        private ItemType selectedItemType = ItemType.Fruit;

        [Header("Price Chart")]
        [SerializeField]
        private RectTransform chartContainer;

        [SerializeField]
        private LineRenderer priceLineRenderer;

        [SerializeField]
        private Text currentPriceText;

        [SerializeField]
        private Text priceChangeText;

        [SerializeField]
        private Text averagePriceText;

        [Header("Chart Settings")]
        [SerializeField]
        private int chartDataPoints = 30;

        [SerializeField]
        private float chartUpdateInterval = 2f;

        [SerializeField]
        private Color priceLineColor = Color.green;

        [SerializeField]
        private Color gridLineColor = new Color(0.3f, 0.3f, 0.3f, 0.5f);

        [Header("Market Info")]
        [SerializeField]
        private Text volatilityText;

        [SerializeField]
        private Text trendText;

        [SerializeField]
        private Text predictionText;

        [SerializeField]
        private GameObject predictionPanel;

        [Header("Quick Actions")]
        [SerializeField]
        private Button buyButton;

        [SerializeField]
        private Button sellButton;

        [SerializeField]
        private Button inventoryButton;

        [SerializeField]
        private Button backButton;

        private Dictionary<ItemType, List<float>> priceHistory = new Dictionary<ItemType, List<float>>();
        private List<GameObject> itemTabs = new List<GameObject>();
        private List<GameObject> chartElements = new List<GameObject>();
        private float chartUpdateTimer = 0f;

        protected override void OnInitialize()
        {
            InitializePriceHistory();
            SetupButtons();
            SetupEventListeners();
            CreateItemTabs();
        }

        protected override void OnShow()
        {
            LogUIAction("Market Analysis panel shown");
            RefreshMarketDisplay();
            UpdateChart();
        }

        protected override void OnHide()
        {
            LogUIAction("Market Analysis panel hidden");
        }

        private void Update()
        {
            if (IsVisible)
            {
                chartUpdateTimer += Time.deltaTime;
                if (chartUpdateTimer >= chartUpdateInterval)
                {
                    UpdatePriceData();
                    UpdateChart();
                    chartUpdateTimer = 0f;
                }
            }
        }

        private void InitializePriceHistory()
        {
            // 各アイテムタイプの価格履歴を初期化
            foreach (ItemType itemType in System.Enum.GetValues(typeof(ItemType)))
            {
                priceHistory[itemType] = new List<float>();

                // 初期データを生成（過去30ポイント分）
                float basePrice = MarketSystem.Instance?.GetBasePrice(itemType) ?? 100f;
                for (int i = 0; i < chartDataPoints; i++)
                {
                    float variation = Random.Range(0.8f, 1.2f);
                    priceHistory[itemType].Add(basePrice * variation);
                }
            }
        }

        private void SetupButtons()
        {
            if (buyButton != null)
                buyButton.onClick.AddListener(OnBuyPressed);

            if (sellButton != null)
                sellButton.onClick.AddListener(OnSellPressed);

            if (inventoryButton != null)
                inventoryButton.onClick.AddListener(OnInventoryPressed);

            if (backButton != null)
                backButton.onClick.AddListener(OnBackPressed);
        }

        private void SetupEventListeners()
        {
            EventBus.Subscribe<PriceChangedEvent>(OnPriceChanged);
            EventBus.Subscribe<SeasonChangedEvent>(OnSeasonChanged);
            EventBus.Subscribe<MarketEventTriggeredEvent>(OnMarketEventTriggered);
        }

        private void CreateItemTabs()
        {
            if (itemTabContainer == null || itemTabPrefab == null)
                return;

            foreach (ItemType itemType in System.Enum.GetValues(typeof(ItemType)))
            {
                var tabGO = Instantiate(itemTabPrefab, itemTabContainer);
                var tabButton = tabGO.GetComponent<Button>();
                var tabText = tabGO.GetComponentInChildren<Text>();

                if (tabText != null)
                    tabText.text = GetItemDisplayName(itemType);

                if (tabButton != null)
                {
                    ItemType capturedType = itemType;
                    tabButton.onClick.AddListener(() => OnItemTabSelected(capturedType));
                }

                itemTabs.Add(tabGO);

                // 初期選択状態を設定
                if (itemType == selectedItemType)
                {
                    SetTabActive(tabGO, true);
                }
                else
                {
                    SetTabActive(tabGO, false);
                }
            }
        }

        private void RefreshMarketDisplay()
        {
            UpdateMarketOverview();
            UpdatePriceInfo();
            UpdateMarketCondition();
            UpdatePrediction();
        }

        private void UpdateMarketOverview()
        {
            if (marketTitleText != null)
                marketTitleText.text = "マーケット分析";

            if (currentSeasonText != null && TimeManager.Instance != null)
            {
                currentSeasonText.text = $"現在の季節: {GetSeasonDisplayName(TimeManager.Instance.CurrentSeason)}";
            }
        }

        private void UpdatePriceInfo()
        {
            if (MarketSystem.Instance == null)
                return;

            float currentPrice = MarketSystem.Instance.GetCurrentPrice(selectedItemType);
            var priceHistory = MarketSystem.Instance.GetPriceHistory(selectedItemType);

            // 現在価格
            if (currentPriceText != null)
                currentPriceText.text = $"現在価格: {currentPrice:F0}G";

            // 価格変動
            if (priceChangeText != null && priceHistory.Count > 1)
            {
                float previousPrice = priceHistory[priceHistory.Count - 2];
                float change = currentPrice - previousPrice;
                float changePercent = (change / previousPrice) * 100f;

                string changeColor = change >= 0 ? "#00FF00" : "#FF0000";
                string changeSymbol = change >= 0 ? "+" : "";
                priceChangeText.text =
                    $"<color={changeColor}>{changeSymbol}{change:F0}G ({changeSymbol}{changePercent:F1}%)</color>";
            }

            // 平均価格
            if (averagePriceText != null)
            {
                float averagePrice = CalculateAveragePrice(selectedItemType);
                averagePriceText.text = $"平均価格: {averagePrice:F0}G";
            }
        }

        private void UpdateMarketCondition()
        {
            if (marketConditionText == null)
                return;

            // 市場状況の分析
            float volatility = CalculateVolatility(selectedItemType);
            string condition = volatility switch
            {
                > 0.2f => "非常に不安定",
                > 0.1f => "やや不安定",
                > 0.05f => "安定",
                _ => "非常に安定",
            };

            marketConditionText.text = $"市場状況: {condition}";

            // ボラティリティ表示
            if (volatilityText != null)
                volatilityText.text = $"変動率: {volatility * 100f:F1}%";

            // トレンド表示
            if (trendText != null)
            {
                string trend = AnalyzeTrend(selectedItemType);
                trendText.text = $"トレンド: {trend}";
            }
        }

        private void UpdatePrediction()
        {
            // ランクに応じて予測情報を表示
            var playerData = GameManager.Instance?.GetPlayerData();
            if (playerData == null || predictionPanel == null)
                return;

            bool showPrediction = playerData.CurrentRank >= MerchantRank.Skilled;
            predictionPanel.SetActive(showPrediction);

            if (showPrediction && predictionText != null)
            {
                string prediction = GeneratePrediction(selectedItemType);
                predictionText.text = prediction;
            }
        }

        private void UpdateChart()
        {
            ClearChartElements();
            DrawGrid();
            DrawPriceLine();
            DrawLabels();
        }

        private void DrawGrid()
        {
            if (chartContainer == null)
                return;

            // 横線（価格レベル）
            int horizontalLines = 5;
            float height = chartContainer.rect.height;

            for (int i = 0; i <= horizontalLines; i++)
            {
                float y = (height / horizontalLines) * i;
                CreateGridLine(new Vector2(0, y), new Vector2(chartContainer.rect.width, y), true);
            }

            // 縦線（時間軸）
            int verticalLines = 6;
            float width = chartContainer.rect.width;

            for (int i = 0; i <= verticalLines; i++)
            {
                float x = (width / verticalLines) * i;
                CreateGridLine(new Vector2(x, 0), new Vector2(x, height), false);
            }
        }

        private void CreateGridLine(Vector2 start, Vector2 end, bool isHorizontal)
        {
            var lineGO = new GameObject(isHorizontal ? "HGridLine" : "VGridLine");
            lineGO.transform.SetParent(chartContainer, false);

            var lineImage = lineGO.AddComponent<Image>();
            lineImage.color = gridLineColor;

            var rectTransform = lineGO.GetComponent<RectTransform>();
            rectTransform.anchorMin = Vector2.zero;
            rectTransform.anchorMax = Vector2.zero;
            rectTransform.pivot = new Vector2(0, 0);

            if (isHorizontal)
            {
                rectTransform.anchoredPosition = start;
                rectTransform.sizeDelta = new Vector2(end.x - start.x, 1);
            }
            else
            {
                rectTransform.anchoredPosition = start;
                rectTransform.sizeDelta = new Vector2(1, end.y - start.y);
            }

            chartElements.Add(lineGO);
        }

        private void DrawPriceLine()
        {
            if (!priceHistory.ContainsKey(selectedItemType))
                return;

            var prices = priceHistory[selectedItemType];
            if (prices.Count < 2)
                return;

            // 価格の最小値と最大値を取得
            float minPrice = float.MaxValue;
            float maxPrice = float.MinValue;

            foreach (float price in prices)
            {
                minPrice = Mathf.Min(minPrice, price);
                maxPrice = Mathf.Max(maxPrice, price);
            }

            float priceRange = maxPrice - minPrice;
            if (priceRange < 1f)
                priceRange = 1f;

            // 線を描画
            float width = chartContainer.rect.width;
            float height = chartContainer.rect.height;
            float pointSpacing = width / (prices.Count - 1);

            for (int i = 0; i < prices.Count - 1; i++)
            {
                float x1 = i * pointSpacing;
                float x2 = (i + 1) * pointSpacing;

                float y1 = ((prices[i] - minPrice) / priceRange) * height;
                float y2 = ((prices[i + 1] - minPrice) / priceRange) * height;

                CreatePriceLine(new Vector2(x1, y1), new Vector2(x2, y2));

                // データポイントマーカー
                if (i % 5 == 0)
                {
                    CreateDataPoint(new Vector2(x1, y1), prices[i]);
                }
            }
        }

        private void CreatePriceLine(Vector2 start, Vector2 end)
        {
            var lineGO = new GameObject("PriceLine");
            lineGO.transform.SetParent(chartContainer, false);

            var lineImage = lineGO.AddComponent<Image>();
            lineImage.color = priceLineColor;

            var rectTransform = lineGO.GetComponent<RectTransform>();
            rectTransform.anchorMin = Vector2.zero;
            rectTransform.anchorMax = Vector2.zero;
            rectTransform.pivot = new Vector2(0, 0.5f);

            // 線の角度と長さを計算
            Vector2 direction = end - start;
            float angle = Mathf.Atan2(direction.y, direction.x) * Mathf.Rad2Deg;
            float length = direction.magnitude;

            rectTransform.anchoredPosition = start;
            rectTransform.sizeDelta = new Vector2(length, 2);
            rectTransform.rotation = Quaternion.Euler(0, 0, angle);

            chartElements.Add(lineGO);
        }

        private void CreateDataPoint(Vector2 position, float price)
        {
            var pointGO = new GameObject("DataPoint");
            pointGO.transform.SetParent(chartContainer, false);

            var pointImage = pointGO.AddComponent<Image>();
            pointImage.color = Color.white;

            var rectTransform = pointGO.GetComponent<RectTransform>();
            rectTransform.anchorMin = Vector2.zero;
            rectTransform.anchorMax = Vector2.zero;
            rectTransform.pivot = new Vector2(0.5f, 0.5f);
            rectTransform.anchoredPosition = position;
            rectTransform.sizeDelta = new Vector2(6, 6);

            chartElements.Add(pointGO);
        }

        private void DrawLabels()
        {
            // 価格ラベルと時間ラベルの描画（簡易実装）
        }

        private void ClearChartElements()
        {
            foreach (var element in chartElements)
            {
                if (element != null)
                    Destroy(element);
            }
            chartElements.Clear();
        }

        private void UpdatePriceData()
        {
            if (MarketSystem.Instance == null)
                return;

            // 現在の価格を履歴に追加
            float currentPrice = MarketSystem.Instance.GetCurrentPrice(selectedItemType);

            if (!priceHistory.ContainsKey(selectedItemType))
                priceHistory[selectedItemType] = new List<float>();

            var history = priceHistory[selectedItemType];
            history.Add(currentPrice);

            // 古いデータを削除
            while (history.Count > chartDataPoints)
            {
                history.RemoveAt(0);
            }
        }

        // ヘルパーメソッド
        private string GetItemDisplayName(ItemType itemType)
        {
            return itemType switch
            {
                ItemType.Fruit => "くだもの",
                ItemType.Potion => "ポーション",
                ItemType.Weapon => "武器",
                ItemType.Accessory => "アクセサリー",
                ItemType.MagicBook => "魔法書",
                ItemType.Gem => "宝石",
                _ => "不明",
            };
        }

        private string GetSeasonDisplayName(Season season)
        {
            return season switch
            {
                Season.Spring => "春",
                Season.Summer => "夏",
                Season.Autumn => "秋",
                Season.Winter => "冬",
                _ => "不明",
            };
        }

        private float CalculateAveragePrice(ItemType itemType)
        {
            if (!priceHistory.ContainsKey(itemType))
                return 0f;

            var prices = priceHistory[itemType];
            if (prices.Count == 0)
                return 0f;

            float sum = 0f;
            foreach (float price in prices)
            {
                sum += price;
            }

            return sum / prices.Count;
        }

        private float CalculateVolatility(ItemType itemType)
        {
            if (!priceHistory.ContainsKey(itemType))
                return 0f;

            var prices = priceHistory[itemType];
            if (prices.Count < 2)
                return 0f;

            float average = CalculateAveragePrice(itemType);
            float sumSquaredDiff = 0f;

            foreach (float price in prices)
            {
                float diff = price - average;
                sumSquaredDiff += diff * diff;
            }

            float variance = sumSquaredDiff / prices.Count;
            float standardDeviation = Mathf.Sqrt(variance);

            return standardDeviation / average; // 変動係数
        }

        private string AnalyzeTrend(ItemType itemType)
        {
            if (!priceHistory.ContainsKey(itemType))
                return "不明";

            var prices = priceHistory[itemType];
            if (prices.Count < 10)
                return "データ不足";

            // 簡易的なトレンド分析（移動平均の比較）
            float recentAverage = 0f;
            float oldAverage = 0f;

            int recentCount = 5;
            for (int i = prices.Count - recentCount; i < prices.Count; i++)
            {
                recentAverage += prices[i];
            }
            recentAverage /= recentCount;

            for (int i = 0; i < recentCount; i++)
            {
                oldAverage += prices[i];
            }
            oldAverage /= recentCount;

            float trendPercent = ((recentAverage - oldAverage) / oldAverage) * 100f;

            return trendPercent switch
            {
                > 10f => "強い上昇傾向",
                > 3f => "上昇傾向",
                < -10f => "強い下降傾向",
                < -3f => "下降傾向",
                _ => "横ばい",
            };
        }

        private string GeneratePrediction(ItemType itemType)
        {
            // ランクに応じた予測精度
            var playerData = GameManager.Instance?.GetPlayerData();
            if (playerData == null)
                return "予測不可";

            string trend = AnalyzeTrend(itemType);
            float volatility = CalculateVolatility(itemType);

            // 季節要因
            var season = TimeManager.Instance?.CurrentSeason ?? Season.Spring;
            string seasonalFactor = GetSeasonalFactor(itemType, season);

            // 予測文の生成
            return playerData.CurrentRank switch
            {
                MerchantRank.Master =>
                    $"詳細予測: {trend}が継続。{seasonalFactor}。変動率{volatility * 100f:F1}%で推移予想。",
                MerchantRank.Veteran => $"予測: {trend}の可能性大。{seasonalFactor}。",
                MerchantRank.Skilled => $"簡易予測: {trend}の兆候あり。",
                _ => "予測機能はまだ使用できません",
            };
        }

        private string GetSeasonalFactor(ItemType itemType, Season season)
        {
            return (itemType, season) switch
            {
                (ItemType.Fruit, Season.Summer) => "夏は需要増で価格上昇",
                (ItemType.Fruit, Season.Winter) => "冬は供給減で価格高騰",
                (ItemType.Potion, Season.Spring) => "春は冒険者増で需要拡大",
                (ItemType.Weapon, Season.Autumn) => "秋は戦争需要で価格上昇",
                _ => "季節要因は標準的",
            };
        }

        // UIイベントハンドラー
        private void OnItemTabSelected(ItemType itemType)
        {
            selectedItemType = itemType;

            // タブの選択状態を更新
            for (int i = 0; i < itemTabs.Count; i++)
            {
                SetTabActive(itemTabs[i], (ItemType)i == itemType);
            }

            RefreshMarketDisplay();
            UpdateChart();
        }

        private void SetTabActive(GameObject tab, bool active)
        {
            var tabImage = tab.GetComponent<Image>();
            if (tabImage != null)
            {
                tabImage.color = active ? Color.white : new Color(0.7f, 0.7f, 0.7f);
            }
        }

        private void OnBuyPressed()
        {
            LogUIAction($"Buy {selectedItemType} pressed");
            ShowTradeDialog(true);
        }

        private void OnSellPressed()
        {
            LogUIAction($"Sell {selectedItemType} pressed");
            ShowTradeDialog(false);
        }

        private void OnInventoryPressed()
        {
            LogUIAction("Inventory button pressed");
            UIManager.Instance.ShowPanel(UIType.Inventory);
        }

        private void ShowTradeDialog(bool isBuying)
        {
            // 取引ダイアログの表示（簡易実装）
            UIManager.Instance.ShowModal(
                UIType.TradeConfirmation,
                (confirmed) =>
                {
                    if (confirmed)
                    {
                        ExecuteTrade(selectedItemType, isBuying);
                    }
                }
            );
        }

        private void ExecuteTrade(ItemType itemType, bool isBuying)
        {
            // 取引実行（将来実装）
            LogUIAction($"Trade executed: {(isBuying ? "Buy" : "Sell")} {itemType}");
        }

        // イベントハンドラー
        private void OnPriceChanged(PriceChangedEvent evt)
        {
            if (evt.ItemType == selectedItemType)
            {
                UpdatePriceInfo();
            }
        }

        private void OnSeasonChanged(SeasonChangedEvent evt)
        {
            RefreshMarketDisplay();
        }

        private void OnMarketEventTriggered(MarketEventTriggeredEvent evt)
        {
            // 市場イベント発生時の処理
            RefreshMarketDisplay();
        }

        private void OnDestroy()
        {
            // イベント解除
            EventBus.Unsubscribe<PriceChangedEvent>(OnPriceChanged);
            EventBus.Unsubscribe<SeasonChangedEvent>(OnSeasonChanged);
            EventBus.Unsubscribe<MarketEventTriggeredEvent>(OnMarketEventTriggered);

            // ボタンイベント解除
            if (buyButton != null)
                buyButton.onClick.RemoveListener(OnBuyPressed);

            if (sellButton != null)
                sellButton.onClick.RemoveListener(OnSellPressed);

            if (inventoryButton != null)
                inventoryButton.onClick.RemoveListener(OnInventoryPressed);

            if (backButton != null)
                backButton.onClick.RemoveListener(OnBackPressed);

            ClearChartElements();
        }
    }
}
