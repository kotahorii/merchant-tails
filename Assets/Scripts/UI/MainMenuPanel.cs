using UnityEngine;
using UnityEngine.UI;
using MerchantTails.Core;
using MerchantTails.Data;

namespace MerchantTails.UI
{
    /// <summary>
    /// メインメニュー画面のUI制御
    /// ゲーム開始、継続、設定、終了機能を提供
    /// </summary>
    public class MainMenuPanel : UIPanel
    {
        [Header("Main Menu Buttons")]
        [SerializeField] private Button newGameButton;
        [SerializeField] private Button continueButton;
        [SerializeField] private Button tutorialButton;
        [SerializeField] private Button settingsButton;
        [SerializeField] private Button creditsButton;
        [SerializeField] private Button exitButton;
        
        [Header("Game Info Display")]
        [SerializeField] private Text gameVersionText;
        [SerializeField] private Text lastSaveDateText;
        [SerializeField] private GameObject saveDataPanel;
        [SerializeField] private Text playerNameText;
        [SerializeField] private Text playerMoneyText;
        [SerializeField] private Text playerRankText;
        
        [Header("Background")]
        [SerializeField] private Image backgroundImage;
        [SerializeField] private Sprite[] backgroundSprites;
        
        private bool hasSaveData = false;
        
        protected override void OnInitialize()
        {
            SetupButtons();
            CheckSaveData();
            SetupBackground();
            UpdateVersionInfo();
        }
        
        protected override void OnShow()
        {
            LogUIAction("Main Menu shown");
            CheckSaveData(); // 表示時に再チェック
            UpdateSaveDataDisplay();
        }
        
        private void SetupButtons()
        {
            if (newGameButton != null)
                newGameButton.onClick.AddListener(OnNewGamePressed);
            
            if (continueButton != null)
                continueButton.onClick.AddListener(OnContinuePressed);
            
            if (tutorialButton != null)
                tutorialButton.onClick.AddListener(OnTutorialPressed);
            
            if (settingsButton != null)
                settingsButton.onClick.AddListener(OnSettingsPressed);
            
            if (creditsButton != null)
                creditsButton.onClick.AddListener(OnCreditsPressed);
            
            if (exitButton != null)
                exitButton.onClick.AddListener(OnExitPressed);
        }
        
        private void CheckSaveData()
        {
            // セーブデータの存在確認
            hasSaveData = GameManager.Instance.HasSaveData();
            
            // Continueボタンの有効/無効制御
            if (continueButton != null)
            {
                continueButton.interactable = hasSaveData;
                
                // ボタンの見た目も変更
                var buttonImage = continueButton.GetComponent<Image>();
                if (buttonImage != null)
                {
                    buttonImage.color = hasSaveData ? Color.white : new Color(1f, 1f, 1f, 0.5f);
                }
            }
            
            // セーブデータパネルの表示/非表示
            if (saveDataPanel != null)
            {
                saveDataPanel.SetActive(hasSaveData);
            }
        }
        
        private void UpdateSaveData()
        {
            if (!hasSaveData) return;
            
            var playerData = GameManager.Instance.GetPlayerData();
            if (playerData == null) return;
            
            // プレイヤー情報の表示
            if (playerNameText != null)
                playerNameText.text = playerData.PlayerName;
            
            if (playerMoneyText != null)
                playerMoneyText.text = $"{playerData.CurrentMoney:N0}G";
            
            if (playerRankText != null)
                playerRankText.text = GetRankDisplayName(playerData.CurrentRank);
            
            // 最終セーブ日時の表示
            if (lastSaveDateText != null)
            {
                var lastSaveDate = GameManager.Instance.GetLastSaveDate();
                lastSaveDateText.text = $"最終プレイ: {lastSaveDate:yyyy/MM/dd HH:mm}";
            }
        }
        
        private string GetRankDisplayName(MerchantRank rank)
        {
            return rank switch
            {
                MerchantRank.Apprentice => "見習い商人",
                MerchantRank.Skilled => "一人前商人",
                MerchantRank.Veteran => "ベテラン商人",
                MerchantRank.Master => "マスター商人",
                _ => "商人"
            };
        }
        
        private void SetupBackground()
        {
            if (backgroundImage != null && backgroundSprites != null && backgroundSprites.Length > 0)
            {
                // 時間帯に応じて背景を変更
                var currentHour = System.DateTime.Now.Hour;
                int spriteIndex = currentHour switch
                {
                    >= 6 and < 12 => 0,  // 朝
                    >= 12 and < 18 => 1, // 昼
                    >= 18 and < 22 => 2, // 夕方
                    _ => 3               // 夜
                };
                
                if (spriteIndex < backgroundSprites.Length)
                {
                    backgroundImage.sprite = backgroundSprites[spriteIndex];
                }
            }
        }
        
        private void UpdateVersionInfo()
        {
            if (gameVersionText != null)
            {
                gameVersionText.text = $"Ver. {Application.version}";
            }
        }
        
        // ボタンイベントハンドラー
        private void OnNewGamePressed()
        {
            LogUIAction("New Game button pressed");
            
            if (hasSaveData)
            {
                // セーブデータがある場合は確認ダイアログを表示
                ShowNewGameConfirmDialog();
            }
            else
            {
                StartNewGame();
            }
        }
        
