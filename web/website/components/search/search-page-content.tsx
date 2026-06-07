"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { FormEvent, useState } from "react";
import { ProviderCard } from "@/components/providers/provider-card";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { useEmployeeSearch, useServiceSearch } from "@/hooks/use-search";
import { customerBookUrl } from "@/lib/portal-url";
import { formatCurrency } from "@/lib/utils";

type Tab = "services" | "providers";

export function SearchPageContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const initialQ = searchParams.get("q") ?? "";
  const initialTab = (searchParams.get("tab") as Tab) || "services";

  const [query, setQuery] = useState(initialQ);
  const [activeQuery, setActiveQuery] = useState(initialQ);
  const [tab, setTab] = useState<Tab>(initialTab);

  const services = useServiceSearch({
    q: activeQuery,
    enabled: tab === "services" && Boolean(activeQuery.trim()),
  });
  const employees = useEmployeeSearch({
    q: activeQuery,
    enabled: tab === "providers" && Boolean(activeQuery.trim()),
  });

  const onSearch = (e: FormEvent) => {
    e.preventDefault();
    const trimmed = query.trim();
    setActiveQuery(trimmed);
    const params = new URLSearchParams();
    if (trimmed) params.set("q", trimmed);
    params.set("tab", tab);
    router.replace(`/search?${params.toString()}`);
  };

  const isLoading = tab === "services" ? services.isLoading : employees.isLoading;

  return (
    <div className="container mx-auto px-4 py-10">
      <h1 className="mb-2 text-3xl font-bold">Search</h1>
      <p className="mb-8 text-muted-foreground">
        Find services and providers across the marketplace.
      </p>

      <form onSubmit={onSearch} className="mb-6 flex max-w-xl gap-2">
        <Input
          placeholder="Search services or providers..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <Button type="submit">Search</Button>
      </form>

      <div className="mb-6 flex gap-2">
        <Button
          variant={tab === "services" ? "default" : "outline"}
          onClick={() => setTab("services")}
        >
          Services
        </Button>
        <Button
          variant={tab === "providers" ? "default" : "outline"}
          onClick={() => setTab("providers")}
        >
          Providers
        </Button>
      </div>

      {!activeQuery.trim() && (
        <p className="text-sm text-muted-foreground">
          Enter a search term to see results.
        </p>
      )}

      {activeQuery.trim() && isLoading && (
        <div className="grid gap-4 md:grid-cols-2">
          <Skeleton className="h-32" />
          <Skeleton className="h-32" />
        </div>
      )}

      {tab === "services" && activeQuery.trim() && !isLoading && (
        <div className="grid gap-4 md:grid-cols-2">
          {(services.data ?? []).length === 0 ? (
            <p className="text-sm text-muted-foreground">No services found.</p>
          ) : (
            services.data?.map((service) => (
              <Card key={service.id}>
                <CardContent className="space-y-3 pt-6">
                  <div>
                    <p className="font-semibold">{service.title}</p>
                    {service.employee_display_name && (
                      <p className="text-sm text-muted-foreground">
                        by {service.employee_display_name}
                      </p>
                    )}
                  </div>
                  {service.description && (
                    <p className="line-clamp-2 text-sm text-muted-foreground">
                      {service.description}
                    </p>
                  )}
                  <div className="flex items-center justify-between">
                    <span className="font-medium">{formatCurrency(service.price)}</span>
                    <Button size="sm" asChild>
                      <Link href={customerBookUrl(service.id)}>Book now</Link>
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))
          )}
        </div>
      )}

      {tab === "providers" && activeQuery.trim() && !isLoading && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {(employees.data ?? []).length === 0 ? (
            <p className="text-sm text-muted-foreground">No providers found.</p>
          ) : (
            employees.data?.map((provider) => (
              <ProviderCard
                key={provider.id}
                provider={{
                  id: provider.id,
                  display_name: provider.display_name,
                  bio: provider.bio,
                  experience_years: provider.experience_years,
                  profile_photo_url: provider.profile_photo_url,
                  location_text: provider.location_text,
                  languages: provider.languages ?? [],
                  skills: provider.skills ?? [],
                  rating: provider.rating,
                  total_reviews: 0,
                }}
              />
            ))
          )}
        </div>
      )}
    </div>
  );
}
