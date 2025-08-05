using System.Collections.Generic;
using MerchantTails.Data;

namespace MerchantTails.Core
{
    /// <summary>
    /// ShopInvestmentSystemのインターフェース
    /// </summary>
    public interface IShopInvestmentSystem
    {
        bool InvestInUpgrade(ShopUpgradeType upgradeType, float amount);
        ShopInvestmentData GetInvestment(ShopUpgradeType upgradeType);
        List<ShopInvestmentData> GetAllInvestments();
        float GetTotalInvestment();
        int GetUpgradeLevel(ShopUpgradeType upgradeType);
        float GetUpgradeEffect(ShopUpgradeType upgradeType);
        bool CanInvest(ShopUpgradeType upgradeType, float amount);
        
        // ShopInvestmentSystem静的プロパティを設定するためのメソッド
        void RegisterAsInstance();
    }

    /// <summary>
    /// 店舗投資データ
    /// </summary>
    public class ShopInvestmentData
    {
        public ShopUpgradeType upgradeType;
        public int level;
        public float totalInvested;
        public float currentEffect;
        public float nextLevelCost;
        public string description;
    }
}