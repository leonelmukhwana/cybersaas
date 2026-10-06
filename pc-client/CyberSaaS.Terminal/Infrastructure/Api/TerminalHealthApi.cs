using System;
using System.Net.Http;
using System.Net.Http.Json;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading.Tasks;
using CyberSaaS.Terminal.Storage;

namespace CyberSaaS.Terminal.Infrastructure.Api;

public sealed class TerminalHealthApi
{
    private readonly HttpClient _httpClient;

    public TerminalHealthApi(HttpClient httpClient)
    {
        _httpClient = httpClient;
    }

    public async Task<TerminalHealthResponse?> ReportHealthAsync(
        TerminalRegistration registration,
        TerminalHealthReportRequest request)
    {
        if (registration == null)
        {
            throw new ArgumentNullException(
                nameof(registration));
        }

        if (request == null)
        {
            throw new ArgumentNullException(
                nameof(request));
        }

        using HttpRequestMessage httpRequest =
            CreateAuthorizedRequest(
                HttpMethod.Post,
                "terminals/health",
                registration);

        httpRequest.Content =
            JsonContent.Create(
                request,
                options: JsonOptions);

        HttpResponseMessage response =
            await _httpClient.SendAsync(
                httpRequest);

        string body =
            await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
        {
            throw new HttpRequestException(
                ExtractError(body));
        }

        if (string.IsNullOrWhiteSpace(body))
        {
            return null;
        }

        TerminalHealthResponse? result =
            JsonSerializer.Deserialize<
                TerminalHealthResponse>(
                    body,
                    JsonOptions);

        return result;
    }

    private static HttpRequestMessage CreateAuthorizedRequest(
        HttpMethod method,
        string endpoint,
        TerminalRegistration registration)
    {
        HttpRequestMessage request =
            new HttpRequestMessage(
                method,
                endpoint);

        request.Headers.Authorization =
            new System.Net.Http.Headers
                .AuthenticationHeaderValue(
                    "Bearer",
                    registration.Credential);

        return request;
    }

    private static string ExtractError(
        string body)
    {
        if (string.IsNullOrWhiteSpace(body))
        {
            return "The server returned an error.";
        }

        try
        {
            using JsonDocument document =
                JsonDocument.Parse(body);

            if (document.RootElement.TryGetProperty(
                    "error",
                    out JsonElement errorElement))
            {
                string? error =
                    errorElement.GetString();

                if (!string.IsNullOrWhiteSpace(error))
                {
                    return error;
                }
            }
        }
        catch
        {
            // Return the response body below.
        }

        return body;
    }

    private static readonly JsonSerializerOptions
        JsonOptions =
            new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true
            };
}

// =================================================================
// HEALTH REPORT REQUEST
// =================================================================

public sealed class TerminalHealthReportRequest
{
    [JsonPropertyName("status")]
    public string Status { get; set; } =
        "healthy";

    [JsonPropertyName("health_message")]
    public string HealthMessage { get; set; } =
        string.Empty;

    [JsonPropertyName("issues")]
    public string[] Issues { get; set; } =
        Array.Empty<string>();

    // -------------------------------------------------------------
    // CPU
    // -------------------------------------------------------------

    [JsonPropertyName("cpu_usage_percent")]
    public double CPUUsagePercent { get; set; }

    // -------------------------------------------------------------
    // MEMORY
    // -------------------------------------------------------------

    [JsonPropertyName("memory_available")]
    public bool MemoryAvailable { get; set; }

    [JsonPropertyName("memory_total_bytes")]
    public long MemoryTotal { get; set; }

    [JsonPropertyName("memory_used_bytes")]
    public long MemoryUsed { get; set; }

    [JsonPropertyName("memory_free_bytes")]
    public long MemoryFree { get; set; }

    // -------------------------------------------------------------
    // DISK
    // -------------------------------------------------------------

    [JsonPropertyName("disk_available")]
    public bool DiskAvailable { get; set; }

    [JsonPropertyName("disk_total_bytes")]
    public long DiskTotal { get; set; }

    [JsonPropertyName("disk_free_bytes")]
    public long DiskFree { get; set; }

    // -------------------------------------------------------------
    // NETWORK
    // -------------------------------------------------------------

    [JsonPropertyName("network_adapter_available")]
    public bool NetworkAdapterAvailable { get; set; }

    [JsonPropertyName("lan_available")]
    public bool LANAvailable { get; set; }

    [JsonPropertyName("internet_available")]
    public bool InternetAvailable { get; set; }

    [JsonPropertyName("dns_available")]
    public bool DNSAvailable { get; set; }

