using NUnit.Framework;
using UnityEngine;
using MerchantTails.Data;

namespace MerchantTails.Tests
{
    /// <summary>
    /// PlayerDataのユニットテスト
    /// セーブ/ロード、ランクアップ、データ整合性を検証
    /// </summary>
    public class PlayerDataTests
    {
        private PlayerData playerData;
        
        [SetUp]
        public void SetUp()
        {
            // テスト用PlayerDataインスタンスを作成
            playerData = ScriptableObject.CreateInstance<PlayerData>();
        }
        
        [TearDown] 
        public void TearDown()
        {
            if (playerData != null)
            {
                Object.DestroyImmediate(playerData);
            }
        }
        
        [Test]
        public void PlayerData_InitialValues_AreCorrect()
        {
            // 初期値の確認
            Assert.AreEqual("新米商人", playerData.PlayerName);
            Assert.AreEqual(1000, playerData.CurrentMoney);
            Assert.AreEqual(MerchantRank.Apprentice, playerData.CurrentRank);
            Assert.AreEqual(0, playerData.TotalProfit);
            Assert.AreEqual(0, playerData.SuccessfulTransactions);
            Assert.IsFalse(playerData.TutorialCompleted);
            Assert.AreEqual(1, playerData.DaysSinceStart);
            Assert.AreEqual(Season.Spring, playerData.CurrentSeason);
        }
        
        [Test]
        public void AddMoney_PositiveAmount_IncreasesCurrentMoney()
        {
            // Arrange
            int initialMoney = playerData.CurrentMoney;
            int addAmount = 500;
            
            // Act
            playerData.AddMoney(addAmount);
            
            // Assert
            Assert.AreEqual(initialMoney + addAmount, playerData.CurrentMoney);
        }
        
        [Test]
        public void AddMoney_NegativeAmount_DoesNotChangeCurrentMoney()
        {
            // Arrange
            int initialMoney = playerData.CurrentMoney;
            
            // Act
            playerData.AddMoney(-100);
            
            // Assert
            Assert.AreEqual(initialMoney, playerData.CurrentMoney);
        }
        
        [Test]
        public void SpendMoney_SufficientFunds_DecreasesCurrentMoneyAndReturnsTrue()
        {
            // Arrange
            playerData.AddMoney(1000); // 2000G total
            int spendAmount = 500;
            
            // Act
            bool result = playerData.SpendMoney(spendAmount);
            
            // Assert
            Assert.IsTrue(result);
            Assert.AreEqual(1500, playerData.CurrentMoney);
        }
        
        [Test]
        public void SpendMoney_InsufficientFunds_DoesNotChangeMoneyAndReturnsFalse()
        {
            // Arrange
            int initialMoney = playerData.CurrentMoney; // 1000G
            int spendAmount = 2000;
            
            // Act
            bool result = playerData.SpendMoney(spendAmount);
            
            // Assert
            Assert.IsFalse(result);
            Assert.AreEqual(initialMoney, playerData.CurrentMoney);
        }
        
        [Test]
        public void UpdateRank_SkilledRankMoney_PromotesToSkilled()
        {
            // Arrange
            playerData.AddMoney(4000); // 5000G total - Skilled rank threshold
            
            // Act
            playerData.UpdateRank();
            
            // Assert
            Assert.AreEqual(MerchantRank.Skilled, playerData.CurrentRank);
        }
        
        [Test]
        public void UpdateRank_VeteranRankMoney_PromotesToVeteran()
        {
            // Arrange
            playerData.AddMoney(9000); // 10000G total - Veteran rank threshold
            
            // Act
            playerData.UpdateRank();
            
            // Assert
            Assert.AreEqual(MerchantRank.Veteran, playerData.CurrentRank);
        }
        
