
using System;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Management;
using System.Net;
using System.Net.Http;
using System.Net.NetworkInformation;
using System.Reflection;
using System.Runtime.InteropServices;
using System.Threading.Tasks;
using Microsoft.EntityFrameworkCore;
using System.Printing;
using CyberSaaS.Terminal.Infrastructure.Database;

namespace CyberSaaS.Terminal.Infrastructure.Diagnostics;

public sealed class TerminalDiagnosticsService
{
    private const long MinimumDiskFreeBytes =
        500L * 1024 * 1024;

    private const double HighCpuThresholdPercent = 95.0;

    private const double HighMemoryThresholdPercent = 95.0;

    private const double LowDiskThresholdPercent = 95.0;

    public async Task<TerminalDiagnosticsResult> RunAsync(
        string apiBaseUrl)
    {
        var result =
            new TerminalDiagnosticsResult();

        // ---------------------------------------------------------
        // 1. NETWORK ADAPTER
        // ---------------------------------------------------------

        NetworkInterface[] interfaces =
            GetActiveNetworkInterfaces();

        result.NetworkAdapterAvailable =
            interfaces.Length > 0;

        // ---------------------------------------------------------
        // 2. LAN
        // ---------------------------------------------------------

        result.LANAvailable =
            interfaces.Any(IsLanInterface);

        // ---------------------------------------------------------
        // 3. INTERNET
        // ---------------------------------------------------------

        result.InternetAvailable =
            await CheckInternetAsync();

        // ---------------------------------------------------------
        // 4. DNS
        // ---------------------------------------------------------

        result.DnsAvailable =
            await CheckDnsAsync();

        // ---------------------------------------------------------
        // 5. CYBERSAAS API + LATENCY
        // ---------------------------------------------------------

        ApiCheckResult apiResult =
            await CheckApiAsync(apiBaseUrl);

        result.ApiAvailable =
            apiResult.Available;

        result.BackendLatencyMS =
            apiResult.LatencyMilliseconds;

        // ---------------------------------------------------------
        // 6. LOCAL DATABASE
        // ---------------------------------------------------------

        result.DatabaseAvailable =
            CheckDatabase();

        // ---------------------------------------------------------
        // 7. DISK
        // ---------------------------------------------------------

        DiskDiagnostics disk =
            CheckDisk();

        result.DiskAvailable =
            disk.Available;

        result.DiskTotalBytes =
            disk.TotalBytes;

        result.DiskFreeBytes =
            disk.FreeBytes;

        result.DiskUsedBytes =
            disk.UsedBytes;

        // ---------------------------------------------------------
        // 8. MEMORY
        // ---------------------------------------------------------

        MemoryDiagnostics memory =
            GetMemoryDiagnostics();

        result.MemoryAvailable =
            memory.Available;

        result.MemoryTotalBytes =
            memory.TotalBytes;

        result.MemoryUsedBytes =
            memory.UsedBytes;

        result.MemoryFreeBytes =
            memory.FreeBytes;

        result.MemoryUsagePercent =
            memory.UsagePercent;

        // ---------------------------------------------------------
        // 9. CPU
        // ---------------------------------------------------------

        result.CPUUsagePercent =
            await GetCpuUsagePercentAsync();

        // ---------------------------------------------------------
        // 10. PRINTER
        // ---------------------------------------------------------

        result.PrinterAvailable =
            CheckPrinter();

        // ---------------------------------------------------------
        // 11. SOUND
        // ---------------------------------------------------------

        SoundDiagnostics sound =
            CheckSound();

        result.SoundAvailable =
            sound.Available;

        result.SoundIssue =
            sound.Issue;

        // ---------------------------------------------------------
        // 12. DRIVERS
        // ---------------------------------------------------------

        DriverDiagnostics drivers =
            CheckDrivers();

        result.DriversAvailable =
            drivers.Available;

        result.DriverIssue =
            drivers.Issue;

        // ---------------------------------------------------------
        // 13. SECURITY / ANTIVIRUS
        // ---------------------------------------------------------

        SecurityDiagnostics security =
            CheckSecurity();

        result.SecurityAvailable =
            security.Available;

        result.AntivirusEnabled =
            security.AntivirusEnabled;

        result.SecurityIssue =
            security.Issue;

        // ---------------------------------------------------------
        // 14. OPERATING SYSTEM
        // ---------------------------------------------------------

        result.OSName =
            RuntimeInformation.OSDescription;

        result.OSVersion =
            Environment.OSVersion.Version.ToString();

        result.OSArchitecture =
            RuntimeInformation.OSArchitecture.ToString();

        // ---------------------------------------------------------
        // 15. CLIENT
        // ---------------------------------------------------------

        result.ClientVersion =
            GetClientVersion();

        result.ClientStatus =
            "running";

        // ---------------------------------------------------------
        // 16. UPTIME
        // ---------------------------------------------------------

        result.SystemUptime =
            GetSystemUptime();

        // ---------------------------------------------------------
        // 17. OFFLINE / SYNC QUEUE
        // ---------------------------------------------------------

        SyncDiagnostics sync =
            GetSyncDiagnostics();

        result.OfflineQueueCount =
            sync.Total;

        result.SyncPendingCount =
            sync.Pending;

        result.SyncFailedCount =
            sync.Failed;

        result.SyncStatus =
            sync.Status;

        result.LastSyncAt =
            sync.LastSyncAt;

        // ---------------------------------------------------------
        // 18. NETWORK ISSUE
        // ---------------------------------------------------------

        result.NetworkIssue =
            DetermineNetworkIssue(result);

        return result;
    }

