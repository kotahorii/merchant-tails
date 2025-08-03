using UnityEngine;
using UnityEngine.UI;
using TMPro;
using MerchantTails.Data;

namespace MerchantTails.UI
{
    /// <summary>
    /// アイテム別統計表示UI
    /// </summary>
    public class StatisticsItem : MonoBehaviour
    {
        [Header("UI Elements")]
        [SerializeField] private Image itemIcon;
        [SerializeField] private TextMeshProUGUI itemNameText;
        [SerializeField] private TextMeshProUGUI purchasedText;
        [SerializeField] private TextMeshProUGUI soldText;
        [SerializeField] private TextMeshProUGUI netProfitText;
        [SerializeField] private TextMeshProUGUI avgBuyPriceText;
        [SerializeField] private TextMeshProUGUI avgSellPriceText;
        [SerializeField] private Image profitIndicator;
        
        [Header("Visual Settings")]
        [SerializeField] private Color profitColor = Color.green;
        [SerializeField] private Color lossColor = Color.red;
        [SerializeField] private Color neutralColor = Color.gray;
        
        private ItemType itemType;
        private ItemStatistics statistics;
        
        public void Setup(ItemType type, ItemStatistics stats)
        {
            itemType = type;
            statistics = stats;
            UpdateDisplay();
        }
        
        private void UpdateDisplay()
        {
            if (statistics == null) return;
            
            // アイテム名
            if (itemNameText != null)
            {
                itemNameText.text = GetItemName(itemType);
            }
            
            // アイコン
            if (itemIcon != null)
            {
                itemIcon.color = GetItemColor(itemType);
            }
            
            // 購入数
            if (purchasedText != null)
            {
                purchasedText.text = $"仕入: {statistics.totalPurchased}個";
            }
            
            // 販売数
            if (soldText != null)
            {
                soldText.text = $"販売: {statistics.totalSold}個";
            }
            
            // 純利益
            if (netProfitText != null)
            {
                float profit = statistics.NetProfit;
                netProfitText.text = $"{profit:+#,0;-#,0;0}G";
                netProfitText.color = profit >= 0 ? profitColor : lossColor;
            }
            
            // 平均仕入価格
            if (avgBuyPriceText != null)
            {
                avgBuyPriceText.text = $"仕入: {statistics.AverageBuyPrice:N0}G";
            }
            
            // 平均販売価格
            if (avgSellPriceText != null)
            {
                avgSellPriceText.text = $"販売: {statistics.AverageSellPrice:N0}G";
            }
            
            // 利益インジケーター
            UpdateProfitIndicator();
        }
        
        private void UpdateProfitIndicator()
        {
            if (profitIndicator == null) return;
            
            float profit = statistics.NetProfit;
            
            if (profit > 0)
            {
                profitIndicator.color = profitColor;
                profitIndicator.fillAmount = 1f;
            }
            else if (profit < 0)
            {
                profitIndicator.color = lossColor;
                profitIndicator.fillAmount = 1f;
            }
            else
            {
                profitIndicator.color = neutralColor;
                profitIndicator.fillAmount = 0.5f;
            }
        }
        
        private string GetItemName(ItemType itemType)
        {
            return itemType switch
            {
                ItemType.Fruit => "くだもの",
                ItemType.Potion => "ポーション",
                ItemType.Weapon => "武器",
                ItemType.Accessory => "アクセサリー",
                ItemType.MagicBook => "魔法書",
                ItemType.Gem => "宝石",
                _ => "不明"
            };
        }
        
        private Color GetItemColor(ItemType itemType)
        {
            return itemType switch
            {
                ItemType.Fruit => new Color(1f, 0.6f, 0.3f),      // オレンジ
                ItemType.Potion => new Color(0.3f, 0.8f, 1f),     // 水色
                ItemType.Weapon => new Color(0.7f, 0.7f, 0.7f),   // グレー
                ItemType.Accessory => new Color(1f, 0.8f, 0.3f),  // ゴールド
                ItemType.MagicBook => new Color(0.6f, 0.3f, 1f),  // 紫
                ItemType.Gem => new Color(1f, 0.3f, 0.6f),        // ピンク
                _ => Color.white
            };
        }
    }
}