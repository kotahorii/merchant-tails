using System;
using System.Collections;
using MerchantTails.Core;
using MerchantTails.Data;
using NUnit.Framework;
using UnityEngine;
using UnityEngine.TestTools;

namespace MerchantTails.Testing
{
    /// <summary>
    /// テストの基底クラス
    /// 共通のセットアップとクリーンアップ処理を提供
    /// </summary>
    public abstract class TestBase
    {
        protected GameObject testGameObject;
        protected GameManager gameManager;
        protected TimeManager timeManager;
        protected MarketSystem marketSystem;
        protected InventorySystem inventorySystem;

        // EventSystem is replaced by EventBus static class

        [SetUp]
        public virtual void Setup()
        {
            // テスト用GameObjectを作成
            testGameObject = new GameObject("TestGameObject");

            // 基本的なマネージャーをセットアップ
            SetupGameManager();
            SetupTimeManager();
            SetupMarketSystem();
            SetupInventorySystem();
            SetupEventSystem();

            // エラーハンドラーを初期化
            ErrorHandler.Initialize();
            ErrorHandler.SetDebugMode(true);
        }

        [TearDown]
        public virtual void Teardown()
        {
            // エラー統計をリセット
            ErrorHandler.ResetErrorStatistics();

            // GameObjectを破棄
            if (testGameObject != null)
            {
                UnityEngine.Object.Destroy(testGameObject);
            }

            // 静的インスタンスをクリア
            ClearStaticInstances();
        }

        protected virtual void SetupGameManager()
        {
            gameManager = testGameObject.AddComponent<GameManager>();
            gameManager.PlayerData = ScriptableObject.CreateInstance<PlayerData>();
            gameManager.PlayerData.SetPlayerName("TestPlayer");
            gameManager.PlayerData.SetMoney(1000);
            gameManager.PlayerData.SetRank(MerchantRank.Apprentice);
        }

        protected virtual void SetupTimeManager()
        {
            timeManager = testGameObject.AddComponent<TimeManager>();
            // デフォルトの時間設定
            timeManager.LoadTimeData(1, Season.Spring, DayPhase.Morning, 0f);
        }

        protected virtual void SetupMarketSystem()
        {
            marketSystem = testGameObject.AddComponent<MarketSystem>();
            // テスト用の価格設定
            foreach (ItemType itemType in Enum.GetValues(typeof(ItemType)))
            {
                var testPrice = GetTestBasePrice(itemType);
                // MarketSystemの初期化処理に依存
            }
        }

        protected virtual void SetupInventorySystem()
        {
            inventorySystem = testGameObject.AddComponent<InventorySystem>();
        }

        protected virtual void SetupEventSystem()
        {
            eventSystem = testGameObject.AddComponent<EventSystem>();
        }

        protected virtual void ClearStaticInstances()
        {
            // リフレクションを使用して静的インスタンスをクリア
            var types = new[]
            {
                typeof(GameManager),
                typeof(TimeManager),
                typeof(MarketSystem),
                typeof(InventorySystem),
                typeof(EventSystem),
                typeof(SaveSystem),
                typeof(ResourceManager),
                typeof(UpdateManager),
                typeof(MemoryOptimizationSystem),
                typeof(FrameRateStabilizer),
                typeof(LocalizationManager),
            };

            foreach (var type in types)
            {
                var instanceField = type.GetField(
                    "instance",
                    System.Reflection.BindingFlags.NonPublic | System.Reflection.BindingFlags.Static
                );

                if (instanceField != null)
                {
                    instanceField.SetValue(null, null);
                }
            }
        }

        // ヘルパーメソッド
        protected float GetTestBasePrice(ItemType itemType)
        {
            return itemType switch
            {
                ItemType.Fruit => 10f,
                ItemType.Potion => 50f,
                ItemType.Weapon => 200f,
                ItemType.Accessory => 100f,
                ItemType.MagicBook => 300f,
                ItemType.Gem => 500f,
                _ => 100f,
            };
        }

        protected InventoryItem CreateTestItem(ItemType type, int quantity = 1, float? price = null)
        {
            return new InventoryItem
            {
                uniqueId = Guid.NewGuid().ToString(),
                itemType = type,
                purchasePrice = price ?? GetTestBasePrice(type),
                quality = ItemQuality.Common,
                purchaseDay = 1,
                expiryDay = -1,
                location = InventoryLocation.Trading,
            };
        }

        protected void AdvanceTime(float hours)
        {
            for (int i = 0; i < hours * 4; i++) // 15分ごとに進める
            {
                timeManager.AdvanceTime(0.25f);
            }
        }

        protected void AdvanceDays(int days)
        {
            for (int i = 0; i < days; i++)
            {
                timeManager.AdvanceDay();
            }
        }

        // アサーションヘルパー
        protected void AssertFloatEquals(float expected, float actual, float tolerance = 0.01f)
        {
            Assert.That(actual, Is.EqualTo(expected).Within(tolerance), $"Expected {expected} but was {actual}");
        }

        protected void AssertInRange(float value, float min, float max)
        {
            Assert.That(value, Is.InRange(min, max), $"Value {value} is not in range [{min}, {max}]");
        }

        // コルーチンテスト用のヘルパー
        protected IEnumerator WaitForCondition(Func<bool> condition, float timeout = 5f)
        {
            float elapsed = 0f;
            while (!condition() && elapsed < timeout)
            {
                yield return null;
                elapsed += Time.deltaTime;
            }

            if (elapsed >= timeout)
            {
                Assert.Fail($"Condition was not met within {timeout} seconds");
            }
        }

        protected IEnumerator WaitForSeconds(float seconds)
        {
            yield return new WaitForSeconds(seconds);
        }
    }
}
