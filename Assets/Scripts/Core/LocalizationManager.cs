using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using TMPro;
using UnityEngine;

namespace MerchantTails.Core
{
    /// <summary>
    /// 多言語対応を管理するシステム
    /// 日本語と英語の切り替えをサポート
    /// </summary>
    public class LocalizationManager : MonoBehaviour
    {
        private static LocalizationManager instance;
        public static LocalizationManager Instance => instance;

        [Header("Language Settings")]
        [SerializeField]
        private SystemLanguage defaultLanguage = SystemLanguage.Japanese;

        [SerializeField]
        private List<SystemLanguage> supportedLanguages = new List<SystemLanguage>
        {
            SystemLanguage.Japanese,
            SystemLanguage.English,
        };

        [Header("Resource Settings")]
        [SerializeField]
        private string localizationFolder = "Localization";

        [SerializeField]
        private string localizationFilePrefix = "locale_";

        [SerializeField]
        private string localizationFileExtension = ".json";

        [Header("Font Settings")]
        [SerializeField]
        private TMP_FontAsset japaneseFontAsset;

        [SerializeField]
        private TMP_FontAsset englishFontAsset;

        [SerializeField]
        private Font japaneseLegacyFont;

        [SerializeField]
        private Font englishLegacyFont;

        [Header("Text Settings")]
        [SerializeField]
        private float japaneseTextScale = 1f;

        [SerializeField]
        private float englishTextScale = 0.9f;

        [SerializeField]
        private bool autoAdjustTextSize = true;

        // 現在の言語
        private SystemLanguage currentLanguage;
        private Dictionary<string, string> currentLocalization = new Dictionary<string, string>();
        private Dictionary<SystemLanguage, Dictionary<string, string>> localizationCache = new Dictionary<SystemLanguage, Dictionary<string, string>>();

        // テキストコンポーネントの管理
        private List<LocalizedText> registeredTexts = new List<LocalizedText>();
        private Dictionary<string, List<LocalizedText>> textsByKey = new Dictionary<string, List<LocalizedText>>();

        public SystemLanguage CurrentLanguage => currentLanguage;
        public bool IsJapanese => currentLanguage == SystemLanguage.Japanese;
        public bool IsEnglish => currentLanguage == SystemLanguage.English;

        // イベント
        public event Action<SystemLanguage> OnLanguageChanged;

        private void Awake()
        {
            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;
            DontDestroyOnLoad(gameObject);

            Initialize();
        }

        private void OnDestroy()
        {
            if (instance == this)
            {
                instance = null;
            }
        }

        private void Initialize()
        {
            // 保存された言語設定を読み込み
            string savedLanguage = PlayerPrefs.GetString("Language", "");
            if (!string.IsNullOrEmpty(savedLanguage) && Enum.TryParse<SystemLanguage>(savedLanguage, out SystemLanguage saved))
            {
                currentLanguage = saved;
            }
            else
            {
                // システム言語を確認
                if (supportedLanguages.Contains(Application.systemLanguage))
                {
                    currentLanguage = Application.systemLanguage;
                }
                else
                {
                    currentLanguage = defaultLanguage;
                }
            }

            // 初期言語をロード
            LoadLanguage(currentLanguage);
        }

        /// <summary>
        /// 言語を変更
        /// </summary>
        public void SetLanguage(SystemLanguage language)
        {
            if (!supportedLanguages.Contains(language))
            {
                ErrorHandler.LogWarning($"Unsupported language: {language}", null, "LocalizationManager");
                return;
            }

            if (language == currentLanguage)
                return;

            currentLanguage = language;
            PlayerPrefs.SetString("Language", language.ToString());
            PlayerPrefs.Save();

            LoadLanguage(language);

            // イベント発行
            OnLanguageChanged?.Invoke(language);
            EventBus.Publish(new LanguageChangedEvent(language));

            ErrorHandler.LogInfo($"Language changed to: {language}", "LocalizationManager");
        }

