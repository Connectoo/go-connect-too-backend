"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchProfile, updateProfile } from "@/services/profile";

export function useProfile() {
  return useQuery({
    queryKey: ["employee", "profile"],
    queryFn: fetchProfile,
  });
}

export function useProfileMutations() {
  const qc = useQueryClient();
  return {
    update: useMutation({
      mutationFn: updateProfile,
      onSuccess: () =>
        qc.invalidateQueries({ queryKey: ["employee", "profile"] }),
    }),
  };
}
