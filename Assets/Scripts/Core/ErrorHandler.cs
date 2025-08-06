using System;
using System.Collections.Generic;
using System.IO;
using System.Text;
using UnityEngine;

namespace MerchantTails.Core
{
    /// <summary>
    /// ゲーム全体のエラーハンドリングとログ管理
    /// 例外処理、デバッグ情報、エラー回復処理を統合管理
    /// </summary>
    public static class ErrorHandler
    {
        public static event Action<string, LogLevel> OnErrorLogged;
        public static event Action<Exception> OnCriticalError;
        public static event Action<string> OnRecoveryAttempted;

        private static bool debugMode = true;
        private static int maxLogHistory = 100;
        private static Queue<LogEntry> logHistory = new Queue<LogEntry>();

        // エラー統計
        private static Dictionary<string, int> errorCounts = new Dictionary<string, int>();
        private static Dictionary<string, DateTime> lastErrorTimes = new Dictionary<string, DateTime>();
        private static int totalErrors = 0;
        private static int criticalErrors = 0;

        // ログファイル設定
        private static bool enableFileLogging = true;
        private static string logFilePath;
        private static StreamWriter logWriter;
        private static readonly object logLock = new object();

        public enum LogLevel
        {
            Debug,
            Info,
            Warning,
            Error,
            Critical,
        }

        public struct LogEntry
        {
            public DateTime timestamp;
            public string message;
            public LogLevel level;
            public string stackTrace;
        }

        /// <summary>
        /// システム初期化時のエラーハンドリング設定
        /// </summary>
        public static void Initialize()
        {
            Application.logMessageReceived += HandleUnityLog;

            // ログファイルパスの設定
            string logDirectory = Path.Combine(Application.persistentDataPath, "Logs");
            if (!Directory.Exists(logDirectory))
            {
                Directory.CreateDirectory(logDirectory);
            }

            logFilePath = Path.Combine(logDirectory, $"game_log_{DateTime.Now:yyyy-MM-dd_HH-mm-ss}.txt");

            if (enableFileLogging)
            {
                try
                {
                    logWriter = new StreamWriter(logFilePath, true);
                    logWriter.AutoFlush = true;
                }
                catch (Exception e)
                {
                    Debug.LogError($"Failed to initialize log file: {e.Message}");
                    enableFileLogging = false;
                }
            }

            Debug.Log("[ErrorHandler] Error handling system initialized");
        }

        /// <summary>
        /// システム終了時のクリーンアップ
        /// </summary>
        public static void Cleanup()
        {
            Application.logMessageReceived -= HandleUnityLog;

            lock (logLock)
            {
                if (logWriter != null)
                {
                    logWriter.Close();
                    logWriter.Dispose();
                    logWriter = null;
                }
            }
        }

        /// <summary>
        /// 安全な処理実行（例外キャッチ付き）
        /// </summary>
        public static bool SafeExecute(Action action, string context = "Unknown")
        {
            try
            {
                action?.Invoke();
                return true;
            }
            catch (Exception e)
            {
                LogError($"Error in {context}: {e.Message}", e);
                return false;
            }
        }

        /// <summary>
        /// 安全な処理実行（戻り値付き）
        /// </summary>
        public static T SafeExecute<T>(Func<T> func, T defaultValue = default(T), string context = "Unknown")
        {
            try
            {
                return func != null ? func() : defaultValue;
            }
            catch (Exception e)
            {
                LogError($"Error in {context}: {e.Message}", e);
                return defaultValue;
            }
        }

        /// <summary>
        /// デバッグログ出力
        /// </summary>
        public static void LogDebug(string message, string context = "")
        {
            if (debugMode)
            {
                string fullMessage = string.IsNullOrEmpty(context) ? message : $"[{context}] {message}";
                Log(fullMessage, LogLevel.Debug);
                Debug.Log(fullMessage);
            }
        }