        private void OnContinuePressed()
        {
            LogUIAction("Continue button pressed");
            
            if (hasSaveData)
            {
                ContinueGame();
            }
        }
        
        private void OnTutorialPressed()
        {
            LogUIAction("Tutorial button pressed");
            StartTutorial();
        }
        
        private void OnSettingsPressed()
        {
            LogUIAction("Settings button pressed");
            UIManager.Instance.ShowModal(UIType.Settings);
        }
        
        private void OnCreditsPressed()
        {
            LogUIAction("Credits button pressed");
            ShowCredits();
        }
        
        private void OnExitPressed()
        {
            LogUIAction("Exit button pressed");
            ShowExitConfirmDialog();
        }
        
        // ゲーム制御メソッド
        private void StartNewGame()
        {
            UIManager.Instance.ShowLoading("新しいゲームを開始しています...");
            
            // 新規ゲーム開始の処理
            GameManager.Instance.StartNewGame();
        }
        
        private void ContinueGame()
        {
            UIManager.Instance.ShowLoading("セーブデータを読み込んでいます...");
            
            // セーブデータから継続
            bool loadSuccess = GameManager.Instance.LoadGame();
            
            if (!loadSuccess)
            {
                UIManager.Instance.HideLoading();
                ShowLoadErrorDialog();
            }
        }
        
        private void StartTutorial()
        {
            // チュートリアル開始
            GameManager.Instance.StartTutorial();
        }
        
        private void ShowCredits()
        {
            // クレジット表示（将来実装）
            ErrorHandler.LogInfo("Credits not implemented yet", "MainMenuPanel");
        }
        
        // 確認ダイアログ
        private void ShowNewGameConfirmDialog()
        {
            var confirmDialog = CreateConfirmDialog(
                "新しいゲーム",
                "現在のセーブデータが上書きされます。\n本当に新しいゲームを開始しますか？",
                "開始する",
                "キャンセル",
                (confirmed) =>
                {
                    if (confirmed)
                    {
                        StartNewGame();
                    }
                }
            );
        }
        
        private void ShowExitConfirmDialog()
        {
            var confirmDialog = CreateConfirmDialog(
                "ゲーム終了",
                "ゲームを終了しますか？",
                "終了する",
                "キャンセル",
                (confirmed) =>
                {
                    if (confirmed)
                    {
                        Application.Quit();
                        
                        #if UNITY_EDITOR
                        UnityEditor.EditorApplication.isPlaying = false;
                        #endif
                    }
                }
            );
        }
        
        private void ShowLoadErrorDialog()
        {
            var errorDialog = CreateConfirmDialog(
                "読み込みエラー",
                "セーブデータの読み込みに失敗しました。\nファイルが破損している可能性があります。",
                "OK",
                null,
                null
            );
        }
        
        private GameObject CreateConfirmDialog(string title, string message, string confirmText, string cancelText, System.Action<bool> callback)
        {
            // 簡易確認ダイアログの作成（将来的にはプリハブ化）
            var dialogGO = new GameObject("ConfirmDialog");
            var rectTransform = dialogGO.AddComponent<RectTransform>();
            var canvasGroup = dialogGO.AddComponent<CanvasGroup>();
            
            // ダイアログの基本設定
            rectTransform.SetParent(UIManager.Instance.transform, false);
            rectTransform.anchorMin = Vector2.zero;
            rectTransform.anchorMax = Vector2.one;
            rectTransform.sizeDelta = Vector2.zero;
            
            // 背景
            var bgImage = dialogGO.AddComponent<Image>();
            bgImage.color = new Color(0, 0, 0, 0.7f);
            
            // ダイアログパネル
            var panelGO = new GameObject("Panel");
            var panelRect = panelGO.AddComponent<RectTransform>();
            var panelImage = panelGO.AddComponent<Image>();
            
            panelRect.SetParent(rectTransform, false);
            panelRect.anchoredPosition = Vector2.zero;
            panelRect.sizeDelta = new Vector2(400, 200);
            panelImage.color = Color.white;
            
            // コールバック設定（簡易実装）
            if (callback != null)
            {
                var button = dialogGO.AddComponent<Button>();
                button.onClick.AddListener(() =>
                {
                    callback(true);
                    Destroy(dialogGO);
                });
            }
            
            return dialogGO;
        }
        
        private void UpdateSaveDataDisplay()
        {
            if (hasSaveData)
            {
                UpdateSaveData();
            }
        }
        
        private void OnDestroy()
        {
            // ボタンイベントの解除
            if (newGameButton != null)
                newGameButton.onClick.RemoveListener(OnNewGamePressed);
            
            if (continueButton != null)
                continueButton.onClick.RemoveListener(OnContinuePressed);
            
            if (tutorialButton != null)
                tutorialButton.onClick.RemoveListener(OnTutorialPressed);
            
            if (settingsButton != null)
                settingsButton.onClick.RemoveListener(OnSettingsPressed);
            
            if (creditsButton != null)
                creditsButton.onClick.RemoveListener(OnCreditsPressed);
            
            if (exitButton != null)
                exitButton.onClick.RemoveListener(OnExitPressed);
        }
    }
}