    // =============================================================
    // NETWORK
    // =============================================================

    private static NetworkInterface[] GetActiveNetworkInterfaces()
    {
        try
        {
            return NetworkInterface
                .GetAllNetworkInterfaces()
                .Where(
                    n =>
                        n.OperationalStatus ==
                        OperationalStatus.Up &&
                        n.NetworkInterfaceType !=
                        NetworkInterfaceType.Loopback &&
                        n.NetworkInterfaceType !=
                        NetworkInterfaceType.Tunnel)
                .ToArray();
        }
        catch
        {
            return Array.Empty<NetworkInterface>();
        }
    }

    private static bool IsLanInterface(
        NetworkInterface networkInterface)
    {
        if (networkInterface.OperationalStatus !=
            OperationalStatus.Up)
        {
            return false;
        }

        if (networkInterface.NetworkInterfaceType ==
            NetworkInterfaceType.Loopback ||
            networkInterface.NetworkInterfaceType ==
            NetworkInterfaceType.Tunnel)
        {
            return false;
        }

        // Both Ethernet and Wi-Fi provide local network
        // connectivity. We deliberately do not depend on a
        // GigabitEthernet enum because it is not available in
        // the .NET NetworkInterfaceType enum.
        return
            networkInterface.NetworkInterfaceType ==
            NetworkInterfaceType.Ethernet
            ||
            networkInterface.NetworkInterfaceType ==
            NetworkInterfaceType.Wireless80211
            ||
            networkInterface.NetworkInterfaceType ==
            NetworkInterfaceType.FastEthernetFx
            ||
            networkInterface.NetworkInterfaceType ==
            NetworkInterfaceType.FastEthernetT;
    }

    private static async Task<bool> CheckInternetAsync()
    {
        try
        {
            using var client =
                new HttpClient
                {
                    Timeout =
                        TimeSpan.FromSeconds(5)
                };

            using var request =
                new HttpRequestMessage(
                    HttpMethod.Get,
                    "https://www.google.com/generate_204");

            using HttpResponseMessage response =
                await client.SendAsync(request);

            return true;
        }
        catch
        {
            return false;
        }
    }

    private static async Task<bool> CheckDnsAsync()
    {
        try
        {
            IPAddress[] addresses =
                await Dns.GetHostAddressesAsync(
                    "example.com");

            return addresses.Length > 0;
        }
        catch
        {
            return false;
        }
    }

    private static async Task<ApiCheckResult> CheckApiAsync(
        string apiBaseUrl)
    {
        if (string.IsNullOrWhiteSpace(apiBaseUrl))
        {
            return new ApiCheckResult
            {
                Available = false,
                LatencyMilliseconds = 0
            };
        }

        try
        {
            using var client =
                new HttpClient
                {
                    Timeout =
                        TimeSpan.FromSeconds(5)
                };

            Stopwatch stopwatch =
                Stopwatch.StartNew();

            using HttpResponseMessage response =
                await client.GetAsync(apiBaseUrl);

            stopwatch.Stop();

            return new ApiCheckResult
            {
                Available =
                    response.IsSuccessStatusCode ||
                    response.StatusCode ==
                        HttpStatusCode.Unauthorized ||
                    response.StatusCode ==
                        HttpStatusCode.Forbidden,

                LatencyMilliseconds =
                    (int)stopwatch.ElapsedMilliseconds
            };
        }
        catch
        {
            return new ApiCheckResult
            {
                Available = false,
                LatencyMilliseconds = 0
            };
        }
    }

