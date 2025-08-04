using System.Collections;
using System.IO;
using System.Linq;
using MerchantTails.Core;
using MerchantTails.Data;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.TestTools;

namespace MerchantTails.Testing
{
    /// <summary>
    /// セーブ/ロード整合性テスト
    /// データの永続化と復元が正しく動作することを確認
    /// </summary>
    public class SaveLoadTests : TestBase
    {
        private SaveSystem saveSystem;
        private BankSystem bankSystem;
        private AchievementSystem achievementSystem;
        private FeatureUnlockSystem featureUnlockSystem;
        private ShopInvestmentSystem shopInvestmentSystem;
        private MerchantInvestmentSystem merchantInvestmentSystem;

        public override void Setup()
        {
            base.Setup();

            saveSystem = testGameObject.AddComponent<SaveSystem>();
            bankSystem = testGameObject.AddComponent<BankSystem>();
            achievementSystem = testGameObject.AddComponent<AchievementSystem>();
            featureUnlockSystem = testGameObject.AddComponent<FeatureUnlockSystem>();
            shopInvestmentSystem = testGameObject.AddComponent<ShopInvestmentSystem>();
            merchantInvestmentSystem = testGameObject.AddComponent<MerchantInvestmentSystem>();
        }

        public override void Teardown()
        {
            // テスト用のセーブファイルを削除
            saveSystem?.DeleteAllSaves();
            base.Teardown();
        }

        [UnityTest]
        public IEnumerator SaveAndLoad_PlayerData()
        {
            // Arrange
            string testName = "TestMerchant";
            int testMoney = 54321;
            var testRank = MerchantRank.Veteran;
            int testTransactions = 150;
            float testProfit = 12345.67f;

            gameManager.PlayerData.SetPlayerName(testName);
            gameManager.PlayerData.SetMoney(testMoney);
            gameManager.PlayerData.SetRank(testRank);
            gameManager.PlayerData.TotalTransactions = testTransactions;
            gameManager.PlayerData.TotalProfit = testProfit;

            // Act - Use async methods
            var saveTask = saveSystem.SaveAsync(0);
            yield return new WaitUntil(() => saveTask.IsCompleted);
            Assert.IsTrue(saveTask.Result, "Save should succeed");

            // Reset data
            gameManager.PlayerData.SetPlayerName("Default");
            gameManager.PlayerData.SetMoney(0);
            gameManager.PlayerData.SetRank(MerchantRank.Apprentice);
            gameManager.PlayerData.TotalTransactions = 0;
            gameManager.PlayerData.TotalProfit = 0;

            var loadTask = saveSystem.LoadAsync(0);
            yield return new WaitUntil(() => loadTask.IsCompleted);

            // Assert
            Assert.IsTrue(loadTask.Result, "Load should succeed");
            Assert.AreEqual(testName, gameManager.PlayerData.PlayerName);
            Assert.AreEqual(testMoney, gameManager.PlayerData.CurrentMoney);
            Assert.AreEqual(testRank, gameManager.PlayerData.CurrentRank);
            Assert.AreEqual(testTransactions, gameManager.PlayerData.TotalTransactions);
            AssertFloatEquals(testProfit, gameManager.PlayerData.TotalProfit);
        }

        [Test]
        public void SaveAndLoad_TimeData()
        {
            // Arrange
            int testDay = 123;
            var testSeason = Season.Winter;
            var testPhase = DayPhase.Evening;
            float testProgress = 0.75f;

            timeManager.LoadTimeData(testDay, testSeason, testPhase, testProgress);

            // Act
            saveSystem.Save(0);
            timeManager.LoadTimeData(1, Season.Spring, DayPhase.Morning, 0f);
            saveSystem.Load(0);

            // Assert
            Assert.AreEqual(testDay, timeManager.CurrentDay);
            Assert.AreEqual(testSeason, timeManager.CurrentSeason);
            Assert.AreEqual(testPhase, timeManager.CurrentPhase);
            AssertFloatEquals(testProgress, timeManager.DayProgress);
        }

        [Test]
        public void SaveAndLoad_ComplexInventory()
        {
            // Arrange
            var items = new[]
            {
                CreateTestItem(ItemType.Fruit, 25, 15f),
                CreateTestItem(ItemType.Potion, 10, 75f),
                CreateTestItem(ItemType.Weapon, 3, 350f),
                CreateTestItem(ItemType.Accessory, 7, 150f),
                CreateTestItem(ItemType.MagicBook, 2, 500f),
                CreateTestItem(ItemType.Gem, 5, 1000f),
            };

            foreach (var item in items)
            {
                item.quality = (ItemQuality)Random.Range(0, 3);
                item.condition = Random.Range(0.5f, 1f);
                inventorySystem.AddItem(item);
            }

            // Some items in shop
            inventorySystem.TransferToShop(ItemType.Fruit, 10);
            inventorySystem.TransferToShop(ItemType.Potion, 5);

            // Act
            saveSystem.Save(0);
            inventorySystem.ClearInventory();
            saveSystem.Load(0);

            // Assert
            var allItems = inventorySystem.GetAllItems();
            Assert.AreEqual(6, allItems.Count, "Should have all 6 item types");

            Assert.AreEqual(15, inventorySystem.GetItemCount(ItemType.Fruit, false), "Storage fruit count");
            Assert.AreEqual(10, inventorySystem.GetItemCount(ItemType.Fruit, true), "Shop fruit count");
            Assert.AreEqual(5, inventorySystem.GetItemCount(ItemType.Potion, false), "Storage potion count");
            Assert.AreEqual(5, inventorySystem.GetItemCount(ItemType.Potion, true), "Shop potion count");
        }

