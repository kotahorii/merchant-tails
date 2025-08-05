using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Security.Cryptography;
using System.Text;
using System.Threading.Tasks;
using MerchantTails.Data;
using Newtonsoft.Json; // Unity 6の新しいJSON統合
using Unity.Burst;
using Unity.Collections;
using Unity.Jobs;
using UnityEngine;

namespace MerchantTails.Core
{
    /// <summary>
    /// セーブデータを管理するシステム
    /// JSON形式での永続化、暗号化、複数スロット対応
    /// </summary>
    public class SaveSystem : MonoBehaviour
    {
        private static SaveSystem instance;
        public static SaveSystem Instance => instance;

        [Header("Save Settings")]
        [SerializeField]
        private int maxSaveSlots = 3;

        [SerializeField]
        private bool enableEncryption = true;

        [SerializeField]
        private bool enableAutoSave = true;

        [SerializeField]
        private float autoSaveInterval = 300f; // 5分ごと

        [SerializeField]
        private bool enableBackup = true;

        [SerializeField]
        private int maxBackupCount = 3;

        private string savePath;
        private string backupPath;
        private string encryptionKey = "MerchantTails2025"; // 実際のプロダクトではより安全な方法で管理
        private float lastAutoSaveTime;
        private int currentSlot = 0;

        // セーブデータのキャッシュ
        private SaveData currentSaveData;
        private Dictionary<string, object> additionalData = new Dictionary<string, object>();

        public int CurrentSlot => currentSlot;
        public bool HasSaveData => File.Exists(GetSaveFilePath(currentSlot));
        public SaveData CurrentSaveData => currentSaveData;

        // イベント
        public event Action<int> OnSaveCompleted;
        public event Action<int> OnLoadCompleted;
        public event Action<string> OnSaveError;
        public event Action<string> OnLoadError;

        private void Awake()
        {
            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;
            DontDestroyOnLoad(gameObject);

            InitializePaths();
        }

        private void OnDestroy()
        {
            if (instance == this)
            {
                instance = null;
            }
        }

        private void Start()
        {
            SubscribeToEvents();
            LoadSettings();
        }

        private void Update()
        {
            // オートセーブの処理
            if (enableAutoSave && Time.time - lastAutoSaveTime > autoSaveInterval)
            {
                _ = AutoSaveAsync(); // Unity 6の非同期処理
                lastAutoSaveTime = Time.time;
            }
        }

        private void InitializePaths()
        {
            // セーブファイルのパス設定
#if UNITY_EDITOR
            savePath = Path.Combine(Application.dataPath, "../Saves");
            backupPath = Path.Combine(Application.dataPath, "../Saves/Backups");
#else
            savePath = Path.Combine(Application.persistentDataPath, "Saves");
            backupPath = Path.Combine(Application.persistentDataPath, "Saves/Backups");
#endif

            // ディレクトリが存在しない場合は作成
            if (!Directory.Exists(savePath))
            {
                Directory.CreateDirectory(savePath);
            }

            if (!Directory.Exists(backupPath))
            {
                Directory.CreateDirectory(backupPath);
            }
        }

        private void SubscribeToEvents()
        {
            // ゲームの重要なイベントで自動セーブ
            EventBus.Subscribe<DayChangedEvent>(OnDayChanged);
            EventBus.Subscribe<RankChangedEvent>(OnRankChanged);
            EventBus.Subscribe<AchievementUnlockedEvent>(OnAchievementUnlocked);
        }

        private void OnDayChanged(DayChangedEvent e)
        {
            if (enableAutoSave)
            {
                AutoSave();
            }
        }

        private void OnRankChanged(RankChangedEvent e)
        {
            if (enableAutoSave)
            {
                AutoSave();
            }
        }

        private void OnAchievementUnlockedEvent(AchievementUnlockedEvent e)
        {
            if (enableAutoSave)
            {
                AutoSave();
            }
        }

