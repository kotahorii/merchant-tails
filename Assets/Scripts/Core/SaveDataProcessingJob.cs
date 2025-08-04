using Unity.Burst;
using Unity.Collections;
using Unity.Jobs;
using Unity.Mathematics;
using System.Text;

namespace MerchantTails.Core
{
    /// <summary>
    /// セーブデータの圧縮・暗号化を並列処理するJob
    /// Unity 6のJob Systemを活用
    /// </summary>
    [BurstCompile]
    public struct SaveDataCompressionJob : IJob
    {
        [ReadOnly]
        public NativeArray<byte> uncompressedData;

        [WriteOnly]
        public NativeArray<byte> compressedData;

        [WriteOnly]
        public NativeArray<int> compressedSize;

        public void Execute()
        {
            // 簡易的なRLE圧縮を実装
            int writeIndex = 0;
            int readIndex = 0;

            while (readIndex < uncompressedData.Length && writeIndex < compressedData.Length - 2)
            {
                byte currentByte = uncompressedData[readIndex];
                int count = 1;

                // 同じバイトが連続する数をカウント
                while (readIndex + count < uncompressedData.Length &&
                       count < 255 &&
                       uncompressedData[readIndex + count] == currentByte)
                {
                    count++;
                }

                // 圧縮データを書き込み
                if (count > 3) // 4バイト以上の連続なら圧縮
                {
                    compressedData[writeIndex++] = 0xFF; // マーカー
                    compressedData[writeIndex++] = (byte)count;
                    compressedData[writeIndex++] = currentByte;
                    readIndex += count;
                }
                else
                {
                    // 圧縮しない場合はそのまま書き込み
                    for (int i = 0; i < count && writeIndex < compressedData.Length; i++)
                    {
                        if (uncompressedData[readIndex] == 0xFF && writeIndex < compressedData.Length - 1)
                        {
                            // エスケープ処理
                            compressedData[writeIndex++] = 0xFF;
                            compressedData[writeIndex++] = 0x00;
                        }
                        else
                        {
                            compressedData[writeIndex++] = uncompressedData[readIndex];
                        }
                        readIndex++;
                    }
                }
            }

            compressedSize[0] = writeIndex;
        }
    }

    /// <summary>
    /// セーブデータの暗号化を並列処理するJob
    /// </summary>
    [BurstCompile]
    public struct SaveDataEncryptionJob : IJobParallelFor
    {
        [ReadOnly]
        public NativeArray<byte> plainData;

        [WriteOnly]
        public NativeArray<byte> encryptedData;

        [ReadOnly]
        public uint encryptionKey;

        public void Execute(int index)
        {
            // 簡易的なXOR暗号化（実際のプロダクトではより強力な暗号化を使用）
            var random = new Unity.Mathematics.Random(encryptionKey + (uint)index);
            byte keyByte = (byte)(random.NextUInt() & 0xFF);
            encryptedData[index] = (byte)(plainData[index] ^ keyByte);
        }
    }

    /// <summary>
    /// セーブデータの復号化を並列処理するJob
    /// </summary>
    [BurstCompile]
    public struct SaveDataDecryptionJob : IJobParallelFor
    {
        [ReadOnly]
        public NativeArray<byte> encryptedData;

        [WriteOnly]
        public NativeArray<byte> decryptedData;

        [ReadOnly]
        public uint encryptionKey;

        public void Execute(int index)
        {
            // 復号化（暗号化と同じXOR処理）
            var random = new Unity.Mathematics.Random(encryptionKey + (uint)index);
            byte keyByte = (byte)(random.NextUInt() & 0xFF);
            decryptedData[index] = (byte)(encryptedData[index] ^ keyByte);
        }
    }

    /// <summary>
    /// セーブデータの解凍を並列処理するJob
    /// </summary>
    [BurstCompile]
    public struct SaveDataDecompressionJob : IJob
    {
        [ReadOnly]
        public NativeArray<byte> compressedData;

        [ReadOnly]
        public int compressedSize;

        [WriteOnly]
        public NativeArray<byte> decompressedData;

        [WriteOnly]
        public NativeArray<int> decompressedSize;

        public void Execute()
        {
            int readIndex = 0;
            int writeIndex = 0;

            while (readIndex < compressedSize && writeIndex < decompressedData.Length)
            {
                byte currentByte = compressedData[readIndex];

                if (currentByte == 0xFF && readIndex + 1 < compressedSize)
                {
                    byte nextByte = compressedData[readIndex + 1];

                    if (nextByte == 0x00)
                    {
                        // エスケープされた0xFF
                        decompressedData[writeIndex++] = 0xFF;
                        readIndex += 2;
                    }
                    else if (readIndex + 2 < compressedSize)
                    {
                        // RLE圧縮データ
                        int count = nextByte;
                        byte valueByte = compressedData[readIndex + 2];

                        for (int i = 0; i < count && writeIndex < decompressedData.Length; i++)
                        {
                            decompressedData[writeIndex++] = valueByte;
                        }

                        readIndex += 3;
                    }
                    else
                    {
                        readIndex++;
                    }
                }
                else
                {
                    // 非圧縮データ
                    decompressedData[writeIndex++] = currentByte;
                    readIndex++;
                }
            }

            decompressedSize[0] = writeIndex;
        }
    }
}
