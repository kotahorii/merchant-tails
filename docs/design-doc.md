# マーチャントテイル ～商人物語～ 設計書

## 1. システム概要

### 1.1 アーキテクチャ概要

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Presentation  │    │    Business     │    │      Data       │
│     Layer       │◄──►│     Logic       │◄──►│     Layer       │
│                 │    │     Layer       │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
│ UI/Input Manager│    │ Game Manager    │    │ Save System     │
│ Scene Manager   │    │ Market System   │    │ Data Models     │
│ Audio Manager   │    │ Inventory System│    │ Configuration   │
└─────────────────┘    │ Event System    │    └─────────────────┘
                       │ Tutorial System │
                       └─────────────────┘
```

### 1.2 技術スタック

- **エンジン**: Unity 2022.3 LTS
- **言語**: C#
- **アーキテクチャパターン**: MVC + Observer Pattern
- **データ永続化**: JSON + PlayerPrefs
- **状態管理**: Finite State Machine

## 2. コアシステム設計

### 2.1 ゲームマネージャー

```csharp
public class GameManager : MonoBehaviour
{
    public GameState CurrentState { get; private set; }
    public TimeManager TimeManager { get; private set; }
    public PlayerData PlayerData { get; private set; }

    public void ChangeState(GameState newState);
    public void SaveGame();
    public void LoadGame();
}

public enum GameState
{
    MainMenu,
    Tutorial,
    Shopping,
    StoreManagement,
    MarketView,
    Paused
}
```

### 2.2 時間管理システム

```csharp
public class TimeManager : MonoBehaviour
{
    public Season CurrentSeason { get; private set; }
    public int CurrentDay { get; private set; }
    public DayPhase CurrentPhase { get; private set; }

    public event Action<DayPhase> OnPhaseChanged;
    public event Action<Season> OnSeasonChanged;

    private void AdvanceTime();
    private void TriggerPhaseEvents();
}

public enum Season { Spring, Summer, Autumn, Winter }
public enum DayPhase { Morning, Afternoon, Evening, Night }
```

### 2.3 市場システム

```csharp
public class MarketSystem : MonoBehaviour
{
    private Dictionary<ItemType, MarketData> marketPrices;
    private EventSystem eventSystem;

    public float GetCurrentPrice(ItemType itemType);
    public float GetBasePrice(ItemType itemType);
    public List<PriceHistory> GetPriceHistory(ItemType itemType);

    private void UpdatePrices();
    private void ApplySeasonalEffects();
    private void ApplyEventEffects();
}

[System.Serializable]
public class MarketData
{
    public float basePrice;
    public float currentPrice;
    public float volatility;
    public List<PriceHistory> history;
    public SeasonalModifier seasonalModifier;
}
```

### 2.4 在庫管理システム

```csharp
public class InventorySystem : MonoBehaviour
{
    private Dictionary<ItemType, InventorySlot> inventory;

    public bool AddItem(ItemType itemType, int quantity, float purchasePrice);
    public bool RemoveItem(ItemType itemType, int quantity);
    public bool TransferToMarket(ItemType itemType, int quantity);
    public bool TransferToStore(ItemType itemType, int quantity);

    public int GetTotalQuantity(ItemType itemType);
    public int GetStoreQuantity(ItemType itemType);
    public int GetMarketQuantity(ItemType itemType);
}

[System.Serializable]
public class InventorySlot
{
    public ItemType itemType;
    public int storeQuantity;
    public int marketQuantity;
    public float averagePurchasePrice;
    public DateTime lastUpdated;
}
```

## 3. データモデル設計

### 3.1 プレイヤーデータ

```csharp
[System.Serializable]
public class PlayerData
{
    public string playerName;
    public float gold;
    public int currentDay;
    public Season currentSeason;
    public MerchantRank rank;
    public ShopData shopData;
    public Dictionary<ItemType, InventorySlot> inventory;
    public List<Transaction> transactionHistory;
    public TutorialProgress tutorialProgress;
    public AchievementData achievements;
}