    // =============================================================
    // DATABASE
    // =============================================================

    private static bool CheckDatabase()
    {
        try
        {
            using var db =
                new TerminalDbContext();

            db.Database.OpenConnection();

            db.Database.CloseConnection();

            return true;
        }
        catch
        {
            return false;
        }
    }

    // =============================================================
    // DISK
    // =============================================================

    private static DiskDiagnostics CheckDisk()
    {
        try
        {
            string systemRoot =
                Path.GetPathRoot(
                    Environment.SystemDirectory)
                ?? "C:\\";

            var drive =
                new DriveInfo(systemRoot);

            if (!drive.IsReady)
            {
                return new DiskDiagnostics();
            }

            long total =
                drive.TotalSize;

            long free =
                drive.AvailableFreeSpace;

            long used =
                Math.Max(
                    0,
                    total - free);

            double usagePercent =
                total > 0
                    ? used * 100.0 / total
                    : 0;

            bool available =
                free >
                MinimumDiskFreeBytes &&
                usagePercent <
                LowDiskThresholdPercent;

            return new DiskDiagnostics
            {
                Available = available,
                TotalBytes = total,
                FreeBytes = free,
                UsedBytes = used
            };
        }
        catch
        {
            return new DiskDiagnostics();
        }
    }

    // =============================================================
    // MEMORY
    // =============================================================

    private static MemoryDiagnostics GetMemoryDiagnostics()
    {
        try
        {
            MEMORYSTATUSEX memoryStatus =
                new MEMORYSTATUSEX();

            if (!GlobalMemoryStatusEx(
                    memoryStatus))
            {
                return new MemoryDiagnostics();
            }

            long total =
                (long)memoryStatus.ullTotalPhys;

            long free =
                (long)memoryStatus.ullAvailPhys;

            long used =
                Math.Max(
                    0,
                    total - free);

            double usagePercent =
                total > 0
                    ? used * 100.0 / total
                    : 0;

            return new MemoryDiagnostics
            {
                Available =
                    usagePercent <
                    HighMemoryThresholdPercent,

                TotalBytes = total,

                UsedBytes = used,

                FreeBytes = free,

                UsagePercent =
                    usagePercent
            };
        }
        catch
        {
            return new MemoryDiagnostics();
        }
    }

    // =============================================================
    // CPU
    // =============================================================

    private static async Task<double>
        GetCpuUsagePercentAsync()
    {
        try
        {
            using PerformanceCounter cpuCounter =
                new PerformanceCounter(
                    "Processor",
                    "% Processor Time",
                    "_Total");

            // First reading initializes the counter.
            _ = cpuCounter.NextValue();

            await Task.Delay(500);

            float cpuUsage =
                cpuCounter.NextValue();

            return Math.Clamp(
                cpuUsage,
                0,
                100);
        }
        catch
        {
            // PerformanceCounter can be unavailable on some
            // Windows configurations. Fall back to WMI.
            return await GetCpuUsageFromWmiAsync();
        }
    }

    private static async Task<double>
        GetCpuUsageFromWmiAsync()
    {
        try
        {
            using var searcher =
                new ManagementObjectSearcher(
                    "SELECT LoadPercentage " +
                    "FROM Win32_Processor");

            ManagementObjectCollection processors =
                searcher.Get();

            double total = 0;
            int count = 0;

            foreach (ManagementObject processor
                     in processors)
            {
                object? load =
                    processor["LoadPercentage"];

                if (load == null)
                {
                    continue;
                }

                total +=
                    Convert.ToDouble(load);

                count++;
            }

            await Task.CompletedTask;

            if (count == 0)
            {
                return 0;
            }

            return Math.Clamp(
                total / count,
                0,
                100);
        }
        catch
        {
            return 0;
        }
    }

    // =============================================================
    // PRINTER
    // =============================================================

