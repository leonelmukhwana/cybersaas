
"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

import { useAuthStore } from "@/store/auth.store";
import { subscriptionService } from "@/services/subscription.service";

import type { SubscriptionPlan } from "@/types/subscription";

type PaymentMethod = "stk" | "paybill";

export default function PaymentPage() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const planId = searchParams.get("plan");

  const token = useAuthStore((state) => state.token);

  const [plan, setPlan] = useState<SubscriptionPlan | null>(null);

  const [paymentMethod, setPaymentMethod] =
    useState<PaymentMethod>("stk");

  const [phoneNumber, setPhoneNumber] = useState("");
  const [mpesaReference, setMpesaReference] = useState("");

  const [loading, setLoading] = useState(true);
  const [paying, setPaying] = useState(false);

  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  /*
   * Load the selected package.
   */
  useEffect(() => {
    if (!token) {
      router.replace("/login/owner");
      return;
    }

    if (!planId) {
      setError("No subscription package was selected.");
      setLoading(false);
      return;
    }

    const currentToken = token;

    async function loadPlan() {
      try {
        setLoading(true);
        setError("");

        const plans =
          await subscriptionService.getPlans(currentToken);

        const selectedPlan = plans.find(
          (item) => item.id === planId
        );

        if (!selectedPlan) {
          setError(
            "The selected subscription package could not be found."
          );
          return;
        }

        setPlan(selectedPlan);
      } catch (err) {
        console.error("LOAD PAYMENT PLAN ERROR:", err);

        setError(
          err instanceof Error
            ? err.message
            : "Unable to load the selected package."
        );
      } finally {
        setLoading(false);
      }
    }

    loadPlan();
  }, [token, planId, router]);

  const formattedAmount = useMemo(() => {
    if (!plan) {
      return "";
    }

    const amount = Number(plan.monthly_price);

    if (Number.isNaN(amount)) {
      return `KES ${plan.monthly_price}`;
    }

    return new Intl.NumberFormat("en-KE", {
      style: "currency",
      currency: "KES",
      maximumFractionDigits: 0,
    }).format(amount);
  }, [plan]);

  function handlePhoneChange(
    event: React.ChangeEvent<HTMLInputElement>
  ) {
    const value = event.target.value;

    setPhoneNumber(value);
    setError("");
    setSuccess("");
  }

  function handleReferenceChange(
    event: React.ChangeEvent<HTMLInputElement>
  ) {
    const value = event.target.value.toUpperCase();

    setMpesaReference(value);
    setError("");
    setSuccess("");
  }

  async function handleStkPayment() {
    if (!phoneNumber.trim()) {
      setError("Enter the M-Pesa phone number.");
      return;
    }

    /*
     * Payment API will be connected here after we confirm
     * the CreatePaymentRequest used by your backend.
     */
    setPaying(true);
    setError("");
    setSuccess("");

    try {
      console.log("STK PAYMENT:", {
        plan_id: planId,
        phone_number: phoneNumber.trim(),
        amount: plan?.monthly_price,
      });

      /*
       * TEMPORARY UI FLOW
       *
       * Replace this section with the real API call once
       * we match the backend CreatePaymentRequest.
       */

      await new Promise((resolve) =>
        setTimeout(resolve, 800)
      );

      setSuccess(
        "Payment request is ready. We will connect the M-Pesa STK Push next."
      );
    } catch (err) {
      console.error("STK PAYMENT ERROR:", err);

      setError(
        err instanceof Error
          ? err.message
          : "Unable to initiate payment."
      );
    } finally {
      setPaying(false);
    }
  }

  async function handleManualPayment() {
    if (!mpesaReference.trim()) {
      setError("Enter your M-Pesa transaction code.");
      return;
    }

    setPaying(true);
    setError("");
    setSuccess("");

    try {
      console.log("MANUAL PAYMENT:", {
        plan_id: planId,
        reference: mpesaReference.trim(),
        amount: plan?.monthly_price,
      });

      /*
       * Manual payment API will be connected after we confirm
       * the backend request structure.
       */

      await new Promise((resolve) =>
        setTimeout(resolve, 800)
      );

      setSuccess(
        "Your payment reference has been captured. Verification will be connected next."
      );
    } catch (err) {
      console.error("MANUAL PAYMENT ERROR:", err);

      setError(
        err instanceof Error
          ? err.message
          : "Unable to submit payment."
      );
    } finally {
      setPaying(false);
    }
  }

  if (loading) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-50">
        <p className="text-sm text-slate-500">
          Loading payment...
        </p>
      </main>
    );
  }

  if (error && !plan) {
    return (
      <main className="flex min-h-screen items-center justify-center bg-slate-50 px-4">
        <div className="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-8 text-center shadow-sm">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-[#0757B8] text-xl font-bold text-white">
            C
          </div>

          <h1 className="mt-5 text-xl font-bold text-slate-950">
            Payment unavailable
          </h1>

          <p className="mt-2 text-sm text-slate-500">
            {error}
          </p>

          <Button
            type="button"
            onClick={() =>
              router.push(
                "/dashboard/owner/subscription"
              )
            }
            className="mt-6 w-full bg-[#0757B8] hover:bg-[#064A9D]"
          >
            Back to Packages
          </Button>
        </div>
      </main>
    );
  }

  if (!plan) {
    return null;
  }

  return (
    <main className="min-h-screen bg-slate-50 px-4 py-8 md:py-12">
      <div className="mx-auto max-w-5xl">
        {/* TOP */}
        <div className="mb-8">
          <Link
            href="/dashboard/owner/subscription"
            className="text-sm font-medium text-[#0757B8] hover:underline"
          >
            ← Back to packages
          </Link>

          <div className="mt-6">
            <h1 className="text-3xl font-bold tracking-tight text-slate-950">
              Complete Payment
            </h1>

            <p className="mt-2 text-sm text-slate-500">
              Complete your CyberSaaS subscription payment.
            </p>
          </div>
        </div>

        <div className="grid gap-6 lg:grid-cols-[320px_1fr]">
          {/* PACKAGE SUMMARY */}
          <section className="h-fit rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
            <p className="text-sm font-medium text-slate-500">
              Selected package
            </p>

            <h2 className="mt-2 text-2xl font-bold text-slate-950">
              {plan.name}
            </h2>

            <div className="mt-6 rounded-xl bg-slate-50 p-5">
              <p className="text-sm text-slate-500">
                Amount to pay
              </p>

              <p className="mt-1 text-3xl font-bold text-[#0757B8]">
                {formattedAmount}
              </p>

              <p className="mt-1 text-sm text-slate-500">
                {plan.is_lifetime
                  ? "One-time payment"
                  : "Monthly subscription"}
              </p>
            </div>

            <div className="mt-6 space-y-4">
              <div className="flex justify-between text-sm">
                <span className="text-slate-500">
                  Branches
                </span>

                <span className="font-semibold text-slate-900">
                  {plan.included_branches}
                </span>
              </div>

              <div className="flex justify-between text-sm">
                <span className="text-slate-500">
                  Terminals
                </span>

                <span className="font-semibold text-slate-900">
                  {plan.included_terminals}
                </span>
              </div>
            </div>

            <div className="mt-6 border-t border-slate-100 pt-5">
              <p className="text-xs leading-5 text-slate-500">
                The amount shown here comes directly from the
                selected subscription package.
              </p>
            </div>
          </section>

          {/* PAYMENT CARD */}
          <section className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm md:p-8">
            <div>
              <h2 className="text-xl font-bold text-slate-950">
                M-Pesa Payment
              </h2>

              <p className="mt-1 text-sm text-slate-500">
                Choose how you want to pay.
              </p>
            </div>

            {/* METHODS */}
            <div className="mt-6 grid gap-3 md:grid-cols-2">
              <button
                type="button"
                onClick={() => {
                  setPaymentMethod("stk");
                  setError("");
                  setSuccess("");
                }}
                className={`rounded-xl border p-5 text-left transition ${
                  paymentMethod === "stk"
                    ? "border-[#0757B8] bg-blue-50 ring-1 ring-[#0757B8]"
                    : "border-slate-200 hover:border-slate-300"
                }`}
              >
                <div className="flex items-center gap-3">
                  <span
                    className={`flex h-5 w-5 items-center justify-center rounded-full border ${
                      paymentMethod === "stk"
                        ? "border-[#0757B8]"
                        : "border-slate-300"
                    }`}
                  >
                    {paymentMethod === "stk" && (
                      <span className="h-2.5 w-2.5 rounded-full bg-[#0757B8]" />
                    )}
                  </span>

                  <span className="font-semibold text-slate-900">
                    M-Pesa Express
                  </span>
                </div>

                <p className="mt-3 pl-8 text-xs text-slate-500">
                  Enter your phone number and receive an
                  M-Pesa payment prompt.
                </p>
              </button>

              <button
                type="button"
                onClick={() => {
                  setPaymentMethod("paybill");
                  setError("");
                  setSuccess("");
                }}
                className={`rounded-xl border p-5 text-left transition ${
                  paymentMethod === "paybill"
                    ? "border-[#0757B8] bg-blue-50 ring-1 ring-[#0757B8]"
                    : "border-slate-200 hover:border-slate-300"
                }`}
              >
                <div className="flex items-center gap-3">
                  <span
                    className={`flex h-5 w-5 items-center justify-center rounded-full border ${
                      paymentMethod === "paybill"
                        ? "border-[#0757B8]"
                        : "border-slate-300"
                    }`}
                  >
                    {paymentMethod === "paybill" && (
                      <span className="h-2.5 w-2.5 rounded-full bg-[#0757B8]" />
                    )}
                  </span>

                  <span className="font-semibold text-slate-900">
                    Paybill
                  </span>
                </div>

                <p className="mt-3 pl-8 text-xs text-slate-500">
                  Make the payment manually using M-Pesa.
                </p>
              </button>
            </div>

            {/* ALERTS */}
            {error && (
              <div
                role="alert"
                className="mt-6 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
              >
                {error}
              </div>
            )}

            {success && (
              <div
                role="status"
                className="mt-6 rounded-xl border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-700"
              >
                {success}
              </div>
            )}

            {/* STK PAYMENT */}
            {paymentMethod === "stk" && (
              <div className="mt-8">
                <div className="rounded-xl border border-slate-200 p-5">
                  <h3 className="font-semibold text-slate-950">
                    Pay with M-Pesa Express
                  </h3>

                  <p className="mt-1 text-sm text-slate-500">
                    Enter the M-Pesa number you want to pay
                    with.
                  </p>

                  <div className="mt-6">
                    <Label htmlFor="phone">
                      M-Pesa phone number
                    </Label>

                    <Input
                      id="phone"
                      name="phone"
                      type="tel"
                      inputMode="numeric"
                      autoComplete="tel"
                      placeholder="0712 345 678"
                      value={phoneNumber}
                      onChange={handlePhoneChange}
                      disabled={paying}
                      className="mt-2 h-12"
                    />
                  </div>

                  <div className="mt-6 rounded-lg bg-slate-50 px-4 py-3">
                    <div className="flex justify-between text-sm">
                      <span className="text-slate-500">
                        Amount
                      </span>

                      <span className="font-bold text-slate-950">
                        {formattedAmount}
                      </span>
                    </div>
                  </div>

                  <Button
                    type="button"
                    onClick={handleStkPayment}
                    disabled={paying}
                    className="mt-6 h-12 w-full bg-[#0757B8] text-sm font-semibold hover:bg-[#064A9D]"
                  >
                    {paying
                      ? "Processing..."
                      : `Pay ${formattedAmount}`}
                  </Button>

                  <p className="mt-4 text-center text-xs leading-5 text-slate-400">
                    After you continue, an M-Pesa prompt will
                    be sent to this phone number. Enter your
                    M-Pesa PIN to complete the payment.
                  </p>
                </div>
              </div>
            )}

            {/* PAYBILL */}
            {paymentMethod === "paybill" && (
              <div className="mt-8">
                <div className="rounded-xl border border-slate-200 p-5">
                  <h3 className="font-semibold text-slate-950">
                    Pay manually using M-Pesa
                  </h3>

                  <p className="mt-1 text-sm text-slate-500">
                    Use the payment details below to complete
                    your payment.
                  </p>

                  <div className="mt-6 divide-y divide-slate-100 rounded-xl bg-slate-50">
                    <div className="flex items-center justify-between px-5 py-4">
                      <span className="text-sm text-slate-500">
                        Paybill Number
                      </span>

                      <span className="font-bold text-slate-950">
                        PAYBILL
                      </span>
                    </div>

                    <div className="flex items-center justify-between px-5 py-4">
                      <span className="text-sm text-slate-500">
                        Account Number
                      </span>

                      <span className="font-bold text-slate-950">
                        YOUR ACCOUNT
                      </span>
                    </div>

                    <div className="flex items-center justify-between px-5 py-4">
                      <span className="text-sm text-slate-500">
                        Amount
                      </span>

                      <span className="font-bold text-[#0757B8]">
                        {formattedAmount}
                      </span>
                    </div>
                  </div>

                  <div className="mt-6 rounded-xl border border-slate-200 p-5">
                    <h4 className="font-semibold text-slate-900">
                      How to pay
                    </h4>

                    <ol className="mt-4 space-y-2 text-sm leading-6 text-slate-600">
                      <li>
                        1. Open M-Pesa on your phone.
                      </li>
                      <li>
                        2. Select <strong>Lipa na M-Pesa</strong>.
                      </li>
                      <li>
                        3. Select <strong>Pay Bill</strong>.
                      </li>
                      <li>
                        4. Enter the Paybill number shown above.
                      </li>
                      <li>
                        5. Enter the account number shown above.
                      </li>
                      <li>
                        6. Enter the amount shown above.
                      </li>
                      <li>
                        7. Enter your M-Pesa PIN.
                      </li>
                    </ol>
                  </div>

                  <div className="mt-6">
                    <Label htmlFor="reference">
                      M-Pesa transaction code
                    </Label>

                    <Input
                      id="reference"
                      name="reference"
                      type="text"
                      placeholder="e.g. QH7ABC1234"
                      value={mpesaReference}
                      onChange={handleReferenceChange}
                      disabled={paying}
                      className="mt-2 h-12 uppercase"
                    />

                    <p className="mt-2 text-xs text-slate-500">
                      Enter the transaction code from your
                      M-Pesa confirmation message.
                    </p>
                  </div>

                  <Button
                    type="button"
                    onClick={handleManualPayment}
                    disabled={paying}
                    className="mt-6 h-12 w-full bg-[#0757B8] text-sm font-semibold hover:bg-[#064A9D]"
                  >
                    {paying
                      ? "Submitting..."
                      : "I've Made the Payment"}
                  </Button>
                </div>
              </div>
            )}
          </section>
        </div>
      </div>
    </main>
  );
}