        [Test]
        public void SaveAndLoad_MarketHistory()
        {
            // Arrange
            // Generate price history
            for (int i = 0; i < 10; i++)
            {
                marketSystem.UpdatePrices();
            }

            var originalHistory = new System.Collections.Generic.Dictionary<ItemType, System.Collections.Generic.List<float>>();
            foreach (ItemType itemType in System.Enum.GetValues(typeof(ItemType)))
            {
                originalHistory[itemType] = marketSystem.GetPriceHistory(itemType).ToList();
            }

            // Act
            saveSystem.Save(0);

            // Clear market data (would need method in MarketSystem)
            marketSystem.UpdatePrices(); // This changes current prices

            saveSystem.Load(0);

            // Assert
            foreach (ItemType itemType in System.Enum.GetValues(typeof(ItemType)))
            {
                var loadedHistory = marketSystem.GetPriceHistory(itemType);
                var originalList = originalHistory[itemType];

                Assert.AreEqual(originalList.Count, loadedHistory.Count,
                    $"History count for {itemType} should match");

                // Check at least some prices match
                if (originalList.Count > 0)
                {
                    AssertFloatEquals(originalList[0], loadedHistory[0], 0.01f);
                    AssertFloatEquals(originalList[originalList.Count - 1],
                        loadedHistory[loadedHistory.Count - 1], 0.01f);
                }
            }
        }

        [UnityTest]
        public IEnumerator SaveAndLoad_BankDeposits()
        {
            // Arrange
            var deposit1 = bankSystem.CreateDeposit(1000f, 7);
            var deposit2 = bankSystem.CreateDeposit(5000f, 30);

            // Advance some days
            AdvanceDays(3);
            yield return null;

            // Act
            saveSystem.Save(0);

            // Clear bank data
            bankSystem.ClearAllDeposits();

            saveSystem.Load(0);

            // Assert
            var deposits = bankSystem.GetAllDeposits();
            Assert.AreEqual(2, deposits.Count, "Should have 2 deposits");

            var loaded1 = deposits.FirstOrDefault(d => d.amount == 1000f);
            var loaded2 = deposits.FirstOrDefault(d => d.amount == 5000f);

            Assert.IsNotNull(loaded1, "First deposit should exist");
            Assert.IsNotNull(loaded2, "Second deposit should exist");
            Assert.AreEqual(7, loaded1.termDays);
            Assert.AreEqual(30, loaded2.termDays);
        }

        [Test]
        public void SaveAndLoad_UnlockedFeatures()
        {
            // Arrange
            featureUnlockSystem.UnlockFeature(GameFeature.PricePrediction);
            featureUnlockSystem.UnlockFeature(GameFeature.BankAccount);
            featureUnlockSystem.UnlockFeature(GameFeature.ShopInvestment);

            // Act
            saveSystem.Save(0);

            // Reset features
            featureUnlockSystem.ResetAllFeatures();

            saveSystem.Load(0);

            // Assert
            Assert.IsTrue(featureUnlockSystem.IsFeatureUnlocked(GameFeature.PricePrediction));
            Assert.IsTrue(featureUnlockSystem.IsFeatureUnlocked(GameFeature.BankAccount));
            Assert.IsTrue(featureUnlockSystem.IsFeatureUnlocked(GameFeature.ShopInvestment));
            Assert.IsFalse(featureUnlockSystem.IsFeatureUnlocked(GameFeature.MarketManipulation));
        }

        [Test]
        public void SaveAndLoad_Achievements()
        {
            // Arrange
            var achievement1 = new Achievement
            {
                id = "first_sale",
                name = "First Sale",
                description = "Complete your first sale",
                isUnlocked = true,
            };

            var achievement2 = new Achievement
            {
                id = "master_trader",
                name = "Master Trader",
                description = "Reach Master rank",
                isUnlocked = true,
            };

            achievementSystem.UnlockAchievement("first_sale");
            achievementSystem.UnlockAchievement("master_trader");

            // Act
            saveSystem.Save(0);
            achievementSystem.ResetAllAchievements();
            saveSystem.Load(0);

            // Assert
            var unlocked = achievementSystem.GetUnlockedAchievements();
            Assert.AreEqual(2, unlocked.Count);
            Assert.IsTrue(unlocked.Any(a => a.id == "first_sale"));
            Assert.IsTrue(unlocked.Any(a => a.id == "master_trader"));
        }