    private static bool CheckPrinter()
    {
        try
        {
            using var printServer =
                new LocalPrintServer();

            PrintQueueCollection queues =
                printServer.GetPrintQueues(
                    new[]
                    {
                        EnumeratedPrintQueueTypes.Local
                    });

            foreach (PrintQueue queue in queues)
            {
                try
                {
                    queue.Refresh();

                    if (!queue.IsOffline &&
                        !queue.IsNotAvailable)
                    {
                        return true;
                    }
                }
                catch
                {
                    // Check the next printer.
                }
                finally
                {
                    queue.Dispose();
                }
            }

            return false;
        }
        catch
        {
            return false;
        }
    }

    // =============================================================
    // SOUND
    // =============================================================

    private static SoundDiagnostics CheckSound()
    {
        try
        {
            uint deviceCount =
                waveOutGetNumDevs();

            if (deviceCount > 0)
            {
                return new SoundDiagnostics
                {
                    Available = true,
                    Issue = string.Empty
                };
            }

            return new SoundDiagnostics
            {
                Available = false,
                Issue =
                    "No Windows audio output device was detected."
            };
        }
        catch (Exception ex)
        {
            return new SoundDiagnostics
            {
                Available = false,
                Issue =
                    $"Sound check failed: {ex.Message}"
            };
        }
    }

    // =============================================================
    // DRIVERS
    // =============================================================

    private static DriverDiagnostics CheckDrivers()
    {
        try
        {
            using var searcher =
                new ManagementObjectSearcher(
                    "SELECT Name, Status, ConfigManagerErrorCode " +
                    "FROM Win32_PnPEntity");

            ManagementObjectCollection devices =
                searcher.Get();

            int problemCount = 0;

            foreach (ManagementObject device in devices)
            {
                object? errorCode =
                    device["ConfigManagerErrorCode"];

                if (errorCode == null)
                {
                    continue;
                }

                uint code =
                    Convert.ToUInt32(errorCode);

                if (code != 0)
                {
                    problemCount++;
                }
            }

            if (problemCount == 0)
            {
                return new DriverDiagnostics
                {
                    Available = true,
                    Issue = string.Empty
                };
            }

            return new DriverDiagnostics
            {
                Available = false,
                Issue =
                    $"{problemCount} device driver problem(s) detected."
            };
        }
        catch (Exception ex)
        {
            return new DriverDiagnostics
            {
                Available = false,
                Issue =
                    $"Driver check failed: {ex.Message}"
            };
        }
    }

    // =============================================================
    // SECURITY / ANTIVIRUS
    // =============================================================

    private static SecurityDiagnostics CheckSecurity()
    {
        try
        {
            using var searcher =
                new ManagementObjectSearcher(
                    @"root\SecurityCenter2",
                    "SELECT displayName, productState " +
                    "FROM AntiVirusProduct");

            ManagementObjectCollection products =
                searcher.Get();

            if (products.Count == 0)
            {
                return new SecurityDiagnostics
                {
                    Available = false,
                    AntivirusEnabled = false,
                    Issue =
                        "No registered antivirus product was detected."
                };
            }

            bool antivirusEnabled = false;

            foreach (ManagementObject product in products)
            {
                object? productState =
                    product["productState"];

                if (productState == null)
                {
                    continue;
                }

                uint state =
                    Convert.ToUInt32(productState);

                // Windows Security Center productState
                // uses the middle byte for real-time
                // protection state.
                uint realTimeProtectionState =
                    (state >> 8) & 0xFF;

                if (realTimeProtectionState ==
                    0x10 ||
                    realTimeProtectionState ==
                    0x11)
                {
                    antivirusEnabled = true;
                }
            }

            if (antivirusEnabled)
            {
                return new SecurityDiagnostics
                {
                    Available = true,
                    AntivirusEnabled = true,
                    Issue = string.Empty
                };
            }

            return new SecurityDiagnostics
            {
                Available = false,
                AntivirusEnabled = false,
                Issue =
                    "An antivirus product is registered, but active protection could not be confirmed."
            };
        }
        catch (Exception ex)
        {
            return new SecurityDiagnostics
            {
                Available = false,
                AntivirusEnabled = false,
                Issue =
                    $"Security check failed: {ex.Message}"
            };
        }
    }

    // =============================================================
    // CLIENT VERSION
    // =============================================================

    private static string GetClientVersion()
    {
        try
        {
            Version? version =
                Assembly
                    .GetExecutingAssembly()
                    .GetName()
                    .Version;

            return version?.ToString()
                   ?? "unknown";
        }
        catch
        {
            return "unknown";
        }
    }

    // =============================================================
    // UPTIME
    // =============================================================

