
import { create } from "zustand";
import { reportService } from "@/services/report.service";
import { useAuthStore } from "@/store/auth.store";
import {
  FetchOwnerReportFilters,
  OwnerReportSummaryResponse,
} from "@/types/reports";

interface ReportState {
  data: OwnerReportSummaryResponse | null;
  loading: boolean;
  error: string | null;
  filters: FetchOwnerReportFilters;
  setFilters: (filters: Partial<FetchOwnerReportFilters>) => void;
  fetchReport: () => Promise<void>;
}

export const useReportStore = create<ReportState>((set, get) => ({
  data: null,
  loading: false,
  error: null,
  filters: {},

  setFilters: (newFilters) => {
    set((state) => ({
      filters: {
        ...state.filters,
        ...newFilters,
      },
    }));
  },

  fetchReport: async () => {
    const token = useAuthStore.getState().token;

    if (!token) {
      return;
    }

    set({
      loading: true,
      error: null,
    });

    try {
      const reportData = await reportService.getOwnerSummary(
        token,
        get().filters
      );

      set({
        data: reportData,
        loading: false,
      });
    } catch (err) {
      set({
        error:
          err instanceof Error
            ? err.message
            : "Failed to load report data.",
        loading: false,
      });
    }
  },
}));
