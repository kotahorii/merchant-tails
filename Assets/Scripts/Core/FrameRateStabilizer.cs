using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.Rendering;

namespace MerchantTails.Core
{
    /// <summary>
    /// フレームレートを安定化させるシステム
    /// 動的な品質調整と負荷分散を行う
    /// </summary>
    public class FrameRateStabilizer : MonoBehaviour
    {
        private static FrameRateStabilizer instance;
        public static FrameRateStabilizer Instance => instance;

        [Header("Target Settings")]
        [SerializeField]
        private int targetFrameRate = 60;

        [SerializeField]
        private int minimumAcceptableFrameRate = 30;

        [SerializeField]
        private float frameRateTolerance = 5f; // ±5 FPS

        [Header("Quality Adjustment")]
        [SerializeField]
        private bool enableDynamicQuality = true;

        [SerializeField]
        private float qualityAdjustmentSpeed = 0.1f;

        [SerializeField]
        private float qualityCheckInterval = 2f;

        [Header("Load Balancing")]
        [SerializeField]
        private int maxOperationsPerFrame = 10;

        [SerializeField]
        private float maxFrameTime = 16.67f; // 60 FPS target

        [SerializeField]
        private bool enableTimeSlicing = true;

        [Header("Rendering Optimization")]
        [SerializeField]
        private bool enableLODOptimization = true;

        [SerializeField]
        private bool enableDynamicBatching = true;

        [SerializeField]
        private bool enableOcclusionCulling = true;

        // フレームレート計測
        private float currentFrameRate;
        private float averageFrameRate;
        private Queue<float> frameRateHistory = new Queue<float>();
        private const int frameRateHistorySize = 60;

        // 品質調整
        private float currentQualityLevel = 1f;
        private int lastQualityIndex;
        private float lastQualityCheckTime;

        // 負荷分散
        private Queue<Action> pendingOperations = new Queue<Action>();
        private float frameStartTime;

        // レンダリング設定
        private Dictionary<Camera, float> originalLODBias = new Dictionary<Camera, float>();
        private Dictionary<Light, LightShadows> originalShadowSettings = new Dictionary<Light, LightShadows>();

        public float CurrentFrameRate => currentFrameRate;
        public float AverageFrameRate => averageFrameRate;
        public float CurrentQualityLevel => currentQualityLevel;
        public bool IsStable => Mathf.Abs(averageFrameRate - targetFrameRate) <= frameRateTolerance;

        // イベント
        public event Action<float> OnFrameRateChanged;
        public event Action<float> OnQualityLevelChanged;
        public event Action OnFrameRateStabilized;
        public event Action OnFrameRateUnstable;

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

            RestoreOriginalSettings();
        }

        private void Initialize()
        {
            // フレームレート設定
            Application.targetFrameRate = targetFrameRate;
            QualitySettings.vSyncCount = 0;

            // 初期品質設定
            lastQualityIndex = QualitySettings.GetQualityLevel();

            // レンダリング最適化
            ApplyRenderingOptimizations();

            // 定期的な品質調整を開始
            StartCoroutine(QualityAdjustmentCoroutine());
        }

        private void Update()
        {
            frameStartTime = Time.realtimeSinceStartup;

            // フレームレート計測
            MeasureFrameRate();

            // 負荷分散処理
            ProcessPendingOperations();
        }

        /// <summary>
        /// フレームレートを計測
        /// </summary>
        private void MeasureFrameRate()
        {
            currentFrameRate = 1f / Time.deltaTime;

            // 履歴に追加
            frameRateHistory.Enqueue(currentFrameRate);
            if (frameRateHistory.Count > frameRateHistorySize)
            {
                frameRateHistory.Dequeue();
            }

            // 平均を計算
            float sum = 0f;
            foreach (float fps in frameRateHistory)
            {
                sum += fps;
            }
            averageFrameRate = sum / frameRateHistory.Count;

            OnFrameRateChanged?.Invoke(averageFrameRate);
        }

