using System.Collections.Generic;

namespace MerchantTails.Core
{
    /// <summary>
    /// BankSystemのインターフェース (BankSystemはすでにCoreにあるため、このインターフェースはCoreの外部から使用される)
    /// </summary>
    public interface IBankSystem
    {
        float TotalInterestEarned { get; }
        
        bool CreateDeposit(float amount, int termDays);
        bool WithdrawDeposit(string depositId);
        List<Deposit> GetAllDeposits();
        Deposit GetDeposit(string depositId);
        float GetTotalDeposits();
        float GetAvailableBalance();
        void ProcessMaturedDeposits();
        
        // BankSystem静的プロパティを設定するためのメソッド
        void RegisterAsInstance();
    }
}