        /// <summary>
        /// 情報ログ出力
        /// </summary>
        public static void LogInfo(string message, string context = "")
        {
            string fullMessage = string.IsNullOrEmpty(context) ? message : $"[{context}] {message}";
            Log(fullMessage, LogLevel.Info);
            Debug.Log(fullMessage);
        }

        /// <summary>
        /// 警告ログ出力
        /// </summary>
        public static void LogWarning(string message, string context = "")
        {
            string fullMessage = string.IsNullOrEmpty(context) ? message : $"[{context}] {message}";
            Log(fullMessage, LogLevel.Warning);
            Debug.LogWarning(fullMessage);
        }

        /// <summary>
        /// エラーログ出力
        /// </summary>
        public static void LogError(string message, Exception exception = null, string context = "")
        {
            string fullMessage = string.IsNullOrEmpty(context) ? message : $"[{context}] {message}";

            if (exception != null)
            {
                fullMessage += $"\nException: {exception.GetType().Name}\nStack Trace: {exception.StackTrace}";
            }

            Log(fullMessage, LogLevel.Error, exception?.StackTrace);
            Debug.LogError(fullMessage);

            // エラー統計を更新
            totalErrors++;
            UpdateErrorStatistics(context);

            // Publish error event for other systems to handle
            if (EventBus != null)
            {
                try
                {
                    EventBus.Publish(new ErrorOccurredEvent(message, exception, context));
                }
                catch
                {
                    // Avoid recursive error handling
                }
            }
        }

        /// <summary>
        /// 致命的エラーログ出力
        /// </summary>
        public static void LogCritical(string message, Exception exception = null, string context = "")
        {
            string fullMessage =
                $"CRITICAL ERROR - {(string.IsNullOrEmpty(context) ? message : $"[{context}] {message}")}";

            if (exception != null)
            {
                fullMessage += $"\nException: {exception.GetType().Name}\nStack Trace: {exception.StackTrace}";
            }

            Log(fullMessage, LogLevel.Critical, exception?.StackTrace);
            Debug.LogError(fullMessage);

            // エラー統計を更新
            totalErrors++;
            criticalErrors++;
            UpdateErrorStatistics(context);

            // 致命的エラーイベントを発行
            OnCriticalError?.Invoke(exception);

            // 自動セーブを試みる
            TryEmergencySave();

            // In a production game, this might trigger automatic save or crash reporting
        }

        /// <summary>
        /// システムの健全性チェック
        /// </summary>
        public static bool CheckSystemHealth()
        {
            bool allHealthy = true;

            try
            {
                // Check core systems
                if (GameManager.Instance == null)
                {
                    LogWarning("GameManager instance is null", "HealthCheck");
                    allHealthy = false;
                }

                if (TimeManager.Instance == null)
                {
                    LogWarning("TimeManager instance is null", "HealthCheck");
                    allHealthy = false;
                }

                var marketSystem = ServiceLocator.GetService<IMarketSystem>();
                if (marketSystem == null)
                {
                    LogWarning("MarketSystem service is null", "HealthCheck");
                    allHealthy = false;
                }

                var inventorySystem = ServiceLocator.GetService<IInventorySystem>();
                if (inventorySystem == null)
                {
                    LogWarning("InventorySystem service is null", "HealthCheck");
                    allHealthy = false;
                }

                // Check memory usage
                long memoryUsage = GC.GetTotalMemory(false);
                if (memoryUsage > 500 * 1024 * 1024) // 500MB threshold
                {
                    LogWarning($"High memory usage detected: {memoryUsage / (1024 * 1024)}MB", "HealthCheck");
                }

                if (allHealthy)
                {
                    LogDebug("System health check passed", "HealthCheck");
                }

                return allHealthy;
            }
            catch (Exception e)
            {
                LogError("Error during system health check", e, "HealthCheck");
                return false;
            }
        }

