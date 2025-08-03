using System.Collections;
using MerchantTails.Core;
using MerchantTails.Data;
using MerchantTails.Events;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.TestTools;

namespace MerchantTails.Testing
{
    /// <summary>
    /// システム間の統合テスト
    /// 複数のシステムが連携して動作することを確認
    /// </summary>
    public class IntegrationTests : TestBase
    {
        private BankSystem bankSystem;
        private FeatureUnlockSystem featureUnlockSystem;
        private AchievementSystem achievementSystem;

        public override void Setup()
        {
            base.Setup();

            // 追加のシステムをセットアップ
            bankSystem = testGameObject.AddComponent<BankSystem>();
            featureUnlockSystem = testGameObject.AddComponent<FeatureUnlockSystem>();
            achievementSystem = testGameObject.AddComponent<AchievementSystem>();
        }

        [UnityTest]
        public IEnumerator BuyAndSellFlow_CompletesSuccessfully()
        {
            // Arrange
            var itemType = ItemType.Fruit;
            float buyPrice = marketSystem.GetCurrentPrice(itemType);
            int quantity = 10;
            float totalCost = buyPrice * quantity;

            // プレイヤーに十分なお金を与える
            gameManager.PlayerData.SetMoney(totalCost * 2);
            float initialMoney = gameManager.PlayerData.CurrentMoney;

            // Act - 購入
            var buyItem = CreateTestItem(itemType, quantity, buyPrice);
            bool buySuccess = inventorySystem.AddItem(buyItem);
            if (buySuccess)
            {
                gameManager.PlayerData.SpendMoney(totalCost);
            }

            // 時間を進めて価格を変動させる
            yield return null;
            AdvanceTime(24f);
            marketSystem.UpdatePrices();

            float sellPrice = marketSystem.GetCurrentPrice(itemType);

            // 販売
            bool sellSuccess = inventorySystem.RemoveItem(buyItem.id, quantity);
            if (sellSuccess)
            {
                float revenue = sellPrice * quantity;
                gameManager.PlayerData.AddMoney(revenue);
            }

            // Assert
            Assert.IsTrue(buySuccess, "Buy should succeed");
            Assert.IsTrue(sellSuccess, "Sell should succeed");

            float finalMoney = gameManager.PlayerData.CurrentMoney;
            float profit = finalMoney - initialMoney;

            // 利益または損失が発生したことを確認
            Assert.AreNotEqual(initialMoney, finalMoney,
                "Money should change after buy/sell cycle");
        }

        [UnityTest]
        public IEnumerator DayProgressionFlow_UpdatesAllSystems()
        {
            // Arrange
            int initialDay = timeManager.CurrentDay;
            float initialMarketPrice = marketSystem.GetCurrentPrice(ItemType.Potion);

            // 腐りやすいアイテムを追加
            var perishableItem = CreateTestItem(ItemType.Fruit, 5);
            inventorySystem.AddItem(perishableItem);

            // Act - 1日進める
            timeManager.AdvanceDay();
            yield return WaitForSeconds(0.1f);

            // Assert
            Assert.AreEqual(initialDay + 1, timeManager.CurrentDay,
                "Day should advance");

            // 市場価格が更新されたか確認
            float newMarketPrice = marketSystem.GetCurrentPrice(ItemType.Potion);
            Assert.AreNotEqual(initialMarketPrice, newMarketPrice,
                "Market prices should update with new day");

            // アイテムの状態が更新されたか確認
            var items = inventorySystem.GetItemsOfType(ItemType.Fruit);
            if (items.Count > 0)
            {
                Assert.Less(items[0].condition, 1f,
                    "Perishable item condition should degrade");
            }
        }

        [UnityTest]
        public IEnumerator RankProgression_UnlocksFeatures()
        {
            // Arrange
            gameManager.PlayerData.SetRank(MerchantRank.Apprentice);
            gameManager.PlayerData.SetMoney(5000); // Skilled rank requirement

            // Act - ランクアップ条件を満たす
            float totalAssets = 5000;
            var assetEvent = new AssetCalculatedEvent(totalAssets);
            EventBus.Publish(assetEvent);

            yield return WaitForSeconds(0.1f);

            // Assert
            Assert.AreEqual(MerchantRank.Skilled, gameManager.PlayerData.CurrentRank,
                "Should rank up to Skilled");

            // 機能が解放されたか確認
            Assert.IsTrue(featureUnlockSystem.IsFeatureUnlocked(GameFeature.PricePrediction),
                "Price prediction should be unlocked at Skilled rank");
            Assert.IsTrue(featureUnlockSystem.IsFeatureUnlocked(GameFeature.BankAccount),
                "Bank account should be unlocked at Skilled rank");
        }

        [UnityTest]
        public IEnumerator EventTriggering_AffectsMarketPrices()
        {
            // Arrange
            var itemType = ItemType.Weapon;
            float originalPrice = marketSystem.GetCurrentPrice(itemType);

            // Act - イベントをトリガー
            var gameEvent = new GameEvent
            {
                id = "dragon_defeated",
                name = "Dragon Defeated",
                description = "The dragon has been defeated!",
                eventType = EventType.Major,
                duration = 3,
                priceModifiers = new System.Collections.Generic.Dictionary<ItemType, float>
                {
                    { ItemType.Weapon, 1.5f } // 武器の需要が50%増加
                }
            };

            eventSystem.TriggerEvent(gameEvent);
            yield return WaitForSeconds(0.1f);

            marketSystem.UpdatePrices();
            float eventPrice = marketSystem.GetCurrentPrice(itemType);

            // Assert
            Assert.Greater(eventPrice, originalPrice,
                "Weapon price should increase during dragon defeat event");
        }

