"use client";

import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { verifyEmail } from "@/services/auth";

export function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const attempted = useRef(false);

  const mutation = useMutation({
    mutationFn: (t: string) => verifyEmail(t),
  });

  useEffect(() => {
    if (token && !attempted.current) {
      attempted.current = true;
      mutation.mutate(token);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- run once per token
  }, [token]);

  if (!token) {
    return (
      <Card className="w-full max-w-md">
        <CardContent className="pt-6">
          <p className="text-sm text-destructive">Missing verification token.</p>
          <Button asChild className="mt-4" variant="outline">
            <Link href="/login">Sign in</Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Verify your email</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {mutation.isPending && (
          <p className="text-sm text-muted-foreground">Verifying your email…</p>
        )}
        {mutation.isSuccess && (
          <>
            <p className="text-sm text-green-600">Email verified successfully.</p>
            <Button asChild className="w-full">
              <Link href="/login">Sign in</Link>
            </Button>
          </>
        )}
        {mutation.isError && (
          <>
            <p className="text-sm text-destructive">
              {(mutation.error as Error)?.message || "Verification failed"}
            </p>
            <Button asChild variant="outline" className="w-full">
              <Link href="/login">Sign in to resend verification</Link>
            </Button>
          </>
        )}
      </CardContent>
    </Card>
  );
}
