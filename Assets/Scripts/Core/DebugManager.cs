using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using MerchantTails.Data;
using MerchantTails.Events;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.Core
{
    /// <summary>
    /// デバッグ機能を管理するシステム
    /// チート機能、ログ表示、パフォーマンス監視などを提供
    /// </summary>
    public class DebugManager : MonoBehaviour
    {
        private static DebugManager instance;
        public static DebugManager Instance => instance;

        [Header("Debug Settings")]
        [SerializeField]
        private bool enableDebugMode = true;

        [SerializeField]
        private KeyCode debugMenuKey = KeyCode.F1;

        [SerializeField]
        private bool showFPS = true;

        [SerializeField]
        private bool showMemoryUsage = true;

        [SerializeField]
        private bool showPerformanceStats = true;

        [Header("Console Settings")]
        [SerializeField]
        private int maxConsoleLines = 100;

        [SerializeField]
        private float consoleHeight = 300f;

        [SerializeField]
        private Font consoleFont;

        [SerializeField]
        private int consoleFontSize = 12;

        [Header("Cheat Settings")]
        [SerializeField]
        private float moneyMultiplier = 1000f;

        [SerializeField]
        private int quickDayAdvance = 7;

        // UI要素
        private bool isDebugMenuOpen = false;
        private bool isConsoleOpen = false;
        private List<DebugLogEntry> consoleEntries = new List<DebugLogEntry>();
        private Vector2 consoleScrollPosition;
        private string consoleCommand = "";

        // パフォーマンス計測
        private float fpsUpdateInterval = 0.5f;
        private float fpsAccumulator = 0f;
        private int fpsFrameCount = 0;
        private float fpsTimer = 0f;
        private float currentFPS = 0f;

        // チートコマンド
        private Dictionary<string, Action<string[]>> cheatCommands = new Dictionary<string, Action<string[]>>();

        // デバッグ情報
        private StringBuilder debugInfoBuilder = new StringBuilder();
        private GUIStyle debugStyle;
        private GUIStyle consoleStyle;

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

            // ログハンドラーを解除
            Application.logMessageReceived -= HandleLog;
        }

        private void Initialize()
        {
            // デバッグビルドでのみ有効化
            if (!Debug.isDebugBuild)
            {
                enableDebugMode = false;
                return;
            }

            // ログハンドラーを登録
            Application.logMessageReceived += HandleLog;

            // チートコマンドを登録
            RegisterCheatCommands();

            // スタイル初期化
            InitializeStyles();
        }

        private void InitializeStyles()
        {
            debugStyle = new GUIStyle();
            debugStyle.normal.textColor = Color.white;
            debugStyle.fontSize = 14;
            debugStyle.fontStyle = FontStyle.Bold;

            consoleStyle = new GUIStyle();
            consoleStyle.normal.textColor = Color.white;
            consoleStyle.fontSize = consoleFontSize;
            if (consoleFont != null)
            {
                consoleStyle.font = consoleFont;
            }
        }

        private void Update()
        {
            if (!enableDebugMode)
                return;

            // デバッグメニューのトグル
            if (Input.GetKeyDown(debugMenuKey))
            {
                isDebugMenuOpen = !isDebugMenuOpen;
            }

            // コンソールのトグル（Tilde/Backquote）
            if (Input.GetKeyDown(KeyCode.BackQuote))
            {
                isConsoleOpen = !isConsoleOpen;
                if (isConsoleOpen)
                {
                    consoleCommand = "";
                }
            }

            // FPS計測
            if (showFPS)
            {
                UpdateFPS();
            }

            // デバッグキー処理
            HandleDebugKeys();
        }

        private void UpdateFPS()
        {
            fpsTimer += Time.deltaTime;
            fpsAccumulator += 1f / Time.deltaTime;
            fpsFrameCount++;

            if (fpsTimer >= fpsUpdateInterval)
            {
                currentFPS = fpsAccumulator / fpsFrameCount;
                fpsTimer = 0f;
                fpsAccumulator = 0f;
                fpsFrameCount = 0;
            }
        }

        private void HandleDebugKeys()
        {
            // 即座の時間操作
            if (Input.GetKey(KeyCode.LeftShift))
            {
                if (Input.GetKeyDown(KeyCode.T))
                {
                    TimeManager.Instance?.AdvanceTime(1f); // 1時間進める
                    LogDebug("Advanced time by 1 hour");
                }
                else if (Input.GetKeyDown(KeyCode.D))
                {
                    TimeManager.Instance?.AdvanceDay();
                    LogDebug("Advanced to next day");
                }
                else if (Input.GetKeyDown(KeyCode.S))
                {
                    TimeManager.Instance?.AdvanceSeason();
                    LogDebug("Advanced to next season");
                }
            }

            // 即座の金額操作
            if (Input.GetKey(KeyCode.LeftControl))
            {
                if (Input.GetKeyDown(KeyCode.M))
                {
                    AddMoney(moneyMultiplier);
                }
                else if (Input.GetKeyDown(KeyCode.N))
                {
                    AddMoney(-moneyMultiplier);
                }
            }

            // セーブ/ロード
            if (Input.GetKey(KeyCode.LeftAlt))
            {
                if (Input.GetKeyDown(KeyCode.S))
                {
                    SaveSystem.Instance?.QuickSave();
                    LogDebug("Quick save completed");
                }
                else if (Input.GetKeyDown(KeyCode.L))
                {
                    SaveSystem.Instance?.QuickLoad();
                    LogDebug("Quick load completed");
                }
            }
        }

        private void OnGUI()
        {
            if (!enableDebugMode)
                return;

            // FPS表示
            if (showFPS)
            {
                DrawFPS();
            }

            // メモリ使用量表示
            if (showMemoryUsage)
            {
                DrawMemoryUsage();
            }

            // パフォーマンス統計表示
            if (showPerformanceStats)
            {
                DrawPerformanceStats();
            }

            // デバッグメニュー
            if (isDebugMenuOpen)
            {
                DrawDebugMenu();
            }

            // コンソール
            if (isConsoleOpen)
            {
                DrawConsole();
            }
        }

        private void DrawFPS()
        {
            Color originalColor = GUI.color;
            GUI.color =
                currentFPS < 30 ? Color.red
                : currentFPS < 50 ? Color.yellow
                : Color.green;
            GUI.Label(new Rect(10, 10, 200, 25), $"FPS: {currentFPS:F1}", debugStyle);
            GUI.color = originalColor;
        }

        private void DrawMemoryUsage()
        {
            long totalMemory = GC.GetTotalMemory(false);
            string memoryText = $"Memory: {FormatBytes(totalMemory)}";
            GUI.Label(new Rect(10, 35, 200, 25), memoryText, debugStyle);
        }

        private void DrawPerformanceStats()
        {
            if (UpdateManager.Instance != null)
            {
                var stats = UpdateManager.Instance.GetStats();
                string perfText = $"Updates: {stats.normalUpdateCount} | Performance: {stats.performanceLevel}";
                GUI.Label(new Rect(10, 60, 300, 25), perfText, debugStyle);
            }
        }

        private void DrawDebugMenu()
        {
            float width = 400f;
            float height = 600f;
            float x = (Screen.width - width) / 2;
            float y = (Screen.height - height) / 2;

            GUI.Box(new Rect(x, y, width, height), "Debug Menu");

            float buttonY = y + 30;
            float buttonHeight = 30f;
            float spacing = 5f;

            // ゲーム状態操作
            GUI.Label(new Rect(x + 10, buttonY, width - 20, 25), "Game State", debugStyle);
            buttonY += 30;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Add 1000G"))
            {
                AddMoney(1000);
            }
            buttonY += buttonHeight + spacing;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Add 10000G"))
            {
                AddMoney(10000);
            }
            buttonY += buttonHeight + spacing;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Advance 1 Day"))
            {
                TimeManager.Instance?.AdvanceDay();
            }
            buttonY += buttonHeight + spacing;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Advance 7 Days"))
            {
                for (int i = 0; i < 7; i++)
                {
                    TimeManager.Instance?.AdvanceDay();
                }
            }
            buttonY += buttonHeight + spacing;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Advance Season"))
            {
                TimeManager.Instance?.AdvanceSeason();
            }
            buttonY += buttonHeight + spacing;

            // アイテム操作
            GUI.Label(new Rect(x + 10, buttonY + 10, width - 20, 25), "Items", debugStyle);
            buttonY += 40;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Add Random Items"))
            {
                AddRandomItems();
            }
            buttonY += buttonHeight + spacing;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Clear Inventory"))
            {
                ClearInventory();
            }
            buttonY += buttonHeight + spacing;

            // ランク操作
            GUI.Label(new Rect(x + 10, buttonY + 10, width - 20, 25), "Rank", debugStyle);
            buttonY += 40;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Set Apprentice"))
            {
                SetMerchantRank(MerchantRank.Apprentice);
            }
            buttonY += buttonHeight + spacing;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Set Skilled"))
            {
                SetMerchantRank(MerchantRank.Skilled);
            }
            buttonY += buttonHeight + spacing;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Set Veteran"))
            {
                SetMerchantRank(MerchantRank.Veteran);
            }
            buttonY += buttonHeight + spacing;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Set Master"))
            {
                SetMerchantRank(MerchantRank.Master);
            }
            buttonY += buttonHeight + spacing;

            // その他
            GUI.Label(new Rect(x + 10, buttonY + 10, width - 20, 25), "Other", debugStyle);
            buttonY += 40;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Trigger Random Event"))
            {
                TriggerRandomEvent();
            }
            buttonY += buttonHeight + spacing;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Reset Tutorial"))
            {
                ResetTutorial();
            }
            buttonY += buttonHeight + spacing;

            if (GUI.Button(new Rect(x + 10, buttonY, width - 20, buttonHeight), "Clear Save Data"))
            {
                ClearSaveData();
            }
            buttonY += buttonHeight + spacing;

            // 閉じるボタン
            if (GUI.Button(new Rect(x + 10, y + height - 40, width - 20, 30), "Close"))
            {
                isDebugMenuOpen = false;
            }
        }

        private void DrawConsole()
        {
            float width = Screen.width * 0.8f;
            float x = (Screen.width - width) / 2;
            float y = 10;

            // 背景
            GUI.Box(new Rect(x, y, width, consoleHeight), "Debug Console");

            // ログ表示エリア
            float logHeight = consoleHeight - 60;
            GUILayout.BeginArea(new Rect(x + 5, y + 25, width - 10, logHeight));
            consoleScrollPosition = GUILayout.BeginScrollView(consoleScrollPosition);

            foreach (var entry in consoleEntries)
            {
                Color originalColor = GUI.color;
                GUI.color = GetLogColor(entry.type);
                GUILayout.Label($"[{entry.timestamp:HH:mm:ss}] {entry.message}", consoleStyle);
                GUI.color = originalColor;
            }

            GUILayout.EndScrollView();
            GUILayout.EndArea();

            // コマンド入力
            float inputY = y + consoleHeight - 30;
            consoleCommand = GUI.TextField(new Rect(x + 5, inputY, width - 70, 25), consoleCommand);

            if (GUI.Button(new Rect(x + width - 60, inputY, 55, 25), "Execute") ||
                (Event.current.type == EventType.KeyDown && Event.current.keyCode == KeyCode.Return))
            {
                ExecuteConsoleCommand(consoleCommand);
                consoleCommand = "";
                GUI.FocusControl(null);
            }
        }

        private void HandleLog(string logString, string stackTrace, LogType type)
        {
            if (consoleEntries.Count >= maxConsoleLines)
            {
                consoleEntries.RemoveAt(0);
            }

            consoleEntries.Add(new DebugLogEntry
            {
                message = logString,
                stackTrace = stackTrace,
                type = type,
                timestamp = DateTime.Now,
            });

            // 自動スクロール
            consoleScrollPosition = new Vector2(0, float.MaxValue);
        }

        private Color GetLogColor(LogType type)
        {
            return type switch
            {
                LogType.Error or LogType.Exception => Color.red,
                LogType.Warning => Color.yellow,
                LogType.Log => Color.white,
                _ => Color.gray,
            };
        }

        // チートコマンド登録
        private void RegisterCheatCommands()
        {
            cheatCommands["money"] = (args) =>
            {
                if (args.Length > 0 && float.TryParse(args[0], out float amount))
                {
                    AddMoney(amount);
                }
            };

            cheatCommands["day"] = (args) =>
            {
                if (args.Length > 0 && int.TryParse(args[0], out int days))
                {
                    for (int i = 0; i < days; i++)
                    {
                        TimeManager.Instance?.AdvanceDay();
                    }
                }
            };

            cheatCommands["season"] = (args) =>
            {
                if (args.Length > 0)
                {
                    if (Enum.TryParse<Season>(args[0], true, out Season season))
                    {
                        SetSeason(season);
                    }
                }
                else
                {
                    TimeManager.Instance?.AdvanceSeason();
                }
            };

            cheatCommands["rank"] = (args) =>
            {
                if (args.Length > 0 && Enum.TryParse<MerchantRank>(args[0], true, out MerchantRank rank))
                {
                    SetMerchantRank(rank);
                }
            };

            cheatCommands["item"] = (args) =>
            {
                if (args.Length >= 2)
                {
                    if (Enum.TryParse<ItemType>(args[0], true, out ItemType type) &&
                        int.TryParse(args[1], out int quantity))
                    {
                        AddItem(type, quantity);
                    }
                }
            };

            cheatCommands["event"] = (args) =>
            {
                if (args.Length > 0)
                {
                    TriggerEvent(args[0]);
                }
                else
                {
                    TriggerRandomEvent();
                }
            };

            cheatCommands["save"] = (args) => SaveSystem.Instance?.QuickSave();
            cheatCommands["load"] = (args) => SaveSystem.Instance?.QuickLoad();
            cheatCommands["clear"] = (args) => consoleEntries.Clear();
            cheatCommands["help"] = (args) => ShowHelp();
        }

        private void ExecuteConsoleCommand(string command)
        {
            if (string.IsNullOrWhiteSpace(command))
                return;

            LogDebug($"> {command}");

            string[] parts = command.Split(' ');
            string cmd = parts[0].ToLower();
            string[] args = parts.Skip(1).ToArray();

            if (cheatCommands.TryGetValue(cmd, out Action<string[]> action))
            {
                try
                {
                    action(args);
                }
                catch (Exception e)
                {
                    LogError($"Command error: {e.Message}");
                }
            }
            else
            {
                LogError($"Unknown command: {cmd}");
            }
        }

        // チート機能
        private void AddMoney(float amount)
        {
            var gameManager = GameManager.Instance;
            if (gameManager != null && gameManager.PlayerData != null)
            {
                gameManager.PlayerData.CurrentMoney += amount;
                LogDebug($"Added {amount:C} to player money. New total: {gameManager.PlayerData.CurrentMoney:C}");
                EventBus.Publish(new MoneyChangedEvent(amount));
            }
        }

        private void SetMerchantRank(MerchantRank rank)
        {
            var gameManager = GameManager.Instance;
            if (gameManager != null && gameManager.PlayerData != null)
            {
                gameManager.PlayerData.CurrentRank = rank;
                LogDebug($"Set merchant rank to: {rank}");
                EventBus.Publish(new RankChangedEvent(rank));
            }
        }

        private void AddItem(ItemType type, int quantity)
        {
            var inventory = InventorySystem.Instance;
            if (inventory != null)
            {
                var item = new ItemData
                {
                    id = Guid.NewGuid().ToString(),
                    type = type,
                    basePrice = MarketSystem.Instance?.GetCurrentPrice(type) ?? 100f,
                    currentPrice = MarketSystem.Instance?.GetCurrentPrice(type) ?? 100f,
                    quantity = quantity,
                    quality = ItemQuality.Normal,
                    isInShop = false,
                };

                inventory.AddItem(item);
                LogDebug($"Added {quantity}x {type} to inventory");
            }
        }

        private void AddRandomItems()
        {
            foreach (ItemType type in Enum.GetValues(typeof(ItemType)))
            {
                int quantity = UnityEngine.Random.Range(5, 20);
                AddItem(type, quantity);
            }
        }

        private void ClearInventory()
        {
            var inventory = InventorySystem.Instance;
            if (inventory != null)
            {
                var items = inventory.GetAllItems().ToList();
                foreach (var item in items)
                {
                    inventory.RemoveItem(item.id);
                }
                LogDebug("Cleared all inventory");
            }
        }

        private void SetSeason(Season season)
        {
            var timeManager = TimeManager.Instance;
            if (timeManager != null)
            {
                // リフレクションでプライベートフィールドを設定
                var field = typeof(TimeManager).GetField("currentSeason", System.Reflection.BindingFlags.NonPublic | System.Reflection.BindingFlags.Instance);
                field?.SetValue(timeManager, season);
                EventBus.Publish(new SeasonChangedEvent(season));
                LogDebug($"Set season to: {season}");
            }
        }

        private void TriggerRandomEvent()
        {
            var eventSystem = EventSystem.Instance;
            if (eventSystem != null)
            {
                eventSystem.TriggerRandomEvent();
                LogDebug("Triggered random event");
            }
        }

        private void TriggerEvent(string eventId)
        {
            var eventSystem = EventSystem.Instance;
            if (eventSystem != null)
            {
                // イベントIDで特定のイベントをトリガー
                LogDebug($"Triggered event: {eventId}");
            }
        }

        private void ResetTutorial()
        {
            var tutorialSystem = TutorialSystem.Instance;
            if (tutorialSystem != null)
            {
                PlayerPrefs.DeleteKey("TutorialCompleted");
                PlayerPrefs.DeleteKey("TutorialStep");
                LogDebug("Tutorial progress reset");
            }
        }

        private void ClearSaveData()
        {
            SaveSystem.Instance?.DeleteAllSaves();
            PlayerPrefs.DeleteAll();
            LogDebug("All save data cleared");
        }

        private void ShowHelp()
        {
            LogDebug("Available commands:");
            LogDebug("  money <amount> - Add/subtract money");
            LogDebug("  day <count> - Advance days");
            LogDebug("  season [name] - Change/advance season");
            LogDebug("  rank <rank> - Set merchant rank");
            LogDebug("  item <type> <quantity> - Add items");
            LogDebug("  event [id] - Trigger event");
            LogDebug("  save - Quick save");
            LogDebug("  load - Quick load");
            LogDebug("  clear - Clear console");
            LogDebug("  help - Show this help");
        }

        // ユーティリティ
        private string FormatBytes(long bytes)
        {
            string[] sizes = { "B", "KB", "MB", "GB" };
            double len = bytes;
            int order = 0;

            while (len >= 1024 && order < sizes.Length - 1)
            {
                order++;
                len = len / 1024;
            }

            return $"{len:0.##} {sizes[order]}";
        }

        // パブリックログメソッド
        public void LogDebug(string message)
        {
            Debug.Log($"[DEBUG] {message}");
        }

        public void LogWarning(string message)
        {
            Debug.LogWarning($"[DEBUG] {message}");
        }

        public void LogError(string message)
        {
            Debug.LogError($"[DEBUG] {message}");
        }

        // 構造体
        private struct DebugLogEntry
        {
            public string message;
            public string stackTrace;
            public LogType type;
            public DateTime timestamp;
        }
    }
}
