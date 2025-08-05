using System.Collections.Generic;
using MerchantTails.Data;

namespace MerchantTails.Core
{
    /// <summary>
    /// MarketSystemのインターフェース
    /// </summary>
    public interface IMarketSystem
    {
        float GetCurrentPrice(ItemType itemType);
        float GetPreviousPrice(ItemType itemType);
        float GetPriceChange(ItemType itemType);
        float GetPriceChangePercentage(ItemType itemType);
        List<float> GetPriceHistory(ItemType itemType);
        void LoadPriceHistory(ItemType itemType, List<float> prices);
        float GetAveragePrice(ItemType itemType, int days);
        float GetMinPrice(ItemType itemType, int days);
        float GetMaxPrice(ItemType itemType, int days);
        float GetVolatility(ItemType itemType, int days);
        
        // MarketSystem静的プロパティを設定するためのメソッド
        void RegisterAsInstance();
    }
}