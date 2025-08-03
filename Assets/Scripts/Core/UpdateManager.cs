using System;
using System.Collections.Generic;
using MerchantTails.Events;
using UnityEngine;

namespace MerchantTails.Core
{
    /// <summary>
    /// Update処理を一元管理し、パフォーマンスを最適化するシステム
    /// フレームレート別の更新処理とオブジェクトプーリングを提供
    /// </summary>
    public class UpdateManager : MonoBehaviour
    {
        private static UpdateManager instance;
        public static UpdateManager Instance => instance;

        [Header("Update Settings")]
        [SerializeField]
        private int targetFrameRate = 60;

        [SerializeField]
        private float fixedUpdateRate = 0.02f; // 50Hz

        [SerializeField]
        private float slowUpdateRate = 1f; // 1Hz

        [SerializeField]
        private float verySlowUpdateRate = 5f; // 0.2Hz

        [Header("Performance Settings")]
        [SerializeField]
        private bool enableDynamicFramerate = true;

        [SerializeField]
        private int lowPerformanceThreshold = 30; // FPS

        [SerializeField]
        private int criticalPerformanceThreshold = 20; // FPS

        [SerializeField]
        private float performanceCheckInterval = 2f;

        // 更新対象の管理
        private HashSet<IUpdatable> updatables = new HashSet<IUpdatable>();
        private HashSet<IFixedUpdatable> fixedUpdatables = new HashSet<IFixedUpdatable>();
        private HashSet<ISlowUpdatable> slowUpdatables = new HashSet<ISlowUpdatable>();
        private HashSet<IVerySlowUpdatable> verySlowUpdatables = new HashSet<IVerySlowUpdatable>();

        // 優先度付き更新リスト
        private SortedList<int, List<IUpdatable>> prioritizedUpdatables = new SortedList<int, List<IUpdatable>>();

        // タイミング管理
        private float slowUpdateTimer = 0f;
        private float verySlowUpdateTimer = 0f;
        private float performanceCheckTimer = 0f;
        private float deltaTime = 0f;
        private float currentFPS = 0f;

        // パフォーマンス統計
        private int frameCount = 0;
        private float fpsAccumulator = 0f;
        private PerformanceLevel currentPerformanceLevel = PerformanceLevel.Normal;

        // オブジェクトプール
        private Dictionary<Type, ObjectPool> objectPools = new Dictionary<Type, ObjectPool>();

        public float DeltaTime => deltaTime;
        public float CurrentFPS => currentFPS;
        public PerformanceLevel CurrentPerformanceLevel => currentPerformanceLevel;

        // イベント
        public event Action<PerformanceLevel> OnPerformanceLevelChanged;

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
            // フレームレート設定
            Application.targetFrameRate = targetFrameRate;
            QualitySettings.vSyncCount = 0;

            // Fixed Update レート設定
            Time.fixedDeltaTime = fixedUpdateRate;
        }

        private void Update()
        {
            // デルタタイム計算
            deltaTime = Time.deltaTime;

            // FPS計算
            UpdateFPS();

            // パフォーマンスチェック
            performanceCheckTimer += deltaTime;
            if (performanceCheckTimer >= performanceCheckInterval)
            {
                CheckPerformance();
                performanceCheckTimer = 0f;
            }

            // 通常更新
            UpdateNormal();

            // スロー更新
            slowUpdateTimer += deltaTime;
            if (slowUpdateTimer >= slowUpdateRate)
            {
                UpdateSlow();
                slowUpdateTimer = 0f;
            }

            // ベリースロー更新
            verySlowUpdateTimer += deltaTime;
            if (verySlowUpdateTimer >= verySlowUpdateRate)
            {
                UpdateVerySlow();
                verySlowUpdateTimer = 0f;
            }
        }

        private void FixedUpdate()
        {
            UpdateFixed();
        }

        /// <summary>
        /// 通常更新処理
        /// </summary>
        private void UpdateNormal()
        {
            // 優先度順に更新
            foreach (var priority in prioritizedUpdatables)
            {
                foreach (var updatable in priority.Value)
                {
                    if (updatable.IsActive)
                    {
                        try
                        {
                            updatable.OnUpdate(deltaTime);
                        }
                        catch (Exception e)
                        {
                            ErrorHandler.LogError($"Update error: {e.Message}", "UpdateManager");
                        }
                    }
                }
            }

            // 優先度なしの更新
            foreach (var updatable in updatables)
            {
                if (updatable.IsActive)
                {
                    try
                    {
                        updatable.OnUpdate(deltaTime);
                    }
                    catch (Exception e)
                    {
                        ErrorHandler.LogError($"Update error: {e.Message}", "UpdateManager");
                    }
                }
            }
        }

