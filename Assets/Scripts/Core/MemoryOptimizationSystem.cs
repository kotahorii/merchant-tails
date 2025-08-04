using System;
using System.Collections.Generic;
using MerchantTails.Events;
using UnityEngine;
using UnityEngine.Profiling;

namespace MerchantTails.Core
{
    /// <summary>
    /// メモリ管理とガベージコレクションを最適化するシステム
    /// </summary>
    public class MemoryOptimizationSystem : MonoBehaviour
    {
        private static MemoryOptimizationSystem instance;
        public static MemoryOptimizationSystem Instance => instance;

        [Header("Memory Settings")]
        [SerializeField]
        private long memoryWarningThreshold = 500L * 1024L * 1024L; // 500MB

        [SerializeField]
        private long memoryCriticalThreshold = 800L * 1024L * 1024L; // 800MB

        [SerializeField]
        private float memoryCheckInterval = 10f; // 10秒ごと

        [Header("GC Settings")]
        [SerializeField]
        private bool enableIncrementalGC = true;

        [SerializeField]
        private ulong incrementalGCTimeSlice = 3000000; // 3ms

        [SerializeField]
        private float manualGCInterval = 300f; // 5分ごと

        [SerializeField]
        private int gcGenerationThreshold = 2; // 第2世代まで

        [Header("String Pool Settings")]
        [SerializeField]
        private int stringPoolMaxSize = 1000;

        [SerializeField]
        private bool enableStringPooling = true;

        // メモリ状態
        private float lastMemoryCheckTime;
        private float lastManualGCTime;
        private long lastTotalMemory;
        private long peakMemoryUsage;
        private MemoryStatus currentMemoryStatus = MemoryStatus.Normal;

        // 文字列プール
        private Dictionary<string, string> stringPool = new Dictionary<string, string>();
        private Queue<string> stringPoolQueue = new Queue<string>();

        // オブジェクトプール（汎用）
        private Dictionary<Type, Stack<object>> reusableObjects = new Dictionary<Type, Stack<object>>();

        // 配列プール
        private Dictionary<int, Stack<Array>> arrayPools = new Dictionary<int, Stack<Array>>();

        // メモリプロファイリング
        private Dictionary<string, long> memoryAllocations = new Dictionary<string, long>();
        private bool isProfilingEnabled = false;

        public MemoryStatus CurrentMemoryStatus => currentMemoryStatus;
        public long CurrentMemoryUsage => GC.GetTotalMemory(false);
        public long PeakMemoryUsage => peakMemoryUsage;

        // イベント
        public event Action<MemoryStatus> OnMemoryStatusChanged;
        public event Action<long> OnMemoryWarning;
        public event Action OnMemoryCritical;

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

            // クリーンアップ
            ClearAllPools();
        }

        private void Initialize()
        {
            // インクリメンタルGCの設定
            if (enableIncrementalGC)
            {
                GarbageCollector.GCMode = GarbageCollector.Mode.Enabled;
                GarbageCollector.incrementalTimeSliceNanoseconds = incrementalGCTimeSlice;
            }

            // 初期メモリ状態を記録
            lastTotalMemory = GC.GetTotalMemory(false);
            peakMemoryUsage = lastTotalMemory;
        }

        private void Update()
        {
            // メモリチェック
            if (Time.time - lastMemoryCheckTime > memoryCheckInterval)
            {
                CheckMemoryStatus();
                lastMemoryCheckTime = Time.time;
            }

            // 定期的なGC
            if (Time.time - lastManualGCTime > manualGCInterval)
            {
                PerformScheduledGC();
                lastManualGCTime = Time.time;
            }
        }

        /// <summary>
        /// メモリ状態をチェック
        /// </summary>
        private void CheckMemoryStatus()
        {
            long currentMemory = GC.GetTotalMemory(false);

            // ピーク使用量の更新
            if (currentMemory > peakMemoryUsage)
            {
                peakMemoryUsage = currentMemory;
            }

            // ステータスの判定
            MemoryStatus newStatus = MemoryStatus.Normal;
            if (currentMemory > memoryCriticalThreshold)
            {
                newStatus = MemoryStatus.Critical;
                OnMemoryCritical?.Invoke();
                EmergencyMemoryCleanup();
            }
            else if (currentMemory > memoryWarningThreshold)
            {
                newStatus = MemoryStatus.Warning;
                OnMemoryWarning?.Invoke(currentMemory);
            }

            // ステータス変更通知
            if (newStatus != currentMemoryStatus)
            {
                currentMemoryStatus = newStatus;
                OnMemoryStatusChanged?.Invoke(newStatus);
                EventBus.Publish(new MemoryStatusChangedEvent(newStatus, currentMemory));

                ErrorHandler.LogWarning($"Memory status changed: {newStatus} ({FormatBytes(currentMemory)})", "MemoryOptimization");
            }

            // メモリ増加率のチェック
            float memoryGrowthRate = (float)(currentMemory - lastTotalMemory) / lastTotalMemory;
            if (memoryGrowthRate > 0.2f) // 20%以上の増加
            {
                ErrorHandler.LogWarning($"Rapid memory growth detected: {memoryGrowthRate:P0}", "MemoryOptimization");
                TriggerIncrementalGC();
            }

            lastTotalMemory = currentMemory;
        }