        /// <summary>
        /// 品質調整コルーチン
        /// </summary>
        private IEnumerator QualityAdjustmentCoroutine()
        {
            bool wasStable = true;

            while (true)
            {
                yield return new WaitForSeconds(qualityCheckInterval);

                if (!enableDynamicQuality)
                    continue;

                bool isStable = IsStable;

                // 安定性の変化を通知
                if (isStable != wasStable)
                {
                    if (isStable)
                    {
                        OnFrameRateStabilized?.Invoke();
                        EventBus.Publish(new FrameRateStabilizedEvent(averageFrameRate));
                    }
                    else
                    {
                        OnFrameRateUnstable?.Invoke();
                        EventBus.Publish(new FrameRateUnstableEvent(averageFrameRate));
                    }
                    wasStable = isStable;
                }

                // 品質調整
                AdjustQuality();
            }
        }

        /// <summary>
        /// 品質を調整
        /// </summary>
        private void AdjustQuality()
        {
            float frameRateDiff = averageFrameRate - targetFrameRate;
            float adjustment = 0f;

            if (frameRateDiff < -frameRateTolerance)
            {
                // フレームレートが低い - 品質を下げる
                adjustment = -qualityAdjustmentSpeed;
            }
            else if (frameRateDiff > frameRateTolerance && currentQualityLevel < 1f)
            {
                // フレームレートに余裕がある - 品質を上げる
                adjustment = qualityAdjustmentSpeed;
            }

            if (adjustment != 0f)
            {
                SetQualityLevel(currentQualityLevel + adjustment);
            }

            // 極端にフレームレートが低い場合
            if (averageFrameRate < minimumAcceptableFrameRate)
            {
                EmergencyQualityReduction();
            }
        }

        /// <summary>
        /// 品質レベルを設定
        /// </summary>
        private void SetQualityLevel(float level)
        {
            currentQualityLevel = Mathf.Clamp01(level);

            // Unity品質設定の調整
            int qualityIndex = Mathf.RoundToInt(currentQualityLevel * (QualitySettings.names.Length - 1));
            if (qualityIndex != QualitySettings.GetQualityLevel())
            {
                QualitySettings.SetQualityLevel(qualityIndex, true);
            }

            // カスタム品質調整
            ApplyCustomQualitySettings();

            OnQualityLevelChanged?.Invoke(currentQualityLevel);
            ErrorHandler.LogInfo($"Quality level adjusted to: {currentQualityLevel:F2}", "FrameRateStabilizer");
        }

        /// <summary>
        /// カスタム品質設定を適用
        /// </summary>
        private void ApplyCustomQualitySettings()
        {
            // LODバイアス
            if (enableLODOptimization)
            {
                float lodBias = Mathf.Lerp(2f, 0f, currentQualityLevel);
                QualitySettings.lodBias = lodBias;
            }

            // シャドウ品質
            if (currentQualityLevel < 0.3f)
            {
                QualitySettings.shadows = ShadowQuality.Disable;
            }
            else if (currentQualityLevel < 0.6f)
            {
                QualitySettings.shadows = ShadowQuality.HardOnly;
            }
            else
            {
                QualitySettings.shadows = ShadowQuality.All;
            }

            // パーティクル品質
            float particleRaycastBudget = Mathf.Lerp(64, 512, currentQualityLevel);
            QualitySettings.particleRaycastBudget = Mathf.RoundToInt(particleRaycastBudget);

            // テクスチャ品質
            if (currentQualityLevel < 0.4f)
            {
                QualitySettings.globalTextureMipmapLimit = 2;
            }
            else if (currentQualityLevel < 0.7f)
            {
                QualitySettings.globalTextureMipmapLimit = 1;
            }
            else
            {
                QualitySettings.globalTextureMipmapLimit = 0;
            }
        }

        /// <summary>
        /// 緊急品質低下
        /// </summary>
        private void EmergencyQualityReduction()
        {
            ErrorHandler.LogWarning("Emergency quality reduction triggered", "FrameRateStabilizer");

            // 最低品質に設定
            SetQualityLevel(0f);

            // 追加の最適化
            QualitySettings.shadows = ShadowQuality.Disable;
            QualitySettings.globalTextureMipmapLimit = 3;
            QualitySettings.lodBias = 3f;
            QualitySettings.particleRaycastBudget = 32;

            // すべてのライトのシャドウを無効化
            foreach (Light light in FindObjectsOfType<Light>())
            {
                if (!originalShadowSettings.ContainsKey(light))
                {
                    originalShadowSettings[light] = light.shadows;
                }
                light.shadows = LightShadows.None;
            }
        }