        [Test]
        public void SaveAndLoad_Investments()
        {
            // Arrange
            gameManager.PlayerData.SetMoney(10000);

            // Shop investments
            shopInvestmentSystem.Invest(ShopUpgradeType.Storage, 1000f);
            shopInvestmentSystem.Invest(ShopUpgradeType.Display, 500f);

            // Merchant investments
            var merchant = new MerchantInfo
            {
                id = "merchant_1",
                name = "Test Merchant",
                dividendRate = 0.05f,
            };
            merchantInvestmentSystem.InvestInMerchant(merchant.id, 2000f);

            // Act
            saveSystem.Save(0);

            // Reset investments
            shopInvestmentSystem.ResetAllInvestments();
            merchantInvestmentSystem.ClearAllInvestments();

            saveSystem.Load(0);

            // Assert
            var storageInv = shopInvestmentSystem.GetInvestment(ShopUpgradeType.Storage);
            var displayInv = shopInvestmentSystem.GetInvestment(ShopUpgradeType.Display);

            Assert.IsNotNull(storageInv);
            Assert.IsNotNull(displayInv);
            AssertFloatEquals(1000f, storageInv.totalInvested);
            AssertFloatEquals(500f, displayInv.totalInvested);

            var merchantInv = merchantInvestmentSystem.GetInvestment("merchant_1");
            Assert.IsNotNull(merchantInv);
            AssertFloatEquals(2000f, merchantInv.totalInvested);
        }

        [Test]
        public void SaveSlots_IndependentData()
        {
            // Arrange
            gameManager.PlayerData.SetMoney(1111);
            saveSystem.Save(0);

            gameManager.PlayerData.SetMoney(2222);
            saveSystem.Save(1);

            gameManager.PlayerData.SetMoney(3333);
            saveSystem.Save(2);

            // Act & Assert
            saveSystem.Load(0);
            Assert.AreEqual(1111, gameManager.PlayerData.CurrentMoney);

            saveSystem.Load(1);
            Assert.AreEqual(2222, gameManager.PlayerData.CurrentMoney);

            saveSystem.Load(2);
            Assert.AreEqual(3333, gameManager.PlayerData.CurrentMoney);
        }

        [Test]
        public void SaveFileCorruption_HandledGracefully()
        {
            // Arrange
            saveSystem.Save(0);

            // Corrupt the save file
            string savePath = Path.Combine(Application.persistentDataPath, "Saves", "save_slot0.json");
            if (File.Exists(savePath))
            {
                File.WriteAllText(savePath, "CORRUPTED DATA!@#$%");
            }

            // Act
            bool loadSuccess = saveSystem.Load(0);

            // Assert
            Assert.IsFalse(loadSuccess, "Load should fail with corrupted data");
            // Game should still be playable with default state
            Assert.IsNotNull(gameManager);
            Assert.IsNotNull(gameManager.PlayerData);
        }

        [UnityTest]
        public IEnumerator AutoSave_WorksCorrectly()
        {
            // Arrange
            gameManager.PlayerData.SetMoney(7777);

            // Enable auto-save with short interval
            saveSystem.EnableAutoSave = true;
            saveSystem.AutoSaveInterval = 1f; // 1 second for testing

            // Act
            yield return new WaitForSeconds(1.5f);

            // Change data and load to verify auto-save worked
            gameManager.PlayerData.SetMoney(0);
            saveSystem.Load(saveSystem.CurrentSlot);

            // Assert
            Assert.AreEqual(7777, gameManager.PlayerData.CurrentMoney,
                "Auto-save should have preserved the data");
        }

        [Test]
        public void BackupSystem_CreatesAndRestores()
        {
            // Arrange
            gameManager.PlayerData.SetMoney(9999);
            timeManager.LoadTimeData(50, Season.Autumn, DayPhase.Night, 0.9f);

            // Act
            saveSystem.Save(0);

            // Modify and save again (should create backup)
            gameManager.PlayerData.SetMoney(1111);
            saveSystem.Save(0);

            // Load backup
            bool backupLoaded = saveSystem.LoadBackup();

            // Assert
            Assert.IsTrue(backupLoaded, "Backup should load successfully");
            Assert.AreEqual(9999, gameManager.PlayerData.CurrentMoney,
                "Backup should restore original data");
        }

        [Test]
        public void EmergencySave_CreatesSpecialSave()
        {
            // Arrange
            gameManager.PlayerData.SetMoney(55555);
            var specialItem = CreateTestItem(ItemType.Gem, 99, 9999f);
            inventorySystem.AddItem(specialItem);

            // Act
            bool emergencySaved = saveSystem.EmergencySave();

            // Reset data
            gameManager.PlayerData.SetMoney(0);
            inventorySystem.ClearInventory();

            // Load emergency save
            string emergencyPath = Path.Combine(Application.persistentDataPath, "Saves", "emergency_save.json");
            Assert.IsTrue(File.Exists(emergencyPath), "Emergency save file should exist");

            // Note: Would need to add LoadEmergencySave method to SaveSystem for complete test

            // Assert
            Assert.IsTrue(emergencySaved, "Emergency save should succeed");
        }
    }
}
