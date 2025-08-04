using System;
using MerchantTails.Data;
using MerchantTails.Systems;
using UnityEngine;

namespace MerchantTails.Core
{
    public class GameManager : MonoBehaviour
    {
        [Header("Game State")]
        [SerializeField]
        private GameState currentState = GameState.MainMenu;

        [Header("Core Systems")]
        [SerializeField]
        private TimeManager timeManager;

        [SerializeField]
        private PlayerData playerData;

        [SerializeField]
        private AssetCalculator assetCalculator;

        [SerializeField]
        private FeatureUnlockSystem featureUnlockSystem;

        [Header("Tutorial")]
        [SerializeField]
        private bool tutorialCompleted = false;

        public static GameManager Instance { get; private set; }

        public GameState CurrentState
        {
            get => currentState;
            private set => currentState = value;
        }

        public TimeManager TimeManager => timeManager;
        public PlayerData PlayerData => playerData;
        public AssetCalculator AssetCalculator => assetCalculator;
        public FeatureUnlockSystem FeatureUnlockSystem => featureUnlockSystem;
        public bool IsTutorialCompleted => tutorialCompleted;

        public event Action<GameState> OnGameStateChanged;

        private void Awake()
        {
            if (Instance == null)
            {
                Instance = this;
                DontDestroyOnLoad(gameObject);
                InitializeGame();
            }
            else
            {
                Destroy(gameObject);
            }
        }

        private void Start()
        {
            ChangeState(GameState.MainMenu);
        }

        private void InitializeGame()
        {
            Debug.Log("[GameManager] Initializing Merchant Tales...");

            // Initialize core systems
            if (timeManager == null)
                timeManager = FindObjectOfType<TimeManager>();

            if (playerData == null)
                playerData = CreateDefaultPlayerData();

            if (assetCalculator == null)
                assetCalculator = FindObjectOfType<AssetCalculator>();

            if (featureUnlockSystem == null)
                featureUnlockSystem = FindObjectOfType<FeatureUnlockSystem>();
        }

        public void ChangeState(GameState newState)
        {
            if (currentState == newState)
                return;

            Debug.Log($"[GameManager] State changed: {currentState} -> {newState}");

            GameState previousState = currentState;
            currentState = newState;

            OnGameStateChanged?.Invoke(newState);

            HandleStateTransition(previousState, newState);
        }

        private void HandleStateTransition(GameState from, GameState to)
        {
            // Handle state exit logic
            switch (from)
            {
                case GameState.Shopping:
                case GameState.StoreManagement:
                case GameState.MarketView:
                    // Auto-save when leaving gameplay states
                    SaveGame();
                    break;
            }

            // Handle state entry logic
            switch (to)
            {
                case GameState.MainMenu:
                    HandleMainMenuEntry();
                    break;
                case GameState.Tutorial:
                    HandleTutorialEntry();
                    break;
                case GameState.Shopping:
                case GameState.StoreManagement:
                case GameState.MarketView:
                    HandleGameplayEntry();
                    break;
            }
        }

        private void HandleMainMenuEntry()
        {
            Debug.Log("[GameManager] Entering Main Menu");
            // UI loading, music transition, etc.
        }

        private void HandleTutorialEntry()
        {
            Debug.Log("[GameManager] Starting Tutorial");
            // Tutorial initialization
        }

        private void HandleGameplayEntry()
        {
            Debug.Log("[GameManager] Entering Gameplay");
            // Ensure systems are ready for gameplay
        }

        public void SaveGame()
        {
            Debug.Log("[GameManager] Saving game...");
            // TODO: Implement save system
            PlayerPrefs.SetInt("TutorialCompleted", tutorialCompleted ? 1 : 0);
            PlayerPrefs.Save();
        }

        public void LoadGame()
        {
            Debug.Log("[GameManager] Loading game...");
            // TODO: Implement load system
            tutorialCompleted = PlayerPrefs.GetInt("TutorialCompleted", 0) == 1;
        }

        public void SetTutorialCompleted(bool completed)
        {
            tutorialCompleted = completed;
            PlayerPrefs.SetInt("TutorialCompleted", completed ? 1 : 0);
            PlayerPrefs.Save();
        }

        public bool HasSaveData()
        {
            // Check if save data exists
            return PlayerPrefs.HasKey("SaveData_Exists") && PlayerPrefs.GetInt("SaveData_Exists", 0) == 1;
        }

        public void PauseGame()
        {
            ChangeState(GameState.Paused);
            Time.timeScale = 0f;
        }

        public void ResumeGame()
        {
            // Return to previous non-paused state
            // For now, default to StoreManagement
            ChangeState(GameState.StoreManagement);
            Time.timeScale = 1f;
        }

        public void QuitGame()
        {
            Debug.Log("[GameManager] Quitting game...");
            SaveGame();

#if UNITY_EDITOR
            UnityEditor.EditorApplication.isPlaying = false;
#else
            Application.Quit();
#endif
        }

        private PlayerData CreateDefaultPlayerData()
        {
            var newPlayerData = ScriptableObject.CreateInstance<PlayerData>();
            // TODO: Initialize with default values
            return newPlayerData;
        }

        private void OnApplicationPause(bool pauseStatus)
        {
            if (pauseStatus)
            {
                SaveGame();
            }
        }

        private void OnApplicationFocus(bool hasFocus)
        {
            if (!hasFocus)
            {
                SaveGame();
            }
        }
    }
}
