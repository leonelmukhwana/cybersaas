
import Dexie, { Table } from "dexie";

/* =========================================================
   TYPES
========================================================= */

export type UserRole = "attendant";

export type AccountStatus =
  | "active"
  | "inactive"
  | "suspended";

export type CustomerType =
  | "adult"
  | "child";

export type SyncOperation =
  | "create"
  | "update"
  | "confirm"
  | "print"
  | "reprint";

export type SyncEntity =
  | "customer"
  | "sale"
  | "payment"
  | "receipt"
  | "receipt_reprint";

export type SyncStatus =
  | "pending"
  | "processing"
  | "failed";

/* =========================================================
   SYNC QUEUE
========================================================= */

export interface SyncQueueItem {
  id: string;

  /*
   * Idempotency key.
   *
   * This MUST remain the same for every retry
   * of the same operation.
   */
  client_operation_id: string;

  tenant_id: string;
  user_id: string;
  branch_id: string;

  entity: SyncEntity;
  operation: SyncOperation;

  entity_id: string;

  payload: Record<string, unknown>;

  status: SyncStatus;

  attempts: number;

  last_error?: string | null;

  created_at: string;
  updated_at: string;
}

/* =========================================================
   OFFLINE ATTENDANT CREDENTIAL
========================================================= */

export interface OfflineCredential {
  id: string;

  user_id: string;
  tenant_id: string;

  email: string;

  role: UserRole;

  branch_id: string;
  branch_name: string;

  /*
   * NEVER store the actual password.
   */
  password_verifier: string;
  password_salt: string;

  account_status: AccountStatus;

  credential_version: number;

  last_online_verified_at: string;

  created_at: string;
  updated_at: string;
}

/* =========================================================
   LOCAL BRANCH
========================================================= */

export interface LocalBranch {
  id: string;

  tenant_id: string;

  name: string;

  address?: string | null;
  phone?: string | null;

  status:
    | "active"
    | "inactive";

  attendant_user_id: string;

  created_at: string;
  updated_at: string;
}

/* =========================================================
   LOCAL SERVICE
========================================================= */

export interface LocalService {
  id: string;

  tenant_id: string;
  branch_id: string;

  name: string;

  description?: string | null;

  price: string;

  active: boolean;

  created_at: string;
  updated_at: string;

  /*
   * When this service was cached locally.
   */
  cached_at: string;
}

/* =========================================================
   LOCAL CUSTOMER
========================================================= */

export interface LocalCustomer {
  id: string;

  tenant_id: string;
  branch_id: string;

  customer_type: CustomerType;

  full_name: string;

  phone?: string | null;

  parent_name?: string | null;
  parent_phone?: string | null;

  sync_status:
    | "synced"
    | "pending"
    | "failed";

  server_updated_at?: string | null;

  created_at: string;
  updated_at: string;
}

/* =========================================================
   LOCAL SALE
========================================================= */

export type LocalSaleStatus =
  | "completed"
  | "voided"
  | "refunded";

export interface LocalSale {
  id: string;

  tenant_id: string;
  branch_id: string;

  customer_id?: string | null;

  session_id?: string | null;
  session_started_at?: string | null;

  terminal_id?: string | null;
  attendant_id?: string | null;

  subtotal: string;

  discount_type?: string | null;
  discount_value: string;
  discount_amount: string;

  total_amount: string;

  currency: "KES";

  status: LocalSaleStatus;

  client_operation_id: string;

  sync_status:
    | "synced"
    | "pending"
    | "failed";

  server_updated_at?: string | null;

  created_at: string;
  updated_at: string;
}

/* =========================================================
   LOCAL SALE ITEM
========================================================= */

export interface LocalSaleItem {
  id: string;

  sale_id: string;

  service_id?: string | null;

  description: string;

  quantity: string;

  unit_price: string;

  line_total: string;

  created_at: string;
}

/* =========================================================
   LOCAL PAYMENT
========================================================= */

export type LocalPaymentMethod =
  | "cash"
  | "mpesa"
  | "other";

export type LocalPaymentStatus =
  | "pending"
  | "confirmed"
  | "failed"
  | "refunded";

export interface LocalPayment {
  id: string;

  tenant_id: string;
  branch_id: string;
  sale_id: string;

  method: LocalPaymentMethod;

  status: LocalPaymentStatus;

  amount: string;

  phone?: string | null;

  external_reference?: string | null;

  mpesa_receipt_number?: string | null;

  provider_request_id?: string | null;

