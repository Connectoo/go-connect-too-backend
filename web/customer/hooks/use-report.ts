"use client";

import { useMutation } from "@tanstack/react-query";
import { submitReport } from "@/services/report";
import type { CreateReportInput } from "@/types/report";

export function useSubmitReport() {
  return useMutation({
    mutationFn: (body: CreateReportInput) => submitReport(body),
  });
}
