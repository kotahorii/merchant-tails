using System;
using System.Collections;
using System.Collections.Generic;
using MerchantTails.Core;
using MerchantTails.Data;
using UnityEngine;

namespace MerchantTails.Tutorial
{
    /// <summary>
    /// チュートリアルシステムの管理クラス
    /// 8ステップの段階的チュートリアルを提供
    /// </summary>
    public class TutorialSystem : MonoBehaviour
    {
        private static TutorialSystem instance;
        public static TutorialSystem Instance => instance;

        [SerializeField]
        private TutorialStep[] tutorialSteps;

        [SerializeField]
        private float defaultStepDelay = 0.5f;

        private int currentStepIndex = -1;
        private bool isActive = false;
        private bool isWaitingForAction = false;
        private bool canSkip = true;

        public bool IsActive => isActive;
        public bool IsCompleted => currentStepIndex >= tutorialSteps.Length - 1;
        public int CurrentStep => currentStepIndex;
        public float Progress => (float)(currentStepIndex + 1) / tutorialSteps.Length;

        public event Action<int> OnStepStarted;
        public event Action<int> OnStepCompleted;
        public event Action OnTutorialCompleted;
        public event Action OnTutorialSkipped;

        private void Awake()
        {
            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;

            InitializeTutorialSteps();
            LoadTutorialProgress();
        }

        private void OnDestroy()
        {
            if (instance == this)
            {
                instance = null;
            }
        }

        private void InitializeTutorialSteps()
        {
            // デフォルトの8ステップチュートリアル
            if (tutorialSteps == null || tutorialSteps.Length == 0)
            {
                tutorialSteps = new TutorialStep[]
                {
                    new TutorialStep
                    {
                        stepName = "Welcome",
                        title = "魔法都市エルムへようこそ！",
                        description = "あなたは新米商人として、この街で道具屋を営むことになりました。",
                        targetUI = UIType.MainMenu,
                        requiredAction = TutorialAction.None,
                        highlightArea = new Rect(Screen.width * 0.5f - 200, Screen.height * 0.5f - 100, 400, 200),
                    },
                    new TutorialStep
                    {
                        stepName = "ShopOverview",
                        title = "お店の管理画面",
                        description = "ここがあなたのお店です。商品を仕入れて、お客様に販売しましょう。",
                        targetUI = UIType.ShopManagement,
                        requiredAction = TutorialAction.OpenShop,
                        highlightArea = new Rect(50, 50, 300, 400),
                    },
                    new TutorialStep
                    {
                        stepName = "MarketBasics",
                        title = "相場を確認しよう",
                        description = "商品の価格は日々変動します。安く仕入れて高く売るのが商売の基本です。",
                        targetUI = UIType.MarketAnalysis,
                        requiredAction = TutorialAction.OpenMarket,
                        highlightArea = new Rect(100, 100, 600, 400),
                    },
                    new TutorialStep
                    {
                        stepName = "FirstPurchase",
                        title = "商品を仕入れよう",
                        description = "まずは「くだもの」を10個仕入れてみましょう。",
                        targetUI = UIType.Inventory,
                        requiredAction = TutorialAction.BuyItem,
                        requiredItemType = ItemType.Fruit,
                        requiredQuantity = 10,
                        highlightArea = new Rect(200, 200, 400, 300),
                    },
                    new TutorialStep
                    {
                        stepName = "StockManagement",
                        title = "在庫を店頭に並べよう",
                        description = "仕入れた商品を店頭に移動させて、お客様が買えるようにしましょう。",
                        targetUI = UIType.Inventory,
                        requiredAction = TutorialAction.MoveToStorefront,
                        highlightArea = new Rect(300, 300, 200, 100),
                    },
                    new TutorialStep
                    {
                        stepName = "TimeProgression",
                        title = "時間を進めよう",
                        description = "時間を進めると、お客様が来店して商品を購入していきます。",
                        targetUI = UIType.ShopManagement,
                        requiredAction = TutorialAction.AdvanceTime,
                        highlightArea = new Rect(Screen.width - 250, 50, 200, 100),
                    },
                    new TutorialStep
                    {
                        stepName = "ProfitCheck",
                        title = "利益を確認しよう",
                        description = "商人手帳で今日の売上と利益を確認できます。",
                        targetUI = UIType.ShopManagement,
                        requiredAction = TutorialAction.CheckProfit,
                        highlightArea = new Rect(50, Screen.height - 150, 300, 100),
                    },
                    new TutorialStep
                    {
                        stepName = "TutorialComplete",
                        title = "チュートリアル完了！",
                        description = "基本的な商売の流れを理解しました。さあ、一人前の商人を目指しましょう！",
                        targetUI = UIType.ShopManagement,
                        requiredAction = TutorialAction.None,
                        isLastStep = true,
                    },
                };
            }
        }