  provider_transaction_id?: string | null;

  failure_reason?: string | null;

  confirmed_at?: string | null;

  client_operation_id: string;

  sync_status:
    | "synced"
    | "pending"
    | "failed";

  server_updated_at?: string | null;

  created_at: string;
  updated_at: string;
}

/* =========================================================
   LOCAL RECEIPT
========================================================= */

export interface LocalReceipt {
  id: string;

  tenant_id: string;
  branch_id: string;

  sale_id: string;

  payment_id?: string | null;

  receipt_number: string;

  receipt_type: "sale";

  issued_at: string;

  printed_at?: string | null;

  reprint_count: number;

  sync_status:
    | "synced"
    | "pending"
    | "failed";

  server_updated_at?: string | null;

  created_at: string;
  updated_at: string;
}

/* =========================================================
   LOCAL RECEIPT REPRINT
========================================================= */

export interface LocalReceiptReprint {
  id: string;

  receipt_id: string;

  user_id?: string | null;
  terminal_id?: string | null;

  reason?: string | null;

  client_operation_id: string;

  sync_status:
    | "synced"
    | "pending"
    | "failed";

  server_updated_at?: string | null;

  created_at: string;
}

/* =========================================================
   SYNC METADATA
========================================================= */

export interface SyncMetadata {
  key: string;
  value: string;
  updated_at: string;
}

/* =========================================================
   DATABASE
========================================================= */

class CyberSaaSDatabase extends Dexie {
  offlineCredentials!: Table<
    OfflineCredential,
    string
  >;

  branches!: Table<
    LocalBranch,
    string
  >;

  customers!: Table<
    LocalCustomer,
    string
  >;

  services!: Table<
    LocalService,
    string
  >;

  sales!: Table<
    LocalSale,
    string
  >;

  saleItems!: Table<
    LocalSaleItem,
    string
  >;

  payments!: Table<
    LocalPayment,
    string
  >;

  receipts!: Table<
    LocalReceipt,
    string
  >;

  receiptReprints!: Table<
    LocalReceiptReprint,
    string
  >;

  syncQueue!: Table<
    SyncQueueItem,
    string
  >;

  syncMetadata!: Table<
    SyncMetadata,
    string
  >;

  constructor() {
    super("cybersaas-offline");

    /* =====================================================
       VERSION 1
    ===================================================== */

    this.version(1).stores({
      offlineCredentials:
        "id, user_id, tenant_id, email, role, branch_id, account_status",

      branches:
        "id, tenant_id, attendant_user_id, status",

      customers:
        "id, tenant_id, branch_id, customer_type, full_name, phone, sync_status, updated_at",

      syncQueue:
        "id, client_operation_id, tenant_id, user_id, branch_id, entity, operation, entity_id, status, created_at",

      syncMetadata:
        "key",
    });

    /* =====================================================
       VERSION 2

       client_operation_id becomes unique.
    ===================================================== */

    this.version(2).stores({
      offlineCredentials:
        "id, user_id, tenant_id, email, role, branch_id, account_status",

      branches:
        "id, tenant_id, attendant_user_id, status",

      customers:
        "id, tenant_id, branch_id, customer_type, full_name, phone, sync_status, updated_at",

      syncQueue:
        "id, &client_operation_id, tenant_id, user_id, branch_id, entity, operation, entity_id, status, created_at",

      syncMetadata:
        "key",
    });

    /* =====================================================
       VERSION 3

       Sales, payments, receipts and related tables.
    ===================================================== */

    this.version(3).stores({
      offlineCredentials:
        "id, user_id, tenant_id, branch_id, email",

      branches:
        "id, tenant_id, attendant_user_id",

      customers:
        "id, tenant_id, branch_id, sync_status",

      sales:
        "id, tenant_id, branch_id, customer_id, session_id, terminal_id, attendant_id, status, sync_status, client_operation_id",

      saleItems:
        "id, sale_id, service_id",

      payments:
        "id, tenant_id, branch_id, sale_id, status, method, client_operation_id",

      receipts:
        "id, tenant_id, branch_id, sale_id, receipt_number, sync_status",

      receiptReprints:
        "id, receipt_id, user_id, terminal_id, sync_status, client_operation_id",

      syncQueue:
        "id, client_operation_id, tenant_id, user_id, branch_id, entity, operation, entity_id, status, created_at",

      syncMetadata:
        "key",
    });

    /* =====================================================
       VERSION 4

       Service catalogue for offline operation.
    ===================================================== */

    this.version(4).stores({
      offlineCredentials:
        "id, user_id, tenant_id, branch_id, email",

      branches:
        "id, tenant_id, attendant_user_id",

      customers:
        "id, tenant_id, branch_id, sync_status",

      services:
        "id, tenant_id, branch_id, active, name",

      sales:
        "id, tenant_id, branch_id, customer_id, session_id, terminal_id, attendant_id, status, sync_status, client_operation_id",

      saleItems:
        "id, sale_id, service_id",

      payments:
        "id, tenant_id, branch_id, sale_id, status, method, client_operation_id",

      receipts:
        "id, tenant_id, branch_id, sale_id, receipt_number, sync_status",

      receiptReprints:
        "id, receipt_id, user_id, terminal_id, sync_status, client_operation_id",

      syncQueue:
        "id, client_operation_id, tenant_id, user_id, branch_id, entity, operation, entity_id, status, created_at",

      syncMetadata:
        "key",
    });
  }
}

