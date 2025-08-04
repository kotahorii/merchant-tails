namespace MerchantTails.Data
{
    /// <summary>
    /// 資産の内訳
    /// </summary>
    [System.Serializable]
    public struct AssetBreakdown
    {
        public float cash;
        public float inventoryValue;
        public float bankDeposits;
        public float investments;
        public float totalAssets;
    }

    /// <summary>
    /// 日次資産レポート
    /// </summary>
    [System.Serializable]
    public struct DailyAssetReport
    {
        public int day;
        public float totalAssets;
        public float dailyProfit;
        public float profitPercentage;
        public AssetBreakdown breakdown;
    }
}