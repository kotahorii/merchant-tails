using MerchantTails.Data;
using MerchantTails.Inventory;
using MerchantTails.Market;
using NUnit.Framework;
using UnityEngine;

namespace MerchantTails.Tests
{
    /// <summary>
    /// データ整合性とScriptableObjectのテスト
    /// データの妥当性、シリアライゼーション、列挙型処理を検証
    /// </summary>
    public class DataIntegrityTests
    {
        [Test]
        public void ItemType_AllValues_AreValid()
        {
            // 全ItemType値が有効であることを確認
            var itemTypes = System.Enum.GetValues(typeof(ItemType));

            Assert.Greater(itemTypes.Length, 0, "ItemType should have at least one value");

            foreach (ItemType itemType in itemTypes)
            {
                Assert.IsTrue(
                    System.Enum.IsDefined(typeof(ItemType), itemType),
                    $"ItemType {itemType} should be properly defined"
                );
            }
        }

        [Test]
        public void Season_AllValues_AreValid()
        {
            // 全Season値が有効であることを確認
            var seasons = System.Enum.GetValues(typeof(Season));

            Assert.AreEqual(4, seasons.Length, "Should have exactly 4 seasons");

            foreach (Season season in seasons)
            {
                Assert.IsTrue(
                    System.Enum.IsDefined(typeof(Season), season),
                    $"Season {season} should be properly defined"
                );
            }
        }

        [Test]
        public void MerchantRank_AllValues_AreValid()
        {
            // 全MerchantRank値が有効であることを確認
            var ranks = System.Enum.GetValues(typeof(MerchantRank));

            Assert.AreEqual(4, ranks.Length, "Should have exactly 4 merchant ranks");

            foreach (MerchantRank rank in ranks)
            {
                Assert.IsTrue(
                    System.Enum.IsDefined(typeof(MerchantRank), rank),
                    $"MerchantRank {rank} should be properly defined"
                );
            }
        }

        [Test]
        public void GameState_AllValues_AreValid()
        {
            // 全GameState値が有効であることを確認
            var gameStates = System.Enum.GetValues(typeof(MerchantTails.Data.GameState));

            Assert.Greater(gameStates.Length, 0, "GameState should have at least one value");

            foreach (MerchantTails.Data.GameState state in gameStates)
            {
                Assert.IsTrue(
                    System.Enum.IsDefined(typeof(MerchantTails.Data.GameState), state),
                    $"GameState {state} should be properly defined"
                );
            }
        }

        [Test]
        public void InventoryLocation_AllValues_AreValid()
        {
            // 全InventoryLocation値が有効であることを確認
            var locations = System.Enum.GetValues(typeof(InventoryLocation));

            Assert.AreEqual(2, locations.Length, "Should have exactly 2 inventory locations");

            foreach (InventoryLocation location in locations)
            {
                Assert.IsTrue(
                    System.Enum.IsDefined(typeof(InventoryLocation), location),
                    $"InventoryLocation {location} should be properly defined"
                );
            }
        }

        [Test]
        public void DayPhase_AllValues_AreValid()
        {
            // 全DayPhase値が有効であることを確認
            var phases = System.Enum.GetValues(typeof(DayPhase));

            Assert.AreEqual(4, phases.Length, "Should have exactly 4 day phases");

            foreach (DayPhase phase in phases)
            {
                Assert.IsTrue(
                    System.Enum.IsDefined(typeof(DayPhase), phase),
                    $"DayPhase {phase} should be properly defined"
                );
            }
        }

        [Test]
        public void MarketData_CanBeCreated_WithValidDefaults()
        {
            // MarketDataのインスタンス化テスト
            var marketData = ScriptableObject.CreateInstance<MarketData>();

            Assert.IsNotNull(marketData, "MarketData should be creatable");

            // クリーンアップ
            Object.DestroyImmediate(marketData);
        }

        [Test]
        public void InventoryData_CanBeCreated_WithValidDefaults()
        {
            // InventoryDataのインスタンス化テスト
            var inventoryData = ScriptableObject.CreateInstance<InventoryData>();

            Assert.IsNotNull(inventoryData, "InventoryData should be creatable");

            // クリーンアップ
            Object.DestroyImmediate(inventoryData);
        }

        [Test]
        public void PlayerData_CanBeCreated_WithValidDefaults()
        {
            // PlayerDataのインスタンス化テスト
            var playerData = ScriptableObject.CreateInstance<PlayerData>();

            Assert.IsNotNull(playerData, "PlayerData should be creatable");
            Assert.IsNotNull(playerData.PlayerName, "Player name should not be null");
            Assert.GreaterOrEqual(playerData.CurrentMoney, 0, "Current money should not be negative");

            // クリーンアップ
            Object.DestroyImmediate(playerData);
        }

