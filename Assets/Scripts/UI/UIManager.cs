using System;
using System.Collections.Generic;
using MerchantTails.Core;
using MerchantTails.Data;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.UI
{
    /// <summary>
    /// UI管理の中心クラス
    /// 画面遷移、モーダル管理、UIスタックを制御
    /// </summary>
    public class UIManager : MonoBehaviour
    {
        public static UIManager Instance { get; private set; }

        [Header("UI Canvas References")]
        [SerializeField]
        private Canvas mainUICanvas;

        [SerializeField]
        private Canvas modalCanvas;

        [SerializeField]
        private Canvas overlayCanvas;

        [Header("UI Panels")]
        [SerializeField]
        private GameObject mainMenuPanel;

        [SerializeField]
        private GameObject gameHUDPanel;

        [SerializeField]
        private GameObject shopManagementPanel;

        [SerializeField]
        private GameObject marketAnalysisPanel;

        [SerializeField]
        private GameObject inventoryPanel;

        [SerializeField]
        private GameObject settingsPanel;

        [SerializeField]
        private GameObject tutorialPanel;

        [Header("Loading & Transition")]
        [SerializeField]
        private GameObject loadingPanel;

        [SerializeField]
        private Image transitionOverlay;

        [SerializeField]
        private float transitionDuration = 0.5f;

        private Stack<UIPanel> uiStack = new Stack<UIPanel>();
        private Dictionary<UIType, UIPanel> uiPanels = new Dictionary<UIType, UIPanel>();
        private bool isTransitioning = false;

        private void Awake()
        {
            if (Instance == null)
            {
                Instance = this;
                DontDestroyOnLoad(gameObject);
                InitializeUI();
            }
            else
            {
                Destroy(gameObject);
            }
        }

        private void Start()
        {
            // イベント登録
            EventBus.Subscribe<GameStateChangedEvent>(OnGameStateChanged);
        }

        private void OnDestroy()
        {
            EventBus.Unsubscribe<GameStateChangedEvent>(OnGameStateChanged);
        }

        private void InitializeUI()
        {
            // UIパネルの初期化
            RegisterUIPanel(UIType.MainMenu, mainMenuPanel);
            RegisterUIPanel(UIType.GameHUD, gameHUDPanel);
            RegisterUIPanel(UIType.ShopManagement, shopManagementPanel);
            RegisterUIPanel(UIType.MarketAnalysis, marketAnalysisPanel);
            RegisterUIPanel(UIType.Inventory, inventoryPanel);
            RegisterUIPanel(UIType.Settings, settingsPanel);
            RegisterUIPanel(UIType.Tutorial, tutorialPanel);

            // 初期状態では全て非表示
            HideAllPanels();

            // メインメニューを表示
            ShowPanel(UIType.MainMenu);

            ErrorHandler.LogInfo("UIManager initialized", "UIManager");
        }

        private void RegisterUIPanel(UIType uiType, GameObject panelObject)
        {
            if (panelObject != null)
            {
                var uiPanel = panelObject.GetComponent<UIPanel>();
                if (uiPanel == null)
                {
                    uiPanel = panelObject.AddComponent<UIPanel>();
                }

                uiPanel.Initialize(uiType);
                uiPanels[uiType] = uiPanel;
            }
        }

        public void ShowPanel(UIType uiType, bool addToStack = true)
        {
            if (isTransitioning)
                return;

            if (uiPanels.TryGetValue(uiType, out UIPanel panel))
            {
                StartCoroutine(ShowPanelCoroutine(panel, addToStack));
            }
            else
            {
                ErrorHandler.LogError($"UI Panel {uiType} not found", null, "UIManager");
            }
        }

        public void HidePanel(UIType uiType)
        {
            if (uiPanels.TryGetValue(uiType, out UIPanel panel))
            {
                panel.Hide();

                // スタックから削除
                if (uiStack.Count > 0 && uiStack.Peek() == panel)
                {
                    uiStack.Pop();
                }
            }
        }

        public void HideAllPanels()
        {
            foreach (var panel in uiPanels.Values)
            {
                panel.Hide();
            }
            uiStack.Clear();
        }

        public void GoBack()
        {
            if (uiStack.Count > 1)
            {
                var currentPanel = uiStack.Pop();
                currentPanel.Hide();

                var previousPanel = uiStack.Peek();
                previousPanel.Show();
            }
        }

        public void ShowModal(UIType modalType, System.Action<bool> onResult = null)
        {
            // モーダル表示の実装
            if (uiPanels.TryGetValue(modalType, out UIPanel modal))
            {
                modal.SetParent(modalCanvas.transform);
                modal.Show();

                // モーダル結果のコールバック設定
                modal.SetModalCallback(onResult);
            }
        }

        public void HideModal(UIType modalType, bool result = false)
        {
            if (uiPanels.TryGetValue(modalType, out UIPanel modal))
            {
                modal.Hide();
                modal.TriggerModalCallback(result);
            }
        }

        public void ShowLoading(string message = "Loading...")
        {
            if (loadingPanel != null)
            {
                loadingPanel.SetActive(true);
                // ローディングメッセージの設定
                var loadingText = loadingPanel.GetComponentInChildren<Text>();
                if (loadingText != null)
                {
                    loadingText.text = message;
                }
            }
        }

        public void HideLoading()
        {
            if (loadingPanel != null)
            {
                loadingPanel.SetActive(false);
            }
        }

        private System.Collections.IEnumerator ShowPanelCoroutine(UIPanel panel, bool addToStack)
        {
            isTransitioning = true;

            // フェードアウト
            yield return StartCoroutine(FadeOut());

            // 現在のパネルを隠す
            if (uiStack.Count > 0)
            {
                uiStack.Peek().Hide();
            }

            // 新しいパネルを表示
            panel.Show();

            if (addToStack)
            {
                uiStack.Push(panel);
            }

            // フェードイン
            yield return StartCoroutine(FadeIn());

            isTransitioning = false;
        }

        private System.Collections.IEnumerator FadeOut()
        {
            if (transitionOverlay != null)
            {
                transitionOverlay.gameObject.SetActive(true);
                float elapsed = 0f;

                while (elapsed < transitionDuration)
                {
                    elapsed += Time.deltaTime;
                    float alpha = Mathf.Lerp(0f, 1f, elapsed / transitionDuration);
                    transitionOverlay.color = new Color(0, 0, 0, alpha);
                    yield return null;
                }

                transitionOverlay.color = new Color(0, 0, 0, 1f);
            }
        }

        private System.Collections.IEnumerator FadeIn()
        {
            if (transitionOverlay != null)
            {
                float elapsed = 0f;

                while (elapsed < transitionDuration)
                {
                    elapsed += Time.deltaTime;
                    float alpha = Mathf.Lerp(1f, 0f, elapsed / transitionDuration);
                    transitionOverlay.color = new Color(0, 0, 0, alpha);
                    yield return null;
                }

                transitionOverlay.color = new Color(0, 0, 0, 0f);
                transitionOverlay.gameObject.SetActive(false);
            }
        }

        private void OnGameStateChanged(GameStateChangedEvent evt)
        {
            // ゲーム状態に応じてUIを切り替え
            switch (evt.NewState)
            {
                case GameState.MainMenu:
                    ShowPanel(UIType.MainMenu, false);
                    break;

                case GameState.Shopping:
                    ShowPanel(UIType.ShopManagement, false);
                    break;

                case GameState.MarketView:
                    ShowPanel(UIType.MarketAnalysis, false);
                    break;

                case GameState.StoreManagement:
                    ShowPanel(UIType.Inventory, false);
                    break;

                case GameState.Tutorial:
                    ShowPanel(UIType.Tutorial, false);
                    break;

                case GameState.Paused:
                    ShowModal(UIType.Settings);
                    break;
            }
        }

        public bool IsCurrentPanel(UIType uiType)
        {
            return uiStack.Count > 0 && uiStack.Peek().UIType == uiType;
        }

        public UIPanel GetPanel(UIType uiType)
        {
            uiPanels.TryGetValue(uiType, out UIPanel panel);
            return panel;
        }

        public int GetStackDepth()
        {
            return uiStack.Count;
        }

        /// <summary>
        /// 通知を表示
        /// </summary>
        public void ShowNotification(
            string title,
            string message,
            float duration = 3f,
            NotificationType type = NotificationType.Info
        )
        {
            ErrorHandler.SafeExecute(
                () =>
                {
                    // TODO: 通知UIの実装
                    ErrorHandler.LogInfo($"[{type}] {title}: {message}", "Notification");
                },
                "UIManager.ShowNotification"
            );
        }

        /// <summary>
        /// 確認ダイアログを表示
        /// </summary>
        public void ShowConfirmDialog(string title, string message, Action onConfirm, Action onCancel)
        {
            ErrorHandler.SafeExecute(
                () =>
                {
                    // TODO: 確認ダイアログUIの実装
                    ErrorHandler.LogInfo($"Confirm Dialog: {title} - {message}", "UIManager");
                    // 仮実装：常に確認を実行
                    onConfirm?.Invoke();
                },
                "UIManager.ShowConfirmDialog"
            );
        }

        public enum NotificationType
        {
            Info,
            Success,
            Warning,
            Error,
        }

        // デバッグ用
        public void LogUIState()
        {
            ErrorHandler.LogInfo($"UI Stack depth: {uiStack.Count}", "UIManager");
            ErrorHandler.LogInfo($"Is transitioning: {isTransitioning}", "UIManager");

            if (uiStack.Count > 0)
            {
                ErrorHandler.LogInfo($"Current panel: {uiStack.Peek().UIType}", "UIManager");
            }
        }
    }
}
