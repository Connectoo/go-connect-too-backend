"use client";

import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { CategoryCard } from "@/components/categories/category-card";
import { ProviderCard } from "@/components/providers/provider-card";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useHome } from "@/hooks/use-public";
import { formatCurrency } from "@/lib/utils";

export default function HomePage() {
  const { data, isLoading } = useHome();

  return (
    <div>
      <section className="border-b bg-gradient-to-b from-primary/5 to-background">
        <div className="container mx-auto px-4 py-16 md:py-24">
          <div className="max-w-2xl space-y-6">
            <h1 className="text-4xl font-bold tracking-tight md:text-5xl">
              Book trusted local service professionals
            </h1>
            <p className="text-lg text-muted-foreground">
              Compare providers, browse categories, and book services with confidence.
            </p>
            <div className="flex flex-wrap gap-3">
              <Button asChild size="lg">
                <Link href="/providers">
                  Browse providers <ArrowRight className="h-4 w-4" />
                </Link>
              </Button>
              <Button asChild variant="outline" size="lg">
                <Link href="/categories">View categories</Link>
              </Button>
            </div>
          </div>
          {data?.stats && (
            <div className="mt-12 grid gap-4 sm:grid-cols-3">
              {[
                { label: "Categories", value: data.stats.categories_count },
                { label: "Providers", value: data.stats.providers_count },
                { label: "Services", value: data.stats.services_count },
              ].map((stat) => (
                <Card key={stat.label}>
                  <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">
                      {stat.label}
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-3xl font-bold">{stat.value}</p>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      </section>

      <section className="container mx-auto px-4 py-12">
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-2xl font-semibold">Popular categories</h2>
          <Button variant="ghost" asChild>
            <Link href="/categories">See all</Link>
          </Button>
        </div>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {isLoading
            ? Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-32" />
              ))
            : data?.featured_categories.map((category) => (
                <CategoryCard key={category.id} category={category} />
              ))}
        </div>
      </section>

      <section className="container mx-auto px-4 py-12">
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-2xl font-semibold">Top providers</h2>
          <Button variant="ghost" asChild>
            <Link href="/providers">See all</Link>
          </Button>
        </div>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {isLoading
            ? Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-48" />
              ))
            : data?.featured_providers.map((provider) => (
                <ProviderCard key={provider.id} provider={provider} />
              ))}
        </div>
      </section>

      {data?.featured_services && data.featured_services.length > 0 && (
        <section className="container mx-auto px-4 py-12">
          <h2 className="mb-6 text-2xl font-semibold">Featured services</h2>
          <div className="grid gap-4 md:grid-cols-2">
            {data.featured_services.map((service) => (
              <Card key={service.id}>
                <CardHeader>
                  <CardTitle className="text-base">{service.title}</CardTitle>
                </CardHeader>
                <CardContent className="flex items-center justify-between">
                  <p className="text-sm text-muted-foreground line-clamp-2">
                    {service.description}
                  </p>
                  <span className="font-semibold">{formatCurrency(service.price)}</span>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