        /// <summary>
        /// 現在のゲーム状態をセーブ（Unity 6 Job System対応）
        /// </summary>
        public async Task<bool> SaveAsync(int slot = -1)
        {
            if (slot == -1)
                slot = currentSlot;

            try
            {
                // セーブデータの収集
                currentSaveData = CollectSaveData();
                currentSaveData.saveSlot = slot;
                currentSaveData.saveTime = DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss");
                currentSaveData.playTime = Time.time;
                currentSaveData.version = Application.version;

                // Unity 6の新しいNewtonsoft.Json統合を使用
                string json = JsonConvert.SerializeObject(currentSaveData, Formatting.Indented);
                byte[] jsonBytes = Encoding.UTF8.GetBytes(json);

                // Job Systemで圧縮と暗号化を並列処理
                byte[] processedData = jsonBytes;

                if (enableEncryption)
                {
                    processedData = await ProcessDataWithJobsAsync(jsonBytes, true);
                }

                // バックアップの作成
                if (enableBackup)
                {
                    CreateBackup(slot);
                }

                // ファイルに書き込み（非同期）
                string filePath = GetSaveFilePath(slot);
                await File.WriteAllBytesAsync(filePath, processedData);

                OnSaveCompleted?.Invoke(slot);
                EventBus.Publish(new SaveCompletedEvent(slot));
                ErrorHandler.LogInfo($"Save completed: Slot {slot}", "SaveSystem");
                return true;
            }
            catch (Exception e)
            {
                OnSaveError?.Invoke(e.Message);
                ErrorHandler.LogError($"Save failed: {e.Message}", "SaveSystem");
                return false;
            }
        }

        /// <summary>
        /// セーブデータをロード（Unity 6 Job System対応）
        /// </summary>
        public async Task<bool> LoadAsync(int slot = -1)
        {
            if (slot == -1)
                slot = currentSlot;

            string filePath = GetSaveFilePath(slot);

            if (!File.Exists(filePath))
            {
                OnLoadError?.Invoke($"Save file not found: Slot {slot}");
                return false;
            }

            try
            {
                // ファイルを読み込み（非同期）
                byte[] encryptedData = await File.ReadAllBytesAsync(filePath);

                // Job Systemで復号化と解凍を並列処理
                byte[] jsonBytes = encryptedData;

                if (enableEncryption)
                {
                    jsonBytes = await ProcessDataWithJobsAsync(encryptedData, false);
                }

                // Unity 6の新しいNewtonsoft.Json統合を使用
                string json = Encoding.UTF8.GetString(jsonBytes);
                currentSaveData = JsonConvert.DeserializeObject<SaveData>(json);
                currentSlot = slot;

                // ゲーム状態に適用
                ApplySaveData(currentSaveData);

                OnLoadCompleted?.Invoke(slot);
                EventBus.Publish(new LoadCompletedEvent(slot));
                ErrorHandler.LogInfo($"Load completed: Slot {slot}", "SaveSystem");
                return true;
            }
            catch (Exception e)
            {
                OnLoadError?.Invoke(e.Message);
                ErrorHandler.LogError($"Load failed: {e.Message}", "SaveSystem");
                return false;
            }
        }


        /// <summary>
        /// オートセーブ（Unity 6対応）
        /// </summary>
        private async Task AutoSaveAsync()
        {
            await SaveAsync(currentSlot);
            ErrorHandler.LogInfo("Auto save completed", "SaveSystem");
        }

        private void AutoSave()
        {
            _ = AutoSaveAsync();
        }

