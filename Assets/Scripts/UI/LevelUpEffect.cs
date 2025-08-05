using System.Collections;
using MerchantTails.Core;
using MerchantTails.Data;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.UI
{
    /// <summary>
    /// ランクアップ時の演出を管理
    /// </summary>
    public class LevelUpEffect : MonoBehaviour
    {
        [Header("UI References")]
        [SerializeField]
        private GameObject effectPanel;

        [SerializeField]
        private Image backgroundOverlay;

        [SerializeField]
        private RectTransform effectContainer;

        [SerializeField]
        private TextMeshProUGUI rankUpText;

        [SerializeField]
        private TextMeshProUGUI newRankText;

        [SerializeField]
        private TextMeshProUGUI previousRankText;

        [SerializeField]
        private Image rankIcon;

        [SerializeField]
        private GameObject unlockedFeaturesPanel;

        [SerializeField]
        private Transform featureListContainer;

        [SerializeField]
        private GameObject featureItemPrefab;

        [Header("Animation Settings")]
        [SerializeField]
        private float fadeInDuration = 0.5f;

        [SerializeField]
        private float displayDuration = 3f;

        [SerializeField]
        private float fadeOutDuration = 0.5f;

        [SerializeField]
        private AnimationCurve scaleCurve = AnimationCurve.EaseInOut(0, 0, 1, 1);

        [SerializeField]
        private AnimationCurve bounceCurve;

        [Header("Effects")]
        [SerializeField]
        private ParticleSystem sparkleEffect;

        [SerializeField]
        private ParticleSystem confettiEffect;

        [SerializeField]
        private AudioClip rankUpSound;

        [SerializeField]
        private AudioClip featureUnlockSound;

        [Header("Rank Icons")]
        [SerializeField]
        private Sprite apprenticeIcon;

        [SerializeField]
        private Sprite skilledIcon;

        [SerializeField]
        private Sprite veteranIcon;

        [SerializeField]
        private Sprite masterIcon;

        private AudioSource audioSource;
        private Coroutine currentEffectCoroutine;
        private FeatureUnlockSystem featureUnlockSystem;

        private void Awake()
        {
            audioSource = GetComponent<AudioSource>();
            if (audioSource == null)
            {
                audioSource = gameObject.AddComponent<AudioSource>();
            }

            // 初期状態で非表示
            if (effectPanel != null)
            {
                effectPanel.SetActive(false);
            }
        }

        private void Start()
        {
            featureUnlockSystem = FeatureUnlockSystem.Instance;
            SubscribeToEvents();
        }

        private void OnDestroy()
        {
            UnsubscribeFromEvents();
        }

        private void SubscribeToEvents()
        {
            EventBus.Subscribe<RankChangedEvent>(OnRankChanged);
        }

        private void UnsubscribeFromEvents()
        {
            EventBus.Unsubscribe<RankChangedEvent>(OnRankChanged);
        }

        private void OnRankChanged(RankChangedEvent e)
        {
            if (currentEffectCoroutine != null)
            {
                StopCoroutine(currentEffectCoroutine);
            }

            currentEffectCoroutine = StartCoroutine(PlayRankUpEffect(e.PreviousRank, e.NewRank));
        }

        /// <summary>
        /// ランクアップ演出を再生
        /// </summary>
        private IEnumerator PlayRankUpEffect(MerchantRank previousRank, MerchantRank newRank)
        {
            // UI要素の準備
            PrepareUI(previousRank, newRank);

            // パネルを表示
            effectPanel.SetActive(true);

            // フェードイン
            yield return StartCoroutine(FadeIn());

            // サウンド再生
            if (rankUpSound != null && audioSource != null)
            {
                audioSource.PlayOneShot(rankUpSound);
            }

            // パーティクル再生
            if (sparkleEffect != null)
            {
                sparkleEffect.Play();
            }

            if (confettiEffect != null)
            {
                confettiEffect.Play();
            }

            // スケールアニメーション
            yield return StartCoroutine(ScaleAnimation());

            // 新機能の表示
            var unlockedFeatures = GetUnlockedFeaturesForRank(newRank);
            if (unlockedFeatures.Count > 0)
            {
                yield return new WaitForSeconds(0.5f);
                yield return StartCoroutine(ShowUnlockedFeatures(unlockedFeatures));
            }

            // 表示時間待機
            yield return new WaitForSeconds(displayDuration);

            // フェードアウト
            yield return StartCoroutine(FadeOut());

            // パネルを非表示
            effectPanel.SetActive(false);
            currentEffectCoroutine = null;
        }

        /// <summary>
        /// UI要素を準備
        /// </summary>
        private void PrepareUI(MerchantRank previousRank, MerchantRank newRank)
        {
            // テキスト設定
            rankUpText.text = "ランクアップ！";
            previousRankText.text = GetRankDisplayName(previousRank);
            newRankText.text = GetRankDisplayName(newRank);

            // アイコン設定
            rankIcon.sprite = GetRankIcon(newRank);

            // 色設定
            Color rankColor = GetRankColor(newRank);
            newRankText.color = rankColor;
            rankIcon.color = rankColor;

            // 初期状態
            backgroundOverlay.color = new Color(0, 0, 0, 0);
            effectContainer.localScale = Vector3.zero;
            unlockedFeaturesPanel.SetActive(false);
        }

        /// <summary>
        /// フェードイン演出
        /// </summary>
        private IEnumerator FadeIn()
        {
            float elapsed = 0f;

            while (elapsed < fadeInDuration)
            {
                elapsed += Time.deltaTime;
                float t = elapsed / fadeInDuration;

                // 背景のフェード
                backgroundOverlay.color = new Color(0, 0, 0, Mathf.Lerp(0, 0.8f, t));

                // コンテナのスケール
                float scale = scaleCurve.Evaluate(t);
                effectContainer.localScale = Vector3.one * scale;

                yield return null;
            }

            backgroundOverlay.color = new Color(0, 0, 0, 0.8f);
            effectContainer.localScale = Vector3.one;
        }

        /// <summary>
        /// スケールアニメーション（バウンス効果）
        /// </summary>
        private IEnumerator ScaleAnimation()
        {
            if (bounceCurve == null || bounceCurve.length == 0)
            {
                yield break;
            }

            float duration = 1f;
            float elapsed = 0f;

            while (elapsed < duration)
            {
                elapsed += Time.deltaTime;
                float t = elapsed / duration;

                float scale = bounceCurve.Evaluate(t);
                effectContainer.localScale = Vector3.one * scale;

                yield return null;
            }

            effectContainer.localScale = Vector3.one;
        }

        /// <summary>
        /// 解放された機能を表示
        /// </summary>
        private IEnumerator ShowUnlockedFeatures(List<GameFeature> features)
        {
            // 既存のアイテムをクリア
            foreach (Transform child in featureListContainer)
            {
                Destroy(child.gameObject);
            }

            unlockedFeaturesPanel.SetActive(true);

            // フェードイン
            CanvasGroup panelGroup = unlockedFeaturesPanel.GetComponent<CanvasGroup>();
            if (panelGroup == null)
            {
                panelGroup = unlockedFeaturesPanel.AddComponent<CanvasGroup>();
            }

            float fadeTime = 0.3f;
            float elapsed = 0f;

            while (elapsed < fadeTime)
            {
                elapsed += Time.deltaTime;
                panelGroup.alpha = elapsed / fadeTime;
                yield return null;
            }

            // 機能をひとつずつ表示
            foreach (var feature in features)
            {
                if (featureUnlockSound != null)
                {
                    audioSource.PlayOneShot(featureUnlockSound, 0.5f);
                }

                GameObject item = Instantiate(featureItemPrefab, featureListContainer);
                TextMeshProUGUI text = item.GetComponentInChildren<TextMeshProUGUI>();
                if (text != null)
                {
                    text.text = $"• {GetFeatureDisplayName(feature)}";
                }

                // スケールアニメーション
                item.transform.localScale = Vector3.zero;
                float itemAnimTime = 0.3f;
                float itemElapsed = 0f;

                while (itemElapsed < itemAnimTime)
                {
                    itemElapsed += Time.deltaTime;
                    float t = itemElapsed / itemAnimTime;
                    item.transform.localScale = Vector3.one * scaleCurve.Evaluate(t);
                    yield return null;
                }

                yield return new WaitForSeconds(0.1f);
            }
        }

        /// <summary>
        /// フェードアウト演出
        /// </summary>
        private IEnumerator FadeOut()
        {
            float elapsed = 0f;
            CanvasGroup panelGroup = unlockedFeaturesPanel.GetComponent<CanvasGroup>();

            while (elapsed < fadeOutDuration)
            {
                elapsed += Time.deltaTime;
                float t = elapsed / fadeOutDuration;

                // 背景のフェード
                backgroundOverlay.color = new Color(0, 0, 0, Mathf.Lerp(0.8f, 0, t));

                // コンテナのスケール
                float scale = scaleCurve.Evaluate(1 - t);
                effectContainer.localScale = Vector3.one * scale;

                // 機能パネルのフェード
                if (panelGroup != null && unlockedFeaturesPanel.activeSelf)
                {
                    panelGroup.alpha = 1 - t;
                }

                yield return null;
            }
        }

        /// <summary>
        /// ランクに応じて解放される機能を取得
        /// </summary>
        private List<GameFeature> GetUnlockedFeaturesForRank(MerchantRank rank)
        {
            if (featureUnlockSystem == null)
                return new List<GameFeature>();

            return featureUnlockSystem.GetFeaturesForRank(rank);
        }

        /// <summary>
        /// ランクの表示名を取得
        /// </summary>
        private string GetRankDisplayName(MerchantRank rank)
        {
            return rank switch
            {
                MerchantRank.Apprentice => "見習い商人",
                MerchantRank.Skilled => "一人前商人",
                MerchantRank.Veteran => "ベテラン商人",
                MerchantRank.Master => "マスター商人",
                _ => rank.ToString(),
            };
        }

        /// <summary>
        /// ランクアイコンを取得
        /// </summary>
        private Sprite GetRankIcon(MerchantRank rank)
        {
            return rank switch
            {
                MerchantRank.Apprentice => apprenticeIcon,
                MerchantRank.Skilled => skilledIcon,
                MerchantRank.Veteran => veteranIcon,
                MerchantRank.Master => masterIcon,
                _ => null,
            };
        }

        /// <summary>
        /// ランクの色を取得
        /// </summary>
        private Color GetRankColor(MerchantRank rank)
        {
            return rank switch
            {
                MerchantRank.Apprentice => new Color(0.7f, 0.7f, 0.7f), // グレー
                MerchantRank.Skilled => new Color(0.2f, 0.8f, 0.2f), // 緑
                MerchantRank.Veteran => new Color(0.2f, 0.4f, 0.9f), // 青
                MerchantRank.Master => new Color(0.9f, 0.7f, 0.1f), // 金
                _ => Color.white,
            };
        }

        /// <summary>
        /// 機能の表示名を取得
        /// </summary>
        private string GetFeatureDisplayName(GameFeature feature)
        {
            return feature switch
            {
                GameFeature.BasicTrading => "基本取引",
                GameFeature.SimpleInventory => "シンプル在庫管理",
                GameFeature.SeasonalInfo => "季節情報",
                GameFeature.BasicPriceHistory => "基本価格履歴",
                GameFeature.PricePrediction => "価格予測機能",
                GameFeature.MarketTrends => "市場トレンド分析",
                GameFeature.BankAccount => "商人銀行口座",
                GameFeature.AdvancedInventory => "高度な在庫管理",
                GameFeature.AdvancedAnalytics => "高度な分析機能",
                GameFeature.EventPrediction => "イベント予測",
                GameFeature.ShopInvestment => "店舗投資",
                GameFeature.AutoPricing => "自動価格設定",
                GameFeature.MerchantNetwork => "商人ネットワーク",
                GameFeature.MarketManipulation => "市場操作",
                GameFeature.ExclusiveDeals => "独占取引",
                GameFeature.FullAutomation => "完全自動化",
                _ => feature.ToString(),
            };
        }

        /// <summary>
        /// テスト用：ランクアップ演出を手動で再生
        /// </summary>
        [ContextMenu("Test Rank Up Effect")]
        public void TestRankUpEffect()
        {
            if (Application.isPlaying)
            {
                StartCoroutine(PlayRankUpEffect(MerchantRank.Apprentice, MerchantRank.Skilled));
            }
        }
    }
}
