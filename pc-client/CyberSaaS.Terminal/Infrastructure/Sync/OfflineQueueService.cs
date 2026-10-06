using System;
using System.Collections.Generic;
using System.Linq;
using System.Text.Json;
using System.Text.Json.Serialization;
using CyberSaaS.Terminal.Infrastructure.Database;

namespace CyberSaaS.Terminal.Infrastructure.Sync
{
    public sealed class OfflineQueueService
    {
        public void Enqueue<T>(
            string operationId,
            string operationType,
            T payload)
        {
            if (string.IsNullOrWhiteSpace(operationId))
            {
                throw new ArgumentException(
                    "Operation ID is required.",
                    nameof(operationId));
            }

            if (string.IsNullOrWhiteSpace(operationType))
            {
                throw new ArgumentException(
                    "Operation type is required.",
                    nameof(operationType));
            }

            string json =
                JsonSerializer.Serialize(payload);

            using var db = new TerminalDbContext();

            bool alreadyExists =
                db.OfflineQueueItems.Any(
                    x => x.OperationId == operationId);

            if (alreadyExists)
            {
                return;
            }

            db.OfflineQueueItems.Add(
                new OfflineQueueItem
                {
                    OperationId = operationId,
                    OperationType = operationType,
                    Payload = json,
                    CreatedAtUtc = DateTime.UtcNow,
                    AttemptCount = 0,
                    Status = "pending"
                });

            db.SaveChanges();
        }

        public List<OfflineQueueItem> GetPending(
            int limit = 50)
        {
            if (limit <= 0)
            {
                limit = 50;
            }

            if (limit > 500)
            {
                limit = 500;
            }

            using var db = new TerminalDbContext();

            return db.OfflineQueueItems
                .Where(x => x.Status == "pending")
                .OrderBy(x => x.Id)
                .Take(limit)
                .ToList();
        }


        public void ReplaceQueuedSessionId(
    string clientOperationId,
    string oldSessionId,
    string serverSessionId)
{
    if (string.IsNullOrWhiteSpace(clientOperationId) ||
        string.IsNullOrWhiteSpace(oldSessionId) ||
        string.IsNullOrWhiteSpace(serverSessionId))
    {
        return;
    }

    using var db = new TerminalDbContext();

    var items =
        db.OfflineQueueItems
            .Where(x =>
                x.Status == "pending" &&
                x.OperationType != "start_session")
            .OrderBy(x => x.Id)
            .ToList();

    bool changed = false;

    foreach (OfflineQueueItem item in items)
    {
        OfflineSessionOperationPayload? operation;

        try
        {
            operation =
                JsonSerializer.Deserialize<OfflineSessionOperationPayload>(
                    item.Payload);
        }
        catch
        {
            continue;
        }

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

        if (!string.Equals(
                operation.SessionId,
                oldSessionId,
                StringComparison.Ordinal))
        {
            continue;
        }

        operation.SessionId = serverSessionId;

        item.Payload =
            JsonSerializer.Serialize(operation);

        changed = true;
    }

    if (changed)
    {
        db.SaveChanges();
    }
}

    private sealed class OfflineSessionOperationPayload
    {
        [JsonPropertyName("session_id")]
        public string SessionId { get; set; } = string.Empty;

        [JsonPropertyName("client_operation_id")]
        public string ClientOperationId { get; set; } = string.Empty;

        [JsonPropertyName("occurred_at_utc")]
        public DateTime OccurredAtUtc { get; set; }
    }

        public void MarkAttempt(
            long id,
            string? error = null)
        {
            using var db = new TerminalDbContext();

            OfflineQueueItem? item =
                db.OfflineQueueItems
                    .FirstOrDefault(x => x.Id == id);

            if (item == null)
            {
                return;
            }

            item.AttemptCount++;
            item.LastAttemptAtUtc = DateTime.UtcNow;
            item.LastError = error;

            db.SaveChanges();
        }

        public void MarkCompleted(long id)
        {
            using var db = new TerminalDbContext();

            OfflineQueueItem? item =
                db.OfflineQueueItems
                    .FirstOrDefault(x => x.Id == id);

            if (item == null)
            {
                return;
            }

            item.Status = "completed";
            item.LastError = null;

            db.SaveChanges();
        }

        public void MarkFailed(
            long id,
            string error)
        {
            using var db = new TerminalDbContext();

            OfflineQueueItem? item =
                db.OfflineQueueItems
                    .FirstOrDefault(x => x.Id == id);

            if (item == null)
            {
                return;
            }

            item.Status = "failed";
            item.LastError = error;
            item.LastAttemptAtUtc = DateTime.UtcNow;

            db.SaveChanges();
        }

        public void DeleteCompleted()
        {
            using var db = new TerminalDbContext();

            List<OfflineQueueItem> completed =
                db.OfflineQueueItems
                    .Where(x => x.Status == "completed")
                    .ToList();

            if (completed.Count == 0)
            {
                return;
            }

            db.OfflineQueueItems.RemoveRange(completed);

            db.SaveChanges();
        }
    }
}