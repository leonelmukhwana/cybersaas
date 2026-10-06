using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using CyberSaaS.Terminal.Infrastructure.Api;
using CyberSaaS.Terminal.Infrastructure.Diagnostics;
using CyberSaaS.Terminal.Storage;

namespace CyberSaaS.Terminal.Infrastructure.Health;

public sealed class TerminalHealthReportingService
{
    private readonly TerminalDiagnosticsService _diagnosticsService;

    private readonly TerminalHealthApi _healthApi;

    public TerminalHealthReportingService(
        TerminalDiagnosticsService diagnosticsService,
        TerminalHealthApi healthApi)
    {
        _diagnosticsService =
            diagnosticsService;

        _healthApi =
            healthApi;
    }

    public async Task<TerminalHealthResponse?> CheckAndReportAsync(
        TerminalRegistration registration,
        string apiBaseUrl)
    {
        TerminalDiagnosticsResult diagnostics =
            await _diagnosticsService.RunAsync(
                apiBaseUrl);

        TerminalHealthReportRequest request =
            BuildReport(diagnostics);

        try
        {
            return await _healthApi.ReportHealthAsync(
                registration,
                request);
        }
        catch
        {
            // The terminal may be offline.
            //
            // Health reporting must NEVER stop:
            // - customer lookup
            // - sessions
            // - payments
            // - receipts
            // - offline queue
            // - terminal operation
            return null;
        }
    }

    private static TerminalHealthReportRequest BuildReport(
        TerminalDiagnosticsResult diagnostics)
    {
        List<string> issues =
            new();

        // =========================================================
        // NETWORK
        // =========================================================

        if (!diagnostics.NetworkAdapterAvailable)
        {
            issues.Add(
                "Network adapter unavailable.");
        }

        if (!diagnostics.LANAvailable)
        {
            issues.Add(
                "LAN connection unavailable.");
        }

        if (!diagnostics.InternetAvailable)
        {
            issues.Add(
                "Internet connection unavailable.");
        }

        if (!diagnostics.DnsAvailable)
        {
            issues.Add(
                "DNS resolution unavailable.");
        }

        if (!diagnostics.ApiAvailable)
        {
            issues.Add(
                "CyberSaaS API unavailable.");
        }

        if (diagnostics.BackendLatencyMS >= 2000)
        {
            issues.Add(
                "CyberSaaS API latency is high.");
        }

        // =========================================================
        // DATABASE
        // =========================================================

        if (!diagnostics.DatabaseAvailable)
        {
            issues.Add(
                "Local database unavailable.");
        }

        // =========================================================
        // CPU
        // =========================================================

        if (diagnostics.CPUUsagePercent >= 95)
        {
            issues.Add(
                $"CPU usage is high ({diagnostics.CPUUsagePercent:F1}%).");
        }

        // =========================================================
        // MEMORY
        // =========================================================

        if (!diagnostics.MemoryAvailable)
        {
            issues.Add(
                $"Memory usage is high ({diagnostics.MemoryUsagePercent:F1}%).");
        }

        // =========================================================
        // DISK
        // =========================================================

        if (!diagnostics.DiskAvailable)
        {
            issues.Add(
                "Disk space is low.");
        }

        if (diagnostics.DiskTotalBytes > 0)
        {
            double diskUsage =
                diagnostics.DiskUsedBytes *
                100.0 /
                diagnostics.DiskTotalBytes;

            if (diskUsage >= 95)
            {
                issues.Add(
                    $"Disk usage is high ({diskUsage:F1}%).");
            }
        }

        // =========================================================
        // PRINTER
        // =========================================================

        if (!diagnostics.PrinterAvailable)
        {
            issues.Add(
                "No available printer was detected.");
        }

        // =========================================================
        // SOUND
        // =========================================================

        if (!diagnostics.SoundAvailable)
        {
            if (!string.IsNullOrWhiteSpace(
                    diagnostics.SoundIssue))
            {
                issues.Add(
                    diagnostics.SoundIssue);
            }
            else
            {
                issues.Add(
                    "Sound output is unavailable.");
            }
        }

        // =========================================================
        // DRIVERS
        // =========================================================

        if (!diagnostics.DriversAvailable)
        {
            if (!string.IsNullOrWhiteSpace(
                    diagnostics.DriverIssue))
            {
                issues.Add(
                    diagnostics.DriverIssue);
            }
            else
            {
                issues.Add(
                    "One or more device driver problems were detected.");
            }
        }

        // =========================================================
        // SECURITY
        // =========================================================

        if (!diagnostics.SecurityAvailable)
        {
            if (!string.IsNullOrWhiteSpace(
                    diagnostics.SecurityIssue))
            {
                issues.Add(
                    diagnostics.SecurityIssue);
            }
            else
            {
                issues.Add(
                    "Security protection requires attention.");
            }
        }

        if (!diagnostics.AntivirusEnabled)
        {
            issues.Add(
                "Active antivirus protection could not be confirmed.");
        }

        // =========================================================
        // SYNC
        // =========================================================

        if (diagnostics.SyncPendingCount > 0)
        {
            issues.Add(
                $"{diagnostics.SyncPendingCount} operation(s) pending synchronization.");
        }

        if (diagnostics.SyncFailedCount > 0)
        {
            issues.Add(
                $"{diagnostics.SyncFailedCount} synchronization operation(s) failed.");
        }

        // =========================================================
        // NETWORK MESSAGE
        // =========================================================

        if (!string.IsNullOrWhiteSpace(
                diagnostics.NetworkIssue) &&
            diagnostics.NetworkIssue !=
                "No basic network problem detected.")
        {
            issues.Add(
                diagnostics.NetworkIssue);
        }

        // =========================================================
        // REMOVE DUPLICATES
        // =========================================================

        issues =
            issues
                .Where(
                    x =>
                        !string.IsNullOrWhiteSpace(x))
                .Distinct(
                    StringComparer.OrdinalIgnoreCase)
                .ToList();

        // =========================================================
        // STATUS
        // =========================================================

        string status =
            DetermineStatus(
                diagnostics,
                issues);

        // =========================================================
        // HEALTH REQUEST
        // =========================================================

        return new TerminalHealthReportRequest
        {
            Status =
                status,

            HealthMessage =
                BuildHealthMessage(
                    diagnostics,
                    status),

            Issues =
                issues.ToArray(),

            // -----------------------------------------------------
            // CPU
            // -----------------------------------------------------

            CPUUsagePercent =
                diagnostics.CPUUsagePercent,

            // -----------------------------------------------------
            // MEMORY
            // -----------------------------------------------------

            MemoryAvailable =
                diagnostics.MemoryAvailable,

            MemoryTotal =
                diagnostics.MemoryTotalBytes,

            MemoryUsed =
                diagnostics.MemoryUsedBytes,

            MemoryFree =
                diagnostics.MemoryFreeBytes,

            // -----------------------------------------------------
            // DISK
            // -----------------------------------------------------

            DiskAvailable =
                diagnostics.DiskAvailable,

            DiskTotal =
                diagnostics.DiskTotalBytes,

            DiskFree =
                diagnostics.DiskFreeBytes,

            // -----------------------------------------------------
            // NETWORK
            // -----------------------------------------------------

            NetworkAdapterAvailable =
                diagnostics.NetworkAdapterAvailable,

            LANAvailable =
                diagnostics.LANAvailable,

            InternetAvailable =
                diagnostics.InternetAvailable,

            DNSAvailable =
                diagnostics.DnsAvailable,

            APIAvailable =
                diagnostics.ApiAvailable,

            BackendLatencyMS =
                diagnostics.BackendLatencyMS,

            NetworkIssue =
                diagnostics.NetworkIssue,

            // -----------------------------------------------------
            // LOCAL SERVICES
            // -----------------------------------------------------

            DatabaseAvailable =
                diagnostics.DatabaseAvailable,

            PrinterAvailable =
                diagnostics.PrinterAvailable,

            SoundAvailable =
                diagnostics.SoundAvailable,

            SoundIssue =
                diagnostics.SoundIssue,

            DriversAvailable =
                diagnostics.DriversAvailable,

            DriverIssue =
                diagnostics.DriverIssue,

            // -----------------------------------------------------
            // SECURITY
            // -----------------------------------------------------

            SecurityAvailable =
                diagnostics.SecurityAvailable,

            AntivirusEnabled =
                diagnostics.AntivirusEnabled,

            SecurityIssue =
                diagnostics.SecurityIssue,

            // -----------------------------------------------------
            // OPERATING SYSTEM
            // -----------------------------------------------------

            OSName =
                diagnostics.OSName,

            OSVersion =
                diagnostics.OSVersion,

            OSArchitecture =
                diagnostics.OSArchitecture,

            // -----------------------------------------------------
            // CLIENT
            // -----------------------------------------------------

            ClientVersion =
                diagnostics.ClientVersion,

            ClientStatus =
                diagnostics.ClientStatus,

            // -----------------------------------------------------
            // OFFLINE / SYNC
            // -----------------------------------------------------

            OfflineQueueCount =
                diagnostics.OfflineQueueCount,

            SyncPendingCount =
                diagnostics.SyncPendingCount,

            SyncFailedCount =
                diagnostics.SyncFailedCount,

            SyncStatus =
                diagnostics.SyncStatus,

            LastSyncAt =
                diagnostics.LastSyncAt,

            // -----------------------------------------------------
            // UPTIME
            // -----------------------------------------------------

            UptimeSeconds =
                (long)
                diagnostics
                    .SystemUptime
                    .TotalSeconds
        };
    }

