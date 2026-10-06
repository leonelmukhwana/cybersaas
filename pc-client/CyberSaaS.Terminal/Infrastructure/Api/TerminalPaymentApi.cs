using System.Net.Http;
using System.Net.Http.Headers;
using System.Net.Http.Json;
using System.Text.Json.Serialization;

namespace CyberSaaS.Terminal.Infrastructure.Api;

// ============================================================
// CASH PAYMENT
// ============================================================

public sealed class TerminalCashPaymentRequest
{
    [JsonPropertyName("session_id")]
    public string SessionId { get; set; } = string.Empty;

    [JsonPropertyName("amount")]
    public string Amount { get; set; } = string.Empty;

    [JsonPropertyName("client_operation_id")]
    public string ClientOperationId { get; set; } = string.Empty;
}

public sealed class TerminalCashPaymentResponse
{
    [JsonPropertyName("payment_id")]
    public string PaymentId { get; set; } = string.Empty;

    [JsonPropertyName("sale_id")]
    public string SaleId { get; set; } = string.Empty;

    [JsonPropertyName("session_id")]
    public string SessionId { get; set; } = string.Empty;

    [JsonPropertyName("amount")]
    public string Amount { get; set; } = string.Empty;

    [JsonPropertyName("method")]
    public string Method { get; set; } = string.Empty;

    [JsonPropertyName("status")]
    public string Status { get; set; } = string.Empty;

    [JsonPropertyName("currency")]
    public string Currency { get; set; } = "KES";

    [JsonPropertyName("confirmed_at")]
    public DateTime? ConfirmedAt { get; set; }
}

// ============================================================
// M-PESA REQUEST
// ============================================================

public sealed class TerminalMpesaPaymentRequest
{
    [JsonPropertyName("session_id")]
    public string SessionId { get; set; } = string.Empty;

    [JsonPropertyName("amount")]
    public string Amount { get; set; } = string.Empty;

    [JsonPropertyName("phone_number")]
    public string PhoneNumber { get; set; } = string.Empty;

    [JsonPropertyName("client_operation_id")]
    public string ClientOperationId { get; set; } = string.Empty;
}

// ============================================================
// M-PESA STK RESPONSE
// ============================================================

public sealed class MpesaStkResponse
{
    [JsonPropertyName("payment_id")]
    public string PaymentId { get; set; } = string.Empty;

    [JsonPropertyName("sale_id")]
    public string SaleId { get; set; } = string.Empty;

    [JsonPropertyName("session_id")]
    public string SessionId { get; set; } = string.Empty;

    [JsonPropertyName("amount")]
    public string Amount { get; set; } = string.Empty;

    [JsonPropertyName("method")]
    public string Method { get; set; } = string.Empty;

    [JsonPropertyName("status")]
    public string Status { get; set; } = string.Empty;

    [JsonPropertyName("currency")]
    public string Currency { get; set; } = "KES";

    [JsonPropertyName("phone_number")]
    public string PhoneNumber { get; set; } = string.Empty;

    [JsonPropertyName("checkout_request_id")]
    public string CheckoutRequestId { get; set; } = string.Empty;

    [JsonPropertyName("merchant_request_id")]
    public string MerchantRequestId { get; set; } = string.Empty;

    [JsonPropertyName("confirmed_at")]
    public DateTime? ConfirmedAt { get; set; }

    [JsonPropertyName("message")]
    public string Message { get; set; } = string.Empty;
}

// ============================================================
// M-PESA PAYMENT STATUS
// ============================================================

public sealed class MpesaPaymentStatusResponse
{
    [JsonPropertyName("payment_id")]
    public string PaymentId { get; set; } = string.Empty;

    [JsonPropertyName("sale_id")]
    public string SaleId { get; set; } = string.Empty;

    [JsonPropertyName("session_id")]
    public string SessionId { get; set; } = string.Empty;

    [JsonPropertyName("amount")]
    public string Amount { get; set; } = string.Empty;

    [JsonPropertyName("method")]
    public string Method { get; set; } = string.Empty;

    [JsonPropertyName("status")]
    public string Status { get; set; } = string.Empty;

    [JsonPropertyName("currency")]
    public string Currency { get; set; } = "KES";

    [JsonPropertyName("receipt_number")]
    public string ReceiptNumber { get; set; } = string.Empty;

    [JsonPropertyName("confirmed_at")]
    public DateTime? ConfirmedAt { get; set; }

    [JsonPropertyName("message")]
    public string Message { get; set; } = string.Empty;
}

