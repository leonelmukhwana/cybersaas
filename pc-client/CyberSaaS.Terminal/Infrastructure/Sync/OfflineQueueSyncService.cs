using System;
using System.Net.Http;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using CyberSaaS.Terminal.Infrastructure.Api;
using CyberSaaS.Terminal.Infrastructure.Database;
using CyberSaaS.Terminal.Storage;
using System.Text.Json.Serialization;

namespace CyberSaaS.Terminal.Infrastructure.Sync;

public sealed class OfflineQueueSyncService
{
    private readonly OfflineQueueService _queue;
    private readonly TerminalSessionApi _sessionApi;
    private readonly TerminalRegistration _registration;

    private int _syncRunning;

    public OfflineQueueSyncService(
        OfflineQueueService queue,
        TerminalSessionApi sessionApi,
        TerminalRegistration registration)
    {
        _queue = queue;
        _sessionApi = sessionApi;
        _registration = registration;
    }

    public async Task SyncPendingAsync()
    {
        /*
         * Prevent two heartbeat ticks from synchronizing the same
         * queue simultaneously.
         */
        if (Interlocked.Exchange(
                ref _syncRunning,
                1) == 1)
        {
            return;
        }

        try
        {
            var items =
                _queue.GetPending();

            foreach (OfflineQueueItem item in items)
            {
                try
                {
                    _queue.MarkAttempt(item.Id);

                    switch (item.OperationType)
                    {
                        case "start_session":
                            await SyncStartSessionAsync(item);
                            break;

                        case "pause_session":
                            await SyncPauseSessionAsync(item);
                            break;

                        case "resume_session":
                            await SyncResumeSessionAsync(item);
                            break;

                        case "end_session":
                            await SyncEndSessionAsync(item);
                            break;

                        default:
                            _queue.MarkFailed(
                                item.Id,
                                $"Unknown operation type: {item.OperationType}");
                            break;
                    }
                }
                catch (HttpRequestException ex)
                {
                    /*
                     * Network problem:
                     * leave it pending so the next heartbeat retries it.
                     */
                    _queue.MarkAttempt(
                        item.Id,
                        ex.Message);
                }
                catch (Exception ex)
                {
                    /*
                     * Do not lose queued data because of a temporary
                     * backend/API problem.
                     *
                     * The operation remains pending and will be retried.
                     */
                    _queue.MarkAttempt(
                        item.Id,
                        ex.Message);
                }
            }
        }
        finally
        {
            Interlocked.Exchange(
                ref _syncRunning,
                0);
        }
    }

    private async Task SyncStartSessionAsync(
        OfflineQueueItem item)
    {
        TerminalStartSessionRequest? request =
            Deserialize<TerminalStartSessionRequest>(
                item.Payload);

        if (request == null)
        {
            _queue.MarkFailed(
                item.Id,
                "Queued session request could not be read.");

            return;
        }

        if (string.IsNullOrWhiteSpace(
                request.ClientOperationId))
        {
            _queue.MarkFailed(
                item.Id,
                "Queued session is missing client operation ID.");

            return;
        }

        try
        {
            TerminalSessionData? result =
                await _sessionApi.StartSessionAsync(
                    _registration,
                    request);

            if (result == null ||
                string.IsNullOrWhiteSpace(result.Id))
            {
                _queue.MarkAttempt(
                    item.Id,
                    "Backend returned an invalid session response.");

                return;
            }

            /*
             * IMPORTANT:
             *
             * Offline sessions use an ID such as:
             *
             * offline-xxxxxxxx
             *
             * The backend creates the real UUID.
             *
             * Replace that temporary ID in every queued
             * pause/resume/end operation belonging to this
             * client operation.
             */
            string oldOfflineSessionId =
                FindOfflineSessionId(
                    request.ClientOperationId);

            if (!string.IsNullOrWhiteSpace(
                    oldOfflineSessionId))
            {
                _queue.ReplaceQueuedSessionId(
                    request.ClientOperationId,
                    oldOfflineSessionId,
                    result.Id);
            }

            /*
             * Also update the local active session if it still exists.
             *
             * This is useful when the session has not yet been removed
             * by the offline end flow.
             */
            UpdateLocalActiveSession(
                request.ClientOperationId,
                result);

            _queue.MarkCompleted(item.Id);
        }
        catch (InvalidOperationException ex)
        {
            /*
             * A 409 such as:
             *
             * "terminal already has an active session"
             *
             * is a permanent conflict for this queued start.
             *
             * Do not retry it forever on every heartbeat.
             */
            if (ex.Message.Contains(
                    "terminal already has an active session",
                    StringComparison.OrdinalIgnoreCase))
            {
                _queue.MarkFailed(
                    item.Id,
                    ex.Message);

                return;
            }

            /*
             * Other API errors remain pending for now.
             */
            _queue.MarkAttempt(
                item.Id,
                ex.Message);
        }
    }

