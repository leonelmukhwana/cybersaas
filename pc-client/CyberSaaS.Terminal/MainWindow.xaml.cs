
using Microsoft.EntityFrameworkCore;
using System;
using System.Diagnostics;
using System.Linq;
using System.Net.Http;
using System.Net.Http.Json;
using System.Text.Json;
using System.Text.Json.Serialization;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Threading;
using CyberSaaS.Terminal.Infrastructure.Api;
using CyberSaaS.Terminal.Infrastructure.Database;
using CyberSaaS.Terminal.Infrastructure.Diagnostics;
using CyberSaaS.Terminal.Infrastructure.Health;
using CyberSaaS.Terminal.Infrastructure.Sync;
using CyberSaaS.Terminal.Storage;
using CyberSaaS.Terminal.UI.Views;

namespace CyberSaaS.Terminal;

public partial class MainWindow : Window
{
    private const string ApiBaseUrl =
        "https://cybersaas.onrender.com/api/";

#if DEBUG
    private const bool IsTestBuild = true;
#else
    private const bool IsTestBuild = false;
#endif

    private readonly HttpClient _httpClient;
    private readonly TerminalSessionApi _terminalSessionApi;
    private readonly TerminalCustomerApi _terminalCustomerApi;
    private readonly OfflineQueueService _offlineQueueService;
    private OfflineQueueSyncService? _offlineQueueSyncService;
    private readonly DispatcherTimer _heartbeatTimer;
    private readonly TerminalDiagnosticsService _diagnosticsService;
    private readonly TerminalHealthReportingService _healthReportingService;

    private TerminalPaymentApi? _terminalPaymentApi;

    private TerminalRegistration? _registration;
    private bool _heartbeatRunning;

    public MainWindow()
    {
        InitializeComponent();

        InitializeDatabase();

        _httpClient = new HttpClient
        {
            BaseAddress = new Uri(ApiBaseUrl),
            Timeout = TimeSpan.FromSeconds(10)
        };

        _terminalSessionApi =
            new TerminalSessionApi(_httpClient);

        _terminalCustomerApi =
            new TerminalCustomerApi(_httpClient);

        _offlineQueueService =
            new OfflineQueueService();

        _diagnosticsService =
            new TerminalDiagnosticsService();

        TerminalHealthApi terminalHealthApi =
            new TerminalHealthApi(_httpClient);

        _healthReportingService =
            new TerminalHealthReportingService(
                _diagnosticsService,
                terminalHealthApi);

        _heartbeatTimer = new DispatcherTimer
        {
            Interval = TimeSpan.FromSeconds(30)
        };

        _heartbeatTimer.Tick += HeartbeatTimer_Tick;

        StartTerminal();
    }