        /// <summary>
        /// 言語データをロード
        /// </summary>
        private void LoadLanguage(SystemLanguage language)
        {
            // キャッシュチェック
            if (localizationCache.TryGetValue(language, out Dictionary<string, string> cached))
            {
                currentLocalization = cached;
                UpdateAllTexts();
                return;
            }

            // ファイルからロード
            string fileName = $"{localizationFilePrefix}{GetLanguageCode(language)}";
            string filePath = $"{localizationFolder}/{fileName}";

            TextAsset localizationFile = Resources.Load<TextAsset>(filePath);
            if (localizationFile == null)
            {
                ErrorHandler.LogError($"Localization file not found: {filePath}", null, "LocalizationManager");
                CreateDefaultLocalization(language);
                return;
            }

            try
            {
                // JSONをパース
                LocalizationData data = JsonUtility.FromJson<LocalizationData>(localizationFile.text);
                currentLocalization = new Dictionary<string, string>();

                foreach (var entry in data.entries)
                {
                    currentLocalization[entry.key] = entry.value;
                }

                // キャッシュに保存
                localizationCache[language] = currentLocalization;

                // すべてのテキストを更新
                UpdateAllTexts();
            }
            catch (Exception e)
            {
                ErrorHandler.LogError($"Failed to load localization: {e.Message}", e, "LocalizationManager");
                CreateDefaultLocalization(language);
            }
        }

        /// <summary>
        /// デフォルトのローカライゼーションを作成
        /// </summary>
        private void CreateDefaultLocalization(SystemLanguage language)
        {
            currentLocalization = new Dictionary<string, string>();

            if (language == SystemLanguage.Japanese)
            {
                // 日本語のデフォルト
                currentLocalization["game.title"] = "マーチャントテイル ～商人物語～";
                currentLocalization["menu.start"] = "ゲーム開始";
                currentLocalization["menu.continue"] = "続きから";
                currentLocalization["menu.settings"] = "設定";
                currentLocalization["menu.quit"] = "終了";
                currentLocalization["shop.name"] = "道具屋";
                currentLocalization["item.fruit"] = "くだもの";
                currentLocalization["item.potion"] = "ポーション";
                currentLocalization["item.weapon"] = "武器";
                currentLocalization["item.accessory"] = "アクセサリー";
                currentLocalization["item.magicbook"] = "魔法書";
                currentLocalization["item.gem"] = "宝石";
                currentLocalization["rank.apprentice"] = "見習い商人";
                currentLocalization["rank.skilled"] = "一人前商人";
                currentLocalization["rank.veteran"] = "ベテラン商人";
                currentLocalization["rank.master"] = "マスター商人";
                currentLocalization["season.spring"] = "春";
                currentLocalization["season.summer"] = "夏";
                currentLocalization["season.autumn"] = "秋";
                currentLocalization["season.winter"] = "冬";
                currentLocalization["phase.morning"] = "朝";
                currentLocalization["phase.afternoon"] = "昼";
                currentLocalization["phase.evening"] = "夕方";
                currentLocalization["phase.night"] = "夜";
                currentLocalization["ui.money"] = "所持金";
                currentLocalization["ui.day"] = "日目";
                currentLocalization["ui.buy"] = "購入";
                currentLocalization["ui.sell"] = "売却";
                currentLocalization["ui.confirm"] = "確認";
                currentLocalization["ui.cancel"] = "キャンセル";
                currentLocalization["ui.back"] = "戻る";
                currentLocalization["tutorial.welcome"] = "商人の世界へようこそ！";
                currentLocalization["tutorial.objective"] = "商品を安く買って高く売り、利益を上げましょう。";
            }
            else
            {
                // 英語のデフォルト
                currentLocalization["game.title"] = "Merchant Tales";
                currentLocalization["menu.start"] = "Start Game";
                currentLocalization["menu.continue"] = "Continue";
                currentLocalization["menu.settings"] = "Settings";
                currentLocalization["menu.quit"] = "Quit";
                currentLocalization["shop.name"] = "Item Shop";
                currentLocalization["item.fruit"] = "Fruit";
                currentLocalization["item.potion"] = "Potion";
                currentLocalization["item.weapon"] = "Weapon";
                currentLocalization["item.accessory"] = "Accessory";
                currentLocalization["item.magicbook"] = "Magic Book";
                currentLocalization["item.gem"] = "Gem";
                currentLocalization["rank.apprentice"] = "Apprentice Merchant";
                currentLocalization["rank.skilled"] = "Skilled Merchant";
                currentLocalization["rank.veteran"] = "Veteran Merchant";
                currentLocalization["rank.master"] = "Master Merchant";
                currentLocalization["season.spring"] = "Spring";
                currentLocalization["season.summer"] = "Summer";
                currentLocalization["season.autumn"] = "Autumn";
                currentLocalization["season.winter"] = "Winter";
                currentLocalization["phase.morning"] = "Morning";
                currentLocalization["phase.afternoon"] = "Afternoon";
                currentLocalization["phase.evening"] = "Evening";
                currentLocalization["phase.night"] = "Night";
                currentLocalization["ui.money"] = "Money";
                currentLocalization["ui.day"] = "Day";
                currentLocalization["ui.buy"] = "Buy";
                currentLocalization["ui.sell"] = "Sell";
                currentLocalization["ui.confirm"] = "Confirm";
                currentLocalization["ui.cancel"] = "Cancel";
                currentLocalization["ui.back"] = "Back";
                currentLocalization["tutorial.welcome"] = "Welcome to the world of merchants!";
                currentLocalization["tutorial.objective"] = "Buy low, sell high, and make profits!";
            }

            localizationCache[language] = currentLocalization;
            UpdateAllTexts();
        }

