using System;
using MerchantTails.Core;
using TMPro;
using UnityEngine;
using UnityEngine.Audio;
using UnityEngine.UI;

namespace MerchantTails.UI
{
    /// <summary>
    /// 設定画面UI
    /// </summary>
    public class SettingsPanel : UIPanel
    {
        [Header("Audio Settings")]
        [SerializeField]
        private Slider masterVolumeSlider;

        [SerializeField]
        private TextMeshProUGUI masterVolumeText;

        [SerializeField]
        private Slider bgmVolumeSlider;

        [SerializeField]
        private TextMeshProUGUI bgmVolumeText;

        [SerializeField]
        private Slider sfxVolumeSlider;

        [SerializeField]
        private TextMeshProUGUI sfxVolumeText;

        [SerializeField]
        private AudioMixer audioMixer;

        [Header("Graphics Settings")]
        [SerializeField]
        private TMP_Dropdown resolutionDropdown;

        [SerializeField]
        private Toggle fullscreenToggle;

        [SerializeField]
        private TMP_Dropdown qualityDropdown;

        [SerializeField]
        private Toggle vsyncToggle;

        [Header("Game Settings")]
        [SerializeField]
        private Slider gameSpeedSlider;

        [SerializeField]
        private TextMeshProUGUI gameSpeedText;

        [SerializeField]
        private Toggle autoSaveToggle;

        [SerializeField]
        private TMP_Dropdown autoSaveIntervalDropdown;

        [SerializeField]
        private Toggle tutorialHintsToggle;

        [Header("Language Settings")]
        [SerializeField]
        private TMP_Dropdown languageDropdown;

        [Header("Control Buttons")]
        [SerializeField]
        private Button applyButton;

        [SerializeField]
        private Button resetButton;

        [SerializeField]
        private Button backButton;

        private GameSettings currentSettings;
        private GameSettings tempSettings;

        protected override void Awake()
        {
            base.Awake();
            SetupButtons();
            LoadSettings();
        }

        private void SetupButtons()
        {
            // スライダーのイベント設定
            if (masterVolumeSlider != null)
                masterVolumeSlider.onValueChanged.AddListener(OnMasterVolumeChanged);

            if (bgmVolumeSlider != null)
                bgmVolumeSlider.onValueChanged.AddListener(OnBGMVolumeChanged);

            if (sfxVolumeSlider != null)
                sfxVolumeSlider.onValueChanged.AddListener(OnSFXVolumeChanged);

            if (gameSpeedSlider != null)
                gameSpeedSlider.onValueChanged.AddListener(OnGameSpeedChanged);

            // トグルのイベント設定
            if (fullscreenToggle != null)
                fullscreenToggle.onValueChanged.AddListener(OnFullscreenChanged);

            if (vsyncToggle != null)
                vsyncToggle.onValueChanged.AddListener(OnVSyncChanged);

            if (autoSaveToggle != null)
                autoSaveToggle.onValueChanged.AddListener(OnAutoSaveChanged);

            if (tutorialHintsToggle != null)
                tutorialHintsToggle.onValueChanged.AddListener(OnTutorialHintsChanged);

            // ドロップダウンのイベント設定
            if (resolutionDropdown != null)
                resolutionDropdown.onValueChanged.AddListener(OnResolutionChanged);

            if (qualityDropdown != null)
                qualityDropdown.onValueChanged.AddListener(OnQualityChanged);

            if (autoSaveIntervalDropdown != null)
                autoSaveIntervalDropdown.onValueChanged.AddListener(OnAutoSaveIntervalChanged);

            if (languageDropdown != null)
                languageDropdown.onValueChanged.AddListener(OnLanguageChanged);

            // ボタンのイベント設定
            if (applyButton != null)
                applyButton.onClick.AddListener(ApplySettings);

            if (resetButton != null)
                resetButton.onClick.AddListener(ResetToDefaults);

            if (backButton != null)
                backButton.onClick.AddListener(OnBackButtonClicked);
        }

        private void LoadSettings()
        {
            currentSettings = GameSettings.Load();
            tempSettings = new GameSettings(currentSettings);

            InitializeResolutionOptions();
            InitializeQualityOptions();
            InitializeLanguageOptions();
            InitializeAutoSaveOptions();

            ApplySettingsToUI();
        }

