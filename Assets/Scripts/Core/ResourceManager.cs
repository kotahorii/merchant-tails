using System;
using System.Collections.Generic;
using UnityEngine;
// TODO: Addressables support - currently disabled
// using UnityEngine.AddressableAssets;
// using UnityEngine.ResourceManagement.AsyncOperations;

namespace MerchantTails.Core
{
    /// <summary>
    /// リソースの読み込みとキャッシュを管理するシステム
    /// メモリ効率を考慮した動的なリソース管理
    /// </summary>
    public class ResourceManager : MonoBehaviour
    {
        private static ResourceManager instance;
        public static ResourceManager Instance => instance;

        [Header("Cache Settings")]
        [SerializeField]
        private int maxCacheSize = 100; // 最大キャッシュ数

        [SerializeField]
        private float cacheExpirationTime = 300f; // キャッシュ有効期限（秒）

        [SerializeField]
        private bool useAddressables = false; // Addressables使用フラグ

        [Header("Memory Management")]
        [SerializeField]
        private float memoryCheckInterval = 30f; // メモリチェック間隔

        [SerializeField]
        private float memoryWarningThreshold = 0.8f; // メモリ警告閾値（0-1）

        [SerializeField]
        private float memoryCriticalThreshold = 0.9f; // メモリ危険閾値（0-1）

        // キャッシュ管理
        private Dictionary<string, CachedResource> resourceCache = new Dictionary<string, CachedResource>();
        private Queue<string> cacheQueue = new Queue<string>(); // LRU実装用
        private float lastMemoryCheckTime;

        // 非同期ロード管理
        // TODO: Addressables support
        // private Dictionary<string, AsyncOperationHandle> activeHandles = new Dictionary<string, AsyncOperationHandle>();
        private Dictionary<string, List<Action<UnityEngine.Object>>> pendingCallbacks =
            new Dictionary<string, List<Action<UnityEngine.Object>>>();

        // リソースパス定義
        public static class ResourcePaths
        {
            public const string Sprites = "Sprites/";
            public const string Prefabs = "Prefabs/";
            public const string Audio = "Audio/";
            public const string Materials = "Materials/";
            public const string UI = "UI/";
            public const string Data = "Data/";
            public const string Effects = "Effects/";
        }

        private void Awake()
        {
            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;
            DontDestroyOnLoad(gameObject);
        }

        private void OnDestroy()
        {
            if (instance == this)
            {
                instance = null;
            }

            // リソースのクリーンアップ
            ClearAllCache();
            ReleaseAllHandles();
        }

        private void Update()
        {
            // 定期的なメモリチェック
            if (Time.time - lastMemoryCheckTime > memoryCheckInterval)
            {
                CheckMemoryUsage();
                lastMemoryCheckTime = Time.time;
            }

            // キャッシュの有効期限チェック
            CheckCacheExpiration();
        }

        /// <summary>
        /// リソースを同期的に読み込み
        /// </summary>
        public T Load<T>(string path)
            where T : UnityEngine.Object
        {
            // キャッシュチェック
            if (TryGetFromCache<T>(path, out T cached))
            {
                return cached;
            }

            // リソース読み込み
            T resource = Resources.Load<T>(path);
            if (resource != null)
            {
                AddToCache(path, resource);
            }
            else
            {
                ErrorHandler.LogWarning($"Resource not found: {path}", null, "ResourceManager");
            }

            return resource;
        }

        /// <summary>
        /// リソースを非同期で読み込み
        /// </summary>
        public void LoadAsync<T>(string path, Action<T> onComplete) where T : UnityEngine.Object
        {
            // キャッシュチェック
            if (TryGetFromCache<T>(path, out T cached))
            {
                onComplete?.Invoke(cached);
                return;
            }

            // 既にロード中の場合はコールバックを追加
            if (pendingCallbacks.ContainsKey(path))
            {
                pendingCallbacks[path].Add(obj => onComplete?.Invoke(obj as T));
                return;
            }

            // 新規非同期ロード
            pendingCallbacks[path] = new List<Action<UnityEngine.Object>> { obj => onComplete?.Invoke(obj as T) };

            // TODO: Addressables support
            // if (useAddressables)
            // {
            //     LoadWithAddressables<T>(path);
            // }
            // else
            // {
                StartCoroutine(LoadWithResourcesAsync<T>(path));
            // }
        }

        /// <summary>
        /// Resources.LoadAsyncを使用した非同期読み込み
        /// </summary>
        private System.Collections.IEnumerator LoadWithResourcesAsync<T>(string path) where T : UnityEngine.Object
        {
            ResourceRequest request = Resources.LoadAsync<T>(path);
            yield return request;

            if (request.asset != null)
            {
                AddToCache(path, request.asset);
                InvokeCallbacks(path, request.asset);
            }
            else
            {
                ErrorHandler.LogWarning($"Async resource not found: {path}", "ResourceManager");
                InvokeCallbacks(path, null);
            }

            pendingCallbacks.Remove(path);
        }

