using System;
using System.Collections.Generic;
using UnityEngine;
using MerchantTails.Data;
using MerchantTails.Events;

namespace MerchantTails.Core
{
    /// <summary>
    /// 他商人への出資システム
    /// 配当による不労所得を管理
    /// </summary>
    public class MerchantInvestmentSystem : MonoBehaviour
    {
        private static MerchantInvestmentSystem instance;
        public static MerchantInvestmentSystem Instance => instance;

        [Header("Investment Settings")]
        [SerializeField] private int dividendFrequency = 7; // 配当頻度（日）
        [SerializeField] private float baseReturnRate = 0.1f; // 基本リターン率 10%/週
        [SerializeField] private float riskFactor = 0.3f; // リスク要因（配当のばらつき）

        [Header("Available Merchants")]
        [SerializeField] private List<MerchantProfile> availableMerchants = new List<MerchantProfile>();

        private Dictionary<string, MerchantInvestment> activeInvestments = new Dictionary<string, MerchantInvestment>();
        private PlayerData playerData;
        private FeatureUnlockSystem featureUnlockSystem;
        private int lastDividendDay = 0;
        private float totalDividendsEarned = 0f;

        // プロパティ
        public float TotalInvestmentValue => GetTotalInvestmentValue();
        public float TotalDividendsEarned => totalDividendsEarned;
        public bool IsUnlocked => featureUnlockSystem != null && featureUnlockSystem.IsFeatureUnlocked(GameFeature.MerchantNetwork);

        // イベント
        public event Action<MerchantProfile, float> OnInvestmentMade;
        public event Action<string, float> OnDividendReceived;
        public event Action<MerchantProfile> OnMerchantBankrupt;

        private void Awake()
        {
            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;

            InitializeMerchants();
        }

        private void OnDestroy()
        {
            if (instance == this)
            {
                instance = null;
            }
            UnsubscribeFromEvents();
        }

        private void Start()
        {
            playerData = GameManager.Instance?.PlayerData;
            featureUnlockSystem = FeatureUnlockSystem.Instance;

            SubscribeToEvents();
            LoadInvestmentData();
        }

        private void InitializeMerchants()
        {
            if (availableMerchants.Count == 0)
            {
                availableMerchants = new List<MerchantProfile>
                {
                    new MerchantProfile
                    {
                        id = "merchant_anna",
                        name = "商人アンナ",
                        description = "堅実な経営で知られる女性商人",
                        speciality = "ポーション取引",
                        riskLevel = RiskLevel.Low,
                        returnMultiplier = 0.8f,
                        minInvestment = 1000,
                        portraitName = "anna_portrait",
                    },
                    new MerchantProfile
                    {
                        id = "merchant_boris",
                        name = "商人ボリス",
                        description = "武器の目利きに定評がある",
                        speciality = "武器・防具",
                        riskLevel = RiskLevel.Medium,
                        returnMultiplier = 1.0f,
                        minInvestment = 2000,
                        portraitName = "boris_portrait",
                    },
                    new MerchantProfile
                    {
                        id = "merchant_clara",
                        name = "商人クララ",
                        description = "宝石取引で一攫千金を狙う",
                        speciality = "宝石・アクセサリー",
                        riskLevel = RiskLevel.High,
                        returnMultiplier = 1.5f,
                        minInvestment = 5000,
                        portraitName = "clara_portrait",
                    },
                    new MerchantProfile
                    {
                        id = "merchant_darius",
                        name = "商人ダリウス",
                        description = "魔法書の権威として知られる",
                        speciality = "魔法書・古文書",
                        riskLevel = RiskLevel.Medium,
                        returnMultiplier = 1.2f,
                        minInvestment = 3000,
                        portraitName = "darius_portrait",
                    },
                    new MerchantProfile
                    {
                        id = "merchant_elena",
                        name = "商人エレナ",
                        description = "季節商品の仕入れが得意",
                        speciality = "果物・季節商品",
                        riskLevel = RiskLevel.Low,
                        returnMultiplier = 0.9f,
                        minInvestment = 1500,
                        portraitName = "elena_portrait",
                    },
                };
            }
        }

