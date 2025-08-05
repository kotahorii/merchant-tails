using System.Collections.Generic;

namespace MerchantTails.Core
{
    /// <summary>
    /// MerchantInvestmentSystemのインターフェース
    /// </summary>
    public interface IMerchantInvestmentSystem
    {
        bool InvestInMerchant(string merchantId, float amount);
        MerchantInvestmentData GetInvestment(string merchantId);
        List<MerchantData> GetAvailableMerchants();
        List<MerchantInvestmentData> GetActiveInvestments();
        float GetTotalDividends();
        bool CanInvest(string merchantId, float amount);
        void ProcessDividends();
        
        // MerchantInvestmentSystem静的プロパティを設定するためのメソッド
        void RegisterAsInstance();
    }

    /// <summary>
    /// 商人データ
    /// </summary>
    public class MerchantData
    {
        public string id;
        public string name;
        public string description;
        public float minimumInvestment;
        public float expectedReturn;
        public float riskLevel;
        public string specialization;
    }

    /// <summary>
    /// 商人投資データ
    /// </summary>
    public class MerchantInvestmentData
    {
        public string merchantId;
        public float totalInvested;
        public float totalDividends;
        public int lastInvestmentDay;
        public int lastDividendDay;
        public bool isActive;
    }
}