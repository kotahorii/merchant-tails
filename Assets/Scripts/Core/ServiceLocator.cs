using System;
using System.Collections.Generic;
using UnityEngine;

namespace MerchantTails.Core
{
    /// <summary>
    /// サービスロケーターパターンを実装し、Coreアセンブリから他のアセンブリへの依存を解決
    /// </summary>
    public static class ServiceLocator
    {
        private static Dictionary<Type, object> services = new Dictionary<Type, object>();

        /// <summary>
        /// サービスを登録
        /// </summary>
        public static void RegisterService<T>(T service) where T : class
        {
            var type = typeof(T);
            if (services.ContainsKey(type))
            {
                Debug.LogWarning($"Service {type.Name} is already registered. Overwriting...");
            }
            services[type] = service;
            Debug.Log($"Service {type.Name} registered");
        }

        /// <summary>
        /// サービスを取得
        /// </summary>
        public static T GetService<T>() where T : class
        {
            var type = typeof(T);
            if (services.TryGetValue(type, out var service))
            {
                return service as T;
            }
            Debug.LogError($"Service {type.Name} not found");
            return null;
        }

        /// <summary>
        /// サービスが登録されているか確認
        /// </summary>
        public static bool HasService<T>() where T : class
        {
            return services.ContainsKey(typeof(T));
        }

        /// <summary>
        /// すべてのサービスをクリア
        /// </summary>
        public static void Clear()
        {
            services.Clear();
            Debug.Log("All services cleared");
        }
    }
}