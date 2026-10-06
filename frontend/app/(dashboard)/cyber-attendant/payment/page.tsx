"use client";

import {
  Suspense,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  ArrowRight,
  Banknote,
  CheckCircle2,
  CreditCard,
  RefreshCw,
  Smartphone,
  Wifi,
  WifiOff,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

import { db } from "@/lib/offline/db";

import { useAuthStore } from "@/store/auth.store";
import type { SaleResponse } from "@/types/sale";
import { createOfflinePayment } from "@/services/offline-sale.service";
import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

type PaymentMethod =
  | "cash"
  | "mpesa"
  | "other";

function PaymentsPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const {
    user,
    isAuthenticated,
  } = useAuthStore();

  const saleId =
    searchParams.get("saleId");

  const [sale, setSale] =
    useState<SaleResponse | null>(null);

  const [method, setMethod] =
    useState<PaymentMethod>("cash");

  const [phone, setPhone] =
    useState("");

  const [externalReference, setExternalReference] =
    useState("");

  const [isOnline, setIsOnline] =
    useState(true);

  const [loading, setLoading] =
    useState(true);

  const [processing, setProcessing] =
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
    const loadSale = async () => {
      if (!saleId) {
        setError(
          "No sale was provided.",
        );

        setLoading(false);
        return;
      }

      try {
        const localSale =
          await db.sales.get(saleId);

        if (!localSale) {
          setError(
            "Sale could not be found on this device.",
          );

          setLoading(false);
          return;
        }

        const items =
          await db.saleItems
            .where("sale_id")
            .equals(saleId)
            .toArray();

        setSale({
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
        });
      } catch (err) {
        setError(
          err instanceof Error
            ? err.message
            : "Unable to load the sale.",
        );
      } finally {
        setLoading(false);
      }
    };

    void loadSale();
  }, [saleId]);

  const total = useMemo(
    () =>
      Number(
        sale?.total_amount ?? 0,
      ),
    [sale],
  );

  const handlePayment = async () => {
    setError("");

    if (!sale) {
      setError(
        "Sale is not available.",
      );
      return;
    }

    if (!user) {
      setError(
        "You are not authenticated.",
      );
      return;
    }

    if (!user.tenant_id) {
      setError(
        "Your account is missing a tenant ID.",
      );
      return;
    }

    if (
      method === "mpesa" &&
      !phone.trim()
    ) {
      setError(
        "Enter the customer's M-Pesa phone number.",
      );
      return;
    }

    setProcessing(true);

    try {
      const payment =
        await createOfflinePayment(
          {
            branch_id:
              sale.branch_id,

            sale_id:
              sale.id,

            method,

            amount:
              total.toFixed(2),

            phone:
              method === "mpesa"
                ? phone.trim()
                : null,

            external_reference:
              externalReference.trim() ||
              null,
          },
          user.tenant_id,
          user.id,
          method === "cash",
        );

      /*
       * Only confirmed payments can
       * proceed to receipt creation.
       */
      if (
        payment.status ===
        "confirmed"
      ) {
        router.push(
          `/cyber-attendant/receipt?saleId=${sale.id}&paymentId=${payment.id}`,
        );

        return;
      }

      /*
       * M-Pesa / other payments remain
       * pending until actual confirmation.
       */
      router.push(
        `/cyber-attendant/payment?saleId=${sale.id}&pendingPaymentId=${payment.id}`,
      );
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Unable to create payment.",
      );
    } finally {
      setProcessing(false);
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
          title="Payment"
          description="Select how the customer is paying for this sale."
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

        {!sale ? (
          <Card>
            <CardContent className="py-10 text-center text-sm text-slate-500">
              Sale not found.
            </CardContent>
          </Card>
        ) : (
          <div className="grid gap-6 lg:grid-cols-[1fr_380px]">
            <Card className="border-slate-200 shadow-sm">
              <CardHeader>
                <CardTitle>
                  Payment Method
                </CardTitle>
              </CardHeader>

              <CardContent className="space-y-4">
                <button
                  type="button"
                  onClick={() =>
                    setMethod("cash")
                  }
                  className={`flex w-full items-center gap-4 rounded-lg border p-4 text-left ${
                    method === "cash"
                      ? "border-[#0757B8] bg-blue-50"
                      : "border-slate-200"
                  }`}
                >
                  <Banknote className="h-6 w-6 text-[#0757B8]" />

                  <div>
                    <p className="font-medium text-slate-900">
                      Cash
                    </p>

                    <p className="text-sm text-slate-500">
                      Confirm cash immediately.
                    </p>
                  </div>
                </button>

                <button
                  type="button"
                  onClick={() =>
                    setMethod("mpesa")
                  }
                  className={`flex w-full items-center gap-4 rounded-lg border p-4 text-left ${
                    method === "mpesa"
                      ? "border-[#0757B8] bg-blue-50"
                      : "border-slate-200"
                  }`}
                >
                  <Smartphone className="h-6 w-6 text-[#0757B8]" />

                  <div>
                    <p className="font-medium text-slate-900">
                      M-Pesa
                    </p>

                    <p className="text-sm text-slate-500">
                      Payment remains pending until confirmed.
                    </p>
                  </div>
                </button>

                <button
                  type="button"
                  onClick={() =>
                    setMethod("other")
                  }
                  className={`flex w-full items-center gap-4 rounded-lg border p-4 text-left ${
                    method === "other"
                      ? "border-[#0757B8] bg-blue-50"
                      : "border-slate-200"
                  }`}
                >
                  <CreditCard className="h-6 w-6 text-[#0757B8]" />

                  <div>
                    <p className="font-medium text-slate-900">
                      Other
                    </p>

                    <p className="text-sm text-slate-500">
                      Record another payment method.
                    </p>
                  </div>
                </button>

                {method === "mpesa" && (
                  <div className="space-y-4 border-t border-slate-200 pt-4">
                    <div className="space-y-2">
                      <Label htmlFor="phone">
                        M-Pesa Phone Number
                      </Label>

                      <Input
                        id="phone"
                        placeholder="07XXXXXXXX"
                        value={phone}
                        onChange={(e) =>
                          setPhone(
                            e.target.value,
                          )
                        }
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="reference">
                        External Reference
                      </Label>

                      <Input
                        id="reference"
                        placeholder="Payment reference"
                        value={
                          externalReference
                        }
                        onChange={(e) =>
                          setExternalReference(
                            e.target.value,
                          )
                        }
                      />
                    </div>
                  </div>
                )}

                {method === "other" && (
                  <div className="space-y-2 border-t border-slate-200 pt-4">
                    <Label htmlFor="otherReference">
                      Reference
                    </Label>

                    <Input
                      id="otherReference"
                      placeholder="Payment reference"
                      value={
                        externalReference
                      }
                      onChange={(e) =>
                        setExternalReference(
                          e.target.value,
                        )
                      }
                    />
                  </div>
                )}

                <Button
                  type="button"
                  disabled={processing}
                  onClick={handlePayment}
                  className="w-full bg-[#0757B8] hover:bg-[#064A9D]"
                >
                  {processing ? (
                    <>
                      <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                      Processing...
                    </>
                  ) : (
                    <>
                      Complete Payment
                      <ArrowRight className="ml-2 h-4 w-4" />
                    </>
                  )}
                </Button>
              </CardContent>
            </Card>

            <Card className="h-fit border-slate-200 shadow-sm">
              <CardHeader>
                <CardTitle>
                  Sale Summary
                </CardTitle>
              </CardHeader>

              <CardContent className="space-y-4">
                {sale.items?.map(
                  (item) => (
                    <div
                      key={item.id}
                      className="flex justify-between gap-4 text-sm"
                    >
                      <div>
                        <p className="font-medium text-slate-900">
                          {item.description}
                        </p>

                        <p className="text-slate-500">
                          {item.quantity} × KES{" "}
                          {Number(
                            item.unit_price,
                          ).toFixed(2)}
                        </p>
                      </div>

                      <span className="font-medium">
                        KES{" "}
                        {Number(
                          item.line_total,
                        ).toFixed(2)}
                      </span>
                    </div>
                  ),
                )}

                <div className="space-y-2 border-t border-slate-200 pt-4">
                  <div className="flex justify-between text-sm">
                    <span className="text-slate-500">
                      Subtotal
                    </span>

                    <span>
                      KES{" "}
                      {Number(
                        sale.subtotal,
                      ).toFixed(2)}
                    </span>
                  </div>

                  <div className="flex justify-between text-sm">
                    <span className="text-slate-500">
                      Discount
                    </span>

                    <span>
                      KES{" "}
                      {Number(
                        sale.discount_amount,
                      ).toFixed(2)}
                    </span>
                  </div>

                  <div className="flex justify-between border-t border-slate-200 pt-3">
                    <span className="font-semibold">
                      Amount Due
                    </span>

                    <span className="text-xl font-bold text-[#0757B8]">
                      KES{" "}
                      {total.toFixed(2)}
                    </span>
                  </div>
                </div>

                {method === "cash" && (
                  <div className="rounded-lg border border-green-200 bg-green-50 p-3 text-sm text-green-700">
                    <div className="flex gap-2">
                      <CheckCircle2 className="h-4 w-4 shrink-0" />
                      Cash can be confirmed locally even when offline.
                    </div>
                  </div>
                )}

                {method === "mpesa" && (
                  <div className="rounded-lg border border-amber-200 bg-amber-50 p-3 text-sm text-amber-700">
                    M-Pesa must remain pending until the actual payment is confirmed.
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        )}
      </div>
    </DashboardShell>
  );
}

export default function PaymentsPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-[400px] items-center justify-center text-sm text-slate-500">
          Loading payment...
        </div>
      }
    >
      <PaymentsPageContent />
    </Suspense>
  );
}