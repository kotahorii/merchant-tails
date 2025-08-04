using MerchantTails.Data;

namespace MerchantTails.Events
{
    /// <summary>
    /// 資産変動イベント
    /// </summary>
    public class AssetChangedEvent : BaseGameEvent
    {
        public float TotalAssets { get; }
        public AssetBreakdown Breakdown { get; }

        public AssetChangedEvent(float totalAssets, AssetBreakdown breakdown)
        {
            TotalAssets = totalAssets;
            Breakdown = breakdown;
        }
    }

    /// <summary>
    /// 日次資産レポートイベント
    /// </summary>
    public class DailyAssetReportEvent : BaseGameEvent
    {
        public DailyAssetReport Report { get; }

        public DailyAssetReportEvent(DailyAssetReport report)
        {
            Report = report;
        }
    }
}