        /// <summary>
        /// チュートリアルを開始
        /// </summary>
        public void StartTutorial(bool fromBeginning = true)
        {
            if (isActive)
                return;

            ErrorHandler.LogInfo("Starting tutorial", "TutorialSystem");

            isActive = true;
            if (fromBeginning)
            {
                currentStepIndex = -1;
            }

            // Subscribe to game events
            SubscribeToEvents();

            // Start first step
            NextStep();
        }

        /// <summary>
        /// チュートリアルをスキップ
        /// </summary>
        public void SkipTutorial()
        {
            if (!isActive || !canSkip)
                return;

            ErrorHandler.LogInfo("Skipping tutorial", "TutorialSystem");

            isActive = false;
            currentStepIndex = tutorialSteps.Length - 1;

            UnsubscribeFromEvents();
            SaveTutorialProgress();

            OnTutorialSkipped?.Invoke();
            OnTutorialCompleted?.Invoke();

            // Unlock all basic features
            UnlockBasicFeatures();
        }

        /// <summary>
        /// 次のステップに進む
        /// </summary>
        public void NextStep()
        {
            if (!isActive || isWaitingForAction)
                return;

            currentStepIndex++;

            if (currentStepIndex >= tutorialSteps.Length)
            {
                CompleteTutorial();
                return;
            }

            StartCoroutine(ShowStepCoroutine());
        }

        private IEnumerator ShowStepCoroutine()
        {
            var step = tutorialSteps[currentStepIndex];

            ErrorHandler.LogInfo($"Starting tutorial step: {step.stepName}", "TutorialSystem");

            // Notify step started
            OnStepStarted?.Invoke(currentStepIndex);

            // Wait for delay
            yield return new WaitForSeconds(defaultStepDelay);

            // Show tutorial UI
            ShowTutorialUI(step);

            // If this step requires an action, wait for it
            if (step.requiredAction != TutorialAction.None)
            {
                isWaitingForAction = true;
            }
            else
            {
                // Auto-advance after reading time
                yield return new WaitForSeconds(step.displayDuration);
                CompleteCurrentStep();
            }
        }

        private void ShowTutorialUI(TutorialStep step)
        {
            // Show the tutorial panel as a notification
            var uiManager = ServiceLocator.GetService<IUIManager>();
            if (uiManager != null && step != null)
            {
                string message = step.description;
                if (!string.IsNullOrEmpty(step.instruction))
                {
                    message += "\n\n" + step.instruction;
                }
                uiManager.ShowNotification(step.stepName, message, 0f, NotificationType.Info);
            }

            // Navigate to target UI if needed
            // Note: UI navigation will be handled by the UI system when it's implemented
            if (step.targetUI != UIType.None)
            {
                // TODO: Implement UI navigation when UIManager.ShowPanel is available
                ErrorHandler.LogInfo($"Tutorial requested navigation to {step.targetUI}", "TutorialSystem");
            }
        }

        /// <summary>
        /// 現在のステップを完了
        /// </summary>
        public void CompleteCurrentStep()
        {
            if (!isActive || currentStepIndex < 0)
                return;

            var step = tutorialSteps[currentStepIndex];

            ErrorHandler.LogInfo($"Completing tutorial step: {step.stepName}", "TutorialSystem");

            isWaitingForAction = false;

            // Hide tutorial UI
            // Note: Tutorial notifications will auto-hide based on duration
            ErrorHandler.LogInfo("Tutorial step completed, notification will auto-hide", "TutorialSystem");

            // Notify step completed
            OnStepCompleted?.Invoke(currentStepIndex);

            // Publish event
            EventBus.Publish(new TutorialStepCompletedEvent(currentStepIndex, step.stepName, step.isLastStep));

            SaveTutorialProgress();

            if (step.isLastStep)
            {
                CompleteTutorial();
            }
            else
            {
                // Auto-advance to next step
                NextStep();
            }
        }

        private void CompleteTutorial()
        {
            ErrorHandler.LogInfo("Tutorial completed!", "TutorialSystem");

            isActive = false;
            UnsubscribeFromEvents();

            OnTutorialCompleted?.Invoke();

            // Unlock all basic features
            UnlockBasicFeatures();

            SaveTutorialProgress();
        }