    [JsonPropertyName("api_available")]
    public bool APIAvailable { get; set; }

    [JsonPropertyName("backend_latency_ms")]
    public int BackendLatencyMS { get; set; }

    [JsonPropertyName("network_issue")]
    public string NetworkIssue { get; set; } =
        string.Empty;

    // -------------------------------------------------------------
    // DATABASE
    // -------------------------------------------------------------

    [JsonPropertyName("database_available")]
    public bool DatabaseAvailable { get; set; }

    // -------------------------------------------------------------
    // PRINTER
    // -------------------------------------------------------------

    [JsonPropertyName("printer_available")]
    public bool PrinterAvailable { get; set; }

    // -------------------------------------------------------------
    // SOUND
    // -------------------------------------------------------------

    [JsonPropertyName("sound_available")]
    public bool SoundAvailable { get; set; }

    [JsonPropertyName("sound_issue")]
    public string SoundIssue { get; set; } =
        string.Empty;

    // -------------------------------------------------------------
    // DRIVERS
    // -------------------------------------------------------------

    [JsonPropertyName("drivers_available")]
    public bool DriversAvailable { get; set; }

    [JsonPropertyName("driver_issue")]
    public string DriverIssue { get; set; } =
        string.Empty;

    // -------------------------------------------------------------
    // SECURITY
    // -------------------------------------------------------------

    [JsonPropertyName("security_available")]
    public bool SecurityAvailable { get; set; }

    [JsonPropertyName("antivirus_enabled")]
    public bool AntivirusEnabled { get; set; }

    [JsonPropertyName("security_issue")]
    public string SecurityIssue { get; set; } =
        string.Empty;

    // -------------------------------------------------------------
    // OPERATING SYSTEM
    // -------------------------------------------------------------

    [JsonPropertyName("os_name")]
    public string OSName { get; set; } =
        string.Empty;

    [JsonPropertyName("os_version")]
    public string OSVersion { get; set; } =
        string.Empty;

    [JsonPropertyName("os_architecture")]
    public string OSArchitecture { get; set; } =
        string.Empty;

    // -------------------------------------------------------------
    // CYBERSAAS CLIENT
    // -------------------------------------------------------------

    [JsonPropertyName("client_version")]
    public string ClientVersion { get; set; } =
        string.Empty;

    [JsonPropertyName("client_status")]
    public string ClientStatus { get; set; } =
        "running";

    // -------------------------------------------------------------
    // OFFLINE / SYNC
    // -------------------------------------------------------------

    [JsonPropertyName("offline_queue_count")]
    public int OfflineQueueCount { get; set; }

    [JsonPropertyName("sync_pending_count")]
    public int SyncPendingCount { get; set; }

    [JsonPropertyName("sync_failed_count")]
    public int SyncFailedCount { get; set; }

    [JsonPropertyName("sync_status")]
    public string SyncStatus { get; set; } =
        "unknown";

    [JsonPropertyName("last_sync_at")]
    public DateTime? LastSyncAt { get; set; }

    // -------------------------------------------------------------
    // UPTIME
    // -------------------------------------------------------------

    [JsonPropertyName("uptime_seconds")]
    public long UptimeSeconds { get; set; }
}

// =================================================================
// HEALTH RESPONSE
// =================================================================

public sealed class TerminalHealthResponse
{
    [JsonPropertyName("health")]
    public TerminalHealthData? Health { get; set; }
}

// =================================================================
// HEALTH DATA
// =================================================================

public sealed class TerminalHealthData
{
    [JsonPropertyName("terminal_id")]
    public string TerminalId { get; set; } =
        string.Empty;

    [JsonPropertyName("status")]
    public string Status { get; set; } =
        "healthy";

    [JsonPropertyName("health_message")]
    public string HealthMessage { get; set; } =
        string.Empty;

    [JsonPropertyName("issues")]
    public string[] Issues { get; set; } =
        Array.Empty<string>();

    // -------------------------------------------------------------
    // CPU
    // -------------------------------------------------------------

    [JsonPropertyName("cpu_usage_percent")]
    public double CPUUsagePercent { get; set; }

    // -------------------------------------------------------------
    // MEMORY
    // -------------------------------------------------------------

    [JsonPropertyName("memory_available")]
    public bool MemoryAvailable { get; set; }

    [JsonPropertyName("memory_total_bytes")]
    public long MemoryTotal { get; set; }

    [JsonPropertyName("memory_used_bytes")]
    public long MemoryUsed { get; set; }

