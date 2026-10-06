
"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import {
  ArrowRight,
  Minus,
  Plus,
  RefreshCw,
  ShoppingCart,
  Trash2,
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

import { useAuthStore } from "@/store/auth.store";

import type { Service } from "@/types/service";
import type { SaleItemInput } from "@/types/sale";

import { serviceService } from "@/services/service.service";
import { offlineServiceService } from "@/services/offline-service.service";
import { attendantService } from "@/services/attendant.service";
import { createOfflineSale } from "@/services/offline-sale.service";

import DashboardShell from "@/components/dashboard/DashboardShell";
import PageHeader from "@/components/dashboard/PageHeader";

interface CartItem extends SaleItemInput {
  id: string;
  service_name: string;
}

export default function SalesPage() {
  const router = useRouter();

  const {
    user,
    isAuthenticated,
    offlineBranchId,
    offlineBranchName,
  } = useAuthStore();

  const [branchId, setBranchId] = useState<string | null>(null);
  const [services, setServices] = useState<Service[]>([]);

  const [customerId, setCustomerId] = useState("");
  const [discountValue, setDiscountValue] = useState("0");

  const [cart, setCart] = useState<CartItem[]>([]);

  const [isOnline, setIsOnline] = useState(true);
  const [loadingServices, setLoadingServices] = useState(false);
  const [loading, setLoading] = useState(false);

  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  /*
   * Track internet connectivity.
   */
  useEffect(() => {
    setIsOnline(navigator.onLine);

    const handleOnline = () => {
      setIsOnline(true);
    };

    const handleOffline = () => {
      setIsOnline(false);
    };

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  /*
   * Resolve the branch for the current user.
   *
   * Offline attendant:
   *   Uses the branch stored during offline authentication.
   *
   * Online attendant:
   *   Gets the assigned branch from GET /api/attendants/me.
   *
   * Owner/testing:
   *   Falls back to ?branchId=... when provided.
   */
  useEffect(() => {
    if (!user || !isAuthenticated) {
      return;
    }

    setError("");

    /*
     * Offline attendant already has the branch
     * stored during offline authentication.
     */
    if (offlineBranchId) {
      setBranchId(offlineBranchId);
      return;
    }

    /*
     * Online attendant gets their assigned branch
     * from the backend.
     */
    if (user.role === "attendant" && isOnline) {
      let cancelled = false;

      const loadAttendantBranch = async () => {
        try {
          const token =
            useAuthStore.getState().token;

          if (!token) {
            throw new Error(
              "Authentication token is missing.",
            );
          }

          const attendant =
            await attendantService.me(token);

          if (cancelled) {
            return;
          }

          if (!attendant.branch_id) {
            setBranchId(null);

            setError(
              "You are not assigned to a branch.",
            );

            return;
          }

          setBranchId(attendant.branch_id);
        } catch (err) {
          if (!cancelled) {
            setBranchId(null);

            setError(
              err instanceof Error
                ? err.message
                : "Unable to load your assigned branch.",
            );
          }
        }
      };

      loadAttendantBranch();

      return () => {
        cancelled = true;
      };
    }

    /*
     * Owner/testing fallback.
     */
    const params = new URLSearchParams(
      window.location.search,
    );

    const queryBranchId =
      params.get("branchId");

    if (queryBranchId) {
      setBranchId(queryBranchId);
    }
  }, [
    user,
    isAuthenticated,
    offlineBranchId,
    isOnline,
  ]);

  /*
   * Load services.
   *
   * ONLINE:
   *   1. Get services from backend.
   *   2. Cache them in IndexedDB.
   *   3. Display them.
   *
   * OFFLINE:
   *   1. Get services from IndexedDB.
   *   2. Display them.
   */
  useEffect(() => {
    if (!branchId || !user || !isAuthenticated) {
      return;
    }

    let cancelled = false;

    const loadServices = async () => {
      setLoadingServices(true);
      setError("");

      try {
        /*
         * ONLINE
         */
        if (isOnline) {
          const token =
            useAuthStore.getState().token;

          if (!token) {
            throw new Error(
              "Authentication token is missing.",
            );
          }

          const response =
            await serviceService.list(
              token,
              {
                branch_id: branchId,
                active_only: true,
                limit: 100,
                offset: 0,
              },
            );

          /*
           * Save the latest server services
           * into IndexedDB.
           */
          await offlineServiceService.replaceServices(
            branchId,
            response.services,
          );

          if (!cancelled) {
            setServices(response.services);
          }

          return;
        }

        /*
         * OFFLINE
         */
        const localServices =
          await offlineServiceService.list(
            branchId,
            true,
          );

        if (!cancelled) {
          setServices(localServices);

          if (localServices.length === 0) {
            setError(
              "No cached services are available for this branch. Connect to the internet once to download the services.",
            );
          }
        }
      } catch (err) {
        if (!cancelled) {
          setServices([]);

          setError(
            err instanceof Error
              ? err.message
              : "Unable to load services.",
          );
        }
      } finally {
        if (!cancelled) {
          setLoadingServices(false);
        }
      }
    };

    loadServices();

    return () => {
      cancelled = true;
    };
  }, [
    branchId,
    user,
    isAuthenticated,
    isOnline,
  ]);

  /*
   * Calculate subtotal.
   */
  const subtotal = useMemo(() => {
    return cart.reduce(
      (total, item) =>
        total +
        Number(item.unit_price) *
          item.quantity,
      0,
    );
  }, [cart]);

  /*
   * Calculate discount.
   */
  const discount = Math.max(
    0,
    Number(discountValue) || 0,
  );

  /*
   * Calculate total.
   */
  const total = Math.max(
    0,
    subtotal - discount,
  );

  /*
   * Add service to cart.
   */
  const addService = (
    service: Service,
  ) => {
    setError("");
    setSuccess("");

    setCart((current) => {
      const existing =
        current.find(
          (item) =>
            item.service_id ===
            service.id,
        );

      if (existing) {
        return current.map((item) =>
          item.service_id ===
          service.id
            ? {
                ...item,
                quantity:
                  item.quantity + 1,
              }
            : item,
        );
      }

      return [
        ...current,
        {
          id: crypto.randomUUID(),
          service_id: service.id,
          service_name: service.name,
          description: service.name,
          quantity: 1,
          unit_price: service.price,
        },
      ];
    });
  };

  /*
   * Remove item from cart.
   */
  const removeItem = (
    id: string,
  ) => {
    setCart((current) =>
      current.filter(
        (item) =>
          item.id !== id,
      ),
    );
  };

  /*
   * Change quantity.
   */
  const changeQuantity = (
    id: string,
    amount: number,
  ) => {
    setCart((current) =>
      current.map((item) => {
        if (item.id !== id) {
          return item;
        }

        return {
          ...item,
          quantity: Math.max(
            1,
            item.quantity + amount,
          ),
        };
      }),
    );
  };

  /*
   * Create sale.
   *
   * The sale is created through the
   * offline-first sale service regardless
   * of current connectivity.
   */
  const handleCreateSale = async () => {
    setError("");
    setSuccess("");

    if (!isAuthenticated || !user) {
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

    if (!branchId) {
      setError(
        "No branch is selected.",
      );

      return;
    }

    if (cart.length === 0) {
      setError(
        "Add at least one service.",
      );

      return;
    }

    if (discount > subtotal) {
      setError(
        "Discount cannot be greater than the subtotal.",
      );

      return;
    }

    setLoading(true);

    try {
      const items: SaleItemInput[] =
        cart.map((item) => ({
          service_id:
            item.service_id ?? null,
          description:
            item.description,
          quantity:
            item.quantity,
          unit_price:
            item.unit_price,
        }));

      const sale =
        await createOfflineSale(
          {
            branch_id: branchId,
            customer_id:
              customerId.trim() ||
              null,
            session_id: null,
            session_started_at: null,
            terminal_id: null,
            attendant_id: user.id,
            discount_type:
              discount > 0
                ? "fixed"
                : null,
            discount_value:
              discount.toFixed(2),
            items,
          },
          user.tenant_id,
          user.id,
        );

      setSuccess(
        isOnline
          ? "Sale saved and queued for synchronization."
          : "Sale saved offline.",
      );

      /*
       * Continue to payment.
       */
      router.push(
        `/cyber-attendant/payment?saleId=${sale.id}`,
      );
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Unable to create the sale.",
      );
    } finally {
      setLoading(false);
    }
  };

  /*
   * Authentication guard.
   */
  if (!isAuthenticated || !user) {
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
          title="New Sale"
          description="Select services, add them to the sale, and continue to payment."
        />

        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm">
            {isOnline ? (
              <>
                <Wifi className="h-4 w-4 text-green-600" />

                <span>
                  Online
                </span>
              </>
            ) : (
              <>
                <WifiOff className="h-4 w-4 text-amber-600" />

                <span>
                  Offline mode
                </span>
              </>
            )}
          </div>

          {offlineBranchName && (
            <div className="rounded-lg bg-slate-50 px-3 py-2 text-sm text-slate-600">
              Branch:{" "}
              <span className="font-medium text-slate-900">
                {offlineBranchName}
              </span>
            </div>
          )}
        </div>

        {error && (
          <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {success && (
          <div className="rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-700">
            {success}
          </div>
        )}

        <div className="grid gap-6 lg:grid-cols-[1fr_420px]">
          <Card className="border-slate-200 shadow-sm">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <ShoppingCart className="h-5 w-5 text-[#0757B8]" />

                Available Services
              </CardTitle>
            </CardHeader>

            <CardContent>
              {loadingServices ? (
                <div className="flex items-center justify-center py-12">
                  <RefreshCw className="mr-2 h-5 w-5 animate-spin text-[#0757B8]" />

                  <span className="text-sm text-slate-500">
                    Loading services...
                  </span>
                </div>
              ) : services.length === 0 ? (
                <div className="rounded-lg border border-dashed border-slate-300 py-10 text-center">
                  {isOnline ? (
                    <>
                      <ShoppingCart className="mx-auto mb-3 h-8 w-8 text-slate-400" />

                      <p className="font-medium text-slate-700">
                        No active services
                      </p>

                      <p className="mt-1 text-sm text-slate-500">
                        Ask the Cyber Owner to add services for this branch.
                      </p>
                    </>
                  ) : (
                    <>
                      <WifiOff className="mx-auto mb-3 h-8 w-8 text-slate-400" />

                      <p className="font-medium text-slate-700">
                        No cached services
                      </p>

                      <p className="mt-1 text-sm text-slate-500">
                        Connect to the internet once to download services for offline use.
                      </p>
                    </>
                  )}
                </div>
              ) : (
                <div className="grid gap-3 sm:grid-cols-2">
                  {services.map(
                    (service) => (
                      <div
                        key={service.id}
                        className="rounded-lg border border-slate-200 p-4 transition hover:border-[#0757B8] hover:shadow-sm"
                      >
                        <div className="flex items-start justify-between gap-3">
                          <div className="min-w-0">
                            <h3 className="font-medium text-slate-900">
                              {service.name}
                            </h3>

                            {service.description && (
                              <p className="mt-1 text-sm text-slate-500">
                                {
                                  service.description
                                }
                              </p>
                            )}

                            <p className="mt-3 text-lg font-bold text-[#0757B8]">
                              KES{" "}
                              {Number(
                                service.price,
                              ).toFixed(
                                2,
                              )}
                            </p>
                          </div>
                        </div>

                        <Button
                          type="button"
                          onClick={() =>
                            addService(
                              service,
                            )
                          }
                          className="mt-4 w-full bg-[#0757B8] hover:bg-[#064A9D]"
                        >
                          <Plus className="mr-2 h-4 w-4" />

                          Add to Sale
                        </Button>
                      </div>
                    ),
                  )}
                </div>
              )}
            </CardContent>
          </Card>

          <div className="space-y-6">
            <Card className="border-slate-200 shadow-sm">
              <CardHeader>
                <CardTitle>
                  Current Sale
                </CardTitle>
              </CardHeader>

              <CardContent className="space-y-3">
                {cart.length === 0 ? (
                  <div className="rounded-lg border border-dashed border-slate-300 py-8 text-center">
                    <ShoppingCart className="mx-auto mb-3 h-7 w-7 text-slate-400" />

                    <p className="text-sm text-slate-500">
                      No services added yet.
                    </p>
                  </div>
                ) : (
                  cart.map((item) => {
                    const lineTotal =
                      Number(
                        item.unit_price,
                      ) *
                      item.quantity;

                    return (
                      <div
                        key={item.id}
                        className="rounded-lg border border-slate-200 p-3"
                      >
                        <div className="flex items-start justify-between gap-3">
                          <div>
                            <p className="font-medium text-slate-900">
                              {
                                item.service_name
                              }
                            </p>

                            <p className="text-sm text-slate-500">
                              KES{" "}
                              {Number(
                                item.unit_price,
                              ).toFixed(
                                2,
                              )}{" "}
                              ×{" "}
                              {
                                item.quantity
                              }
                            </p>
                          </div>

                          <p className="font-semibold text-slate-900">
                            KES{" "}
                            {lineTotal.toFixed(
                              2,
                            )}
                          </p>
                        </div>

                        <div className="mt-3 flex items-center justify-between">
                          <div className="flex items-center rounded-md border border-slate-200">
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              onClick={() =>
                                changeQuantity(
                                  item.id,
                                  -1,
                                )
                              }
                            >
                              <Minus className="h-4 w-4" />
                            </Button>

                            <span className="w-8 text-center text-sm">
                              {
                                item.quantity
                              }
                            </span>

                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              onClick={() =>
                                changeQuantity(
                                  item.id,
                                  1,
                                )
                              }
                            >
                              <Plus className="h-4 w-4" />
                            </Button>
                          </div>

                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            onClick={() =>
                              removeItem(
                                item.id,
                              )
                            }
                            className="text-red-600 hover:bg-red-50"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    );
                  })
                )}
              </CardContent>
            </Card>

            <Card className="border-slate-200 shadow-sm">
              <CardHeader>
                <CardTitle>
                  Customer & Summary
                </CardTitle>
              </CardHeader>

              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="customerId">
                    Customer ID{" "}
                    <span className="text-slate-400">
                      (optional)
                    </span>
                  </Label>

                  <Input
                    id="customerId"
                    placeholder="Customer UUID"
                    value={customerId}
                    onChange={(e) =>
                      setCustomerId(
                        e.target.value,
                      )
                    }
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="discount">
                    Discount
                  </Label>

                  <Input
                    id="discount"
                    type="number"
                    min="0"
                    step="0.01"
                    value={
                      discountValue
                    }
                    onChange={(e) =>
                      setDiscountValue(
                        e.target.value,
                      )
                    }
                  />
                </div>

                <div className="space-y-3 border-t border-slate-200 pt-4">
                  <div className="flex justify-between text-sm">
                    <span className="text-slate-500">
                      Subtotal
                    </span>

                    <span>
                      KES{" "}
                      {subtotal.toFixed(
                        2,
                      )}
                    </span>
                  </div>

                  <div className="flex justify-between text-sm">
                    <span className="text-slate-500">
                      Discount
                    </span>

                    <span>
                      KES{" "}
                      {discount.toFixed(
                        2,
                      )}
                    </span>
                  </div>

                  <div className="flex justify-between border-t border-slate-200 pt-3">
                    <span className="font-semibold">
                      Total
                    </span>

                    <span className="text-xl font-bold text-[#0757B8]">
                      KES{" "}
                      {total.toFixed(
                        2,
                      )}
                    </span>
                  </div>
                </div>

                <Button
                  type="button"
                  disabled={
                    loading ||
                    cart.length === 0
                  }
                  onClick={
                    handleCreateSale
                  }
                  className="w-full bg-[#0757B8] hover:bg-[#064A9D]"
                >
                  {loading ? (
                    <>
                      <RefreshCw className="mr-2 h-4 w-4 animate-spin" />

                      Creating Sale...
                    </>
                  ) : (
                    <>
                      Continue to Payment

                      <ArrowRight className="ml-2 h-4 w-4" />
                    </>
                  )}
                </Button>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </DashboardShell>
  );
}
