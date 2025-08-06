using System;
using UnityEngine;

namespace MerchantTails.Data
{
    /// <summary>
    /// プレイヤーの基本データを管理するScriptableObject
    /// セーブデータの基盤となるクラス
    /// </summary>
    [CreateAssetMenu(fileName = "PlayerData", menuName = "MerchantTails/Player Data")]
    public class PlayerData : ScriptableObject
    {
        [Header("Basic Information")]
        [SerializeField]
        private string playerName = "新米商人";

        [SerializeField]
        private int currentMoney = 1000;

        [SerializeField]
        private MerchantRank currentRank = MerchantRank.Apprentice;

        [Header("Game Progress")]
        [SerializeField]
        private int totalProfit = 0;

        [SerializeField]
        private int successfulTransactions = 0;

        [SerializeField]
        private bool tutorialCompleted = false;

        [Header("Gameplay Stats")]
        [SerializeField]
        private int daysSinceStart = 1;

        [SerializeField]
        private Season currentSeason = Season.Spring;

        // Properties for external access
        public string PlayerName
        {
            get => playerName;
            set => playerName = value;
        }

        public int CurrentMoney
        {
            get => currentMoney;
            private set => currentMoney = Mathf.Max(0, value);
        }

        public MerchantRank CurrentRank
        {
            get => currentRank;
            private set => currentRank = value;
        }

        public int TotalProfit
        {
            get => totalProfit;
            private set => totalProfit = value;
        }

        public int SuccessfulTransactions
        {
            get => successfulTransactions;
            private set => successfulTransactions = value;
        }

        public bool TutorialCompleted
        {
            get => tutorialCompleted;
            set => tutorialCompleted = value;
        }

        public int DaysSinceStart
        {
            get => daysSinceStart;
            set => daysSinceStart = Mathf.Max(1, value);
        }

        public Season CurrentSeason
        {
            get => currentSeason;
            set => currentSeason = value;
        }

        // Events for UI updates
        public event Action<int> OnMoneyChanged;
        public event Action<MerchantRank> OnRankChanged;
        public event Action<int> OnProfitChanged;

        /// <summary>
        /// プレイヤーのお金を変更する
        /// </summary>
        /// <param name="amount">変更量（正の値で増加、負の値で減少）</param>
        /// <returns>取引が成功したかどうか</returns>
        public bool ChangeMoney(int amount)
        {
            int newAmount = currentMoney + amount;

            // お金が足りない場合は失敗
            if (newAmount < 0)
            {
                Debug.LogWarning($"[PlayerData] Insufficient funds. Current: {currentMoney}, Required: {-amount}");
                return false;
            }

            currentMoney = newAmount;
            OnMoneyChanged?.Invoke(currentMoney);

            Debug.Log($"[PlayerData] Money changed by {amount}. New total: {currentMoney}");
            return true;
        }

        /// <summary>
        /// お金を消費する
        /// </summary>
        /// <param name="amount">消費する金額</param>
        /// <returns>消費が成功したかどうか</returns>
        public bool SpendMoney(int amount)
        {
            if (amount < 0)
            {
                Debug.LogWarning($"[PlayerData] SpendMoney called with negative amount: {amount}");
                return false;
            }
            return ChangeMoney(-amount);
        }

        /// <summary>
        /// お金を獲得する
        /// </summary>
        /// <param name="amount">獲得する金額</param>
        public void EarnMoney(int amount)
        {
            if (amount < 0)
            {
                Debug.LogWarning($"[PlayerData] EarnMoney called with negative amount: {amount}");
                return;
            }
            ChangeMoney(amount);
        }

        /// <summary>
        /// 利益を記録し、ランクアップをチェックする
        /// </summary>
        /// <param name="profit">今回の利益</param>
        public void RecordProfit(int profit)
        {
            totalProfit += profit;
            OnProfitChanged?.Invoke(totalProfit);

            if (profit > 0)
            {
                successfulTransactions++;
                CheckRankUp();
            }

            Debug.Log($"[PlayerData] Profit recorded: {profit}. Total profit: {totalProfit}");
        }

        /// <summary>
        /// ランクアップの条件をチェックし、必要に応じてランクアップする
        /// </summary>
        private void CheckRankUp()
        {
            MerchantRank newRank = CalculateRankFromAssets();

            if (newRank != currentRank)
            {
                currentRank = newRank;
                OnRankChanged?.Invoke(currentRank);
                Debug.Log($"[PlayerData] Rank up! New rank: {currentRank}");
            }
        }

        /// <summary>
        /// ランクを手動で設定する
        /// </summary>
        public void SetRank(MerchantRank newRank)
        {
            if (currentRank != newRank)
            {
                currentRank = newRank;
                OnRankChanged?.Invoke(currentRank);
                Debug.Log($"[PlayerData] Rank changed to: {currentRank}");
            }
        }

        /// <summary>
        /// 総資産（現金 + 商品価値）から商人ランクを計算する
        /// </summary>
        /// <returns>現在の総資産に応じたランク</returns>
        private MerchantRank CalculateRankFromAssets()
        {
            // TODO: 在庫価値も含めた総資産計算が必要
            int totalAssets = currentMoney;

            if (totalAssets >= 10000)
                return MerchantRank.Master;
            if (totalAssets >= 5000)
                return MerchantRank.Veteran;
            if (totalAssets >= 1000)
                return MerchantRank.Skilled;
            return MerchantRank.Apprentice;
        }

        /// <summary>
        /// プレイヤーデータを初期状態にリセット
        /// </summary>
        public void ResetToDefault()
        {
            playerName = "新米商人";
            currentMoney = 1000;
            currentRank = MerchantRank.Apprentice;
            totalProfit = 0;
            successfulTransactions = 0;
            tutorialCompleted = false;
            daysSinceStart = 1;
            currentSeason = Season.Spring;

            Debug.Log("[PlayerData] Reset to default values");
        }

        /// <summary>
        /// デバッグ用：現在の状態をログ出力
        /// </summary>
        public void LogCurrentState()
        {
            Debug.Log(
                $"[PlayerData] Player: {playerName}, Money: {currentMoney}, Rank: {currentRank}, "
                    + $"Profit: {totalProfit}, Transactions: {successfulTransactions}, Day: {daysSinceStart}"
            );
        }
    }
}