        private void UnlockBasicFeatures()
        {
            // Enable all basic game features
            if (GameManager.Instance != null)
            {
                GameManager.Instance.SetTutorialCompleted(true);
            }
        }

        #region Event Handling

        private void SubscribeToEvents()
        {
            EventBus.Subscribe<TransactionCompletedEvent>(OnTransactionCompleted);
            EventBus.Subscribe<GameStateChangedEvent>(OnGameStateChanged);
            EventBus.Subscribe<PhaseChangedEvent>(OnPhaseChanged);
        }

        private void UnsubscribeFromEvents()
        {
            EventBus.Unsubscribe<TransactionCompletedEvent>(OnTransactionCompleted);
            EventBus.Unsubscribe<GameStateChangedEvent>(OnGameStateChanged);
            EventBus.Unsubscribe<PhaseChangedEvent>(OnPhaseChanged);
        }

        private void OnTransactionCompleted(TransactionCompletedEvent e)
        {
            if (!isActive || !isWaitingForAction)
                return;

            var step = tutorialSteps[currentStepIndex];

            // Check if this transaction completes the current step
            if (step.requiredAction == TutorialAction.BuyItem && e.IsPurchase)
            {
                if (step.requiredItemType == ItemType.None || step.requiredItemType == e.ItemType)
                {
                    if (step.requiredQuantity <= 0 || e.Quantity >= step.requiredQuantity)
                    {
                        CompleteCurrentStep();
                    }
                }
            }
        }

        private void OnGameStateChanged(GameStateChangedEvent e)
        {
            if (!isActive || !isWaitingForAction)
                return;

            var step = tutorialSteps[currentStepIndex];

            // Check state-based actions
            switch (step.requiredAction)
            {
                case TutorialAction.OpenShop:
                    if (e.NewState == GameState.StoreManagement)
                        CompleteCurrentStep();
                    break;

                case TutorialAction.OpenMarket:
                    if (e.NewState == GameState.MarketView)
                        CompleteCurrentStep();
                    break;
            }
        }

        private void OnPhaseChanged(PhaseChangedEvent e)
        {
            if (!isActive || !isWaitingForAction)
                return;

            var step = tutorialSteps[currentStepIndex];

            if (step.requiredAction == TutorialAction.AdvanceTime)
            {
                CompleteCurrentStep();
            }
        }

        #endregion

        #region Save/Load

        private void SaveTutorialProgress()
        {
            PlayerPrefs.SetInt("Tutorial_CurrentStep", currentStepIndex);
            PlayerPrefs.SetInt("Tutorial_Completed", IsCompleted ? 1 : 0);
            PlayerPrefs.Save();
        }

        private void LoadTutorialProgress()
        {
            currentStepIndex = PlayerPrefs.GetInt("Tutorial_CurrentStep", -1);
            bool completed = PlayerPrefs.GetInt("Tutorial_Completed", 0) == 1;

            if (completed)
            {
                currentStepIndex = tutorialSteps.Length - 1;
            }
        }

        #endregion

        /// <summary>
        /// 特定のアクションがチュートリアルで要求されているか確認
        /// </summary>
        public bool IsActionRequired(TutorialAction action)
        {
            if (!isActive || !isWaitingForAction || currentStepIndex < 0)
                return false;

            return tutorialSteps[currentStepIndex].requiredAction == action;
        }

        /// <summary>
        /// 現在のステップ情報を取得
        /// </summary>
        public TutorialStep GetCurrentStep()
        {
            if (currentStepIndex < 0 || currentStepIndex >= tutorialSteps.Length)
                return null;

            return tutorialSteps[currentStepIndex];
        }
    }

    /// <summary>
    /// チュートリアルステップの定義
    /// </summary>
    [Serializable]
    public class TutorialStep
    {
        public string stepName;
        public string title;
        public string description;
        public UIType targetUI = UIType.None;
        public TutorialAction requiredAction = TutorialAction.None;
        public ItemType requiredItemType = ItemType.None;
        public int requiredQuantity = 0;
        public Rect highlightArea;
        public float displayDuration = 5f;
        public bool canSkip = true;
        public bool isLastStep = false;
    }

    /// <summary>
    /// チュートリアルで要求されるアクション
    /// </summary>
    public enum TutorialAction
    {
        None,
        OpenShop,
        OpenMarket,
        BuyItem,
        SellItem,
        MoveToStorefront,
        MoveToTrading,
        AdvanceTime,
        CheckProfit,
        OpenSettings,
        SaveGame,
    }
}
