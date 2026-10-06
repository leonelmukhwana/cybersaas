using System;

namespace CyberSaaS.Terminal.Infrastructure.Database;

public sealed class BranchBillingConfigLocal
{
    public int Id { get; set; }

    public string BranchId { get; set; } =
        string.Empty;

    public string RatePerMinute { get; set; } =
        "0.00";

    public string MinimumCharge { get; set; } =
        "0.00";

    public string Currency { get; set; } =
        "KES";

    public DateTime UpdatedAtUtc { get; set; }
}