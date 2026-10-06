import { db } from "@/lib/offline/db";
import type {
  CreatePaymentInput,
  CreateSaleInput,
  PaymentResponse,
  ReceiptResponse,
  SaleResponse,
} from "@/types/sale";

function generateId(): string {
  return crypto.randomUUID();
}

function generateClientOperationId(): string {
  return crypto.randomUUID();
}

function calculateLineTotal(
  quantity: number,
  unitPrice: string,
): string {
  const price = Number(unitPrice);

  if (!Number.isFinite(price) || !Number.isFinite(quantity)) {
    return "0.00";
  }

  return (quantity * price).toFixed(2);
}

function calculateSubtotal(
  items: CreateSaleInput["items"],
): string {
  return items
    .reduce(
      (total, item) =>
        total +
        Number(
          calculateLineTotal(
            item.quantity,
            item.unit_price,
          ),
        ),
      0,
    )
    .toFixed(2);
}

function calculateDiscount(
  subtotal: string,
  discountType?: string | null,
  discountValue = "0",
): string {
  const amount = Number(subtotal);
  const value = Number(discountValue);

  if (
    !Number.isFinite(amount) ||
    !Number.isFinite(value)
  ) {
    return "0.00";
  }

  if (!discountType || value <= 0) {
    return "0.00";
  }

  if (discountType === "percentage") {
    return ((amount * value) / 100).toFixed(2);
  }

  return Math.min(value, amount).toFixed(2);
}

function calculateTotal(
  subtotal: string,
  discountAmount: string,
): string {
  return Math.max(
    0,
    Number(subtotal) - Number(discountAmount),
  ).toFixed(2);
}

/**
 * Create a sale locally.
 *
 * The sale is stored in IndexedDB and added to the
 * synchronization queue.
 */
export async function createOfflineSale(
  input: CreateSaleInput,
  tenantId: string,
  userId: string,
): Promise<SaleResponse> {
  const now = new Date().toISOString();

  const saleId = generateId();
  const clientOperationId =
    generateClientOperationId();

  const subtotal = calculateSubtotal(input.items);

  const discountAmount = calculateDiscount(
    subtotal,
    input.discount_type,
    input.discount_value,
  );

  const totalAmount = calculateTotal(
    subtotal,
    discountAmount,
  );

  const sale: SaleResponse = {
    id: saleId,
    tenant_id: tenantId,
    branch_id: input.branch_id,

    customer_id: input.customer_id ?? null,
    session_id: input.session_id ?? null,
    session_started_at:
      input.session_started_at ?? null,
    terminal_id: input.terminal_id ?? null,
    attendant_id:
      input.attendant_id ?? userId,

    status: "completed",

    subtotal,
    discount_type:
      input.discount_type ?? null,
    discount_value:
      input.discount_value || "0.00",
    discount_amount: discountAmount,
    total_amount: totalAmount,
    currency: "KES",

    created_at: now,
    updated_at: now,

    items: input.items.map((item) => ({
      id: generateId(),
      sale_id: saleId,
      service_id: item.service_id ?? null,
      description: item.description,
      quantity: item.quantity,
      unit_price: item.unit_price,
      line_total: calculateLineTotal(
        item.quantity,
        item.unit_price,
      ),
    })),
  };

  /*
   * Save sale locally.
   */
  await db.sales.put({
    id: sale.id,
    tenant_id: sale.tenant_id,
    branch_id: sale.branch_id,

    customer_id: sale.customer_id,
    session_id: sale.session_id,
    session_started_at:
      sale.session_started_at,
    terminal_id: sale.terminal_id,
    attendant_id: sale.attendant_id,

    subtotal: sale.subtotal,
    discount_type: sale.discount_type,
    discount_value: sale.discount_value,
    discount_amount: sale.discount_amount,

    total_amount: sale.total_amount,
    currency: "KES",

    status: sale.status,

    client_operation_id:
      clientOperationId,

    sync_status: "pending",

    created_at: sale.created_at,
    updated_at: sale.updated_at,
  });

  /*
   * Save sale items locally.
   */
  await db.saleItems.bulkPut(
    sale.items.map((item) => ({
      id: item.id,
      sale_id: item.sale_id,
      service_id: item.service_id,

      description: item.description,

      quantity: String(item.quantity),

      unit_price: item.unit_price,
      line_total: item.line_total,

      created_at: now,
    })),
  );

  /*
   * Add sale to synchronization queue.
   */
  await db.syncQueue.put({
    id: generateId(),

    client_operation_id:
      clientOperationId,

    tenant_id: tenantId,
    user_id: userId,
    branch_id: input.branch_id,

    entity: "sale",
    operation: "create",

    entity_id: saleId,

    payload: {
      branch_id: input.branch_id,

      customer_id:
        input.customer_id ?? null,

      session_id:
        input.session_id ?? null,

      session_started_at:
        input.session_started_at ?? null,

      terminal_id:
        input.terminal_id ?? null,

      attendant_id:
        input.attendant_id ?? userId,

      discount_type:
        input.discount_type ?? null,

      discount_value:
        input.discount_value || "0.00",

      client_operation_id:
        clientOperationId,

      items: sale.items.map((item) => ({
        service_id: item.service_id,

        description:
          item.description,

        quantity: item.quantity,

        unit_price:
          item.unit_price,
      })),
    },

    status: "pending",

    attempts: 0,

    last_error: null,

    created_at: now,
    updated_at: now,
  });

  return sale;
}

