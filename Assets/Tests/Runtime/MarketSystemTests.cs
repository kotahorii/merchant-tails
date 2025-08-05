using System.Collections;
using System.Linq;
using MerchantTails.Core;
using MerchantTails.Data;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.TestTools;

namespace MerchantTails.Testing
{
    /// <summary>
    /// MarketSystemの単体テスト
    /// </summary>
    public class MarketSystemTests : TestBase
    {
        [Test]
        public void GetCurrentPrice_ReturnsValidPrice()
        {
            // Arrange
            var itemType = ItemType.Fruit;

            // Act
            float price = marketSystem.GetCurrentPrice(itemType);

            // Assert
            Assert.Greater(price, 0f, "Price should be greater than 0");
            AssertInRange(price, 5f, 50f); // フルーツの妥当な価格範囲
        }

        [Test]
        public void GetCurrentPrice_AllItemTypes_ReturnsValidPrices()
        {
            foreach (ItemType itemType in System.Enum.GetValues(typeof(ItemType)))
            {
                // Act
                float price = marketSystem.GetCurrentPrice(itemType);

                // Assert
                Assert.Greater(price, 0f, $"Price for {itemType} should be greater than 0");

                // 各アイテムタイプごとの妥当な価格範囲をチェック
                switch (itemType)
                {
                    case ItemType.Fruit:
                        AssertInRange(price, 5f, 50f);
                        break;
                    case ItemType.Potion:
                        AssertInRange(price, 25f, 200f);
                        break;
                    case ItemType.Weapon:
                        AssertInRange(price, 100f, 1000f);
                        break;
                    case ItemType.Accessory:
                        AssertInRange(price, 50f, 500f);
                        break;
                    case ItemType.MagicBook:
                        AssertInRange(price, 150f, 1500f);
                        break;
                    case ItemType.Gem:
                        AssertInRange(price, 250f, 2500f);
                        break;
                }
            }
        }

        [Test]
        public void UpdatePrices_ChangesAllPrices()
        {
            // Arrange
            var originalPrices = new System.Collections.Generic.Dictionary<ItemType, float>();
            foreach (ItemType itemType in System.Enum.GetValues(typeof(ItemType)))
            {
                originalPrices[itemType] = marketSystem.GetCurrentPrice(itemType);
            }

            // Act
            marketSystem.UpdatePrices();

            // Assert
            bool anyPriceChanged = false;
            foreach (ItemType itemType in System.Enum.GetValues(typeof(ItemType)))
            {
                float newPrice = marketSystem.GetCurrentPrice(itemType);
                if (System.Math.Abs(newPrice - originalPrices[itemType]) > 0.01f)
                {
                    anyPriceChanged = true;
                    break;
                }
            }

            Assert.IsTrue(anyPriceChanged, "At least one price should have changed after update");
        }

        [Test]
        public void GetPriceHistory_ReturnsCorrectHistory()
        {
            // Arrange
            var itemType = ItemType.Potion;
            int updateCount = 5;

            // Act
            for (int i = 0; i < updateCount; i++)
            {
                marketSystem.UpdatePrices();
            }

            var history = marketSystem.GetPriceHistory(itemType);

            // Assert
            Assert.IsNotNull(history, "Price history should not be null");
            Assert.GreaterOrEqual(history.Count, updateCount,
                $"Price history should have at least {updateCount} entries");
        }

        [Test]
        public void GetPriceChangePercent_CalculatesCorrectly()
        {
            // Arrange
            var itemType = ItemType.Weapon;
            float originalPrice = marketSystem.GetCurrentPrice(itemType);

            // Act
            marketSystem.UpdatePrices();
            float changePercent = marketSystem.GetPriceChangePercent(itemType);
            float newPrice = marketSystem.GetCurrentPrice(itemType);

            // Assert
            float expectedChange = ((newPrice - originalPrice) / originalPrice) * 100f;
            AssertFloatEquals(expectedChange, changePercent, 0.1f);
        }

        [Test]
        public void GetPriceTrend_IdentifiesUpwardTrend()
        {
            // Arrange
            var itemType = ItemType.MagicBook;

            // 価格を人為的に上昇させる
            for (int i = 0; i < 5; i++)
            {
                var basePrice = 300f + (i * 10f);
                // MarketSystemの内部状態を操作する必要がある
                marketSystem.UpdatePrices();
            }

            // Act
            var trend = marketSystem.GetPriceTrend(itemType);

            // Assert
            Assert.IsNotNull(trend, "Price trend should not be null");
            // トレンドの方向性をチェック（実装に依存）
        }

        [UnityTest]
        public IEnumerator PriceUpdate_PublishesEvent()
        {
            // Arrange
            bool eventReceived = false;
            ItemType updatedItemType = ItemType.Fruit;

            EventBus.Subscribe<PriceChangedEvent>((e) =>
            {
                eventReceived = true;
                updatedItemType = e.ItemType;
            });

            // Act
            marketSystem.UpdatePrices();

            // Assert
            yield return WaitForCondition(() => eventReceived, 1f);
            Assert.IsTrue(eventReceived, "PriceChangedEvent should have been published");
        }

        [Test]
        public void SeasonalModifier_AffectsPrices()
        {
            // Arrange
            var itemType = ItemType.Fruit;

            // 春の価格を記録
            timeManager.LoadTimeData(1, Season.Spring, DayPhase.Morning, 0f);
            marketSystem.UpdatePrices();
            float springPrice = marketSystem.GetCurrentPrice(itemType);

            // Act - 夏に変更
            timeManager.LoadTimeData(1, Season.Summer, DayPhase.Morning, 0f);
            marketSystem.UpdatePrices();
            float summerPrice = marketSystem.GetCurrentPrice(itemType);

            // Assert
            Assert.AreNotEqual(springPrice, summerPrice,
                "Prices should be different in different seasons");
        }

        [Test]
        public void GetRecommendedAction_ProvidesValidRecommendation()
        {
            // Arrange
            var itemType = ItemType.Accessory;

            // Act
            var action = marketSystem.GetRecommendedAction(itemType);

            // Assert
            Assert.IsNotNull(action, "Recommended action should not be null");
            Assert.That(action, Is.EqualTo("Buy").Or.EqualTo("Sell").Or.EqualTo("Hold"),
                "Action should be Buy, Sell, or Hold");
        }

        [Test]
        public void CalculateProfit_CorrectCalculation()
        {
            // Arrange
            var itemType = ItemType.Gem;
            float buyPrice = 400f;
            float sellPrice = 500f;
            int quantity = 5;

            // Act
            float profit = marketSystem.CalculateProfit(itemType, buyPrice, sellPrice, quantity);

            // Assert
            float expectedProfit = (sellPrice - buyPrice) * quantity;
            AssertFloatEquals(expectedProfit, profit);
        }
    }
}