    private async Task SyncPauseSessionAsync(
        OfflineQueueItem item)
    {
        OfflineSessionOperationPayload? operation =
            Deserialize<OfflineSessionOperationPayload>(
                item.Payload);

        if (operation == null)
        {
            _queue.MarkFailed(
                item.Id,
                "Queued pause operation could not be read.");

            return;
        }

        if (string.IsNullOrWhiteSpace(
                operation.SessionId))
        {
            _queue.MarkAttempt(
                item.Id,
                "Pause operation has no session ID.");

            return;
        }

        /*
         * If the ID is still an offline ID, the start operation
         * has not successfully synchronized yet.
         *
         * Keep this operation pending.
         */
        if (IsOfflineSessionId(
                operation.SessionId))
        {
            _queue.MarkAttempt(
                item.Id,
                "Waiting for offline session start to synchronize.");

            return;
        }

        try
        {
            TerminalSessionData? result =
                await _sessionApi.PauseSessionAsync(
                    _registration,
                    operation.SessionId);

            if (result != null)
            {
                _queue.MarkCompleted(item.Id);
                return;
            }

            _queue.MarkAttempt(
                item.Id,
                "Backend returned an empty pause response.");
        }
        catch (HttpRequestException)
        {
            throw;
        }
        catch (Exception ex)
        {
            /*
             * The request may actually have reached the backend
             * before the connection failed.
             *
             * Check the server before retrying blindly.
             */
            if (await IsSessionPausedAsync(
                    operation.SessionId))
            {
                _queue.MarkCompleted(item.Id);
                return;
            }

            _queue.MarkAttempt(
                item.Id,
                ex.Message);
        }
    }

    private async Task SyncResumeSessionAsync(
        OfflineQueueItem item)
    {
        OfflineSessionOperationPayload? operation =
            Deserialize<OfflineSessionOperationPayload>(
                item.Payload);

        if (operation == null)
        {
            _queue.MarkFailed(
                item.Id,
                "Queued resume operation could not be read.");

            return;
        }

        if (string.IsNullOrWhiteSpace(
                operation.SessionId))
        {
            _queue.MarkAttempt(
                item.Id,
                "Resume operation has no session ID.");

            return;
        }

        if (IsOfflineSessionId(
                operation.SessionId))
        {
            _queue.MarkAttempt(
                item.Id,
                "Waiting for offline session start to synchronize.");

            return;
        }

        try
        {
            TerminalSessionData? result =
                await _sessionApi.ResumeSessionAsync(
                    _registration,
                    operation.SessionId);

            if (result != null)
            {
                _queue.MarkCompleted(item.Id);
                return;
            }

            _queue.MarkAttempt(
                item.Id,
                "Backend returned an empty resume response.");
        }
        catch (HttpRequestException)
        {
            throw;
        }
        catch (Exception ex)
        {
            /*
             * If the resume request actually succeeded but the
             * response was lost, the backend should now show
             * the session as running.
             */
            if (await IsSessionRunningAsync(
                    operation.SessionId))
            {
                _queue.MarkCompleted(item.Id);
                return;
            }

            _queue.MarkAttempt(
                item.Id,
                ex.Message);
        }
    }

    private async Task SyncEndSessionAsync(
        OfflineQueueItem item)
    {
        OfflineSessionOperationPayload? operation =
            Deserialize<OfflineSessionOperationPayload>(
                item.Payload);

        if (operation == null)
        {
            _queue.MarkFailed(
                item.Id,
                "Queued end operation could not be read.");

            return;
        }

        if (string.IsNullOrWhiteSpace(
                operation.SessionId))
        {
            _queue.MarkAttempt(
                item.Id,
                "End operation has no session ID.");

            return;
        }

        if (IsOfflineSessionId(
                operation.SessionId))
        {
            _queue.MarkAttempt(
                item.Id,
                "Waiting for offline session start to synchronize.");

            return;
        }

        try
        {
            TerminalSessionWithBilling? result =
                await _sessionApi.EndSessionAsync(
                    _registration,
                    operation.SessionId);

            if (result != null)
            {
                _queue.MarkCompleted(item.Id);
                return;
            }

            _queue.MarkAttempt(
                item.Id,
                "Backend returned an empty end-session response.");
        }
        catch (HttpRequestException)
        {
            throw;
        }
        catch (Exception ex)
        {
            /*
             * If the backend already ended the session but the
             * response was lost, verify the server state.
             */
            if (await IsSessionCompletedAsync(
                    operation.SessionId))
            {
                _queue.MarkCompleted(item.Id);
                return;
            }

            _queue.MarkAttempt(
                item.Id,
                ex.Message);
        }
    }

    private async Task<bool> IsSessionPausedAsync(
        string sessionId)
    {
        try
        {
            TerminalSessionData? session =
                await _sessionApi.GetSessionAsync(
                    _registration,
                    sessionId);

            return session != null &&
                   string.Equals(
                       session.Status,
                       "paused",
                       StringComparison.OrdinalIgnoreCase);
        }
        catch
        {
            return false;
        }
    }

