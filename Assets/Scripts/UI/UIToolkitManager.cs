using System;
using System.Collections.Generic;
using MerchantTails.Core;
using MerchantTails.Data;
using UnityEngine;
using UnityEngine.UIElements;

namespace MerchantTails.UI
{
    /// <summary>
    /// Unity 6のUI Toolkitを使用した新しいUI管理システム
    /// USS/UXMLベースのモダンなUI実装
    /// </summary>
    public class UIToolkitManager : MonoBehaviour
    {
        public static UIToolkitManager Instance { get; private set; }

        [Header("UI Document References")]
        [SerializeField]
        private UIDocument mainUIDocument;

        [SerializeField]
        private UIDocument modalDocument;

        [SerializeField]
        private UIDocument overlayDocument;

        [Header("Visual Tree Assets")]
        [SerializeField]
        private VisualTreeAsset mainMenuTemplate;

        [SerializeField]
        private VisualTreeAsset gameHUDTemplate;

        [SerializeField]
        private VisualTreeAsset shopManagementTemplate;

        [SerializeField]
        private VisualTreeAsset marketAnalysisTemplate;

        [SerializeField]
        private VisualTreeAsset inventoryTemplate;

        [SerializeField]
        private VisualTreeAsset settingsTemplate;

        [SerializeField]
        private VisualTreeAsset tutorialTemplate;

        [SerializeField]
        private VisualTreeAsset loadingTemplate;

        [Header("Style Sheets")]
        [SerializeField]
        private StyleSheet mainStyleSheet;

        [SerializeField]
        private StyleSheet themeStyleSheet;

        [Header("Transition Settings")]
        [SerializeField]
        private float transitionDuration = 500f; // ミリ秒

        private Stack<UIToolkitPanel> uiStack = new Stack<UIToolkitPanel>();
        private Dictionary<UIType, UIToolkitPanel> uiPanels = new Dictionary<UIType, UIToolkitPanel>();
        private Dictionary<UIType, VisualTreeAsset> uiTemplates = new Dictionary<UIType, VisualTreeAsset>();

        private VisualElement rootElement;
        private VisualElement modalRoot;
        private VisualElement overlayRoot;
        private VisualElement transitionOverlay;

        private bool isTransitioning = false;

