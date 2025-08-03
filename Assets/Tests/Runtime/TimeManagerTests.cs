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
    /// TimeManagerの単体テスト
    /// </summary>
    public class TimeManagerTests : TestBase
    {
        [Test]
        public void InitialState_IsCorrect()
        {
            // Assert
            Assert.AreEqual(1, timeManager.CurrentDay, "Initial day should be 1");
            Assert.AreEqual(Season.Spring, timeManager.CurrentSeason, "Initial season should be Spring");
            Assert.AreEqual(DayPhase.Morning, timeManager.CurrentPhase, "Initial phase should be Morning");
            Assert.AreEqual(0f, timeManager.DayProgress, "Initial day progress should be 0");
        }

        [Test]
        public void AdvanceTime_SingleHour_UpdatesCorrectly()
        {
            // Arrange
            float initialProgress = timeManager.DayProgress;

            // Act
            timeManager.AdvanceTime(1f);

            // Assert
            float expectedProgress = initialProgress + (1f / 24f);
            AssertFloatEquals(expectedProgress, timeManager.DayProgress, 0.01f);
        }

        [Test]
        public void AdvanceTime_FullDay_AdvancesToNextDay()
        {
            // Arrange
            int initialDay = timeManager.CurrentDay;

            // Act
            timeManager.AdvanceTime(24f);

            // Assert
            Assert.AreEqual(initialDay + 1, timeManager.CurrentDay,
                "Day should advance by 1");
            Assert.AreEqual(DayPhase.Morning, timeManager.CurrentPhase,
                "Should start new day in Morning phase");
        }

        [Test]
        public void AdvanceDay_IncrementsDayCounter()
        {
            // Arrange
            int initialDay = timeManager.CurrentDay;

            // Act
            timeManager.AdvanceDay();

            // Assert
            Assert.AreEqual(initialDay + 1, timeManager.CurrentDay);
            Assert.AreEqual(0f, timeManager.DayProgress, "Day progress should reset to 0");
        }

        [Test]
        public void GetDayPhase_ReturnsCorrectPhase()
        {
            // Test Morning (0.0 - 0.25)
            timeManager.LoadTimeData(1, Season.Spring, DayPhase.Morning, 0.1f);
            Assert.AreEqual(DayPhase.Morning, timeManager.CurrentPhase);

            // Test Afternoon (0.25 - 0.5)
            timeManager.LoadTimeData(1, Season.Spring, DayPhase.Morning, 0.3f);
            Assert.AreEqual(DayPhase.Afternoon, timeManager.GetDayPhaseFromProgress(0.3f));

            // Test Evening (0.5 - 0.75)
            timeManager.LoadTimeData(1, Season.Spring, DayPhase.Morning, 0.6f);
            Assert.AreEqual(DayPhase.Evening, timeManager.GetDayPhaseFromProgress(0.6f));

            // Test Night (0.75 - 1.0)
            timeManager.LoadTimeData(1, Season.Spring, DayPhase.Morning, 0.8f);
            Assert.AreEqual(DayPhase.Night, timeManager.GetDayPhaseFromProgress(0.8f));
        }

        [Test]
        public void AdvanceSeason_CyclesThroughSeasons()
        {
            // Arrange
            timeManager.LoadTimeData(1, Season.Spring, DayPhase.Morning, 0f);

            // Act & Assert
            timeManager.AdvanceSeason();
            Assert.AreEqual(Season.Summer, timeManager.CurrentSeason);

            timeManager.AdvanceSeason();
            Assert.AreEqual(Season.Autumn, timeManager.CurrentSeason);

            timeManager.AdvanceSeason();
            Assert.AreEqual(Season.Winter, timeManager.CurrentSeason);

            timeManager.AdvanceSeason();
            Assert.AreEqual(Season.Spring, timeManager.CurrentSeason,
                "Should cycle back to Spring");
        }

        [Test]
        public void GetTimeString_FormatsCorrectly()
        {
            // Arrange
            timeManager.LoadTimeData(15, Season.Summer, DayPhase.Afternoon, 0.375f);

            // Act
            string timeString = timeManager.GetTimeString();

            // Assert
            Assert.IsNotNull(timeString);
            Assert.IsTrue(timeString.Contains("15"), "Should contain day number");
            // Format might vary, so just check it's not empty
            Assert.IsTrue(timeString.Length > 0);
        }

        [UnityTest]
        public IEnumerator DayChange_PublishesEvent()
        {
            // Arrange
            bool eventReceived = false;
            int newDay = 0;

            EventBus.Subscribe<DayChangedEvent>((e) =>
            {
                eventReceived = true;
                newDay = e.NewDay;
            });

            // Act
            timeManager.AdvanceDay();

            // Assert
            yield return WaitForCondition(() => eventReceived, 1f);
            Assert.IsTrue(eventReceived, "DayChangedEvent should have been published");
            Assert.AreEqual(timeManager.CurrentDay, newDay);
        }

        [UnityTest]
        public IEnumerator SeasonChange_PublishesEvent()
        {
            // Arrange
            bool eventReceived = false;
            Season newSeason = Season.Spring;

            EventBus.Subscribe<SeasonChangedEvent>((e) =>
            {
                eventReceived = true;
                newSeason = e.NewSeason;
            });

            // Act
            timeManager.AdvanceSeason();

            // Assert
            yield return WaitForCondition(() => eventReceived, 1f);
            Assert.IsTrue(eventReceived, "SeasonChangedEvent should have been published");
            Assert.AreEqual(timeManager.CurrentSeason, newSeason);
        }

        [UnityTest]
        public IEnumerator PhaseChange_PublishesEvent()
        {
            // Arrange
            bool eventReceived = false;
            DayPhase newPhase = DayPhase.Morning;

            EventBus.Subscribe<PhaseChangedEvent>((e) =>
            {
                eventReceived = true;
                newPhase = e.NewPhase;
            });

            // Act - Advance to next phase
            timeManager.AdvanceTime(6f); // 6 hours should change phase

            // Assert
            yield return WaitForCondition(() => eventReceived, 1f);
            Assert.IsTrue(eventReceived, "PhaseChangedEvent should have been published");
        }

        [Test]
        public void GetDaysInSeason_ReturnsCorrectValue()
        {
            // Assume 30 days per season
            int expectedDays = 30;

            // Act
            int daysInSeason = timeManager.GetDaysInSeason();

            // Assert
            Assert.AreEqual(expectedDays, daysInSeason);
        }

        [Test]
        public void GetDayOfSeason_CalculatesCorrectly()
        {
            // Arrange - Day 35 should be day 5 of second season
            timeManager.LoadTimeData(35, Season.Summer, DayPhase.Morning, 0f);

            // Act
            int dayOfSeason = timeManager.GetDayOfSeason();

            // Assert
            Assert.AreEqual(5, dayOfSeason, "Day 35 should be day 5 of the season");
        }

        [Test]
        public void IsWeekend_IdentifiesCorrectly()
        {
            // Assuming days 6, 7, 13, 14, etc. are weekends
            // Test weekday
            timeManager.LoadTimeData(3, Season.Spring, DayPhase.Morning, 0f);
            Assert.IsFalse(timeManager.IsWeekend(), "Day 3 should not be weekend");

            // Test weekend
            timeManager.LoadTimeData(7, Season.Spring, DayPhase.Morning, 0f);
            Assert.IsTrue(timeManager.IsWeekend(), "Day 7 should be weekend");
        }

        [Test]
        public void TimeSpeed_AffectsProgression()
        {
            // Arrange
            float normalSpeed = 1f;
            float fastSpeed = 2f;

            timeManager.SetTimeSpeed(normalSpeed);
            float initialProgress = timeManager.DayProgress;

            // Act - Advance with normal speed
            timeManager.AdvanceTime(1f);
            float progressAfterNormal = timeManager.DayProgress;

            // Reset and test with fast speed
            timeManager.LoadTimeData(1, Season.Spring, DayPhase.Morning, initialProgress);
            timeManager.SetTimeSpeed(fastSpeed);
            timeManager.AdvanceTime(1f);
            float progressAfterFast = timeManager.DayProgress;

            // Assert
            float normalDelta = progressAfterNormal - initialProgress;
            float fastDelta = progressAfterFast - initialProgress;
            AssertFloatEquals(normalDelta * 2f, fastDelta, 0.01f);
        }

        [Test]
        public void PauseTime_StopsProgression()
        {
            // Arrange
            timeManager.SetPaused(false);
            float initialProgress = timeManager.DayProgress;

            // Act
            timeManager.SetPaused(true);
            timeManager.AdvanceTime(5f); // Try to advance while paused

            // Assert
            Assert.AreEqual(initialProgress, timeManager.DayProgress,
                "Time should not progress while paused");
        }
    }
}