/* =========================================================
   DATABASE INSTANCE
========================================================= */

export const db =
  new CyberSaaSDatabase();

/* =========================================================
   OFFLINE CREDENTIAL HELPERS
========================================================= */

export async function getOfflineCredential(): Promise<
  OfflineCredential | undefined
> {
  return db.offlineCredentials
    .toCollection()
    .first();
}

export async function saveOfflineCredential(
  credential: OfflineCredential
): Promise<void> {
  await db.transaction(
    "rw",
    db.offlineCredentials,
    db.branches,
    async () => {
      await db.offlineCredentials.clear();

      await db.offlineCredentials.put(
        credential
      );

      await db.branches.clear();
    }
  );
}

export async function disableOfflineCredential(
  userId: string
): Promise<void> {
  const credential =
    await db.offlineCredentials
      .where("user_id")
      .equals(userId)
      .first();

  if (!credential) {
    return;
  }

  await db.offlineCredentials.put({
    ...credential,

    account_status: "inactive",

    updated_at:
      new Date().toISOString(),
  });
}

export async function clearOfflineCredential(): Promise<void> {
  await db.transaction(
    "rw",
    db.offlineCredentials,
    db.branches,
    async () => {
      await db.offlineCredentials.clear();
      await db.branches.clear();
    }
  );
}

/* =========================================================
   BRANCH HELPERS
========================================================= */

export async function saveLocalBranch(
  branch: LocalBranch
): Promise<void> {
  await db.branches.put(branch);
}

export async function getLocalBranch(): Promise<
  LocalBranch | undefined
> {
  const credential =
    await getOfflineCredential();

  if (!credential) {
    return undefined;
  }

  return db.branches
    .where("attendant_user_id")
    .equals(credential.user_id)
    .first();
}

/* =========================================================
   SERVICE HELPERS
========================================================= */

/*
 * Save services to the local service catalogue.
 */
export async function saveLocalServices(
  services: LocalService[]
): Promise<void> {
  if (services.length === 0) {
    return;
  }

  await db.services.bulkPut(
    services
  );
}

/*
 * Get services belonging to a branch.
 *
 * activeOnly=true is used by the sales screen.
 */
export async function getLocalServices(
  branchId: string,
  activeOnly = true
): Promise<LocalService[]> {
  const services =
    await db.services
      .where("branch_id")
      .equals(branchId)
      .toArray();

  if (activeOnly) {
    return services.filter(
      (service) => service.active
    );
  }

  return services;
}

/*
 * Get one cached service.
 */
export async function getLocalService(
  serviceId: string
): Promise<
  LocalService | undefined
> {
  return db.services.get(
    serviceId
  );
}

/*
 * Replace the cached service catalogue
 * for a branch with the latest server catalogue.
 */
export async function replaceLocalServices(
  branchId: string,
  services: LocalService[]
): Promise<void> {
  await db.transaction(
    "rw",
    db.services,
    async () => {
      await db.services
        .where("branch_id")
        .equals(branchId)
        .delete();

      if (services.length > 0) {
        await db.services.bulkPut(
          services
        );
      }
    }
  );
}

/* =========================================================
   CUSTOMER HELPERS
========================================================= */

export async function saveLocalCustomer(
  customer: LocalCustomer
): Promise<void> {
  await db.customers.put(
    customer
  );
}

export async function getLocalCustomers(): Promise<
  LocalCustomer[]
> {
  const credential =
    await getOfflineCredential();

  if (!credential) {
    return [];
  }

  return db.customers
    .where("branch_id")
    .equals(credential.branch_id)
    .toArray();
}

