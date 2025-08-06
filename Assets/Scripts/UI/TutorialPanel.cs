using System.Collections;
using MerchantTails.Core;
using MerchantTails.Tutorial;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.UI
{
    /// <summary>
    /// チュートリアル表示用UIパネル
    /// ステップごとの説明表示とハイライト機能を提供
    /// </summary>
    public class TutorialPanel : UIPanel
    {
        private static TutorialPanel instance;
        public static TutorialPanel Instance => instance;

        [Header("Tutorial UI Elements")]
        [SerializeField]
        private TextMeshProUGUI titleText;

        [SerializeField]
        private TextMeshProUGUI descriptionText;

        [SerializeField]
        private TextMeshProUGUI stepCounterText;

        [SerializeField]
        private Image progressBar;

        [SerializeField]
        private Button nextButton;

        [SerializeField]
        private Button skipButton;

        [SerializeField]
        private GameObject highlightOverlay;

        [SerializeField]
        private RectTransform highlightFrame;

        [SerializeField]
        private GameObject arrowPointer;

        [Header("Animation Settings")]
        [SerializeField]
        private float typewriterSpeed = 0.05f;

        [SerializeField]
        private float highlightPulseDuration = 1f;

        [SerializeField]
        private AnimationCurve highlightPulseCurve = AnimationCurve.EaseInOut(0, 0.9f, 1, 1.1f);

        private TutorialStep currentStep;
        private Coroutine typewriterCoroutine;
        private Coroutine highlightCoroutine;

        protected override void Awake()
        {
            base.Awake();

            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;

            SetupButtons();
        }

        protected virtual void OnDestroy()
        {

            if (instance == this)
            {
                instance = null;
            }
        }

        private void SetupButtons()
        {
            if (nextButton != null)
            {
                nextButton.onClick.AddListener(OnNextButtonClicked);
            }

            if (skipButton != null)
            {
                skipButton.onClick.AddListener(OnSkipButtonClicked);
            }
        }

        /// <summary>
        /// チュートリアルステップを表示
        /// </summary>
        public void ShowStep(TutorialStep step)
        {
            ErrorHandler.SafeExecute(
                () =>
                {
                    currentStep = step;
                    Show();
                    UpdateStepDisplay();
                },
                "TutorialPanel.ShowStep"
            );
        }

        private void UpdateStepDisplay()
        {
            if (currentStep == null)
                return;

            // Update title
            if (titleText != null)
            {
                titleText.text = currentStep.title;
            }

            // Start typewriter effect for description
            if (descriptionText != null)
            {
                if (typewriterCoroutine != null)
                {
                    StopCoroutine(typewriterCoroutine);
                }
                typewriterCoroutine = StartCoroutine(TypewriterEffect(currentStep.description));
            }

            // Update step counter
            if (stepCounterText != null && TutorialSystem.Instance != null)
            {
                int currentStepNum = TutorialSystem.Instance.CurrentStep + 1;
                int totalSteps = 8; // Default tutorial has 8 steps
                stepCounterText.text = $"ステップ {currentStepNum}/{totalSteps}";
            }

            // Update progress bar
            if (progressBar != null && TutorialSystem.Instance != null)
            {
                progressBar.fillAmount = TutorialSystem.Instance.Progress;
            }

            // Update buttons
            UpdateButtons();

            // Show highlight if needed
            if (currentStep.highlightArea.width > 0 && currentStep.highlightArea.height > 0)
            {
                ShowHighlight(currentStep.highlightArea);
            }
            else
            {
                HideHighlight();
            }
        }

        private IEnumerator TypewriterEffect(string text)
        {
            descriptionText.text = "";

            foreach (char c in text)
            {
                descriptionText.text += c;
                yield return new WaitForSeconds(typewriterSpeed);
            }

            typewriterCoroutine = null;
        }

        private void UpdateButtons()
        {
            // Next button
            if (nextButton != null)
            {
                bool showNext = currentStep.requiredAction == TutorialAction.None;
                nextButton.gameObject.SetActive(showNext);

                if (showNext)
                {
                    nextButton.GetComponentInChildren<TextMeshProUGUI>().text = currentStep.isLastStep
                        ? "完了"
                        : "次へ";
                }
            }

            // Skip button
            if (skipButton != null)
            {
                skipButton.gameObject.SetActive(currentStep.canSkip && !currentStep.isLastStep);
            }
        }

        private void ShowHighlight(Rect area)
        {
            if (highlightOverlay == null || highlightFrame == null)
                return;

            highlightOverlay.SetActive(true);

            // Position highlight frame
            highlightFrame.anchorMin = Vector2.zero;
            highlightFrame.anchorMax = Vector2.zero;
            highlightFrame.anchoredPosition = new Vector2(area.x + area.width * 0.5f, area.y + area.height * 0.5f);
            highlightFrame.sizeDelta = new Vector2(area.width, area.height);

            // Start pulse animation
            if (highlightCoroutine != null)
            {
                StopCoroutine(highlightCoroutine);
            }
            highlightCoroutine = StartCoroutine(HighlightPulseAnimation());

            // Position arrow pointer if needed
            if (arrowPointer != null)
            {
                arrowPointer.SetActive(true);
                PositionArrowPointer(area);
            }
        }

        private void HideHighlight()
        {
            if (highlightOverlay != null)
            {
                highlightOverlay.SetActive(false);
            }

            if (arrowPointer != null)
            {
                arrowPointer.SetActive(false);
            }

            if (highlightCoroutine != null)
            {
                StopCoroutine(highlightCoroutine);
                highlightCoroutine = null;
            }
        }

        private IEnumerator HighlightPulseAnimation()
        {
            float time = 0;

            while (true)
            {
                time += Time.deltaTime;
                float normalizedTime = (time % highlightPulseDuration) / highlightPulseDuration;
                float scale = highlightPulseCurve.Evaluate(normalizedTime);

                if (highlightFrame != null)
                {
                    highlightFrame.localScale = Vector3.one * scale;
                }

                yield return null;
            }
        }

        private void PositionArrowPointer(Rect targetArea)
        {
            // Position arrow to point at the highlight area
            RectTransform arrowRect = arrowPointer.GetComponent<RectTransform>();
            if (arrowRect != null)
            {
                // Simple positioning - place arrow above the highlight
                arrowRect.anchoredPosition = new Vector2(
                    targetArea.x + targetArea.width * 0.5f,
                    targetArea.y + targetArea.height + 50f
                );

                // Rotate to point down
                arrowRect.rotation = Quaternion.Euler(0, 0, -90);
            }
        }

        private void OnNextButtonClicked()
        {
            ErrorHandler.SafeExecute(
                () =>
                {
                    if (TutorialSystem.Instance != null)
                    {
                        if (currentStep.isLastStep)
                        {
                            TutorialSystem.Instance.CompleteCurrentStep();
                        }
                        else
                        {
                            TutorialSystem.Instance.NextStep();
                        }
                    }
                },
                "TutorialPanel.OnNextButtonClicked"
            );
        }

        private void OnSkipButtonClicked()
        {
            ErrorHandler.SafeExecute(
                () =>
                {
                    if (TutorialSystem.Instance != null)
                    {
                        // Show confirmation dialog
                        ShowSkipConfirmation();
                    }
                },
                "TutorialPanel.OnSkipButtonClicked"
            );
        }

        private void ShowSkipConfirmation()
        {
            if (UIManager.Instance != null)
            {
                UIManager.Instance.ShowConfirmDialog(
                    "チュートリアルをスキップ",
                    "チュートリアルをスキップしますか？\n基本的な操作方法を後で確認することもできます。",
                    () =>
                    {
                        TutorialSystem.Instance.SkipTutorial();
                        Hide();
                    },
                    null
                );
            }
        }

        protected override void OnShow()
        {
            base.OnShow();

            // Reset animations
            if (highlightFrame != null)
            {
                highlightFrame.localScale = Vector3.one;
            }
        }

        protected override void OnHide()
        {
            base.OnHide();

            // Stop coroutines
            if (typewriterCoroutine != null)
            {
                StopCoroutine(typewriterCoroutine);
                typewriterCoroutine = null;
            }

            if (highlightCoroutine != null)
            {
                StopCoroutine(highlightCoroutine);
                highlightCoroutine = null;
            }

            HideHighlight();
        }

        /// <summary>
        /// 特定のUI要素をハイライト（外部から呼び出し可能）
        /// </summary>
        public void HighlightUIElement(RectTransform element)
        {
            if (element == null)
                return;

            // Get world corners
            Vector3[] corners = new Vector3[4];
            element.GetWorldCorners(corners);

            // Convert to screen space
            Vector2 min = RectTransformUtility.WorldToScreenPoint(null, corners[0]);
            Vector2 max = RectTransformUtility.WorldToScreenPoint(null, corners[2]);

            // Create rect
            Rect highlightRect = new Rect(min.x, min.y, max.x - min.x, max.y - min.y);

            ShowHighlight(highlightRect);
        }
    }
}
