namespace MerchantTails.Data
{
    /// <summary>
    /// 在庫の保管場所を表すenum
    /// </summary>
    public enum InventoryLocation
    {
        /// <summary>店頭販売用在庫</summary>
        Storefront,
        /// <summary>相場取引用在庫</summary>
        Trading,
    }
}