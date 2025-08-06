using System;
using System.Collections;
using MerchantTails.Data;
using UnityEngine;

namespace MerchantTails.Core
{
    /// <summary>
    /// ゲーム内時間の管理を行うシステム
    /// 季節、日数、時間フェーズを制御し、関連するイベントを発行する
    /// </summary>
    public class TimeManager : MonoBehaviour
    {
        [Header("Time Settings")]
        [SerializeField]
        private float realTimePerGameDay = 120f; // 実時間2分 = ゲーム内1日

        [SerializeField]
        private bool autoAdvanceTime = true;

        [SerializeField]
        private bool pauseTimeProgression = false;

        [Header("Current Time State")]
        [SerializeField]
        private Season currentSeason = Season.Spring;

        [SerializeField]
        private int currentDay = 1;

        [SerializeField]
        private DayPhase currentPhase = DayPhase.Morning;

        [SerializeField]
        private int currentYear = 1;

        [Header("Season Settings")]
        [SerializeField]
        private int daysPerSeason = 30;

        // Public Properties
        public Season CurrentSeason
        {
            get => currentSeason;
            private set => currentSeason = value;
        }

        public int CurrentDay
        {
            get => currentDay;
            private set => currentDay = value;
        }

        public DayPhase CurrentPhase
        {
            get => currentPhase;
            private set => currentPhase = value;
        }

        public int CurrentYear
        {
            get => currentYear;
            private set => currentYear = value;
        }

        public float TimeSpeedMultiplier { get; set; } = 1f;

        /// <summary>
        /// 現在の日の進行度（0.0 = 日の始まり、1.0 = 日の終わり）
        /// </summary>
        public float DayProgress
        {
            get
            {
                float phaseProgress = (int)currentPhase * 0.25f; // Each phase is 25% of a day
                float currentPhaseProgress = phaseTimer / phaseDuration * 0.25f;
                return Mathf.Clamp01(phaseProgress + currentPhaseProgress);
            }
        }

        // Events
        public event Action<DayPhase> OnPhaseChanged;
        public event Action<Season> OnSeasonChanged;
        public event Action<int> OnDayChanged;
        public event Action<int> OnYearChanged;

        // Internal state
        private float phaseTimer = 0f;
        private float phaseDuration => (realTimePerGameDay / 4f) / TimeSpeedMultiplier; // 4 phases per day
        private Coroutine timeCoroutine;

        public static TimeManager Instance { get; private set; }

        private void Awake()
        {
            if (Instance == null)
            {
                Instance = this;
                DontDestroyOnLoad(gameObject);
            }
            else
            {
                Destroy(gameObject);
                return;
            }
        }

        private void Start()
        {
            InitializeTimeSystem();
            if (autoAdvanceTime)
            {
                StartTimeProgression();
            }
        }

        private void InitializeTimeSystem()
        {
            Debug.Log(
                $"[TimeManager] Initializing time system - Day: {currentDay}, Season: {currentSeason}, Phase: {currentPhase}"
            );

            // Publish initial state events
            EventBus.Publish(new PhaseChangedEvent(currentPhase, currentPhase, currentDay));
            EventBus.Publish(new SeasonChangedEvent(currentSeason, currentSeason, currentYear));
        }

        public void StartTimeProgression()
        {
            if (timeCoroutine != null)
            {
                StopCoroutine(timeCoroutine);
            }

            timeCoroutine = StartCoroutine(TimeProgressionCoroutine());
            Debug.Log("[TimeManager] Time progression started");
        }

        public void StopTimeProgression()
        {
            if (timeCoroutine != null)
            {
                StopCoroutine(timeCoroutine);
                timeCoroutine = null;
            }
            Debug.Log("[TimeManager] Time progression stopped");
        }

        public void PauseTime()
        {
            pauseTimeProgression = true;
            Debug.Log("[TimeManager] Time progression paused");
        }

        public void ResumeTime()
        {
            pauseTimeProgression = false;
            Debug.Log("[TimeManager] Time progression resumed");
        }

        private IEnumerator TimeProgressionCoroutine()
        {
            while (true)
            {
                if (!pauseTimeProgression)
                {
                    phaseTimer += Time.deltaTime;

                    if (phaseTimer >= phaseDuration)
                    {
                        AdvancePhase();
                        phaseTimer = 0f;
                    }
                }

                yield return null;
            }
        }

        private void AdvancePhase()
        {
            DayPhase previousPhase = currentPhase;

            // Advance to next phase
            currentPhase = GetNextPhase(currentPhase);

            // If we've cycled back to Morning, advance the day
            if (currentPhase == DayPhase.Morning)
            {
                AdvanceDay();
            }

            // Trigger phase change events
            TriggerPhaseEvents(previousPhase, currentPhase);

            Debug.Log($"[TimeManager] Phase advanced: {previousPhase} -> {currentPhase} (Day {currentDay})");
        }

        private void AdvanceDay()
        {
            int previousDay = currentDay;
            currentDay++;

            // Check for season change
            int dayInSeason = (currentDay - 1) % daysPerSeason + 1;
            if (dayInSeason == 1 && currentDay > 1) // New season (except first day)
            {
                AdvanceSeason();
            }

            // Trigger day change events
            OnDayChanged?.Invoke(currentDay);
            EventBus.Publish(new DayChangedEvent(previousDay, currentDay, currentSeason, currentYear));

            Debug.Log($"[TimeManager] Day advanced: {previousDay} -> {currentDay}");
        }

