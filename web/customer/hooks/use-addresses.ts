"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createAddress,
  deleteAddress,
  fetchAddresses,
  updateAddress,
} from "@/services/profile";
import type { AddressInput } from "@/types/address";

export function useAddresses() {
  return useQuery({
    queryKey: ["customer", "addresses"],
    queryFn: fetchAddresses,
  });
}

export function useAddressMutations() {
  const qc = useQueryClient();
  const invalidate = () =>
    qc.invalidateQueries({ queryKey: ["customer", "addresses"] });

  return {
    create: useMutation({
      mutationFn: (body: AddressInput) => createAddress(body),
      onSuccess: invalidate,
    }),
    update: useMutation({
      mutationFn: ({ id, ...body }: AddressInput & { id: string }) =>
        updateAddress(id, body),
      onSuccess: invalidate,
    }),
    remove: useMutation({
      mutationFn: deleteAddress,
      onSuccess: invalidate,
    }),
  };
}