// ============================================================
// TERMINAL PAYMENT API
// ============================================================

public sealed class TerminalPaymentApi
{
    private readonly HttpClient _httpClient;
    private readonly string _terminalCredential;

    public TerminalPaymentApi(
        HttpClient httpClient,
        string terminalCredential)
    {
        _httpClient = httpClient;
        _terminalCredential = terminalCredential;
    }

    // ========================================================
    // CASH
    // ========================================================

    public async Task<TerminalCashPaymentResponse> CreateCashPaymentAsync(
        TerminalCashPaymentRequest request)
    {
        using var httpRequest = new HttpRequestMessage(
            HttpMethod.Post,
            "terminal-payments/cash");

        httpRequest.Headers.Authorization =
            new AuthenticationHeaderValue(
                "Bearer",
                _terminalCredential);

        httpRequest.Content = JsonContent.Create(request);

        using HttpResponseMessage response =
            await _httpClient.SendAsync(httpRequest);

        if (!response.IsSuccessStatusCode)
        {
            string body = await response.Content.ReadAsStringAsync();

            throw new HttpRequestException(
                $"Cash payment failed: {(int)response.StatusCode} {body}");
        }

        TerminalCashPaymentResponse? result =
            await response.Content.ReadFromJsonAsync<TerminalCashPaymentResponse>();

        if (result == null)
        {
            throw new InvalidOperationException(
                "Cash payment response was empty.");
        }

        return result;
    }

    // ========================================================
    // M-PESA STK
    // ========================================================

    public async Task<MpesaStkResponse> SendMpesaStkAsync(
        TerminalMpesaPaymentRequest request)
    {
        using var httpRequest = new HttpRequestMessage(
            HttpMethod.Post,
            "terminal-payments/mpesa");

        httpRequest.Headers.Authorization =
            new AuthenticationHeaderValue(
                "Bearer",
                _terminalCredential);

        httpRequest.Content = JsonContent.Create(request);

        using HttpResponseMessage response =
            await _httpClient.SendAsync(httpRequest);

        if (!response.IsSuccessStatusCode)
        {
            string body = await response.Content.ReadAsStringAsync();

            throw new HttpRequestException(
                $"M-Pesa STK request failed: {(int)response.StatusCode} {body}");
        }

        MpesaStkResponse? result =
            await response.Content.ReadFromJsonAsync<MpesaStkResponse>();

        if (result == null)
        {
            throw new InvalidOperationException(
                "M-Pesa STK response was empty.");
        }

        return result;
    }

    // ========================================================
    // M-PESA STK - EXISTING WINDOW OVERLOAD
    // ========================================================

    public Task<MpesaStkResponse> SendMpesaStkAsync(
        string sessionId,
        string amount,
        string phoneNumber)
    {
        var request = new TerminalMpesaPaymentRequest
        {
            SessionId = sessionId,
            Amount = amount,
            PhoneNumber = phoneNumber,
            ClientOperationId = Guid.NewGuid().ToString()
        };

        return SendMpesaStkAsync(request);
    }

    // ========================================================
    // M-PESA STATUS
    // ========================================================

    public async Task<MpesaPaymentStatusResponse> GetMpesaPaymentStatusAsync(
        string paymentId)
    {
        if (string.IsNullOrWhiteSpace(paymentId))
        {
            throw new ArgumentException(
                "Payment ID is required.",
                nameof(paymentId));
        }

        using var httpRequest = new HttpRequestMessage(
            HttpMethod.Get,
            $"terminal-payments/mpesa/{Uri.EscapeDataString(paymentId)}/status");

        httpRequest.Headers.Authorization =
            new AuthenticationHeaderValue(
                "Bearer",
                _terminalCredential);

        using HttpResponseMessage response =
            await _httpClient.SendAsync(httpRequest);

        if (!response.IsSuccessStatusCode)
        {
            string body = await response.Content.ReadAsStringAsync();

            throw new HttpRequestException(
                $"M-Pesa payment status failed: {(int)response.StatusCode} {body}");
        }

        MpesaPaymentStatusResponse? result =
            await response.Content.ReadFromJsonAsync<MpesaPaymentStatusResponse>();

        if (result == null)
        {
            throw new InvalidOperationException(
                "M-Pesa payment status response was empty.");
        }

        return result;
    }

    // ========================================================
    // MODERN ALIAS
    // ========================================================

    public Task<MpesaStkResponse> CreateMpesaPaymentAsync(
        TerminalMpesaPaymentRequest request)
    {
        return SendMpesaStkAsync(request);
    }
}