        private void SubscribeToEvents()
        {
            EventBus.Subscribe<DayChangedEvent>(OnDayChanged);
            EventBus.Subscribe<FeatureUnlockedEvent>(OnFeatureUnlocked);
        }

        private void UnsubscribeFromEvents()
        {
            EventBus.Unsubscribe<DayChangedEvent>(OnDayChanged);
            EventBus.Unsubscribe<FeatureUnlockedEvent>(OnFeatureUnlocked);
        }

        private void OnDayChanged(DayChangedEvent e)
        {
            if (!IsUnlocked) return;

            // 配当計算
            if (e.NewDay - lastDividendDay >= dividendFrequency)
            {
                CalculateDividends(e.NewDay);
                lastDividendDay = e.NewDay;
            }

            // 商人の状態更新
            UpdateMerchantStatuses(e.NewDay);
        }

        private void OnFeatureUnlocked(FeatureUnlockedEvent e)
        {
            if (e.Feature == GameFeature.MerchantNetwork)
            {
                ErrorHandler.LogInfo("Merchant investment system unlocked!", "MerchantInvestment");
                EventBus.Publish(new MerchantNetworkUnlockedEvent());
            }
        }

        /// <summary>
        /// 商人に出資
        /// </summary>
        public bool InvestInMerchant(string merchantId, float amount)
        {
            if (!IsUnlocked)
            {
                ErrorHandler.LogWarning("Merchant network feature is not unlocked", "MerchantInvestment");
                return false;
            }

            var merchant = GetMerchant(merchantId);
            if (merchant == null)
            {
                ErrorHandler.LogError($"Merchant not found: {merchantId}", "MerchantInvestment");
                return false;
            }

            if (amount < merchant.minInvestment)
            {
                ErrorHandler.LogWarning($"Amount below minimum: {amount} < {merchant.minInvestment}", "MerchantInvestment");
                return false;
            }

            if (playerData.CurrentMoney < amount)
            {
                ErrorHandler.LogWarning("Insufficient funds for investment", "MerchantInvestment");
                return false;
            }

            // 出資処理
            if (playerData.ChangeMoney(-(int)amount))
            {
                MerchantInvestment investment;
                if (activeInvestments.TryGetValue(merchantId, out investment))
                {
                    // 既存の投資に追加
                    investment.totalInvested += amount;
                    investment.lastInvestmentDay = TimeManager.Instance?.CurrentDay ?? 1;
                }
                else
                {
                    // 新規投資
                    investment = new MerchantInvestment
                    {
                        merchantId = merchantId,
                        totalInvested = amount,
                        totalDividends = 0,
                        lastInvestmentDay = TimeManager.Instance?.CurrentDay ?? 1,
                        lastDividendDay = TimeManager.Instance?.CurrentDay ?? 1,
                        isActive = true,
                    };
                    activeInvestments[merchantId] = investment;
                }

                OnInvestmentMade?.Invoke(merchant, amount);
                EventBus.Publish(new MerchantInvestmentMadeEvent(merchant, amount));
                ErrorHandler.LogInfo($"Invested {amount}G in {merchant.name}", "MerchantInvestment");

                SaveInvestmentData();
                return true;
            }

            return false;
        }

        /// <summary>
        /// 出資を引き上げ
        /// </summary>
        public bool WithdrawInvestment(string merchantId, float amount)
        {
            if (!activeInvestments.TryGetValue(merchantId, out var investment) || !investment.isActive)
            {
                ErrorHandler.LogWarning($"No active investment found for merchant: {merchantId}", "MerchantInvestment");
                return false;
            }

            if (amount > investment.totalInvested)
            {
                ErrorHandler.LogWarning($"Withdrawal amount exceeds investment: {amount} > {investment.totalInvested}", "MerchantInvestment");
                return false;
            }

            // 引き上げ処理（手数料10%）
            float withdrawalFee = amount * 0.1f;
            float netAmount = amount - withdrawalFee;

            investment.totalInvested -= amount;
            playerData.ChangeMoney((int)netAmount);

            if (investment.totalInvested <= 0)
            {
                investment.isActive = false;
            }

            EventBus.Publish(new MerchantInvestmentWithdrawnEvent(merchantId, amount, withdrawalFee));
            ErrorHandler.LogInfo($"Withdrew {amount}G (net: {netAmount}G) from {merchantId}", "MerchantInvestment");

            SaveInvestmentData();
            return true;
        }

