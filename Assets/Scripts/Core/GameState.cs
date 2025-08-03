namespace MerchantTails.Core
{
    /// <summary>
    /// ゲームの状態を表すenum
    /// システム全体の状態管理に使用
    /// </summary>
    public enum GameState
    {
        /// <summary>メインメニュー画面</summary>
        MainMenu,

        /// <summary>チュートリアル画面</summary>
        Tutorial,

        /// <summary>ショッピング（商品購入）画面</summary>
        Shopping,

        /// <summary>店舗管理（商品販売）画面</summary>
        StoreManagement,

        /// <summary>市場確認（相場チェック）画面</summary>
        MarketView,

        /// <summary>一時停止状態</summary>
        Paused
    }
}