export async function getLocalCustomer(
  customerId: string
): Promise<
  LocalCustomer | undefined
> {
  return db.customers.get(
    customerId
  );
}

/*
 * Create a local customer and its sync operation
 * atomically.
 */
export async function createLocalCustomer(
  customer: LocalCustomer,
  syncItem: SyncQueueItem
): Promise<void> {
  await db.transaction(
    "rw",
    db.customers,
    db.syncQueue,
    async () => {
      await db.customers.put(
        customer
      );

      await db.syncQueue.put(
        syncItem
      );
    }
  );
}

/*
 * Update a local customer and queue the update
 * atomically.
 */
export async function updateLocalCustomer(
  customer: LocalCustomer,
  syncItem: SyncQueueItem
): Promise<void> {
  await db.transaction(
    "rw",
    db.customers,
    db.syncQueue,
    async () => {
      await db.customers.put(
        customer
      );

      await db.syncQueue.put(
        syncItem
      );
    }
  );
}

/* =========================================================
   SYNC QUEUE HELPERS
========================================================= */

export async function addToSyncQueue(
  item: SyncQueueItem
): Promise<void> {
  await db.syncQueue.put(
    item
  );
}

export async function recoverProcessingSyncItems(): Promise<void> {
  const processing =
    await db.syncQueue
      .where("status")
      .equals("processing")
      .toArray();

  if (processing.length === 0) {
    return;
  }

  const now =
    new Date().toISOString();

  await db.transaction(
    "rw",
    db.syncQueue,
    async () => {
      for (const item of processing) {
        await db.syncQueue.put({
          ...item,

          status: "pending",

          last_error:
            item.last_error ??
            "Synchronization interrupted and will be retried.",

          updated_at: now,
        });
      }
    }
  );
}

export async function getPendingSyncItems(): Promise<
  SyncQueueItem[]
> {
  await recoverProcessingSyncItems();

  return db.syncQueue
    .where("status")
    .equals("pending")
    .sortBy("created_at");
}

export async function markSyncProcessing(
  id: string
): Promise<void> {
  const item =
    await db.syncQueue.get(id);

  if (!item) {
    return;
  }

  await db.syncQueue.put({
    ...item,

    status: "processing",

    attempts:
      item.attempts + 1,

    updated_at:
      new Date().toISOString(),
  });
}

export async function markSyncCompleted(
  id: string
): Promise<void> {
  await db.syncQueue.delete(
    id
  );
}

export async function markSyncFailed(
  id: string,
  error: string
): Promise<void> {
  const item =
    await db.syncQueue.get(id);

  if (!item) {
    return;
  }

  await db.syncQueue.put({
    ...item,

    status: "failed",

    last_error: error,

    updated_at:
      new Date().toISOString(),
  });
}

export async function retryFailedSyncItems(): Promise<void> {
  const failed =
    await db.syncQueue
      .where("status")
      .equals("failed")
      .toArray();

  if (failed.length === 0) {
    return;
  }

  const now =
    new Date().toISOString();

  await db.transaction(
    "rw",
    db.syncQueue,
    async () => {
      for (const item of failed) {
        await db.syncQueue.put({
          ...item,

          status: "pending",

          last_error: null,

          updated_at: now,
        });
      }
    }
  );
}

export async function getSyncItemByOperationId(
  clientOperationId: string
): Promise<
  SyncQueueItem | undefined
> {
  return db.syncQueue
    .where("client_operation_id")
    .equals(clientOperationId)
    .first();
}

/* =========================================================
   SYNC METADATA
========================================================= */

export async function setSyncMetadata(
  key: string,
  value: string
): Promise<void> {
  await db.syncMetadata.put({
    key,
    value,

    updated_at:
      new Date().toISOString(),
  });
}

export async function getSyncMetadata(
  key: string
): Promise<string | null> {
  const item =
    await db.syncMetadata.get(
      key
    );

  return item?.value ?? null;
}

/* =========================================================
   CLEAR OFFLINE DATABASE
========================================================= */

export async function clearOfflineDatabase(): Promise<void> {
  await db.offlineCredentials.clear();
  await db.branches.clear();
  await db.customers.clear();
  await db.services.clear();
  await db.sales.clear();
  await db.saleItems.clear();
  await db.payments.clear();
  await db.receipts.clear();
  await db.receiptReprints.clear();
  await db.syncQueue.clear();
  await db.syncMetadata.clear();
}