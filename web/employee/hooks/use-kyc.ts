"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { changePassword, fetchKYC, submitKYC } from "@/services/kyc";
import type { ChangePasswordRequest } from "@/types/profile";
import type { SubmitKYCRequest } from "@/types/kyc";

export function useKYC() {
  return useQuery({
    queryKey: ["employee", "kyc"],
    queryFn: fetchKYC,
    retry: false,
  });
}

export function useKYCMutations() {
  const qc = useQueryClient();
  return {
    submit: useMutation({
      mutationFn: (body: SubmitKYCRequest) => submitKYC(body),
      onSuccess: () => qc.invalidateQueries({ queryKey: ["employee", "kyc"] }),
    }),
    changePassword: useMutation({
      mutationFn: (body: ChangePasswordRequest) => changePassword(body),
    }),
  };
}
