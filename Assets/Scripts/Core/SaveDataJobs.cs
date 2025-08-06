using Unity.Burst;
using Unity.Collections;
using Unity.Jobs;

namespace MerchantTails.Core
{
    /// <summary>
    /// セーブデータ圧縮用Job
    /// </summary>
    [BurstCompile]
    public struct SaveDataCompressionJob : IJob
    {
        [ReadOnly]
        public NativeArray<byte> uncompressedData;
        public NativeArray<byte> compressedData;
        public NativeArray<int> compressedSize;

        public void Execute()
        {
            // 簡単な圧縮実装（本番環境では適切な圧縮アルゴリズムを使用）
            int size = uncompressedData.Length;
            for (int i = 0; i < size; i++)
            {
                compressedData[i] = uncompressedData[i];
            }
            compressedSize[0] = size;
        }
    }

    /// <summary>
    /// セーブデータ暗号化用Job
    /// </summary>
    [BurstCompile]
    public struct SaveDataEncryptionJob : IJobParallelFor
    {
        [ReadOnly]
        public NativeArray<byte> plainData;
        public NativeArray<byte> encryptedData;
        public uint encryptionKey;

        public void Execute(int index)
        {
            // 簡単なXOR暗号化（本番環境では適切な暗号化を使用）
            encryptedData[index] = (byte)(plainData[index] ^ (encryptionKey >> (index % 4) * 8));
        }
    }

    /// <summary>
    /// セーブデータ復号化用Job
    /// </summary>
    [BurstCompile]
    public struct SaveDataDecryptionJob : IJobParallelFor
    {
        [ReadOnly]
        public NativeArray<byte> encryptedData;
        public NativeArray<byte> decryptedData;
        public uint encryptionKey;

        public void Execute(int index)
        {
            // 簡単なXOR復号化（本番環境では適切な復号化を使用）
            decryptedData[index] = (byte)(encryptedData[index] ^ (encryptionKey >> (index % 4) * 8));
        }
    }

    /// <summary>
    /// セーブデータ解凍用Job
    /// </summary>
    [BurstCompile]
    public struct SaveDataDecompressionJob : IJob
    {
        [ReadOnly]
        public NativeArray<byte> compressedData;
        public int compressedSize;
        public NativeArray<byte> decompressedData;
        public NativeArray<int> decompressedSize;

        public void Execute()
        {
            // 簡単な解凍実装（本番環境では適切な解凍アルゴリズムを使用）
            for (int i = 0; i < compressedSize; i++)
            {
                decompressedData[i] = compressedData[i];
            }
            decompressedSize[0] = compressedSize;
        }
    }
}