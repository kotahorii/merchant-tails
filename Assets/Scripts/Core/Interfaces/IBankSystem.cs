using System.Collections.Generic;
using MerchantTails.Data;

namespace MerchantTails.Core
{
    /// <summary>
    /// BankSystemのインターフェース (BankSystemはすでにCoreにあるため、このインターフェースはCoreの外部から使用される)
    /// </summary>
    public interface IBankSystem
    {
        float RegularDeposit { get; }
        float TotalDeposits { get; }
        float TotalInterestEarned { get; }
        float CurrentInterestRate { get; }
        bool IsUnlocked { get; }
        
        bool Deposit(float amount);
        bool Withdraw(float amount);
        bool CreateTermDeposit(int typeIndex, float amount);
        bool BreakTermDeposit(string depositId, bool isEarlyWithdrawal = false);
        List<TermDeposit> GetTermDeposits();
        TermDepositType GetTermDepositType(int index);
        float GetTotalDeposits();
        
        // BankSystem静的プロパティを設定するためのメソッド
        void RegisterAsInstance();
    }
}