        /// <summary>
        /// スケジュールされたGCを実行
        /// </summary>
        private void PerformScheduledGC()
        {
            if (currentMemoryStatus == MemoryStatus.Normal)
            {
                // 通常時は軽いGCのみ
                GC.Collect(0, GCCollectionMode.Optimized);
            }
            else
            {
                // メモリ逼迫時はより積極的に
                GC.Collect(gcGenerationThreshold, GCCollectionMode.Forced);
                GC.WaitForPendingFinalizers();
                GC.Collect();
            }

            ErrorHandler.LogInfo($"Scheduled GC completed. Memory: {FormatBytes(GC.GetTotalMemory(false))}", "MemoryOptimization");
        }

        /// <summary>
        /// インクリメンタルGCをトリガー
        /// </summary>
        private void TriggerIncrementalGC()
        {
            if (enableIncrementalGC)
            {
                GarbageCollector.CollectIncremental(incrementalGCTimeSlice);
            }
        }

        /// <summary>
        /// 緊急メモリクリーンアップ
        /// </summary>
        private void EmergencyMemoryCleanup()
        {
            ErrorHandler.LogError("Emergency memory cleanup triggered", null, "MemoryOptimization");

            // すべてのキャッシュをクリア
            ResourceManager.Instance?.ClearAllCache();

            // 文字列プールをクリア
            ClearStringPool();

            // オブジェクトプールを縮小
            TrimObjectPools();

            // アセットのアンロード
            Resources.UnloadUnusedAssets();

            // 強制GC
            GC.Collect(GC.MaxGeneration, GCCollectionMode.Forced, true);
            GC.WaitForPendingFinalizers();
            GC.Collect();

            ErrorHandler.LogInfo($"Emergency cleanup completed. Memory: {FormatBytes(GC.GetTotalMemory(false))}", "MemoryOptimization");
        }

        // 文字列プール管理
        /// <summary>
        /// 文字列をプールから取得
        /// </summary>
        public string GetPooledString(string str)
        {
            if (!enableStringPooling || string.IsNullOrEmpty(str))
                return str;

            if (stringPool.TryGetValue(str, out string pooled))
            {
                return pooled;
            }

            // プールサイズ制限
            if (stringPool.Count >= stringPoolMaxSize)
            {
                RemoveOldestStringFromPool();
            }

            stringPool[str] = str;
            stringPoolQueue.Enqueue(str);
            return str;
        }

        private void RemoveOldestStringFromPool()
        {
            if (stringPoolQueue.Count > 0)
            {
                string oldest = stringPoolQueue.Dequeue();
                stringPool.Remove(oldest);
            }
        }

        private void ClearStringPool()
        {
            stringPool.Clear();
            stringPoolQueue.Clear();
        }

        // オブジェクトプール管理
        /// <summary>
        /// 再利用可能オブジェクトを取得
        /// </summary>
        public T GetReusable<T>() where T : class, new()
        {
            Type type = typeof(T);

            if (reusableObjects.TryGetValue(type, out Stack<object> pool) && pool.Count > 0)
            {
                return (T)pool.Pop();
            }

            return new T();
        }

        /// <summary>
        /// オブジェクトをプールに返却
        /// </summary>
        public void ReturnReusable<T>(T obj) where T : class
        {
            if (obj == null)
                return;

            Type type = typeof(T);

            if (!reusableObjects.ContainsKey(type))
            {
                reusableObjects[type] = new Stack<object>();
            }

            // リセット可能な場合はリセット
            if (obj is IPoolable poolable)
            {
                poolable.Reset();
            }

            reusableObjects[type].Push(obj);
        }

        // 配列プール管理
        /// <summary>
        /// 配列を取得
        /// </summary>
        public T[] GetArray<T>(int size)
        {
            int roundedSize = GetRoundedSize(size);

            if (arrayPools.TryGetValue(roundedSize, out Stack<Array> pool) && pool.Count > 0)
            {
                return (T[])pool.Pop();
            }

            return new T[roundedSize];
        }

        /// <summary>
        /// 配列を返却
        /// </summary>
        public void ReturnArray<T>(T[] array)
        {
            if (array == null)
                return;

            // 配列をクリア
            Array.Clear(array, 0, array.Length);

            int size = array.Length;
            if (!arrayPools.ContainsKey(size))
            {
                arrayPools[size] = new Stack<Array>();
            }

            arrayPools[size].Push(array);
        }

