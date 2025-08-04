using System;
using System.Collections.Generic;
using System.Linq;
using MerchantTails.Core;
using MerchantTails.Data;
using MerchantTails.Events;
using Unity.Burst;
using Unity.Collections;
using Unity.Jobs;
using Unity.Mathematics;
using UnityEngine;

namespace MerchantTails.Market
{
    /// <summary>
    /// 市場価格変動システム
    /// 6種類の商品の価格を管理し、季節・イベント・需給に基づく変動を処理
    /// </summary>
    public class MarketSystem : MonoBehaviour
    {
        [Header("Market Configuration")]
        [SerializeField]
        private MarketConfiguration marketConfig;

        [SerializeField]
        private bool enablePriceFluctuations = true;

        [SerializeField]
        private float fluctuationIntensity = 1.0f;

        [Header("Price History")]
        [SerializeField]
        private int maxHistoryDays = 90;

        // Market data storage
        private Dictionary<ItemType, MarketData> marketPrices;
        private Dictionary<ItemType, List<PriceHistory>> priceHistories;
        private List<MarketEvent> activeMarketEvents;

        // Job System用のNativeArrays
        private NativeArray<float> nativeBasePrices;
        private NativeArray<float> nativeCurrentPrices;
        private NativeArray<float> nativeVolatilities;
        private NativeArray<float> nativeDemands;
        private NativeArray<float> nativeSupplies;
        private NativeArray<float> nativeSeasonalModifiers;
        private NativeArray<float> nativeEventModifiers;
        private NativeArray<float> nativePriceChanges;
        private JobHandle priceCalculationHandle;

        // Seasonal price modifiers
        private readonly Dictionary<ItemType, Dictionary<Season, float>> seasonalModifiers =
            new Dictionary<ItemType, Dictionary<Season, float>>();

        public static MarketSystem Instance { get; private set; }

        // Events
        public event Action<ItemType, float, float> OnPriceChanged;

        private void Awake()
        {
            if (Instance == null)
            {
                Instance = this;
                DontDestroyOnLoad(gameObject);
                InitializeMarketSystem();
            }
            else
            {
                Destroy(gameObject);
            }
        }

        private void Start()
        {
            SubscribeToEvents();
        }

        private void InitializeMarketSystem()
        {
            Debug.Log("[MarketSystem] Initializing market system...");

            marketPrices = new Dictionary<ItemType, MarketData>();
            priceHistories = new Dictionary<ItemType, List<PriceHistory>>();
            activeMarketEvents = new List<MarketEvent>();

            // Initialize market data for all item types
            var itemTypes = Enum.GetValues(typeof(ItemType));
            int itemCount = itemTypes.Length;

            // Initialize NativeArrays for Job System
            nativeBasePrices = new NativeArray<float>(itemCount, Allocator.Persistent);
            nativeCurrentPrices = new NativeArray<float>(itemCount, Allocator.Persistent);
            nativeVolatilities = new NativeArray<float>(itemCount, Allocator.Persistent);
            nativeDemands = new NativeArray<float>(itemCount, Allocator.Persistent);
            nativeSupplies = new NativeArray<float>(itemCount, Allocator.Persistent);
            nativeSeasonalModifiers = new NativeArray<float>(itemCount, Allocator.Persistent);
            nativeEventModifiers = new NativeArray<float>(itemCount, Allocator.Persistent);
            nativePriceChanges = new NativeArray<float>(itemCount, Allocator.Persistent);
            
            int index = 0;
            foreach (ItemType itemType in itemTypes)
            {
                var marketData = CreateInitialMarketData(itemType);
                marketPrices[itemType] = marketData;
                priceHistories[itemType] = new List<PriceHistory>();

                // Initialize native arrays
                nativeBasePrices[index] = marketData.basePrice;
                nativeCurrentPrices[index] = marketData.currentPrice;
                nativeVolatilities[index] = marketData.volatility;
                nativeDemands[index] = marketData.demand;
                nativeSupplies[index] = marketData.supply;
                nativeSeasonalModifiers[index] = 1.0f;
                nativeEventModifiers[index] = 1.0f;
                
                // Record initial price
                RecordPrice(itemType, marketData.currentPrice);
                index++;
            }

            InitializeSeasonalModifiers();
            Debug.Log($"[MarketSystem] Initialized {marketPrices.Count} market items with Job System support");
        }