/**
 * Create a payment locally.
 */
export async function createOfflinePayment(
  input: CreatePaymentInput,
  tenantId: string,
  userId: string,
  confirmed = false,
): Promise<PaymentResponse> {
  const now = new Date().toISOString();

  const paymentId = generateId();

  const clientOperationId =
    generateClientOperationId();

  const payment: PaymentResponse = {
    id: paymentId,

    tenant_id: tenantId,
    branch_id: input.branch_id,
    sale_id: input.sale_id,

    method: input.method,

    status: confirmed
      ? "confirmed"
      : "pending",

    amount: input.amount,

    phone:
      input.phone ?? null,

    external_reference:
      input.external_reference ?? null,

    mpesa_receipt_number: null,

    provider_request_id: null,

    provider_transaction_id: null,

    failure_reason: null,

    confirmed_at:
      confirmed ? now : null,

    created_at: now,
    updated_at: now,
  };

  /*
   * Save payment locally.
   */
  await db.payments.put({
    id: payment.id,

    tenant_id: payment.tenant_id,
    branch_id: payment.branch_id,
    sale_id: payment.sale_id,

    method: payment.method,
    status: payment.status,

    amount: payment.amount,

    phone:
      input.phone ?? null,

    external_reference:
      input.external_reference ?? null,

    mpesa_receipt_number: null,

    provider_request_id: null,

    provider_transaction_id: null,

    failure_reason: null,

    confirmed_at:
      payment.confirmed_at ?? null,

    client_operation_id:
      clientOperationId,

    sync_status: "pending",

    created_at: now,
    updated_at: now,
  });

  /*
   * Queue payment creation.
   */
  await db.syncQueue.put({
    id: generateId(),

    client_operation_id:
      clientOperationId,

    tenant_id: tenantId,
    user_id: userId,
    branch_id: input.branch_id,

    entity: "payment",
    operation: "create",

    entity_id: paymentId,

    payload: {
      branch_id: input.branch_id,

      sale_id: input.sale_id,

      method: input.method,

      amount: input.amount,

      phone:
        input.phone ?? null,

      external_reference:
        input.external_reference ?? null,

      client_operation_id:
        clientOperationId,
    },

    status: "pending",

    attempts: 0,

    last_error: null,

    created_at: now,
    updated_at: now,
  });

  /*
   * Cash can be confirmed locally because
   * the attendant has physically received it.
   *
   * M-Pesa and other methods remain pending
   * until there is actual confirmation.
   */
  if (confirmed) {
    await db.syncQueue.put({
      id: generateId(),

      client_operation_id:
        generateClientOperationId(),

      tenant_id: tenantId,
      user_id: userId,
      branch_id: input.branch_id,

      entity: "payment",
      operation: "confirm",

      entity_id: paymentId,

      payload: {},

      status: "pending",

      attempts: 0,

      last_error: null,

      created_at: now,
      updated_at: now,
    });
  }

  return payment;
}

/**
 * Create a local receipt.
 *
 * The LOCAL receipt number is temporary.
 * The server will provide the official receipt number
 * after synchronization.
 */
export async function createOfflineReceipt(
  sale: SaleResponse,
  payment: PaymentResponse,
  tenantId: string,
): Promise<ReceiptResponse> {
  const now = new Date().toISOString();

  const receiptId = generateId();

  const receipt: ReceiptResponse = {
    id: receiptId,

    tenant_id: tenantId,
    branch_id: sale.branch_id,

    sale_id: sale.id,

    payment_id: payment.id,

    receipt_number:
      `LOCAL-${sale.id}`,

    receipt_type: "sale",

    issued_at: now,

    printed_at: null,

    reprint_count: 0,

    created_at: now,
    updated_at: now,
  };

  /*
   * Save receipt locally.
   */
  await db.receipts.put({
    id: receipt.id,

    tenant_id: receipt.tenant_id,
    branch_id: receipt.branch_id,

    sale_id: receipt.sale_id,

    payment_id:
      receipt.payment_id,

    receipt_number:
      receipt.receipt_number,

    receipt_type:
      receipt.receipt_type,

    issued_at:
      receipt.issued_at,

    printed_at: null,

    reprint_count: 0,

    sync_status: "pending",

    created_at:
      receipt.created_at,

    updated_at:
      receipt.updated_at,
  });

  return receipt;
}

/**
 * Complete the complete local transaction:
 *
 * Sale
 *   ↓
 * Payment
 *   ↓
 * Receipt
 */
export async function completeOfflineSale(
  saleInput: CreateSaleInput,
  paymentInput: CreatePaymentInput,
  tenantId: string,
  userId: string,
): Promise<{
  sale: SaleResponse;
  payment: PaymentResponse;
  receipt: ReceiptResponse;
}> {
  const sale =
    await createOfflineSale(
      saleInput,
      tenantId,
      userId,
    );

  const payment =
    await createOfflinePayment(
      {
        ...paymentInput,
        sale_id: sale.id,
      },
      tenantId,
      userId,
      paymentInput.method === "cash",
    );

  const receipt =
    await createOfflineReceipt(
      sale,
      payment,
      tenantId,
    );

  return {
    sale,
    payment,
    receipt,
  };
}