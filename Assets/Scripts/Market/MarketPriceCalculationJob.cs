using Unity.Burst;
using Unity.Collections;
using Unity.Jobs;
using Unity.Mathematics;

namespace MerchantTails.Market
{
    /// <summary>
    /// Job Systemを使用した価格計算の並列処理
    /// Unity 6の新機能を活用してパフォーマンスを向上
    /// </summary>
    [BurstCompile]
    public struct MarketPriceCalculationJob : IJobParallelFor
    {
        // 入力データ
        [ReadOnly]
        public NativeArray<float> basePrices;
        [ReadOnly]
        public NativeArray<float> volatilities;
        [ReadOnly]
        public NativeArray<float> demands;
        [ReadOnly]
        public NativeArray<float> supplies;
        [ReadOnly]
        public NativeArray<float> seasonalModifiers;
        [ReadOnly]
        public NativeArray<float> eventModifiers;
        [ReadOnly]
        public float globalMarketTrend;
        [ReadOnly]
        public uint randomSeed;
        [ReadOnly]
        public float deltaTime;

        // 出力データ
        [WriteOnly]
        public NativeArray<float> calculatedPrices;
        [WriteOnly]
        public NativeArray<float> priceChanges;

        public void Execute(int index)
        {
            // ランダム値生成（Burst対応）
            var random = new Unity.Mathematics.Random(randomSeed + (uint)index);

            // 基本価格
            float basePrice = basePrices[index];
            float volatility = volatilities[index];

            // 需給バランスの計算
            float demandSupplyRatio = demands[index] / math.max(supplies[index], 0.1f);
            float demandModifier = math.clamp(demandSupplyRatio, 0.5f, 2.0f);

            // ランダム変動
            float randomFluctuation = 1.0f + (random.NextFloat(-1f, 1f) * volatility * 0.1f);

            // 季節とイベントの影響
            float seasonalEffect = seasonalModifiers[index];
            float eventEffect = eventModifiers[index];

            // グローバルトレンドの影響（アイテムタイプごとに異なる影響度）
            float trendInfluence = 1.0f + (globalMarketTrend * volatility * 0.2f);

            // 最終価格の計算
            float newPrice = basePrice * demandModifier * randomFluctuation *
                           seasonalEffect * eventEffect * trendInfluence;

            // 価格の範囲制限（基本価格の0.3倍～3倍）
            newPrice = math.clamp(newPrice, basePrice * 0.3f, basePrice * 3.0f);

            // 価格変動の滑らかさ（急激な変動を抑制）
            float currentPrice = calculatedPrices[index];
            if (currentPrice > 0)
            {
                float maxChangeRate = volatility * deltaTime;
                float priceChange = newPrice - currentPrice;
                priceChange = math.clamp(priceChange, -currentPrice * maxChangeRate, currentPrice * maxChangeRate);
                newPrice = currentPrice + priceChange;
            }

            // 結果の保存
            calculatedPrices[index] = newPrice;
            priceChanges[index] = (newPrice - currentPrice) / math.max(currentPrice, 0.01f);
        }
    }

    /// <summary>
    /// 価格履歴の更新を並列処理するJob
    /// </summary>
    [BurstCompile]
    public struct PriceHistoryUpdateJob : IJob
    {
        [ReadOnly]
        public NativeArray<float> newPrices;
        [ReadOnly]
        public int currentDay;

        // 価格履歴（循環バッファ）
        public NativeArray<float> priceHistory;
        public NativeArray<int> historyDays;

        [ReadOnly]
        public int historySize;
        [ReadOnly]
        public int itemCount;

        public void Execute()
        {
            for (int itemIndex = 0; itemIndex < itemCount; itemIndex++)
            {
                int baseIndex = itemIndex * historySize;
                int writeIndex = currentDay % historySize;

                priceHistory[baseIndex + writeIndex] = newPrices[itemIndex];
                historyDays[baseIndex + writeIndex] = currentDay;
            }
        }
    }

    /// <summary>
    /// 価格トレンド分析Job
    /// </summary>
    [BurstCompile]
    public struct PriceTrendAnalysisJob : IJobParallelFor
    {
        [ReadOnly]
        public NativeArray<float> priceHistory;
        [ReadOnly]
        public int historySize;
        [ReadOnly]
        public int analysisWindow; // 分析する日数

        [WriteOnly]
        public NativeArray<float> trendSlopes; // トレンドの傾き
        [WriteOnly]
        public NativeArray<float> volatilityScores; // ボラティリティスコア

        public void Execute(int itemIndex)
        {
            int baseIndex = itemIndex * historySize;
            float sumX = 0;
            float sumY = 0;
            float sumXY = 0;
            float sumX2 = 0;
            float validDataPoints = 0;

            // 最小二乗法でトレンドラインを計算
            for (int i = 0; i < analysisWindow && i < historySize; i++)
            {
                float price = priceHistory[baseIndex + i];
                if (price > 0)
                {
                    sumX += i;
                    sumY += price;
                    sumXY += i * price;
                    sumX2 += i * i;
                    validDataPoints++;
                }
            }

            if (validDataPoints > 1)
            {
                float slope = (validDataPoints * sumXY - sumX * sumY) /
                            (validDataPoints * sumX2 - sumX * sumX);
                trendSlopes[itemIndex] = slope;

                // ボラティリティの計算
                float avgPrice = sumY / validDataPoints;
                float variance = 0;

                for (int i = 0; i < analysisWindow && i < historySize; i++)
                {
                    float price = priceHistory[baseIndex + i];
                    if (price > 0)
                    {
                        float diff = price - avgPrice;
                        variance += diff * diff;
                    }
                }

                volatilityScores[itemIndex] = math.sqrt(variance / validDataPoints) / avgPrice;
            }
            else
            {
                trendSlopes[itemIndex] = 0;
                volatilityScores[itemIndex] = 0;
            }
        }
    }
}