        /// <summary>
        /// 配当を計算して支払い
        /// </summary>
        private void CalculateDividends(int currentDay)
        {
            foreach (var kvp in activeInvestments)
            {
                if (!kvp.Value.isActive) continue;

                var merchant = GetMerchant(kvp.Key);
                if (merchant == null) continue;

                // 基本配当計算
                float baseReturn = kvp.Value.totalInvested * baseReturnRate * merchant.returnMultiplier;

                // リスクによる変動
                float riskModifier = GetRiskModifier(merchant.riskLevel);
                float dividend = baseReturn * riskModifier;

                // 最低保証（低リスクのみ）
                if (merchant.riskLevel == RiskLevel.Low && dividend < baseReturn * 0.5f)
                {
                    dividend = baseReturn * 0.5f;
                }

                // 配当支払い
                if (dividend > 0)
                {
                    playerData.ChangeMoney((int)dividend);
                    kvp.Value.totalDividends += dividend;
                    totalDividendsEarned += dividend;

                    OnDividendReceived?.Invoke(merchant.name, dividend);
                    EventBus.Publish(new DividendReceivedEvent(merchant, dividend));
                    ErrorHandler.LogInfo($"Dividend received from {merchant.name}: {dividend}G", "MerchantInvestment");
                }

                kvp.Value.lastDividendDay = currentDay;
            }
        }

        /// <summary>
        /// リスクに基づく配当変動を計算
        /// </summary>
        private float GetRiskModifier(RiskLevel riskLevel)
        {
            float baseModifier = 1f;
            float variance = 0f;

            switch (riskLevel)
            {
                case RiskLevel.Low:
                    variance = riskFactor * 0.3f; // ±9%
                    break;
                case RiskLevel.Medium:
                    variance = riskFactor * 0.6f; // ±18%
                    break;
                case RiskLevel.High:
                    variance = riskFactor; // ±30%
                    break;
            }

            // -variance から +variance*2 の範囲で変動（高リスクほど上振れの可能性）
            return baseModifier + UnityEngine.Random.Range(-variance, variance * 2f);
        }

        /// <summary>
        /// 商人の状態を更新（破産チェックなど）
        /// </summary>
        private void UpdateMerchantStatuses(int currentDay)
        {
            foreach (var merchant in availableMerchants)
            {
                // 高リスク商人の破産チェック（1%の確率）
                if (merchant.riskLevel == RiskLevel.High && UnityEngine.Random.Range(0f, 1f) < 0.01f)
                {
                    if (activeInvestments.TryGetValue(merchant.id, out var investment) && investment.isActive)
                    {
                        // 破産処理
                        investment.isActive = false;
                        OnMerchantBankrupt?.Invoke(merchant);
                        EventBus.Publish(new MerchantBankruptEvent(merchant, investment.totalInvested));
                        
                        ErrorHandler.LogWarning($"{merchant.name} has gone bankrupt! Lost investment: {investment.totalInvested}G", "MerchantInvestment");
                    }
                }
            }
        }

        /// <summary>
        /// 特定の商人情報を取得
        /// </summary>
        public MerchantProfile GetMerchant(string merchantId)
        {
            return availableMerchants.Find(m => m.id == merchantId);
        }

        /// <summary>
        /// 投資情報を取得
        /// </summary>
        public MerchantInvestment GetInvestment(string merchantId)
        {
            return activeInvestments.TryGetValue(merchantId, out var investment) ? investment : null;
        }