        [Test]
        public void UpdateRank_MasterRankMoney_PromotesToMaster()
        {
            // Arrange
            playerData.AddMoney(15000); // 16000G total - Master rank threshold
            
            // Act
            playerData.UpdateRank();
            
            // Assert
            Assert.AreEqual(MerchantRank.Master, playerData.CurrentRank);
        }
        
        [Test]
        public void AddTransaction_ValidTransaction_UpdatesStats()
        {
            // Arrange
            int profit = 200;
            
            // Act
            playerData.AddTransaction(profit);
            
            // Assert
            Assert.AreEqual(1, playerData.SuccessfulTransactions);
            Assert.AreEqual(profit, playerData.TotalProfit);
            Assert.AreEqual(1200, playerData.CurrentMoney); // 1000 + 200
        }
        
        [Test]
        public void AddTransaction_MultipleTransactions_AccumulatesStats()
        {
            // Arrange & Act
            playerData.AddTransaction(100);
            playerData.AddTransaction(150);
            playerData.AddTransaction(75);
            
            // Assert
            Assert.AreEqual(3, playerData.SuccessfulTransactions);
            Assert.AreEqual(325, playerData.TotalProfit);
            Assert.AreEqual(1325, playerData.CurrentMoney);
        }
        
        [Test]
        public void CompleteTutorial_SetsFlag_ToTrue()
        {
            // Act
            playerData.CompleteTutorial();
            
            // Assert
            Assert.IsTrue(playerData.TutorialCompleted);
        }
        
        [Test]
        public void AdvanceDay_IncrementsDay_UpdatesSeason()
        {
            // Arrange
            int initialDays = playerData.DaysSinceStart;
            
            // Act
            playerData.AdvanceDay();
            
            // Assert
            Assert.AreEqual(initialDays + 1, playerData.DaysSinceStart);
        }
        
        [Test]
        public void AdvanceDay_After30Days_ChangesSeason()
        {
            // Arrange - 春から夏への変更をテスト
            for (int i = 0; i < 29; i++)
            {
                playerData.AdvanceDay(); // 30日目まで進める
            }
            Assert.AreEqual(Season.Spring, playerData.CurrentSeason);
            
            // Act
            playerData.AdvanceDay(); // 31日目
            
            // Assert
            Assert.AreEqual(Season.Summer, playerData.CurrentSeason);
        }
        
        [Test]
        public void GetSaveData_ReturnsCorrectJsonString()
        {
            // Arrange
            playerData.AddMoney(500);
            playerData.CompleteTutorial();
            
            // Act
            string saveData = playerData.GetSaveData();
            
            // Assert
            Assert.IsNotNull(saveData);
            Assert.IsTrue(saveData.Contains("1500")); // money
            Assert.IsTrue(saveData.Contains("true")); // tutorial completed
        }
        
        [Test]
        [TestCase(-100)]
        [TestCase(0)]
        public void AddMoney_InvalidAmounts_DoesNotChangeMoney(int invalidAmount)
        {
            // Arrange
            int initialMoney = playerData.CurrentMoney;
            
            // Act
            playerData.AddMoney(invalidAmount);
            
            // Assert
            Assert.AreEqual(initialMoney, playerData.CurrentMoney);
        }
        
        [Test]
        public void RankProgression_FullCycle_WorksCorrectly()
        {
            // Test complete rank progression
            Assert.AreEqual(MerchantRank.Apprentice, playerData.CurrentRank);
            
            // To Skilled (5000G)
            playerData.AddMoney(4000);
            playerData.UpdateRank();
            Assert.AreEqual(MerchantRank.Skilled, playerData.CurrentRank);
            
            // To Veteran (10000G)
            playerData.AddMoney(5000);
            playerData.UpdateRank();
            Assert.AreEqual(MerchantRank.Veteran, playerData.CurrentRank);
            
            // To Master (15000G+)
            playerData.AddMoney(5000);
            playerData.UpdateRank();
            Assert.AreEqual(MerchantRank.Master, playerData.CurrentRank);
        }
    }
}