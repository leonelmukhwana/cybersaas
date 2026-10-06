"use client";

import {
  Suspense,
  useEffect,
  useState,
} from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  CheckCircle2,
  Printer,
  RefreshCw,
  Wifi,
  WifiOff,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

import { db } from "@/lib/offline/db";

import { useAuthStore } from "@/store/auth.store";
import type {
  PaymentResponse,
  SaleResponse,
} from "@/types/sale";
import { createOfflineReceipt } from "@/services/offline-sale.service";
import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

interface ReceiptData {
  id: string;
  receiptNumber: string;
  issuedAt: string;
  saleId: string;
  paymentId: string | null;
  total: string;
  items: {
    description: string;
    quantity: number;
    unitPrice: string;
    lineTotal: string;
  }[];
}

function ReceiptsPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const {
    user,
    isAuthenticated,
  } = useAuthStore();

  const saleId =
    searchParams.get("saleId");

  const paymentId =
    searchParams.get("paymentId");

  const [receipt, setReceipt] =
    useState<ReceiptData | null>(null);

  const [isOnline, setIsOnline] =
    useState(true);

  const [loading, setLoading] =
    useState(true);

  const [printing, setPrinting] =
    useState(false);

  const [error, setError] =
    useState("");

  useEffect(() => {
    setIsOnline(navigator.onLine);

    const handleOnline = () =>
      setIsOnline(true);

    const handleOffline = () =>
      setIsOnline(false);

    window.addEventListener(
      "online",
      handleOnline,
    );

    window.addEventListener(
      "offline",
      handleOffline,
    );

    return () => {
      window.removeEventListener(
        "online",
        handleOnline,
      );

      window.removeEventListener(
        "offline",
        handleOffline,
      );
    };
  }, []);

  useEffect(() => {
    const loadReceipt =
      async () => {
        if (!saleId || !paymentId) {
          setError(
            "Sale or payment information is missing.",
          );

          setLoading(false);
          return;
        }

        if (!user?.tenant_id) {
          setError(
            "Your account is missing a tenant ID.",
          );

          setLoading(false);
          return;
        }

        try {
          const localSale =
            await db.sales.get(
              saleId,
            );

          const localPayment =
            await db.payments.get(
              paymentId,
            );

          if (!localSale) {
            setError(
              "Sale could not be found.",
            );

            setLoading(false);
            return;
          }

          if (!localPayment) {
            setError(
              "Payment could not be found.",
            );

            setLoading(false);
            return;
          }

          if (
            localPayment.status !==
            "confirmed"
          ) {
            setError(
              "The payment is not confirmed yet. A receipt cannot be issued.",
            );

            setLoading(false);
            return;
          }

          const existingReceipt =
            await db.receipts
              .where("sale_id")
              .equals(saleId)
              .first();

          const items =
            await db.saleItems
              .where("sale_id")
              .equals(saleId)
              .toArray();

          const sale: SaleResponse = {
            id: localSale.id,
            tenant_id:
              localSale.tenant_id,
            branch_id:
              localSale.branch_id,
            customer_id:
              localSale.customer_id,
            session_id:
              localSale.session_id,
            session_started_at:
              localSale.session_started_at ?? null,
            terminal_id:
              localSale.terminal_id,
            attendant_id:
              localSale.attendant_id,
            status:
              localSale.status,
            subtotal:
              localSale.subtotal,
            discount_type:
              localSale.discount_type,
            discount_value:
              localSale.discount_value,
            discount_amount:
              localSale.discount_amount,
            total_amount:
              localSale.total_amount,
            currency:
              localSale.currency,
            created_at:
              localSale.created_at,
            updated_at:
              localSale.updated_at,
            items: items.map(
              (item) => ({
                id: item.id,
                sale_id:
                  item.sale_id,
                service_id:
                  item.service_id,
                description:
                  item.description,
                quantity:
                  Number(item.quantity),
                unit_price:
                  item.unit_price,
                line_total:
                  item.line_total,
              }),
            ),
          };

          const payment: PaymentResponse =
            {
              id: localPayment.id,
              tenant_id:
                localPayment.tenant_id,
              branch_id:
                localPayment.branch_id,
              sale_id:
                localPayment.sale_id,
              method:
                localPayment.method,
              status:
                localPayment.status,
              amount:
                localPayment.amount,
              phone:
                localPayment.phone,
              external_reference:
                localPayment.external_reference,
              mpesa_receipt_number:
                localPayment.mpesa_receipt_number,
              provider_request_id:
                localPayment.provider_request_id,
              provider_transaction_id:
                localPayment.provider_transaction_id,
              failure_reason:
                localPayment.failure_reason,
              confirmed_at:
                localPayment.confirmed_at,
              created_at:
                localPayment.created_at,
              updated_at:
                localPayment.updated_at,
            };

          let receiptData:
            ReceiptData;

          if (existingReceipt) {
            receiptData = {
              id:
                existingReceipt.id,
              receiptNumber:
                existingReceipt.receipt_number,
              issuedAt:
                existingReceipt.issued_at,
              saleId:
                existingReceipt.sale_id,
              paymentId:
                existingReceipt.payment_id ??
                null,
              total:
                sale.total_amount,
              items:
                sale.items.map(
                  (item) => ({
                    description:
                      item.description,
                    quantity:
                      item.quantity,
                    unitPrice:
                      item.unit_price,
                    lineTotal:
                      item.line_total,
                  }),
                ),
            };
          } else {
            const newReceipt =
              await createOfflineReceipt(
                sale,
                payment,
                user.tenant_id,
              );

            receiptData = {
              id:
                newReceipt.id,
              receiptNumber:
                newReceipt.receipt_number,
              issuedAt:
                newReceipt.issued_at,
              saleId:
                newReceipt.sale_id,
              paymentId:
                newReceipt.payment_id ??
                null,
              total:
                sale.total_amount,
              items:
                sale.items.map(
                  (item) => ({
                    description:
                      item.description,
                    quantity:
                      item.quantity,
                    unitPrice:
                      item.unit_price,
                    lineTotal:
                      item.line_total,
                  }),
                ),
            };
          }

          setReceipt(receiptData);
        } catch (err) {
          setError(
            err instanceof Error
              ? err.message
              : "Unable to create the receipt.",
          );
        } finally {
          setLoading(false);
        }
      };

    void loadReceipt();
  }, [
    saleId,
    paymentId,
    user?.tenant_id,
  ]);

  const handlePrint =
    async () => {
      if (!receipt) {
        return;
      }

      setPrinting(true);

      try {
        await db.receipts.update(
          receipt.id,
          {
            printed_at:
              new Date().toISOString(),
            updated_at:
              new Date().toISOString(),
          },
        );

        window.print();
      } finally {
        setPrinting(false);
      }
    };

  if (
    !isAuthenticated ||
    !user
  ) {
    return (
      <DashboardShell role="attendant">
        <div className="p-6">
          <Card>
            <CardContent className="py-10 text-center">
              <p className="text-sm text-slate-600">
                Please sign in to continue.
              </p>
            </CardContent>
          </Card>
        </div>
      </DashboardShell>
    );
  }

  if (loading) {
    return (
      <DashboardShell
        role={
          user.role === "owner"
            ? "owner"
            : "attendant"
        }
      >
        <div className="flex min-h-[400px] items-center justify-center">
          <RefreshCw className="h-6 w-6 animate-spin text-[#0757B8]" />
        </div>
      </DashboardShell>
    );
  }

  return (
    <DashboardShell
      role={
        user.role === "owner"
          ? "owner"
          : "attendant"
      }
    >
      <div className="space-y-6">
        <PageHeader
          title="Receipt"
          description="Review and print the customer receipt."
        />

        <div className="flex w-fit items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm">
          {isOnline ? (
            <>
              <Wifi className="h-4 w-4 text-green-600" />
              Online
            </>
          ) : (
            <>
              <WifiOff className="h-4 w-4 text-amber-600" />
              Offline mode
            </>
          )}
        </div>

        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {receipt && (
          <div className="mx-auto max-w-2xl">
            <Card className="border-slate-200 shadow-sm print:shadow-none">
              <CardHeader className="border-b border-slate-200 text-center">
                <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-green-50">
                  <CheckCircle2 className="h-6 w-6 text-green-600" />
                </div>

                <CardTitle>
                  Payment Complete
                </CardTitle>

                <p className="text-sm text-slate-500">
                  Receipt{" "}
                  {receipt.receiptNumber}
                </p>
              </CardHeader>

              <CardContent className="space-y-6 p-6">
                <div className="text-center">
                  <p className="text-xs uppercase tracking-wide text-slate-400">
                    Cyber Café Receipt
                  </p>

                  <p className="mt-1 text-sm text-slate-500">
                    {new Date(
                      receipt.issuedAt,
                    ).toLocaleString()}
                  </p>
                </div>

                <div className="space-y-3">
                  {receipt.items.map(
                    (item, index) => (
                      <div
                        key={`${item.description}-${index}`}
                        className="flex justify-between gap-4 border-b border-slate-100 pb-3"
                      >
                        <div>
                          <p className="font-medium text-slate-900">
                            {item.description}
                          </p>

                          <p className="text-sm text-slate-500">
                            {item.quantity} × KES{" "}
                            {Number(
                              item.unitPrice,
                            ).toFixed(2)}
                          </p>
                        </div>

                        <span className="font-medium">
                          KES{" "}
                          {Number(
                            item.lineTotal,
                          ).toFixed(2)}
                        </span>
                      </div>
                    ),
                  )}
                </div>

                <div className="border-t border-slate-200 pt-4">
                  <div className="flex justify-between">
                    <span className="text-lg font-semibold">
                      Total Paid
                    </span>

                    <span className="text-xl font-bold text-[#0757B8]">
                      KES{" "}
                      {Number(
                        receipt.total,
                      ).toFixed(2)}
                    </span>
                  </div>
                </div>

                {receipt.receiptNumber.startsWith(
                  "LOCAL-",
                ) && (
                  <div className="rounded-lg border border-amber-200 bg-amber-50 p-3 text-sm text-amber-700 print:hidden">
                    This is a temporary offline
                    receipt number. The official
                    receipt number will be assigned
                    when the transaction synchronizes.
                  </div>
                )}

                <div className="flex flex-col gap-3 sm:flex-row print:hidden">
                  <Button
                    type="button"
                    onClick={
                      handlePrint
                    }
                    disabled={
                      printing
                    }
                    className="flex-1 bg-[#0757B8] hover:bg-[#064A9D]"
                  >
                    {printing ? (
                      <>
                        <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                        Preparing...
                      </>
                    ) : (
                      <>
                        <Printer className="mr-2 h-4 w-4" />
                        Print Receipt
                      </>
                    )}
                  </Button>

                  <Button
                    type="button"
                    variant="outline"
                    onClick={() =>
                      router.push(
                        "/sales",
                      )
                    }
                    className="flex-1"
                  >
                    New Sale
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        )}
      </div>
    </DashboardShell>
  );
}

export default function ReceiptsPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-[400px] items-center justify-center text-sm text-slate-500">
          Loading receipt...
        </div>
      }
    >
      <ReceiptsPageContent />
    </Suspense>
  );
}