public enum MerchantRank
{
    Apprentice,    // 見習い (～1,000G)
    Skilled,       // 一人前 (～5,000G)
    Veteran,       // ベテラン (～10,000G)
    Master         // マスター (10,000G～)
}
```

### 3.2 商品データ

```csharp
[System.Serializable]
public class ItemData
{
    public ItemType itemType;
    public string itemName;
    public string description;
    public Sprite icon;
    public float basePrice;
    public float volatility;
    public int shelfLife;
    public SeasonalModifier seasonalModifier;
    public EventSensitivity eventSensitivity;
    public InvestmentType investmentType;
}

public enum ItemType
{
    Fruit,       // くだもの (短期投資)
    Potion,      // ポーション (成長株)
    Weapon,      // 武器 (優良株)
    Accessory,   // アクセサリー (投機株)
    MagicBook,   // 魔法書 (債券)
    Gem          // 宝石 (ハイリスク投資)
}

public enum InvestmentType
{
    ShortTerm,   // 短期投資
    GrowthStock, // 成長株
    BlueChip,    // 優良株
    Speculative, // 投機株
    Bond,        // 債券
    HighRisk     // ハイリスク投資
}
```

### 3.3 イベントシステム

```csharp
[System.Serializable]
public class GameEvent
{
    public string eventId;
    public string eventName;
    public string description;
    public EventType eventType;
    public int duration;
    public Dictionary<ItemType, float> priceModifiers;
    public Dictionary<ItemType, float> demandModifiers;
    public List<string> prerequisiteEvents;
    public bool isRepeatable;
}

public enum EventType
{
    Regular,     // 定期イベント (給料日、ギルド定例会)
    Seasonal,    // 季節イベント
    Major,       // 大型イベント (ドラゴン討伐、収穫祭)
    Random       // ランダムイベント
}

public class EventSystem : MonoBehaviour
{
    private Queue<GameEvent> scheduledEvents;
    private List<GameEvent> activeEvents;

    public void ScheduleEvent(GameEvent gameEvent, int daysUntilEvent);
    public void TriggerEvent(GameEvent gameEvent);
    public void EndEvent(string eventId);
    public List<GameEvent> GetActiveEvents();
}
```

## 4. UI/UX システム設計

### 4.1 画面遷移

```csharp
public class SceneManager : MonoBehaviour
{
    private Stack<UIScreen> screenStack;
    private Dictionary<ScreenType, UIScreen> screens;

    public void PushScreen(ScreenType screenType);
    public void PopScreen();
    public void ReplaceScreen(ScreenType screenType);
    public void ClearStack();
}

public enum ScreenType
{
    MainMenu,
    Shop,
    Market,
    Inventory,
    Journal,
    Settings,
    Tutorial
}
```

### 4.2 情報表示システム

```csharp
public class InfoDisplaySystem : MonoBehaviour
{
    public void ShowPriceChart(ItemType itemType, int dayRange);
    public void ShowProfitLossGraph();
    public void ShowInventoryStatus();
    public void ShowMarketTrends();

    // 段階的情報開示
    public bool IsFeatureUnlocked(FeatureType feature, MerchantRank playerRank);
}

public enum FeatureType
{
    BasicTrading,
    PriceForecasting,
    AdvancedAnalytics,
    InvestmentOptions
}
```

## 5. チュートリアルシステム

### 5.1 段階的学習設計

```csharp
public class TutorialSystem : MonoBehaviour
{
    private List<TutorialStep> tutorialSteps;
    private int currentStepIndex;

    public void StartTutorial();
    public void NextStep();
    public void SkipTutorial();
    public bool IsTutorialComplete();
}

[System.Serializable]
public class TutorialStep
{
    public string stepId;
    public string title;
    public string description;
    public TutorialType type;
    public object targetObject;
    public bool isCompleted;
}

public enum TutorialType
{
    Introduction,
    BasicPurchase,
    BasicSale,
    InventoryManagement,
    MarketAnalysis,
    SeasonalEffects,
    EventResponse,
    InvestmentBasics
}
```

## 6. セーブ/ロードシステム

### 6.1 データ永続化

```csharp
public class SaveSystem : MonoBehaviour
{
    private const string SAVE_KEY = "MerchantTales_SaveData";

    public void SaveGame(PlayerData playerData);
    public PlayerData LoadGame();
    public bool HasSaveData();
    public void DeleteSaveData();

    // オートセーブ
    private void AutoSave();
    private IEnumerator AutoSaveCoroutine();
}

