using System;

namespace CyberSaaS.Terminal.Infrastructure.Database;

public sealed class OfflineQueueItem
{
    public long Id { get; set; }

    public string OperationId { get; set; } = string.Empty;

    public string OperationType { get; set; } = string.Empty;

    public string Payload { get; set; } = string.Empty;

    public DateTime CreatedAtUtc { get; set; }

    public int AttemptCount { get; set; }

    public DateTime? LastAttemptAtUtc { get; set; }

    public string Status { get; set; } = "pending";

    public string? LastError { get; set; }
}