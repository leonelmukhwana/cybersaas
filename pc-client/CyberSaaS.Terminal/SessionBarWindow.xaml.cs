using Microsoft.EntityFrameworkCore;
using System;
using System.Globalization;
using System.Linq;
using System.Net.Http;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Media;
using System.Windows.Threading;
using CyberSaaS.Terminal.Infrastructure.Api;
using CyberSaaS.Terminal.Infrastructure.Database;
using CyberSaaS.Terminal.Infrastructure.Sync;
using CyberSaaS.Terminal.Infrastructure.Voice;
using CyberSaaS.Terminal.Storage;

namespace CyberSaaS.Terminal;

public partial class SessionBarWindow : Window
{
    private readonly TerminalRegistration _registration;
    private readonly TerminalSessionApi _sessionApi;
    private readonly TerminalPaymentApi _paymentApi;
    private readonly TerminalSessionData _session;
    private readonly OfflineQueueService _offlineQueueService;

    private readonly DispatcherTimer _timer;
    private readonly TerminalVoiceService _voiceService;

    private DateTime _startedAtUtc;
    private DateTime? _pausedAtUtc;

    private readonly int? _allowedMinutes;
    private readonly string _sessionType;

    private long _totalPausedSeconds;

    private bool _ending;
    private bool _changingPauseState;
    private bool _sessionEndedSuccessfully;

    private bool _fiveMinuteWarningGiven;
    private bool _oneMinuteWarningGiven;
    private bool _expiredWarningGiven;

    /*
     * Remote/server-side session monitoring.
     *
     * The attendant dashboard can end a session remotely.
     * The terminal must detect that completed state and open
     * the payment page without trying to end the session again.
     */
    private int _remoteStateCheckTicks;
    private bool _serverEndedDetected;
    private bool _remoteStateCheckInProgress;

    private string _lastKnownCurrency = "KES";

    public SessionBarWindow(
        TerminalRegistration registration,
        TerminalSessionApi sessionApi,
        TerminalPaymentApi paymentApi,
        TerminalSessionData session)
    {
        InitializeComponent();

        _registration = registration;
        _sessionApi = sessionApi;
        _paymentApi = paymentApi;
        _session = session;

        _offlineQueueService =
            new OfflineQueueService();

        _voiceService =
            new TerminalVoiceService();

        _sessionType =
            string.IsNullOrWhiteSpace(session.SessionType)
                ? "pay_after"
                : session.SessionType;

        _startedAtUtc =
            session.StartedAt.ToUniversalTime();

        _pausedAtUtc =
            session.PausedAt?.ToUniversalTime();

        _totalPausedSeconds =
            session.TotalPausedSeconds;

        _allowedMinutes =
            session.AllowedMinutes;

        /*
         * Remember the currency from the local billing
         * configuration immediately.
         */
        _lastKnownCurrency =
            GetBillingCurrency();

        UpdatePauseButton();

        if (_sessionType.Equals(
                "prepaid",
                StringComparison.OrdinalIgnoreCase))
        {
            SessionTitleTextBlock.Text =
                "Prepaid Time Remaining";
        }
        else
        {
            SessionTitleTextBlock.Text =
                "Pay After";
        }

        PositionWindow();

        _timer = new DispatcherTimer
        {
            Interval = TimeSpan.FromSeconds(1)
        };

        _timer.Tick += Timer_Tick;
        _timer.Start();

        UpdateDisplay();
    }

    private void PositionWindow()
    {
        Left =
            SystemParameters.WorkArea.Right -
            Width -
            10;

        Top =
            SystemParameters.WorkArea.Top +
            10;
    }

    private async void Timer_Tick(
        object? sender,
        EventArgs e)
    {
        if (_ending ||
            _serverEndedDetected)
        {
            return;
        }

        UpdateDisplay();

        /*
         * IMPORTANT:
         *
         * Check the server BEFORE returning for a paused
         * session. This means an attendant can end the
         * session from the dashboard even while the customer
         * session is paused.
         */
        _remoteStateCheckTicks++;

        if (_remoteStateCheckTicks >= 2)
        {
            _remoteStateCheckTicks = 0;

            await CheckRemoteSessionStateAsync();

            if (_serverEndedDetected)
            {
                return;
            }
        }

        if (_pausedAtUtc.HasValue)
        {
            return;
        }

        if (_sessionType.Equals(
                "prepaid",
                StringComparison.OrdinalIgnoreCase))
        {
            await HandlePrepaidExpiry();
        }
        else
        {
            await RefreshBillingPreview();
        }
    }

