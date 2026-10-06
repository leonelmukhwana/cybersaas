using System;
using System.Net.Http;
using System.Net.Http.Json;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading.Tasks;
using CyberSaaS.Terminal.Storage;

namespace CyberSaaS.Terminal.Infrastructure.Api;

public sealed class TerminalSessionApi
{
    private readonly HttpClient _httpClient;

    public TerminalSessionApi(HttpClient httpClient)
    {
        _httpClient = httpClient;
    }

    public async Task<TerminalSessionData?> StartSessionAsync(
        TerminalRegistration registration,
        TerminalStartSessionRequest request)
    {
        using HttpRequestMessage httpRequest =
            CreateAuthorizedRequest(
                HttpMethod.Post,
                "terminal-sessions",
                registration);

        httpRequest.Content =
            JsonContent.Create(request);

        HttpResponseMessage response =
            await _httpClient.SendAsync(httpRequest);

        string body =
            await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
            throw new InvalidOperationException(
                ExtractError(body));

        TerminalSessionApiResponse? result =
            JsonSerializer.Deserialize<TerminalSessionApiResponse>(
                body,
                JsonOptions);

        return result?.Data;
    }

    public async Task<TerminalSessionData?> GetSessionAsync(
        TerminalRegistration registration,
        string sessionId)
    {
        using HttpRequestMessage httpRequest =
            CreateAuthorizedRequest(
                HttpMethod.Get,
                $"terminal-sessions/{sessionId}",
                registration);

        HttpResponseMessage response =
            await _httpClient.SendAsync(httpRequest);

        string body =
            await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
            throw new InvalidOperationException(
                ExtractError(body));

        TerminalSessionApiResponse? result =
            JsonSerializer.Deserialize<TerminalSessionApiResponse>(
                body,
                JsonOptions);

        return result?.Data;
    }

    public async Task<TerminalSessionData?> GetActiveSessionAsync(
        TerminalRegistration registration)
    {
        using HttpRequestMessage httpRequest =
            CreateAuthorizedRequest(
                HttpMethod.Get,
                "terminal-sessions/active",
                registration);

        HttpResponseMessage response =
            await _httpClient.SendAsync(httpRequest);

        string body =
            await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
            throw new InvalidOperationException(
                ExtractError(body));

        TerminalSessionApiResponse? result =
            JsonSerializer.Deserialize<TerminalSessionApiResponse>(
                body,
                JsonOptions);

        return result?.Data;
    }

    public async Task<TerminalSessionData?> PauseSessionAsync(
        TerminalRegistration registration,
        string sessionId)
    {
        using HttpRequestMessage httpRequest =
            CreateAuthorizedRequest(
                HttpMethod.Post,
                $"terminal-sessions/{sessionId}/pause",
                registration);

        HttpResponseMessage response =
            await _httpClient.SendAsync(httpRequest);

        string body =
            await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
            throw new InvalidOperationException(
                ExtractError(body));

        TerminalSessionApiResponse? result =
            JsonSerializer.Deserialize<TerminalSessionApiResponse>(
                body,
                JsonOptions);

        return result?.Data;
    }

    public async Task<TerminalSessionData?> ResumeSessionAsync(
        TerminalRegistration registration,
        string sessionId)
    {
        using HttpRequestMessage httpRequest =
            CreateAuthorizedRequest(
                HttpMethod.Post,
                $"terminal-sessions/{sessionId}/resume",
                registration);

        HttpResponseMessage response =
            await _httpClient.SendAsync(httpRequest);

        string body =
            await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
            throw new InvalidOperationException(
                ExtractError(body));

        TerminalSessionApiResponse? result =
            JsonSerializer.Deserialize<TerminalSessionApiResponse>(
                body,
                JsonOptions);

        return result?.Data;
    }

    public async Task<TerminalBillingPreview?> GetBillingPreviewAsync(
        TerminalRegistration registration,
        string sessionId)
    {
        using HttpRequestMessage httpRequest =
            CreateAuthorizedRequest(
                HttpMethod.Get,
                $"terminal-sessions/{sessionId}/billing-preview",
                registration);

        HttpResponseMessage response =
            await _httpClient.SendAsync(httpRequest);

        string body =
            await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
            throw new InvalidOperationException(
                ExtractError(body));

        TerminalBillingPreviewApiResponse? result =
            JsonSerializer.Deserialize<TerminalBillingPreviewApiResponse>(
                body,
                JsonOptions);

        return result?.Data;
    }