    [JsonPropertyName("memory_free_bytes")]
    public long MemoryFree { get; set; }

    // -------------------------------------------------------------
    // DISK
    // -------------------------------------------------------------

    [JsonPropertyName("disk_available")]
    public bool DiskAvailable { get; set; }

    [JsonPropertyName("disk_total_bytes")]
    public long DiskTotal { get; set; }

    [JsonPropertyName("disk_free_bytes")]
    public long DiskFree { get; set; }

    // -------------------------------------------------------------
    // NETWORK
    // -------------------------------------------------------------

    [JsonPropertyName("network_adapter_available")]
    public bool NetworkAdapterAvailable { get; set; }

    [JsonPropertyName("lan_available")]
    public bool LANAvailable { get; set; }

    [JsonPropertyName("internet_available")]
    public bool InternetAvailable { get; set; }

    [JsonPropertyName("dns_available")]
    public bool DNSAvailable { get; set; }

    [JsonPropertyName("api_available")]
    public bool APIAvailable { get; set; }

    [JsonPropertyName("backend_latency_ms")]
    public int BackendLatencyMS { get; set; }

    [JsonPropertyName("network_issue")]
    public string NetworkIssue { get; set; } =
        string.Empty;

    // -------------------------------------------------------------
    // DATABASE
    // -------------------------------------------------------------

    [JsonPropertyName("database_available")]
    public bool DatabaseAvailable { get; set; }

    // -------------------------------------------------------------
    // PRINTER
    // -------------------------------------------------------------

    [JsonPropertyName("printer_available")]
    public bool PrinterAvailable { get; set; }

    // -------------------------------------------------------------
    // SOUND
    // -------------------------------------------------------------

    [JsonPropertyName("sound_available")]
    public bool SoundAvailable { get; set; }

    [JsonPropertyName("sound_issue")]
    public string SoundIssue { get; set; } =
        string.Empty;

    // -------------------------------------------------------------
    // DRIVERS
    // -------------------------------------------------------------

    [JsonPropertyName("drivers_available")]
    public bool DriversAvailable { get; set; }

    [JsonPropertyName("driver_issue")]
    public string DriverIssue { get; set; } =
        string.Empty;

    // -------------------------------------------------------------
    // SECURITY
    // -------------------------------------------------------------

    [JsonPropertyName("security_available")]
    public bool SecurityAvailable { get; set; }

    [JsonPropertyName("antivirus_enabled")]
    public bool AntivirusEnabled { get; set; }

    [JsonPropertyName("security_issue")]
    public string SecurityIssue { get; set; } =
        string.Empty;

    // -------------------------------------------------------------
    // OPERATING SYSTEM
    // -------------------------------------------------------------

    [JsonPropertyName("os_name")]
    public string OSName { get; set; } =
        string.Empty;

    [JsonPropertyName("os_version")]
    public string OSVersion { get; set; } =
        string.Empty;

    [JsonPropertyName("os_architecture")]
    public string OSArchitecture { get; set; } =
        string.Empty;

    // -------------------------------------------------------------
    // CLIENT
    // -------------------------------------------------------------

    [JsonPropertyName("client_version")]
    public string ClientVersion { get; set; } =
        string.Empty;

    [JsonPropertyName("client_status")]
    public string ClientStatus { get; set; } =
        "running";

    // -------------------------------------------------------------
    // OFFLINE / SYNC
    // -------------------------------------------------------------

    [JsonPropertyName("offline_queue_count")]
    public int OfflineQueueCount { get; set; }

    [JsonPropertyName("sync_pending_count")]
    public int SyncPendingCount { get; set; }

    [JsonPropertyName("sync_failed_count")]
    public int SyncFailedCount { get; set; }

    [JsonPropertyName("sync_status")]
    public string SyncStatus { get; set; } =
        "unknown";

    [JsonPropertyName("last_sync_at")]
    public DateTime? LastSyncAt { get; set; }

    // -------------------------------------------------------------
    // UPTIME
    // -------------------------------------------------------------

    [JsonPropertyName("uptime_seconds")]
    public long UptimeSeconds { get; set; }

    // -------------------------------------------------------------
    // SERVER TIMESTAMPS
    // -------------------------------------------------------------

    [JsonPropertyName("last_health_check")]
    public DateTime LastHealthCheck { get; set; }

    [JsonPropertyName("last_seen_at")]
    public DateTime LastSeenAt { get; set; }

    [JsonPropertyName("created_at")]
    public DateTime CreatedAt { get; set; }

    [JsonPropertyName("updated_at")]
    public DateTime UpdatedAt { get; set; }
}