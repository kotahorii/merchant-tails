using System.Collections.Generic;
using MerchantTails.Core;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.UI
{
    /// <summary>
    /// 店舗投資のUIパネル
    /// </summary>
    public class ShopInvestmentPanel : MonoBehaviour
    {
        [Header("Overview")]
        [SerializeField]
        private TextMeshProUGUI totalInvestmentText;

        [SerializeField]
        private TextMeshProUGUI currentMoneyText;

        [Header("Current Bonuses")]
        [SerializeField]
        private TextMeshProUGUI storageBonusText;

        [SerializeField]
        private TextMeshProUGUI efficiencyBonusText;

        [SerializeField]
        private TextMeshProUGUI customerBonusText;

        [SerializeField]
        private TextMeshProUGUI qualityBonusText;

        [Header("Category Tabs")]
        [SerializeField]
        private Toggle[] categoryTabs;

        [SerializeField]
        private TextMeshProUGUI[] categoryProgressTexts;

        [Header("Upgrade List")]
        [SerializeField]
        private Transform upgradeListContainer;

        [SerializeField]
        private GameObject upgradeItemPrefab;

        [Header("UI Settings")]
        [SerializeField]
        private Color maxedColor = new Color(1f, 0.8f, 0.2f);

        [SerializeField]
        private Color unavailableColor = new Color(0.5f, 0.5f, 0.5f);

        [SerializeField]
        private Color purchasableColor = Color.white;

        private ShopInvestmentSystem investmentSystem;
        private PlayerData playerData;
        private List<ShopUpgradeItemUI> upgradeItems = new List<ShopUpgradeItemUI>();
        private UpgradeCategory currentCategory = UpgradeCategory.Storage;

        private void Start()
        {
            investmentSystem = ShopInvestmentSystem.Instance;
            playerData = GameManager.Instance?.PlayerData;

            if (investmentSystem == null)
            {
                ErrorHandler.LogError("ShopInvestmentSystem not found!", null, "ShopInvestmentPanel");
                return;
            }

            InitializeCategoryTabs();
            SubscribeToEvents();
            RefreshDisplay();
        }

        private void OnDestroy()
        {
            UnsubscribeFromEvents();
        }

        private void InitializeCategoryTabs()
        {
            for (int i = 0; i < categoryTabs.Length; i++)
            {
                int index = i;
                categoryTabs[i]
                    .onValueChanged.AddListener(
                        (isOn) =>
                        {
                            if (isOn)
                            {
                                currentCategory = (UpgradeCategory)index;
                                RefreshUpgradeList();
                            }
                        }
                    );
            }
        }

        private void SubscribeToEvents()
        {
            if (investmentSystem != null)
            {
                investmentSystem.OnUpgradePurchased += OnUpgradePurchased;
                investmentSystem.OnUpgradeMaxed += OnUpgradeMaxed;
                investmentSystem.OnBonusesUpdated += OnBonusesUpdated;
            }

            if (playerData != null)
            {
                playerData.OnMoneyChanged += OnMoneyChanged;
            }
        }

        private void UnsubscribeFromEvents()
        {
            if (investmentSystem != null)
            {
                investmentSystem.OnUpgradePurchased -= OnUpgradePurchased;
                investmentSystem.OnUpgradeMaxed -= OnUpgradeMaxed;
                investmentSystem.OnBonusesUpdated -= OnBonusesUpdated;
            }

            if (playerData != null)
            {
                playerData.OnMoneyChanged -= OnMoneyChanged;
            }
        }

        private void RefreshDisplay()
        {
            UpdateOverview();
            UpdateBonuses();
            UpdateCategoryProgress();
            RefreshUpgradeList();
        }

        private void UpdateOverview()
        {
            if (investmentSystem != null && totalInvestmentText != null)
            {
                totalInvestmentText.text = $"総投資額: {investmentSystem.GetTotalInvestment():N0}G";
            }

            if (playerData != null && currentMoneyText != null)
            {
                currentMoneyText.text = $"所持金: {playerData.CurrentMoney:N0}G";
            }
        }

        private void UpdateBonuses()
        {
            if (investmentSystem == null)
                return;

            if (storageBonusText != null)
                storageBonusText.text = $"保管容量: +{(investmentSystem.StorageCapacityMultiplier - 1) * 100:F0}%";

            if (efficiencyBonusText != null)
                efficiencyBonusText.text =
                    $"取引効率: +{(investmentSystem.TransactionEfficiencyMultiplier - 1) * 100:F0}%";

            if (customerBonusText != null)
                customerBonusText.text = $"来客数: +{(investmentSystem.CustomerFlowMultiplier - 1) * 100:F0}%";

            if (qualityBonusText != null)
                qualityBonusText.text = $"品質保持: +{(investmentSystem.ItemQualityMultiplier - 1) * 100:F0}%";
        }

        private void UpdateCategoryProgress()
        {
            for (
                int i = 0;
                i < categoryProgressTexts.Length && i < System.Enum.GetValues(typeof(UpgradeCategory)).Length;
                i++
            )
            {
                var category = (UpgradeCategory)i;
                var upgrades = investmentSystem.GetUpgradesByCategory(category);

                int totalLevels = 0;
                int maxPossibleLevels = 0;

                foreach (var upgrade in upgrades)
                {
                    var progress = investmentSystem.GetProgress(upgrade.id);
                    if (progress != null)
                    {
                        totalLevels += progress.currentLevel;
                        maxPossibleLevels += upgrade.maxLevel;
                    }
                }

                if (categoryProgressTexts[i] != null)
                {
                    if (maxPossibleLevels > 0)
                    {
                        float percentage = (float)totalLevels / maxPossibleLevels * 100;
                        categoryProgressTexts[i].text = $"{percentage:F0}%";
                    }
                    else
                    {
                        categoryProgressTexts[i].text = "0%";
                    }
                }
            }
        }

        private void RefreshUpgradeList()
        {
            // 既存のアイテムをクリア
            foreach (var item in upgradeItems)
            {
                Destroy(item.gameObject);
            }
            upgradeItems.Clear();

            // カテゴリ別のアップグレードを取得
            var upgrades = investmentSystem.GetUpgradesByCategory(currentCategory);

            // アップグレードアイテムを生成
            foreach (var upgrade in upgrades)
            {
                CreateUpgradeItem(upgrade);
            }

            // レイアウトを更新
            LayoutRebuilder.ForceRebuildLayoutImmediate(upgradeListContainer as RectTransform);
        }

        private void CreateUpgradeItem(ShopUpgrade upgrade)
        {
            GameObject itemObj = Instantiate(upgradeItemPrefab, upgradeListContainer);
            ShopUpgradeItemUI itemUI = itemObj.GetComponent<ShopUpgradeItemUI>();

            if (itemUI == null)
            {
                itemUI = itemObj.AddComponent<ShopUpgradeItemUI>();
            }

            var progress = investmentSystem.GetProgress(upgrade.id);
            int nextCost =
                progress.currentLevel < upgrade.maxLevel
                    ? investmentSystem.CalculateUpgradeCost(upgrade, progress.currentLevel)
                    : 0;

            itemUI.Setup(
                upgrade,
                progress,
                playerData?.CurrentMoney ?? 0,
                nextCost,
                purchasableColor,
                unavailableColor,
                maxedColor
            );
            itemUI.OnPurchase += OnPurchaseUpgrade;

            upgradeItems.Add(itemUI);
        }

        private void OnPurchaseUpgrade(string upgradeId)
        {
            if (investmentSystem.PurchaseUpgrade(upgradeId))
            {
                RefreshDisplay();
            }
            else
            {
                // エラー表示
                if (UIManager.Instance != null)
                {
                    UIManager.Instance.ShowNotification(
                        "購入失敗",
                        "アップグレードを購入できませんでした",
                        3f,
                        UIManager.NotificationType.Error
                    );
                }
            }
        }

        // イベントハンドラ
        private void OnUpgradePurchased(ShopUpgrade upgrade, int newLevel)
        {
            RefreshDisplay();

            // 成功通知
            if (UIManager.Instance != null)
            {
                UIManager.Instance.ShowNotification(
                    "アップグレード完了",
                    $"{upgrade.name} レベル{newLevel}",
                    3f,
                    UIManager.NotificationType.Success
                );
            }
        }

        private void OnUpgradeMaxed(ShopUpgrade upgrade)
        {
            // MAX達成通知
            if (UIManager.Instance != null)
            {
                UIManager.Instance.ShowNotification(
                    "最大レベル達成！",
                    $"{upgrade.name}が最大レベルに到達しました",
                    5f,
                    UIManager.NotificationType.Success
                );
            }
        }

        private void OnBonusesUpdated()
        {
            UpdateBonuses();
        }

        private void OnMoneyChanged(int newAmount)
        {
            UpdateOverview();

            // 購入可能状態を更新
            foreach (var item in upgradeItems)
            {
                item.UpdateAffordability(newAmount);
            }
        }
    }

    /// <summary>
    /// 個別のアップグレード表示UI
    /// </summary>
    public class ShopUpgradeItemUI : MonoBehaviour
    {
        [Header("UI References")]
        [SerializeField]
        private Image icon;

        [SerializeField]
        private TextMeshProUGUI nameText;

        [SerializeField]
        private TextMeshProUGUI descriptionText;

        [SerializeField]
        private TextMeshProUGUI levelText;

        [SerializeField]
        private Slider levelProgressBar;

        [SerializeField]
        private TextMeshProUGUI effectText;

        [SerializeField]
        private TextMeshProUGUI costText;

        [SerializeField]
        private Button purchaseButton;

        [SerializeField]
        private GameObject maxedIndicator;

        [SerializeField]
        private GameObject rankRequirementIndicator;

        [SerializeField]
        private TextMeshProUGUI rankRequirementText;

        private ShopUpgrade upgrade;
        private ShopUpgradeProgress progress;
        private int nextCost;
        private Color purchasableColor;
        private Color unavailableColor;
        private Color maxedColor;

        public event System.Action<string> OnPurchase;

        public void Setup(
            ShopUpgrade shopUpgrade,
            ShopUpgradeProgress upgradeProgress,
            int playerMoney,
            int cost,
            Color purchasable,
            Color unavailable,
            Color maxed
        )
        {
            upgrade = shopUpgrade;
            progress = upgradeProgress;
            nextCost = cost;
            purchasableColor = purchasable;
            unavailableColor = unavailable;
            maxedColor = maxed;

            // 基本情報
            if (nameText != null)
                nameText.text = upgrade.name;
            if (descriptionText != null)
                descriptionText.text = upgrade.description;

            // レベル表示
            bool isMaxed = progress.currentLevel >= upgrade.maxLevel;

            if (levelText != null)
            {
                levelText.text = $"Lv.{progress.currentLevel}/{upgrade.maxLevel}";
                levelText.color = isMaxed ? maxedColor : purchasableColor;
            }

            if (levelProgressBar != null)
            {
                levelProgressBar.value = (float)progress.currentLevel / upgrade.maxLevel;
                var fillImage = levelProgressBar.fillRect.GetComponent<Image>();
                if (fillImage != null)
                {
                    fillImage.color = isMaxed ? maxedColor : purchasableColor;
                }
            }

            // 効果表示
            if (effectText != null)
            {
                float currentEffect = upgrade.effectPerLevel * progress.currentLevel * 100;
                float nextEffect = upgrade.effectPerLevel * (progress.currentLevel + 1) * 100;

                if (isMaxed)
                {
                    effectText.text = $"効果: +{currentEffect:F0}%";
                }
                else
                {
                    effectText.text = $"効果: +{currentEffect:F0}% → +{nextEffect:F0}%";
                }
            }

            // コスト表示
            if (costText != null)
            {
                if (isMaxed)
                {
                    costText.text = "MAX";
                    costText.color = maxedColor;
                }
                else
                {
                    costText.text = $"{nextCost:N0}G";
                    costText.color = playerMoney >= nextCost ? purchasableColor : unavailableColor;
                }
            }

            // 購入ボタン
            if (purchaseButton != null)
            {
                purchaseButton.onClick.RemoveAllListeners();
                purchaseButton.onClick.AddListener(OnPurchaseClick);
                purchaseButton.interactable = !isMaxed && playerMoney >= nextCost;
            }

            // MAXインジケータ
            if (maxedIndicator != null)
            {
                maxedIndicator.SetActive(isMaxed);
            }

            // ランク要件
            bool meetsRankRequirement = GameManager.Instance?.PlayerData?.CurrentRank >= upgrade.requiredRank;

            if (rankRequirementIndicator != null)
            {
                rankRequirementIndicator.SetActive(
                    !meetsRankRequirement && upgrade.requiredRank > MerchantRank.Apprentice
                );
            }

            if (rankRequirementText != null && !meetsRankRequirement)
            {
                rankRequirementText.text = $"{upgrade.requiredRank}以上";
            }

            // 全体の色調整
            if (!meetsRankRequirement || (playerMoney < nextCost && !isMaxed))
            {
                SetUIColor(unavailableColor);
            }
        }

        public void UpdateAffordability(int playerMoney)
        {
            bool canAfford = playerMoney >= nextCost;
            bool isMaxed = progress.currentLevel >= upgrade.maxLevel;

            if (purchaseButton != null)
            {
                purchaseButton.interactable = !isMaxed && canAfford;
            }

            if (costText != null && !isMaxed)
            {
                costText.color = canAfford ? purchasableColor : unavailableColor;
            }
        }

        private void OnPurchaseClick()
        {
            OnPurchase?.Invoke(upgrade.id);
        }

        private void SetUIColor(Color color)
        {
            if (nameText != null)
                nameText.color = color;
            if (descriptionText != null)
                descriptionText.color = color;
            if (effectText != null)
                effectText.color = color;
            if (icon != null)
                icon.color = color;
        }
    }
}