    private async Task CheckRemoteSessionStateAsync()
    {
        /*
         * Offline-created sessions do not exist on the server
         * yet, so there is nothing to poll.
         */
        if (_session.Id.StartsWith(
                "offline-",
                StringComparison.OrdinalIgnoreCase))
        {
            return;
        }

        if (_remoteStateCheckInProgress ||
            _ending ||
            _serverEndedDetected)
        {
            return;
        }

        _remoteStateCheckInProgress = true;

        try
        {
            TerminalSessionData? serverSession =
                await _sessionApi.GetSessionAsync(
                    _registration,
                    _session.Id);

            if (serverSession == null)
            {
                return;
            }

            /*
             * THE IMPORTANT CASE:
             *
             * The attendant has ended the session from the
             * web dashboard.
             */
            if (serverSession.Status.Equals(
                    "completed",
                    StringComparison.OrdinalIgnoreCase))
            {
                await HandleRemoteSessionEndAsync(
                    serverSession);

                return;
            }

            /*
             * If the session was cancelled remotely,
             * terminate the local active session as well.
             */
            if (serverSession.Status.Equals(
                    "cancelled",
                    StringComparison.OrdinalIgnoreCase))
            {
                _serverEndedDetected = true;
                _sessionEndedSuccessfully = true;

                _timer.Stop();

                RemoveLocalActiveSession();

                Hide();

                ShowMainWindow();

                Close();

                return;
            }

            /*
             * Synchronize remote pause state.
             */
            if (serverSession.Status.Equals(
                    "paused",
                    StringComparison.OrdinalIgnoreCase))
            {
                if (serverSession.PausedAt.HasValue)
                {
                    _pausedAtUtc =
                        serverSession.PausedAt
                            .Value
                            .ToUniversalTime();

                    _totalPausedSeconds =
                        serverSession.TotalPausedSeconds;

                    UpdatePauseButton();
                    SaveLocalSession();
                }

                return;
            }

            /*
             * Synchronize remote active/resumed state.
             */
            if (serverSession.Status.Equals(
                    "active",
                    StringComparison.OrdinalIgnoreCase))
            {
                _startedAtUtc =
                    serverSession.StartedAt.ToUniversalTime();

                _totalPausedSeconds =
                    serverSession.TotalPausedSeconds;

                /*
                 * If the attendant resumed the session remotely,
                 * clear our local paused state.
                 */
                if (!serverSession.PausedAt.HasValue)
                {
                    _pausedAtUtc = null;
                }
                else
                {
                    _pausedAtUtc =
                        serverSession.PausedAt
                            .Value
                            .ToUniversalTime();
                }

                UpdatePauseButton();
                SaveLocalSession();
            }
        }
        catch
        {
            /*
             * A temporary network failure must NOT end the
             * customer's local session.
             *
             * The existing offline billing/session behavior
             * continues.
             */
        }
        finally
        {
            _remoteStateCheckInProgress = false;
        }
    }