    private async Task<bool> IsSessionRunningAsync(
        string sessionId)
    {
        try
        {
            TerminalSessionData? session =
                await _sessionApi.GetSessionAsync(
                    _registration,
                    sessionId);

            return session != null &&
                   (
                       string.Equals(
                           session.Status,
                           "active",
                           StringComparison.OrdinalIgnoreCase)
                       ||
                       string.Equals(
                           session.Status,
                           "running",
                           StringComparison.OrdinalIgnoreCase)
                   );
        }
        catch
        {
            return false;
        }
    }

    private async Task<bool> IsSessionCompletedAsync(
        string sessionId)
    {
        try
        {
            TerminalSessionData? session =
                await _sessionApi.GetSessionAsync(
                    _registration,
                    sessionId);

            if (session == null)
            {
                return false;
            }

            return
                string.Equals(
                    session.Status,
                    "completed",
                    StringComparison.OrdinalIgnoreCase)
                ||
                string.Equals(
                    session.Status,
                    "cancelled",
                    StringComparison.OrdinalIgnoreCase);
        }
        catch
        {
            return false;
        }
    }

    private static bool IsOfflineSessionId(
        string sessionId)
    {
        return sessionId.StartsWith(
            "offline-",
            StringComparison.OrdinalIgnoreCase);
    }

    private string FindOfflineSessionId(
        string clientOperationId)
    {
        /*
         * First look at the local active session.
         */
        try
        {
            using var db = new TerminalDbContext();

            ActiveSessionLocal? localSession =
                db.ActiveSessions
                    .FirstOrDefault(
                        x =>
                            x.ClientOperationId ==
                            clientOperationId);

            if (localSession != null &&
                IsOfflineSessionId(
                    localSession.SessionId))
            {
                return localSession.SessionId;
            }
        }
        catch
        {
            /*
             * Queue synchronization must not crash because
             * the optional local mapping is unavailable.
             */
        }

        /*
         * If the active-session record was already removed,
         * inspect the pending operations.
         */
        foreach (OfflineQueueItem pending
                 in _queue.GetPending())
        {
            OfflineSessionOperationPayload? operation =
                Deserialize<OfflineSessionOperationPayload>(
                    pending.Payload);

            if (operation == null)
            {
                continue;
            }

            if (!string.Equals(
                    operation.ClientOperationId,
                    clientOperationId,
                    StringComparison.Ordinal))
            {
                continue;
            }

            if (IsOfflineSessionId(
                    operation.SessionId))
            {
                return operation.SessionId;
            }
        }

        return string.Empty;
    }

    private static void UpdateLocalActiveSession(
        string clientOperationId,
        TerminalSessionData serverSession)
    {
        try
        {
            using var db = new TerminalDbContext();

            ActiveSessionLocal? localSession =
                db.ActiveSessions
                    .FirstOrDefault(
                        x =>
                            x.ClientOperationId ==
                            clientOperationId);

            if (localSession == null)
            {
                return;
            }

            localSession.SessionId =
                serverSession.Id;

            localSession.CustomerId =
                serverSession.CustomerId;

            localSession.TerminalId =
                serverSession.TerminalId;

            localSession.BranchId =
                serverSession.BranchId;

            localSession.SessionType =
                serverSession.SessionType;

            localSession.State =
                MapServerState(
                    serverSession.Status);

            localSession.StartedAtUtc =
                serverSession.StartedAt.ToUniversalTime();

            localSession.PausedAtUtc =
                serverSession.PausedAt?
                    .ToUniversalTime();

            localSession.EndedAtUtc =
                serverSession.EndedAt?
                    .ToUniversalTime();

            localSession.TotalPausedSeconds =
                serverSession.TotalPausedSeconds;

            localSession.PrepaidAmount =
                serverSession.PrepaidAmount;

            localSession.AllowedMinutes =
                serverSession.AllowedMinutes;

            localSession.RatePerMinute =
                serverSession.RatePerMinute;

            localSession.MinimumCharge =
                serverSession.MinimumCharge;

            localSession.UpdatedAtUtc =
                DateTime.UtcNow;

            db.SaveChanges();
        }
        catch
        {
            /*
             * Server synchronization is the important operation.
             * Failure to update the optional local record must not
             * cause the queue item to be replayed unnecessarily.
             */
        }
    }

    private static string MapServerState(
        string status)
    {
        if (string.Equals(
                status,
                "paused",
                StringComparison.OrdinalIgnoreCase))
        {
            return "paused";
        }

        if (string.Equals(
                status,
                "completed",
                StringComparison.OrdinalIgnoreCase))
        {
            return "completed";
        }

        return "running";
    }

    private static T? Deserialize<T>(
        string json)
    {
        try
        {
            return JsonSerializer.Deserialize<T>(
                json);
        }
        catch
        {
            return default;
        }
    }

    private sealed class OfflineSessionOperationPayload
    {
        public string SessionId { get; set; } =
            string.Empty;

        public string ClientOperationId { get; set; } =
            string.Empty;

        public DateTime OccurredAtUtc { get; set; }
    }
}