"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { clearCustomerAuth } from "@/lib/auth";
import { useProfile, useProfileMutations } from "@/hooks/use-profile";

export default function AccountSettingsPage() {
  const router = useRouter();
  const { data } = useProfile();
  const { deactivate } = useProfileMutations();
  const [confirmOpen, setConfirmOpen] = useState(false);

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <Link
          href="/profile"
          className="mb-2 inline-block text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to profile
        </Link>
        <h1 className="text-2xl font-bold">Account</h1>
        <p className="text-sm text-muted-foreground">Manage your account status.</p>
      </div>

      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>Deactivate account</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Deactivating your account will prevent you from signing in and making new
            bookings. Your existing booking history will be retained.
          </p>
          {data?.status === "inactive" ? (
            <p className="text-sm font-medium text-muted-foreground">
              Your account is already inactive.
            </p>
          ) : (
            <Button variant="destructive" onClick={() => setConfirmOpen(true)}>
              Deactivate account
            </Button>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title="Deactivate account?"
        description="You will be signed out and will not be able to sign in again until an admin reactivates your account."
        confirmLabel="Deactivate"
        variant="destructive"
        loading={deactivate.isPending}
        onConfirm={async () => {
          await deactivate.mutateAsync();
          clearCustomerAuth();
          router.push("/login");
        }}
      />
    </div>
  );
}
