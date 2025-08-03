using System;
using System.Collections.Generic;
using UnityEngine;

namespace MerchantTails.Core
{
    /// <summary>
    /// システム間の疎結合な通信を提供するイベントバス
    /// Observer Patternを実装し、型安全なイベント配信を行う
    /// </summary>
    public static class EventBus
    {
        private static readonly Dictionary<Type, List<object>> eventHandlers = new Dictionary<Type, List<object>>();
        private static readonly object lockObject = new object();
        
        /// <summary>
        /// イベントハンドラーを登録する
        /// </summary>
        /// <typeparam name="T">イベントの型</typeparam>
        /// <param name="handler">イベントハンドラー</param>
        public static void Subscribe<T>(Action<T> handler) where T : IGameEvent
        {
            lock (lockObject)
            {
                Type eventType = typeof(T);
                
                if (!eventHandlers.ContainsKey(eventType))
                {
                    eventHandlers[eventType] = new List<object>();
                }
                
                eventHandlers[eventType].Add(handler);
                
                Debug.Log($"[EventBus] Subscribed to {eventType.Name}. Total subscribers: {eventHandlers[eventType].Count}");
            }
        }
        
        /// <summary>
        /// イベントハンドラーの登録を解除する
        /// </summary>
        /// <typeparam name="T">イベントの型</typeparam>
        /// <param name="handler">イベントハンドラー</param>
        public static void Unsubscribe<T>(Action<T> handler) where T : IGameEvent
        {
            lock (lockObject)
            {
                Type eventType = typeof(T);
                
                if (eventHandlers.ContainsKey(eventType))
                {
                    eventHandlers[eventType].Remove(handler);
                    
                    if (eventHandlers[eventType].Count == 0)
                    {
                        eventHandlers.Remove(eventType);
                    }
                    
                    Debug.Log($"[EventBus] Unsubscribed from {eventType.Name}");
                }
            }
        }
        
        /// <summary>
        /// イベントを発行する
        /// </summary>
        /// <typeparam name="T">イベントの型</typeparam>
        /// <param name="gameEvent">発行するイベント</param>
        public static void Publish<T>(T gameEvent) where T : IGameEvent
        {
            lock (lockObject)
            {
                Type eventType = typeof(T);
                
                if (eventHandlers.ContainsKey(eventType))
                {
                    var handlers = eventHandlers[eventType];
                    Debug.Log($"[EventBus] Publishing {eventType.Name} to {handlers.Count} subscribers");
                    
                    // イベント配信中にハンドラーリストが変更される可能性があるため、コピーを作成
                    var handlersCopy = new List<object>(handlers);
                    
                    foreach (var handler in handlersCopy)
                    {
                        try
                        {
                            ((Action<T>)handler)?.Invoke(gameEvent);
                        }
                        catch (Exception ex)
                        {
                            Debug.LogError($"[EventBus] Error in event handler for {eventType.Name}: {ex.Message}");
                        }
                    }
                }
                else
                {
                    Debug.LogWarning($"[EventBus] No subscribers for event {eventType.Name}");
                }
            }
        }
        
        /// <summary>
        /// 指定した型のすべてのイベントハンドラーを削除する
        /// </summary>
        /// <typeparam name="T">イベントの型</typeparam>
        public static void Clear<T>() where T : IGameEvent
        {
            lock (lockObject)
            {
                Type eventType = typeof(T);
                
                if (eventHandlers.ContainsKey(eventType))
                {
                    eventHandlers.Remove(eventType);
                    Debug.Log($"[EventBus] Cleared all handlers for {eventType.Name}");
                }
            }
        }
        
        /// <summary>
        /// すべてのイベントハンドラーを削除する（主にシーン切り替え時やゲーム終了時に使用）
        /// </summary>
        public static void ClearAll()
        {
            lock (lockObject)
            {
                int totalHandlers = 0;
                foreach (var kvp in eventHandlers)
                {
                    totalHandlers += kvp.Value.Count;
                }
                
                eventHandlers.Clear();
                Debug.Log($"[EventBus] Cleared all event handlers. Total removed: {totalHandlers}");
            }
        }
        
        /// <summary>
        /// 現在登録されているイベントハンドラーの統計情報を取得する
        /// </summary>
        /// <returns>イベントタイプごとのハンドラー数</returns>
        public static Dictionary<string, int> GetSubscriberStats()
        {
            lock (lockObject)
            {
                var stats = new Dictionary<string, int>();
                
                foreach (var kvp in eventHandlers)
                {
                    stats[kvp.Key.Name] = kvp.Value.Count;
                }
                
                return stats;
            }
        }
        
        /// <summary>
        /// デバッグ用：現在の登録状況をログ出力
        /// </summary>
        public static void LogCurrentState()
        {
            lock (lockObject)
            {
                Debug.Log($"[EventBus] Current event handlers: {eventHandlers.Count} event types");
                
                foreach (var kvp in eventHandlers)
                {
                    Debug.Log($"  - {kvp.Key.Name}: {kvp.Value.Count} handlers");
                }
            }
        }
    }
    
    /// <summary>
    /// すべてのゲームイベントが実装すべきインターフェース
    /// イベントの型安全性を保証する
    /// </summary>
    public interface IGameEvent
    {
        /// <summary>イベントが発生した時刻</summary>
        DateTime Timestamp { get; }
    }
    
    /// <summary>
    /// ゲームイベントの基底クラス
    /// 共通機能を提供する
    /// </summary>
    public abstract class BaseGameEvent : IGameEvent
    {
        public DateTime Timestamp { get; private set; }
        
        protected BaseGameEvent()
        {
            Timestamp = DateTime.Now;
        }
    }
}