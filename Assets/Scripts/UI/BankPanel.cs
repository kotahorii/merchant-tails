using System;
using System.Collections.Generic;
using MerchantTails.Core;
using MerchantTails.Data;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.UI
{
    /// <summary>
    /// 商人銀行のUIパネル
    /// </summary>
    public class BankPanel : MonoBehaviour
    {
        [Header("Main UI References")]
        [SerializeField]
        private TextMeshProUGUI currentMoneyText;

        [SerializeField]
        private TextMeshProUGUI regularDepositText;

        [SerializeField]
        private TextMeshProUGUI totalDepositText;

        [SerializeField]
        private TextMeshProUGUI interestRateText;

        [SerializeField]
        private TextMeshProUGUI totalInterestText;

        [Header("Regular Deposit")]
        [SerializeField]
        private TMP_InputField depositAmountInput;

        [SerializeField]
        private Button depositButton;

        [SerializeField]
        private TMP_InputField withdrawAmountInput;

        [SerializeField]
        private Button withdrawButton;

        [SerializeField]
        private TextMeshProUGUI withdrawFeeText;

        [Header("Term Deposits")]
        [SerializeField]
        private Transform termDepositTypeContainer;

        [SerializeField]
        private GameObject termDepositTypePrefab;

        [SerializeField]
        private Transform activeTermDepositsContainer;

        [SerializeField]
        private GameObject activeTermDepositPrefab;

        [Header("Transaction History")]
        [SerializeField]
        private Transform transactionHistoryContainer;

        [SerializeField]
        private GameObject transactionItemPrefab;

        [SerializeField]
        private int maxHistoryItems = 20;

        [Header("UI Settings")]
        [SerializeField]
        private Color depositColor = Color.green;

        [SerializeField]
        private Color withdrawColor = Color.red;

        [SerializeField]
        private Color interestColor = Color.cyan;

        [SerializeField]
        private Color maturedColor = new Color(1f, 0.8f, 0.2f);

        private BankSystem bankSystem;
        private PlayerData playerData;
        private List<TermDepositTypeUI> termDepositTypes = new List<TermDepositTypeUI>();
        private List<ActiveTermDepositUI> activeTermDeposits = new List<ActiveTermDepositUI>();
        private Queue<GameObject> transactionHistory = new Queue<GameObject>();

        private void Start()
        {
            bankSystem = BankSystem.Instance;
            playerData = GameManager.Instance?.PlayerData;

            if (bankSystem == null)
            {
                ErrorHandler.LogError("BankSystem not found!", null, "BankPanel");
                return;
            }

            InitializeUI();
            SubscribeToEvents();
            RefreshDisplay();
        }

        private void OnDestroy()
        {
            UnsubscribeFromEvents();
        }

        private void InitializeUI()
        {
            // 入金ボタン
            depositButton.onClick.AddListener(OnDepositClick);
            depositAmountInput.onValueChanged.AddListener(OnDepositAmountChanged);

            // 引き出しボタン
            withdrawButton.onClick.AddListener(OnWithdrawClick);
            withdrawAmountInput.onValueChanged.AddListener(OnWithdrawAmountChanged);

            // 定期預金タイプの表示
            InitializeTermDepositTypes();
        }

        private void SubscribeToEvents()
        {
            if (bankSystem != null)
            {
                bankSystem.OnDepositChanged += OnDepositChanged;
                bankSystem.OnInterestEarned += OnInterestEarned;
                bankSystem.OnTermDepositCreated += OnTermDepositCreated;
                bankSystem.OnTermDepositMatured += OnTermDepositMatured;
            }

            EventBus.Subscribe<BankTransactionEvent>(OnBankTransaction);
            EventBus.Subscribe<MoneyChangedEvent>(OnMoneyChanged);
        }

        private void UnsubscribeFromEvents()
        {
            if (bankSystem != null)
            {
                bankSystem.OnDepositChanged -= OnDepositChanged;
                bankSystem.OnInterestEarned -= OnInterestEarned;
                bankSystem.OnTermDepositCreated -= OnTermDepositCreated;
                bankSystem.OnTermDepositMatured -= OnTermDepositMatured;
            }

            EventBus.Unsubscribe<BankTransactionEvent>(OnBankTransaction);
            EventBus.Unsubscribe<MoneyChangedEvent>(OnMoneyChanged);
        }

        private void RefreshDisplay()
        {
            UpdateMoneyDisplay();
            UpdateDepositDisplay();
            UpdateInterestDisplay();
            RefreshActiveTermDeposits();
        }

        private void UpdateMoneyDisplay()
        {
            if (playerData != null && currentMoneyText != null)
            {
                currentMoneyText.text = $"所持金: {playerData.CurrentMoney:N0}G";
            }
        }

        private void UpdateDepositDisplay()
        {
            if (bankSystem != null)
            {
                regularDepositText.text = $"普通預金: {bankSystem.RegularDeposit:N0}G";
                totalDepositText.text = $"総預金額: {bankSystem.TotalDeposits:N0}G";
            }
        }

        private void UpdateInterestDisplay()
        {
            if (bankSystem != null)
            {
                float rate = bankSystem.CurrentInterestRate;
                interestRateText.text = $"現在の利率: {rate * 100:F2}%/日";
                totalInterestText.text = $"累計利息: {bankSystem.TotalInterestEarned:N0}G";

                // ランクボーナスの表示
                if (playerData != null)
                {
                    string bonus = "";
                    switch (playerData.CurrentRank)
                    {
                        case MerchantRank.Veteran:
                            bonus = " (ベテランボーナス +1%)";
                            break;
                        case MerchantRank.Master:
                            bonus = " (マスターボーナス +3%)";
                            break;
                    }
                    interestRateText.text += bonus;
                }
            }
        }

        private void InitializeTermDepositTypes()
        {
            // 既存のUIをクリア
            foreach (var ui in termDepositTypes)
            {
                Destroy(ui.gameObject);
            }
            termDepositTypes.Clear();

            // 定期預金タイプを表示
            for (int i = 0; i < 3; i++) // 最大3種類
            {
                var type = bankSystem.GetTermDepositType(i);
                if (type != null)
                {
                    CreateTermDepositTypeUI(i, type);
                }
            }
        }

        private void CreateTermDepositTypeUI(int index, TermDepositType type)
        {
            GameObject obj = Instantiate(termDepositTypePrefab, termDepositTypeContainer);
            TermDepositTypeUI ui = obj.GetComponent<TermDepositTypeUI>();

            if (ui == null)
            {
                ui = obj.AddComponent<TermDepositTypeUI>();
            }

            ui.Setup(index, type, playerData?.CurrentMoney ?? 0);
            ui.OnCreateDeposit += OnCreateTermDeposit;
            termDepositTypes.Add(ui);
        }

        private void RefreshActiveTermDeposits()
        {
            // 既存のUIをクリア
            foreach (var ui in activeTermDeposits)
            {
                Destroy(ui.gameObject);
            }
            activeTermDeposits.Clear();

            // アクティブな定期預金を表示
            var deposits = bankSystem.GetTermDeposits();
            foreach (var deposit in deposits)
            {
                CreateActiveTermDepositUI(deposit);
            }
        }

        private void CreateActiveTermDepositUI(TermDeposit deposit)
        {
            GameObject obj = Instantiate(activeTermDepositPrefab, activeTermDepositsContainer);
            ActiveTermDepositUI ui = obj.GetComponent<ActiveTermDepositUI>();

            if (ui == null)
            {
                ui = obj.AddComponent<ActiveTermDepositUI>();
            }

            var type = bankSystem.GetTermDepositType(deposit.typeIndex);
            int currentDay = TimeManager.Instance?.CurrentDay ?? 1;

            ui.Setup(deposit, type, currentDay, maturedColor);
            ui.OnWithdraw += OnWithdrawTermDeposit;
            activeTermDeposits.Add(ui);
        }

        // UI イベントハンドラ
        private void OnDepositClick()
        {
            if (float.TryParse(depositAmountInput.text, out float amount))
            {
                if (bankSystem.Deposit(amount))
                {
                    depositAmountInput.text = "";
                    RefreshDisplay();
                }
                else
                {
                    // エラー表示
                    if (UIManager.Instance != null)
                    {
                        UIManager.Instance.ShowNotification(
                            "入金失敗",
                            "所持金が不足しています",
                            3f,
                            UIManager.NotificationType.Error
                        );
                    }
                }
            }
        }

        private void OnWithdrawClick()
        {
            if (float.TryParse(withdrawAmountInput.text, out float amount))
            {
                if (bankSystem.Withdraw(amount))
                {
                    withdrawAmountInput.text = "";
                    RefreshDisplay();
                }
                else
                {
                    // エラー表示
                    if (UIManager.Instance != null)
                    {
                        UIManager.Instance.ShowNotification(
                            "引き出し失敗",
                            "預金残高が不足しています",
                            3f,
                            UIManager.NotificationType.Error
                        );
                    }
                }
            }
        }

        private void OnDepositAmountChanged(string value)
        {
            // 入金可能かチェック
            bool canDeposit = false;
            if (float.TryParse(value, out float amount) && amount > 0)
            {
                canDeposit = playerData?.CurrentMoney >= amount;
            }
            depositButton.interactable = canDeposit;
        }

        private void OnWithdrawAmountChanged(string value)
        {
            // 引き出し可能かチェック
            bool canWithdraw = false;
            float fee = 0;

            if (float.TryParse(value, out float amount) && amount > 0)
            {
                fee = amount * 0.01f; // 1%手数料
                float totalRequired = amount + fee;
                canWithdraw = bankSystem?.RegularDeposit >= totalRequired;
            }

            withdrawButton.interactable = canWithdraw;

            // 手数料表示
            if (withdrawFeeText != null)
            {
                withdrawFeeText.text = fee > 0 ? $"手数料: {fee:N0}G" : "";
            }
        }

        private void OnCreateTermDeposit(int typeIndex, float amount)
        {
            if (bankSystem.CreateTermDeposit(typeIndex, amount))
            {
                RefreshDisplay();
            }
        }

        private void OnWithdrawTermDeposit(string depositId, bool isEarly)
        {
            if (isEarly)
            {
                // 早期解約の確認
                if (UIManager.Instance != null)
                {
                    UIManager.Instance.ShowConfirmDialog(
                        "早期解約の確認",
                        "早期解約するとペナルティが発生します。本当に解約しますか？",
                        () =>
                        {
                            bankSystem.BreakTermDeposit(depositId, true);
                            RefreshDisplay();
                        },
                        null
                    );
                }
            }
            else
            {
                bankSystem.BreakTermDeposit(depositId, false);
                RefreshDisplay();
            }
        }

        // イベントハンドラ
        private void OnDepositChanged(float newBalance)
        {
            UpdateDepositDisplay();
        }

        private void OnInterestEarned(float interest)
        {
            UpdateInterestDisplay();
            AddTransactionHistory($"利息: +{interest:N0}G", interestColor);
        }

        private void OnTermDepositCreated(TermDeposit deposit)
        {
            RefreshActiveTermDeposits();
        }

        private void OnTermDepositMatured(TermDeposit deposit)
        {
            RefreshActiveTermDeposits();
        }

        private void OnBankTransaction(BankTransactionEvent e)
        {
            string message = "";
            Color color = depositColor;

            switch (e.Type)
            {
                case BankTransactionType.Deposit:
                    message = $"入金: +{e.Amount:N0}G";
                    color = depositColor;
                    break;
                case BankTransactionType.Withdrawal:
                    message = $"引き出し: -{e.Amount:N0}G";
                    if (e.Fee > 0)
                        message += $" (手数料: {e.Fee:N0}G)";
                    color = withdrawColor;
                    break;
            }

            AddTransactionHistory(message, color);
        }

        private void OnMoneyChanged(MoneyChangedEvent e)
        {
            UpdateMoneyDisplay();
        }

        private void AddTransactionHistory(string message, Color color)
        {
            if (transactionHistoryContainer == null || transactionItemPrefab == null)
                return;

            GameObject item = Instantiate(transactionItemPrefab, transactionHistoryContainer);
            TextMeshProUGUI text = item.GetComponentInChildren<TextMeshProUGUI>();

            if (text != null)
            {
                text.text = $"{System.DateTime.Now:HH:mm} {message}";
                text.color = color;
            }

            transactionHistory.Enqueue(item);

            // 履歴の上限を管理
            while (transactionHistory.Count > maxHistoryItems)
            {
                GameObject oldItem = transactionHistory.Dequeue();
                Destroy(oldItem);
            }

            // 最新の履歴を上に表示
            item.transform.SetSiblingIndex(0);
        }
    }

    /// <summary>
    /// 定期預金タイプのUI
    /// </summary>
    public class TermDepositTypeUI : MonoBehaviour
    {
        [SerializeField]
        private TextMeshProUGUI nameText;

        [SerializeField]
        private TextMeshProUGUI durationText;

        [SerializeField]
        private TextMeshProUGUI interestRateText;

        [SerializeField]
        private TextMeshProUGUI minDepositText;

        [SerializeField]
        private TMP_InputField amountInput;

        [SerializeField]
        private Button createButton;

        private int typeIndex;
        private TermDepositType type;
        private float playerMoney;

        public event System.Action<int, float> OnCreateDeposit;

        public void Setup(int index, TermDepositType depositType, float currentMoney)
        {
            typeIndex = index;
            type = depositType;
            playerMoney = currentMoney;

            if (nameText != null)
                nameText.text = type.name;
            if (durationText != null)
                durationText.text = $"期間: {type.durationDays}日";
            if (interestRateText != null)
                interestRateText.text = $"利率: {type.interestRate * 100:F0}%";
            if (minDepositText != null)
                minDepositText.text = $"最低預金額: {type.minDeposit:N0}G";

            if (amountInput != null)
            {
                amountInput.onValueChanged.AddListener(OnAmountChanged);
            }

            if (createButton != null)
            {
                createButton.onClick.AddListener(OnCreateClick);
                createButton.interactable = false;
            }
        }

        private void OnAmountChanged(string value)
        {
            bool canCreate = false;
            if (float.TryParse(value, out float amount))
            {
                canCreate = amount >= type.minDeposit && amount <= playerMoney;
            }

            if (createButton != null)
            {
                createButton.interactable = canCreate;
            }
        }

        private void OnCreateClick()
        {
            if (float.TryParse(amountInput.text, out float amount))
            {
                OnCreateDeposit?.Invoke(typeIndex, amount);
                amountInput.text = "";
            }
        }
    }

    /// <summary>
    /// アクティブな定期預金のUI
    /// </summary>
    public class ActiveTermDepositUI : MonoBehaviour
    {
        [SerializeField]
        private TextMeshProUGUI typeText;

        [SerializeField]
        private TextMeshProUGUI principalText;

        [SerializeField]
        private TextMeshProUGUI maturityText;

        [SerializeField]
        private TextMeshProUGUI statusText;

        [SerializeField]
        private Slider progressBar;

        [SerializeField]
        private Button withdrawButton;

        [SerializeField]
        private TextMeshProUGUI withdrawButtonText;

        private TermDeposit deposit;
        private TermDepositType type;
        private bool isMatured;

        public event System.Action<string, bool> OnWithdraw;

        public void Setup(TermDeposit termDeposit, TermDepositType depositType, int currentDay, Color maturedColor)
        {
            deposit = termDeposit;
            type = depositType;
            isMatured = deposit.isMatured;

            if (typeText != null)
                typeText.text = type.name;
            if (principalText != null)
                principalText.text = $"元本: {deposit.principal:N0}G";

            // 満期情報
            int daysRemaining = deposit.maturityDay - currentDay;
            float maturityAmount = deposit.principal * (1 + deposit.interestRate);

            if (maturityText != null)
            {
                if (isMatured)
                {
                    maturityText.text = $"満期額: {maturityAmount:N0}G";
                    maturityText.color = maturedColor;
                }
                else
                {
                    maturityText.text = $"満期まで: {daysRemaining}日";
                }
            }

            // ステータス
            if (statusText != null)
            {
                if (isMatured)
                {
                    statusText.text = "満期";
                    statusText.color = maturedColor;
                }
                else
                {
                    float progress = (float)(currentDay - deposit.startDay) / (deposit.maturityDay - deposit.startDay);
                    statusText.text = $"運用中 ({progress * 100:F0}%)";
                }
            }

            // プログレスバー
            if (progressBar != null)
            {
                float progress = (float)(currentDay - deposit.startDay) / (deposit.maturityDay - deposit.startDay);
                progressBar.value = Mathf.Clamp01(progress);
            }

            // 引き出しボタン
            if (withdrawButton != null)
            {
                withdrawButton.onClick.AddListener(OnWithdrawClick);

                if (withdrawButtonText != null)
                {
                    if (isMatured)
                    {
                        withdrawButtonText.text = "引き出す";
                    }
                    else
                    {
                        float penalty = deposit.principal * type.earlyWithdrawalPenalty;
                        withdrawButtonText.text = $"早期解約 (ペナルティ: {penalty:N0}G)";
                    }
                }
            }
        }

        private void OnWithdrawClick()
        {
            OnWithdraw?.Invoke(deposit.id, !isMatured);
        }
    }
}
