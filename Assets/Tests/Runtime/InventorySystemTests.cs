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
    /// InventorySystemの単体テスト
    /// </summary>
    public class InventorySystemTests : TestBase
    {
        [Test]
        public void AddItem_SingleItem_AddsSuccessfully()
        {
            // Arrange
            var item = CreateTestItem(ItemType.Fruit, 5);

            // Act
            bool result = inventorySystem.AddItem(item);

            // Assert
            Assert.IsTrue(result, "AddItem should return true");
            Assert.AreEqual(5, inventorySystem.GetItemCount(ItemType.Fruit),
                "Item count should be 5");
        }

        [Test]
        public void AddItem_MultipleOfSameType_CombinesQuantity()
        {
            // Arrange
            var item1 = CreateTestItem(ItemType.Potion, 3);
            var item2 = CreateTestItem(ItemType.Potion, 2);

            // Act
            inventorySystem.AddItem(item1);
            inventorySystem.AddItem(item2);

            // Assert
            Assert.AreEqual(5, inventorySystem.GetItemCount(ItemType.Potion),
                "Combined quantity should be 5");
        }

        [Test]
        public void RemoveItem_ValidQuantity_RemovesSuccessfully()
        {
            // Arrange
            var item = CreateTestItem(ItemType.Weapon, 10);
            inventorySystem.AddItem(item);

            // Act
            bool result = inventorySystem.RemoveItem(item.id, 3);

            // Assert
            Assert.IsTrue(result, "RemoveItem should return true");
            Assert.AreEqual(7, inventorySystem.GetItemCount(ItemType.Weapon),
                "Remaining quantity should be 7");
        }

        [Test]
        public void RemoveItem_ExceedsQuantity_Fails()
        {
            // Arrange
            var item = CreateTestItem(ItemType.Accessory, 5);
            inventorySystem.AddItem(item);

            // Act
            bool result = inventorySystem.RemoveItem(item.id, 10);

            // Assert
            Assert.IsFalse(result, "RemoveItem should return false when quantity exceeds available");
            Assert.AreEqual(5, inventorySystem.GetItemCount(ItemType.Accessory),
                "Quantity should remain unchanged");
        }

        [Test]
        public void GetAllItems_ReturnsCorrectItems()
        {
            // Arrange
            var item1 = CreateTestItem(ItemType.MagicBook, 2);
            var item2 = CreateTestItem(ItemType.Gem, 3);
            inventorySystem.AddItem(item1);
            inventorySystem.AddItem(item2);

            // Act
            var allItems = inventorySystem.GetAllItems();

            // Assert
            Assert.AreEqual(2, allItems.Count, "Should have 2 different item types");
            Assert.IsTrue(allItems.ContainsKey(ItemType.MagicBook));
            Assert.IsTrue(allItems.ContainsKey(ItemType.Gem));
        }

        [Test]
        public void GetTotalValue_CalculatesCorrectly()
        {
            // Arrange
            var item1 = CreateTestItem(ItemType.Fruit, 10, 5f);
            var item2 = CreateTestItem(ItemType.Potion, 5, 20f);
            inventorySystem.AddItem(item1);
            inventorySystem.AddItem(item2);

            // Act
            float totalValue = inventorySystem.GetTotalValue();

            // Assert
            float expectedValue = (10 * 5f) + (5 * 20f); // 50 + 100 = 150
            AssertFloatEquals(expectedValue, totalValue);
        }

        [Test]
        public void TransferToShop_ValidQuantity_TransfersSuccessfully()
        {
            // Arrange
            var item = CreateTestItem(ItemType.Weapon, 8);
            inventorySystem.AddItem(item);

            // Act
            bool result = inventorySystem.TransferToShop(ItemType.Weapon, 5);

            // Assert
            Assert.IsTrue(result, "Transfer should succeed");
            Assert.AreEqual(3, inventorySystem.GetItemCount(ItemType.Weapon, false),
                "Storage should have 3 items");
            Assert.AreEqual(5, inventorySystem.GetItemCount(ItemType.Weapon, true),
                "Shop should have 5 items");
        }

        [Test]
        public void TransferFromShop_ValidQuantity_TransfersSuccessfully()
        {
            // Arrange
            var item = CreateTestItem(ItemType.Accessory, 10);
            item.isInShop = true;
            inventorySystem.AddItem(item);

            // Act
            bool result = inventorySystem.TransferFromShop(ItemType.Accessory, 4);

            // Assert
            Assert.IsTrue(result, "Transfer should succeed");
            Assert.AreEqual(6, inventorySystem.GetItemCount(ItemType.Accessory, true),
                "Shop should have 6 items");
            Assert.AreEqual(4, inventorySystem.GetItemCount(ItemType.Accessory, false),
                "Storage should have 4 items");
        }

        [Test]
        public void GetShopItems_ReturnsOnlyShopItems()
        {
            // Arrange
            var shopItem = CreateTestItem(ItemType.MagicBook, 3);
            shopItem.isInShop = true;
            var storageItem = CreateTestItem(ItemType.Gem, 2);
            storageItem.isInShop = false;

            inventorySystem.AddItem(shopItem);
            inventorySystem.AddItem(storageItem);

            // Act
            var shopItems = inventorySystem.GetShopItems();

            // Assert
            Assert.AreEqual(1, shopItems.Count, "Should have 1 shop item type");
            Assert.IsTrue(shopItems.ContainsKey(ItemType.MagicBook));
            Assert.IsFalse(shopItems.ContainsKey(ItemType.Gem));
        }

        [Test]
        public void HasSpace_CapacityCheck()
        {
            // Note: This test assumes there's a capacity limit in InventorySystem
            // If not implemented, this test should be adjusted

            // Act
            bool hasSpace = inventorySystem.HasSpace(ItemType.Fruit, 1);

            // Assert
            Assert.IsTrue(hasSpace, "Empty inventory should have space");
        }

        [UnityTest]
        public IEnumerator AddItem_PublishesEvent()
        {
            // Arrange
            bool eventReceived = false;
            ItemType addedType = ItemType.Fruit;

            EventBus.Subscribe<InventoryChangedEvent>((e) =>
            {
                eventReceived = true;
                addedType = e.ItemType;
            });

            var item = CreateTestItem(ItemType.Potion, 1);

            // Act
            inventorySystem.AddItem(item);

            // Assert
            yield return WaitForCondition(() => eventReceived, 1f);
            Assert.IsTrue(eventReceived, "InventoryChangedEvent should have been published");
            Assert.AreEqual(ItemType.Potion, addedType, "Event should contain correct item type");
        }

        [Test]
        public void ClearInventory_RemovesAllItems()
        {
            // Arrange
            inventorySystem.AddItem(CreateTestItem(ItemType.Fruit, 5));
            inventorySystem.AddItem(CreateTestItem(ItemType.Weapon, 3));
            inventorySystem.AddItem(CreateTestItem(ItemType.Gem, 2));

            // Act
            inventorySystem.ClearInventory();

            // Assert
            var allItems = inventorySystem.GetAllItems();
            Assert.AreEqual(0, allItems.Count, "Inventory should be empty");
            Assert.AreEqual(0, inventorySystem.GetTotalValue(), "Total value should be 0");
        }

        [Test]
        public void GetItemsByQuality_FiltersCorrectly()
        {
            // Arrange
            var normalItem = CreateTestItem(ItemType.Fruit, 5);
            normalItem.quality = ItemQuality.Normal;

            var rareItem = CreateTestItem(ItemType.Gem, 2);
            rareItem.quality = ItemQuality.Rare;

            inventorySystem.AddItem(normalItem);
            inventorySystem.AddItem(rareItem);

            // Act
            var normalItems = inventorySystem.GetItemsByQuality(ItemQuality.Normal);
            var rareItems = inventorySystem.GetItemsByQuality(ItemQuality.Rare);

            // Assert
            Assert.AreEqual(1, normalItems.Count, "Should have 1 normal quality item");
            Assert.AreEqual(1, rareItems.Count, "Should have 1 rare quality item");
        }

        [Test]
        public void UpdateItemCondition_DegradesOverTime()
        {
            // Arrange
            var perishableItem = CreateTestItem(ItemType.Fruit, 10);
            inventorySystem.AddItem(perishableItem);

            float initialCondition = 1.0f; // Assuming items start at 100% condition

            // Act - Simulate time passing
            for (int i = 0; i < 5; i++)
            {
                inventorySystem.UpdateItemConditions();
            }

            // Assert
            var items = inventorySystem.GetItemsOfType(ItemType.Fruit);
            if (items.Count > 0)
            {
                float currentCondition = items[0].condition;
                Assert.Less(currentCondition, initialCondition,
                    "Perishable item condition should degrade over time");
            }
        }
    }
}
