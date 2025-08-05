using System;
using System.Collections.Generic;
using MerchantTails.Data;
using MerchantTails.UI;
using UnityEngine;

namespace MerchantTails.Core
{
    /// <summary>
    /// 商人銀行システム
    /// 預金、複利計算、定期預金などを管理
    /// </summary>
    public class BankSystem : MonoBehaviour
    {
        private static BankSystem instance;
        public static BankSystem Instance => instance;

        [Header("Bank Settings")]
        [SerializeField]
        private float baseInterestRate = 0.02f; // 基本利率 2%/日

        [SerializeField]
        private float veteranBonusRate = 0.01f; // ベテランボーナス 1%

        [SerializeField]
        private float masterBonusRate = 0.02f; // マスターボーナス 2%

        [SerializeField]
        private int compoundFrequency = 1; // 複利計算頻度（日）

        [SerializeField]
        private float withdrawalFee = 0.01f; // 引き出し手数料 1%

        [Header("Term Deposit Settings")]
        [SerializeField]
        private List<TermDepositType> termDepositTypes = new List<TermDepositType>();

        private PlayerData playerData;
        private FeatureUnlockSystem featureUnlockSystem;

        // 預金データ
        private float regularDeposit = 0f;
        private List<TermDeposit> termDeposits = new List<TermDeposit>();
        private float totalInterestEarned = 0f;
        private int lastCompoundDay = 0;

        // プロパティ
        public float RegularDeposit => regularDeposit;
        public float TotalDeposits => GetTotalDeposits();
        public float TotalInterestEarned => totalInterestEarned;
        public float CurrentInterestRate => GetCurrentInterestRate();
        public bool IsUnlocked =>
            featureUnlockSystem != null && featureUnlockSystem.IsFeatureUnlocked(GameFeature.BankAccount);

        // イベント
        public event Action<float> OnDepositChanged;
        public event Action<float> OnInterestEarned;
        public event Action<TermDeposit> OnTermDepositCreated;
        public event Action<TermDeposit> OnTermDepositMatured;

        private void Awake()
        {
            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;

            InitializeTermDepositTypes();
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
            LoadBankData();
        }

