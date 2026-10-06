

import { createLocalCustomer, getLocalCustomer, getLocalCustomers, LocalCustomer, SyncQueueItem, updateLocalCustomer } from "@/lib/offline/db";
import type {
  CreateCustomerRequest,
  Customer,
  UpdateCustomerRequest,
} from "@/types/customer";

/* =========================================================
   HELPERS
========================================================= */

function localCustomerToCustomer(
  customer: LocalCustomer
): Customer {
  return {
    id: customer.id,
    tenant_id: customer.tenant_id,
    branch_id: customer.branch_id,
    customer_type: customer.customer_type,
    full_name: customer.full_name,
    phone: customer.phone ?? null,
    parent_name: customer.parent_name ?? null,
    parent_phone: customer.parent_phone ?? null,
    created_at: customer.created_at,
    updated_at: customer.updated_at,
  };
}

/* =========================================================
   CREATE CUSTOMER OFFLINE
========================================================= */

export async function createOfflineCustomer(
  input: CreateCustomerRequest,
  tenantId: string,
  userId: string
): Promise<Customer> {
  if (!tenantId) {
    throw new Error(
      "Tenant information is required."
    );
  }

  if (!userId) {
    throw new Error(
      "Attendant information is required."
    );
  }

  if (!input.branch_id) {
    throw new Error(
      "Assigned branch is required."
    );
  }

  if (!input.full_name.trim()) {
    throw new Error(
      "Customer name is required."
    );
  }

  if (
    input.customer_type === "child" &&
    (!input.parent_name?.trim() ||
      !input.parent_phone?.trim())
  ) {
    throw new Error(
      "Parent or guardian name and phone are required for a child."
    );
  }

  const now = new Date().toISOString();

  /*
   * The customer gets one permanent local UUID.
   */
  const customerId = crypto.randomUUID();

  /*
   * IMPORTANT:
   *
   * This operation ID identifies this exact operation.
   *
   * If synchronization fails and the operation is retried,
   * the sync worker MUST reuse this same ID.
   *
   * It must never generate another ID for a retry.
   */
  const clientOperationId =
    crypto.randomUUID();

  /*
   * Sensitive ID values are NOT stored in the
   * LocalCustomer cache.
   */
  const customer: LocalCustomer = {
    id: customerId,

    tenant_id: tenantId,
    branch_id: input.branch_id,

    customer_type: input.customer_type,

    full_name: input.full_name.trim(),

    phone:
      input.phone?.trim() || null,

    parent_name:
      input.parent_name?.trim() || null,

    parent_phone:
      input.parent_phone?.trim() || null,

    sync_status: "pending",

    server_updated_at: null,

    created_at: now,
    updated_at: now,
  };

  /*
   * The queue payload contains the complete request
   * required by the backend.
   *
   * These values are NOT copied into LocalCustomer.
   */
  const payload: Record<string, unknown> = {
    branch_id: input.branch_id,

    customer_type:
      input.customer_type,

    full_name:
      input.full_name.trim(),

    phone:
      input.phone?.trim() || null,

    id_number:
      input.id_number?.trim() || null,

    parent_name:
      input.parent_name?.trim() || null,

    parent_phone:
      input.parent_phone?.trim() || null,

    parent_id_number:
      input.parent_id_number?.trim() || null,
  };

  const syncItem: SyncQueueItem = {
    id: crypto.randomUUID(),

    /*
     * NEVER change this value during retry.
     */
    client_operation_id:
      clientOperationId,

    tenant_id: tenantId,
    user_id: userId,
    branch_id: input.branch_id,

    entity: "customer",
    operation: "create",

    entity_id: customerId,

    payload,

    status: "pending",

    attempts: 0,

    last_error: null,

    created_at: now,
    updated_at: now,
  };

  /*
   * Customer and sync queue are saved atomically.
   *
   * This prevents:
   *
   * customer saved + queue missing
   *
   * or
   *
   * queue saved + customer missing
   */
  await createLocalCustomer(
    customer,
    syncItem
  );

  return localCustomerToCustomer(
    customer
  );
}