        /// <summary>
        /// エラー回復処理
        /// </summary>
        public static bool AttemptRecovery(string systemName)
        {
            LogInfo($"Attempting recovery for {systemName}", "Recovery");
            OnRecoveryAttempted?.Invoke(systemName);

            try
            {
                bool result = false;
                switch (systemName.ToLower())
                {
                    case "gamemanager":
                        result = RecoverGameManager();
                        break;
                    case "timemanager":
                        result = RecoverTimeManager();
                        break;
                    case "marketsystem":
                        result = RecoverMarketSystem();
                        break;
                    case "inventorysystem":
                        result = RecoverInventorySystem();
                        break;
                    case "savedata":
                        result = RecoverSaveData();
                        break;
                    case "all":
                        result = RecoverAllSystems();
                        break;
                    default:
                        LogWarning($"No recovery procedure defined for {systemName}", "Recovery");
                        return false;
                }

                if (result)
                {
                    LogInfo($"Recovery successful for {systemName}", "Recovery");
                }
                else
                {
                    LogError($"Recovery failed for {systemName}", null, "Recovery");
                }

                return result;
            }
            catch (Exception e)
            {
                LogError($"Recovery failed for {systemName}", e, "Recovery");
                return false;
            }
        }

        private static bool RecoverGameManager()
        {
            if (GameManager.Instance == null)
            {
                var gameObject = new GameObject("GameManager_Recovery");
                gameObject.AddComponent<GameManager>();
                LogInfo("GameManager recovery attempted", "Recovery");
                return GameManager.Instance != null;
            }
            return true;
        }

        private static bool RecoverTimeManager()
        {
            if (TimeManager.Instance == null && GameManager.Instance != null)
            {
                GameManager.Instance.gameObject.AddComponent<TimeManager>();
                LogInfo("TimeManager recovery attempted", "Recovery");
                return TimeManager.Instance != null;
            }
            return TimeManager.Instance != null;
        }

        private static bool RecoverMarketSystem()
        {
            var marketSystem = ServiceLocator.GetService<IMarketSystem>();
            if (marketSystem == null)
            {
                LogWarning("MarketSystem service not found", "Recovery");
                return false;
            }
            return true;
        }

        private static bool RecoverInventorySystem()
        {
            var inventorySystem = ServiceLocator.GetService<IInventorySystem>();
            if (inventorySystem == null)
            {
                LogWarning("InventorySystem service not found", "Recovery");
                return false;
            }
            return true;
        }

        private static void Log(string message, LogLevel level, string stackTrace = "")
        {
            var entry = new LogEntry
            {
                timestamp = DateTime.Now,
                message = message,
                level = level,
                stackTrace = stackTrace ?? "",
            };

            logHistory.Enqueue(entry);

            // Maintain log history size
            while (logHistory.Count > maxLogHistory)
            {
                logHistory.Dequeue();
            }

            // ファイルに書き込み
            if (enableFileLogging && logWriter != null)
            {
                lock (logLock)
                {
                    try
                    {
                        logWriter.WriteLine($"[{entry.timestamp:yyyy-MM-dd HH:mm:ss}] [{level}] {message}");
                        if (!string.IsNullOrEmpty(stackTrace))
                        {
                            logWriter.WriteLine($"Stack Trace: {stackTrace}");
                        }
                    }
                    catch
                    {
                        // ログ書き込みエラーは無視
                    }
                }
            }

            OnErrorLogged?.Invoke(message, level);
        }

        private static void HandleUnityLog(string logString, string stackTrace, LogType type)
        {
            LogLevel level = type switch
            {
                LogType.Error => LogLevel.Error,
                LogType.Exception => LogLevel.Critical,
                LogType.Warning => LogLevel.Warning,
                LogType.Log => LogLevel.Info,
                _ => LogLevel.Debug,
            };

            // Only log if it's not already handled by our system
            if (!logString.StartsWith("[") || level == LogLevel.Critical)
            {
                Log(logString, level, stackTrace);
            }
        }