        /// <summary>
        /// ローカライズされたテキストを取得
        /// </summary>
        public string GetText(string key, params object[] args)
        {
            if (string.IsNullOrEmpty(key))
                return "";

            if (currentLocalization.TryGetValue(key, out string value))
            {
                // フォーマット引数がある場合
                if (args != null && args.Length > 0)
                {
                    try
                    {
                        return string.Format(value, args);
                    }
                    catch
                    {
                        return value;
                    }
                }
                return value;
            }

            ErrorHandler.LogWarning($"Localization key not found: {key}", "LocalizationManager");
            return $"[{key}]";
        }

        /// <summary>
        /// 複数形のテキストを取得
        /// </summary>
        public string GetPluralText(string key, int count)
        {
            string pluralKey = count == 1 ? $"{key}.singular" : $"{key}.plural";
            return GetText(pluralKey, count);
        }

        /// <summary>
        /// テキストコンポーネントを登録
        /// </summary>
        public void RegisterText(LocalizedText localizedText)
        {
            if (localizedText == null)
                return;

            if (!registeredTexts.Contains(localizedText))
            {
                registeredTexts.Add(localizedText);
            }

            // キーごとのリストに追加
            if (!string.IsNullOrEmpty(localizedText.LocalizationKey))
            {
                if (!textsByKey.ContainsKey(localizedText.LocalizationKey))
                {
                    textsByKey[localizedText.LocalizationKey] = new List<LocalizedText>();
                }
                textsByKey[localizedText.LocalizationKey].Add(localizedText);
            }

            // 即座に更新
            UpdateText(localizedText);
        }

        /// <summary>
        /// テキストコンポーネントの登録を解除
        /// </summary>
        public void UnregisterText(LocalizedText localizedText)
        {
            if (localizedText == null)
                return;

            registeredTexts.Remove(localizedText);

            if (!string.IsNullOrEmpty(localizedText.LocalizationKey) && textsByKey.ContainsKey(localizedText.LocalizationKey))
            {
                textsByKey[localizedText.LocalizationKey].Remove(localizedText);
            }
        }

        /// <summary>
        /// すべてのテキストを更新
        /// </summary>
        private void UpdateAllTexts()
        {
            foreach (var text in registeredTexts)
            {
                if (text != null)
                {
                    UpdateText(text);
                }
            }

            // フォント更新
            UpdateFonts();
        }

        /// <summary>
        /// 個別のテキストを更新
        /// </summary>
        private void UpdateText(LocalizedText localizedText)
        {
            if (localizedText == null || string.IsNullOrEmpty(localizedText.LocalizationKey))
                return;

            string text = GetText(localizedText.LocalizationKey, localizedText.FormatArgs);
            localizedText.SetText(text);
        }

        /// <summary>
        /// 特定のキーのテキストを更新
        /// </summary>
        public void UpdateTextsByKey(string key)
        {
            if (textsByKey.TryGetValue(key, out List<LocalizedText> texts))
            {
                foreach (var text in texts)
                {
                    if (text != null)
                    {
                        UpdateText(text);
                    }
                }
            }
        }

        /// <summary>
        /// フォントを更新
        /// </summary>
        private void UpdateFonts()
        {
            // TextMeshPro
            if (japaneseFontAsset != null && englishFontAsset != null)
            {
                TMP_FontAsset targetFont = IsJapanese ? japaneseFontAsset : englishFontAsset;
                float targetScale = IsJapanese ? japaneseTextScale : englishTextScale;

                foreach (var tmp in FindObjectsOfType<TextMeshProUGUI>())
                {
                    tmp.font = targetFont;
                    if (autoAdjustTextSize)
                    {
                        tmp.transform.localScale = Vector3.one * targetScale;
                    }
                }
            }

            // Legacy Text
            if (japaneseLegacyFont != null && englishLegacyFont != null)
            {
                Font targetFont = IsJapanese ? japaneseLegacyFont : englishLegacyFont;

                foreach (var text in FindObjectsOfType<UnityEngine.UI.Text>())
                {
                    text.font = targetFont;
                }
            }
        }