    private async Task HandleRemoteSessionEndAsync(
        TerminalSessionData serverSession)
    {
        if (_serverEndedDetected ||
            _ending)
        {
            return;
        }

        _serverEndedDetected = true;
        _sessionEndedSuccessfully = true;
        _ending = true;

        /*
         * The backend has already completed the session.
         *
         * DO NOT call EndSessionAsync().
         */
        _timer.Stop();

        /*
         * Remove the terminal's local active-session record.
         * This prevents the terminal from recovering this
         * completed session again.
         */
        RemoveLocalActiveSession();

        /*
         * The completed session contains final_amount.
         *
         * Use it when available.
         */
        string amount;

        if (!string.IsNullOrWhiteSpace(
                serverSession.FinalAmount))
        {
            amount =
                serverSession.FinalAmount;
        }
        else
        {
            /*
             * Normally final_amount should be present for a
             * completed session. This fallback keeps the
             * payment page usable if an older backend response
             * does not contain it.
             */
            amount =
                CalculateCurrentPayAfterAmount()
                    .ToString(
                        "0.00",
                        CultureInfo.InvariantCulture);
        }

        string currency =
            string.IsNullOrWhiteSpace(
                _lastKnownCurrency)
                ? GetBillingCurrency()
                : _lastKnownCurrency;

        /*
         * Make sure the SessionBar cannot continue running.
         */
        Hide();

        /*
         * Open exactly the same online payment page used when
         * the customer ends a pay-after session on the terminal.
         *
         * online = true enables Cash and M-Pesa.
         */
        OpenPaymentPage(
            amount,
            currency,
            true);
    }

    private void UpdateDisplay()
    {
        if (_pausedAtUtc.HasValue)
        {
            StatusTextBlock.Text =
                "PAUSED";

            StatusTextBlock.Foreground =
                Brushes.DarkOrange;

            if (_sessionType.Equals(
                    "prepaid",
                    StringComparison.OrdinalIgnoreCase))
            {
                TimeSpan remaining =
                    CalculatePrepaidRemaining();

                SessionValueTextBlock.Text =
                    FormatDuration(remaining);
            }
            else
            {
                UpdateLocalPayAfterAmount();
            }

            return;
        }

        StatusTextBlock.Foreground =
            (Brush)Application.Current.Resources[
                "CyberPrimaryBrush"];

        TimeSpan activeElapsed =
            GetActiveElapsed();

        if (_sessionType.Equals(
                "prepaid",
                StringComparison.OrdinalIgnoreCase))
        {
            int allowed =
                _allowedMinutes ?? 0;

            TimeSpan remaining =
                TimeSpan.FromMinutes(allowed) -
                activeElapsed;

            if (remaining < TimeSpan.Zero)
            {
                remaining =
                    TimeSpan.Zero;
            }

            CheckPrepaidVoiceWarnings(remaining);

            SessionValueTextBlock.Text =
                FormatDuration(remaining);

            if (remaining == TimeSpan.Zero)
            {
                SessionValueTextBlock.Foreground =
                    Brushes.Red;

                StatusTextBlock.Text =
                    "TIME EXPIRED";

                return;
            }

            if (remaining.TotalMinutes <= 5)
            {
                SessionValueTextBlock.Foreground =
                    Brushes.Red;

                StatusTextBlock.Text =
                    "ENDING SOON";
            }
            else
            {
                SessionValueTextBlock.Foreground =
                    (Brush)Application.Current.Resources[
                        "CyberPrimaryBrush"];

                StatusTextBlock.Text =
                    "RUNNING";
            }
        }
        else
        {
            UpdateLocalPayAfterAmount();

            StatusTextBlock.Text =
                "RUNNING";
        }
    }

    private void CheckPrepaidVoiceWarnings(
        TimeSpan remaining)
    {
        if (!_sessionType.Equals(
                "prepaid",
                StringComparison.OrdinalIgnoreCase))
        {
            return;
        }

        if (!_fiveMinuteWarningGiven &&
            remaining.TotalSeconds <= 300 &&
            remaining.TotalSeconds > 60)
        {
            _fiveMinuteWarningGiven = true;

            _voiceService.Speak(
                "Your session will expire in five minutes.");
        }

        if (!_oneMinuteWarningGiven &&
            remaining.TotalSeconds <= 60 &&
            remaining.TotalSeconds > 0)
        {
            _oneMinuteWarningGiven = true;

            _voiceService.Speak(
                "Your session will expire in one minute.");
        }

        if (!_expiredWarningGiven &&
            remaining <= TimeSpan.Zero)
        {
            _expiredWarningGiven = true;

            _voiceService.Speak(
                "Your session has expired.");
        }
    }

