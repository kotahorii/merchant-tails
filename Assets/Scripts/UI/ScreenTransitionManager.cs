using System;
using System.Collections;
using UnityEngine;
using UnityEngine.UI;
using MerchantTails.Core;

namespace MerchantTails.UI
{
    /// <summary>
    /// 画面遷移エフェクトを管理するクラス
    /// </summary>
    public class ScreenTransitionManager : MonoBehaviour
    {
        private static ScreenTransitionManager instance;
        public static ScreenTransitionManager Instance => instance;

        [Header("Transition Elements")]
        [SerializeField] private Canvas transitionCanvas;
        [SerializeField] private Image fadeImage;
        [SerializeField] private Image circleWipeImage;
        [SerializeField] private Image slideImage;
        [SerializeField] private Material circleWipeMaterial;

        [Header("Transition Settings")]
        [SerializeField] private float defaultTransitionDuration = 0.5f;
        [SerializeField] private AnimationCurve defaultTransitionCurve = AnimationCurve.EaseInOut(0, 0, 1, 1);
        [SerializeField] private Color fadeColor = Color.black;

        private bool isTransitioning = false;
        private Coroutine currentTransition;

        public bool IsTransitioning => isTransitioning;

        private void Awake()
        {
            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;
            DontDestroyOnLoad(gameObject);

            InitializeTransitionElements();
        }

        private void OnDestroy()
        {
            if (instance == this)
            {
                instance = null;
            }
        }

        private void InitializeTransitionElements()
        {
            // トランジションキャンバスの設定
            if (transitionCanvas != null)
            {
                transitionCanvas.renderMode = RenderMode.ScreenSpaceOverlay;
                transitionCanvas.sortingOrder = 9999; // 最前面に表示
            }

            // 初期状態では非表示
            HideAllTransitionElements();
        }

        private void HideAllTransitionElements()
        {
            if (fadeImage != null)
                fadeImage.gameObject.SetActive(false);
            if (circleWipeImage != null)
                circleWipeImage.gameObject.SetActive(false);
            if (slideImage != null)
                slideImage.gameObject.SetActive(false);
        }

        /// <summary>
        /// フェード遷移を実行
        /// </summary>
        public void DoFadeTransition(Action onTransitionMiddle = null, float duration = -1)
        {
            if (isTransitioning) return;

            if (currentTransition != null)
                StopCoroutine(currentTransition);

            currentTransition = StartCoroutine(
                FadeTransitionCoroutine(onTransitionMiddle, duration > 0 ? duration : defaultTransitionDuration)
            );
        }

        /// <summary>
        /// サークルワイプ遷移を実行
        /// </summary>
        public void DoCircleWipeTransition(
            Vector2 centerPosition,
            Action onTransitionMiddle = null,
            float duration = -1
        )
        {
            if (isTransitioning) return;

            if (currentTransition != null)
                StopCoroutine(currentTransition);

            currentTransition = StartCoroutine(
                CircleWipeTransitionCoroutine(
                    centerPosition,
                    onTransitionMiddle,
                    duration > 0 ? duration : defaultTransitionDuration
                )
            );
        }

        /// <summary>
        /// スライド遷移を実行
        /// </summary>
        public void DoSlideTransition(
            SlideDirection direction,
            Action onTransitionMiddle = null,
            float duration = -1
        )
        {
            if (isTransitioning) return;

            if (currentTransition != null)
                StopCoroutine(currentTransition);

            currentTransition = StartCoroutine(
                SlideTransitionCoroutine(direction, onTransitionMiddle, duration > 0 ? duration : defaultTransitionDuration)
            );
        }

        private IEnumerator FadeTransitionCoroutine(Action onTransitionMiddle, float duration)
        {
            isTransitioning = true;

            if (fadeImage == null)
                yield break;

            fadeImage.gameObject.SetActive(true);
            fadeImage.color = new Color(fadeColor.r, fadeColor.g, fadeColor.b, 0);

            // フェードアウト
            float halfDuration = duration * 0.5f;
            yield return StartCoroutine(FadeCoroutine(fadeImage, 0, 1, halfDuration));

            // 遷移中の処理
            onTransitionMiddle?.Invoke();

            // フェードイン
            yield return StartCoroutine(FadeCoroutine(fadeImage, 1, 0, halfDuration));

            fadeImage.gameObject.SetActive(false);
            isTransitioning = false;
        }