        [Test]
        public void EnumConversion_InvalidValues_HandleGracefully()
        {
            // 無効な列挙型値の処理テスト

            // ItemType無効値テスト
            bool isValidItemType = System.Enum.IsDefined(typeof(ItemType), 999);
            Assert.IsFalse(isValidItemType, "Invalid ItemType value should not be defined");

            // Season無効値テスト
            bool isValidSeason = System.Enum.IsDefined(typeof(Season), 10);
            Assert.IsFalse(isValidSeason, "Invalid Season value should not be defined");

            // MerchantRank無効値テスト
            bool isValidRank = System.Enum.IsDefined(typeof(MerchantRank), -1);
            Assert.IsFalse(isValidRank, "Invalid MerchantRank value should not be defined");
        }

        [Test]
        public void EnumParsing_StringValues_WorksCorrectly()
        {
            // 文字列からの列挙型解析テスト

            // ItemType解析
            bool fruitParseSuccess = System.Enum.TryParse<ItemType>("Fruit", out ItemType parsedFruit);
            Assert.IsTrue(fruitParseSuccess, "Should parse 'Fruit' to ItemType.Fruit");
            Assert.AreEqual(ItemType.Fruit, parsedFruit);

            // Season解析
            bool springParseSuccess = System.Enum.TryParse<Season>("Spring", out Season parsedSpring);
            Assert.IsTrue(springParseSuccess, "Should parse 'Spring' to Season.Spring");
            Assert.AreEqual(Season.Spring, parsedSpring);

            // 無効文字列の解析
            bool invalidParseSuccess = System.Enum.TryParse<ItemType>("InvalidItem", out ItemType _);
            Assert.IsFalse(invalidParseSuccess, "Should not parse invalid enum string");
        }

        [Test]
        public void ItemTypeCount_MatchesExpectedNumber()
        {
            // アイテムタイプ数が期待値と一致することを確認
            var itemTypes = System.Enum.GetValues(typeof(ItemType));
            Assert.AreEqual(
                6,
                itemTypes.Length,
                "Should have exactly 6 item types (Fruit, Potion, Weapon, Accessory, MagicBook, Gem)"
            );
        }

        [Test]
        public void SeasonProgression_IsLogical()
        {
            // 季節の進行が論理的であることを確認
            Assert.AreEqual(0, (int)Season.Spring, "Spring should be first season (0)");
            Assert.AreEqual(1, (int)Season.Summer, "Summer should be second season (1)");
            Assert.AreEqual(2, (int)Season.Autumn, "Autumn should be third season (2)");
            Assert.AreEqual(3, (int)Season.Winter, "Winter should be fourth season (3)");
        }

        [Test]
        public void MerchantRankProgression_IsLogical()
        {
            // 商人ランクの進行が論理的であることを確認
            Assert.AreEqual(0, (int)MerchantRank.Apprentice, "Apprentice should be first rank (0)");
            Assert.AreEqual(1, (int)MerchantRank.Skilled, "Skilled should be second rank (1)");
            Assert.AreEqual(2, (int)MerchantRank.Veteran, "Veteran should be third rank (2)");
            Assert.AreEqual(3, (int)MerchantRank.Master, "Master should be fourth rank (3)");
        }

        [Test]
        public void DayPhaseProgression_IsLogical()
        {
            // 日の時間帯の進行が論理的であることを確認
            Assert.AreEqual(0, (int)DayPhase.Morning, "Morning should be first phase (0)");
            Assert.AreEqual(1, (int)DayPhase.Afternoon, "Afternoon should be second phase (1)");
            Assert.AreEqual(2, (int)DayPhase.Evening, "Evening should be third phase (2)");
            Assert.AreEqual(3, (int)DayPhase.Night, "Night should be fourth phase (3)");
        }

        [Test]
        public void JsonSerialization_AllEnums_CanBeConverted()
        {
            // JSON変換可能性のテスト
            try
            {
                // ItemType
                string itemTypeJson = JsonUtility.ToJson(new { itemType = ItemType.Fruit });
                Assert.IsNotEmpty(itemTypeJson, "ItemType should be JSON serializable");

                // Season
                string seasonJson = JsonUtility.ToJson(new { season = Season.Spring });
                Assert.IsNotEmpty(seasonJson, "Season should be JSON serializable");

                // MerchantRank
                string rankJson = JsonUtility.ToJson(new { rank = MerchantRank.Apprentice });
                Assert.IsNotEmpty(rankJson, "MerchantRank should be JSON serializable");
            }
            catch (System.Exception e)
            {
                Assert.Fail($"Enum serialization failed: {e.Message}");
            }
        }

        [Test]
        public void ScriptableObjectSerialization_PlayerData_WorksCorrectly()
        {
            // PlayerDataのシリアライゼーションテスト
            var playerData = ScriptableObject.CreateInstance<PlayerData>();

            try
            {
                // JSON変換テスト
                string saveData = playerData.GetSaveData();
                Assert.IsNotNull(saveData, "Save data should not be null");
                Assert.IsTrue(saveData.Length > 0, "Save data should not be empty");

                // JSON解析テスト
                var parsedData = JsonUtility.FromJson<PlayerData>(saveData);
                Assert.IsNotNull(parsedData, "Should be able to parse save data back to PlayerData");
            }
            catch (System.Exception e)
            {
                Assert.Fail($"PlayerData serialization failed: {e.Message}");
            }
            finally
            {
                Object.DestroyImmediate(playerData);
            }
        }
    }
}