        private MarketData CreateInitialMarketData(ItemType itemType)
        {
            var basePrice = GetBasePrice(itemType);
            return new MarketData
            {
                itemType = itemType,
                basePrice = basePrice,
                currentPrice = basePrice,
                volatility = GetItemVolatility(itemType),
                demand = 1.0f,
                supply = 1.0f,
                lastUpdateDay = 1,
            };
        }

        private void InitializeSeasonalModifiers()
        {
            // くだもの (Fruit) - 夏に需要増
            seasonalModifiers[ItemType.Fruit] = new Dictionary<Season, float>
            {
                { Season.Spring, 1.0f },
                { Season.Summer, 1.3f },
                { Season.Autumn, 0.9f },
                { Season.Winter, 0.7f },
            };

            // ポーション (Potion) - 夏と冬に需要増
            seasonalModifiers[ItemType.Potion] = new Dictionary<Season, float>
            {
                { Season.Spring, 1.0f },
                { Season.Summer, 1.2f },
                { Season.Autumn, 1.0f },
                { Season.Winter, 1.3f },
            };

            // 武器 (Weapon) - 秋に需要増（戦争シーズン）
            seasonalModifiers[ItemType.Weapon] = new Dictionary<Season, float>
            {
                { Season.Spring, 1.0f },
                { Season.Summer, 0.9f },
                { Season.Autumn, 1.4f },
                { Season.Winter, 1.1f },
            };

            // アクセサリー (Accessory) - 春と秋に需要増（社交シーズン）
            seasonalModifiers[ItemType.Accessory] = new Dictionary<Season, float>
            {
                { Season.Spring, 1.3f },
                { Season.Summer, 0.8f },
                { Season.Autumn, 1.2f },
                { Season.Winter, 1.0f },
            };

            // 魔法書 (MagicBook) - 冬に需要増（研究シーズン）
            seasonalModifiers[ItemType.MagicBook] = new Dictionary<Season, float>
            {
                { Season.Spring, 1.0f },
                { Season.Summer, 0.8f },
                { Season.Autumn, 1.1f },
                { Season.Winter, 1.4f },
            };

            // 宝石 (Gem) - 一年中安定だが春に若干増
            seasonalModifiers[ItemType.Gem] = new Dictionary<Season, float>
            {
                { Season.Spring, 1.1f },
                { Season.Summer, 1.0f },
                { Season.Autumn, 1.0f },
                { Season.Winter, 1.0f },
            };
        }

        private void SubscribeToEvents()
        {
            // Time-based events
            EventBus.Subscribe<PhaseChangedEvent>(OnPhaseChanged);
            EventBus.Subscribe<SeasonChangedEvent>(OnSeasonChanged);
            EventBus.Subscribe<DayChangedEvent>(OnDayChanged);

            // Game events that affect market
            EventBus.Subscribe<GameEventTriggeredEvent>(OnGameEventTriggered);
            EventBus.Subscribe<TransactionCompletedEvent>(OnTransactionCompleted);
        }

        private void OnDestroy()
        {
            // Job完了を待つ
            priceCalculationHandle.Complete();
            
            // NativeArraysの解放
            if (nativeBasePrices.IsCreated) nativeBasePrices.Dispose();
            if (nativeCurrentPrices.IsCreated) nativeCurrentPrices.Dispose();
            if (nativeVolatilities.IsCreated) nativeVolatilities.Dispose();
            if (nativeDemands.IsCreated) nativeDemands.Dispose();
            if (nativeSupplies.IsCreated) nativeSupplies.Dispose();
            if (nativeSeasonalModifiers.IsCreated) nativeSeasonalModifiers.Dispose();
            if (nativeEventModifiers.IsCreated) nativeEventModifiers.Dispose();
            if (nativePriceChanges.IsCreated) nativePriceChanges.Dispose();
            
            // Unsubscribe from events
            EventBus.Unsubscribe<PhaseChangedEvent>(OnPhaseChanged);
            EventBus.Unsubscribe<SeasonChangedEvent>(OnSeasonChanged);
            EventBus.Unsubscribe<DayChangedEvent>(OnDayChanged);
            EventBus.Unsubscribe<GameEventTriggeredEvent>(OnGameEventTriggered);
            EventBus.Unsubscribe<TransactionCompletedEvent>(OnTransactionCompleted);
        }

