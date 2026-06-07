import { Suspense } from "react";
import { CheckoutContent } from "@/components/subscription/checkout-content";
import { Skeleton } from "@/components/ui/skeleton";

export default function SubscriptionCheckoutPage() {
  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Checkout</h1>
        <p className="text-sm text-muted-foreground">
          Complete payment to activate your subscription.
        </p>
      </div>

      <Suspense fallback={<Skeleton className="h-48 max-w-md" />}>
        <CheckoutContent />
      </Suspense>
    </div>
  );
}