        /// <summary>
        /// 総投資額を取得
        /// </summary>
        public float GetTotalInvestmentValue()
        {
            float total = 0f;
            foreach (var investment in activeInvestments.Values)
            {
                if (investment.isActive)
                {
                    total += investment.totalInvested;
                }
            }
            return total;
        }

        /// <summary>
        /// 利用可能な商人リストを取得
        /// </summary>
        public List<MerchantProfile> GetAvailableMerchants()
        {
            return new List<MerchantProfile>(availableMerchants);
        }

        // セーブ/ロード
        private void SaveInvestmentData()
        {
            var data = new MerchantInvestmentSaveData
            {
                activeInvestments = activeInvestments,
                totalDividendsEarned = totalDividendsEarned,
                lastDividendDay = lastDividendDay,
            };

            string json = JsonUtility.ToJson(data);
            PlayerPrefs.SetString($"MerchantInvestment_{playerData?.PlayerName}", json);
            PlayerPrefs.Save();
        }

        private void LoadInvestmentData()
        {
            string key = $"MerchantInvestment_{playerData?.PlayerName}";
            if (PlayerPrefs.HasKey(key))
            {
                string json = PlayerPrefs.GetString(key);
                var data = JsonUtility.FromJson<MerchantInvestmentSaveData>(json);

                activeInvestments = data.activeInvestments ?? new Dictionary<string, MerchantInvestment>();
                totalDividendsEarned = data.totalDividendsEarned;
                lastDividendDay = data.lastDividendDay;
            }
        }
    }

    /// <summary>
    /// 商人プロフィール
    /// </summary>
    [Serializable]
    public class MerchantProfile
    {
        public string id;
        public string name;
        public string description;
        public string speciality;
        public RiskLevel riskLevel;
        public float returnMultiplier;
        public float minInvestment;
        public string portraitName;
    }

    /// <summary>
    /// リスクレベル
    /// </summary>
    public enum RiskLevel
    {
        Low,    // 低リスク・低リターン
        Medium, // 中リスク・中リターン
        High,   // 高リスク・高リターン
    }

    /// <summary>
    /// 投資データ
    /// </summary>
    [Serializable]
    public class MerchantInvestment
    {
        public string merchantId;
        public float totalInvested;
        public float totalDividends;
        public int lastInvestmentDay;
        public int lastDividendDay;
        public bool isActive;
    }

    /// <summary>
    /// 保存データ
    /// </summary>
    [Serializable]
    public class MerchantInvestmentSaveData
    {
        public Dictionary<string, MerchantInvestment> activeInvestments;
        public float totalDividendsEarned;
        public int lastDividendDay;
    }

    // イベント
    public class MerchantNetworkUnlockedEvent : BaseGameEvent
    {
    }

    public class MerchantInvestmentMadeEvent : BaseGameEvent
    {
        public MerchantProfile Merchant { get; }
        public float Amount { get; }

        public MerchantInvestmentMadeEvent(MerchantProfile merchant, float amount)
        {
            Merchant = merchant;
            Amount = amount;
        }
    }

    public class MerchantInvestmentWithdrawnEvent : BaseGameEvent
    {
        public string MerchantId { get; }
        public float Amount { get; }
        public float Fee { get; }

        public MerchantInvestmentWithdrawnEvent(string merchantId, float amount, float fee)
        {
            MerchantId = merchantId;
            Amount = amount;
            Fee = fee;
        }
    }

    public class DividendReceivedEvent : BaseGameEvent
    {
        public MerchantProfile Merchant { get; }
        public float Amount { get; }

        public DividendReceivedEvent(MerchantProfile merchant, float amount)
        {
            Merchant = merchant;
            Amount = amount;
        }
    }

    public class MerchantBankruptEvent : BaseGameEvent
    {
        public MerchantProfile Merchant { get; }
        public float LostAmount { get; }

        public MerchantBankruptEvent(MerchantProfile merchant, float lostAmount)
        {
            Merchant = merchant;
            LostAmount = lostAmount;
        }
    }
}