        // Unity 6 Job System対応の価格更新メソッド
        public void UpdatePrices()
        {
            if (!enablePriceFluctuations)
                return;

            // Job System用データの準備
            PrepareJobData();

            // Job Systemで価格計算を並列実行
            var priceJob = new MarketPriceCalculationJob
            {
                basePrices = nativeBasePrices,
                volatilities = nativeVolatilities,
                demands = nativeDemands,
                supplies = nativeSupplies,
                seasonalModifiers = nativeSeasonalModifiers,
                eventModifiers = nativeEventModifiers,
                globalMarketTrend = CalculateGlobalMarketTrend(),
                randomSeed = (uint)UnityEngine.Random.Range(0, int.MaxValue),
                deltaTime = Time.deltaTime,
                calculatedPrices = nativeCurrentPrices,
                priceChanges = nativePriceChanges
            };

            // Jobのスケジューリング
            priceCalculationHandle = priceJob.Schedule(marketPrices.Count, 1);
            
            // Jobの完了を待つ
            priceCalculationHandle.Complete();
            
            // 結果をDictionaryに反映
            ApplyJobResults();

            Debug.Log("[MarketSystem] Prices updated using Job System");
        }
        
        private void PrepareJobData()
        {
            int index = 0;
            foreach (var kvp in marketPrices)
            {
                var marketData = kvp.Value;
                nativeCurrentPrices[index] = marketData.currentPrice;
                nativeDemands[index] = marketData.demand;
                nativeSupplies[index] = marketData.supply;
                nativeSeasonalModifiers[index] = GetSeasonalModifier(kvp.Key);
                nativeEventModifiers[index] = GetEventModifier(kvp.Key);
                index++;
            }
        }
        
        private void ApplyJobResults()
        {
            int index = 0;
            foreach (var kvp in marketPrices.ToList())
            {
                var itemType = kvp.Key;
                var marketData = kvp.Value;
                
                float oldPrice = marketData.currentPrice;
                float newPrice = nativeCurrentPrices[index];
                
                marketData.currentPrice = newPrice;
                marketData.lastUpdateDay = GameManager.Instance.TimeManager.CurrentDay;
                
                // Record price history
                RecordPrice(itemType, newPrice);
                
                // Publish price change event
                if (Math.Abs(oldPrice - newPrice) > 0.01f)
                {
                    OnPriceChanged?.Invoke(itemType, oldPrice, newPrice);
                    EventBus.Publish(new PriceChangedEvent(itemType, oldPrice, newPrice));
                }
                
                index++;
            }
        }
        
        private float CalculateGlobalMarketTrend()
        {
            // グローバル市場トレンドの計算（-1.0 ~ 1.0）
            float trend = 0f;
            foreach (var history in priceHistories.Values)
            {
                if (history.Count > 10)
                {
                    float recentAvg = history.TakeLast(5).Average(h => h.price);
                    float olderAvg = history.Skip(history.Count - 10).Take(5).Average(h => h.price);
                    trend += (recentAvg - olderAvg) / olderAvg;
                }
            }
            return Mathf.Clamp(trend / priceHistories.Count, -1f, 1f);
        }
        
        private float GetSeasonalModifier(ItemType itemType)
        {
            var currentSeason = GameManager.Instance.TimeManager.CurrentSeason;
            if (seasonalModifiers.ContainsKey(itemType) && seasonalModifiers[itemType].ContainsKey(currentSeason))
            {
                return seasonalModifiers[itemType][currentSeason];
            }
            return 1.0f;
        }
        
        private float GetEventModifier(ItemType itemType)
        {
            float modifier = 1.0f;
            foreach (var evt in activeMarketEvents)
            {
                if (evt.priceModifiers.ContainsKey(itemType))
                {
                    modifier *= evt.priceModifiers[itemType];
                }
            }
            return modifier;
        }

        // Event handlers
        private void OnPhaseChanged(PhaseChangedEvent evt)
        {
            // Minor price fluctuations during phase changes
            if (enablePriceFluctuations)
            {
                ApplyMinorFluctuations();
            }
        }