        private void InitializeResolutionOptions()
        {
            if (resolutionDropdown == null)
                return;

            resolutionDropdown.ClearOptions();
            var resolutions = Screen.resolutions;
            var options = new System.Collections.Generic.List<string>();

            int currentResolutionIndex = 0;
            for (int i = 0; i < resolutions.Length; i++)
            {
                string option = $"{resolutions[i].width} x {resolutions[i].height}";
                options.Add(option);

                if (resolutions[i].width == Screen.width && resolutions[i].height == Screen.height)
                {
                    currentResolutionIndex = i;
                }
            }

            resolutionDropdown.AddOptions(options);
            resolutionDropdown.value = currentResolutionIndex;
            resolutionDropdown.RefreshShownValue();
        }

        private void InitializeQualityOptions()
        {
            if (qualityDropdown == null)
                return;

            qualityDropdown.ClearOptions();
            qualityDropdown.AddOptions(new System.Collections.Generic.List<string> { "低", "中", "高", "最高" });
        }

        private void InitializeLanguageOptions()
        {
            if (languageDropdown == null)
                return;

            languageDropdown.ClearOptions();
            languageDropdown.AddOptions(new System.Collections.Generic.List<string> { "日本語", "English" });
        }

        private void InitializeAutoSaveOptions()
        {
            if (autoSaveIntervalDropdown == null)
                return;

            autoSaveIntervalDropdown.ClearOptions();
            autoSaveIntervalDropdown.AddOptions(
                new System.Collections.Generic.List<string> { "1分", "3分", "5分", "10分", "15分" }
            );
        }

        private void ApplySettingsToUI()
        {
            // 音声設定
            if (masterVolumeSlider != null)
            {
                masterVolumeSlider.value = tempSettings.masterVolume;
                if (masterVolumeText != null)
                    masterVolumeText.text = $"{(int)(tempSettings.masterVolume * 100)}%";
            }

            if (bgmVolumeSlider != null)
            {
                bgmVolumeSlider.value = tempSettings.bgmVolume;
                if (bgmVolumeText != null)
                    bgmVolumeText.text = $"{(int)(tempSettings.bgmVolume * 100)}%";
            }

            if (sfxVolumeSlider != null)
            {
                sfxVolumeSlider.value = tempSettings.sfxVolume;
                if (sfxVolumeText != null)
                    sfxVolumeText.text = $"{(int)(tempSettings.sfxVolume * 100)}%";
            }

            // グラフィックス設定
            if (fullscreenToggle != null)
                fullscreenToggle.isOn = tempSettings.fullscreen;

            if (vsyncToggle != null)
                vsyncToggle.isOn = tempSettings.vsync;

            if (qualityDropdown != null)
                qualityDropdown.value = tempSettings.qualityLevel;

            // ゲーム設定
            if (gameSpeedSlider != null)
            {
                gameSpeedSlider.value = tempSettings.gameSpeed;
                if (gameSpeedText != null)
                    gameSpeedText.text = $"{tempSettings.gameSpeed:F1}x";
            }

            if (autoSaveToggle != null)
                autoSaveToggle.isOn = tempSettings.autoSave;

            if (autoSaveIntervalDropdown != null)
            {
                int intervalIndex = GetAutoSaveIntervalIndex(tempSettings.autoSaveInterval);
                autoSaveIntervalDropdown.value = intervalIndex;
            }

            if (tutorialHintsToggle != null)
                tutorialHintsToggle.isOn = tempSettings.showTutorialHints;

            // 言語設定
            if (languageDropdown != null)
                languageDropdown.value = tempSettings.language == "ja" ? 0 : 1;
        }

        private int GetAutoSaveIntervalIndex(int interval)
        {
            return interval switch
            {
                60 => 0, // 1分
                180 => 1, // 3分
                300 => 2, // 5分
                600 => 3, // 10分
                900 => 4, // 15分
                _ => 2, // デフォルト5分
            };
        }

        private int GetAutoSaveIntervalFromIndex(int index)
        {
            return index switch
            {
                0 => 60, // 1分
                1 => 180, // 3分
                2 => 300, // 5分
                3 => 600, // 10分
                4 => 900, // 15分
                _ => 300, // デフォルト5分
            };
        }

        #region Event Handlers

        private void OnMasterVolumeChanged(float value)
        {
            tempSettings.masterVolume = value;
            if (masterVolumeText != null)
                masterVolumeText.text = $"{(int)(value * 100)}%";

            // リアルタイムプレビュー
            ApplyAudioSettings();
        }

        private void OnBGMVolumeChanged(float value)
        {
            tempSettings.bgmVolume = value;
            if (bgmVolumeText != null)
                bgmVolumeText.text = $"{(int)(value * 100)}%";

            ApplyAudioSettings();
        }

