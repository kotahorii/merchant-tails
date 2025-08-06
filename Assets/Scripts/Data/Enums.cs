namespace MerchantTails.Data
{
    /// <summary>
    /// 商人のランク
    /// プレイヤーの進行度と解放される機能を制御
    /// </summary>
    public enum MerchantRank
    {
        /// <summary>見習い（~1,000G）- 基本機能のみ</summary>
        Apprentice,

        /// <summary>一人前（~5,000G）- 価格予測機能解放</summary>
        Skilled,

        /// <summary>ベテラン（~10,000G）- 高度な分析機能解放</summary>
        Veteran,

        /// <summary>マスター（10,000G+）- 全機能解放</summary>
        Master,
    }

    /// <summary>
    /// 季節
    /// 価格変動と商品の需要に影響
    /// </summary>
    public enum Season
    {
        /// <summary>春 - バランスの取れた需要</summary>
        Spring,

        /// <summary>夏 - ポーションとくだものの需要増</summary>
        Summer,

        /// <summary>秋 - 武器とアクセサリーの需要増</summary>
        Autumn,

        /// <summary>冬 - 魔法書と宝石の需要増</summary>
        Winter,
    }

    /// <summary>
    /// 1日の時間フェーズ
    /// NPCの行動パターンと市場の動きに影響
    /// </summary>
    public enum DayPhase
    {
        /// <summary>朝（6:00-12:00）- 仕入れに最適</summary>
        Morning,

        /// <summary>昼（12:00-18:00）- 活発な取引時間</summary>
        Afternoon,

        /// <summary>夕方（18:00-21:00）- 最後の販売チャンス</summary>
        Evening,

        /// <summary>夜（21:00-6:00）- 市場休止、計画時間</summary>
        Night,
    }

    /// <summary>
    /// 商品アイテムの種類
    /// それぞれ異なる投資特性を持つ
    /// </summary>
    public enum ItemType
    {
        /// <summary>なし - 特定のアイテムを指定しない場合</summary>
        None,

        /// <summary>くだもの - 短期取引、腐敗リスク</summary>
        Fruit,

        /// <summary>ポーション - 成長株、イベント駆動</summary>
        Potion,

        /// <summary>武器 - 優良株、安定価格</summary>
        Weapon,

        /// <summary>アクセサリー - 投機株、トレンド駆動</summary>
        Accessory,

        /// <summary>魔法書 - 債券、高価格安定</summary>
        MagicBook,

        /// <summary>宝石 - ハイリスク、予測困難</summary>
        Gem,
    }

    /// <summary>
    /// 商品の品質ランク
    /// 価格と需要に影響
    /// </summary>
    public enum ItemQuality
    {
        /// <summary>粗悪品 - 低価格、低需要</summary>
        Poor,

        /// <summary>普通品 - 標準価格、標準需要</summary>
        Common,

        /// <summary>良品 - 高価格、高需要</summary>
        Good,

        /// <summary>優良品 - 最高価格、最高需要</summary>
        Excellent,
    }

    /// <summary>
    /// 店舗アップグレードの種類
    /// </summary>
    public enum ShopUpgradeType
    {
        /// <summary>陳列棚 - 在庫上限を増加</summary>
        DisplayShelf,

        /// <summary>看板 - 客足を増加</summary>
        Signboard,

        /// <summary>保管庫 - 品質劣化を抑制</summary>
        Storage,

        /// <summary>装飾 - 価格交渉力を向上</summary>
        Decoration,
    }
}