        private void InitializeTermDepositTypes()
        {
            if (termDepositTypes.Count == 0)
            {
                termDepositTypes = new List<TermDepositType>
                {
                    new TermDepositType
                    {
                        name = "短期定期（7日）",
                        durationDays = 7,
                        interestRate = 0.15f, // 15%
                        minDeposit = 1000,
                        earlyWithdrawalPenalty = 0.5f, // 50%ペナルティ
                    },
                    new TermDepositType
                    {
                        name = "中期定期（30日）",
                        durationDays = 30,
                        interestRate = 0.8f, // 80%
                        minDeposit = 5000,
                        earlyWithdrawalPenalty = 0.7f,
                    },
                    new TermDepositType
                    {
                        name = "長期定期（90日）",
                        durationDays = 90,
                        interestRate = 3.0f, // 300%
                        minDeposit = 10000,
                        earlyWithdrawalPenalty = 0.8f,
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
            if (!IsUnlocked)
                return;

            // 複利計算
            if (e.NewDay - lastCompoundDay >= compoundFrequency)
            {
                CalculateCompoundInterest();
                lastCompoundDay = e.NewDay;
            }

            // 定期預金の満期チェック
            CheckTermDeposits(e.NewDay);
        }

        private void OnFeatureUnlocked(FeatureUnlockedEvent e)
        {
            if (e.Feature == GameFeature.BankAccount)
            {
                // 銀行機能が解放されたら初期化
                ErrorHandler.LogInfo("Bank system unlocked!", "BankSystem");
                EventBus.Publish(new BankUnlockedEvent());
            }
        }

        /// <summary>
        /// 通常預金に入金
        /// </summary>
        public bool Deposit(float amount)
        {
            if (!IsUnlocked)
            {
                ErrorHandler.LogWarning("Bank feature is not unlocked yet", "BankSystem");
                return false;
            }

            if (amount <= 0 || playerData == null)
            {
                return false;
            }

            // プレイヤーの所持金をチェック
            if (playerData.CurrentMoney < amount)
            {
                ErrorHandler.LogWarning(
                    $"Insufficient funds for deposit. Required: {amount}, Available: {playerData.CurrentMoney}",
                    "BankSystem"
                );
                return false;
            }

            // 入金処理
            if (playerData.ChangeMoney(-(int)amount))
            {
                regularDeposit += amount;
                OnDepositChanged?.Invoke(regularDeposit);

                EventBus.Publish(new BankTransactionEvent(BankTransactionType.Deposit, amount, regularDeposit));
                ErrorHandler.LogInfo($"Deposited {amount}G. New balance: {regularDeposit}G", "BankSystem");

                SaveBankData();
                return true;
            }

            return false;
        }

        /// <summary>
        /// 通常預金から引き出し
        /// </summary>
        public bool Withdraw(float amount)
        {
            if (!IsUnlocked)
                return false;

            if (amount <= 0 || regularDeposit < amount)
            {
                ErrorHandler.LogWarning(
                    $"Invalid withdrawal amount: {amount}, Available: {regularDeposit}",
                    "BankSystem"
                );
                return false;
            }

            // 手数料を計算
            float fee = amount * withdrawalFee;
            float totalWithdrawal = amount + fee;

            if (regularDeposit < totalWithdrawal)
            {
                ErrorHandler.LogWarning(
                    $"Insufficient funds including fee. Total required: {totalWithdrawal}",
                    "BankSystem"
                );
                return false;
            }

            // 引き出し処理
            regularDeposit -= totalWithdrawal;
            playerData.ChangeMoney((int)amount); // 手数料を差し引いた額を受け取る

            OnDepositChanged?.Invoke(regularDeposit);
            EventBus.Publish(new BankTransactionEvent(BankTransactionType.Withdrawal, amount, regularDeposit, fee));

            ErrorHandler.LogInfo($"Withdrew {amount}G (fee: {fee}G). New balance: {regularDeposit}G", "BankSystem");
            SaveBankData();
            return true;
        }

        /// <summary>
        /// 定期預金を作成
        /// </summary>
        public bool CreateTermDeposit(int typeIndex, float amount)
        {
            if (!IsUnlocked)
                return false;

            if (typeIndex < 0 || typeIndex >= termDepositTypes.Count)
            {
                ErrorHandler.LogError("Invalid term deposit type", null, "BankSystem");
                return false;
            }

            var type = termDepositTypes[typeIndex];

            if (amount < type.minDeposit)
            {
                ErrorHandler.LogWarning($"Amount below minimum: {amount} < {type.minDeposit}", "BankSystem");
                return false;
            }

            if (playerData.CurrentMoney < amount)
            {
                ErrorHandler.LogWarning("Insufficient funds for term deposit", "BankSystem");
                return false;
            }

            // 定期預金を作成
            if (playerData.ChangeMoney(-(int)amount))
            {
                var termDeposit = new TermDeposit
                {
                    id = Guid.NewGuid().ToString(),
                    typeIndex = typeIndex,
                    principal = amount,
                    interestRate = type.interestRate,
                    startDay = TimeManager.Instance?.CurrentDay ?? 1,
                    maturityDay = (TimeManager.Instance?.CurrentDay ?? 1) + type.durationDays,
                    isMatured = false,
                };

                termDeposits.Add(termDeposit);
                OnTermDepositCreated?.Invoke(termDeposit);

                EventBus.Publish(new TermDepositCreatedEvent(termDeposit));
                ErrorHandler.LogInfo($"Created term deposit: {type.name}, Amount: {amount}G", "BankSystem");

                SaveBankData();
                return true;
            }

            return false;
        }

        /// <summary>
        /// 定期預金を解約（早期解約にはペナルティ）
        /// </summary>
        public bool BreakTermDeposit(string depositId, bool isEarlyWithdrawal = false)
        {
            var deposit = termDeposits.Find(td => td.id == depositId);
            if (deposit == null)
                return false;

            var type = termDepositTypes[deposit.typeIndex];
            float returnAmount = deposit.principal;

            if (deposit.isMatured)
            {
                // 満期の場合は利息込みで返金
                returnAmount = deposit.principal * (1 + deposit.interestRate);
            }
            else if (isEarlyWithdrawal)
            {
                // 早期解約の場合はペナルティ
                returnAmount = deposit.principal * (1 - type.earlyWithdrawalPenalty);
            }

            // 返金処理
            playerData.ChangeMoney((int)returnAmount);
            termDeposits.Remove(deposit);

            EventBus.Publish(new TermDepositWithdrawnEvent(deposit, returnAmount, isEarlyWithdrawal));
            ErrorHandler.LogInfo($"Term deposit withdrawn: {returnAmount}G (Early: {isEarlyWithdrawal})", "BankSystem");

            SaveBankData();
            return true;
        }

        /// <summary>
        /// 複利計算を実行
        /// </summary>
        private void CalculateCompoundInterest()
        {
            if (regularDeposit <= 0)
                return;

            float rate = GetCurrentInterestRate();
            float interest = regularDeposit * rate;

            regularDeposit += interest;
            totalInterestEarned += interest;

            OnDepositChanged?.Invoke(regularDeposit);
            OnInterestEarned?.Invoke(interest);

            EventBus.Publish(new InterestEarnedEvent(interest, regularDeposit, rate));
            ErrorHandler.LogInfo(
                $"Interest earned: {interest}G at {rate * 100}% rate. New balance: {regularDeposit}G",
                "BankSystem"
            );
        }

        /// <summary>
        /// 現在の利率を取得
        /// </summary>
        private float GetCurrentInterestRate()
        {
            float rate = baseInterestRate;

            if (playerData != null)
            {
                switch (playerData.CurrentRank)
                {
                    case MerchantRank.Veteran:
                        rate += veteranBonusRate;
                        break;
                    case MerchantRank.Master:
                        rate += veteranBonusRate + masterBonusRate;
                        break;
                }
            }

            return rate;
        }

        /// <summary>
        /// 定期預金の満期をチェック
        /// </summary>
        private void CheckTermDeposits(int currentDay)
        {
            foreach (var deposit in termDeposits)
            {
                if (!deposit.isMatured && currentDay >= deposit.maturityDay)
                {
                    deposit.isMatured = true;
                    OnTermDepositMatured?.Invoke(deposit);

                    // 通知を表示
                    if (UIManager.Instance != null)
                    {
                        var type = termDepositTypes[deposit.typeIndex];
                        float maturityAmount = deposit.principal * (1 + deposit.interestRate);
                        UIManager.Instance.ShowNotification(
                            "定期預金満期",
                            $"{type.name}が満期になりました！\n元本: {deposit.principal}G → {maturityAmount}G",
                            5f,
                            UIManager.NotificationType.Success
                        );
                    }

                    EventBus.Publish(new TermDepositMaturedEvent(deposit));
                }
            }
        }

        /// <summary>
        /// 総預金額を取得
        /// </summary>
        public float GetTotalDeposits()
        {
            float total = regularDeposit;

            foreach (var deposit in termDeposits)
            {
                if (deposit.isMatured)
                {
                    total += deposit.principal * (1 + deposit.interestRate);
                }
                else
                {
                    total += deposit.principal;
                }
            }

            return total;
        }

        /// <summary>
        /// 定期預金リストを取得
        /// </summary>
        public List<TermDeposit> GetTermDeposits()
        {
            return new List<TermDeposit>(termDeposits);
        }

        /// <summary>
        /// 定期預金タイプを取得
        /// </summary>
        public TermDepositType GetTermDepositType(int index)
        {
            if (index >= 0 && index < termDepositTypes.Count)
            {
                return termDepositTypes[index];
            }
            return null;
        }

        // セーブ/ロード
        private void SaveBankData()
        {
            var data = new BankSaveData
            {
                regularDeposit = regularDeposit,
                termDeposits = termDeposits,
                totalInterestEarned = totalInterestEarned,
                lastCompoundDay = lastCompoundDay,
            };

            string json = JsonUtility.ToJson(data);
            PlayerPrefs.SetString($"BankData_{playerData?.PlayerName}", json);
            PlayerPrefs.Save();
        }

        private void LoadBankData()
        {
            string key = $"BankData_{playerData?.PlayerName}";
            if (PlayerPrefs.HasKey(key))
            {
                string json = PlayerPrefs.GetString(key);
                var data = JsonUtility.FromJson<BankSaveData>(json);

                regularDeposit = data.regularDeposit;
                termDeposits = data.termDeposits ?? new List<TermDeposit>();
                totalInterestEarned = data.totalInterestEarned;
                lastCompoundDay = data.lastCompoundDay;
            }
        }
    }

    /// <summary>
    /// 定期預金タイプ
    /// </summary>
    [Serializable]
    public class TermDepositType
    {
        public string name;
        public int durationDays;
        public float interestRate;
        public float minDeposit;
        public float earlyWithdrawalPenalty;
    }

    /// <summary>
    /// 定期預金データ
    /// </summary>
    [Serializable]
    public class TermDeposit
    {
        public string id;
        public int typeIndex;
        public float principal;
        public float interestRate;
        public int startDay;
        public int maturityDay;
        public bool isMatured;
    }

    /// <summary>
    /// 銀行データの保存形式
    /// </summary>
    [Serializable]
    public class BankSaveData
    {
        public float regularDeposit;
        public List<TermDeposit> termDeposits;
        public float totalInterestEarned;
        public int lastCompoundDay;
    }

    // 銀行関連イベント
    public enum BankTransactionType
    {
        Deposit,
        Withdrawal,
    }

    public class BankTransactionEvent : BaseGameEvent
    {
        public BankTransactionType Type { get; }
        public float Amount { get; }
        public float NewBalance { get; }
        public float Fee { get; }

        public BankTransactionEvent(BankTransactionType type, float amount, float newBalance, float fee = 0)
        {
            Type = type;
            Amount = amount;
            NewBalance = newBalance;
            Fee = fee;
        }
    }

    public class InterestEarnedEvent : BaseGameEvent
    {
        public float Interest { get; }
        public float NewBalance { get; }
        public float Rate { get; }

        public InterestEarnedEvent(float interest, float newBalance, float rate)
        {
            Interest = interest;
            NewBalance = newBalance;
            Rate = rate;
        }
    }

    public class TermDepositCreatedEvent : BaseGameEvent
    {
        public TermDeposit Deposit { get; }

        public TermDepositCreatedEvent(TermDeposit deposit)
        {
            Deposit = deposit;
        }
    }

    public class TermDepositMaturedEvent : BaseGameEvent
    {
        public TermDeposit Deposit { get; }

        public TermDepositMaturedEvent(TermDeposit deposit)
        {
            Deposit = deposit;
        }
    }

    public class TermDepositWithdrawnEvent : BaseGameEvent
    {
        public TermDeposit Deposit { get; }
        public float ReturnAmount { get; }
        public bool IsEarlyWithdrawal { get; }

        public TermDepositWithdrawnEvent(TermDeposit deposit, float returnAmount, bool isEarlyWithdrawal)
        {
            Deposit = deposit;
            ReturnAmount = returnAmount;
            IsEarlyWithdrawal = isEarlyWithdrawal;
        }
    }

    public class BankUnlockedEvent : BaseGameEvent { }
}
