using System;

namespace CyberSaaS.Terminal.Infrastructure.Database;

public sealed class LocalBillingResult
{
    public long ElapsedMinutes { get; init; }

    public long BillableMinutes { get; init; }

    public decimal RatePerMinute { get; init; }

    public decimal MinimumCharge { get; init; }

    public decimal CalculatedAmount { get; init; }

    public string Currency { get; init; } = "KES";
}

public static class LocalBillingCalculator
{
    public static LocalBillingResult Calculate(
        BranchBillingConfigLocal config,
        DateTime startedAtUtc,
        DateTime endedAtUtc)
    {
        if (endedAtUtc < startedAtUtc)
        {
            throw new ArgumentException(
                "Ended time cannot be before started time.");
        }

        decimal rate =
            ParseMoney(config.RatePerMinute);

        decimal minimum =
            ParseMoney(config.MinimumCharge);

        TimeSpan elapsed =
            endedAtUtc - startedAtUtc;

        long elapsedMinutes =
            (long)Math.Floor(
                elapsed.TotalMinutes);

        long billableMinutes =
            elapsedMinutes;

        decimal calculatedAmount =
            billableMinutes * rate;

        if (calculatedAmount < minimum)
        {
            calculatedAmount = minimum;
        }

        return new LocalBillingResult
        {
            ElapsedMinutes = elapsedMinutes,
            BillableMinutes = billableMinutes,
            RatePerMinute = rate,
            MinimumCharge = minimum,
            CalculatedAmount =
                decimal.Round(
                    calculatedAmount,
                    2,
                    MidpointRounding.AwayFromZero),
            Currency =
                string.IsNullOrWhiteSpace(
                    config.Currency)
                    ? "KES"
                    : config.Currency
        };
    }

    private static decimal ParseMoney(
        string value)
    {
        if (string.IsNullOrWhiteSpace(value))
        {
            return 0m;
        }

        if (!decimal.TryParse(
                value,
                System.Globalization.NumberStyles.Number,
                System.Globalization.CultureInfo.InvariantCulture,
                out decimal result))
        {
            return 0m;
        }

        return result;
    }
}