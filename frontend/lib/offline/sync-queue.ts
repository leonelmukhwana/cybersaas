import {
  addToSyncQueue,
  getPendingSyncItems,
  markSyncCompleted,
  markSyncFailed,
  markSyncProcessing,
  recoverProcessingSyncItems,
  retryFailedSyncItems,
  type SyncQueueItem,
} from "@/lib/offline/db";

export interface SyncEventRequest {
  id: string;
  event_type: string;
  entity_type: string;
  entity_id: string;
  payload: Record<string, unknown>;
  user_id?: string;
}

export interface SyncEventResponse {
  id: string;
  status: string;
}

export interface SyncQueueResult {
  processed: number;
  failed: number;
  remaining: number;
}

function mapEventType(operation: SyncQueueItem["operation"]): string {
  switch (operation) {
    case "create":
      return "create";

    case "update":
      return "update";

    case "confirm":
      return "confirm";

    case "print":
      return "print";

    case "reprint":
      return "reprint";

    default:
      return operation;
  }
}

function buildSyncEvent(
  item: SyncQueueItem,
): SyncEventRequest {
  const event: SyncEventRequest = {
    id: item.id,
    event_type: mapEventType(item.operation),
    entity_type: item.entity,
    entity_id: item.entity_id,
    payload: item.payload,
  };

  if (item.user_id) {
    event.user_id = item.user_id;
  }

  return event;
}

export async function enqueueSyncItem(
  item: SyncQueueItem,
): Promise<void> {
  await addToSyncQueue(item);
}

export async function getSyncQueueItems(): Promise<
  SyncQueueItem[]
> {
  return getPendingSyncItems();
}

export async function recoverSyncQueue(): Promise<void> {
  await recoverProcessingSyncItems();
}

export async function retrySyncQueue(): Promise<void> {
  await retryFailedSyncItems();
}

export async function processSyncQueue(
  pushEvent: (
    event: SyncEventRequest,
  ) => Promise<SyncEventResponse>,
): Promise<SyncQueueResult> {
  await recoverProcessingSyncItems();

  const items = await getPendingSyncItems();

  let processed = 0;
  let failed = 0;

  for (const item of items) {
    try {
      await markSyncProcessing(item.id);

      const event = buildSyncEvent(item);

      const response = await pushEvent(event);

      if (
        response.status === "already_known" ||
        response.status === "pending" ||
        response.status === "processed" ||
        response.status === "success"
      ) {
        await markSyncCompleted(item.id);
        processed += 1;
        continue;
      }

      throw new Error(
        `Server returned unexpected sync status: ${response.status}`,
      );
    } catch (error) {
      failed += 1;

      const message =
        error instanceof Error
          ? error.message
          : "Unknown synchronization error";

      await markSyncFailed(
        item.id,
        message,
      );
    }
  }

  const remaining =
    (await getPendingSyncItems()).length;

  return {
    processed,
    failed,
    remaining,
  };
}