using System;
using MerchantTails.Core;
using UnityEngine;
using UnityEngine.UIElements;

namespace MerchantTails.UI.Panels
{
    /// <summary>
    /// メインメニューパネルのUI Toolkit実装
    /// Unity 6の新機能を活用したモダンなUI
    /// </summary>
    public class MainMenuPanel : UIToolkitPanel
    {
        private Button newGameButton;
        private Button continueButton;
        private Button settingsButton;
        private Button creditsButton;
        private Button exitButton;
        private Label versionLabel;

        protected override void OnInitialize()
        {
            base.OnInitialize();
            
            // UI要素の取得
            newGameButton = Element.Q<Button>("new-game-button");
            continueButton = Element.Q<Button>("continue-button");
            settingsButton = Element.Q<Button>("settings-button");
            creditsButton = Element.Q<Button>("credits-button");
            exitButton = Element.Q<Button>("exit-button");
            versionLabel = Element.Q<Label>("version-label");
            
            // イベントハンドラの設定
            SetupEventHandlers();
            
            // 初期状態の設定
            UpdateUI();
        }

        private void SetupEventHandlers()
        {
            if (newGameButton != null)
            {
                newGameButton.clicked += OnNewGameClicked;
            }
            
            if (continueButton != null)
            {
                continueButton.clicked += OnContinueClicked;
            }
            
            if (settingsButton != null)
            {
                settingsButton.clicked += OnSettingsClicked;
            }
            
            if (creditsButton != null)
            {
                creditsButton.clicked += OnCreditsClicked;
            }
            
            if (exitButton != null)
            {
                exitButton.clicked += OnExitClicked;
            }
        }

        protected override void OnShow()
        {
            base.OnShow();
            
            // BGMの再生
            // AudioManager.Instance?.PlayBGM("MainMenuTheme");
            
            UpdateUI();
            
            // Unity 6のアニメーション機能を使用
            AnimateMenuEntrance();
        }

        private void UpdateUI()
        {
            // セーブデータの確認
            bool hasSaveData = SaveSystem.Instance?.HasSaveData ?? false;
            
            if (continueButton != null)
            {
                continueButton.SetEnabled(hasSaveData);
                continueButton.style.opacity = hasSaveData ? 1f : 0.5f;
            }
            
            // バージョン情報の更新
            if (versionLabel != null)
            {
                versionLabel.text = $"Version {Application.version}";
            }
        }

        private void AnimateMenuEntrance()
        {
            // ロゴのアニメーション
            var logoContainer = Element.Q<VisualElement>("logo-container");
            if (logoContainer != null)
            {
                logoContainer.style.opacity = 0;
                logoContainer.style.translate = new Translate(0, -50);
                
                logoContainer.schedule.Execute(() => {
                    logoContainer.style.opacity = 1;
                    logoContainer.style.translate = new Translate(0, 0);
                }).StartingIn(100);
            }
            
            // ボタンの順次アニメーション
            var buttons = Element.Query<Button>(className: "menu-button").ToList();
            for (int i = 0; i < buttons.Count; i++)
            {
                var button = buttons[i];
                button.style.opacity = 0;
                button.style.translate = new Translate(-50, 0);
                
                int delay = 200 + (i * 50);
                button.schedule.Execute(() => {
                    button.style.opacity = 1;
                    button.style.translate = new Translate(0, 0);
                }).StartingIn(delay);
            }
        }

        private void OnNewGameClicked()
        {
            ErrorHandler.LogInfo("New Game clicked", "MainMenuPanel");
            
            // 新規ゲーム開始の確認ダイアログ
            UIToolkitManager.Instance?.ShowModal(UIType.Confirmation, (confirmed) => {
                if (confirmed)
                {
                    StartNewGame();
                }
            });
        }

        private void OnContinueClicked()
        {
            ErrorHandler.LogInfo("Continue clicked", "MainMenuPanel");
            
            // セーブデータの読み込み
            _ = LoadGameAsync();
        }

        private async void LoadGameAsync()
        {
            UIToolkitManager.Instance?.ShowLoading("Loading save data...");
            
            bool success = await SaveSystem.Instance.LoadAsync();
            
            UIToolkitManager.Instance?.HideLoading();
            
            if (success)
            {
                // ゲーム画面へ遷移
                GameManager.Instance?.ChangeState(GameState.Shopping);
            }
            else
            {
                UIToolkitManager.Instance?.ShowNotification(
                    "Load Failed",
                    "Failed to load save data.",
                    3f,
                    UIToolkitManager.NotificationType.Error
                );
            }
        }

        private void OnSettingsClicked()
        {
            ErrorHandler.LogInfo("Settings clicked", "MainMenuPanel");
            UIToolkitManager.Instance?.ShowPanel(UIType.Settings);
        }

        private void OnCreditsClicked()
        {
            ErrorHandler.LogInfo("Credits clicked", "MainMenuPanel");
            UIToolkitManager.Instance?.ShowPanel(UIType.Credits);
        }

        private void OnExitClicked()
        {
            ErrorHandler.LogInfo("Exit clicked", "MainMenuPanel");
            
            // 終了確認ダイアログ
            UIToolkitManager.Instance?.ShowModal(UIType.Confirmation, (confirmed) => {
                if (confirmed)
                {
#if UNITY_EDITOR
                    UnityEditor.EditorApplication.isPlaying = false;
#else
                    Application.Quit();
#endif
                }
            });
        }

        private void StartNewGame()
        {
            // 新規ゲームの初期化
            GameManager.Instance?.StartNewGame();
            
            // チュートリアルの確認
            UIToolkitManager.Instance?.ShowModal(UIType.Tutorial, (showTutorial) => {
                if (showTutorial)
                {
                    GameManager.Instance?.ChangeState(GameState.Tutorial);
                }
                else
                {
                    GameManager.Instance?.ChangeState(GameState.Shopping);
                }
            });
        }

        public new void Cleanup()
        {
            
            // イベントハンドラの解除
            if (newGameButton != null)
            {
                newGameButton.clicked -= OnNewGameClicked;
            }
            
            if (continueButton != null)
            {
                continueButton.clicked -= OnContinueClicked;
            }
            
            if (settingsButton != null)
            {
                settingsButton.clicked -= OnSettingsClicked;
            }
            
            if (creditsButton != null)
            {
                creditsButton.clicked -= OnCreditsClicked;
            }
            
            if (exitButton != null)
            {
                exitButton.clicked -= OnExitClicked;
            }
        }
    }
}