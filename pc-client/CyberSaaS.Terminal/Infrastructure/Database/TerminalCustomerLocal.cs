using System;

namespace CyberSaaS.Terminal.Infrastructure.Database;

public sealed class TerminalCustomerLocal
{
    public int Id { get; set; }

    public string CustomerId { get; set; } =
        string.Empty;

    public string BranchId { get; set; } =
        string.Empty;

    public string CustomerType { get; set; } =
        string.Empty;

    public string FullName { get; set; } =
        string.Empty;

    public string? Phone { get; set; }

    public string? IdNumber { get; set; }

    public string? ParentId { get; set; }

    public string? ParentName { get; set; }

    public DateTime UpdatedAtUtc { get; set; }
}