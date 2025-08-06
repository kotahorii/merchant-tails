using System.Collections.Generic;
using MerchantTails.Data;

namespace MerchantTails.Events
{
    /// <summary>
    /// 資産変動イベント
    /// </summary>
    public class AssetChangedEvent
    {
        public float TotalAssets { get; }
        public Dictionary<string, float> Breakdown { get; }

        public AssetChangedEvent(float totalAssets, Dictionary<string, float> breakdown)
        {
            TotalAssets = totalAssets;
            Breakdown = breakdown;
        }
    }

    /// <summary>
    /// 日次資産レポートイベント
    /// </summary>
    public class DailyAssetReportEvent
    {
        public AssetReport Report { get; }

        public DailyAssetReportEvent(AssetReport report)
        {
            Report = report;
        }
    }

    /// <summary>
    /// 資産レポート
    /// </summary>
    public class AssetReport
    {
        public int day;
        public float totalAssets;
        public float cashOnHand;
        public float inventoryValue;
        public float bankDeposits;
        public float investments;
        public float dailyProfit;
        public float profitPercentage;
        public Dictionary<string, float> breakdown;
    }
}