        /// <summary>
        /// セーブデータを収集
        /// </summary>
        private SaveData CollectSaveData()
        {
            var saveData = new SaveData();

            // プレイヤーデータ
            var playerData = GameManager.Instance?.PlayerData;
            if (playerData != null)
            {
                saveData.playerName = playerData.PlayerName;
                saveData.currentMoney = playerData.CurrentMoney;
                saveData.currentRank = playerData.CurrentRank;
                saveData.totalTransactions = playerData.TotalTransactions;
                saveData.totalProfit = playerData.TotalProfit;
            }

            // 時間データ
            var timeManager = TimeManager.Instance;
            if (timeManager != null)
            {
                saveData.currentDay = timeManager.CurrentDay;
                saveData.currentSeason = timeManager.CurrentSeason;
                saveData.currentPhase = timeManager.CurrentPhase;
                saveData.dayProgress = timeManager.DayProgress;
            }

            // 在庫データ
            var inventorySystem = InventorySystem.Instance;
            if (inventorySystem != null)
            {
                saveData.inventoryData = new SerializableInventory();
                var allItems = inventorySystem.GetAllItems();
                saveData.inventoryData.items = allItems.Select(kvp => new SerializableInventoryItem
                {
                    itemType = kvp.Key.ToString(),
                    count = kvp.Value.count,
                    condition = kvp.Value.condition,
                    averagePurchasePrice = kvp.Value.averagePurchasePrice,
                }).ToList();
            }

            // 市場データ
            var marketSystem = MarketSystem.Instance;
            if (marketSystem != null)
            {
                saveData.marketData = new SerializableMarket();
                saveData.marketData.priceHistories = new List<SerializablePriceHistory>();

                foreach (ItemType itemType in Enum.GetValues(typeof(ItemType)))
                {
                    var history = marketSystem.GetPriceHistory(itemType);
                    if (history != null && history.Count > 0)
                    {
                        saveData.marketData.priceHistories.Add(new SerializablePriceHistory
                        {
                            itemType = itemType.ToString(),
                            prices = history.ToList(),
                        });
                    }
                }
            }

            // 実績データ
            var achievementSystem = AchievementSystem.Instance;
            if (achievementSystem != null)
            {
                saveData.unlockedAchievements = achievementSystem.GetUnlockedAchievements()
                    .Select(a => a.id)
                    .ToList();
            }

            // 機能解放データ
            var featureUnlockSystem = FeatureUnlockSystem.Instance;
            if (featureUnlockSystem != null)
            {
                saveData.unlockedFeatures = featureUnlockSystem.GetUnlockedFeatures()
                    .Select(f => f.ToString())
                    .ToList();
            }

            // 銀行データ
            var bankSystem = BankSystem.Instance;
            if (bankSystem != null)
            {
                saveData.bankData = new SerializableBank();
                saveData.bankData.deposits = bankSystem.GetAllDeposits()
                    .Select(d => new SerializableDeposit
                    {
                        id = d.id,
                        amount = d.amount,
                        termDays = d.termDays,
                        interestRate = d.interestRate,
                        startDay = d.startDay,
                        maturityDay = d.maturityDay,
                        isMatured = d.isMatured,
                    }).ToList();
                saveData.bankData.totalInterestEarned = bankSystem.TotalInterestEarned;
            }

            // 投資データ
            var shopInvestmentSystem = ShopInvestmentSystem.Instance;
            if (shopInvestmentSystem != null)
            {
                saveData.shopInvestments = shopInvestmentSystem.GetAllInvestments()
                    .Select(i => new SerializableShopInvestment
                    {
                        upgradeType = i.upgradeType.ToString(),
                        level = i.level,
                        totalInvested = i.totalInvested,
                    }).ToList();
            }

            var merchantInvestmentSystem = MerchantInvestmentSystem.Instance;
            if (merchantInvestmentSystem != null)
            {
                saveData.merchantInvestments = merchantInvestmentSystem.GetAvailableMerchants()
                    .Select(m =>
                    {
                        var investment = merchantInvestmentSystem.GetInvestment(m.id);
                        if (investment != null && investment.isActive)
                        {
                            return new SerializableMerchantInvestment
                            {
                                merchantId = investment.merchantId,
                                totalInvested = investment.totalInvested,
                                totalDividends = investment.totalDividends,
                                lastInvestmentDay = investment.lastInvestmentDay,
                                lastDividendDay = investment.lastDividendDay,
                            };
                        }
                        return null;
                    })
                    .Where(i => i != null)
                    .ToList();
            }

            // チュートリアル進行状況
            var tutorialSystem = TutorialSystem.Instance;
            if (tutorialSystem != null)
            {
                saveData.tutorialCompleted = tutorialSystem.IsCompleted;
                saveData.currentTutorialStep = tutorialSystem.CurrentStep;
            }

            // 追加データ
            saveData.additionalData = JsonUtility.ToJson(additionalData);

            return saveData;
        }