        private void OnSeasonChanged(SeasonChangedEvent evt)
        {
            Debug.Log($"[MarketSystem] Season changed to {evt.NewSeason}, applying seasonal price adjustments");
            ApplySeasonalEffects(evt.NewSeason);
        }

        private void OnDayChanged(DayChangedEvent evt)
        {
            // Daily market updates
            UpdateDailyPrices(evt.NewDay);
            CleanOldPriceHistory();
        }

        private void OnGameEventTriggered(GameEventTriggeredEvent evt)
        {
            ApplyEventEffects(evt);
        }

        private void OnTransactionCompleted(TransactionCompletedEvent evt)
        {
            // Adjust demand/supply based on player transactions
            AdjustMarketFromTransaction(evt);
        }

        // Public API methods
        public float GetCurrentPrice(ItemType itemType)
        {
            return marketPrices.ContainsKey(itemType) ? marketPrices[itemType].currentPrice : 0f;
        }

        public float GetBasePrice(ItemType itemType)
        {
            return itemType switch
            {
                ItemType.Fruit => 10f, // 短期取引、低価格
                ItemType.Potion => 50f, // 中価格、成長株
                ItemType.Weapon => 200f, // 高価格、安定株
                ItemType.Accessory => 75f, // 中価格、投機株
                ItemType.MagicBook => 300f, // 最高価格、債券
                ItemType.Gem => 150f, // 高価格、ハイリスク
                _ => 100f,
            };
        }

        public float GetItemVolatility(ItemType itemType)
        {
            return itemType switch
            {
                ItemType.Fruit => 0.3f, // 高い変動性（腐敗リスク）
                ItemType.Potion => 0.2f, // 中程度の変動性
                ItemType.Weapon => 0.1f, // 低い変動性（安定）
                ItemType.Accessory => 0.25f, // 高めの変動性（トレンド）
                ItemType.MagicBook => 0.05f, // 非常に低い変動性（安定）
                ItemType.Gem => 0.4f, // 最高の変動性（投機）
                _ => 0.2f,
            };
        }

        public List<PriceHistory> GetPriceHistory(ItemType itemType)
        {
            return priceHistories.ContainsKey(itemType)
                ? new List<PriceHistory>(priceHistories[itemType])
                : new List<PriceHistory>();
        }

        public MarketData GetMarketData(ItemType itemType)
        {
            return marketPrices.ContainsKey(itemType) ? marketPrices[itemType] : null;
        }

        // Price update methods
        private void UpdateDailyPrices(int currentDay)
        {
            foreach (var itemType in marketPrices.Keys.ToList())
            {
                var marketData = marketPrices[itemType];
                if (marketData.lastUpdateDay < currentDay)
                {
                    float oldPrice = marketData.currentPrice;
                    UpdateItemPrice(itemType, currentDay);
                    RecordPrice(itemType, marketData.currentPrice);

                    if (Math.Abs(oldPrice - marketData.currentPrice) > 0.01f)
                    {
                        TriggerPriceChangeEvent(itemType, oldPrice, marketData.currentPrice);
                    }
                }
            }
        }

        private void UpdateItemPrice(ItemType itemType, int currentDay)
        {
            var marketData = marketPrices[itemType];

            // Base volatility factor
            float volatilityFactor = UnityEngine.Random.Range(-marketData.volatility, marketData.volatility);

            // Apply demand/supply influence
            float demandSupplyRatio = marketData.demand / marketData.supply;
            float demandInfluence = (demandSupplyRatio - 1f) * 0.1f;

            // Calculate new price
            float priceChange = (volatilityFactor + demandInfluence) * fluctuationIntensity;
            marketData.currentPrice *= (1f + priceChange);

            // Apply bounds to prevent extreme prices
            float minPrice = marketData.basePrice * 0.3f;
            float maxPrice = marketData.basePrice * 3.0f;
            marketData.currentPrice = Mathf.Clamp(marketData.currentPrice, minPrice, maxPrice);

            // Gradually return demand/supply to equilibrium
            marketData.demand = Mathf.Lerp(marketData.demand, 1f, 0.1f);
            marketData.supply = Mathf.Lerp(marketData.supply, 1f, 0.1f);

            marketData.lastUpdateDay = currentDay;
        }