        /// <summary>
        /// 言語コードを取得
        /// </summary>
        private string GetLanguageCode(SystemLanguage language)
        {
            return language switch
            {
                SystemLanguage.Japanese => "ja",
                SystemLanguage.English => "en",
                _ => "en",
            };
        }

        /// <summary>
        /// サポートされている言語のリストを取得
        /// </summary>
        public List<LanguageOption> GetAvailableLanguages()
        {
            var options = new List<LanguageOption>();

            foreach (var lang in supportedLanguages)
            {
                options.Add(new LanguageOption
                {
                    language = lang,
                    displayName = GetLanguageDisplayName(lang),
                    isActive = lang == currentLanguage,
                });
            }

            return options;
        }

        /// <summary>
        /// 言語の表示名を取得
        /// </summary>
        private string GetLanguageDisplayName(SystemLanguage language)
        {
            return language switch
            {
                SystemLanguage.Japanese => "日本語",
                SystemLanguage.English => "English",
                _ => language.ToString(),
            };
        }

        /// <summary>
        /// ローカライゼーションファイルを保存（エディタ用）
        /// </summary>
        [ContextMenu("Save Localization File")]
        private void SaveLocalizationFile()
        {
#if UNITY_EDITOR
            var data = new LocalizationData
            {
                entries = currentLocalization.Select(kvp => new LocalizationEntry
                {
                    key = kvp.Key,
                    value = kvp.Value,
                }).ToList(),
            };

            string json = JsonUtility.ToJson(data, true);
            string path = $"Assets/Resources/{localizationFolder}/{localizationFilePrefix}{GetLanguageCode(currentLanguage)}{localizationFileExtension}";

            // ディレクトリ作成
            string dir = Path.GetDirectoryName(path);
            if (!Directory.Exists(dir))
            {
                Directory.CreateDirectory(dir);
            }

            File.WriteAllText(path, json);
            UnityEditor.AssetDatabase.Refresh();

            Debug.Log($"Localization saved: {path}");
#endif
        }
    }

    /// <summary>
    /// ローカライゼーションデータ
    /// </summary>
    [Serializable]
    public class LocalizationData
    {
        public List<LocalizationEntry> entries = new List<LocalizationEntry>();
    }

    /// <summary>
    /// ローカライゼーションエントリ
    /// </summary>
    [Serializable]
    public class LocalizationEntry
    {
        public string key;
        public string value;
    }

    /// <summary>
    /// 言語オプション
    /// </summary>
    [Serializable]
    public struct LanguageOption
    {
        public SystemLanguage language;
        public string displayName;
        public bool isActive;
    }

    /// <summary>
    /// ローカライズされたテキストコンポーネント
    /// </summary>
    public class LocalizedText : MonoBehaviour
    {
        [SerializeField]
        private string localizationKey;

        [SerializeField]
        private object[] formatArgs;

        private TextMeshProUGUI tmpText;
        private UnityEngine.UI.Text legacyText;

        public string LocalizationKey => localizationKey;
        public object[] FormatArgs => formatArgs;

        private void Awake()
        {
            tmpText = GetComponent<TextMeshProUGUI>();
            legacyText = GetComponent<UnityEngine.UI.Text>();
        }

        private void Start()
        {
            LocalizationManager.Instance?.RegisterText(this);
        }

        private void OnDestroy()
        {
            LocalizationManager.Instance?.UnregisterText(this);
        }

        public void SetText(string text)
        {
            if (tmpText != null)
            {
                tmpText.text = text;
            }
            else if (legacyText != null)
            {
                legacyText.text = text;
            }
        }

        public void SetKey(string key, params object[] args)
        {
            localizationKey = key;
            formatArgs = args;
            LocalizationManager.Instance?.UpdateTextsByKey(key);
        }
    }

    // イベント
    public class LanguageChangedEvent : BaseGameEvent
    {
        public SystemLanguage NewLanguage { get; }

        public LanguageChangedEvent(SystemLanguage newLanguage)
        {
            NewLanguage = newLanguage;
        }
    }
}
