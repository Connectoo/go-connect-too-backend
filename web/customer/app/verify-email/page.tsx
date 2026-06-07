import { Suspense } from "react";
import { VerifyEmailContent } from "@/components/auth/verify-email-content";

export default function VerifyEmailPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 p-4">
      <Suspense fallback={<p className="text-sm text-muted-foreground">Loading...</p>}>
        <VerifyEmailContent />
      </Suspense>
    </div>
  );
}