        /// <summary>
        /// セーブデータを適用
        /// </summary>
        private void ApplySaveData(SaveData saveData)
        {
            // プレイヤーデータ
            var playerData = GameManager.Instance?.PlayerData;
            if (playerData != null)
            {
                playerData.SetPlayerName(saveData.playerName);
                playerData.SetMoney(saveData.currentMoney);
                playerData.SetRank(saveData.currentRank);
                // 統計データも復元
            }

            // 時間データ
            var timeManager = TimeManager.Instance;
            if (timeManager != null)
            {
                timeManager.LoadTimeData(
                    saveData.currentDay,
                    saveData.currentSeason,
                    saveData.currentPhase,
                    saveData.dayProgress
                );
            }

            // 在庫データ
            var inventorySystem = InventorySystem.Instance;
            if (inventorySystem != null && saveData.inventoryData != null)
            {
                inventorySystem.ClearInventory();
                foreach (var item in saveData.inventoryData.items)
                {
                    if (Enum.TryParse<ItemType>(item.itemType, out var itemType))
                    {
                        inventorySystem.LoadInventoryItem(
                            itemType,
                            item.count,
                            item.condition,
                            item.averagePurchasePrice
                        );
                    }
                }
            }

            // 市場データ
            var marketSystem = MarketSystem.Instance;
            if (marketSystem != null && saveData.marketData != null)
            {
                foreach (var history in saveData.marketData.priceHistories)
                {
                    if (Enum.TryParse<ItemType>(history.itemType, out var itemType))
                    {
                        marketSystem.LoadPriceHistory(itemType, history.prices);
                    }
                }
            }

            // 実績データ
            var achievementSystem = AchievementSystem.Instance;
            if (achievementSystem != null && saveData.unlockedAchievements != null)
            {
                achievementSystem.LoadUnlockedAchievements(saveData.unlockedAchievements);
            }

            // 機能解放データ
            var featureUnlockSystem = FeatureUnlockSystem.Instance;
            if (featureUnlockSystem != null && saveData.unlockedFeatures != null)
            {
                var features = saveData.unlockedFeatures
                    .Select(f => Enum.TryParse<GameFeature>(f, out var feature) ? feature : (GameFeature?)null)
                    .Where(f => f.HasValue)
                    .Select(f => f.Value)
                    .ToList();
                featureUnlockSystem.LoadUnlockedFeatures(features);
            }

            // その他のシステムも同様に復元...

            EventBus.Publish(new SaveDataAppliedEvent());
        }

        /// <summary>
        /// バックアップを作成
        /// </summary>
        private void CreateBackup(int slot)
        {
            string sourceFile = GetSaveFilePath(slot);
            if (!File.Exists(sourceFile))
                return;

            // バックアップファイル名（タイムスタンプ付き）
            string timestamp = DateTime.Now.ToString("yyyyMMdd_HHmmss");
            string backupFile = Path.Combine(backupPath, $"save_slot{slot}_{timestamp}.bak");

            File.Copy(sourceFile, backupFile, true);

            // 古いバックアップを削除
            CleanupOldBackups(slot);
        }

        /// <summary>
        /// 古いバックアップを削除
        /// </summary>
        private void CleanupOldBackups(int slot)
        {
            var backupFiles = Directory.GetFiles(backupPath, $"save_slot{slot}_*.bak")
                .OrderByDescending(f => new FileInfo(f).CreationTime)
                .Skip(maxBackupCount)
                .ToList();

            foreach (var file in backupFiles)
            {
                File.Delete(file);
            }
        }

        /// <summary>
        /// セーブファイルのパスを取得
        /// </summary>
        private string GetSaveFilePath(int slot)
        {
            return Path.Combine(savePath, $"save_slot{slot}.json");
        }

        /// <summary>
        /// 利用可能なセーブスロットを取得
        /// </summary>
        public List<SaveSlotInfo> GetSaveSlots()
        {
            var slots = new List<SaveSlotInfo>();

            for (int i = 0; i < maxSaveSlots; i++)
            {
                var info = new SaveSlotInfo { slot = i };
                string filePath = GetSaveFilePath(i);

                if (File.Exists(filePath))
                {
                    try
                    {
                        string json = File.ReadAllText(filePath);
                        if (enableEncryption)
                        {
                            json = Decrypt(json);
                        }

                        var data = JsonUtility.FromJson<SaveData>(json);
                        info.hasData = true;
                        info.playerName = data.playerName;
                        info.playTime = data.playTime;
                        info.saveTime = data.saveTime;
                        info.currentDay = data.currentDay;
                        info.currentMoney = data.currentMoney;
                    }
                    catch
                    {
                        info.hasData = false;
                    }
                }

                slots.Add(info);
            }

            return slots;
        }

