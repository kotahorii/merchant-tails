namespace MerchantTails.Core
{
    /// <summary>
    /// TutorialSystemのインターフェース
    /// </summary>
    public interface ITutorialSystem
    {
        bool IsCompleted { get; }
        int CurrentStep { get; }
        void StartTutorial();
        void CompleteCurrentStep();
        void SkipTutorial();
        void RestartTutorial();
        bool IsStepCompleted(int step);
        string GetCurrentStepDescription();
        
        // TutorialSystem静的プロパティを設定するためのメソッド
        void RegisterAsInstance();
    }
}