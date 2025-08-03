using System;
using MerchantTails.Data;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

namespace MerchantTails.UI
{
    /// <summary>
    /// 取引履歴の個別アイテムUI
    /// </summary>
    public class TransactionHistoryItem : MonoBehaviour
    {
        [Header("UI Elements")]
        [SerializeField]
        private TextMeshProUGUI timeText;

        [SerializeField]
        private TextMeshProUGUI itemNameText;

        [SerializeField]
        private TextMeshProUGUI quantityText;

        [SerializeField]
        private TextMeshProUGUI priceText;

        [SerializeField]
        private TextMeshProUGUI totalText;

        [SerializeField]
        private Image transactionTypeIcon;

        [SerializeField]
        private Image itemIcon;

        [Header("Visual Settings")]
        [SerializeField]
        private Sprite buyIcon;

        [SerializeField]
        private Sprite sellIcon;

        [SerializeField]
        private Color buyColor = new Color(1f, 0.8f, 0.8f);

        [SerializeField]
        private Color sellColor = new Color(0.8f, 1f, 0.8f);

        private TransactionRecord transaction;

        public void Setup(TransactionRecord record)
        {
            transaction = record;
            UpdateDisplay();
        }

        private void UpdateDisplay()
        {
            if (transaction == null)
                return;

            // 時刻表示
            if (timeText != null)
            {
                timeText.text = FormatTime(transaction.timestamp);
            }

            // アイテム名
            if (itemNameText != null)
            {
                itemNameText.text = GetItemName(transaction.itemType);
            }

            // 数量
            if (quantityText != null)
            {
                quantityText.text = $"×{transaction.quantity}";
            }

            // 単価
            if (priceText != null)
            {
                priceText.text = $"{transaction.unitPrice:N0}G";
            }

            // 合計
            if (totalText != null)
            {
                totalText.text = $"{transaction.totalPrice:N0}G";
                totalText.color = transaction.isPurchase ? buyColor : sellColor;
            }

            // 取引タイプアイコン
            if (transactionTypeIcon != null)
            {
                transactionTypeIcon.sprite = transaction.isPurchase ? buyIcon : sellIcon;
                transactionTypeIcon.color = transaction.isPurchase ? buyColor : sellColor;
            }

            // アイテムアイコン
            if (itemIcon != null)
            {
                // TODO: アイテムタイプに応じたアイコンを設定
                itemIcon.color = GetItemColor(transaction.itemType);
            }

            // 背景色
            var image = GetComponent<Image>();
            if (image != null)
            {
                image.color = transaction.isPurchase
                    ? new Color(1f, 0.95f, 0.95f, 0.3f)
                    : new Color(0.95f, 1f, 0.95f, 0.3f);
            }
        }

        private string FormatTime(DateTime timestamp)
        {
            var now = DateTime.Now;
            var diff = now - timestamp;

            if (diff.TotalMinutes < 1)
                return "たった今";
            else if (diff.TotalMinutes < 60)
                return $"{(int)diff.TotalMinutes}分前";
            else if (diff.TotalHours < 24)
                return $"{(int)diff.TotalHours}時間前";
            else
                return $"{(int)diff.TotalDays}日前";
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
                _ => "不明",
            };
        }

        private Color GetItemColor(ItemType itemType)
        {
            return itemType switch
            {
                ItemType.Fruit => new Color(1f, 0.6f, 0.3f), // オレンジ
                ItemType.Potion => new Color(0.3f, 0.8f, 1f), // 水色
                ItemType.Weapon => new Color(0.7f, 0.7f, 0.7f), // グレー
                ItemType.Accessory => new Color(1f, 0.8f, 0.3f), // ゴールド
                ItemType.MagicBook => new Color(0.6f, 0.3f, 1f), // 紫
                ItemType.Gem => new Color(1f, 0.3f, 0.6f), // ピンク
                _ => Color.white,
            };
        }
    }
}