[System.Serializable]
public class SaveData
{
    public PlayerData playerData;
    public GameSettings gameSettings;
    public DateTime saveTimestamp;
    public string gameVersion;
}
```

## 7. パフォーマンス最適化

### 7.1 メモリ管理

```csharp
public class ResourceManager : MonoBehaviour
{
    private Dictionary<string, Object> loadedAssets;
    private Queue<string> assetQueue;

    public T LoadAsset<T>(string assetPath) where T : Object;
    public void UnloadAsset(string assetPath);
    public void PreloadAssets(List<string> assetPaths);
}
```

### 7.2 更新頻度最適化

```csharp
public class UpdateManager : MonoBehaviour
{
    private List<IUpdatable> frameUpdates;
    private List<IUpdatable> fixedUpdates;
    private List<IUpdatable> slowUpdates;

    // フレーム毎、物理更新、低頻度更新を分離
    private void Update();
    private void FixedUpdate();
    private IEnumerator SlowUpdateCoroutine();
}
```

## 8. モジュラー設計

### 8.1 システム間通信

```csharp
public class EventBus : MonoBehaviour
{
    private Dictionary<Type, List<IEventHandler>> eventHandlers;

    public void Subscribe<T>(IEventHandler<T> handler) where T : IEvent;
    public void Unsubscribe<T>(IEventHandler<T> handler) where T : IEvent;
    public void Publish<T>(T eventData) where T : IEvent;
}

// イベント例
public class PriceChangedEvent : IEvent
{
    public ItemType ItemType { get; set; }
    public float OldPrice { get; set; }
    public float NewPrice { get; set; }
}
```

## 9. テスト戦略

### 9.1 単体テスト

```csharp
[TestFixture]
public class MarketSystemTests
{
    [Test]
    public void PriceCalculation_WithSeasonalModifier_ReturnsCorrectPrice();

    [Test]
    public void EventEffect_OnDragonSlaying_IncreasesWeaponDemand();

    [Test]
    public void InventoryManagement_WhenItemExpires_RemovesFromInventory();
}
```

### 9.2 統合テスト

```csharp
[TestFixture]
public class GameFlowTests
{
    [Test]
    public void CompleteDayFlow_FromMorningToNight_UpdatesAllSystems();

    [Test]
    public void SaveLoadCycle_PreservesAllGameState();
}
```

## 10. 拡張性設計

### 10.1 モジュール化

```csharp
public interface IGameModule
{
    void Initialize();
    void Update();
    void Cleanup();
    bool IsEnabled { get; }
}

public class ModuleManager : MonoBehaviour
{
    private List<IGameModule> modules;

    public void RegisterModule(IGameModule module);
    public void EnableModule<T>() where T : IGameModule;
    public void DisableModule<T>() where T : IGameModule;
}
```

### 10.2 設定システム

```csharp
[System.Serializable]
public class GameSettings
{
    public float masterVolume;
    public bool tutorialEnabled;
    public DifficultyLevel difficulty;
    public LanguageType language;
    public bool autoSaveEnabled;
    public int autoSaveInterval;
}

public enum DifficultyLevel
{
    Easy,    // 価格変動緩やか
    Normal,  // 標準
    Hard     // 価格変動激しい
}
```

## 11. デバッグ・開発支援

### 11.1 デバッグシステム

```csharp
public class DebugManager : MonoBehaviour
{
    [SerializeField] private bool debugMode;

    public void SetGold(float amount);
    public void AdvanceDay(int days);
    public void TriggerEvent(string eventId);
    public void UnlockAllFeatures();
    public void ResetPlayerData();
}
```

## 12. ローカライゼーション

### 12.1 多言語対応

```csharp
public class LocalizationManager : MonoBehaviour
{
    private Dictionary<string, Dictionary<LanguageType, string>> localizedText;

    public string GetLocalizedText(string key);
    public void SetLanguage(LanguageType language);
    public void LoadLanguageFile(LanguageType language);
}

public enum LanguageType
{
    Japanese,
    English
}
```

この Design Document は、PRD で定義された要件を技術的に実装するための詳細な設計指針を提供します。各システムは独立性を保ちながら、イベントバスを通じて効率的に連携し、拡張性とメンテナンス性を確保しています。