        private int GetRoundedSize(int size)
        {
            // 2のべき乗に丸める
            int rounded = 16;
            while (rounded < size)
            {
                rounded *= 2;
            }
            return rounded;
        }

        /// <summary>
        /// オブジェクトプールをトリム
        /// </summary>
        private void TrimObjectPools()
        {
            // 各プールのサイズを半分に
            foreach (var pool in reusableObjects.Values)
            {
                int toRemove = pool.Count / 2;
                for (int i = 0; i < toRemove; i++)
                {
                    if (pool.Count > 0)
                        pool.Pop();
                }
            }

            foreach (var pool in arrayPools.Values)
            {
                int toRemove = pool.Count / 2;
                for (int i = 0; i < toRemove; i++)
                {
                    if (pool.Count > 0)
                        pool.Pop();
                }
            }
        }

        /// <summary>
        /// すべてのプールをクリア
        /// </summary>
        private void ClearAllPools()
        {
            ClearStringPool();
            reusableObjects.Clear();
            arrayPools.Clear();
        }

        // メモリプロファイリング
        /// <summary>
        /// メモリプロファイリングを開始
        /// </summary>
        public void StartProfiling()
        {
            isProfilingEnabled = true;
            memoryAllocations.Clear();
            Profiler.enabled = true;
        }

        /// <summary>
        /// メモリプロファイリングを停止
        /// </summary>
        public void StopProfiling()
        {
            isProfilingEnabled = false;
            Profiler.enabled = false;
        }

        /// <summary>
        /// メモリ割り当てを記録
        /// </summary>
        public void RecordAllocation(string tag, long bytes)
        {
            if (!isProfilingEnabled)
                return;

            if (!memoryAllocations.ContainsKey(tag))
            {
                memoryAllocations[tag] = 0;
            }
            memoryAllocations[tag] += bytes;
        }

        /// <summary>
        /// プロファイリング結果を取得
        /// </summary>
        public Dictionary<string, long> GetProfilingResults()
        {
            return new Dictionary<string, long>(memoryAllocations);
        }

        // ユーティリティ
        /// <summary>
        /// メモリ統計を取得
        /// </summary>
        public MemoryStats GetMemoryStats()
        {
            long totalMemory = GC.GetTotalMemory(false);
            long monoHeap = Profiler.GetMonoHeapSizeLong();
            long monoUsed = Profiler.GetMonoUsedSizeLong();

            return new MemoryStats
            {
                totalMemory = totalMemory,
                peakMemory = peakMemoryUsage,
                monoHeapSize = monoHeap,
                monoUsedSize = monoUsed,
                gen0Collections = GC.CollectionCount(0),
                gen1Collections = GC.CollectionCount(1),
                gen2Collections = GC.CollectionCount(2),
                stringPoolSize = stringPool.Count,
                objectPoolCount = reusableObjects.Count,
                arrayPoolCount = arrayPools.Count,
            };
        }

        /// <summary>
        /// バイト数をフォーマット
        /// </summary>
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

        /// <summary>
        /// メモリ使用量をログ出力
        /// </summary>
        [ContextMenu("Log Memory Usage")]
        public void LogMemoryUsage()
        {
            var stats = GetMemoryStats();
            Debug.Log($"=== Memory Usage ===\n" +
                     $"Total: {FormatBytes(stats.totalMemory)}\n" +
                     $"Peak: {FormatBytes(stats.peakMemory)}\n" +
                     $"Mono Heap: {FormatBytes(stats.monoHeapSize)}\n" +
                     $"Mono Used: {FormatBytes(stats.monoUsedSize)}\n" +
                     $"GC Gen0: {stats.gen0Collections}\n" +
                     $"GC Gen1: {stats.gen1Collections}\n" +
                     $"GC Gen2: {stats.gen2Collections}\n" +
                     $"String Pool: {stats.stringPoolSize}\n" +
                     $"Object Pools: {stats.objectPoolCount}\n" +
                     $"Array Pools: {stats.arrayPoolCount}");
        }
    }

    // インターフェース
    public interface IPoolable
    {
        void Reset();
    }

    // 列挙型
    public enum MemoryStatus
    {
        Normal,
        Warning,
        Critical,
    }

    // 構造体
    [Serializable]
    public struct MemoryStats
    {
        public long totalMemory;
        public long peakMemory;
        public long monoHeapSize;
        public long monoUsedSize;
        public int gen0Collections;
        public int gen1Collections;
        public int gen2Collections;
        public int stringPoolSize;
        public int objectPoolCount;
        public int arrayPoolCount;
    }

    // イベント
    public class MemoryStatusChangedEvent : BaseGameEvent
    {
        public MemoryStatus Status { get; }
        public long MemoryUsage { get; }

        public MemoryStatusChangedEvent(MemoryStatus status, long memoryUsage)
        {
            Status = status;
            MemoryUsage = memoryUsage;
        }
    }
}