        /// <summary>
        /// セーブデータを削除
        /// </summary>
        public void DeleteSave(int slot)
        {
            string filePath = GetSaveFilePath(slot);
            if (File.Exists(filePath))
            {
                File.Delete(filePath);
                ErrorHandler.LogInfo($"Save deleted: Slot {slot}", "SaveSystem");
            }
        }

        /// <summary>
        /// 設定を読み込み
        /// </summary>
        private void LoadSettings()
        {
            enableAutoSave = PlayerPrefs.GetInt("EnableAutoSave", 1) == 1;
            autoSaveInterval = PlayerPrefs.GetFloat("AutoSaveInterval", 300f);
            enableEncryption = PlayerPrefs.GetInt("EnableEncryption", 1) == 1;
            enableBackup = PlayerPrefs.GetInt("EnableBackup", 1) == 1;
        }

        /// <summary>
        /// 設定を保存
        /// </summary>
        public void SaveSettings()
        {
            PlayerPrefs.SetInt("EnableAutoSave", enableAutoSave ? 1 : 0);
            PlayerPrefs.SetFloat("AutoSaveInterval", autoSaveInterval);
            PlayerPrefs.SetInt("EnableEncryption", enableEncryption ? 1 : 0);
            PlayerPrefs.SetInt("EnableBackup", enableBackup ? 1 : 0);
            PlayerPrefs.Save();
        }

        /// <summary>
        /// Job Systemを使用したデータ処理
        /// </summary>
        private async Task<byte[]> ProcessDataWithJobsAsync(byte[] inputData, bool isEncrypting)
        {
            return await Task.Run(() =>
            {
                byte[] result = inputData;
                
                if (isEncrypting)
                {
                    // 圧縮
                    var compressedData = new NativeArray<byte>(inputData.Length * 2, Allocator.TempJob);
                    var uncompressedData = new NativeArray<byte>(inputData, Allocator.TempJob);
                    var compressedSize = new NativeArray<int>(1, Allocator.TempJob);
                    
                    var compressionJob = new SaveDataCompressionJob
                    {
                        uncompressedData = uncompressedData,
                        compressedData = compressedData,
                        compressedSize = compressedSize
                    };
                    
                    var compressionHandle = compressionJob.Schedule();
                    compressionHandle.Complete();
                    
                    int actualSize = compressedSize[0];
                    result = new byte[actualSize];
                    compressedData.Slice(0, actualSize).CopyTo(result);
                    
                    compressedData.Dispose();
                    uncompressedData.Dispose();
                    compressedSize.Dispose();
                    
                    // 暗号化
                    var plainData = new NativeArray<byte>(result, Allocator.TempJob);
                    var encryptedData = new NativeArray<byte>(result.Length, Allocator.TempJob);
                    
                    var encryptionJob = new SaveDataEncryptionJob
                    {
                        plainData = plainData,
                        encryptedData = encryptedData,
                        encryptionKey = (uint)encryptionKey.GetHashCode()
                    };
                    
                    var encryptionHandle = encryptionJob.Schedule(result.Length, 64);
                    encryptionHandle.Complete();
                    
                    encryptedData.CopyTo(result);
                    
                    plainData.Dispose();
                    encryptedData.Dispose();
                }
                else
                {
                    // 復号化
                    var encryptedData = new NativeArray<byte>(inputData, Allocator.TempJob);
                    var decryptedData = new NativeArray<byte>(inputData.Length, Allocator.TempJob);
                    
                    var decryptionJob = new SaveDataDecryptionJob
                    {
                        encryptedData = encryptedData,
                        decryptedData = decryptedData,
                        encryptionKey = (uint)encryptionKey.GetHashCode()
                    };
                    
                    var decryptionHandle = decryptionJob.Schedule(inputData.Length, 64);
                    decryptionHandle.Complete();
                    
                    decryptedData.CopyTo(result);
                    
                    encryptedData.Dispose();
                    decryptedData.Dispose();
                    
                    // 解凍
                    var compressedData = new NativeArray<byte>(result, Allocator.TempJob);
                    var decompressedData = new NativeArray<byte>(result.Length * 10, Allocator.TempJob); // 十分な大きさを確保
                    var decompressedSize = new NativeArray<int>(1, Allocator.TempJob);
                    
                    var decompressionJob = new SaveDataDecompressionJob
                    {
                        compressedData = compressedData,
                        compressedSize = result.Length,
                        decompressedData = decompressedData,
                        decompressedSize = decompressedSize
                    };
                    
                    var decompressionHandle = decompressionJob.Schedule();
                    decompressionHandle.Complete();
                    
                    int actualSize = decompressedSize[0];
                    result = new byte[actualSize];
                    decompressedData.Slice(0, actualSize).CopyTo(result);
                    
                    compressedData.Dispose();
                    decompressedData.Dispose();
                    decompressedSize.Dispose();
                }
                
                return result;
            });
        }

