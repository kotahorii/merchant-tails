namespace MerchantTails.Data
{
    /// <summary>
    /// ゲーム内で解放可能な機能
    /// </summary>
    public enum GameFeature
    {
        // 基本機能
        BasicTrading,
        SimpleInventory,

        // 一人前で解放
        PricePrediction,
        MarketTrends,
        BankAccount,
        SeasonalInfo,
        BasicPriceHistory,

        // ベテランで解放
        AdvancedAnalytics,
        EventPrediction,
        ShopInvestment,
        AutoPricing,
        AdvancedInventory,

        // マスターで解放
        MerchantNetwork,
        MarketManipulation,
        ExclusiveDeals,
        FullAutomation,
    }
}