        // TODO: Addressables support
        // /// <summary>
        // /// Addressablesを使用した非同期読み込み
        // /// </summary>
        // private void LoadWithAddressables<T>(string path) where T : UnityEngine.Object
        // {
        //     var handle = Addressables.LoadAssetAsync<T>(path);
        //     activeHandles[path] = handle;
        //
        //     handle.Completed += (AsyncOperationHandle<T> completedHandle) =>
        //     {
        //         if (completedHandle.Status == AsyncOperationStatus.Succeeded)
        //         {
        //             AddToCache(path, completedHandle.Result);
        //             InvokeCallbacks(path, completedHandle.Result);
        //         }
        //         else
        //         {
        //             ErrorHandler.LogError($"Addressable load failed: {path}", "ResourceManager");
        //             InvokeCallbacks(path, null);
        //         }
        //
        //         activeHandles.Remove(path);
        //         pendingCallbacks.Remove(path);
        //     };
        // }

        /// <summary>
        /// プリロード（事前読み込み）
        /// </summary>
        public void Preload<T>(string[] paths) where T : UnityEngine.Object
        {
            foreach (string path in paths)
            {
                LoadAsync<T>(path, _ => { }); // キャッシュのみ
            }
        }

        /// <summary>
        /// キャッシュから取得を試みる
        /// </summary>
        private bool TryGetFromCache<T>(string path, out T resource) where T : UnityEngine.Object
        {
            if (resourceCache.TryGetValue(path, out CachedResource cached))
            {
                cached.lastAccessTime = Time.time;
                resource = cached.resource as T;

                // LRU更新
                UpdateLRU(path);

                return resource != null;
            }

            resource = null;
            return false;
        }

        /// <summary>
        /// キャッシュに追加
        /// </summary>
        private void AddToCache(string path, UnityEngine.Object resource)
        {
            // キャッシュサイズチェック
            if (resourceCache.Count >= maxCacheSize)
            {
                RemoveOldestCache();
            }

            var cached = new CachedResource
            {
                resource = resource,
                loadTime = Time.time,
                lastAccessTime = Time.time,
                referenceCount = 1,
            };

            resourceCache[path] = cached;
            cacheQueue.Enqueue(path);

            ErrorHandler.LogInfo($"Cached resource: {path}", "ResourceManager");
        }

        /// <summary>
        /// LRU順序を更新
        /// </summary>
        private void UpdateLRU(string path)
        {
            // キューから削除して最後に追加
            var tempQueue = new Queue<string>();
            while (cacheQueue.Count > 0)
            {
                string item = cacheQueue.Dequeue();
                if (item != path)
                {
                    tempQueue.Enqueue(item);
                }
            }
            cacheQueue = tempQueue;
            cacheQueue.Enqueue(path);
        }

        /// <summary>
        /// 最も古いキャッシュを削除
        /// </summary>
        private void RemoveOldestCache()
        {
            if (cacheQueue.Count > 0)
            {
                string oldestPath = cacheQueue.Dequeue();
                RemoveFromCache(oldestPath);
            }
        }

        /// <summary>
        /// キャッシュから削除
        /// </summary>
        public void RemoveFromCache(string path)
        {
            if (resourceCache.TryGetValue(path, out CachedResource cached))
            {
                // Addressablesの場合はリリース
                // TODO: Addressables support
                // if (useAddressables && activeHandles.ContainsKey(path))
                // {
                //     Addressables.Release(activeHandles[path]);
                //     activeHandles.Remove(path);
                // }

                resourceCache.Remove(path);
                ErrorHandler.LogInfo($"Removed from cache: {path}", "ResourceManager");
            }
        }

        /// <summary>
        /// キャッシュの有効期限チェック
        /// </summary>
        private void CheckCacheExpiration()
        {
            List<string> expiredPaths = new List<string>();

            foreach (var kvp in resourceCache)
            {
                if (Time.time - kvp.Value.lastAccessTime > cacheExpirationTime)
                {
                    expiredPaths.Add(kvp.Key);
                }
            }

            foreach (string path in expiredPaths)
            {
                RemoveFromCache(path);
            }
        }

        /// <summary>
        /// メモリ使用状況をチェック
        /// </summary>
        private void CheckMemoryUsage()
        {
            float memoryUsage = GetMemoryUsage();

            if (memoryUsage > memoryCriticalThreshold)
            {
                ErrorHandler.LogError($"Critical memory usage: {memoryUsage:P0}", null, "ResourceManager");
                EmergencyCacheClear();
            }
            else if (memoryUsage > memoryWarningThreshold)
            {
                ErrorHandler.LogWarning($"High memory usage: {memoryUsage:P0}", "ResourceManager");
                ClearUnusedCache();
            }
        }