    private static string DetermineStatus(
        TerminalDiagnosticsResult diagnostics,
        List<string> issues)
    {
        // ---------------------------------------------------------
        // TRUE LOCAL OFFLINE
        //
        // The terminal has lost network connectivity AND its local
        // database is unavailable.
        // ---------------------------------------------------------

        if (!diagnostics.NetworkAdapterAvailable &&
            !diagnostics.DatabaseAvailable)
        {
            return "offline";
        }

        // ---------------------------------------------------------
        // LOCAL TERMINAL IS FUNCTIONAL
        //
        // Internet/API failure alone should NOT make the terminal
        // unusable because CyberSaaS is offline-first.
        // ---------------------------------------------------------

        if (issues.Count > 0)
        {
            return "attention";
        }

        return "healthy";
    }

    private static string BuildHealthMessage(
        TerminalDiagnosticsResult diagnostics,
        string status)
    {
        if (status == "offline")
        {
            return
                "Terminal is offline and requires connectivity or local database attention.";
        }

        if (status == "attention")
        {
            if (!diagnostics.InternetAvailable)
            {
                return
                    "Terminal is operational but internet connectivity requires attention.";
            }

            if (!diagnostics.ApiAvailable)
            {
                return
                    "Terminal is operational but the CyberSaaS API is unreachable.";
            }

            if (diagnostics.SyncFailedCount > 0)
            {
                return
                    "Terminal is operational but synchronization failures require attention.";
            }

            return
                "Terminal is operational but one or more health checks require attention.";
        }

        return
            "Terminal is operating normally.";
    }
}