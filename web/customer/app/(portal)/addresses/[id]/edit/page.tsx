"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { use } from "react";
import { AddressForm } from "@/components/shared/address-form";
import { Skeleton } from "@/components/ui/skeleton";
import { useAddressMutations, useAddresses } from "@/hooks/use-addresses";

export default function EditAddressPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const router = useRouter();
  const { data: addresses, isLoading } = useAddresses();
  const { update } = useAddressMutations();

  const address = addresses?.find((a) => a.id === id);

  if (isLoading) {
    return (
      <div className="p-6 md:p-8">
        <Skeleton className="h-64 w-full max-w-xl" />
      </div>
    );
  }

  if (!address) {
    return (
      <div className="p-6 md:p-8">
        <p className="text-sm text-muted-foreground">Address not found.</p>
        <Link href="/addresses" className="text-sm text-primary underline">
          Back to addresses
        </Link>
      </div>
    );
  }

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <Link
          href="/addresses"
          className="mb-2 inline-block text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to addresses
        </Link>
        <h1 className="text-2xl font-bold">Edit address</h1>
      </div>

      <AddressForm
        defaultValues={{
          label: address.label,
          address_line: address.address_line,
          city: address.city,
          state: address.state,
          country: address.country,
          pincode: address.pincode,
          is_default: address.is_default,
        }}
        submitLabel="Save changes"
        loading={update.isPending}
        error={update.isError ? (update.error as Error)?.message : null}
        onSubmit={(values) =>
          update.mutate(
            { id, ...values },
            { onSuccess: () => router.push("/addresses") },
          )
        }
      />
    </div>
  );
}