        /// <summary>
        /// 固定更新処理
        /// </summary>
        private void UpdateFixed()
        {
            foreach (var fixedUpdatable in fixedUpdatables)
            {
                if (fixedUpdatable.IsActive)
                {
                    try
                    {
                        fixedUpdatable.OnFixedUpdate(fixedUpdateRate);
                    }
                    catch (Exception e)
                    {
                        ErrorHandler.LogError($"Fixed update error: {e.Message}", "UpdateManager");
                    }
                }
            }
        }

        /// <summary>
        /// スロー更新処理
        /// </summary>
        private void UpdateSlow()
        {
            foreach (var slowUpdatable in slowUpdatables)
            {
                if (slowUpdatable.IsActive)
                {
                    try
                    {
                        slowUpdatable.OnSlowUpdate();
                    }
                    catch (Exception e)
                    {
                        ErrorHandler.LogError($"Slow update error: {e.Message}", "UpdateManager");
                    }
                }
            }
        }

        /// <summary>
        /// ベリースロー更新処理
        /// </summary>
        private void UpdateVerySlow()
        {
            foreach (var verySlowUpdatable in verySlowUpdatables)
            {
                if (verySlowUpdatable.IsActive)
                {
                    try
                    {
                        verySlowUpdatable.OnVerySlowUpdate();
                    }
                    catch (Exception e)
                    {
                        ErrorHandler.LogError($"Very slow update error: {e.Message}", "UpdateManager");
                    }
                }
            }
        }

        /// <summary>
        /// FPSを更新
        /// </summary>
        private void UpdateFPS()
        {
            frameCount++;
            fpsAccumulator += 1f / deltaTime;

            if (frameCount >= 30) // 30フレームごとに平均を計算
            {
                currentFPS = fpsAccumulator / frameCount;
                frameCount = 0;
                fpsAccumulator = 0f;
            }
        }

        /// <summary>
        /// パフォーマンスをチェック
        /// </summary>
        private void CheckPerformance()
        {
            PerformanceLevel newLevel = PerformanceLevel.Normal;

            if (currentFPS < criticalPerformanceThreshold)
            {
                newLevel = PerformanceLevel.Critical;
            }
            else if (currentFPS < lowPerformanceThreshold)
            {
                newLevel = PerformanceLevel.Low;
            }

            if (newLevel != currentPerformanceLevel)
            {
                currentPerformanceLevel = newLevel;
                OnPerformanceLevelChanged?.Invoke(currentPerformanceLevel);
                AdjustPerformanceSettings();

                EventBus.Publish(new PerformanceLevelChangedEvent(currentPerformanceLevel));
                ErrorHandler.LogWarning($"Performance level changed: {currentPerformanceLevel}", "UpdateManager");
            }
        }

        /// <summary>
        /// パフォーマンス設定を調整
        /// </summary>
        private void AdjustPerformanceSettings()
        {
            switch (currentPerformanceLevel)
            {
                case PerformanceLevel.Normal:
                    slowUpdateRate = 1f;
                    verySlowUpdateRate = 5f;
                    break;

                case PerformanceLevel.Low:
                    slowUpdateRate = 2f;
                    verySlowUpdateRate = 10f;
                    QualitySettings.SetQualityLevel(QualitySettings.GetQualityLevel() - 1, true);
                    break;

                case PerformanceLevel.Critical:
                    slowUpdateRate = 3f;
                    verySlowUpdateRate = 15f;
                    QualitySettings.SetQualityLevel(0, true);
                    break;
            }
        }

        // 登録・解除メソッド
        public void RegisterUpdatable(IUpdatable updatable, int priority = 0)
        {
            if (priority == 0)
            {
                updatables.Add(updatable);
            }
            else
            {
                if (!prioritizedUpdatables.ContainsKey(priority))
                {
                    prioritizedUpdatables[priority] = new List<IUpdatable>();
                }
                prioritizedUpdatables[priority].Add(updatable);
            }
        }

        public void UnregisterUpdatable(IUpdatable updatable)
        {
            updatables.Remove(updatable);

            foreach (var list in prioritizedUpdatables.Values)
            {
                list.Remove(updatable);
            }
        }

        public void RegisterFixedUpdatable(IFixedUpdatable fixedUpdatable)
        {
            fixedUpdatables.Add(fixedUpdatable);
        }

        public void UnregisterFixedUpdatable(IFixedUpdatable fixedUpdatable)
        {
            fixedUpdatables.Remove(fixedUpdatable);
        }

        public void RegisterSlowUpdatable(ISlowUpdatable slowUpdatable)
        {
            slowUpdatables.Add(slowUpdatable);
        }

        public void UnregisterSlowUpdatable(ISlowUpdatable slowUpdatable)
        {
            slowUpdatables.Remove(slowUpdatable);
        }

        public void RegisterVerySlowUpdatable(IVerySlowUpdatable verySlowUpdatable)
        {
            verySlowUpdatables.Add(verySlowUpdatable);
        }

        public void UnregisterVerySlowUpdatable(IVerySlowUpdatable verySlowUpdatable)
        {
            verySlowUpdatables.Remove(verySlowUpdatable);
        }

