using System;
using MerchantTails.Events;
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

        private static bool debugMode = true;
        private static int maxLogHistory = 100;
        private static System.Collections.Generic.Queue<LogEntry> logHistory =
            new System.Collections.Generic.Queue<LogEntry>();

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
            Debug.Log("[ErrorHandler] Error handling system initialized");
        }

        /// <summary>
        /// システム終了時のクリーンアップ
        /// </summary>
        public static void Cleanup()
        {
            Application.logMessageReceived -= HandleUnityLog;
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

            // Publish error event for other systems to handle
            if (EventBus != null)
            {
                try
                {
                    // Create error event (would need to define ErrorOccurredEvent)
                    // EventBus.Publish(new ErrorOccurredEvent(message, exception));
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

                if (MarketSystem.Instance == null)
                {
                    LogWarning("MarketSystem instance is null", "HealthCheck");
                    allHealthy = false;
                }

                if (InventorySystem.Instance == null)
                {
                    LogWarning("InventorySystem instance is null", "HealthCheck");
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

            try
            {
                switch (systemName.ToLower())
                {
                    case "gamemanager":
                        return RecoverGameManager();
                    case "timemanager":
                        return RecoverTimeManager();
                    case "marketsystem":
                        return RecoverMarketSystem();
                    case "inventorysystem":
                        return RecoverInventorySystem();
                    default:
                        LogWarning($"No recovery procedure defined for {systemName}", "Recovery");
                        return false;
                }
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
            if (MarketSystem.Instance == null && GameManager.Instance != null)
            {
                GameManager.Instance.gameObject.AddComponent<MarketSystem>();
                LogInfo("MarketSystem recovery attempted", "Recovery");
                return MarketSystem.Instance != null;
            }
            return MarketSystem.Instance != null;
        }

        private static bool RecoverInventorySystem()
        {
            if (InventorySystem.Instance == null && GameManager.Instance != null)
            {
                GameManager.Instance.gameObject.AddComponent<InventorySystem>();
                LogInfo("InventorySystem recovery attempted", "Recovery");
                return InventorySystem.Instance != null;
            }
            return InventorySystem.Instance != null;
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
    }
}