    private TimeSpan CalculatePrepaidRemaining()
    {
        int allowed =
            _allowedMinutes ?? 0;

        if (allowed <= 0)
        {
            return TimeSpan.Zero;
        }

        TimeSpan remaining =
            TimeSpan.FromMinutes(allowed) -
            GetActiveElapsed();

        if (remaining < TimeSpan.Zero)
        {
            remaining =
                TimeSpan.Zero;
        }

        return remaining;
    }

    private void UpdateLocalPayAfterAmount()
    {
        BranchBillingConfigLocal? config =
            GetLocalBillingConfig();

        if (config == null)
        {
            SessionValueTextBlock.Text =
                $"Time {FormatDuration(GetActiveElapsed())}";

            SessionValueTextBlock.Foreground =
                (Brush)Application.Current.Resources[
                    "CyberPrimaryBrush"];

            return;
        }

        decimal rate =
            ParseMoney(config.RatePerMinute);

        decimal minimum =
            ParseMoney(config.MinimumCharge);

        if (rate <= 0)
        {
            SessionValueTextBlock.Text =
                $"Time {FormatDuration(GetActiveElapsed())}";

            return;
        }

        TimeSpan activeElapsed =
            GetActiveElapsed();

        long billableMinutes =
            (long)Math.Floor(
                activeElapsed.TotalMinutes);

        decimal amount =
            billableMinutes * rate;

        if (amount < minimum)
        {
            amount = minimum;
        }

        amount =
            decimal.Round(
                amount,
                2,
                MidpointRounding.AwayFromZero);

        string currency =
            string.IsNullOrWhiteSpace(config.Currency)
                ? "KES"
                : config.Currency;

        _lastKnownCurrency =
            currency;

        SessionTitleTextBlock.Text =
            $"Pay After • {currency}";

        SessionValueTextBlock.Text =
            $"{currency} {amount:0.00}";

        SessionValueTextBlock.Foreground =
            (Brush)Application.Current.Resources[
                "CyberPrimaryBrush"];
    }

    private BranchBillingConfigLocal? GetLocalBillingConfig()
    {
        try
        {
            using var db =
                new TerminalDbContext();

            string branchId =
                _session.BranchId;

            if (string.IsNullOrWhiteSpace(branchId))
            {
                return null;
            }

            return db.BranchBillingConfigs
                .FirstOrDefault(
                    x => x.BranchId == branchId);
        }
        catch
        {
            return null;
        }
    }

    private static decimal ParseMoney(
        string? value)
    {
        if (string.IsNullOrWhiteSpace(value))
        {
            return 0m;
        }

        if (!decimal.TryParse(
                value,
                NumberStyles.Number,
                CultureInfo.InvariantCulture,
                out decimal result))
        {
            return 0m;
        }

        return result;
    }

    private TimeSpan GetActiveElapsed()
    {
        DateTime endTime =
            _pausedAtUtc?.ToUniversalTime()
            ?? DateTime.UtcNow;

        TimeSpan totalElapsed =
            endTime -
            _startedAtUtc;

        if (totalElapsed.TotalSeconds < 0)
        {
            totalElapsed =
                TimeSpan.Zero;
        }

        TimeSpan paused =
            TimeSpan.FromSeconds(
                _totalPausedSeconds);

        TimeSpan active =
            totalElapsed -
            paused;

        if (active.TotalSeconds < 0)
        {
            active =
                TimeSpan.Zero;
        }

        return active;
    }

    private decimal CalculateCurrentPayAfterAmount()
    {
        BranchBillingConfigLocal? config =
            GetLocalBillingConfig();

        if (config == null)
        {
            return 0m;
        }

        decimal rate =
            ParseMoney(config.RatePerMinute);

        decimal minimum =
            ParseMoney(config.MinimumCharge);

        if (rate <= 0)
        {
            return 0m;
        }

        long billableMinutes =
            (long)Math.Floor(
                GetActiveElapsed().TotalMinutes);

        decimal amount =
            billableMinutes * rate;

        if (amount < minimum)
        {
            amount = minimum;
        }

        return decimal.Round(
            amount,
            2,
            MidpointRounding.AwayFromZero);
    }

