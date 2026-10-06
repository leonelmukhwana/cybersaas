
using System;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading.Tasks;
using CyberSaaS.Terminal.Storage;

namespace CyberSaaS.Terminal.Infrastructure.Api;

public sealed class TerminalCustomerApi
{
    private readonly HttpClient _httpClient;

    public TerminalCustomerApi(
        HttpClient httpClient)
    {
        _httpClient = httpClient;
    }

    public async Task<TerminalCustomerLookupResponse>
        LookupCustomerAsync(
            TerminalRegistration registration,
            string searchType,
            string searchTerm)
    {
        if (registration == null)
        {
            throw new ArgumentNullException(
                nameof(registration));
        }

        if (string.IsNullOrWhiteSpace(searchType))
        {
            throw new ArgumentException(
                "Customer search type is required.",
                nameof(searchType));
        }

        if (string.IsNullOrWhiteSpace(searchTerm))
        {
            throw new ArgumentException(
                "Customer search value is required.",
                nameof(searchTerm));
        }

        string normalizedSearchType =
            searchType.Trim().ToLowerInvariant();

        if (normalizedSearchType != "adult" &&
            normalizedSearchType != "child")
        {
            throw new ArgumentException(
                "Customer search type must be adult or child.",
                nameof(searchType));
        }

        string normalizedSearchTerm =
            searchTerm.Trim();

        var request =
            new TerminalCustomerLookupRequest
            {
                SearchType = normalizedSearchType,
                SearchTerm = normalizedSearchTerm,

                IdNumber =
                    normalizedSearchType == "adult"
                        ? normalizedSearchTerm
                        : null,

                FullName =
                    normalizedSearchType == "child"
                        ? normalizedSearchTerm
                        : null
            };

        string json =
            JsonSerializer.Serialize(
                request,
                JsonOptions);

        using HttpRequestMessage httpRequest =
            new HttpRequestMessage(
                HttpMethod.Post,
                "terminal-customers/lookup");

        httpRequest.Headers.Authorization =
            new System.Net.Http.Headers
                .AuthenticationHeaderValue(
                    "Bearer",
                    registration.Credential);

        httpRequest.Content =
            new StringContent(
                json,
                Encoding.UTF8,
                "application/json");

        HttpResponseMessage response =
            await _httpClient.SendAsync(
                httpRequest);

        string body =
            await response.Content.ReadAsStringAsync();

        if (!response.IsSuccessStatusCode)
        {
            throw new InvalidOperationException(
                ExtractError(body));
        }

        TerminalCustomerLookupResponse? result =
            JsonSerializer.Deserialize<
                TerminalCustomerLookupResponse>(
                    body,
                    JsonOptions);

        if (result == null)
        {
            throw new InvalidOperationException(
                "The server returned an invalid customer response.");
        }

        if (result.Customer == null ||
            string.IsNullOrWhiteSpace(result.Customer.Id))
        {
            throw new InvalidOperationException(
                "Customer was not found.");
        }

        return result;
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

    private static readonly JsonSerializerOptions
        JsonOptions =
            new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true
            };
}

public sealed class TerminalCustomerLookupRequest
{
    [JsonPropertyName("search_type")]
    public string SearchType { get; set; } =
        string.Empty;

    [JsonPropertyName("search_term")]
    public string SearchTerm { get; set; } =
        string.Empty;

    [JsonPropertyName("id_number")]
    public string? IdNumber { get; set; }

    [JsonPropertyName("full_name")]
    public string? FullName { get; set; }
}

public sealed class TerminalCustomerLookupResponse
{
    [JsonPropertyName("customer")]
    public TerminalCustomer? Customer { get; set; }
}

public sealed class TerminalCustomer
{
    [JsonPropertyName("id")]
    public string Id { get; set; } =
        string.Empty;

    [JsonPropertyName("branch_id")]
    public string BranchId { get; set; } =
        string.Empty;

    [JsonPropertyName("customer_type")]
    public string CustomerType { get; set; } =
        string.Empty;

    [JsonPropertyName("full_name")]
    public string FullName { get; set; } =
        string.Empty;

    [JsonPropertyName("phone")]
    public string? Phone { get; set; }

    [JsonPropertyName("id_number_hash")]
    public string? IdNumberHash { get; set; }

    [JsonPropertyName("parent_id")]
    public string? ParentId { get; set; }

    [JsonPropertyName("parent_name")]
    public string? ParentName { get; set; }

    [JsonPropertyName("updated_at")]
    public DateTime UpdatedAt { get; set; }
}