    private static TimeSpan GetSystemUptime()
    {
        return TimeSpan.FromMilliseconds(
            Environment.TickCount64);
    }

    // =============================================================
    // SYNC / OFFLINE QUEUE
    // =============================================================

    private static SyncDiagnostics
        GetSyncDiagnostics()
    {
        try
        {
            using var db =
                new TerminalDbContext();

            int total =
                db.OfflineQueueItems.Count();

            int pending =
                db.OfflineQueueItems
                    .Count(
                        x =>
                            x.Status ==
                            "pending");

            int failed =
                db.OfflineQueueItems
                    .Count(
                        x =>
                            x.Status ==
                            "failed");

            DateTime? lastSyncAt =
                db.OfflineQueueItems
                    .Where(
                        x =>
                            x.Status ==
                            "completed" &&
                            x.LastAttemptAtUtc !=
                            null)
                    .OrderByDescending(
                        x =>
                            x.LastAttemptAtUtc)
                    .Select(
                        x =>
                            (DateTime?)
                            x.LastAttemptAtUtc)
                    .FirstOrDefault();

            string status;

            if (failed > 0)
            {
                status = "failed";
            }
            else if (pending > 0)
            {
                status = "pending";
            }
            else
            {
                status = "synced";
            }

            return new SyncDiagnostics
            {
                Total = total,
                Pending = pending,
                Failed = failed,
                Status = status,
                LastSyncAt = lastSyncAt
            };
        }
        catch
        {
            return new SyncDiagnostics
            {
                Total = 0,
                Pending = 0,
                Failed = 0,
                Status = "unknown",
                LastSyncAt = null
            };
        }
    }

    // =============================================================
    // NETWORK ISSUE
    // =============================================================

    private static string DetermineNetworkIssue(
        TerminalDiagnosticsResult result)
    {
        if (!result.NetworkAdapterAvailable)
        {
            return
                "No network adapter is connected.";
        }

        if (!result.LANAvailable &&
            !result.InternetAvailable)
        {
            return
                "No active LAN connection was detected and internet access is unavailable.";
        }

        if (!result.InternetAvailable &&
            !result.DnsAvailable)
        {
            return
                "Internet connection and DNS resolution are unavailable.";
        }

        if (!result.InternetAvailable)
        {
            return
                "Internet access is unavailable.";
        }

        if (!result.DnsAvailable)
        {
            return
                "DNS resolution is failing.";
        }

        if (!result.ApiAvailable)
        {
            return
                "Internet works, but the CyberSaaS API is unreachable.";
        }

        if (result.BackendLatencyMS >= 2000)
        {
            return
                "CyberSaaS API response latency is high.";
        }

        return
            "No basic network problem detected.";
    }

    // =============================================================
    // DNS REPAIR
    // =============================================================

    public async Task<NetworkRepairResult>
        FlushDnsAsync()
    {
        try
        {
            var processInfo =
                new ProcessStartInfo
                {
                    FileName = "ipconfig.exe",
                    Arguments = "/flushdns",
                    UseShellExecute = false,
                    CreateNoWindow = true,
                    RedirectStandardOutput = true,
                    RedirectStandardError = true
                };

            using Process? process =
                Process.Start(processInfo);

            if (process == null)
            {
                return new NetworkRepairResult
                {
                    Success = false,
                    Message =
                        "Could not start the DNS repair command."
                };
            }

            await process.WaitForExitAsync();

            if (process.ExitCode == 0)
            {
                return new NetworkRepairResult
                {
                    Success = true,
                    Message =
                        "DNS cache was flushed successfully."
                };
            }

            return new NetworkRepairResult
            {
                Success = false,
                Message =
                    "Windows could not flush the DNS cache."
            };
        }
        catch (Exception ex)
        {
            return new NetworkRepairResult
            {
                Success = false,
                Message =
                    $"DNS repair failed: {ex.Message}"
            };
        }
    }

    // =============================================================
    // WINDOWS MEMORY API
    // =============================================================

    [StructLayout(
        LayoutKind.Sequential,
        CharSet = CharSet.Unicode)]
    private sealed class MEMORYSTATUSEX
    {
        public uint dwLength =
            (uint)Marshal.SizeOf<
                MEMORYSTATUSEX>();

        public uint dwMemoryLoad;

        public ulong ullTotalPhys;

        public ulong ullAvailPhys;

        public ulong ullTotalPageFile;