        /// <summary>
        /// レンダリング最適化を適用
        /// </summary>
        private void ApplyRenderingOptimizations()
        {
            // 動的バッチング
            if (enableDynamicBatching)
            {
                QualitySettings.realtimeReflectionProbes = false;
            }

            // オクルージョンカリング
            foreach (Camera cam in Camera.allCameras)
            {
                cam.useOcclusionCulling = enableOcclusionCulling;
            }
        }

        /// <summary>
        /// 元の設定を復元
        /// </summary>
        private void RestoreOriginalSettings()
        {
            // 品質設定を復元
            QualitySettings.SetQualityLevel(lastQualityIndex, true);

            // ライトのシャドウ設定を復元
            foreach (var kvp in originalShadowSettings)
            {
                if (kvp.Key != null)
                {
                    kvp.Key.shadows = kvp.Value;
                }
            }
        }

        // 負荷分散
        /// <summary>
        /// 時間分割処理を登録
        /// </summary>
        public void RegisterTimeSlicedOperation(Action operation)
        {
            if (!enableTimeSlicing || operation == null)
            {
                operation?.Invoke();
                return;
            }

            pendingOperations.Enqueue(operation);
        }

        /// <summary>
        /// 待機中の処理を実行
        /// </summary>
        private void ProcessPendingOperations()
        {
            if (!enableTimeSlicing || pendingOperations.Count == 0)
                return;

            int operationsProcessed = 0;
            float currentFrameTime = Time.realtimeSinceStartup - frameStartTime;

            while (pendingOperations.Count > 0 &&
                   operationsProcessed < maxOperationsPerFrame &&
                   currentFrameTime < maxFrameTime * 0.8f) // フレーム時間の80%まで
            {
                Action operation = pendingOperations.Dequeue();
                operation?.Invoke();
                operationsProcessed++;

                currentFrameTime = Time.realtimeSinceStartup - frameStartTime;
            }
        }

        /// <summary>
        /// コルーチンの時間分割実行
        /// </summary>
        public IEnumerator TimeSlicedCoroutine(IEnumerator coroutine, float maxTimePerFrame = 0.008f)
        {
            while (coroutine.MoveNext())
            {
                float startTime = Time.realtimeSinceStartup;

                yield return coroutine.Current;

                // フレーム時間チェック
                if (Time.realtimeSinceStartup - startTime > maxTimePerFrame)
                {
                    yield return null; // 次のフレームまで待機
                }
            }
        }

        /// <summary>
        /// パフォーマンス統計を取得
        /// </summary>
        public PerformanceStats GetPerformanceStats()
        {
            return new PerformanceStats
            {
                currentFPS = currentFrameRate,
                averageFPS = averageFrameRate,
                targetFPS = targetFrameRate,
                qualityLevel = currentQualityLevel,
                isStable = IsStable,
                pendingOperations = pendingOperations.Count,
                qualityIndex = QualitySettings.GetQualityLevel(),
            };
        }

        /// <summary>
        /// デバッグ情報を表示
        /// </summary>
        private void OnGUI()
        {
            if (!Debug.isDebugBuild)
                return;

            GUI.color = IsStable ? Color.green : Color.red;
            GUI.Label(new Rect(10, 10, 200, 20), $"FPS: {averageFrameRate:F1} / {targetFrameRate}");
            GUI.Label(new Rect(10, 30, 200, 20), $"Quality: {currentQualityLevel:F2}");
            GUI.Label(new Rect(10, 50, 200, 20), $"Pending Ops: {pendingOperations.Count}");
        }
    }

    // 構造体
    [Serializable]
    public struct PerformanceStats
    {
        public float currentFPS;
        public float averageFPS;
        public int targetFPS;
        public float qualityLevel;
        public bool isStable;
        public int pendingOperations;
        public int qualityIndex;
    }

    // イベント
    public class FrameRateStabilizedEvent : BaseGameEvent
    {
        public float FrameRate { get; }

        public FrameRateStabilizedEvent(float frameRate)
        {
            FrameRate = frameRate;
        }
    }

    public class FrameRateUnstableEvent : BaseGameEvent
    {
        public float FrameRate { get; }

        public FrameRateUnstableEvent(float frameRate)
        {
            FrameRate = frameRate;
        }
    }
}
