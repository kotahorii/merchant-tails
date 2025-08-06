namespace MerchantTails.Data
{
    /// <summary>
    /// UI画面の種類を定義する列挙型
    /// UIManagerで画面管理に使用
    /// </summary>
    public enum UIType
    {
        /// <summary>メインメニュー画面</summary>
        MainMenu,

        /// <summary>ゲーム内HUD</summary>
        GameHUD,

        /// <summary>ショップ管理画面</summary>
        ShopManagement,

        /// <summary>マーケット分析画面</summary>
        MarketAnalysis,

        /// <summary>インベントリ画面</summary>
        Inventory,

        /// <summary>設定画面</summary>
        Settings,

        /// <summary>チュートリアル画面</summary>
        Tutorial,

        /// <summary>商人手帳画面</summary>
        MerchantJournal,

        /// <summary>取引履歴画面</summary>
        TransactionHistory,

        /// <summary>価格チャート画面</summary>
        PriceChart,

        /// <summary>確認ダイアログ</summary>
        ConfirmDialog,

        /// <summary>アイテム詳細モーダル</summary>
        ItemDetail,

        /// <summary>取引確認モーダル</summary>
        TradeConfirmation,

        /// <summary>ローディング画面</summary>
        Loading,

        /// <summary>ゲームオーバー画面</summary>
        GameOver,

        /// <summary>ポーズメニュー</summary>
        PauseMenu,

        /// <summary>確認画面</summary>
        Confirmation,

        /// <summary>クレジット画面</summary>
        Credits,

        /// <summary>該当なし</summary>
        None,
    }

    /// <summary>
    /// UI表示モード
    /// パネルの表示方法を制御
    /// </summary>
    public enum UIDisplayMode
    {
        /// <summary>通常表示（フルスクリーン）</summary>
        Normal,

        /// <summary>モーダル表示（オーバーレイ）</summary>
        Modal,

        /// <summary>ポップアップ表示</summary>
        Popup,

        /// <summary>サイドパネル表示</summary>
        SidePanel,

        /// <summary>通知表示</summary>
        Notification,
    }

    /// <summary>
    /// UI遷移の種類
    /// 画面切り替え時のアニメーション制御
    /// </summary>
    public enum UITransitionType
    {
        /// <summary>フェード</summary>
        Fade,

        /// <summary>スライド（左から）</summary>
        SlideLeft,

        /// <summary>スライド（右から）</summary>
        SlideRight,

        /// <summary>スライド（上から）</summary>
        SlideUp,

        /// <summary>スライド（下から）</summary>
        SlideDown,

        /// <summary>スケール</summary>
        Scale,

        /// <summary>即座に切り替え</summary>
        Instant,
    }

    /// <summary>
    /// UI状態
    /// パネルの現在状態を表す
    /// </summary>
    public enum UIState
    {
        /// <summary>非表示</summary>
        Hidden,

        /// <summary>表示中</summary>
        Visible,

        /// <summary>表示アニメーション中</summary>
        Showing,

        /// <summary>非表示アニメーション中</summary>
        Hiding,

        /// <summary>無効化状態</summary>
        Disabled,
    }

    /// <summary>
    /// UI優先度
    /// 複数のUIが重なった時の表示順制御
    /// </summary>
    public enum UIPriority
    {
        /// <summary>最低優先度（背景UI）</summary>
        Background = 0,

        /// <summary>低優先度（通常UI）</summary>
        Low = 100,

        /// <summary>標準優先度（メインUI）</summary>
        Normal = 200,

        /// <summary>高優先度（重要UI）</summary>
        High = 300,

        /// <summary>最高優先度（システムUI、エラーダイアログ等）</summary>
        Critical = 400,
    }
}