    private void InitializeDatabase()
    {
        using var db = new TerminalDbContext();

        db.Database.EnsureCreated();

        // ---------------------------------------------------------
        // Terminal local state
        // ---------------------------------------------------------
        db.Database.ExecuteSqlRaw(@"
            CREATE TABLE IF NOT EXISTS terminal_local_state (
                Id INTEGER NOT NULL
                    CONSTRAINT PK_terminal_local_state
                    PRIMARY KEY AUTOINCREMENT,

                TerminalId TEXT NOT NULL,
                TenantId TEXT NOT NULL,
                BranchId TEXT NOT NULL,
                TerminalCode TEXT NOT NULL,
                IsOnline INTEGER NOT NULL DEFAULT 0,
                DesiredState TEXT NOT NULL DEFAULT 'unlocked',
                LastSuccessfulHeartbeatUtc TEXT NULL,
                LastProcessedCommandVersion INTEGER NOT NULL DEFAULT 0,
                LastCommandResult TEXT NULL,
                UpdatedAtUtc TEXT NOT NULL
            );
        ");

        // ---------------------------------------------------------
        // Upgrade older terminal_local_state databases.
        // ---------------------------------------------------------
        AddTerminalLocalStateColumn(
            db,
            "LastProcessedCommandVersion",
            "INTEGER NOT NULL DEFAULT 0");

        AddTerminalLocalStateColumn(
            db,
            "LastCommandResult",
            "TEXT NULL");

        // ---------------------------------------------------------
        // Branch billing configuration cache
        // ---------------------------------------------------------
        db.Database.ExecuteSqlRaw(@"
            CREATE TABLE IF NOT EXISTS branch_billing_config (
                Id INTEGER NOT NULL
                    CONSTRAINT PK_branch_billing_config
                    PRIMARY KEY AUTOINCREMENT,

                BranchId TEXT NOT NULL,
                RatePerMinute TEXT NOT NULL,
                MinimumCharge TEXT NOT NULL,
                Currency TEXT NOT NULL,
                UpdatedAtUtc TEXT NOT NULL
            );
        ");

        db.Database.ExecuteSqlRaw(@"
            CREATE UNIQUE INDEX IF NOT EXISTS
            IX_branch_billing_config_BranchId
            ON branch_billing_config (BranchId);
        ");

        // ---------------------------------------------------------
        // Active session
        // ---------------------------------------------------------
        db.Database.ExecuteSqlRaw(@"
            CREATE TABLE IF NOT EXISTS active_session (
                Id INTEGER NOT NULL
                    CONSTRAINT PK_active_session
                    PRIMARY KEY AUTOINCREMENT,

                SessionId TEXT NOT NULL,
                CustomerId TEXT NOT NULL,
                TerminalId TEXT NOT NULL,
                BranchId TEXT NOT NULL,
                SessionType TEXT NOT NULL,
                State TEXT NOT NULL,
                StartedAtUtc TEXT NOT NULL,
                PausedAtUtc TEXT NULL,
                EndedAtUtc TEXT NULL,
                TotalPausedSeconds INTEGER NOT NULL DEFAULT 0,
                PrepaidAmount TEXT NULL,
                AllowedMinutes INTEGER NULL,
                RatePerMinute TEXT NULL,
                MinimumCharge TEXT NULL,
                ClientOperationId TEXT NOT NULL,
                UpdatedAtUtc TEXT NOT NULL
            );
        ");

        // ---------------------------------------------------------
        // Older databases may not have TotalPausedSeconds.
        // ---------------------------------------------------------
        try
        {
            db.Database.ExecuteSqlRaw(@"
                ALTER TABLE active_session
                ADD COLUMN TotalPausedSeconds
                INTEGER NOT NULL DEFAULT 0;
            ");
        }
        catch
        {
            // Column already exists.
        }

        db.Database.ExecuteSqlRaw(@"
            CREATE UNIQUE INDEX IF NOT EXISTS
            IX_active_session_SessionId
            ON active_session (SessionId);
        ");

        db.Database.ExecuteSqlRaw(@"
            CREATE INDEX IF NOT EXISTS
            IX_active_session_TerminalId
            ON active_session (TerminalId);
        ");

        db.Database.ExecuteSqlRaw(@"
            CREATE INDEX IF NOT EXISTS
            IX_active_session_State
            ON active_session (State);
        ");

        // ---------------------------------------------------------
        // Cached customers for offline lookup
        // ---------------------------------------------------------
        db.Database.ExecuteSqlRaw(@"
            CREATE TABLE IF NOT EXISTS terminal_customer (
                Id INTEGER NOT NULL
                    CONSTRAINT PK_terminal_customer
                    PRIMARY KEY AUTOINCREMENT,

                CustomerId TEXT NOT NULL,
                BranchId TEXT NOT NULL,
                CustomerType TEXT NOT NULL,
                FullName TEXT NOT NULL,
                Phone TEXT NULL,
                IdNumber TEXT NULL,
                ParentId TEXT NULL,
                ParentName TEXT NULL,
                UpdatedAtUtc TEXT NOT NULL
            );
        ");

        // ---------------------------------------------------------
        // Upgrade older terminal_customer databases.
        // ---------------------------------------------------------
        MigrateTerminalCustomerTable(db);

        // ---------------------------------------------------------
        // Customer indexes
        // ---------------------------------------------------------
        db.Database.ExecuteSqlRaw(@"
            CREATE INDEX IF NOT EXISTS
            IX_terminal_customer_BranchId_IdNumber
            ON terminal_customer (BranchId, IdNumber);
        ");

        db.Database.ExecuteSqlRaw(@"
            CREATE INDEX IF NOT EXISTS
            IX_terminal_customer_BranchId_FullName
            ON terminal_customer (BranchId, FullName);
        ");

        db.Database.ExecuteSqlRaw(@"
            CREATE UNIQUE INDEX IF NOT EXISTS
            IX_terminal_customer_CustomerId
            ON terminal_customer (CustomerId);
        ");

        // ---------------------------------------------------------
        // Offline operation queue
        // ---------------------------------------------------------
        db.Database.ExecuteSqlRaw(@"
            CREATE TABLE IF NOT EXISTS offline_queue (
                Id INTEGER NOT NULL
                    CONSTRAINT PK_offline_queue
                    PRIMARY KEY AUTOINCREMENT,

                OperationId TEXT NOT NULL,
                OperationType TEXT NOT NULL,
                Payload TEXT NOT NULL,
                CreatedAtUtc TEXT NOT NULL,
                AttemptCount INTEGER NOT NULL DEFAULT 0,
                LastAttemptAtUtc TEXT NULL,
                Status TEXT NOT NULL DEFAULT 'pending',
                LastError TEXT NULL
            );
        ");

        db.Database.ExecuteSqlRaw(@"
            CREATE UNIQUE INDEX IF NOT EXISTS
            IX_offline_queue_OperationId
            ON offline_queue (OperationId);
        ");
    }

    private static void AddTerminalLocalStateColumn(
        TerminalDbContext db,
        string columnName,
        string columnDefinition)
    {
        string sql =
            (columnName, columnDefinition) switch
            {
                (
                    "LastProcessedCommandVersion",
                    "INTEGER NOT NULL DEFAULT 0"
                ) =>
                    @"
                    ALTER TABLE terminal_local_state
                    ADD COLUMN LastProcessedCommandVersion
                    INTEGER NOT NULL DEFAULT 0;
                    ",

                (
                    "LastCommandResult",
                    "TEXT NULL"
                ) =>
                    @"
                    ALTER TABLE terminal_local_state
                    ADD COLUMN LastCommandResult
                    TEXT NULL;
                    ",

                _ =>
                    throw new ArgumentException(
                        "Unsupported terminal local-state column.",
                        nameof(columnName))
            };

        try
        {
            db.Database.ExecuteSqlRaw(sql);
        }
        catch
        {
            // Column already exists.
        }
    }

    private static void MigrateTerminalCustomerTable(
        TerminalDbContext db)
    {
        bool tableExists;

        using (var command =
            db.Database.GetDbConnection().CreateCommand())
        {
            command.CommandText = @"
                SELECT COUNT(*)
                FROM sqlite_master
                WHERE type = 'table'
                AND name = 'terminal_customer';
            ";

            db.Database.OpenConnection();

            tableExists =
                Convert.ToInt32(
                    command.ExecuteScalar()) > 0;

            db.Database.CloseConnection();
        }

        if (!tableExists)
        {
            return;
        }

        bool idNumberIsNotNull = false;
        bool parentIdExists = false;
        bool parentNameExists = false;

        using (var command =
            db.Database.GetDbConnection().CreateCommand())
        {
            command.CommandText =
                "PRAGMA table_info(terminal_customer);";

            db.Database.OpenConnection();

            using var reader =
                command.ExecuteReader();

            while (reader.Read())
            {
                string columnName =
                    reader.GetString(1);

                int notNull =
                    reader.GetInt32(3);

                if (string.Equals(
                        columnName,
                        "IdNumber",
                        StringComparison.OrdinalIgnoreCase))
                {
                    idNumberIsNotNull =
                        notNull == 1;
                }

                if (string.Equals(
                        columnName,
                        "ParentId",
                        StringComparison.OrdinalIgnoreCase))
                {
                    parentIdExists = true;
                }

                if (string.Equals(
                        columnName,
                        "ParentName",
                        StringComparison.OrdinalIgnoreCase))
                {
                    parentNameExists = true;
                }
            }

            db.Database.CloseConnection();
        }

        if (!idNumberIsNotNull &&
            parentIdExists &&
            parentNameExists)
        {
            return;
        }

        db.Database.ExecuteSqlRaw(@"
            CREATE TABLE IF NOT EXISTS terminal_customer_new (
                Id INTEGER NOT NULL
                    CONSTRAINT PK_terminal_customer_new
                    PRIMARY KEY AUTOINCREMENT,

                CustomerId TEXT NOT NULL,
                BranchId TEXT NOT NULL,
                CustomerType TEXT NOT NULL,
                FullName TEXT NOT NULL,
                Phone TEXT NULL,
                IdNumber TEXT NULL,
                ParentId TEXT NULL,
                ParentName TEXT NULL,
                UpdatedAtUtc TEXT NOT NULL
            );
        ");

        if (parentIdExists &&
            parentNameExists)
        {
            db.Database.ExecuteSqlRaw(@"
                INSERT INTO terminal_customer_new
                (
                    Id,
                    CustomerId,
                    BranchId,
                    CustomerType,
                    FullName,
                    Phone,
                    IdNumber,
                    ParentId,
                    ParentName,
                    UpdatedAtUtc
                )
                SELECT
                    Id,
                    CustomerId,
                    BranchId,
                    CustomerType,
                    FullName,
                    Phone,
                    IdNumber,
                    ParentId,
                    ParentName,
                    UpdatedAtUtc
                FROM terminal_customer;
            ");
        }
        else
        {
            db.Database.ExecuteSqlRaw(@"
                INSERT INTO terminal_customer_new
                (
                    Id,
                    CustomerId,
                    BranchId,
                    CustomerType,
                    FullName,
                    Phone,
                    IdNumber,
                    ParentId,
                    ParentName,
                    UpdatedAtUtc
                )
                SELECT
                    Id,
                    CustomerId,
                    BranchId,
                    CustomerType,
                    FullName,
                    Phone,
                    IdNumber,
                    NULL,
                    NULL,
                    UpdatedAtUtc
                FROM terminal_customer;
            ");
        }

        db.Database.ExecuteSqlRaw(@"
            DROP TABLE terminal_customer;
        ");

        db.Database.ExecuteSqlRaw(@"
            ALTER TABLE terminal_customer_new
            RENAME TO terminal_customer;
        ");
    }

    // =============================================================
    // TERMINAL STARTUP
    // =============================================================
    private async void StartTerminal()
    {
        _registration =
            TerminalStorage.Load();

        if (_registration == null)
        {
            ShowRegistrationScreen();
            return;
        }

        _offlineQueueSyncService =
            new OfflineQueueSyncService(
                _offlineQueueService,
                _terminalSessionApi,
                _registration);

        InitializeTerminalPaymentApi();

        // ---------------------------------------------------------
        // IMPORTANT:
        // A registered terminal always starts LOCKED.
        // ---------------------------------------------------------
        ShowLockedScreen();

        await AuthenticateTerminal(
            _registration);

        // ---------------------------------------------------------
        // If Windows/application restarted while a session was
        // active, recover it.
        // Otherwise the terminal remains locked.
        // ---------------------------------------------------------
        await RecoverActiveSession();

        if (!HasActiveLocalSession())
        {
            ShowLockedScreen();
        }
    }

    private bool HasActiveLocalSession()
    {
        using var db =
            new TerminalDbContext();

        if (_registration == null)
        {
            return false;
        }

        return db.ActiveSessions.Any(
            x =>
                x.TerminalId ==
                    _registration.TerminalId &&
                (x.State == "running" ||
                 x.State == "paused"));
    }

    private void InitializeTerminalPaymentApi()
    {
        if (_registration == null ||
            string.IsNullOrWhiteSpace(
                _registration.Credential))
        {
            _terminalPaymentApi = null;
            return;
        }

        _terminalPaymentApi =
            new TerminalPaymentApi(
                _httpClient,
                _registration.Credential);
    }

    private void ShowRegistrationScreen()
    {
        ScreenHost.Content =
            new TerminalRegistrationView(
                RegisterTerminalAsync);
    }

    private void ShowHomeScreen()
    {
        ScreenHost.Content =
            new TerminalHomeView(
                ShowCustomerLookupScreen);
    }

    private void ShowCustomerLookupScreen()
    {
        ScreenHost.Content =
            new CustomerSelectionView(
                LookupCustomerAsync,
                ShowLockedScreen);
    }

    private void ShowStartSessionScreen(
        TerminalCustomer customer)
    {
        ScreenHost.Content =
            new StartSessionView(
                ShowCustomerLookupScreen,
                StartSessionAsync,
                customer);
    }

    // =============================================================
    // RETURN TO LOCKED STATE
    // =============================================================
    public void ReturnToStartScreen()
    {
        ShowLockedScreen();
    }

    // =============================================================
    // BILLING
    // =============================================================
    private async Task RefreshBillingConfigAsync()
    {
        if (_registration == null)
        {
            return;
        }

        try
        {
            TerminalBillingConfig? config =
                await _terminalSessionApi.GetBillingConfigAsync(
                    _registration);

            if (config == null ||
                string.IsNullOrWhiteSpace(
                    config.BranchId))
            {
                return;
            }

            using var db =
                new TerminalDbContext();

            BranchBillingConfigLocal? existing =
                db.BranchBillingConfigs
                    .FirstOrDefault(
                        x =>
                            x.BranchId ==
                            config.BranchId);

            if (existing == null)
            {
                existing =
                    new BranchBillingConfigLocal
                    {
                        BranchId =
                            config.BranchId
                    };

                db.BranchBillingConfigs.Add(
                    existing);
            }

            existing.RatePerMinute =
                config.RatePerMinute;

            existing.MinimumCharge =
                config.MinimumCharge;

            existing.Currency =
                string.IsNullOrWhiteSpace(
                    config.Currency)
                    ? "KES"
                    : config.Currency;

            existing.UpdatedAtUtc =
                DateTime.UtcNow;

            db.SaveChanges();
        }
        catch
        {
            // Backend unavailable.
        }
    }

    // =============================================================
    // TERMINAL REGISTRATION
    // =============================================================
    private async Task RegisterTerminalAsync(
        string licenceKey)
    {
        try
        {
            var request =
                new RegisterTerminalRequest
                {
                    LicenceKey =
                        licenceKey,

                    MachineName =
                        Environment.MachineName,

                    DeviceIdentifier =
                        Environment.MachineName
                };

            HttpResponseMessage response =
                await _httpClient.PostAsJsonAsync(
                    "terminals/register",
                    request);

            string responseBody =
                await response.Content.ReadAsStringAsync();

            if (!response.IsSuccessStatusCode)
            {
                MessageBox.Show(
                    responseBody,
                    "Registration Failed",
                    MessageBoxButton.OK,
                    MessageBoxImage.Error);

                return;
            }

            RegisterTerminalApiResponse? result =
                JsonSerializer.Deserialize<
                    RegisterTerminalApiResponse>(
                    responseBody,
                    JsonOptions);

            if (result?.Data == null)
            {
                MessageBox.Show(
                    "The server returned an invalid registration response.",
                    "Registration Failed",
                    MessageBoxButton.OK,
                    MessageBoxImage.Error);

                return;
            }

            if (string.IsNullOrWhiteSpace(
                    result.Data.Credential))
            {
                MessageBox.Show(
                    "The server did not provide a terminal credential.",
                    "Registration Failed",
                    MessageBoxButton.OK,
                    MessageBoxImage.Error);

                return;
            }

            TerminalStorage.Save(
                new TerminalRegistration
                {
                    TerminalId =
                        result.Data.TerminalId,

                    TenantId =
                        result.Data.TenantId,

                    BranchId =
                        result.Data.BranchId,

                    TerminalCode =
                        result.Data.TerminalCode,

                    Credential =
                        result.Data.Credential,

                    MachineName =
                        request.MachineName,

                    DeviceIdentifier =
                        request.DeviceIdentifier
                });

            _registration =
                TerminalStorage.Load();

            if (_registration == null)
            {
                MessageBox.Show(
                    "Registration succeeded, but the local terminal record could not be loaded.",
                    "Registration Error",
                    MessageBoxButton.OK,
                    MessageBoxImage.Error);

                return;
            }

            _offlineQueueSyncService =
                new OfflineQueueSyncService(
                    _offlineQueueService,
                    _terminalSessionApi,
                    _registration);

            InitializeTerminalPaymentApi();

            // Newly registered terminals also start LOCKED.
            ShowLockedScreen();

            await AuthenticateTerminal(
                _registration);

            await RecoverActiveSession();

            if (!HasActiveLocalSession())
            {
                ShowLockedScreen();
            }
        }
        catch (TaskCanceledException)
        {
            MessageBox.Show(
                "Registration timed out. Check that the CyberSaaS backend is running.",
                "Connection Error",
                MessageBoxButton.OK,
                MessageBoxImage.Error);
        }
        catch (HttpRequestException ex)
        {
            MessageBox.Show(
                ex.Message,
                "Connection Error",
                MessageBoxButton.OK,
                MessageBoxImage.Error);
        }
        catch (Exception ex)
        {
            MessageBox.Show(
                ex.Message,
                "Registration Error",
                MessageBoxButton.OK,
                MessageBoxImage.Error);
        }
    }

    // =============================================================
    // TERMINAL AUTHENTICATION
    // =============================================================
    private async Task AuthenticateTerminal(
        TerminalRegistration registration)
    {
        if (string.IsNullOrWhiteSpace(
                registration.Credential))
        {
            TerminalStorage.Delete();
            _registration = null;
            _heartbeatTimer.Stop();
            ShowRegistrationScreen();
            return;
        }

        InitializeTerminalPaymentApi();

        try
        {
            var request =
                new TerminalAuthRequest
                {
                    Credential =
                        registration.Credential
                };

            HttpResponseMessage response =
                await _httpClient.PostAsJsonAsync(
                    "terminals/auth",
                    request);

            if (response.StatusCode ==
                System.Net.HttpStatusCode.Unauthorized)
            {
                _heartbeatTimer.Stop();

                TerminalStorage.Delete();

                _registration = null;

                ShowRegistrationScreen();

                return;
            }

            if (!response.IsSuccessStatusCode)
            {
                _heartbeatTimer.Stop();
                return;
            }

            string responseBody =
                await response.Content.ReadAsStringAsync();

            TerminalAuthApiResponse? result =
                JsonSerializer.Deserialize<
                    TerminalAuthApiResponse>(
                    responseBody,
                    JsonOptions);

            if (result?.Data == null)
            {
                _heartbeatTimer.Stop();
                return;
            }

            if (!string.Equals(
                    result.Data.TerminalId,
                    registration.TerminalId,
                    StringComparison.OrdinalIgnoreCase))
            {
                _heartbeatTimer.Stop();
                return;
            }

            await SendHeartbeat();

            await ReportTerminalHealth();

            await RefreshBillingConfigAsync();

            _heartbeatTimer.Start();
        }
        catch
        {
            // Keep the local registration during temporary
            // network/backend failures.
            _heartbeatTimer.Stop();
        }
    }

    private async Task ReportTerminalHealth()
    {
        if (_registration == null)
        {
            return;
        }

        try
        {
            await _healthReportingService.CheckAndReportAsync(
                _registration,
                ApiBaseUrl);
        }
        catch
        {
            // Health reporting must never interrupt
            // terminal operation.
        }
    }

    // =============================================================
    // CUSTOMER LOOKUP
    // =============================================================
    private async Task LookupCustomerAsync(
        string searchType,
        string searchTerm)
    {
        if (_registration == null)
        {
            throw new InvalidOperationException(
                "Terminal is not registered.");
        }

        string normalizedSearchType =
            searchType.Trim().ToLowerInvariant();

        string normalizedSearchTerm =
            searchTerm.Trim();

        if (normalizedSearchType != "adult" &&
            normalizedSearchType != "child")
        {
            throw new InvalidOperationException(
                "Invalid customer search type.");
        }

        if (string.IsNullOrWhiteSpace(
                normalizedSearchTerm))
        {
            throw new InvalidOperationException(
                normalizedSearchType == "child"
                    ? "Child name is required."
                    : "ID number is required.");
        }

        try
        {
            TerminalCustomerLookupResponse response =
                await _terminalCustomerApi.LookupCustomerAsync(
                    _registration,
                    normalizedSearchType,
                    normalizedSearchTerm);

            if (response.Customer == null ||
                string.IsNullOrWhiteSpace(
                    response.Customer.Id))
            {
                throw new InvalidOperationException(
                    "Customer was not found.");
            }

            using (var db =
                new TerminalDbContext())
            {
                TerminalCustomerLocal? existing =
                    db.TerminalCustomers.FirstOrDefault(
                        x =>
                            x.CustomerId ==
                                response.Customer!.Id &&
                            x.BranchId ==
                                _registration.BranchId);

                if (existing == null)
                {
                    existing =
                        new TerminalCustomerLocal();

                    db.TerminalCustomers.Add(
                        existing);
                }

                existing.CustomerId =
                    response.Customer.Id;

                existing.BranchId =
                    _registration.BranchId;

                existing.CustomerType =
                    response.Customer.CustomerType;

                existing.FullName =
                    response.Customer.FullName;

                existing.Phone =
                    response.Customer.Phone;

                existing.IdNumber =
                    normalizedSearchType == "adult"
                        ? normalizedSearchTerm
                        : null;

                existing.ParentId =
                    response.Customer.ParentId;

                existing.ParentName =
                    response.Customer.ParentName;

                existing.UpdatedAtUtc =
                    response.Customer.UpdatedAt;

                db.SaveChanges();
            }

            ShowStartSessionScreen(
                response.Customer);
        }
        catch (HttpRequestException)
        {
            using var db =
                new TerminalDbContext();

            TerminalCustomerLocal? localCustomer;

            if (normalizedSearchType == "child")
            {
                localCustomer =
                    db.TerminalCustomers.FirstOrDefault(
                        x =>
                            x.BranchId ==
                                _registration.BranchId &&
                            x.CustomerType.ToLower() ==
                                "child" &&
                            x.FullName.ToLower() ==
                                normalizedSearchTerm.ToLower());
            }
            else
            {
                localCustomer =
                    db.TerminalCustomers.FirstOrDefault(
                        x =>
                            x.BranchId ==
                                _registration.BranchId &&
                            x.CustomerType.ToLower() ==
                                "adult" &&
                            x.IdNumber ==
                                normalizedSearchTerm);
            }

            if (localCustomer == null)
            {
                throw new InvalidOperationException(
                    normalizedSearchType == "child"
                        ? "The backend is unavailable and this child has not been cached on this terminal."
                        : "The backend is unavailable and this customer has not been cached on this terminal.");
            }

            var customer =
                new TerminalCustomer
                {
                    Id =
                        localCustomer.CustomerId,

                    BranchId =
                        localCustomer.BranchId,

                    CustomerType =
                        localCustomer.CustomerType,

                    FullName =
                        localCustomer.FullName,

                    Phone =
                        localCustomer.Phone,

                    ParentId =
                        localCustomer.ParentId,

                    ParentName =
                        localCustomer.ParentName,

                    UpdatedAt =
                        localCustomer.UpdatedAtUtc
                };

            ShowStartSessionScreen(
                customer);
        }
    }

    // =============================================================
    // OFFLINE PREPAID CALCULATION
    // =============================================================
    private static int? CalculateOfflinePrepaidMinutes(
        string? prepaidAmount,
        string ratePerMinute)
    {
        if (string.IsNullOrWhiteSpace(
                prepaidAmount))
        {
            return null;
        }

        if (!decimal.TryParse(
                prepaidAmount,
                System.Globalization.NumberStyles.Number,
                System.Globalization.CultureInfo.InvariantCulture,
                out decimal amount))
        {
            return null;
        }

        if (!decimal.TryParse(
                ratePerMinute,
                System.Globalization.NumberStyles.Number,
                System.Globalization.CultureInfo.InvariantCulture,
                out decimal rate))
        {
            return null;
        }

        if (amount <= 0 ||
            rate <= 0)
        {
            return null;
        }

        decimal minutes =
            Math.Floor(
                amount / rate);

        if (minutes <= 0)
        {
            return null;
        }

        return checked(
            (int)minutes);
    }

    // =============================================================
    // START SESSION
    // =============================================================
    private async Task StartSessionAsync(
        TerminalStartSessionRequest request)
    {
        if (_registration == null)
        {
            throw new InvalidOperationException(
                "Terminal is not registered.");
        }

        using (var db =
            new TerminalDbContext())
        {
            bool activeSessionExists =
                db.ActiveSessions.Any(
                    x =>
                        x.TerminalId ==
                            _registration.TerminalId &&
                        (x.State == "running" ||
                         x.State == "paused"));

            if (activeSessionExists)
            {
                MessageBox.Show(
                    "This terminal already has an active session.\n\n" +
                    "End or complete the current session before starting another one.",
                    "Session Already Active",
                    MessageBoxButton.OK,
                    MessageBoxImage.Warning);

                return;
            }
        }

        try
        {
            TerminalSessionData? result =
                await _terminalSessionApi.StartSessionAsync(
                    _registration,
                    request);

            if (result == null ||
                string.IsNullOrWhiteSpace(
                    result.Id))
            {
                throw new InvalidOperationException(
                    "The server returned an invalid session response.");
            }

            SaveActiveSession(result);

            OpenSessionBar(result);
        }
        catch (HttpRequestException)
        {
            BranchBillingConfigLocal? billingConfig =
                GetLocalBillingConfig();

            if (billingConfig == null)
            {
                MessageBox.Show(
                    "This terminal is offline and does not have cached billing configuration.\n\n" +
                    "Connect to the server before starting a new session.",
                    "Billing Configuration Unavailable",
                    MessageBoxButton.OK,
                    MessageBoxImage.Warning);

                return;
            }

            if (string.IsNullOrWhiteSpace(
                    billingConfig.RatePerMinute))
            {
                MessageBox.Show(
                    "No valid branch billing rate is available offline.",
                    "Billing Configuration Error",
                    MessageBoxButton.OK,
                    MessageBoxImage.Warning);

                return;
            }

            string localSessionId =
                $"offline-{Guid.NewGuid()}";

            DateTime startedAt =
                request.StartedAt ??
                DateTime.UtcNow;

            int? allowedMinutes = null;

            if (request.SessionType.Equals(
                    "prepaid",
                    StringComparison.OrdinalIgnoreCase))
            {
                allowedMinutes =
                    CalculateOfflinePrepaidMinutes(
                        request.PrepaidAmount,
                        billingConfig.RatePerMinute);

                if (allowedMinutes == null)
                {
                    MessageBox.Show(
                        "The prepaid amount is not enough to start an offline session using the cached branch rate.",
                        "Invalid Prepaid Amount",
                        MessageBoxButton.OK,
                        MessageBoxImage.Warning);

                    return;
                }
            }

            var localSession =
                new ActiveSessionLocal
                {
                    SessionId =
                        localSessionId,

                    CustomerId =
                        request.CustomerId,

                    TerminalId =
                        _registration.TerminalId,

                    BranchId =
                        _registration.BranchId,

                    SessionType =
                        request.SessionType,

                    State =
                        "running",

                    StartedAtUtc =
                        startedAt.ToUniversalTime(),

                    PausedAtUtc =
                        null,

                    EndedAtUtc =
                        null,

                    TotalPausedSeconds =
                        0,

                    PrepaidAmount =
                        request.PrepaidAmount,

                    AllowedMinutes =
                        allowedMinutes,

                    RatePerMinute =
                        billingConfig.RatePerMinute,

                    MinimumCharge =
                        billingConfig.MinimumCharge,

                    ClientOperationId =
                        request.ClientOperationId,

                    UpdatedAtUtc =
                        DateTime.UtcNow
                };

            using (var db =
                new TerminalDbContext())
            {
                db.ActiveSessions.Add(
                    localSession);

                db.SaveChanges();
            }

            _offlineQueueService.Enqueue(
                request.ClientOperationId,
                "start_session",
                request);

            var localSessionData =
                new TerminalSessionData
                {
                    Id =
                        localSessionId,

                    TenantId =
                        _registration.TenantId,

                    BranchId =
                        _registration.BranchId,

                    TerminalId =
                        _registration.TerminalId,

                    CustomerId =
                        request.CustomerId,

                    SessionType =
                        request.SessionType,

                    Status =
                        "running",

                    StartedAt =
                        startedAt,

                    PausedAt =
                        null,

                    EndedAt =
                        null,

                    TotalPausedSeconds =
                        0,

                    PrepaidAmount =
                        request.PrepaidAmount,

                    AllowedMinutes =
                        allowedMinutes,

                    RatePerMinute =
                        billingConfig.RatePerMinute,

                    MinimumCharge =
                        billingConfig.MinimumCharge,

                    ClientOperationId =
                        request.ClientOperationId,

                    CreatedAt =
                        startedAt,

                    UpdatedAt =
                        DateTime.UtcNow
                };

            OpenSessionBar(
                localSessionData);

            MessageBox.Show(
                "The backend is unavailable.\n\n" +
                "The session has started locally and will be synchronized when the connection returns.",
                "Offline Session",
                MessageBoxButton.OK,
                MessageBoxImage.Information);
        }
    }

    private BranchBillingConfigLocal? GetLocalBillingConfig()
    {
        if (_registration == null)
        {
            return null;
        }

        using var db =
            new TerminalDbContext();

        return db.BranchBillingConfigs
            .FirstOrDefault(
                x =>
                    x.BranchId ==
                    _registration.BranchId);
    }

    // =============================================================
    // SAVE ACTIVE SESSION
    // =============================================================
    private void SaveActiveSession(
        TerminalSessionData session)
    {
        if (_registration == null)
        {
            return;
        }

        using var db =
            new TerminalDbContext();

        ActiveSessionLocal? existing =
            db.ActiveSessions
                .FirstOrDefault(
                    x =>
                        x.SessionId ==
                        session.Id);

        if (existing == null)
        {
            existing =
                new ActiveSessionLocal
                {
                    SessionId =
                        session.Id
                };

            db.ActiveSessions.Add(
                existing);
        }

        existing.CustomerId =
            session.CustomerId;

        existing.TerminalId =
            session.TerminalId;

        existing.BranchId =
            session.BranchId;

        existing.SessionType =
            session.SessionType;

        existing.State =
            string.Equals(
                session.Status,
                "paused",
                StringComparison.OrdinalIgnoreCase)
                ? "paused"
                : "running";

        existing.StartedAtUtc =
            session.StartedAt.ToUniversalTime();

        existing.PausedAtUtc =
            session.PausedAt?.ToUniversalTime();

        existing.EndedAtUtc =
            session.EndedAt?.ToUniversalTime();

        existing.TotalPausedSeconds =
            session.TotalPausedSeconds;

        existing.PrepaidAmount =
            session.PrepaidAmount;

        existing.AllowedMinutes =
            session.AllowedMinutes;

        existing.RatePerMinute =
            session.RatePerMinute;

        existing.MinimumCharge =
            session.MinimumCharge;

        existing.ClientOperationId =
            session.ClientOperationId;

        existing.UpdatedAtUtc =
            DateTime.UtcNow;

        db.SaveChanges();
    }

    // =============================================================
    // RECOVER ACTIVE SESSION
    // =============================================================
    private async Task RecoverActiveSession()
    {
        if (_registration == null)
        {
            return;
        }

        try
        {
            TerminalSessionData? serverSession =
                await _terminalSessionApi.GetActiveSessionAsync(
                    _registration);

            if (serverSession != null &&
                !string.IsNullOrWhiteSpace(
                    serverSession.Id))
            {
                SaveActiveSession(
                    serverSession);

                OpenSessionBar(
                    serverSession);

                return;
            }
        }
        catch
        {
            // Backend unavailable.
        }

        try
        {
            using var db =
                new TerminalDbContext();

            ActiveSessionLocal? localSession =
                db.ActiveSessions
                    .FirstOrDefault(
                        x =>
                            x.TerminalId ==
                                _registration.TerminalId &&
                            (x.State == "running" ||
                             x.State == "paused"));

            if (localSession == null)
            {
                return;
            }

            try
            {
                TerminalSessionData? serverSession =
                    await _terminalSessionApi.GetSessionAsync(
                        _registration,
                        localSession.SessionId);

                if (serverSession == null)
                {
                    return;
                }

                if (string.Equals(
                        serverSession.Status,
                        "completed",
                        StringComparison.OrdinalIgnoreCase) ||
                    string.Equals(
                        serverSession.Status,
                        "cancelled",
                        StringComparison.OrdinalIgnoreCase))
                {
                    db.ActiveSessions.Remove(
                        localSession);

                    await db.SaveChangesAsync();

                    return;
                }

                SaveActiveSession(
                    serverSession);

                OpenSessionBar(
                    serverSession);

                return;
            }
            catch
            {
                // Backend unavailable.
            }

            var offlineSession =
                new TerminalSessionData
                {
                    Id =
                        localSession.SessionId,

                    TenantId =
                        _registration.TenantId,

                    BranchId =
                        localSession.BranchId,

                    TerminalId =
                        localSession.TerminalId,

                    CustomerId =
                        localSession.CustomerId,

                    SessionType =
                        localSession.SessionType,

                    Status =
                        localSession.State == "paused"
                            ? "paused"
                            : "active",

                    StartedAt =
                        localSession.StartedAtUtc,

                    PausedAt =
                        localSession.PausedAtUtc,

                    EndedAt =
                        localSession.EndedAtUtc,

                    TotalPausedSeconds =
                        localSession.TotalPausedSeconds,

                    PrepaidAmount =
                        localSession.PrepaidAmount,

                    AllowedMinutes =
                        localSession.AllowedMinutes,

                    RatePerMinute =
                        localSession.RatePerMinute,

                    MinimumCharge =
                        localSession.MinimumCharge,

                    ClientOperationId =
                        localSession.ClientOperationId,

                    CreatedAt =
                        localSession.StartedAtUtc,

                    UpdatedAt =
                        localSession.UpdatedAtUtc
                };

            OpenSessionBar(
                offlineSession);
        }
        catch
        {
            // No recoverable local session.
        }
    }

    // =============================================================
    // SESSION BAR
    // =============================================================
    private void OpenSessionBar(
        TerminalSessionData session)
    {
        if (_registration == null)
        {
            return;
        }

        if (_terminalPaymentApi == null)
        {
            InitializeTerminalPaymentApi();
        }

        if (_terminalPaymentApi == null)
        {
            MessageBox.Show(
                "Terminal payment service is not available.",
                "Payment Service",
                MessageBoxButton.OK,
                MessageBoxImage.Error);

            return;
        }

        var sessionBar =
            new SessionBarWindow(
                _registration,
                _terminalSessionApi,
                _terminalPaymentApi,
                session);

        sessionBar.Show();

        Hide();
    }

    // =============================================================
    // HEARTBEAT TIMER
    // =============================================================
    private async void HeartbeatTimer_Tick(
        object? sender,
        EventArgs e)
    {
        await SendHeartbeat();

        await ReportTerminalHealth();

        await RefreshBillingConfigAsync();

        if (_offlineQueueSyncService != null)
        {
            await _offlineQueueSyncService.SyncPendingAsync();
        }
    }

    // =============================================================
    // LOCK SCREEN
    // =============================================================
    public void ShowLockedScreen()
    {
        ScreenHost.Content =
            new TerminalLockedView(
                ShowCustomerLookupScreen);

        Show();

        if (!IsTestBuild)
        {
            WindowState =
                WindowState.Maximized;
        }

        Activate();
        Focus();
    }

    // =============================================================
    // HEARTBEAT
    // =============================================================
    private async Task SendHeartbeat()
    {
        if (_registration == null)
        {
            return;
        }

        if (_heartbeatRunning)
        {
            return;
        }

        _heartbeatRunning = true;

        try
        {
            var request =
                new TerminalHeartbeatRequest
                {
                    MachineName =
                        Environment.MachineName,

                    DeviceIdentifier =
                        Environment.MachineName
                };

            using HttpRequestMessage httpRequest =
                new HttpRequestMessage(
                    HttpMethod.Post,
                    "terminals/heartbeat");

            httpRequest.Headers.Authorization =
                new System.Net.Http.Headers
                    .AuthenticationHeaderValue(
                        "Bearer",
                        _registration.Credential);

            httpRequest.Content =
                JsonContent.Create(request);

            HttpResponseMessage response =
                await _httpClient.SendAsync(
                    httpRequest);

            if (!response.IsSuccessStatusCode)
            {
                return;
            }

            string responseBody =
                await response.Content.ReadAsStringAsync();

            TerminalHeartbeatApiResponse? result =
                JsonSerializer.Deserialize<
                    TerminalHeartbeatApiResponse>(
                        responseBody,
                        JsonOptions);

            if (result?.Data == null)
            {
                return;
            }

            SaveLocalTerminalState(
                result.Data);
        }
        catch
        {
            // Offline is expected.
        }
        finally
        {
            _heartbeatRunning = false;
        }
    }

    // =============================================================
    // SAVE HEARTBEAT STATE + EXECUTE COMMAND
    // =============================================================
    private void SaveLocalTerminalState(
        TerminalHeartbeatData heartbeat)
    {
        if (_registration == null)
        {
            return;
        }

        string desiredState =
            NormalizeTerminalCommand(
                heartbeat.DesiredState);

        using var db =
            new TerminalDbContext();

        TerminalLocalState? state =
            db.TerminalLocalStates
                .FirstOrDefault(
                    x =>
                        x.TerminalId ==
                        _registration.TerminalId);

        // ---------------------------------------------------------
        // First heartbeat for this terminal.
        //
        // The terminal already starts locked by design. Therefore
        // we record the server's current command version without
        // executing an old command immediately.
        // Future versions are treated as new commands.
        // ---------------------------------------------------------
        if (state == null)
        {
            state =
                new TerminalLocalState
                {
                    TerminalId =
                        _registration.TerminalId,

                    TenantId =
                        _registration.TenantId,

                    BranchId =
                        _registration.BranchId,

                    TerminalCode =
                        _registration.TerminalCode
                };

            db.TerminalLocalStates.Add(
                state);

            state.IsOnline =
                true;

            state.DesiredState =
                desiredState;

            state.LastSuccessfulHeartbeatUtc =
                DateTime.UtcNow;

            state.UpdatedAtUtc =
                DateTime.UtcNow;

            db.SaveChanges();

            SetLastProcessedCommandVersion(
                db,
                _registration.TerminalId,
                heartbeat.CommandVersion,
                "initial_state_recorded");

            return;
        }

        long lastProcessedCommandVersion =
            GetLastProcessedCommandVersion(
                db,
                _registration.TerminalId);

        state.IsOnline =
            true;

        state.DesiredState =
            desiredState;

        state.LastSuccessfulHeartbeatUtc =
            DateTime.UtcNow;

        state.UpdatedAtUtc =
            DateTime.UtcNow;

        db.SaveChanges();

        // ---------------------------------------------------------
        // No new command.
        // ---------------------------------------------------------
        if (heartbeat.CommandVersion <=
            lastProcessedCommandVersion)
        {
            return;
        }

        // ---------------------------------------------------------
        // A running/paused session must not be interrupted by the
        // application lock/unlock command.
        //
        // We leave the command version unprocessed so the same
        // command can be applied after the session finishes.
        // ---------------------------------------------------------
        if (HasActiveLocalSession() &&
            IsApplicationLockCommand(desiredState))
        {
            return;
        }

        ExecuteTerminalCommand(
            desiredState,
            heartbeat.CommandVersion);
    }

    private long GetLastProcessedCommandVersion(
        TerminalDbContext db,
        string terminalId)
    {
        using var command =
            db.Database.GetDbConnection().CreateCommand();

        command.CommandText = @"
            SELECT
                COALESCE(
                    LastProcessedCommandVersion,
                    0)
            FROM terminal_local_state
            WHERE TerminalId = $terminalId
            LIMIT 1;
        ";

        var parameter =
            command.CreateParameter();

        parameter.ParameterName =
            "$terminalId";

        parameter.Value =
            terminalId;

        command.Parameters.Add(
            parameter);

        db.Database.OpenConnection();

        try
        {
            object? result =
                command.ExecuteScalar();

            if (result == null ||
                result == DBNull.Value)
            {
                return 0;
            }

            return Convert.ToInt64(result);
        }
        finally
        {
            db.Database.CloseConnection();
        }
    }

    private void SetLastProcessedCommandVersion(
        TerminalDbContext db,
        string terminalId,
        long commandVersion,
        string result)
    {
        using var command =
            db.Database.GetDbConnection().CreateCommand();

        command.CommandText = @"
            UPDATE terminal_local_state
            SET
                LastProcessedCommandVersion =
                    $commandVersion,
                LastCommandResult =
                    $result,
                UpdatedAtUtc =
                    $updatedAtUtc
            WHERE TerminalId =
                $terminalId;
        ";

        var terminalParameter =
            command.CreateParameter();

        terminalParameter.ParameterName =
            "$terminalId";

        terminalParameter.Value =
            terminalId;

        command.Parameters.Add(
            terminalParameter);

        var versionParameter =
            command.CreateParameter();

        versionParameter.ParameterName =
            "$commandVersion";

        versionParameter.Value =
            commandVersion;

        command.Parameters.Add(
            versionParameter);

        var resultParameter =
            command.CreateParameter();

        resultParameter.ParameterName =
            "$result";

        resultParameter.Value =
            result;

        command.Parameters.Add(
            resultParameter);

        var updatedParameter =
            command.CreateParameter();

        updatedParameter.ParameterName =
            "$updatedAtUtc";

        updatedParameter.Value =
            DateTime.UtcNow.ToString("O");

        command.Parameters.Add(
            updatedParameter);

        db.Database.OpenConnection();

        try
        {
            command.ExecuteNonQuery();
        }
        finally
        {
            db.Database.CloseConnection();
        }
    }

    private static string NormalizeTerminalCommand(
        string? desiredState)
    {
        if (string.IsNullOrWhiteSpace(
                desiredState))
        {
            return "UNLOCK";
        }

        return desiredState
            .Trim()
            .Replace("-", "_")
            .Replace(" ", "_")
            .ToUpperInvariant();
    }

    private static bool IsApplicationLockCommand(
        string command)
    {
        return command == "LOCK" ||
               command == "UNLOCK";
    }

    // =============================================================
    // TERMINAL COMMAND EXECUTION
    // =============================================================
    private void ExecuteTerminalCommand(
        string command,
        long commandVersion)
    {
        if (_registration == null)
        {
            return;
        }

        command =
            NormalizeTerminalCommand(command);

        try
        {
            // -----------------------------------------------------
            // RESTART / SHUTDOWN / CANCEL_SHUTDOWN
            //
            // Persist the command version BEFORE executing the
            // Windows command because restart/shutdown may terminate
            // the application immediately.
            // -----------------------------------------------------
            if (command == "RESTART" ||
                command == "SHUTDOWN" ||
                command == "CANCEL_SHUTDOWN")
            {
                using (var db =
                    new TerminalDbContext())
                {
                    SetLastProcessedCommandVersion(
                        db,
                        _registration.TerminalId,
                        commandVersion,
                        $"{command}_processing");
                }

                bool success =
                    ExecuteWindowsShutdownCommand(
                        command);

                if (!success)
                {
                    using var db =
                        new TerminalDbContext();

                    SetLastProcessedCommandVersion(
                        db,
                        _registration.TerminalId,
                        commandVersion,
                        $"{command}_failed");
                }

                return;
            }

            // -----------------------------------------------------
            // LOCK
            // -----------------------------------------------------
            if (command == "LOCK")
            {
                ShowLockedScreen();

                using var db =
                    new TerminalDbContext();

                SetLastProcessedCommandVersion(
                    db,
                    _registration.TerminalId,
                    commandVersion,
                    "LOCK_executed");

                return;
            }

            // -----------------------------------------------------
            // UNLOCK
            // -----------------------------------------------------
            if (command == "UNLOCK")
            {
                ShowHomeScreen();

                using var db =
                    new TerminalDbContext();

                SetLastProcessedCommandVersion(
                    db,
                    _registration.TerminalId,
                    commandVersion,
                    "UNLOCK_executed");

                return;
            }

            // -----------------------------------------------------
            // Unknown command.
            // -----------------------------------------------------
            using (var db =
                new TerminalDbContext())
            {
                SetLastProcessedCommandVersion(
                    db,
                    _registration.TerminalId,
                    commandVersion,
                    $"unknown_command:{command}");
            }
        }
        catch (Exception ex)
        {
            try
            {
                using var db =
                    new TerminalDbContext();

                SetLastProcessedCommandVersion(
                    db,
                    _registration.TerminalId,
                    commandVersion,
                    $"{command}_error:{ex.Message}");
            }
            catch
            {
                // Never allow command handling to terminate
                // the terminal heartbeat loop.
            }
        }
    }

    private static bool ExecuteWindowsShutdownCommand(
        string command)
    {
        if (IsTestBuild)
        {
            return true;
        }

        string arguments;

        switch (command)
        {
            case "RESTART":
                arguments =
                    "/r /t 0";
                break;

            case "SHUTDOWN":
                arguments =
                    "/s /t 0";
                break;

            case "CANCEL_SHUTDOWN":
                arguments =
                    "/a";
                break;

            default:
                return false;
        }

        try
        {
            var startInfo =
                new ProcessStartInfo
                {
                    FileName =
                        "shutdown.exe",

                    Arguments =
                        arguments,

                    UseShellExecute =
                        false,

                    CreateNoWindow =
                        true,

                    WindowStyle =
                        ProcessWindowStyle.Hidden
                };

            using Process? process =
                Process.Start(startInfo);

            if (process == null)
            {
                return false;
            }

            if (command == "CANCEL_SHUTDOWN")
            {
                process.WaitForExit(3000);

                return process.ExitCode == 0;
            }

            return true;
        }
        catch
        {
            return false;
        }
    }

    // =============================================================
    // FORM CLOSE
    // =============================================================
    protected override void OnClosed(
        EventArgs e)
    {
        _heartbeatTimer.Stop();

        _httpClient.Dispose();

        base.OnClosed(e);
    }

    private static readonly JsonSerializerOptions JsonOptions =
        new JsonSerializerOptions
        {
            PropertyNameCaseInsensitive = true
        };
}

// =============================================================
// API DTOs
// =============================================================

public sealed class RegisterTerminalRequest
{
    [JsonPropertyName("licence_key")]
    public string LicenceKey { get; set; } =
        string.Empty;

    [JsonPropertyName("machine_name")]
    public string MachineName { get; set; } =
        string.Empty;

    [JsonPropertyName("device_identifier")]
    public string DeviceIdentifier { get; set; } =
        string.Empty;
}

public sealed class RegisterTerminalApiResponse
{
    [JsonPropertyName("message")]
    public string? Message { get; set; }

    [JsonPropertyName("data")]
    public RegisterTerminalData? Data { get; set; }
}

public sealed class RegisterTerminalData
{
    [JsonPropertyName("terminal_id")]
    public string TerminalId { get; set; } =
        string.Empty;

    [JsonPropertyName("tenant_id")]
    public string TenantId { get; set; } =
        string.Empty;

    [JsonPropertyName("branch_id")]
    public string BranchId { get; set; } =
        string.Empty;

    [JsonPropertyName("terminal_code")]
    public string TerminalCode { get; set; } =
        string.Empty;

    [JsonPropertyName("credential")]
    public string Credential { get; set; } =
        string.Empty;
}

public sealed class TerminalAuthRequest
{
    [JsonPropertyName("credential")]
    public string Credential { get; set; } =
        string.Empty;
}

public sealed class TerminalAuthApiResponse
{
    [JsonPropertyName("message")]
    public string? Message { get; set; }

    [JsonPropertyName("data")]
    public TerminalAuthData? Data { get; set; }
}

public sealed class TerminalAuthData
{
    [JsonPropertyName("terminal_id")]
    public string TerminalId { get; set; } =
        string.Empty;

    [JsonPropertyName("tenant_id")]
    public string TenantId { get; set; } =
        string.Empty;

    [JsonPropertyName("branch_id")]
    public string BranchId { get; set; } =
        string.Empty;
}

public sealed class TerminalHeartbeatRequest
{
    [JsonPropertyName("machine_name")]
    public string? MachineName { get; set; }

    [JsonPropertyName("device_identifier")]
    public string? DeviceIdentifier { get; set; }
}

public sealed class TerminalHeartbeatApiResponse
{
    [JsonPropertyName("data")]
    public TerminalHeartbeatData? Data { get; set; }
}

public sealed class TerminalHeartbeatData
{
    [JsonPropertyName("terminal_id")]
    public string TerminalId { get; set; } =
        string.Empty;

    [JsonPropertyName("server_time")]
    public DateTime ServerTime { get; set; }

    [JsonPropertyName("desired_state")]
    public string? DesiredState { get; set; }

    [JsonPropertyName("command_version")]
    public long CommandVersion { get; set; }
}
