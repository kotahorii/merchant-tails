using System.Collections.Generic;
using MerchantTails.Core;
using MerchantTails.Data;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.UI
{
    /// <summary>
    /// リスク/リターンの関係を説明するUI
    /// 投資の基本概念を視覚的に表現
    /// </summary>
    public class RiskReturnExplanationPanel : UIPanel
    {
        [Header("UI References")]
        [SerializeField]
        private TextMeshProUGUI titleText;

        [SerializeField]
        private TextMeshProUGUI descriptionText;

        [SerializeField]
        private RectTransform graphContainer;

        [SerializeField]
        private GameObject riskReturnPointPrefab;

        [SerializeField]
        private LineRenderer riskReturnLine;

        [SerializeField]
        private TextMeshProUGUI xAxisLabel;

        [SerializeField]
        private TextMeshProUGUI yAxisLabel;

        [Header("Investment Type Display")]
        [SerializeField]
        private Transform investmentTypeContainer;

        [SerializeField]
        private GameObject investmentTypeItemPrefab;

        [Header("Interactive Elements")]
        [SerializeField]
        private Slider riskSlider;

        [SerializeField]
        private TextMeshProUGUI currentRiskText;

        [SerializeField]
        private TextMeshProUGUI expectedReturnText;

        [SerializeField]
        private TextMeshProUGUI volatilityText;

        [SerializeField]
        private Button closeButton;

        [SerializeField]
        private Button detailsButton;

        [Header("Graph Settings")]
        [SerializeField]
        private float graphWidth = 600f;

        [SerializeField]
        private float graphHeight = 400f;

        [SerializeField]
        private AnimationCurve riskReturnCurve = AnimationCurve.Linear(0, 0, 1, 1);

        [SerializeField]
        private Gradient riskColorGradient;

        private List<InvestmentTypeInfo> investmentTypes;
        private Dictionary<string, RiskReturnPoint> pointsDict = new Dictionary<string, RiskReturnPoint>();
        private bool isShowingDetails = false;

        protected override void Awake()
        {
            base.Awake();
            InitializeInvestmentTypes();
            SetupUI();
        }

        private void Start()
        {
            if (riskSlider != null)
            {
                riskSlider.onValueChanged.AddListener(OnRiskSliderChanged);
            }

            if (closeButton != null)
            {
                closeButton.onClick.AddListener(Hide);
            }

            if (detailsButton != null)
            {
                detailsButton.onClick.AddListener(ToggleDetails);
            }

            // 初期状態で非表示
            gameObject.SetActive(false);
        }

        /// <summary>
        /// 投資タイプの初期化
        /// </summary>
        private void InitializeInvestmentTypes()
        {
            investmentTypes = new List<InvestmentTypeInfo>
            {
                new InvestmentTypeInfo
                {
                    id = "fruit_trading",
                    displayName = "果物取引",
                    itemType = ItemType.Fruit,
                    description = "腐りやすいが回転率が高い",
                    riskLevel = 0.2f,
                    expectedReturn = 0.15f,
                    volatility = 0.3f,
                    color = new Color(1f, 0.6f, 0.3f),
                },
                new InvestmentTypeInfo
                {
                    id = "potion_trading",
                    displayName = "ポーション取引",
                    itemType = ItemType.Potion,
                    description = "イベントに左右されやすい",
                    riskLevel = 0.5f,
                    expectedReturn = 0.35f,
                    volatility = 0.5f,
                    color = new Color(0.6f, 0.3f, 1f),
                },
                new InvestmentTypeInfo
                {
                    id = "weapon_trading",
                    displayName = "武器取引",
                    itemType = ItemType.Weapon,
                    description = "安定した需要と価格",
                    riskLevel = 0.3f,
                    expectedReturn = 0.2f,
                    volatility = 0.2f,
                    color = new Color(0.7f, 0.7f, 0.7f),
                },
                new InvestmentTypeInfo
                {
                    id = "accessory_trading",
                    displayName = "アクセサリー取引",
                    itemType = ItemType.Accessory,
                    description = "流行に敏感で変動が激しい",
                    riskLevel = 0.7f,
                    expectedReturn = 0.5f,
                    volatility = 0.7f,
                    color = new Color(1f, 0.9f, 0.3f),
                },
                new InvestmentTypeInfo
                {
                    id = "magic_book_trading",
                    displayName = "魔法書取引",
                    itemType = ItemType.MagicBook,
                    description = "高額だが安定している",
                    riskLevel = 0.4f,
                    expectedReturn = 0.25f,
                    volatility = 0.15f,
                    color = new Color(0.5f, 0.3f, 0.8f),
                },
                new InvestmentTypeInfo
                {
                    id = "gem_trading",
                    displayName = "宝石取引",
                    itemType = ItemType.Gem,
                    description = "ハイリスク・ハイリターン",
                    riskLevel = 0.9f,
                    expectedReturn = 0.8f,
                    volatility = 0.9f,
                    color = new Color(0.3f, 1f, 1f),
                },
                new InvestmentTypeInfo
                {
                    id = "bank_deposit",
                    displayName = "銀行預金",
                    description = "低リスクで確実な利息",
                    riskLevel = 0.1f,
                    expectedReturn = 0.05f,
                    volatility = 0f,
                    color = new Color(0.2f, 0.8f, 0.2f),
                },
                new InvestmentTypeInfo
                {
                    id = "shop_investment",
                    displayName = "店舗投資",
                    description = "設備改善による効率向上",
                    riskLevel = 0.3f,
                    expectedReturn = 0.15f,
                    volatility = 0.1f,
                    color = new Color(0.8f, 0.6f, 0.2f),
                },
                new InvestmentTypeInfo
                {
                    id = "merchant_investment",
                    displayName = "他商人への出資",
                    description = "配当による不労所得",
                    riskLevel = 0.6f,
                    expectedReturn = 0.4f,
                    volatility = 0.4f,
                    color = new Color(0.9f, 0.4f, 0.7f),
                },
            };
        }

        /// <summary>
        /// UIの初期設定
        /// </summary>
        private void SetupUI()
        {
            titleText.text = "リスクとリターンの関係";
            descriptionText.text =
                "投資にはリスクとリターンのバランスが重要です。\n一般的に、高いリターンを得るには高いリスクを取る必要があります。";

            xAxisLabel.text = "リスク";
            yAxisLabel.text = "期待リターン";

            // グラフの描画
            DrawRiskReturnGraph();

            // 投資タイプの表示
            DisplayInvestmentTypes();

            // スライダーの初期設定
            if (riskSlider != null)
            {
                riskSlider.minValue = 0f;
                riskSlider.maxValue = 1f;
                riskSlider.value = 0.5f;
                OnRiskSliderChanged(0.5f);
            }
        }

        /// <summary>
        /// リスク・リターングラフを描画
        /// </summary>
        private void DrawRiskReturnGraph()
        {
            // 既存のポイントをクリア
            foreach (var point in pointsDict.Values)
            {
                if (point.gameObject != null)
                {
                    Destroy(point.gameObject);
                }
            }
            pointsDict.Clear();

            // 投資タイプごとにポイントを配置
            foreach (var investmentType in investmentTypes)
            {
                GameObject pointObj = Instantiate(riskReturnPointPrefab, graphContainer);
                RectTransform rectTransform = pointObj.GetComponent<RectTransform>();

                // グラフ上の位置を計算
                float x = investmentType.riskLevel * graphWidth;
                float y = investmentType.expectedReturn * graphHeight;
                rectTransform.anchoredPosition = new Vector2(x, y);

                // ポイントの設定
                RiskReturnPoint point = pointObj.GetComponent<RiskReturnPoint>();
                if (point == null)
                {
                    point = pointObj.AddComponent<RiskReturnPoint>();
                }

                point.Setup(investmentType, this);
                pointsDict[investmentType.id] = point;

                // 色の設定
                Image pointImage = pointObj.GetComponent<Image>();
                if (pointImage != null)
                {
                    pointImage.color = investmentType.color;
                }
            }

            // リスク・リターン曲線の描画
            DrawRiskReturnCurve();
        }

        /// <summary>
        /// リスク・リターン曲線を描画
        /// </summary>
        private void DrawRiskReturnCurve()
        {
            if (riskReturnLine == null)
                return;

            int pointCount = 50;
            Vector3[] positions = new Vector3[pointCount];

            for (int i = 0; i < pointCount; i++)
            {
                float t = i / (float)(pointCount - 1);
                float risk = t;
                float expectedReturn = riskReturnCurve.Evaluate(t);

                float x = risk * graphWidth;
                float y = expectedReturn * graphHeight;

                positions[i] = new Vector3(x, y, 0);
            }

            riskReturnLine.positionCount = pointCount;
            riskReturnLine.SetPositions(positions);
            riskReturnLine.startColor = Color.gray;
            riskReturnLine.endColor = Color.gray;
        }

        /// <summary>
        /// 投資タイプ一覧を表示
        /// </summary>
        private void DisplayInvestmentTypes()
        {
            // 既存のアイテムをクリア
            foreach (Transform child in investmentTypeContainer)
            {
                Destroy(child.gameObject);
            }

            foreach (var investmentType in investmentTypes)
            {
                GameObject item = Instantiate(investmentTypeItemPrefab, investmentTypeContainer);
                InvestmentTypeItem itemComponent = item.GetComponent<InvestmentTypeItem>();

                if (itemComponent == null)
                {
                    itemComponent = item.AddComponent<InvestmentTypeItem>();
                }

                itemComponent.Setup(investmentType, this);
            }
        }

        /// <summary>
        /// リスクスライダーの値が変更された時
        /// </summary>
        private void OnRiskSliderChanged(float value)
        {
            // リスクレベルに応じた表示更新
            string riskLevelText = value switch
            {
                < 0.2f => "非常に低い",
                < 0.4f => "低い",
                < 0.6f => "中程度",
                < 0.8f => "高い",
                _ => "非常に高い",
            };

            currentRiskText.text = $"リスクレベル: {riskLevelText} ({value:F2})";

            // 期待リターンの計算
            float expectedReturn = riskReturnCurve.Evaluate(value);
            expectedReturnText.text = $"期待リターン: {expectedReturn:P1}";

            // ボラティリティの表示
            float volatility = value * 0.8f + Random.Range(-0.1f, 0.1f);
            volatilityText.text = $"価格変動性: {volatility:P0}";

            // グラフ上のハイライト更新
            UpdateGraphHighlight(value);
        }

        /// <summary>
        /// グラフ上のハイライトを更新
        /// </summary>
        private void UpdateGraphHighlight(float riskLevel)
        {
            foreach (var kvp in pointsDict)
            {
                var point = kvp.Value;
                var info = point.InvestmentInfo;

                // リスクレベルが近いポイントをハイライト
                float distance = Mathf.Abs(info.riskLevel - riskLevel);
                bool isHighlighted = distance < 0.1f;

                point.SetHighlight(isHighlighted);
            }
        }

        /// <summary>
        /// 詳細表示の切り替え
        /// </summary>
        private void ToggleDetails()
        {
            isShowingDetails = !isShowingDetails;

            if (detailsButton != null)
            {
                TextMeshProUGUI buttonText = detailsButton.GetComponentInChildren<TextMeshProUGUI>();
                if (buttonText != null)
                {
                    buttonText.text = isShowingDetails ? "簡易表示" : "詳細表示";
                }
            }

            // 詳細情報の表示/非表示
            foreach (var point in pointsDict.Values)
            {
                point.SetDetailVisibility(isShowingDetails);
            }
        }

        /// <summary>
        /// 特定の投資タイプにフォーカス
        /// </summary>
        public void FocusOnInvestmentType(string investmentTypeId)
        {
            if (pointsDict.TryGetValue(investmentTypeId, out var point))
            {
                // すべてのポイントを暗くする
                foreach (var p in pointsDict.Values)
                {
                    p.SetDimmed(true);
                }

                // 選択したポイントを強調
                point.SetDimmed(false);
                point.SetHighlight(true);

                // スライダーを該当するリスクレベルに設定
                if (riskSlider != null)
                {
                    riskSlider.value = point.InvestmentInfo.riskLevel;
                }
            }
        }

        /// <summary>
        /// フォーカスをクリア
        /// </summary>
        public void ClearFocus()
        {
            foreach (var point in pointsDict.Values)
            {
                point.SetDimmed(false);
                point.SetHighlight(false);
            }
        }

        public override void Show()
        {
            base.Show();
            // アニメーション演出
            AnimateIn();
        }

        private void AnimateIn()
        {
            // グラフポイントのアニメーション
            float delay = 0f;
            foreach (var point in pointsDict.Values)
            {
                point.AnimateIn(delay);
                delay += 0.05f;
            }
        }
    }

    /// <summary>
    /// 投資タイプ情報
    /// </summary>
    [System.Serializable]
    public class InvestmentTypeInfo
    {
        public string id;
        public string displayName;
        public ItemType? itemType; // 商品タイプ（該当する場合）
        public string description;
        public float riskLevel; // 0-1
        public float expectedReturn; // 0-1
        public float volatility; // 0-1
        public Color color;
    }

    /// <summary>
    /// リスク・リターングラフ上のポイント
    /// </summary>
    public class RiskReturnPoint : MonoBehaviour
    {
        private InvestmentTypeInfo investmentInfo;
        private RiskReturnExplanationPanel parentPanel;
        private Image pointImage;
        private TextMeshProUGUI labelText;
        private GameObject detailPanel;
        private Button button;
        private CanvasGroup canvasGroup;

        public InvestmentTypeInfo InvestmentInfo => investmentInfo;

        public void Setup(InvestmentTypeInfo info, RiskReturnExplanationPanel panel)
        {
            investmentInfo = info;
            parentPanel = panel;

            // コンポーネントの取得
            pointImage = GetComponent<Image>();
            if (pointImage == null)
            {
                pointImage = gameObject.AddComponent<Image>();
            }

            button = GetComponent<Button>();
            if (button == null)
            {
                button = gameObject.AddComponent<Button>();
            }

            canvasGroup = GetComponent<CanvasGroup>();
            if (canvasGroup == null)
            {
                canvasGroup = gameObject.AddComponent<CanvasGroup>();
            }

            // ラベルの設定
            labelText = GetComponentInChildren<TextMeshProUGUI>();
            if (labelText != null)
            {
                labelText.text = info.displayName;
            }

            // ボタンイベント
            button.onClick.AddListener(OnClick);

            // 詳細パネルは初期状態で非表示
            detailPanel = transform.Find("DetailPanel")?.gameObject;
            if (detailPanel != null)
            {
                detailPanel.SetActive(false);
            }
        }

        private void OnClick()
        {
            parentPanel.FocusOnInvestmentType(investmentInfo.id);
        }

        public void SetHighlight(bool highlighted)
        {
            transform.localScale = highlighted ? Vector3.one * 1.2f : Vector3.one;
        }

        public void SetDimmed(bool dimmed)
        {
            canvasGroup.alpha = dimmed ? 0.3f : 1f;
        }

        public void SetDetailVisibility(bool visible)
        {
            if (detailPanel != null)
            {
                detailPanel.SetActive(visible);
            }
        }

        public void AnimateIn(float delay)
        {
            transform.localScale = Vector3.zero;
            LeanTween
                .scale(gameObject, Vector3.one, 0.5f)
                .setDelay(delay)
                .setEaseOutBack();
        }
    }

    /// <summary>
    /// 投資タイプリストのアイテム
    /// </summary>
    public class InvestmentTypeItem : MonoBehaviour
    {
        private InvestmentTypeInfo investmentInfo;
        private RiskReturnExplanationPanel parentPanel;
        private Button button;
        private Image iconImage;
        private TextMeshProUGUI nameText;
        private TextMeshProUGUI descriptionText;
        private TextMeshProUGUI riskText;
        private TextMeshProUGUI returnText;

        public void Setup(InvestmentTypeInfo info, RiskReturnExplanationPanel panel)
        {
            investmentInfo = info;
            parentPanel = panel;

            // UIコンポーネントの取得
            button = GetComponent<Button>();
            iconImage = transform.Find("Icon")?.GetComponent<Image>();
            nameText = transform.Find("Name")?.GetComponent<TextMeshProUGUI>();
            descriptionText = transform.Find("Description")?.GetComponent<TextMeshProUGUI>();
            riskText = transform.Find("Risk")?.GetComponent<TextMeshProUGUI>();
            returnText = transform.Find("Return")?.GetComponent<TextMeshProUGUI>();

            // 値の設定
            if (iconImage != null)
            {
                iconImage.color = info.color;
            }

            if (nameText != null)
            {
                nameText.text = info.displayName;
            }

            if (descriptionText != null)
            {
                descriptionText.text = info.description;
            }

            if (riskText != null)
            {
                riskText.text = $"リスク: {GetRiskLevelText(info.riskLevel)}";
            }

            if (returnText != null)
            {
                returnText.text = $"期待リターン: {info.expectedReturn:P0}";
            }

            // ボタンイベント
            if (button != null)
            {
                button.onClick.AddListener(() => parentPanel.FocusOnInvestmentType(info.id));
            }
        }

        private string GetRiskLevelText(float riskLevel)
        {
            return riskLevel switch
            {
                < 0.2f => "★☆☆☆☆",
                < 0.4f => "★★☆☆☆",
                < 0.6f => "★★★☆☆",
                < 0.8f => "★★★★☆",
                _ => "★★★★★",
            };
        }
    }
}
