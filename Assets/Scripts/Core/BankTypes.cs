using System;
using System.Collections.Generic;

namespace MerchantTails.Core
{
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
}