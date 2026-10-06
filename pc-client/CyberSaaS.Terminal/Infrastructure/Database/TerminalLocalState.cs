using System;

namespace CyberSaaS.Terminal.Infrastructure.Database;

public sealed class TerminalLocalState
{
    public int Id { get; set; }

    public string TerminalId { get; set; } = string.Empty;

    public string TenantId { get; set; } = string.Empty;

    public string BranchId { get; set; } = string.Empty;

    public string TerminalCode { get; set; } = string.Empty;

    public bool IsOnline { get; set; }

    public string DesiredState { get; set; } = "unlocked";

    public DateTime? LastSuccessfulHeartbeatUtc { get; set; }

    public DateTime UpdatedAtUtc { get; set; }
}