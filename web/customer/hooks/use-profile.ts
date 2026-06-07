"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  changePassword,
  deactivateAccount,
  fetchProfile,
  resendVerificationEmail,
  updateProfile,
} from "@/services/profile";
import type { ChangePasswordRequest, UpdateProfileRequest } from "@/types/profile";

export function useProfile() {
  return useQuery({
    queryKey: ["customer", "profile"],
    queryFn: fetchProfile,
  });
}

export function useProfileMutations() {
  const qc = useQueryClient();
  return {
    update: useMutation({
      mutationFn: (body: UpdateProfileRequest) => updateProfile(body),
      onSuccess: () => qc.invalidateQueries({ queryKey: ["customer", "profile"] }),
    }),
    changePassword: useMutation({
      mutationFn: (body: ChangePasswordRequest) => changePassword(body),
    }),
    deactivate: useMutation({
      mutationFn: deactivateAccount,
      onSuccess: () => qc.invalidateQueries({ queryKey: ["customer", "profile"] }),
    }),
    resendVerification: useMutation({
      mutationFn: resendVerificationEmail,
    }),
  };
}