        private void Awake()
        {
            if (Instance == null)
            {
                Instance = this;
                DontDestroyOnLoad(gameObject);
                InitializeUIToolkit();
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

            // Unity 6の新しいUI Toolkit機能を活用
            ConfigureRuntimeTheme();
        }

        private void OnDestroy()
        {
            EventBus.Unsubscribe<GameStateChangedEvent>(OnGameStateChanged);

            // クリーンアップ
            foreach (var panel in uiPanels.Values)
            {
                panel.Cleanup();
            }
        }

        private void InitializeUIToolkit()
        {
            // ルート要素の取得
            rootElement = mainUIDocument.rootVisualElement;
            modalRoot = modalDocument?.rootVisualElement;
            overlayRoot = overlayDocument?.rootVisualElement;

            // スタイルシートの適用
            if (mainStyleSheet != null)
            {
                rootElement.styleSheets.Add(mainStyleSheet);
            }
            if (themeStyleSheet != null)
            {
                rootElement.styleSheets.Add(themeStyleSheet);
            }

            // UIテンプレートの登録
            RegisterUITemplates();

            // トランジションオーバーレイの作成
            CreateTransitionOverlay();

            // 初期パネルの作成
            CreateAllPanels();

            // メインメニューを表示
            ShowPanel(UIType.MainMenu);

            ErrorHandler.LogInfo("UIToolkitManager initialized with Unity 6 features", "UIToolkitManager");
        }

        private void RegisterUITemplates()
        {
            uiTemplates[UIType.MainMenu] = mainMenuTemplate;
            uiTemplates[UIType.GameHUD] = gameHUDTemplate;
            uiTemplates[UIType.ShopManagement] = shopManagementTemplate;
            uiTemplates[UIType.MarketAnalysis] = marketAnalysisTemplate;
            uiTemplates[UIType.Inventory] = inventoryTemplate;
            uiTemplates[UIType.Settings] = settingsTemplate;
            uiTemplates[UIType.Tutorial] = tutorialTemplate;
        }

        private void CreateTransitionOverlay()
        {
            transitionOverlay = new VisualElement();
            transitionOverlay.name = "transition-overlay";
            transitionOverlay.style.position = Position.Absolute;
            transitionOverlay.style.width = Length.Percent(100);
            transitionOverlay.style.height = Length.Percent(100);
            transitionOverlay.style.backgroundColor = new Color(0, 0, 0, 0);
            transitionOverlay.style.display = DisplayStyle.None;
            transitionOverlay.pickingMode = PickingMode.Ignore;

            rootElement.Add(transitionOverlay);
        }

        private void CreateAllPanels()
        {
            foreach (var kvp in uiTemplates)
            {
                CreatePanel(kvp.Key, kvp.Value);
            }
        }

        private void CreatePanel(UIType uiType, VisualTreeAsset template)
        {
            if (template == null)
            {
                ErrorHandler.LogWarning($"Template for {uiType} is null", "UIToolkitManager");
                return;
            }

            var panelElement = template.CloneTree();
            panelElement.name = $"{uiType}Panel";
            panelElement.style.position = Position.Absolute;
            panelElement.style.width = Length.Percent(100);
            panelElement.style.height = Length.Percent(100);
            panelElement.style.display = DisplayStyle.None;

            rootElement.Add(panelElement);

            var panel = new UIToolkitPanel(uiType, panelElement);
            panel.Initialize();
            uiPanels[uiType] = panel;
        }

        public void ShowPanel(UIType uiType, bool addToStack = true)
        {
            if (isTransitioning)
                return;

            if (uiPanels.TryGetValue(uiType, out UIToolkitPanel panel))
            {
                ShowPanelAsync(panel, addToStack);
            }
            else
            {
                ErrorHandler.LogError($"UI Panel {uiType} not found", null, "UIToolkitManager");
            }
        }

        private async void ShowPanelAsync(UIToolkitPanel panel, bool addToStack)
        {
            isTransitioning = true;

            // トランジション開始
            await TransitionOut();

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

            // トランジション終了
            await TransitionIn();

            isTransitioning = false;
        }

        private async System.Threading.Tasks.Task TransitionOut()
        {
            transitionOverlay.style.display = DisplayStyle.Flex;

            // Unity 6のUI Toolkitアニメーション
            transitionOverlay.style.opacity = 0;
            transitionOverlay.style.transitionDuration = new List<TimeValue>
            {
                new TimeValue(transitionDuration, TimeUnit.Millisecond),
            };
            transitionOverlay.style.transitionProperty = new List<StylePropertyName>
            {
                new StylePropertyName("opacity"),
            };

            // 次のフレームで開始
            await System.Threading.Tasks.Task.Yield();
            transitionOverlay.style.opacity = 1;

            // アニメーション完了を待つ
            await System.Threading.Tasks.Task.Delay((int)transitionDuration);
        }

        private async System.Threading.Tasks.Task TransitionIn()
        {
            transitionOverlay.style.opacity = 1;

            // 次のフレームで開始
            await System.Threading.Tasks.Task.Yield();
            transitionOverlay.style.opacity = 0;

            // アニメーション完了を待つ
            await System.Threading.Tasks.Task.Delay((int)transitionDuration);

            transitionOverlay.style.display = DisplayStyle.None;
        }

        public void HidePanel(UIType uiType)
        {
            if (uiPanels.TryGetValue(uiType, out UIToolkitPanel panel))
            {
                panel.Hide();

                // スタックから削除
                if (uiStack.Count > 0 && uiStack.Peek() == panel)
                {
                    uiStack.Pop();
                }
            }
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

        public void ShowModal(UIType modalType, Action<bool> onResult = null)
        {
            if (uiPanels.TryGetValue(modalType, out UIToolkitPanel modal))
            {
                // モーダル用のルートに移動
                if (modalRoot != null)
                {
                    modal.Element.RemoveFromHierarchy();
                    modalRoot.Add(modal.Element);
                }

                modal.Show();
                modal.SetModalCallback(onResult);

                // 背景をブロック
                CreateModalBackdrop();
            }
        }

        private void CreateModalBackdrop()
        {
            var backdrop = new VisualElement();
            backdrop.name = "modal-backdrop";
            backdrop.style.position = Position.Absolute;
            backdrop.style.width = Length.Percent(100);
            backdrop.style.height = Length.Percent(100);
            backdrop.style.backgroundColor = new Color(0, 0, 0, 0.5f);
            backdrop.RegisterCallback<ClickEvent>(evt => evt.StopPropagation());

            modalRoot?.Insert(0, backdrop);
        }

        public void HideModal(UIType modalType, bool result = false)
        {
            if (uiPanels.TryGetValue(modalType, out UIToolkitPanel modal))
            {
                modal.Hide();
                modal.TriggerModalCallback(result);

                // バックドロップを削除
                var backdrop = modalRoot?.Q("modal-backdrop");
                backdrop?.RemoveFromHierarchy();

                // 元のルートに戻す
                modal.Element.RemoveFromHierarchy();
                rootElement.Add(modal.Element);
            }
        }

        public void ShowLoading(string message = "Loading...")
        {
            if (loadingTemplate != null && overlayRoot != null)
            {
                var loadingElement = loadingTemplate.CloneTree();
                loadingElement.name = "loading-overlay";

                var messageLabel = loadingElement.Q<Label>("loading-message");
                if (messageLabel != null)
                {
                    messageLabel.text = message;
                }

                overlayRoot.Add(loadingElement);
            }
        }

        public void HideLoading()
        {
            var loadingElement = overlayRoot?.Q("loading-overlay");
            loadingElement?.RemoveFromHierarchy();
        }

        private void ConfigureRuntimeTheme()
        {
            // Unity 6の新しいランタイムテーマ機能
            var panelSettings = mainUIDocument.panelSettings;
            if (panelSettings != null)
            {
                // スケーリング設定
                panelSettings.scaleMode = PanelScaleMode.ScaleWithScreenSize;
                panelSettings.referenceResolution = new Vector2(1920, 1080);

                // アンチエイリアシング
                panelSettings.targetTexture = null; // 画面に直接描画
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

        public void ShowNotification(
            string title,
            string message,
            float duration = 3f,
            NotificationType type = NotificationType.Info
        )
        {
            // Unity 6のUI Toolkitでの通知実装
            var notification = new VisualElement();
            notification.AddToClassList("notification");
            notification.AddToClassList($"notification--{type.ToString().ToLower()}");

            var titleLabel = new Label(title);
            titleLabel.AddToClassList("notification__title");
            notification.Add(titleLabel);

            var messageLabel = new Label(message);
            messageLabel.AddToClassList("notification__message");
            notification.Add(messageLabel);

            // アニメーション設定
            notification.style.opacity = 0;
            notification.style.translate = new Translate(0, -20);
            notification.style.transitionDuration = new List<TimeValue> { new TimeValue(300, TimeUnit.Millisecond) };
            notification.style.transitionProperty = new List<StylePropertyName>
            {
                new StylePropertyName("opacity"),
                new StylePropertyName("translate"),
            };

            overlayRoot?.Add(notification);

            // 表示アニメーション
            notification
                .schedule.Execute(() =>
                {
                    notification.style.opacity = 1;
                    notification.style.translate = new Translate(0, 0);
                })
                .StartingIn(10);

            // 自動非表示
            notification
                .schedule.Execute(() =>
                {
                    notification.style.opacity = 0;
                    notification.style.translate = new Translate(0, -20);
                })
                .StartingIn((long)(duration * 1000));

            notification
                .schedule.Execute(() =>
                {
                    notification.RemoveFromHierarchy();
                })
                .StartingIn((long)(duration * 1000 + 300));
        }

        public enum NotificationType
        {
            Info,
            Success,
            Warning,
            Error,
        }
    }

    /// <summary>
    /// UI Toolkit用のパネルクラス
    /// </summary>
    public class UIToolkitPanel
    {
        private UIType uiType;
        private VisualElement element;
        private Action<bool> modalCallback;

        public UIType UIType => uiType;
        public VisualElement Element => element;
        public bool IsVisible { get; private set; }

        public UIToolkitPanel(UIType type, VisualElement rootElement)
        {
            uiType = type;
            element = rootElement;
        }

        public void Initialize()
        {
            // 共通の初期化処理
            SetupCommonElements();

            // タイプ別の初期化
            OnInitialize();
        }

        private void SetupCommonElements()
        {
            // 戻るボタン
            var backButton = element.Q<Button>("back-button");
            if (backButton != null)
            {
                backButton.clicked += OnBackPressed;
            }

            // 閉じるボタン
            var closeButton = element.Q<Button>("close-button");
            if (closeButton != null)
            {
                closeButton.clicked += OnClosePressed;
            }
        }

        protected virtual void OnInitialize()
        {
            // サブクラスでオーバーライド
        }

        public void Show()
        {
            element.style.display = DisplayStyle.Flex;
            IsVisible = true;

            // Unity 6のアニメーション
            AnimateShow();

            OnShow();
        }

        public void Hide()
        {
            AnimateHide(() =>
            {
                element.style.display = DisplayStyle.None;
                IsVisible = false;
            });

            OnHide();
        }

        private void AnimateShow()
        {
            element.style.opacity = 0;
            element.style.scale = new Scale(new Vector3(0.9f, 0.9f, 1f));

            element
                .schedule.Execute(() =>
                {
                    element.style.opacity = 1;
                    element.style.scale = new Scale(Vector3.one);
                })
                .StartingIn(10);
        }

        private void AnimateHide(Action onComplete = null)
        {
            element.style.opacity = 1;
            element.style.scale = new Scale(Vector3.one);

            element
                .schedule.Execute(() =>
                {
                    element.style.opacity = 0;
                    element.style.scale = new Scale(new Vector3(0.9f, 0.9f, 1f));
                })
                .StartingIn(10);

            element
                .schedule.Execute(() =>
                {
                    onComplete?.Invoke();
                })
                .StartingIn(310);
        }

        protected virtual void OnShow()
        {
            // サブクラスでオーバーライド
        }

        protected virtual void OnHide()
        {
            // サブクラスでオーバーライド
        }

        private void OnBackPressed()
        {
            UIToolkitManager.Instance?.GoBack();
        }

        private void OnClosePressed()
        {
            Hide();
        }

        public void SetModalCallback(Action<bool> callback)
        {
            modalCallback = callback;
        }

        public void TriggerModalCallback(bool result)
        {
            modalCallback?.Invoke(result);
            modalCallback = null;
        }

        public void Cleanup()
        {
            // イベントの解除
            var backButton = element.Q<Button>("back-button");
            if (backButton != null)
            {
                backButton.clicked -= OnBackPressed;
            }

            var closeButton = element.Q<Button>("close-button");
            if (closeButton != null)
            {
                closeButton.clicked -= OnClosePressed;
            }
        }
    }
}