        /// <summary>
        /// メモリ使用率を取得
        /// </summary>
        private float GetMemoryUsage()
        {
            long usedMemory = GC.GetTotalMemory(false);
            long totalMemory = SystemInfo.systemMemorySize * 1024L * 1024L; // MB to bytes
            return (float)usedMemory / totalMemory;
        }

        /// <summary>
        /// 未使用のキャッシュをクリア
        /// </summary>
        private void ClearUnusedCache()
        {
            List<string> unusedPaths = new List<string>();
            float currentTime = Time.time;

            foreach (var kvp in resourceCache)
            {
                // 60秒以上アクセスされていないリソースを削除
                if (currentTime - kvp.Value.lastAccessTime > 60f)
                {
                    unusedPaths.Add(kvp.Key);
                }
            }

            foreach (string path in unusedPaths)
            {
                RemoveFromCache(path);
            }

            // ガベージコレクション
            Resources.UnloadUnusedAssets();
            GC.Collect();
        }

        /// <summary>
        /// 緊急時のキャッシュクリア
        /// </summary>
        private void EmergencyCacheClear()
        {
            ClearAllCache();
            Resources.UnloadUnusedAssets();
            GC.Collect();
            ErrorHandler.LogWarning("Emergency cache clear executed", "ResourceManager");
        }

        /// <summary>
        /// すべてのキャッシュをクリア
        /// </summary>
        public void ClearAllCache()
        {
            foreach (string path in resourceCache.Keys)
            {
                // TODO: Addressables support
                // if (useAddressables && activeHandles.ContainsKey(path))
                // {
                //     Addressables.Release(activeHandles[path]);
                // }
            }

            resourceCache.Clear();
            cacheQueue.Clear();
            // TODO: Addressables support
            // activeHandles.Clear();
            pendingCallbacks.Clear();
        }

        /// <summary>
        /// コールバックを実行
        /// </summary>
        private void InvokeCallbacks(string path, UnityEngine.Object resource)
        {
            if (pendingCallbacks.TryGetValue(path, out List<Action<UnityEngine.Object>> callbacks))
            {
                foreach (var callback in callbacks)
                {
                    callback?.Invoke(resource);
                }
            }
        }

        /// <summary>
        /// すべてのハンドルをリリース
        /// </summary>
        private void ReleaseAllHandles()
        {
            // TODO: Addressables support
            // if (useAddressables)
            // {
            //     foreach (var handle in activeHandles.Values)
            //     {
            //         Addressables.Release(handle);
            //     }
            //     activeHandles.Clear();
            // }
        }

        /// <summary>
        /// スプライトを読み込み
        /// </summary>
        public Sprite LoadSprite(string spriteName)
        {
            return Load<Sprite>($"{ResourcePaths.Sprites}{spriteName}");
        }

        /// <summary>
        /// プレハブを読み込み
        /// </summary>
        public GameObject LoadPrefab(string prefabName)
        {
            return Load<GameObject>($"{ResourcePaths.Prefabs}{prefabName}");
        }

        /// <summary>
        /// オーディオクリップを読み込み
        /// </summary>
        public AudioClip LoadAudioClip(string clipName)
        {
            return Load<AudioClip>($"{ResourcePaths.Audio}{clipName}");
        }

        /// <summary>
        /// UIプレハブを読み込み
        /// </summary>
        public GameObject LoadUIPrefab(string uiName)
        {
            return Load<GameObject>($"{ResourcePaths.UI}{uiName}");
        }

        /// <summary>
        /// ScriptableObjectを読み込み
        /// </summary>
        public T LoadScriptableObject<T>(string name) where T : ScriptableObject
        {
            return Load<T>($"{ResourcePaths.Data}{name}");
        }

        /// <summary>
        /// キャッシュ情報を取得
        /// </summary>
        public CacheInfo GetCacheInfo()
        {
            return new CacheInfo
            {
                totalCached = resourceCache.Count,
                maxCacheSize = maxCacheSize,
                memoryUsage = GetMemoryUsage(),
                oldestCacheAge = GetOldestCacheAge(),
            };
        }

        /// <summary>
        /// 最も古いキャッシュの経過時間を取得
        /// </summary>
        private float GetOldestCacheAge()
        {
            float currentTime = Time.time;
            float oldestAge = 0f;

            foreach (var cached in resourceCache.Values)
            {
                float age = currentTime - cached.loadTime;
                if (age > oldestAge)
                {
                    oldestAge = age;
                }
            }

            return oldestAge;
        }

        /// <summary>
        /// キャッシュされたリソース
        /// </summary>
        private class CachedResource
        {
            public UnityEngine.Object resource;
            public float loadTime;
            public float lastAccessTime;
            public int referenceCount;
        }

        /// <summary>
        /// キャッシュ情報
        /// </summary>
        public struct CacheInfo
        {
            public int totalCached;
            public int maxCacheSize;
            public float memoryUsage;
            public float oldestCacheAge;
        }
    }
}