/* =========================================================
   UPDATE CUSTOMER OFFLINE
========================================================= */

export async function updateOfflineCustomer(
  customerId: string,
  input: UpdateCustomerRequest,
  tenantId: string,
  userId: string
): Promise<Customer> {
  if (!tenantId) {
    throw new Error(
      "Tenant information is required."
    );
  }

  if (!userId) {
    throw new Error(
      "Attendant information is required."
    );
  }

  if (!customerId) {
    throw new Error(
      "Customer ID is required."
    );
  }

  if (!input.full_name.trim()) {
    throw new Error(
      "Customer name is required."
    );
  }

  if (
    input.customer_type === "child" &&
    (!input.parent_name?.trim() ||
      !input.parent_phone?.trim())
  ) {
    throw new Error(
      "Parent or guardian name and phone are required for a child."
    );
  }

  const existing =
    await getLocalCustomer(
      customerId
    );

  if (!existing) {
    throw new Error(
      "Customer is not available offline."
    );
  }

  if (existing.tenant_id !== tenantId) {
    throw new Error(
      "Customer does not belong to this tenant."
    );
  }

  /*
   * The customer already belongs to the attendant's
   * assigned branch because it was loaded through the
   * local branch-scoped customer cache.
   *
   * We therefore preserve the existing branch ID.
   */
  const now = new Date().toISOString();

  /*
   * This identifies THIS update operation.
   *
   * A retry of this operation must reuse this exact ID.
   */
  const clientOperationId =
    crypto.randomUUID();

  const updatedCustomer: LocalCustomer = {
    ...existing,

    customer_type:
      input.customer_type,

    full_name:
      input.full_name.trim(),

    phone:
      input.phone?.trim() || null,

    parent_name:
      input.parent_name?.trim() || null,

    parent_phone:
      input.parent_phone?.trim() || null,

    sync_status: "pending",

    updated_at: now,
  };

  const payload: Record<string, unknown> = {
    customer_type:
      input.customer_type,

    full_name:
      input.full_name.trim(),

    phone:
      input.phone?.trim() || null,

    id_number:
      input.id_number?.trim() || null,

    parent_name:
      input.parent_name?.trim() || null,

    parent_phone:
      input.parent_phone?.trim() || null,

    parent_id_number:
      input.parent_id_number?.trim() || null,
  };

  const syncItem: SyncQueueItem = {
    id: crypto.randomUUID(),

    /*
     * NEVER generate a new ID when retrying this
     * synchronization operation.
     */
    client_operation_id:
      clientOperationId,

    tenant_id: tenantId,

    user_id: userId,

    branch_id:
      existing.branch_id,

    entity: "customer",

    operation: "update",

    entity_id: customerId,

    payload,

    status: "pending",

    attempts: 0,

    last_error: null,

    created_at: now,

    updated_at: now,
  };

  /*
   * Update + queue insertion happen atomically.
   */
  await updateLocalCustomer(
    updatedCustomer,
    syncItem
  );

  return localCustomerToCustomer(
    updatedCustomer
  );
}

/* =========================================================
   LIST OFFLINE CUSTOMERS
========================================================= */

export async function listOfflineCustomers(
  search?: string
): Promise<Customer[]> {
  const customers =
    await getLocalCustomers();

  const normalizedSearch =
    search?.trim().toLowerCase() || "";

  const filteredCustomers =
    normalizedSearch.length === 0
      ? customers
      : customers.filter(
          (customer) => {
            const nameMatches =
              customer.full_name
                .toLowerCase()
                .includes(
                  normalizedSearch
                );

            const phoneMatches =
              customer.phone
                ?.toLowerCase()
                .includes(
                  normalizedSearch
                ) ?? false;

            return (
              nameMatches ||
              phoneMatches
            );
          }
        );

  return filteredCustomers
    .sort((a, b) =>
      b.updated_at.localeCompare(
        a.updated_at
      )
    )
    .map(localCustomerToCustomer);
}