using System;
using UnityEngine;

namespace MerchantTails.Core
{
    /// <summary>
    /// UIManagerのインターフェース
    /// </summary>
    public interface IUIManager
    {
        void ShowNotification(string title, string message, float duration, NotificationType type);
        void ShowDialog(string title, string message, Action onConfirm, Action onCancel);
        void ShowInputDialog(string title, string message, string defaultValue, Action<string> onConfirm, Action onCancel);
        void ShowProgressBar(string title, float progress);
        void HideProgressBar();
        void ShowLoadingScreen(bool show);
        void ShowError(string title, string message);
        void ShowWarning(string title, string message);
        void ShowInfo(string title, string message);
        void ShowSuccess(string title, string message);
        void UpdateMoney(int amount);
        void UpdateRank(string rank);
        void UpdateTime(string timeText);
        void UpdateSeason(string season);
        
        // UIManager静的プロパティを設定するためのメソッド
        void RegisterAsInstance();
    }

    public enum NotificationType
    {
        Info,
        Success,
        Warning,
        Error
    }
}