        private IEnumerator CircleWipeTransitionCoroutine(
            Vector2 centerPosition,
            Action onTransitionMiddle,
            float duration
        )
        {
            isTransitioning = true;

            if (circleWipeImage == null || circleWipeMaterial == null)
                yield break;

            circleWipeImage.gameObject.SetActive(true);
            circleWipeImage.material = new Material(circleWipeMaterial); // マテリアルのコピーを作成

            // 画面座標を正規化座標に変換
            Vector2 normalizedCenter = new Vector2(
                centerPosition.x / Screen.width,
                centerPosition.y / Screen.height
            );

            circleWipeImage.material.SetVector("_Center", normalizedCenter);

            // ワイプアウト
            float halfDuration = duration * 0.5f;
            yield return StartCoroutine(CircleWipeCoroutine(circleWipeImage.material, 0, 1, halfDuration));

            // 遷移中の処理
            onTransitionMiddle?.Invoke();

            // ワイプイン
            yield return StartCoroutine(CircleWipeCoroutine(circleWipeImage.material, 1, 0, halfDuration));

            Destroy(circleWipeImage.material); // コピーしたマテリアルを破棄
            circleWipeImage.gameObject.SetActive(false);
            isTransitioning = false;
        }

        private IEnumerator SlideTransitionCoroutine(
            SlideDirection direction,
            Action onTransitionMiddle,
            float duration
        )
        {
            isTransitioning = true;

            if (slideImage == null)
                yield break;

            slideImage.gameObject.SetActive(true);
            slideImage.color = fadeColor;

            RectTransform rect = slideImage.rectTransform;
            rect.anchorMin = Vector2.zero;
            rect.anchorMax = Vector2.one;
            rect.sizeDelta = Vector2.zero;

            // スライド方向に応じた初期位置と目標位置を設定
            Vector2 startPos, middlePos, endPos;
            switch (direction)
            {
                case SlideDirection.Left:
                    startPos = new Vector2(Screen.width, 0);
                    middlePos = Vector2.zero;
                    endPos = new Vector2(-Screen.width, 0);
                    break;
                case SlideDirection.Right:
                    startPos = new Vector2(-Screen.width, 0);
                    middlePos = Vector2.zero;
                    endPos = new Vector2(Screen.width, 0);
                    break;
                case SlideDirection.Up:
                    startPos = new Vector2(0, -Screen.height);
                    middlePos = Vector2.zero;
                    endPos = new Vector2(0, Screen.height);
                    break;
                case SlideDirection.Down:
                    startPos = new Vector2(0, Screen.height);
                    middlePos = Vector2.zero;
                    endPos = new Vector2(0, -Screen.height);
                    break;
                default:
                    startPos = middlePos = endPos = Vector2.zero;
                    break;
            }

            // スライドイン
            float halfDuration = duration * 0.5f;
            yield return StartCoroutine(SlideCoroutine(rect, startPos, middlePos, halfDuration));

            // 遷移中の処理
            onTransitionMiddle?.Invoke();

            // スライドアウト
            yield return StartCoroutine(SlideCoroutine(rect, middlePos, endPos, halfDuration));

            slideImage.gameObject.SetActive(false);
            isTransitioning = false;
        }

        private IEnumerator FadeCoroutine(Image image, float startAlpha, float endAlpha, float duration)
        {
            float elapsed = 0;
            Color color = image.color;

            while (elapsed < duration)
            {
                elapsed += Time.deltaTime;
                float t = defaultTransitionCurve.Evaluate(elapsed / duration);
                color.a = Mathf.Lerp(startAlpha, endAlpha, t);
                image.color = color;
                yield return null;
            }

            color.a = endAlpha;
            image.color = color;
        }

        private IEnumerator CircleWipeCoroutine(Material material, float startRadius, float endRadius, float duration)
        {
            float elapsed = 0;

            while (elapsed < duration)
            {
                elapsed += Time.deltaTime;
                float t = defaultTransitionCurve.Evaluate(elapsed / duration);
                float radius = Mathf.Lerp(startRadius, endRadius, t);
                material.SetFloat("_Radius", radius);
                yield return null;
            }

            material.SetFloat("_Radius", endRadius);
        }

        private IEnumerator SlideCoroutine(RectTransform rect, Vector2 startPos, Vector2 endPos, float duration)
        {
            float elapsed = 0;

            while (elapsed < duration)
            {
                elapsed += Time.deltaTime;
                float t = defaultTransitionCurve.Evaluate(elapsed / duration);
                rect.anchoredPosition = Vector2.Lerp(startPos, endPos, t);
                yield return null;
            }

            rect.anchoredPosition = endPos;
        }

        /// <summary>
        /// 即座にフェードアウト状態にする
        /// </summary>
        public void SetFadeOut()
        {
            if (fadeImage != null)
            {
                fadeImage.gameObject.SetActive(true);
                fadeImage.color = new Color(fadeColor.r, fadeColor.g, fadeColor.b, 1);
            }
        }

        /// <summary>
        /// 即座にフェードイン状態にする
        /// </summary>
        public void SetFadeIn()
        {
            if (fadeImage != null)
            {
                fadeImage.gameObject.SetActive(false);
                fadeImage.color = new Color(fadeColor.r, fadeColor.g, fadeColor.b, 0);
            }
        }

        public enum SlideDirection
        {
            Left,
            Right,
            Up,
            Down
        }
    }
}