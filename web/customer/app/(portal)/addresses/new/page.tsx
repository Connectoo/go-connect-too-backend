"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { AddressForm } from "@/components/shared/address-form";
import { useAddressMutations } from "@/hooks/use-addresses";

export default function NewAddressPage() {
  const router = useRouter();
  const { create } = useAddressMutations();

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <Link
          href="/addresses"
          className="mb-2 inline-block text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to addresses
        </Link>
        <h1 className="text-2xl font-bold">Add address</h1>
      </div>

      <AddressForm
        submitLabel="Add address"
        loading={create.isPending}
        error={create.isError ? (create.error as Error)?.message : null}
        onSubmit={(values) =>
          create.mutate(values, { onSuccess: () => router.push("/addresses") })
        }
      />
    </div>
  );
}
