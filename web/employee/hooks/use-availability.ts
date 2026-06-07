"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createAvailability,
  deleteAvailability,
  fetchAvailability,
  updateAvailability,
} from "@/services/availability";
import type { UpdateAvailabilityRequest } from "@/types/availability";

export function useAvailability() {
  return useQuery({
    queryKey: ["employee", "availability"],
    queryFn: fetchAvailability,
  });
}

export function useAvailabilityMutations() {
  const qc = useQueryClient();
  const invalidate = () =>
    qc.invalidateQueries({ queryKey: ["employee", "availability"] });

  return {
    create: useMutation({
      mutationFn: createAvailability,
      onSuccess: invalidate,
    }),
    update: useMutation({
      mutationFn: ({
        id,
        body,
      }: {
        id: string;
        body: UpdateAvailabilityRequest;
      }) => updateAvailability(id, body),
      onSuccess: invalidate,
    }),
    remove: useMutation({
      mutationFn: deleteAvailability,
      onSuccess: invalidate,
    }),
  };
}
