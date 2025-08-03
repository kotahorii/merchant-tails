using UnityEngine;
using UnityEditor;
using UnityEditor.Build.Reporting;
using System.Linq;

/// <summary>
/// CI/CD用のビルドスクリプト
/// GitHub Actionsから呼び出されるビルドメソッドを提供
/// </summary>
public class BuildScript
{
    /// <summary>
    /// CI環境でのゲームビルド
    /// </summary>
    public static void BuildGame()
    {
        // コマンドライン引数から設定を取得
        var args = System.Environment.GetCommandLineArgs();

        string buildPath = GetArg("-buildPath", "build");
        string buildTarget = GetArg("-buildTarget", "StandaloneWindows64");
        string buildName = GetArg("-buildName", "MerchantTails");

        // ビルドターゲットを設定
        BuildTarget target = GetBuildTarget(buildTarget);

        // ビルド設定
        BuildPlayerOptions buildPlayerOptions = new BuildPlayerOptions
        {
            scenes = GetScenePaths(),
            locationPathName = $"{buildPath}/{GetExecutableName(target, buildName)}",
            target = target,
            options = GetBuildOptions()
        };

        Debug.Log($"Building for {target} at {buildPlayerOptions.locationPathName}");

        // ビルド実行
        BuildReport report = BuildPipeline.BuildPlayer(buildPlayerOptions);
        BuildSummary summary = report.summary;

        if (summary.result == BuildResult.Succeeded)
        {
            Debug.Log($"Build succeeded: {summary.totalSize} bytes in {summary.totalTime}");
            EditorApplication.Exit(0);
        }
        else
        {
            Debug.LogError($"Build failed: {summary.result}");

            // エラー詳細をログ出力
            foreach (var step in report.steps)
            {
                foreach (var message in step.messages)
                {
                    if (message.type == LogType.Error || message.type == LogType.Exception)
                    {
                        Debug.LogError($"Build Error: {message.content}");
                    }
                }
            }

            EditorApplication.Exit(1);
        }
    }

    /// <summary>
    /// テスト用の軽量ビルド
    /// </summary>
    public static void BuildForTesting()
    {
        BuildPlayerOptions buildPlayerOptions = new BuildPlayerOptions
        {
            scenes = GetScenePaths(),
            locationPathName = "build/test/MerchantTailsTest",
            target = BuildTarget.StandaloneLinux64,
            options = BuildOptions.Development | BuildOptions.AllowDebugging
        };

        Debug.Log("Building test version...");

        BuildReport report = BuildPipeline.BuildPlayer(buildPlayerOptions);
        BuildSummary summary = report.summary;

        if (summary.result == BuildResult.Succeeded)
        {
            Debug.Log("Test build succeeded");
            EditorApplication.Exit(0);
        }
        else
        {
            Debug.LogError("Test build failed");
            EditorApplication.Exit(1);
        }
    }

    /// <summary>
    /// パフォーマンス最適化ビルド
    /// </summary>
    public static void BuildOptimized()
    {
        // パフォーマンス設定
        PlayerSettings.stripEngineCode = true;
        PlayerSettings.stripUnusedMeshComponents = true;
        PlayerSettings.iOS.scriptCallOptimization = ScriptCallOptimizationLevel.FastButNoExceptions;

        BuildPlayerOptions buildPlayerOptions = new BuildPlayerOptions
        {
            scenes = GetScenePaths(),
            locationPathName = "build/optimized/MerchantTailsOptimized",
            target = BuildTarget.StandaloneWindows64,
            options = BuildOptions.None // リリースビルド設定
        };

        Debug.Log("Building optimized version...");

        BuildReport report = BuildPipeline.BuildPlayer(buildPlayerOptions);
        BuildSummary summary = report.summary;

        if (summary.result == BuildResult.Succeeded)
        {
            Debug.Log($"Optimized build succeeded: {summary.totalSize} bytes");
            EditorApplication.Exit(0);
        }
        else
        {
            Debug.LogError("Optimized build failed");
            EditorApplication.Exit(1);
        }
    }

    private static string GetArg(string name, string defaultValue)
    {
        var args = System.Environment.GetCommandLineArgs();
        for (int i = 0; i < args.Length; i++)
        {
            if (args[i] == name && i + 1 < args.Length)
            {
                return args[i + 1];
            }
        }
        return defaultValue;
    }

    private static BuildTarget GetBuildTarget(string targetString)
    {
        return targetString switch
        {
            "StandaloneWindows64" => BuildTarget.StandaloneWindows64,
            "StandaloneLinux64" => BuildTarget.StandaloneLinux64,
            "StandaloneOSX" => BuildTarget.StandaloneOSX,
            "WebGL" => BuildTarget.WebGL,
            "Android" => BuildTarget.Android,
            "iOS" => BuildTarget.iOS,
            _ => BuildTarget.StandaloneWindows64
        };
    }

    private static string GetExecutableName(BuildTarget target, string buildName)
    {
        return target switch
        {
            BuildTarget.StandaloneWindows64 => $"{buildName}.exe",
            BuildTarget.StandaloneLinux64 => buildName,
            BuildTarget.StandaloneOSX => $"{buildName}.app",
            BuildTarget.WebGL => buildName,
            BuildTarget.Android => $"{buildName}.apk",
            BuildTarget.iOS => buildName,
            _ => $"{buildName}.exe"
        };
    }

    private static string[] GetScenePaths()
    {
        return EditorBuildSettings.scenes
            .Where(scene => scene.enabled)
            .Select(scene => scene.path)
            .ToArray();
    }

    private static BuildOptions GetBuildOptions()
    {
        var args = System.Environment.GetCommandLineArgs();

        BuildOptions options = BuildOptions.None;

        if (args.Contains("-development"))
        {
            options |= BuildOptions.Development;
        }

        if (args.Contains("-allowDebugging"))
        {
            options |= BuildOptions.AllowDebugging;
        }

        if (args.Contains("-autoRunPlayer"))
        {
            options |= BuildOptions.AutoRunPlayer;
        }

        return options;
    }

    /// <summary>
    /// ビルド前の検証
    /// </summary>
    [MenuItem("Build/Validate Build Requirements")]
    public static void ValidateBuildRequirements()
    {
        bool isValid = true;

        // シーンの検証
        var scenes = GetScenePaths();
        if (scenes.Length == 0)
        {
            Debug.LogError("No scenes enabled in build settings");
            isValid = false;
        }

        // 必須アセットの検証
        if (!System.IO.Directory.Exists("Assets/Scripts/Core"))
        {
            Debug.LogError("Core scripts directory missing");
            isValid = false;
        }

        // プレイヤー設定の検証
        if (string.IsNullOrEmpty(PlayerSettings.productName))
        {
            Debug.LogError("Product name not set in Player Settings");
            isValid = false;
        }

        if (string.IsNullOrEmpty(PlayerSettings.companyName))
        {
            Debug.LogError("Company name not set in Player Settings");
            isValid = false;
        }

        if (isValid)
        {
            Debug.Log("✅ Build requirements validation passed");
        }
        else
        {
            Debug.LogError("❌ Build requirements validation failed");
        }
    }

    /// <summary>
    /// テスト実行前の環境設定
    /// </summary>
    public static void SetupTestEnvironment()
    {
        // テスト用設定
        PlayerSettings.runInBackground = true;
        PlayerSettings.displayResolutionDialog = ResolutionDialogSetting.Disabled;
        PlayerSettings.defaultIsNativeResolution = false;
        PlayerSettings.defaultScreenWidth = 1280;
        PlayerSettings.defaultScreenHeight = 720;

        Debug.Log("Test environment configured");
    }
}