    public async Task<TerminalSessionWithBilling?> EndSessionAsync(
        TerminalRegistration registration,
        string sessionId)
    {
        using HttpRequestMessage httpRequest =
            CreateAuthorizedRequest(
                HttpMethod.Post,
                $"terminal-sessions/{sessionId}/end",
                registration);

        httpRequest.Content =
            JsonContent.Create(
                new TerminalEndSessionRequest());

        HttpResponseMessage response =
            await _httpClient.SendAsync(httpRequest);

        string body =
            await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
            throw new InvalidOperationException(
                ExtractError(body));

        TerminalSessionWithBillingApiResponse? result =
            JsonSerializer.Deserialize<TerminalSessionWithBillingApiResponse>(
                body,
                JsonOptions);

        return result?.Data;
    }

    public async Task<TerminalBillingConfig?> GetBillingConfigAsync(
        TerminalRegistration registration)
    {
        using HttpRequestMessage httpRequest =
            CreateAuthorizedRequest(
                HttpMethod.Get,
                "terminal-sessions/billing-config",
                registration);

        HttpResponseMessage response =
            await _httpClient.SendAsync(httpRequest);

        string body =
            await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
            throw new InvalidOperationException(
                ExtractError(body));

        TerminalBillingConfigApiResponse? result =
            JsonSerializer.Deserialize<TerminalBillingConfigApiResponse>(
                body,
                JsonOptions);

        return result?.Data;
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
            new System.Net.Http.Headers.AuthenticationHeaderValue(
                "Bearer",
                registration.Credential);

        return request;
    }

    private static string ExtractError(
        string body)
    {
        try
        {
            ErrorResponse? error =
                JsonSerializer.Deserialize<ErrorResponse>(
                    body,
                    JsonOptions);

            if (!string.IsNullOrWhiteSpace(
                    error?.Error))
            {
                return error.Error;
            }
        }
        catch
        {
        }

        return string.IsNullOrWhiteSpace(body)
            ? "The server returned an error."
            : body;
    }

    private static readonly JsonSerializerOptions JsonOptions =
        new JsonSerializerOptions
        {
            PropertyNameCaseInsensitive = true
        };
}

public sealed class TerminalStartSessionRequest
{
    [JsonPropertyName("customer_id")]
    public string CustomerId { get; set; } =
        string.Empty;

    [JsonPropertyName("session_type")]
    public string SessionType { get; set; } =
        "pay_after";

    [JsonPropertyName("prepaid_amount")]
    public string? PrepaidAmount { get; set; }

    [JsonPropertyName("started_at")]
    public DateTime? StartedAt { get; set; }

    [JsonPropertyName("client_operation_id")]
    public string ClientOperationId { get; set; } =
        string.Empty;
}

public sealed class TerminalEndSessionRequest
{
    [JsonPropertyName("ended_at")]
    public DateTime? EndedAt { get; set; }
}

public sealed class TerminalSessionApiResponse
{
    [JsonPropertyName("message")]
    public string? Message { get; set; }

    [JsonPropertyName("data")]
    public TerminalSessionData? Data { get; set; }
}

public sealed class TerminalSessionData
{
    [JsonPropertyName("id")]
    public string Id { get; set; } =
        string.Empty;

    [JsonPropertyName("tenant_id")]
    public string TenantId { get; set; } =
        string.Empty;

    [JsonPropertyName("branch_id")]
    public string BranchId { get; set; } =
        string.Empty;

    [JsonPropertyName("terminal_id")]
    public string TerminalId { get; set; } =
        string.Empty;

    [JsonPropertyName("customer_id")]
    public string CustomerId { get; set; } =
        string.Empty;

    [JsonPropertyName("attendant_id")]
    public string? AttendantId { get; set; }

    [JsonPropertyName("session_type")]
    public string SessionType { get; set; } =
        string.Empty;