        [UnityTest]
        public IEnumerator BankDeposit_GeneratesInterest()
        {
            // Arrange
            float depositAmount = 1000f;
            int termDays = 7;

            // Act - 預金を作成
            var deposit = bankSystem.CreateDeposit(depositAmount, termDays);
            Assert.IsNotNull(deposit, "Deposit should be created");

            // 満期まで日数を進める
            for (int i = 0; i < termDays; i++)
            {
                timeManager.AdvanceDay();
                yield return null;
            }

            // 満期を確認
            bool isMatured = bankSystem.CheckMaturity(deposit.id);
            float maturityAmount = bankSystem.CalculateMaturityAmount(deposit);

            // Assert
            Assert.IsTrue(isMatured, "Deposit should be matured");
            Assert.Greater(maturityAmount, depositAmount,
                "Maturity amount should include interest");
        }

        [UnityTest]
        public IEnumerator ShopInvestment_ImprovesEfficiency()
        {
            // Arrange
            var shopInvestmentSystem = testGameObject.AddComponent<ShopInvestmentSystem>();
            var upgradeType = ShopUpgradeType.Storage;
            float investmentAmount = 500f;

            gameManager.PlayerData.SetMoney(investmentAmount * 2);

            // Act - 投資を実行
            bool investSuccess = shopInvestmentSystem.Invest(upgradeType, investmentAmount);
            yield return WaitForSeconds(0.1f);

            // Assert
            Assert.IsTrue(investSuccess, "Investment should succeed");

            var investment = shopInvestmentSystem.GetInvestment(upgradeType);
            Assert.IsNotNull(investment, "Investment should exist");
            Assert.Greater(investment.level, 0, "Investment level should increase");

            // 効率ボーナスを確認
            float bonus = shopInvestmentSystem.GetEfficiencyBonus(upgradeType);
            Assert.Greater(bonus, 1f, "Should have efficiency bonus");
        }

        [UnityTest]
        public IEnumerator CompleteGameLoop_24Hours()
        {
            // Arrange
            float startMoney = 1000f;
            gameManager.PlayerData.SetMoney(startMoney);

            // 初期在庫を購入
            var fruitItem = CreateTestItem(ItemType.Fruit, 20, 10f);
            inventorySystem.AddItem(fruitItem);
            gameManager.PlayerData.SpendMoney(200f);

            // 店頭に移動
            inventorySystem.TransferToShop(ItemType.Fruit, 10);

            // Act - 24時間のゲームループをシミュレート
            for (int hour = 0; hour < 24; hour++)
            {
                // 時間を進める
                timeManager.AdvanceTime(1f);

                // 4時間ごとに市場価格を更新
                if (hour % 4 == 0)
                {
                    marketSystem.UpdatePrices();
                }

                // 顧客の来店をシミュレート（簡略化）
                if (hour >= 8 && hour <= 20) // 営業時間
                {
                    // ランダムに販売
                    if (Random.Range(0f, 1f) > 0.7f)
                    {
                        var shopItems = inventorySystem.GetShopItems();
                        if (shopItems.Count > 0)
                        {
                            var itemToSell = shopItems.First();
                            float sellPrice = marketSystem.GetCurrentPrice(itemToSell.Key);

                            if (inventorySystem.RemoveItem(itemToSell.Value.First().id, 1))
                            {
                                gameManager.PlayerData.AddMoney(sellPrice * 1.2f); // 20%マークアップ
                            }
                        }
                    }
                }

                yield return null;
            }

            // Assert
            float endMoney = gameManager.PlayerData.CurrentMoney;
            Assert.AreNotEqual(startMoney, endMoney,
                "Money should change after 24 hour cycle");

            // 統計を確認
            Assert.Greater(gameManager.PlayerData.TotalTransactions, 0,
                "Should have completed some transactions");
        }

        [UnityTest]
        public IEnumerator SaveAndLoad_PreservesGameState()
        {
            // Arrange
            var saveSystem = testGameObject.AddComponent<SaveSystem>();

            // ゲーム状態を設定
            gameManager.PlayerData.SetMoney(12345);
            gameManager.PlayerData.SetRank(MerchantRank.Veteran);
            timeManager.LoadTimeData(42, Season.Autumn, DayPhase.Evening, 0.6f);

            var testItem = CreateTestItem(ItemType.Gem, 3, 500f);
            inventorySystem.AddItem(testItem);

            // Act - セーブ
            saveSystem.Save(0);
            yield return WaitForSeconds(0.1f);

            // ゲーム状態をリセット
            gameManager.PlayerData.SetMoney(0);
            gameManager.PlayerData.SetRank(MerchantRank.Apprentice);
            timeManager.LoadTimeData(1, Season.Spring, DayPhase.Morning, 0f);
            inventorySystem.ClearInventory();

            // ロード
            bool loadSuccess = saveSystem.Load(0);
            yield return WaitForSeconds(0.1f);

            // Assert
            Assert.IsTrue(loadSuccess, "Load should succeed");
            Assert.AreEqual(12345, gameManager.PlayerData.CurrentMoney,
                "Money should be restored");
            Assert.AreEqual(MerchantRank.Veteran, gameManager.PlayerData.CurrentRank,
                "Rank should be restored");
            Assert.AreEqual(42, timeManager.CurrentDay,
                "Day should be restored");
            Assert.AreEqual(Season.Autumn, timeManager.CurrentSeason,
                "Season should be restored");
            Assert.AreEqual(3, inventorySystem.GetItemCount(ItemType.Gem),
                "Inventory should be restored");
        }
    }
}