        private void OnSFXVolumeChanged(float value)
        {
            tempSettings.sfxVolume = value;
            if (sfxVolumeText != null)
                sfxVolumeText.text = $"{(int)(value * 100)}%";

            ApplyAudioSettings();
        }

        private void OnGameSpeedChanged(float value)
        {
            tempSettings.gameSpeed = value;
            if (gameSpeedText != null)
                gameSpeedText.text = $"{value:F1}x";
        }

        private void OnFullscreenChanged(bool value)
        {
            tempSettings.fullscreen = value;
        }

        private void OnVSyncChanged(bool value)
        {
            tempSettings.vsync = value;
        }

        private void OnAutoSaveChanged(bool value)
        {
            tempSettings.autoSave = value;
            if (autoSaveIntervalDropdown != null)
                autoSaveIntervalDropdown.interactable = value;
        }

        private void OnTutorialHintsChanged(bool value)
        {
            tempSettings.showTutorialHints = value;
        }

        private void OnResolutionChanged(int index)
        {
            var resolutions = Screen.resolutions;
            if (index < resolutions.Length)
            {
                tempSettings.resolutionWidth = resolutions[index].width;
                tempSettings.resolutionHeight = resolutions[index].height;
            }
        }

        private void OnQualityChanged(int index)
        {
            tempSettings.qualityLevel = index;
        }

        private void OnAutoSaveIntervalChanged(int index)
        {
            tempSettings.autoSaveInterval = GetAutoSaveIntervalFromIndex(index);
        }

        private void OnLanguageChanged(int index)
        {
            tempSettings.language = index == 0 ? "ja" : "en";
        }

        #endregion

        private void ApplyAudioSettings()
        {
            if (audioMixer != null)
            {
                // デシベルに変換（0〜1を-80〜0dBに変換）
                float masterDB = tempSettings.masterVolume > 0 ? Mathf.Log10(tempSettings.masterVolume) * 20 : -80f;
                float bgmDB = tempSettings.bgmVolume > 0 ? Mathf.Log10(tempSettings.bgmVolume) * 20 : -80f;
                float sfxDB = tempSettings.sfxVolume > 0 ? Mathf.Log10(tempSettings.sfxVolume) * 20 : -80f;

                audioMixer.SetFloat("MasterVolume", masterDB);
                audioMixer.SetFloat("BGMVolume", bgmDB);
                audioMixer.SetFloat("SFXVolume", sfxDB);
            }
        }

        private void ApplySettings()
        {
            // 設定を保存
            currentSettings = new GameSettings(tempSettings);
            currentSettings.Save();

            // グラフィックス設定を適用
            Screen.SetResolution(
                currentSettings.resolutionWidth,
                currentSettings.resolutionHeight,
                currentSettings.fullscreen
            );

            QualitySettings.SetQualityLevel(currentSettings.qualityLevel);
            QualitySettings.vSyncCount = currentSettings.vsync ? 1 : 0;

            // ゲーム速度を適用
            Time.timeScale = currentSettings.gameSpeed;

            // 音声設定を適用
            ApplyAudioSettings();

            ErrorHandler.LogInfo("Settings applied successfully", "SettingsPanel");

            // 適用ボタンを一時的に無効化
            if (applyButton != null)
            {
                applyButton.interactable = false;
                Invoke(nameof(EnableApplyButton), 1f);
            }
        }

        private void EnableApplyButton()
        {
            if (applyButton != null)
                applyButton.interactable = true;
        }

        private void ResetToDefaults()
        {
            tempSettings = new GameSettings();
            ApplySettingsToUI();
            ErrorHandler.LogInfo("Settings reset to defaults", "SettingsPanel");
        }

        private void OnBackButtonClicked()
        {
            // 変更があるかチェック
            if (!tempSettings.Equals(currentSettings))
            {
                // 確認ダイアログを表示
                if (UIManager.Instance != null)
                {
                    UIManager.Instance.ShowConfirmDialog(
                        "変更を破棄",
                        "適用されていない変更があります。変更を破棄してよろしいですか？",
                        () => Hide(),
                        null
                    );
                }
            }
            else
            {
                Hide();
            }
        }

        protected override void OnShow()
        {
            base.OnShow();

            // 現在の設定を読み込み
            LoadSettings();
        }

        protected override void OnHide()
        {
            base.OnHide();

            // 一時的な設定をリセット
            tempSettings = new GameSettings(currentSettings);
        }
    }

    /// <summary>
    /// ゲーム設定データ
    /// </summary>
    [Serializable]
    public class GameSettings
    {
        // 音声設定
        public float masterVolume = 0.8f;
        public float bgmVolume = 0.7f;
        public float sfxVolume = 0.8f;