        private void AdvanceSeason()
        {
            Season previousSeason = currentSeason;
            currentSeason = GetNextSeason(currentSeason);

            // If we've cycled back to Spring, advance the year
            if (currentSeason == Season.Spring && previousSeason == Season.Winter)
            {
                AdvanceYear();
            }

            // Trigger season change events
            TriggerSeasonEvents(previousSeason, currentSeason);

            Debug.Log($"[TimeManager] Season advanced: {previousSeason} -> {currentSeason} (Year {currentYear})");
        }

        private void AdvanceYear()
        {
            int previousYear = currentYear;
            currentYear++;

            OnYearChanged?.Invoke(currentYear);
            EventBus.Publish(new YearChangedEvent(previousYear, currentYear));
            Debug.Log($"[TimeManager] Year advanced: {previousYear} -> {currentYear}");
        }

        private void TriggerPhaseEvents(DayPhase previousPhase, DayPhase newPhase)
        {
            OnPhaseChanged?.Invoke(newPhase);
            EventBus.Publish(new PhaseChangedEvent(previousPhase, newPhase, currentDay));
        }

        private void TriggerSeasonEvents(Season previousSeason, Season newSeason)
        {
            OnSeasonChanged?.Invoke(newSeason);
            EventBus.Publish(new SeasonChangedEvent(previousSeason, newSeason, currentYear));
        }

        private DayPhase GetNextPhase(DayPhase currentPhase)
        {
            return currentPhase switch
            {
                DayPhase.Morning => DayPhase.Afternoon,
                DayPhase.Afternoon => DayPhase.Evening,
                DayPhase.Evening => DayPhase.Night,
                DayPhase.Night => DayPhase.Morning,
                _ => DayPhase.Morning,
            };
        }

        private Season GetNextSeason(Season currentSeason)
        {
            return currentSeason switch
            {
                Season.Spring => Season.Summer,
                Season.Summer => Season.Autumn,
                Season.Autumn => Season.Winter,
                Season.Winter => Season.Spring,
                _ => Season.Spring,
            };
        }

        // Public utility methods
        public void SetTimeSpeed(float multiplier)
        {
            TimeSpeedMultiplier = Mathf.Clamp(multiplier, 0.1f, 10f);
            Debug.Log($"[TimeManager] Time speed set to {TimeSpeedMultiplier}x");
        }

        public void SkipToNextPhase()
        {
            AdvancePhase();
            phaseTimer = 0f;
            Debug.Log("[TimeManager] Skipped to next phase");
        }

        public void SkipToNextDay()
        {
            while (currentPhase != DayPhase.Night)
            {
                AdvancePhase();
            }
            AdvancePhase(); // Advance to Morning of next day
            phaseTimer = 0f;
            Debug.Log("[TimeManager] Skipped to next day");
        }

        public float GetPhaseProgress()
        {
            return phaseTimer / phaseDuration;
        }

        /// <summary>
        /// 時間を進める（テスト用）
        /// </summary>
        /// <param name="hours">進める時間（時間単位）</param>
        public void AdvanceTime(float hours)
        {
            float phasesToAdvance = hours / 6f; // 6 hours per phase
            int fullPhases = Mathf.FloorToInt(phasesToAdvance);

            for (int i = 0; i < fullPhases; i++)
            {
                AdvancePhase();
            }

            // Advance partial phase
            float partialPhase = phasesToAdvance - fullPhases;
            if (partialPhase > 0)
            {
                phaseTimer += partialPhase * phaseDuration;
                if (phaseTimer >= phaseDuration)
                {
                    AdvancePhase();
                }
            }
        }

        /// <summary>
        /// 次の日に進める
        /// </summary>
        public void AdvanceDay()
        {
            SkipToNextDay();
        }

        public string GetFormattedTime()
        {
            return $"Year {currentYear}, Day {currentDay} ({currentSeason}) - {currentPhase}";
        }

        public bool IsBusinessHours()
        {
            return currentPhase == DayPhase.Morning
                || currentPhase == DayPhase.Afternoon
                || currentPhase == DayPhase.Evening;
        }

        public bool IsSeasonTransitionDay()
        {
            int dayInSeason = (currentDay - 1) % daysPerSeason + 1;
            return dayInSeason == daysPerSeason; // Last day of season
        }

        // Save/Load support
        public TimeData GetTimeData()
        {
            return new TimeData
            {
                currentSeason = this.currentSeason,
                currentDay = this.currentDay,
                currentPhase = this.currentPhase,
                currentYear = this.currentYear,
                phaseTimer = this.phaseTimer,
            };
        }

        public void LoadTimeData(TimeData timeData)
        {
            currentSeason = timeData.currentSeason;
            currentDay = timeData.currentDay;
            currentPhase = timeData.currentPhase;
            currentYear = timeData.currentYear;
            phaseTimer = timeData.phaseTimer;

            Debug.Log($"[TimeManager] Time data loaded: {GetFormattedTime()}");
        }

        private void OnDestroy()
        {
            if (timeCoroutine != null)
            {
                StopCoroutine(timeCoroutine);
            }
        }

        // Debug methods
        public void LogCurrentTimeState()
        {
            Debug.Log($"[TimeManager] Current time: {GetFormattedTime()}, Progress: {GetPhaseProgress():P1}");
        }
    }

    [System.Serializable]
    public class TimeData
    {
        public Season currentSeason;
        public int currentDay;
        public DayPhase currentPhase;
        public int currentYear;
        public float phaseTimer;
    }
}