        public ulong ullAvailPageFile;

        public ulong ullTotalVirtual;

        public ulong ullAvailVirtual;

        public ulong ullAvailExtendedVirtual;
    }

    [DllImport(
        "kernel32.dll",
        CharSet = CharSet.Unicode)]
    [return: MarshalAs(
        UnmanagedType.Bool)]
    private static extern bool
        GlobalMemoryStatusEx(
            [In, Out]
            MEMORYSTATUSEX lpBuffer);

    // =============================================================
    // WINDOWS AUDIO API
    // =============================================================

    [DllImport(
        "winmm.dll",
        EntryPoint = "waveOutGetNumDevs")]
    private static extern uint
        waveOutGetNumDevs();

    // =============================================================
    // RESULT TYPES
    // =============================================================

    private sealed class ApiCheckResult
    {
        public bool Available { get; set; }

        public int LatencyMilliseconds { get; set; }
    }

    private sealed class DiskDiagnostics
    {
        public bool Available { get; set; }

        public long TotalBytes { get; set; }

        public long FreeBytes { get; set; }

        public long UsedBytes { get; set; }
    }

    private sealed class MemoryDiagnostics
    {
        public bool Available { get; set; }

        public long TotalBytes { get; set; }

        public long UsedBytes { get; set; }

        public long FreeBytes { get; set; }

        public double UsagePercent { get; set; }
    }

    private sealed class SoundDiagnostics
    {
        public bool Available { get; set; }

        public string Issue { get; set; } =
            string.Empty;
    }

    private sealed class DriverDiagnostics
    {
        public bool Available { get; set; }

        public string Issue { get; set; } =
            string.Empty;
    }

    private sealed class SecurityDiagnostics
    {
        public bool Available { get; set; }

        public bool AntivirusEnabled { get; set; }

        public string Issue { get; set; } =
            string.Empty;
    }

    private sealed class SyncDiagnostics
    {
        public int Total { get; set; }

        public int Pending { get; set; }

        public int Failed { get; set; }

        public string Status { get; set; } =
            "unknown";

        public DateTime? LastSyncAt { get; set; }
    }
}

// =================================================================
// PUBLIC DIAGNOSTICS RESULT
// =================================================================

public sealed class TerminalDiagnosticsResult
{
    public bool NetworkAdapterAvailable { get; set; }

    public bool LANAvailable { get; set; }

    public bool InternetAvailable { get; set; }

    public bool DnsAvailable { get; set; }

    public bool ApiAvailable { get; set; }

    public int BackendLatencyMS { get; set; }

    public bool DatabaseAvailable { get; set; }

    public bool DiskAvailable { get; set; }

    public long DiskTotalBytes { get; set; }

    public long DiskUsedBytes { get; set; }

    public long DiskFreeBytes { get; set; }

    public bool MemoryAvailable { get; set; }

    public long MemoryTotalBytes { get; set; }

    public long MemoryUsedBytes { get; set; }

    public long MemoryFreeBytes { get; set; }

    public double MemoryUsagePercent { get; set; }

    public double CPUUsagePercent { get; set; }

    public bool PrinterAvailable { get; set; }

    public bool SoundAvailable { get; set; }

    public string SoundIssue { get; set; } =
        string.Empty;

    public bool DriversAvailable { get; set; }

    public string DriverIssue { get; set; } =
        string.Empty;

    public bool SecurityAvailable { get; set; }

    public bool AntivirusEnabled { get; set; }

    public string SecurityIssue { get; set; } =
        string.Empty;

    public string OSName { get; set; } =
        string.Empty;

    public string OSVersion { get; set; } =
        string.Empty;

    public string OSArchitecture { get; set; } =
        string.Empty;

    public string ClientVersion { get; set; } =
        string.Empty;

    public string ClientStatus { get; set; } =
        "running";

    public TimeSpan SystemUptime { get; set; }

    public int OfflineQueueCount { get; set; }

    public int SyncPendingCount { get; set; }

    public int SyncFailedCount { get; set; }

    public string SyncStatus { get; set; } =
        "unknown";

    public DateTime? LastSyncAt { get; set; }

    public string NetworkIssue { get; set; } =
        string.Empty;
}

// =================================================================
// NETWORK REPAIR RESULT
// =================================================================

public sealed class NetworkRepairResult
{
    public bool Success { get; set; }

    public string Message { get; set; } =
        string.Empty;
}
