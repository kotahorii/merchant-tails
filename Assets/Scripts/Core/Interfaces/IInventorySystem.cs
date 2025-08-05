using System.Collections.Generic;
using MerchantTails.Data;

namespace MerchantTails.Core
{
    /// <summary>
    /// InventorySystemのインターフェース
    /// </summary>
    public interface IInventorySystem
    {
        bool AddItem(ItemType itemType, int count, float purchasePrice);
        bool RemoveItem(ItemType itemType, int count, out float averagePrice);
        int GetItemCount(ItemType itemType);
        Dictionary<ItemType, (int count, float condition, float averagePurchasePrice)> GetAllItems();
        void ClearInventory();
        void LoadInventoryItem(ItemType itemType, int count, float condition, float averagePurchasePrice);
        float GetInventoryValue();
        int GetTotalItemCount();
        bool HasItem(ItemType itemType);
        bool CanAddItem(ItemType itemType, int count);
        
        // InventorySystem静的プロパティを設定するためのメソッド
        void RegisterAsInstance();
    }
}