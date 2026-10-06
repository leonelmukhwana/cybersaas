using System;

namespace CyberSaaS.Terminal.Infrastructure.Database;

public sealed class ActiveSessionLocal
{
    public int Id { get; set; }

    public string SessionId { get; set; } =
        string.Empty;

    public string CustomerId { get; set; } =
        string.Empty;

    public string TerminalId { get; set; } =
        string.Empty;

    public string BranchId { get; set; } =
        string.Empty;

    public string SessionType { get; set; } =
        string.Empty;

    /*
     * running
     * paused
     * completed
     */
    public string State { get; set; } =
        "running";

    public DateTime StartedAtUtc { get; set; }

    public DateTime? PausedAtUtc { get; set; }

    public DateTime? EndedAtUtc { get; set; }

    public long TotalPausedSeconds { get; set; }

    public string? PrepaidAmount { get; set; }

    public int? AllowedMinutes { get; set; }

    public string? RatePerMinute { get; set; }

    public string? MinimumCharge { get; set; }

    public string ClientOperationId { get; set; } =
        string.Empty;

    public DateTime UpdatedAtUtc { get; set; }
}