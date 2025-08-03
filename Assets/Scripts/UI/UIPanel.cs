using MerchantTails.Core;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.UI
{
    /// <summary>
    /// UI画面の基底クラス
    /// 各画面の共通機能を提供
    /// </summary>
    public class UIPanel : MonoBehaviour
    {
        [Header("Panel Settings")]
        [SerializeField]
        private bool hideOnStart = true;

        [SerializeField]
        private bool useAnimations = true;

        [SerializeField]
        private float animationDuration = 0.3f;

        [Header("Navigation")]
        [SerializeField]
        private Button backButton;

        [SerializeField]
        private Button closeButton;

        private CanvasGroup canvasGroup;
        private RectTransform rectTransform;
        private System.Action<bool> modalCallback;

        public UIType UIType { get; private set; }
        public bool IsVisible { get; private set; }

        protected virtual void Awake()
        {
            canvasGroup = GetComponent<CanvasGroup>();
            if (canvasGroup == null)
            {
                canvasGroup = gameObject.AddComponent<CanvasGroup>();
            }

            rectTransform = GetComponent<RectTransform>();

            SetupButtons();
        }

        protected virtual void Start()
        {
            if (hideOnStart)
            {
                Hide(false);
            }
        }

        private void SetupButtons()
        {
            if (backButton != null)
            {
                backButton.onClick.AddListener(OnBackPressed);
            }

            if (closeButton != null)
            {
                closeButton.onClick.AddListener(OnClosePressed);
            }
        }

        public void Initialize(UIType uiType)
        {
            UIType = uiType;
            OnInitialize();
        }

        protected virtual void OnInitialize()
        {
            // サブクラスでオーバーライド
        }

        public virtual void Show(bool animated = true)
        {
            gameObject.SetActive(true);
            IsVisible = true;

            if (animated && useAnimations)
            {
                StartCoroutine(ShowAnimation());
            }
            else
            {
                SetVisibility(true);
            }

            OnShow();
        }

        public virtual void Hide(bool animated = true)
        {
            IsVisible = false;

            if (animated && useAnimations)
            {
                StartCoroutine(HideAnimation());
            }
            else
            {
                SetVisibility(false);
                gameObject.SetActive(false);
            }

            OnHide();
        }

        protected virtual void OnShow()
        {
            // サブクラスでオーバーライド
        }

        protected virtual void OnHide()
        {
            // サブクラスでオーバーライド
        }

        private System.Collections.IEnumerator ShowAnimation()
        {
            // フェードイン + スケールアニメーション
            float elapsed = 0f;
            Vector3 startScale = Vector3.zero;
            Vector3 endScale = Vector3.one;

            canvasGroup.alpha = 0f;
            rectTransform.localScale = startScale;

            while (elapsed < animationDuration)
            {
                elapsed += Time.deltaTime;
                float progress = elapsed / animationDuration;

                // イージング (EaseOutBack)
                float easedProgress = EaseOutBack(progress);

                canvasGroup.alpha = Mathf.Lerp(0f, 1f, progress);
                rectTransform.localScale = Vector3.Lerp(startScale, endScale, easedProgress);

                yield return null;
            }

            canvasGroup.alpha = 1f;
            rectTransform.localScale = endScale;
        }

        private System.Collections.IEnumerator HideAnimation()
        {
            // フェードアウト + スケールアニメーション
            float elapsed = 0f;
            Vector3 startScale = Vector3.one;
            Vector3 endScale = Vector3.zero;

            while (elapsed < animationDuration)
            {
                elapsed += Time.deltaTime;
                float progress = elapsed / animationDuration;

                // イージング (EaseInBack)
                float easedProgress = EaseInBack(progress);

                canvasGroup.alpha = Mathf.Lerp(1f, 0f, progress);
                rectTransform.localScale = Vector3.Lerp(startScale, endScale, easedProgress);

                yield return null;
            }

            canvasGroup.alpha = 0f;
            rectTransform.localScale = endScale;
            gameObject.SetActive(false);
        }

        private void SetVisibility(bool visible)
        {
            canvasGroup.alpha = visible ? 1f : 0f;
            canvasGroup.interactable = visible;
            canvasGroup.blocksRaycasts = visible;
            rectTransform.localScale = visible ? Vector3.one : Vector3.zero;
        }

        // イージング関数
        private float EaseOutBack(float t)
        {
            const float c1 = 1.70158f;
            const float c3 = c1 + 1f;

            return 1f + c3 * Mathf.Pow(t - 1f, 3f) + c1 * Mathf.Pow(t - 1f, 2f);
        }

        private float EaseInBack(float t)
        {
            const float c1 = 1.70158f;
            const float c3 = c1 + 1f;

            return c3 * t * t * t - c1 * t * t;
        }

        // ナビゲーション
        protected virtual void OnBackPressed()
        {
            UIManager.Instance?.GoBack();
        }

        protected virtual void OnClosePressed()
        {
            Hide();
        }

        // モーダル機能
        public void SetModalCallback(System.Action<bool> callback)
        {
            modalCallback = callback;
        }

        public void TriggerModalCallback(bool result)
        {
            modalCallback?.Invoke(result);
            modalCallback = null;
        }

        // 親の変更（モーダル用）
        public void SetParent(Transform newParent)
        {
            transform.SetParent(newParent, false);
        }

        // ユーティリティ
        public void SetInteractable(bool interactable)
        {
            canvasGroup.interactable = interactable;
        }

        public void SetAlpha(float alpha)
        {
            canvasGroup.alpha = alpha;
        }

        // UI要素の検索ヘルパー
        protected T FindUIComponent<T>(string name)
            where T : Component
        {
            Transform found = transform.Find(name);
            return found != null ? found.GetComponent<T>() : null;
        }

        protected T FindUIComponentInChildren<T>(string name)
            where T : Component
        {
            Transform found = transform.Find(name);
            if (found == null)
            {
                // 再帰的に検索
                found = FindChildRecursive(transform, name);
            }

            return found != null ? found.GetComponent<T>() : null;
        }

        private Transform FindChildRecursive(Transform parent, string name)
        {
            foreach (Transform child in parent)
            {
                if (child.name == name)
                {
                    return child;
                }

                Transform found = FindChildRecursive(child, name);
                if (found != null)
                {
                    return found;
                }
            }

            return null;
        }

        // デバッグ
        protected void LogUIAction(string action)
        {
            ErrorHandler.LogInfo($"[{UIType}] {action}", "UIPanel");
        }

        private void OnDestroy()
        {
            // ボタンのイベント解除
            if (backButton != null)
            {
                backButton.onClick.RemoveListener(OnBackPressed);
            }

            if (closeButton != null)
            {
                closeButton.onClick.RemoveListener(OnClosePressed);
            }
        }
    }
}