        // グラフィックス設定
        public int resolutionWidth = 1920;
        public int resolutionHeight = 1080;
        public bool fullscreen = false;
        public int qualityLevel = 2; // 0:低, 1:中, 2:高, 3:最高
        public bool vsync = true;

        // ゲーム設定
        public float gameSpeed = 1.0f;
        public bool autoSave = true;
        public int autoSaveInterval = 300; // 秒
        public bool showTutorialHints = true;

        // 言語設定
        public string language = "ja"; // ja, en

        public GameSettings() { }

        public GameSettings(GameSettings other)
        {
            masterVolume = other.masterVolume;
            bgmVolume = other.bgmVolume;
            sfxVolume = other.sfxVolume;
            resolutionWidth = other.resolutionWidth;
            resolutionHeight = other.resolutionHeight;
            fullscreen = other.fullscreen;
            qualityLevel = other.qualityLevel;
            vsync = other.vsync;
            gameSpeed = other.gameSpeed;
            autoSave = other.autoSave;
            autoSaveInterval = other.autoSaveInterval;
            showTutorialHints = other.showTutorialHints;
            language = other.language;
        }

        public void Save()
        {
            PlayerPrefs.SetFloat("Settings_MasterVolume", masterVolume);
            PlayerPrefs.SetFloat("Settings_BGMVolume", bgmVolume);
            PlayerPrefs.SetFloat("Settings_SFXVolume", sfxVolume);
            PlayerPrefs.SetInt("Settings_ResolutionWidth", resolutionWidth);
            PlayerPrefs.SetInt("Settings_ResolutionHeight", resolutionHeight);
            PlayerPrefs.SetInt("Settings_Fullscreen", fullscreen ? 1 : 0);
            PlayerPrefs.SetInt("Settings_QualityLevel", qualityLevel);
            PlayerPrefs.SetInt("Settings_VSync", vsync ? 1 : 0);
            PlayerPrefs.SetFloat("Settings_GameSpeed", gameSpeed);
            PlayerPrefs.SetInt("Settings_AutoSave", autoSave ? 1 : 0);
            PlayerPrefs.SetInt("Settings_AutoSaveInterval", autoSaveInterval);
            PlayerPrefs.SetInt("Settings_ShowTutorialHints", showTutorialHints ? 1 : 0);
            PlayerPrefs.SetString("Settings_Language", language);
            PlayerPrefs.Save();
        }

        public static GameSettings Load()
        {
            var settings = new GameSettings
            {
                masterVolume = PlayerPrefs.GetFloat("Settings_MasterVolume", 0.8f),
                bgmVolume = PlayerPrefs.GetFloat("Settings_BGMVolume", 0.7f),
                sfxVolume = PlayerPrefs.GetFloat("Settings_SFXVolume", 0.8f),
                resolutionWidth = PlayerPrefs.GetInt("Settings_ResolutionWidth", Screen.width),
                resolutionHeight = PlayerPrefs.GetInt("Settings_ResolutionHeight", Screen.height),
                fullscreen = PlayerPrefs.GetInt("Settings_Fullscreen", 0) == 1,
                qualityLevel = PlayerPrefs.GetInt("Settings_QualityLevel", 2),
                vsync = PlayerPrefs.GetInt("Settings_VSync", 1) == 1,
                gameSpeed = PlayerPrefs.GetFloat("Settings_GameSpeed", 1.0f),
                autoSave = PlayerPrefs.GetInt("Settings_AutoSave", 1) == 1,
                autoSaveInterval = PlayerPrefs.GetInt("Settings_AutoSaveInterval", 300),
                showTutorialHints = PlayerPrefs.GetInt("Settings_ShowTutorialHints", 1) == 1,
                language = PlayerPrefs.GetString("Settings_Language", "ja"),
            };

            return settings;
        }

        public bool Equals(GameSettings other)
        {
            if (other == null)
                return false;

            return masterVolume == other.masterVolume
                && bgmVolume == other.bgmVolume
                && sfxVolume == other.sfxVolume
                && resolutionWidth == other.resolutionWidth
                && resolutionHeight == other.resolutionHeight
                && fullscreen == other.fullscreen
                && qualityLevel == other.qualityLevel
                && vsync == other.vsync
                && gameSpeed == other.gameSpeed
                && autoSave == other.autoSave
                && autoSaveInterval == other.autoSaveInterval
                && showTutorialHints == other.showTutorialHints
                && language == other.language;
        }
    }
}