        private void ApplySeasonalEffects(Season season)
        {
            foreach (var itemType in marketPrices.Keys.ToList())
            {
                if (seasonalModifiers.ContainsKey(itemType) && seasonalModifiers[itemType].ContainsKey(season))
                {
                    var marketData = marketPrices[itemType];
                    float oldPrice = marketData.currentPrice;
                    float seasonalMultiplier = seasonalModifiers[itemType][season];

                    // Adjust demand based on seasonal multiplier
                    marketData.demand *= seasonalMultiplier;

                    // Apply immediate price adjustment (partial)
                    float priceAdjustment = (seasonalMultiplier - 1f) * 0.3f;
                    marketData.currentPrice *= (1f + priceAdjustment);

                    TriggerPriceChangeEvent(itemType, oldPrice, marketData.currentPrice);
                }
            }
        }

        private void ApplyMinorFluctuations()
        {
            foreach (var itemType in marketPrices.Keys.ToList())
            {
                var marketData = marketPrices[itemType];
                float oldPrice = marketData.currentPrice;

                // Small random fluctuation
                float minorChange = UnityEngine.Random.Range(-0.02f, 0.02f) * fluctuationIntensity;
                marketData.currentPrice *= (1f + minorChange);

                // Trigger event only if change is significant
                if (Math.Abs(oldPrice - marketData.currentPrice) > oldPrice * 0.01f)
                {
                    TriggerPriceChangeEvent(itemType, oldPrice, marketData.currentPrice);
                }
            }
        }

        private void ApplyEventEffects(GameEventTriggeredEvent evt)
        {
            for (int i = 0; i < evt.AffectedItems.Length; i++)
            {
                var itemType = evt.AffectedItems[i];
                var modifier = evt.PriceModifiers[i];

                if (marketPrices.ContainsKey(itemType))
                {
                    var marketData = marketPrices[itemType];
                    float oldPrice = marketData.currentPrice;

                    marketData.demand *= modifier;
                    marketData.currentPrice *= (1f + (modifier - 1f) * 0.5f);

                    TriggerPriceChangeEvent(itemType, oldPrice, marketData.currentPrice);
                }
            }

            // Store active event
            activeMarketEvents.Add(
                new MarketEvent
                {
                    eventName = evt.EventName,
                    affectedItems = evt.AffectedItems,
                    modifiers = evt.PriceModifiers,
                    remainingDuration = evt.Duration,
                }
            );
        }

        private void AdjustMarketFromTransaction(TransactionCompletedEvent evt)
        {
            if (marketPrices.ContainsKey(evt.ItemType))
            {
                var marketData = marketPrices[evt.ItemType];
                float influence = evt.Quantity * 0.01f; // Small influence per transaction

                if (evt.IsPurchase)
                {
                    // Player buying increases demand
                    marketData.demand += influence;
                }
                else
                {
                    // Player selling increases supply
                    marketData.supply += influence;
                }
            }
        }

        private void RecordPrice(ItemType itemType, float price)
        {
            var history = new PriceHistory
            {
                day = TimeManager.Instance?.CurrentDay ?? 1,
                price = price,
                timestamp = DateTime.Now,
            };

            priceHistories[itemType].Add(history);
        }

        private void CleanOldPriceHistory()
        {
            int currentDay = TimeManager.Instance?.CurrentDay ?? 1;
            int cutoffDay = currentDay - maxHistoryDays;

            foreach (var itemType in priceHistories.Keys.ToList())
            {
                priceHistories[itemType].RemoveAll(h => h.day < cutoffDay);
            }
        }

        private void TriggerPriceChangeEvent(ItemType itemType, float oldPrice, float newPrice)
        {
            OnPriceChanged?.Invoke(itemType, oldPrice, newPrice);
            EventBus.Publish(new PriceChangedEvent(itemType, oldPrice, newPrice));
        }

        // Debug methods
        public void LogMarketState()
        {
            Debug.Log("[MarketSystem] Current Market State:");
            foreach (var kvp in marketPrices)
            {
                var data = kvp.Value;
                Debug.Log(
                    $"  {kvp.Key}: {data.currentPrice:F2}G (Base: {data.basePrice:F2}G, "
                        + $"Demand: {data.demand:F2}, Supply: {data.supply:F2})"
                );
            }
        }
    }
}