        // 暗号化・復号化（旧実装、削除予定）
        private string Encrypt(string plainText)
        {
            if (string.IsNullOrEmpty(plainText))
                return plainText;

            byte[] plainBytes = Encoding.UTF8.GetBytes(plainText);
            using (Aes aes = Aes.Create())
            {
                aes.Key = Encoding.UTF8.GetBytes(encryptionKey.PadRight(32).Substring(0, 32));
                aes.IV = new byte[16];

                using (var encryptor = aes.CreateEncryptor())
                {
                    byte[] encryptedBytes = encryptor.TransformFinalBlock(plainBytes, 0, plainBytes.Length);
                    return Convert.ToBase64String(encryptedBytes);
                }
            }
        }

        private string Decrypt(string cipherText)
        {
            if (string.IsNullOrEmpty(cipherText))
                return cipherText;

            byte[] cipherBytes = Convert.FromBase64String(cipherText);
            using (Aes aes = Aes.Create())
            {
                aes.Key = Encoding.UTF8.GetBytes(encryptionKey.PadRight(32).Substring(0, 32));
                aes.IV = new byte[16];

                using (var decryptor = aes.CreateDecryptor())
                {
                    byte[] decryptedBytes = decryptor.TransformFinalBlock(cipherBytes, 0, cipherBytes.Length);
                    return Encoding.UTF8.GetString(decryptedBytes);
                }
            }
        }

        /// <summary>
        /// 追加データを登録
        /// </summary>
        public void RegisterAdditionalData(string key, object data)
        {
            additionalData[key] = data;
        }

        /// <summary>
        /// 追加データを取得
        /// </summary>
        public T GetAdditionalData<T>(string key)
        {
            if (additionalData.ContainsKey(key))
            {
                return (T)additionalData[key];
            }
            return default(T);
        }

        /// <summary>
        /// 緊急セーブ
        /// </summary>
        public bool EmergencySave()
        {
            try
            {
                var saveData = CollectSaveData();
                saveData.saveSlot = -1; // 緊急セーブ用特別スロット
                saveData.saveTime = DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss");
                saveData.playTime = Time.time;
                saveData.version = Application.version;

                string json = JsonUtility.ToJson(saveData, true);
                string emergencyPath = Path.Combine(savePath, "emergency_save.json");
                File.WriteAllText(emergencyPath, json);

                ErrorHandler.LogInfo("Emergency save completed", "SaveSystem");
                return true;
            }
            catch (Exception e)
            {
                ErrorHandler.LogError("Emergency save failed", e, "SaveSystem");
                return false;
            }
        }

        /// <summary>
        /// バックアップからロード
        /// </summary>
        public bool LoadBackup()
        {
            try
            {
                // 最新のバックアップファイルを探す
                var backupFiles = Directory.GetFiles(backupPath, $"save_slot{currentSlot}_*.bak")
                    .OrderByDescending(f => new FileInfo(f).CreationTime)
                    .FirstOrDefault();

                if (string.IsNullOrEmpty(backupFiles))
                {
                    ErrorHandler.LogWarning("No backup files found", "SaveSystem");
                    return false;
                }

                // バックアップを現在のセーブファイルにコピー
                string targetPath = GetSaveFilePath(currentSlot);
                File.Copy(backupFiles, targetPath, true);

                // ロード実行
                return Load(currentSlot);
            }
            catch (Exception e)
            {
                ErrorHandler.LogError("Load backup failed", e, "SaveSystem");
                return false;
            }
        }