        /// <summary>
        /// ログ履歴を取得
        /// </summary>
        public static LogEntry[] GetLogHistory()
        {
            return logHistory.ToArray();
        }

        /// <summary>
        /// デバッグモードの切り替え
        /// </summary>
        public static void SetDebugMode(bool enabled)
        {
            debugMode = enabled;
            LogInfo($"Debug mode {(enabled ? "enabled" : "disabled")}", "ErrorHandler");
        }

        // 新しく追加されたメソッド
        private static void UpdateErrorStatistics(string context)
        {
            string key = string.IsNullOrEmpty(context) ? "Unknown" : context;

            if (!errorCounts.ContainsKey(key))
            {
                errorCounts[key] = 0;
            }
            errorCounts[key]++;

            lastErrorTimes[key] = DateTime.Now;
        }

        private static void TryEmergencySave()
        {
            try
            {
                if (SaveSystem.Instance != null)
                {
                    SaveSystem.Instance.EmergencySave();
                    LogInfo("Emergency save completed", "Recovery");
                }
            }
            catch (Exception e)
            {
                LogError("Emergency save failed", e, "Recovery");
            }
        }

        private static bool RecoverSaveData()
        {
            try
            {
                if (SaveSystem.Instance != null)
                {
                    return SaveSystem.Instance.LoadBackup();
                }
                return false;
            }
            catch
            {
                return false;
            }
        }

        private static bool RecoverAllSystems()
        {
            bool allRecovered = true;

            allRecovered &= RecoverGameManager();
            allRecovered &= RecoverTimeManager();
            allRecovered &= RecoverMarketSystem();
            allRecovered &= RecoverInventorySystem();

            return allRecovered;
        }

        /// <summary>
        /// エラー統計を取得
        /// </summary>
        public static ErrorStatistics GetErrorStatistics()
        {
            return new ErrorStatistics
            {
                totalErrors = totalErrors,
                criticalErrors = criticalErrors,
                errorCounts = new Dictionary<string, int>(errorCounts),
                lastErrorTimes = new Dictionary<string, DateTime>(lastErrorTimes),
            };
        }

        /// <summary>
        /// エラー統計をリセット
        /// </summary>
        public static void ResetErrorStatistics()
        {
            totalErrors = 0;
            criticalErrors = 0;
            errorCounts.Clear();
            lastErrorTimes.Clear();
            LogInfo("Error statistics reset", "ErrorHandler");
        }

        /// <summary>
        /// ログファイルパスを取得
        /// </summary>
        public static string GetLogFilePath()
        {
            return logFilePath;
        }

        /// <summary>
        /// エラー発生率を計算
        /// </summary>
        public static float GetErrorRate(string context = null)
        {
            if (string.IsNullOrEmpty(context))
            {
                // 全体のエラー率
                float timeSinceStart = Time.realtimeSinceStartup;
                return timeSinceStart > 0 ? totalErrors / timeSinceStart : 0;
            }
            else
            {
                // 特定コンテキストのエラー率
                if (
                    errorCounts.TryGetValue(context, out int count)
                    && lastErrorTimes.TryGetValue(context, out DateTime lastTime)
                )
                {
                    float timeSinceFirst = (float)(DateTime.Now - lastTime).TotalSeconds;
                    return timeSinceFirst > 0 ? count / timeSinceFirst : 0;
                }
                return 0;
            }
        }
    }

    // 追加の構造体とクラス
    public struct ErrorStatistics
    {
        public int totalErrors;
        public int criticalErrors;
        public Dictionary<string, int> errorCounts;
        public Dictionary<string, DateTime> lastErrorTimes;
    }

    // エラーイベント
    public class ErrorOccurredEvent : BaseGameEvent
    {
        public string Message { get; }
        public Exception Exception { get; }
        public string Context { get; }

        public ErrorOccurredEvent(string message, Exception exception, string context)
        {
            Message = message;
            Exception = exception;
            Context = context;
        }
    }
}