    private string GetBillingCurrency()
    {
        BranchBillingConfigLocal? config =
            GetLocalBillingConfig();

        if (config == null ||
            string.IsNullOrWhiteSpace(config.Currency))
        {
            return "KES";
        }

        return config.Currency;
    }

    private async Task HandlePrepaidExpiry()
    {
        if (_ending ||
            _changingPauseState ||
            _sessionEndedSuccessfully ||
            _pausedAtUtc.HasValue)
        {
            return;
        }

        int allowed =
            _allowedMinutes ?? 0;

        if (allowed <= 0)
        {
            return;
        }

        TimeSpan activeElapsed =
            GetActiveElapsed();

        if (activeElapsed.TotalMinutes < allowed)
        {
            return;
        }

        await EndExpiredPrepaidSession();
    }

    private async Task EndExpiredPrepaidSession()
    {
        if (_ending ||
            _sessionEndedSuccessfully)
        {
            return;
        }

        _ending = true;

        EndSessionButton.IsEnabled =
            false;

        PauseResumeButton.IsEnabled =
            false;

        StatusTextBlock.Text =
            "TIME EXPIRED";

        try
        {
            bool backendAvailable = false;

            try
            {
                TerminalSessionWithBilling? ended =
                    await _sessionApi.EndSessionAsync(
                        _registration,
                        _session.Id);

                if (ended != null)
                {
                    backendAvailable = true;

                    RemoveLocalActiveSession();

                    _sessionEndedSuccessfully = true;

                    _timer.Stop();

                    MessageBox.Show(
                        "Prepaid time has expired.\n\n" +
                        "Session ended successfully.\n\n" +
                        $"Amount: {ended.Currency} " +
                        $"{ended.CalculatedAmount}",
                        "Prepaid Session Ended",
                        MessageBoxButton.OK,
                        MessageBoxImage.Information);

                    ShowMainWindow();
                    Close();

                    return;
                }
            }
            catch (HttpRequestException)
            {
                backendAvailable = false;
            }
            catch (InvalidOperationException)
            {
                backendAvailable = false;
            }

            if (!backendAvailable)
            {
                QueueOfflineEnd();

                RemoveLocalActiveSession();

                _sessionEndedSuccessfully = true;

                _timer.Stop();

                MessageBox.Show(
                    "Prepaid time has expired.\n\n" +
                    "The session has ended locally and will be synchronized when the connection returns.",
                    "Offline Session Ended",
                    MessageBoxButton.OK,
                    MessageBoxImage.Information);

                ShowMainWindow();
                Close();
            }
        }
        catch
        {
            _ending = false;

            EndSessionButton.IsEnabled =
                true;

            PauseResumeButton.IsEnabled =
                true;

            StatusTextBlock.Text =
                "EXPIRED • OFFLINE";
        }
    }

    private async void PauseResumeButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        if (_ending ||
            _changingPauseState)
        {
            return;
        }