    [JsonPropertyName("status")]
    public string Status { get; set; } =
        string.Empty;

    [JsonPropertyName("started_at")]
    public DateTime StartedAt { get; set; }

    [JsonPropertyName("ended_at")]
    public DateTime? EndedAt { get; set; }

    [JsonPropertyName("paused_at")]
    public DateTime? PausedAt { get; set; }

    [JsonPropertyName("total_paused_seconds")]
    public long TotalPausedSeconds { get; set; }

    [JsonPropertyName("prepaid_amount")]
    public string? PrepaidAmount { get; set; }

    [JsonPropertyName("allowed_minutes")]
    public int? AllowedMinutes { get; set; }

    [JsonPropertyName("rate_per_minute")]
    public string? RatePerMinute { get; set; }

    [JsonPropertyName("minimum_charge")]
    public string? MinimumCharge { get; set; }

    [JsonPropertyName("final_amount")]
    public string? FinalAmount { get; set; }

    [JsonPropertyName("client_operation_id")]
    public string ClientOperationId { get; set; } =
        string.Empty;

    [JsonPropertyName("created_at")]
    public DateTime CreatedAt { get; set; }

    [JsonPropertyName("updated_at")]
    public DateTime UpdatedAt { get; set; }
}

public sealed class TerminalBillingPreviewApiResponse
{
    [JsonPropertyName("data")]
    public TerminalBillingPreview? Data { get; set; }
}

public sealed class TerminalBillingPreview
{
    [JsonPropertyName("session_id")]
    public string SessionId { get; set; } =
        string.Empty;

    [JsonPropertyName("elapsed_minutes")]
    public long ElapsedMinutes { get; set; }

    [JsonPropertyName("billable_minutes")]
    public long BillableMinutes { get; set; }

    [JsonPropertyName("rate_per_minute")]
    public string RatePerMinute { get; set; } =
        "0.00";

    [JsonPropertyName("minimum_charge")]
    public string MinimumCharge { get; set; } =
        "0.00";

    [JsonPropertyName("calculated_amount")]
    public string CalculatedAmount { get; set; } =
        "0.00";

    [JsonPropertyName("currency")]
    public string Currency { get; set; } =
        "KES";
}

public sealed class TerminalSessionWithBillingApiResponse
{
    [JsonPropertyName("message")]
    public string? Message { get; set; }

    [JsonPropertyName("data")]
    public TerminalSessionWithBilling? Data { get; set; }
}

public sealed class TerminalSessionWithBilling
{
    [JsonPropertyName("id")]
    public string Id { get; set; } =
        string.Empty;

    [JsonPropertyName("status")]
    public string Status { get; set; } =
        string.Empty;

    [JsonPropertyName("started_at")]
    public DateTime StartedAt { get; set; }

    [JsonPropertyName("ended_at")]
    public DateTime? EndedAt { get; set; }

    [JsonPropertyName("final_amount")]
    public string? FinalAmount { get; set; }

    [JsonPropertyName("elapsed_minutes")]
    public long ElapsedMinutes { get; set; }

    [JsonPropertyName("billable_minutes")]
    public long BillableMinutes { get; set; }

    [JsonPropertyName("calculated_amount")]
    public string CalculatedAmount { get; set; } =
        "0.00";

    [JsonPropertyName("currency")]
    public string Currency { get; set; } =
        "KES";
}

public sealed class TerminalBillingConfigApiResponse
{
    [JsonPropertyName("data")]
    public TerminalBillingConfig? Data { get; set; }
}

public sealed class TerminalBillingConfig
{
    [JsonPropertyName("branch_id")]
    public string BranchId { get; set; } =
        string.Empty;

    [JsonPropertyName("rate_per_minute")]
    public string RatePerMinute { get; set; } =
        "0.00";

    [JsonPropertyName("minimum_charge")]
    public string MinimumCharge { get; set; } =
        "0.00";

    [JsonPropertyName("currency")]
    public string Currency { get; set; } =
        "KES";
}

public sealed class ErrorResponse
{
    [JsonPropertyName("error")]
    public string? Error { get; set; }
}