        /// <summary>
        /// クイックセーブ
        /// </summary>
        public void QuickSave()
        {
            Save(currentSlot);
        }

        /// <summary>
        /// クイックロード
        /// </summary>
        public void QuickLoad()
        {
            Load(currentSlot);
        }

        /// <summary>
        /// すべてのセーブデータを削除
        /// </summary>
        public void DeleteAllSaves()
        {
            for (int i = 0; i < maxSaveSlots; i++)
            {
                DeleteSave(i);
            }

            // 緊急セーブも削除
            string emergencyPath = Path.Combine(savePath, "emergency_save.json");
            if (File.Exists(emergencyPath))
            {
                File.Delete(emergencyPath);
            }

            // バックアップも削除
            if (Directory.Exists(backupPath))
            {
                var backupFiles = Directory.GetFiles(backupPath, "*.bak");
                foreach (var file in backupFiles)
                {
                    File.Delete(file);
                }
            }

            ErrorHandler.LogInfo("All save data deleted", "SaveSystem");
        }
    }

    /// <summary>
    /// セーブデータ構造
    /// </summary>
    [Serializable]
    public class SaveData
    {
        // メタデータ
        public int saveSlot;
        public string saveTime;
        public float playTime;
        public string version;

        // プレイヤーデータ
        public string playerName;
        public int currentMoney;
        public MerchantRank currentRank;
        public int totalTransactions;
        public float totalProfit;

        // 時間データ
        public int currentDay;
        public Season currentSeason;
        public DayPhase currentPhase;
        public float dayProgress;

        // ゲームデータ
        public SerializableInventory inventoryData;
        public SerializableMarket marketData;
        public SerializableBank bankData;
        public List<string> unlockedAchievements;
        public List<string> unlockedFeatures;
        public List<SerializableShopInvestment> shopInvestments;
        public List<SerializableMerchantInvestment> merchantInvestments;

        // チュートリアル
        public bool tutorialCompleted;
        public int currentTutorialStep;

        // 追加データ（拡張用）
        public string additionalData;
    }

    /// <summary>
    /// セーブスロット情報
    /// </summary>
    [Serializable]
    public class SaveSlotInfo
    {
        public int slot;
        public bool hasData;
        public string playerName;
        public float playTime;
        public string saveTime;
        public int currentDay;
        public int currentMoney;
    }

    // シリアライズ可能なデータ構造
    [Serializable]
    public class SerializableInventory
    {
        public List<SerializableInventoryItem> items;
    }

    [Serializable]
    public class SerializableInventoryItem
    {
        public string itemType;
        public int count;
        public float condition;
        public float averagePurchasePrice;
    }

    [Serializable]
    public class SerializableMarket
    {
        public List<SerializablePriceHistory> priceHistories;
    }

    [Serializable]
    public class SerializablePriceHistory
    {
        public string itemType;
        public List<float> prices;
    }

    [Serializable]
    public class SerializableBank
    {
        public List<SerializableDeposit> deposits;
        public float totalInterestEarned;
    }

    [Serializable]
    public class SerializableDeposit
    {
        public string id;
        public float amount;
        public int termDays;
        public float interestRate;
        public int startDay;
        public int maturityDay;
        public bool isMatured;
    }

    [Serializable]
    public class SerializableShopInvestment
    {
        public string upgradeType;
        public int level;
        public float totalInvested;
    }

    [Serializable]
    public class SerializableMerchantInvestment
    {
        public string merchantId;
        public float totalInvested;
        public float totalDividends;
        public int lastInvestmentDay;
        public int lastDividendDay;
    }

    // イベント
    public class SaveCompletedEvent : BaseGameEvent
    {
        public int Slot { get; }

        public SaveCompletedEvent(int slot)
        {
            Slot = slot;
        }
    }

    public class LoadCompletedEvent : BaseGameEvent
    {
        public int Slot { get; }

        public LoadCompletedEvent(int slot)
        {
            Slot = slot;
        }
    }

    public class SaveDataAppliedEvent : BaseGameEvent { }
}
