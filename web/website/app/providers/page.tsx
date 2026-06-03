"use client";

import { Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { ProviderCard } from "@/components/providers/provider-card";
import { Skeleton } from "@/components/ui/skeleton";
import { useProviders } from "@/hooks/use-public";

function ProvidersContent() {
  const searchParams = useSearchParams();
  const category = searchParams.get("category");
  const { data, isLoading } = useProviders(50);

  return (
    <div className="container mx-auto px-4 py-10">
      <div className="mb-8 max-w-2xl">
        <h1 className="text-3xl font-bold">Service providers</h1>
        <p className="mt-2 text-muted-foreground">
          {category
            ? "Providers offering services in your selected category."
            : "Discover verified professionals ready to help."}
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {isLoading
          ? Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-48" />
            ))
          : data?.map((provider) => (
              <ProviderCard key={provider.id} provider={provider} />
            ))}
      </div>
    </div>
  );
}

export default function ProvidersPage() {
  return (
    <Suspense
      fallback={
        <div className="container mx-auto px-4 py-10">
          <Skeleton className="h-10 w-64" />
        </div>
      }
    >
      <ProvidersContent />
    </Suspense>
  );
}