        await TogglePauseResume();
    }

    private async Task TogglePauseResume()
    {
        _changingPauseState = true;

        PauseResumeButton.IsEnabled =
            false;

        try
        {
            if (_pausedAtUtc.HasValue)
            {
                await ResumeSession();
            }
            else
            {
                await PauseSession();
            }
        }
        catch (Exception ex)
        {
            MessageBox.Show(
                ex.Message,
                "Pause / Resume Failed",
                MessageBoxButton.OK,
                MessageBoxImage.Error);
        }
        finally
        {
            _changingPauseState = false;

            if (!_ending)
            {
                PauseResumeButton.IsEnabled =
                    true;
            }
        }
    }

    private async Task PauseSession()
    {
        /*
         * Offline-created sessions do not exist on the backend yet.
         * Therefore they must always be paused locally.
         */
        if (_session.Id.StartsWith(
                "offline-",
                StringComparison.OrdinalIgnoreCase))
        {
            PauseSessionOffline();
            return;
        }

        try
        {
            TerminalSessionData? result =
                await _sessionApi.PauseSessionAsync(
                    _registration,
                    _session.Id);

            if (result == null)
            {
                throw new InvalidOperationException(
                    "The server returned an invalid pause response.");
            }

            ApplySessionState(result);

            SaveLocalSession();

            UpdatePauseButton();
            UpdateDisplay();
        }
        catch (HttpRequestException)
        {
            PauseSessionOffline();
        }
    }

    private async Task ResumeSession()
    {
        /*
         * Offline-created sessions do not exist on the backend yet.
         * Therefore they must always be resumed locally.
         */
        if (_session.Id.StartsWith(
                "offline-",
                StringComparison.OrdinalIgnoreCase))
        {
            ResumeSessionOffline();
            return;
        }

        try
        {
            TerminalSessionData? result =
                await _sessionApi.ResumeSessionAsync(
                    _registration,
                    _session.Id);

            if (result == null)
            {
                throw new InvalidOperationException(
                    "The server returned an invalid resume response.");
            }

            ApplySessionState(result);

            SaveLocalSession();

            UpdatePauseButton();
            UpdateDisplay();
        }
        catch (HttpRequestException)
        {
            ResumeSessionOffline();
        }
    }

    private void PauseSessionOffline()
    {
        if (_pausedAtUtc.HasValue)
        {
            return;
        }

        _pausedAtUtc =
            DateTime.UtcNow;

        SaveLocalSession();

        _offlineQueueService.Enqueue(
            Guid.NewGuid().ToString(),
            "pause_session",
            new OfflineSessionOperation
            {
                SessionId =
                    _session.Id,

                ClientOperationId =
                    _session.ClientOperationId,

                OccurredAtUtc =
                    DateTime.UtcNow
            });

        UpdatePauseButton();
        UpdateDisplay();

        StatusTextBlock.Text =
            "PAUSED • OFFLINE";
    }

    private void ResumeSessionOffline()
    {
        if (!_pausedAtUtc.HasValue)
        {
            return;
        }

        DateTime resumedAt =
            DateTime.UtcNow;

        TimeSpan currentPause =
            resumedAt -
            _pausedAtUtc.Value;

        if (currentPause.TotalSeconds > 0)
        {
            _totalPausedSeconds +=
                (long)currentPause.TotalSeconds;
        }

        _pausedAtUtc =
            null;

        SaveLocalSession();

        _offlineQueueService.Enqueue(
            Guid.NewGuid().ToString(),
            "resume_session",
            new OfflineSessionOperation
            {
                SessionId =
                    _session.Id,

                ClientOperationId =
                    _session.ClientOperationId,

                OccurredAtUtc =
                    resumedAt
            });

        UpdatePauseButton();
        UpdateDisplay();

        StatusTextBlock.Text =
            "RUNNING • OFFLINE";
    }

    private void ApplySessionState(
        TerminalSessionData result)
    {
        _startedAtUtc =
            result.StartedAt.ToUniversalTime();

        _pausedAtUtc =
            result.PausedAt?.ToUniversalTime();

        _totalPausedSeconds =
            result.TotalPausedSeconds;
    }

    private void UpdatePauseButton()
    {
        PauseResumeButton.Content =
            _pausedAtUtc.HasValue
                ? "Resume"
                : "Pause";
    }

    private async Task RefreshBillingPreview()
    {
        if (_ending ||
            _pausedAtUtc.HasValue)
        {
            return;
        }

        /*
         * Offline-created sessions do not have a server
         * billing preview yet.
         */
        if (_session.Id.StartsWith(
                "offline-",
                StringComparison.OrdinalIgnoreCase))
        {
            UpdateLocalPayAfterAmount();

            StatusTextBlock.Text =
                "OFFLINE";

            return;
        }

        try
        {
            TerminalBillingPreview? preview =
                await _sessionApi.GetBillingPreviewAsync(
                    _registration,
                    _session.Id);

            if (preview == null)
            {
                return;
            }

            _lastKnownCurrency =
                string.IsNullOrWhiteSpace(preview.Currency)
                    ? "KES"
                    : preview.Currency;

            SessionTitleTextBlock.Text =
                $"Pay After • {_lastKnownCurrency}";

            SessionValueTextBlock.Text =
                $"{_lastKnownCurrency} {preview.CalculatedAmount}";

            StatusTextBlock.Text =
                "ONLINE";
        }
        catch
        {
            UpdateLocalPayAfterAmount();

            StatusTextBlock.Text =
                "OFFLINE";
        }
    }

    private async void EndSessionButton_Click(
        object sender,
        RoutedEventArgs e)
    {
        if (_ending ||
            _serverEndedDetected)
        {
            return;
        }

        MessageBoxResult result =
            MessageBox.Show(
                "End this customer session?",
                "End Session",
                MessageBoxButton.YesNo,
                MessageBoxImage.Question);

        if (result != MessageBoxResult.Yes)
        {
            return;
        }

        await EndSessionManually();
    }

    private async Task EndSessionManually()
    {
        if (_ending ||
            _serverEndedDetected)
        {
            return;
        }

        _ending = true;

        EndSessionButton.IsEnabled =
            false;

        PauseResumeButton.IsEnabled =
            false;

        StatusTextBlock.Text =
            "ENDING...";

        try
        {
            /*
             * Pay After:
             * End the session first, then open the payment page.
             */
            if (_sessionType.Equals(
                    "pay_after",
                    StringComparison.OrdinalIgnoreCase))
            {
                await EndPayAfterSession();

                return;
            }

            /*
             * Preserve the existing prepaid behavior.
             */
            try
            {
                TerminalSessionWithBilling? ended =
                    await _sessionApi.EndSessionAsync(
                        _registration,
                        _session.Id);

                RemoveLocalActiveSession();

                _sessionEndedSuccessfully = true;

                _timer.Stop();

                MessageBox.Show(
                    ended == null
                        ? "Session ended successfully."
                        : $"Session ended successfully.\n\n" +
                          $"Amount: {ended.Currency} " +
                          $"{ended.CalculatedAmount}",
                    "Session Ended",
                    MessageBoxButton.OK,
                    MessageBoxImage.Information);

                ShowMainWindow();

                Close();

                return;
            }
            catch (HttpRequestException)
            {
            }
            catch (InvalidOperationException)
            {
                if (!_session.Id.StartsWith(
                        "offline-",
                        StringComparison.OrdinalIgnoreCase))
                {
                    throw;
                }
            }

            QueueOfflineEnd();

            RemoveLocalActiveSession();

            _sessionEndedSuccessfully = true;

            _timer.Stop();

            MessageBox.Show(
                "Session ended locally.\n\n" +
                "The end operation has been saved and will be synchronized when the connection returns.",
                "Offline Session Ended",
                MessageBoxButton.OK,
                MessageBoxImage.Information);

            ShowMainWindow();
            Close();
        }
        catch (Exception ex)
        {
            StatusTextBlock.Text =
                "Could not end session.";

            EndSessionButton.IsEnabled =
                true;

            PauseResumeButton.IsEnabled =
                true;

            _ending = false;

            MessageBox.Show(
                ex.Message,
                "End Session Failed",
                MessageBoxButton.OK,
                MessageBoxImage.Error);
        }
    }

    private async Task EndPayAfterSession()
    {
        /*
         * Keep the amount before ending/removing the local session
         * because offline billing depends on the local configuration.
         */
        decimal offlineAmount =
            CalculateCurrentPayAfterAmount();

        string offlineCurrency =
            GetBillingCurrency();

        TerminalSessionWithBilling? ended =
            null;

        bool onlineEndSucceeded = false;

        try
        {
            ended =
                await _sessionApi.EndSessionAsync(
                    _registration,
                    _session.Id);

            if (ended != null)
            {
                onlineEndSucceeded = true;
            }
        }
        catch (HttpRequestException)
        {
            onlineEndSucceeded = false;
        }
        catch (InvalidOperationException)
        {
            if (!_session.Id.StartsWith(
                    "offline-",
                    StringComparison.OrdinalIgnoreCase))
            {
                throw;
            }
        }

        if (onlineEndSucceeded &&
            ended != null)
        {
            RemoveLocalActiveSession();

            _sessionEndedSuccessfully = true;

            _timer.Stop();

            _ending = false;

            Hide();

            OpenPaymentPage(
                ended.CalculatedAmount,
                ended.Currency,
                true);

            return;
        }

        /*
         * Offline session end.
         *
         * The session end is queued, but payment is not marked
         * as successfully paid because the backend cannot be
         * reached.
         */
        QueueOfflineEnd();

        RemoveLocalActiveSession();

        _sessionEndedSuccessfully = true;

        _timer.Stop();

        _ending = false;

        Hide();

        OpenPaymentPage(
            offlineAmount.ToString(
                "0.00",
                CultureInfo.InvariantCulture),
            offlineCurrency,
            false);
    }

    private void OpenPaymentPage(
        string amount,
        string currency,
        bool online)
    {
        var paymentWindow =
            new PaymentPageWindow(
                _registration,
                _paymentApi,
                _session.Id,
                amount,
                currency,
                online);

        paymentWindow.Owner =
            Application.Current.Windows
                .OfType<MainWindow>()
                .FirstOrDefault();

        paymentWindow.Show();
    }

    private void QueueOfflineEnd()
    {
        _offlineQueueService.Enqueue(
            Guid.NewGuid().ToString(),
            "end_session",
            new OfflineSessionOperation
            {
                SessionId =
                    _session.Id,

                ClientOperationId =
                    _session.ClientOperationId,

                OccurredAtUtc =
                    DateTime.UtcNow
            });
    }

    private void SaveLocalSession()
    {
        try
        {
            using var db =
                new TerminalDbContext();

            ActiveSessionLocal? localSession =
                db.ActiveSessions
                    .FirstOrDefault(
                        x =>
                            x.SessionId ==
                            _session.Id);

            if (localSession == null)
            {
                return;
            }

            localSession.State =
                _pausedAtUtc.HasValue
                    ? "paused"
                    : "running";

            localSession.StartedAtUtc =
                _startedAtUtc;

            localSession.PausedAtUtc =
                _pausedAtUtc;

            localSession.TotalPausedSeconds =
                _totalPausedSeconds;

            localSession.UpdatedAtUtc =
                DateTime.UtcNow;

            db.SaveChanges();
        }
        catch
        {
        }
    }

    private void RemoveLocalActiveSession()
    {
        try
        {
            using var db =
                new TerminalDbContext();

            ActiveSessionLocal? localSession =
                db.ActiveSessions
                    .FirstOrDefault(
                        x =>
                            x.SessionId ==
                            _session.Id);

            if (localSession == null)
            {
                return;
            }

            db.ActiveSessions.Remove(
                localSession);

            db.SaveChanges();
        }
        catch
        {
        }
    }

    private void ShowMainWindow()
    {
        Application.Current.Dispatcher.Invoke(
            () =>
            {
                MainWindow? mainWindow =
                    Application.Current.Windows
                        .OfType<MainWindow>()
                        .FirstOrDefault();

                if (mainWindow == null)
                {
                    return;
                }

                mainWindow.Show();

                mainWindow.WindowState =
                    WindowState.Normal;

                mainWindow.Activate();
                mainWindow.Focus();
            });
    }

    private static string FormatDuration(
        TimeSpan duration)
    {
        if (duration.TotalHours >= 1)
        {
            return
                $"{(int)duration.TotalHours:00}:" +
                $"{duration.Minutes:00}:" +
                $"{duration.Seconds:00}";
        }

        return
            $"{duration.Minutes:00}:" +
            $"{duration.Seconds:00}";
    }

    protected override void OnClosed(
        EventArgs e)
    {
        _timer.Stop();

        _voiceService.Dispose();

        if (_sessionEndedSuccessfully &&
            Application.Current.Windows
                .OfType<PaymentPageWindow>()
                .FirstOrDefault() == null)
        {
            ShowMainWindow();
        }

        base.OnClosed(e);
    }
}

public sealed class OfflineSessionOperation
{
    public string SessionId { get; set; } =
        string.Empty;

    public string ClientOperationId { get; set; } =
        string.Empty;

    public DateTime OccurredAtUtc { get; set; }
}