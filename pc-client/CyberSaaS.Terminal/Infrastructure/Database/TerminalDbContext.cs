using System;
using Microsoft.EntityFrameworkCore;

namespace CyberSaaS.Terminal.Infrastructure.Database;

public sealed class TerminalDbContext : DbContext
{
    public DbSet<TerminalLocalState> TerminalLocalStates { get; set; } = null!;

    public DbSet<OfflineQueueItem> OfflineQueueItems { get; set; } = null!;

    public DbSet<ActiveSessionLocal> ActiveSessions { get; set; } = null!;

    public DbSet<TerminalCustomerLocal> TerminalCustomers { get; set; } = null!;

    public DbSet<BranchBillingConfigLocal> BranchBillingConfigs { get; set; } = null!;

    protected override void OnConfiguring(
        DbContextOptionsBuilder optionsBuilder)
    {
        string appData =
            Environment.GetFolderPath(
                Environment.SpecialFolder.LocalApplicationData);

        string directory =
            System.IO.Path.Combine(
                appData,
                "CyberSaaS",
                "Terminal");

        System.IO.Directory.CreateDirectory(directory);

        string databasePath =
            System.IO.Path.Combine(
                directory,
                "terminal.db");

        optionsBuilder.UseSqlite(
            $"Data Source={databasePath}");
    }

    protected override void OnModelCreating(
        ModelBuilder modelBuilder)
    {
        base.OnModelCreating(modelBuilder);

        modelBuilder.Entity<TerminalLocalState>(
            entity =>
            {
                entity.ToTable("terminal_local_state");

                entity.HasKey(x => x.Id);

                entity.Property(x => x.TerminalId)
                    .IsRequired();

                entity.Property(x => x.TenantId)
                    .IsRequired();

                entity.Property(x => x.BranchId)
                    .IsRequired();

                entity.Property(x => x.TerminalCode)
                    .IsRequired();

                entity.Property(x => x.DesiredState)
                    .IsRequired();
            });

        modelBuilder.Entity<OfflineQueueItem>(
            entity =>
            {
                entity.ToTable("offline_queue");

                entity.HasKey(x => x.Id);

                entity.HasIndex(x => x.OperationId)
                    .IsUnique();

                entity.Property(x => x.OperationId)
                    .IsRequired();

                entity.Property(x => x.OperationType)
                    .IsRequired();

                entity.Property(x => x.Payload)
                    .IsRequired();

                entity.Property(x => x.Status)
                    .IsRequired();
            });

        modelBuilder.Entity<ActiveSessionLocal>(
            entity =>
            {
                entity.ToTable("active_session");

                entity.HasKey(x => x.Id);

                entity.HasIndex(x => x.SessionId)
                    .IsUnique();

                entity.HasIndex(x => x.TerminalId);

                entity.HasIndex(x => x.State);

                entity.Property(x => x.SessionId)
                    .IsRequired();

                entity.Property(x => x.CustomerId)
                    .IsRequired();

                entity.Property(x => x.TerminalId)
                    .IsRequired();

                entity.Property(x => x.BranchId)
                    .IsRequired();

                entity.Property(x => x.SessionType)
                    .IsRequired();

                entity.Property(x => x.State)
                    .IsRequired();

                entity.Property(x => x.ClientOperationId)
                    .IsRequired();
            });

        modelBuilder.Entity<TerminalCustomerLocal>(
            entity =>
            {
                entity.ToTable("terminal_customer");

                entity.HasKey(x => x.Id);

                entity.HasIndex(x => new
                {
                    x.BranchId,
                    x.IdNumber
                });

                entity.HasIndex(x => new
                {
                    x.BranchId,
                    x.FullName
                });

                entity.HasIndex(x => x.CustomerId)
                    .IsUnique();

                entity.Property(x => x.CustomerId)
                    .IsRequired();

                entity.Property(x => x.BranchId)
                    .IsRequired();

                entity.Property(x => x.CustomerType)
                    .IsRequired();

                entity.Property(x => x.FullName)
                    .IsRequired();

                entity.Property(x => x.IdNumber)
                    .IsRequired(false);

                entity.Property(x => x.ParentId)
                    .IsRequired(false);

                entity.Property(x => x.ParentName)
                    .IsRequired(false);
            });

        modelBuilder.Entity<BranchBillingConfigLocal>(
            entity =>
            {
                entity.ToTable("branch_billing_config");

                entity.HasKey(x => x.Id);

                entity.HasIndex(x => x.BranchId)
                    .IsUnique();

                entity.Property(x => x.BranchId)
                    .IsRequired();

                entity.Property(x => x.RatePerMinute)
                    .IsRequired();

                entity.Property(x => x.MinimumCharge)
                    .IsRequired();

                entity.Property(x => x.Currency)
                    .IsRequired();
            });
    }
}