        // オブジェクトプール管理
        /// <summary>
        /// オブジェクトプールを作成
        /// </summary>
        public void CreatePool<T>(GameObject prefab, int initialSize = 10, int maxSize = 100) where T : Component
        {
            Type type = typeof(T);
            if (!objectPools.ContainsKey(type))
            {
                var pool = new ObjectPool(prefab, initialSize, maxSize, transform);
                objectPools[type] = pool;
            }
        }

        /// <summary>
        /// プールからオブジェクトを取得
        /// </summary>
        public T GetFromPool<T>() where T : Component
        {
            Type type = typeof(T);
            if (objectPools.TryGetValue(type, out ObjectPool pool))
            {
                GameObject obj = pool.Get();
                return obj?.GetComponent<T>();
            }

            ErrorHandler.LogWarning($"Pool not found for type: {type}", "UpdateManager");
            return null;
        }

        /// <summary>
        /// オブジェクトをプールに返却
        /// </summary>
        public void ReturnToPool<T>(T component) where T : Component
        {
            if (component == null)
                return;

            Type type = typeof(T);
            if (objectPools.TryGetValue(type, out ObjectPool pool))
            {
                pool.Return(component.gameObject);
            }
        }

        /// <summary>
        /// すべてのプールをクリア
        /// </summary>
        public void ClearAllPools()
        {
            foreach (var pool in objectPools.Values)
            {
                pool.Clear();
            }
        }

        /// <summary>
        /// 統計情報を取得
        /// </summary>
        public UpdateStats GetStats()
        {
            return new UpdateStats
            {
                fps = currentFPS,
                performanceLevel = currentPerformanceLevel,
                normalUpdateCount = updatables.Count + GetPrioritizedCount(),
                fixedUpdateCount = fixedUpdatables.Count,
                slowUpdateCount = slowUpdatables.Count,
                verySlowUpdateCount = verySlowUpdatables.Count,
                poolCount = objectPools.Count,
            };
        }

        private int GetPrioritizedCount()
        {
            int count = 0;
            foreach (var list in prioritizedUpdatables.Values)
            {
                count += list.Count;
            }
            return count;
        }

        /// <summary>
        /// オブジェクトプール
        /// </summary>
        private class ObjectPool
        {
            private GameObject prefab;
            private Queue<GameObject> pool;
            private Transform parent;
            private int maxSize;

            public ObjectPool(GameObject prefab, int initialSize, int maxSize, Transform parent)
            {
                this.prefab = prefab;
                this.maxSize = maxSize;
                this.parent = parent;
                pool = new Queue<GameObject>(initialSize);

                // 初期オブジェクトを生成
                for (int i = 0; i < initialSize; i++)
                {
                    CreateObject();
                }
            }

            private GameObject CreateObject()
            {
                GameObject obj = Instantiate(prefab, parent);
                obj.SetActive(false);
                pool.Enqueue(obj);
                return obj;
            }

            public GameObject Get()
            {
                GameObject obj;
                if (pool.Count > 0)
                {
                    obj = pool.Dequeue();
                }
                else
                {
                    obj = CreateObject();
                }

                obj.SetActive(true);
                return obj;
            }

            public void Return(GameObject obj)
            {
                if (obj == null)
                    return;

                obj.SetActive(false);
                obj.transform.SetParent(parent);

                if (pool.Count < maxSize)
                {
                    pool.Enqueue(obj);
                }
                else
                {
                    Destroy(obj);
                }
            }

            public void Clear()
            {
                while (pool.Count > 0)
                {
                    GameObject obj = pool.Dequeue();
                    if (obj != null)
                    {
                        Destroy(obj);
                    }
                }
            }
        }
    }

    // インターフェース定義
    public interface IUpdatable
    {
        bool IsActive { get; }
        void OnUpdate(float deltaTime);
    }

    public interface IFixedUpdatable
    {
        bool IsActive { get; }
        void OnFixedUpdate(float fixedDeltaTime);
    }

    public interface ISlowUpdatable
    {
        bool IsActive { get; }
        void OnSlowUpdate();
    }

    public interface IVerySlowUpdatable
    {
        bool IsActive { get; }
        void OnVerySlowUpdate();
    }

    // 列挙型
    public enum PerformanceLevel
    {
        Normal,
        Low,
        Critical,
    }

    // 統計情報
    [Serializable]
    public struct UpdateStats
    {
        public float fps;
        public PerformanceLevel performanceLevel;
        public int normalUpdateCount;
        public int fixedUpdateCount;
        public int slowUpdateCount;
        public int verySlowUpdateCount;
        public int poolCount;
    }

    // イベント
    public class PerformanceLevelChangedEvent : BaseGameEvent
    {
        public PerformanceLevel NewLevel { get; }

        public PerformanceLevelChangedEvent(PerformanceLevel newLevel)
        {
            NewLevel = newLevel;
        }
    }
}