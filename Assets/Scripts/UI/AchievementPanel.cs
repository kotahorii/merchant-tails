using System.Collections.Generic;
using MerchantTails.Core;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.UI
{
    /// <summary>
    /// 実績一覧を表示するUIパネル
    /// </summary>
    public class AchievementPanel : MonoBehaviour
    {
        [Header("UI References")]
        [SerializeField]
        private Transform achievementListContainer;

        [SerializeField]
        private GameObject achievementItemPrefab;

        [SerializeField]
        private TextMeshProUGUI totalPointsText;

        [SerializeField]
        private TextMeshProUGUI completionText;

        [SerializeField]
        private Slider completionSlider;

        [Header("Category Tabs")]
        [SerializeField]
        private Toggle[] categoryTabs;

        [SerializeField]
        private TextMeshProUGUI[] categoryPointsTexts;

        [Header("Display Settings")]
        [SerializeField]
        private Color unlockedColor = Color.white;

        [SerializeField]
        private Color lockedColor = new Color(0.5f, 0.5f, 0.5f, 0.5f);

        [SerializeField]
        private Color progressBarColor = new Color(1f, 0.8f, 0.2f);

        [SerializeField]
        private Color completedBarColor = new Color(0.2f, 1f, 0.3f);

        private AchievementSystem achievementSystem;
        private List<AchievementItemUI> achievementItems = new List<AchievementItemUI>();
        private AchievementCategory currentCategory = AchievementCategory.Trading;

        private void Start()
        {
            achievementSystem = AchievementSystem.Instance;

            if (achievementSystem == null)
            {
                ErrorHandler.LogError("AchievementSystem not found!", null, "AchievementPanel");
                return;
            }

            InitializeCategoryTabs();
            RefreshDisplay();

            // イベント登録
            achievementSystem.OnAchievementUnlocked += OnAchievementUnlocked;
            achievementSystem.OnAchievementProgress += OnAchievementProgress;
        }

        private void OnDestroy()
        {
            if (achievementSystem != null)
            {
                achievementSystem.OnAchievementUnlocked -= OnAchievementUnlocked;
                achievementSystem.OnAchievementProgress -= OnAchievementProgress;
            }
        }

        private void InitializeCategoryTabs()
        {
            for (int i = 0; i < categoryTabs.Length; i++)
            {
                int index = i; // クロージャのためにローカル変数にコピー
                categoryTabs[i]
                    .onValueChanged.AddListener(
                        (isOn) =>
                        {
                            if (isOn)
                            {
                                currentCategory = (AchievementCategory)index;
                                RefreshAchievementList();
                            }
                        }
                    );
            }
        }

        private void RefreshDisplay()
        {
            UpdateOverallStats();
            UpdateCategoryStats();
            RefreshAchievementList();
        }

        private void UpdateOverallStats()
        {
            int totalPoints = achievementSystem.GetTotalUnlockedPoints();
            float completion = achievementSystem.GetCompletionPercentage();

            totalPointsText.text = $"総ポイント: {totalPoints}";
            completionText.text = $"達成率: {completion:F1}%";
            completionSlider.value = completion / 100f;
        }

        private void UpdateCategoryStats()
        {
            for (
                int i = 0;
                i < categoryPointsTexts.Length && i < System.Enum.GetValues(typeof(AchievementCategory)).Length;
                i++
            )
            {
                var category = (AchievementCategory)i;
                var achievements = achievementSystem.GetAchievementsByCategory(category);

                int unlockedCount = 0;
                int totalPoints = 0;
                int unlockedPoints = 0;

                foreach (var achievement in achievements)
                {
                    totalPoints += achievement.points;
                    var progress = achievementSystem.GetProgress(achievement.id);
                    if (progress != null && progress.unlocked)
                    {
                        unlockedCount++;
                        unlockedPoints += achievement.points;
                    }
                }

                categoryPointsTexts[i].text = $"{unlockedCount}/{achievements.Count} ({unlockedPoints}pt)";
            }
        }

        private void RefreshAchievementList()
        {
            // 既存のアイテムをクリア
            foreach (var item in achievementItems)
            {
                Destroy(item.gameObject);
            }
            achievementItems.Clear();

            // カテゴリ別の実績を取得
            var achievements = achievementSystem.GetAchievementsByCategory(currentCategory);

            // 実績アイテムを生成
            foreach (var achievement in achievements)
            {
                var progress = achievementSystem.GetProgress(achievement.id);

                // 隠し実績で未解除の場合はスキップ
                if (achievement.hidden && progress != null && !progress.unlocked)
                {
                    continue;
                }

                CreateAchievementItem(achievement, progress);
            }

            // レイアウトを更新
            LayoutRebuilder.ForceRebuildLayoutImmediate(achievementListContainer as RectTransform);
        }

        private void CreateAchievementItem(Achievement achievement, AchievementProgress progress)
        {
            GameObject itemObj = Instantiate(achievementItemPrefab, achievementListContainer);
            AchievementItemUI itemUI = itemObj.GetComponent<AchievementItemUI>();

            if (itemUI == null)
            {
                itemUI = itemObj.AddComponent<AchievementItemUI>();
            }

            itemUI.Setup(achievement, progress, unlockedColor, lockedColor, progressBarColor, completedBarColor);
            achievementItems.Add(itemUI);
        }

        private void OnAchievementUnlocked(Achievement achievement)
        {
            RefreshDisplay();
        }

        private void OnAchievementProgress(Achievement achievement, float progressPercentage)
        {
            // 該当する実績アイテムの進捗を更新
            foreach (var item in achievementItems)
            {
                if (item.AchievementId == achievement.id)
                {
                    item.UpdateProgress(progressPercentage);
                    break;
                }
            }
        }
    }

    /// <summary>
    /// 個別の実績表示UI
    /// </summary>
    public class AchievementItemUI : MonoBehaviour
    {
        [Header("UI References")]
        [SerializeField]
        private Image icon;

        [SerializeField]
        private TextMeshProUGUI nameText;

        [SerializeField]
        private TextMeshProUGUI descriptionText;

        [SerializeField]
        private TextMeshProUGUI pointsText;

        [SerializeField]
        private GameObject progressBarContainer;

        [SerializeField]
        private Slider progressBar;

        [SerializeField]
        private TextMeshProUGUI progressText;

        [SerializeField]
        private GameObject unlockedIndicator;

        [SerializeField]
        private TextMeshProUGUI unlockedDateText;

        private Achievement achievement;
        private AchievementProgress progress;
        private Color progressBarColor;
        private Color completedBarColor;

        public string AchievementId => achievement?.id;

        public void Setup(
            Achievement achievement,
            AchievementProgress progress,
            Color unlockedColor,
            Color lockedColor,
            Color progressBarColor,
            Color completedBarColor
        )
        {
            this.achievement = achievement;
            this.progress = progress;
            this.progressBarColor = progressBarColor;
            this.completedBarColor = completedBarColor;

            // 基本情報を設定
            nameText.text = achievement.name;
            descriptionText.text = achievement.description;
            pointsText.text = $"{achievement.points}pt";

            // アイコン設定
            if (icon != null && achievement.icon != null)
            {
                icon.sprite = achievement.icon;
            }

            // 解除状態に応じて表示を更新
            bool isUnlocked = progress != null && progress.unlocked;

            // 色を設定
            Color targetColor = isUnlocked ? unlockedColor : lockedColor;
            if (nameText != null)
                nameText.color = targetColor;
            if (descriptionText != null)
                descriptionText.color = targetColor;
            if (icon != null)
                icon.color = targetColor;

            // 解除インジケータ
            if (unlockedIndicator != null)
            {
                unlockedIndicator.SetActive(isUnlocked);
            }

            // 解除日時
            if (unlockedDateText != null)
            {
                if (isUnlocked && progress.unlockedDate.HasValue)
                {
                    unlockedDateText.text = progress.unlockedDate.Value.ToString("yyyy/MM/dd");
                    unlockedDateText.gameObject.SetActive(true);
                }
                else
                {
                    unlockedDateText.gameObject.SetActive(false);
                }
            }

            // 進捗バー
            if (progressBarContainer != null)
            {
                bool showProgress = achievement.showProgressBar && !isUnlocked;
                progressBarContainer.SetActive(showProgress);

                if (showProgress && progressBar != null)
                {
                    float progressPercentage =
                        progress != null ? progress.currentProgress / achievement.maxProgress : 0f;
                    UpdateProgress(progressPercentage);
                }
            }
        }

        public void UpdateProgress(float progressPercentage)
        {
            if (progressBar != null)
            {
                progressBar.value = progressPercentage;

                // プログレスバーの色を設定
                var fillImage = progressBar.fillRect.GetComponent<Image>();
                if (fillImage != null)
                {
                    fillImage.color = progressPercentage >= 1f ? completedBarColor : progressBarColor;
                }
            }

            if (progressText != null && achievement != null)
            {
                if (achievement.maxProgress > 1)
                {
                    float current = progress != null ? progress.currentProgress : 0f;
                    progressText.text = $"{current:F0}/{achievement.maxProgress:F0}";
                }
                else
                {
                    progressText.text = $"{progressPercentage * 100:F0}%";
                }
            }
        }
    }
}
