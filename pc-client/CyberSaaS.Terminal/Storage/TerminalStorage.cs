using System;
using System.IO;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;

namespace CyberSaaS.Terminal.Storage;

public sealed class TerminalRegistration
{
    public string TerminalId { get; set; } = string.Empty;
    public string TenantId { get; set; } = string.Empty;
    public string BranchId { get; set; } = string.Empty;
    public string TerminalCode { get; set; } = string.Empty;
    public string Credential { get; set; } = string.Empty;
    public string MachineName { get; set; } = string.Empty;
    public string DeviceIdentifier { get; set; } = string.Empty;
}

public static class TerminalStorage
{
    private static readonly string StorageDirectory =
        Path.Combine(
            Environment.GetFolderPath(
                Environment.SpecialFolder.LocalApplicationData),
            "CyberSaaS",
            "Terminal");

    private static readonly string StorageFile =
        Path.Combine(StorageDirectory, "terminal.dat");

    public static bool IsRegistered()
    {
        return File.Exists(StorageFile);
    }

    public static void Save(TerminalRegistration registration)
    {
        Directory.CreateDirectory(StorageDirectory);

        string json = JsonSerializer.Serialize(registration);

        byte[] plainBytes = Encoding.UTF8.GetBytes(json);

        byte[] protectedBytes =
            ProtectedData.Protect(
                plainBytes,
                null,
                DataProtectionScope.CurrentUser);

        File.WriteAllBytes(StorageFile, protectedBytes);
    }

    public static TerminalRegistration? Load()
    {
        if (!File.Exists(StorageFile))
        {
            return null;
        }

        try
        {
            byte[] protectedBytes =
                File.ReadAllBytes(StorageFile);

            byte[] plainBytes =
                ProtectedData.Unprotect(
                    protectedBytes,
                    null,
                    DataProtectionScope.CurrentUser);

            string json =
                Encoding.UTF8.GetString(plainBytes);

            return JsonSerializer.Deserialize<TerminalRegistration>(
                json);
        }
        catch
        {
            return null;
        }
    }

    public static void Delete()
    {
        if (File.Exists(StorageFile))
        {
            File.Delete(StorageFile);
        }
    }
}