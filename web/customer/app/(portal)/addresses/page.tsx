"use client";

import Link from "next/link";
import { useState } from "react";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useAddressMutations, useAddresses } from "@/hooks/use-addresses";

export default function AddressesPage() {
  const { data: addresses, isLoading } = useAddresses();
  const { remove } = useAddressMutations();
  const [deleteId, setDeleteId] = useState<string | null>(null);

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Saved addresses</h1>
          <p className="text-sm text-muted-foreground">
            Manage addresses for your bookings.
          </p>
        </div>
        <Button asChild>
          <Link href="/addresses/new">Add address</Link>
        </Button>
      </div>

      {isLoading && (
        <div className="space-y-3">
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
        </div>
      )}

      {!isLoading && (!addresses || addresses.length === 0) && (
        <Card>
          <CardContent className="py-10 text-center text-sm text-muted-foreground">
            No saved addresses yet.{" "}
            <Link href="/addresses/new" className="text-primary underline">
              Add your first address
            </Link>
          </CardContent>
        </Card>
      )}

      <div className="grid gap-4 md:grid-cols-2">
        {addresses?.map((address) => (
          <Card key={address.id}>
            <CardContent className="space-y-3 pt-6">
              <div className="flex items-center justify-between">
                <p className="font-medium">{address.label}</p>
                {address.is_default && <Badge variant="secondary">Default</Badge>}
              </div>
              <p className="text-sm text-muted-foreground">
                {address.address_line}
                <br />
                {address.city}, {address.state} {address.pincode}
                <br />
                {address.country}
              </p>
              <div className="flex gap-2">
                <Button size="sm" variant="outline" asChild>
                  <Link href={`/addresses/${address.id}/edit`}>Edit</Link>
                </Button>
                <Button size="sm" variant="ghost" onClick={() => setDeleteId(address.id)}>
                  Delete
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <ConfirmDialog
        open={Boolean(deleteId)}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title="Delete address?"
        description="This address will be permanently removed."
        confirmLabel="Delete"
        variant="destructive"
        loading={remove.isPending}
        onConfirm={async () => {
          if (!deleteId) return;
          await remove.mutateAsync(deleteId);
          setDeleteId(null);
        }}
      />
    </div>
  );
}
