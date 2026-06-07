"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useProfileMutations } from "@/hooks/use-profile";

export default function NotificationSettingsPage() {
  const { resendVerification } = useProfileMutations();

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <Link
          href="/profile"
          className="mb-2 inline-block text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to profile
        </Link>
        <h1 className="text-2xl font-bold">Notifications</h1>
        <p className="text-sm text-muted-foreground">
          Email verification and notification preferences.
        </p>
      </div>

      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>Email verification</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Didn&apos;t receive a verification email? We can send another one to your
            registered address.
          </p>

          {resendVerification.isError && (
            <p className="text-sm text-destructive">
              {(resendVerification.error as Error)?.message ||
                "Failed to send verification email"}
            </p>
          )}
          {resendVerification.isSuccess && (
            <p className="text-sm text-green-600">Verification email sent.</p>
          )}

          <Button
            onClick={() => resendVerification.mutate()}
            disabled={resendVerification.isPending}
          >
            {resendVerification.isPending ? "Sending..." : "Resend verification email"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
