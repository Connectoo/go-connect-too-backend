"use client";

import Link from "next/link";
import { use } from "react";
import { Star } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useProvider, useServices } from "@/hooks/use-public";
import { formatCurrency } from "@/lib/utils";

export default function ProviderProfilePage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const { data: provider, isLoading } = useProvider(id);
  const { data: services } = useServices({ limit: 20 });

  const providerServices =
    services?.filter((s) => s.employee_id === id) ?? [];

  if (isLoading) {
    return (
      <div className="container mx-auto space-y-4 px-4 py-10">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  if (!provider) {
    return (
      <div className="container mx-auto px-4 py-10">
        <p>Provider not found.</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-10">
      <div className="grid gap-8 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <div>
            <h1 className="text-3xl font-bold">
              {provider.display_name ?? "Service Provider"}
            </h1>
            {provider.location_text && (
              <p className="mt-1 text-muted-foreground">{provider.location_text}</p>
            )}
            {provider.rating != null && (
              <p className="mt-2 flex items-center gap-1 text-sm">
                <Star className="h-4 w-4 fill-amber-400 text-amber-400" />
                {provider.rating.toFixed(1)} · {provider.total_reviews} reviews
              </p>
            )}
          </div>

          {provider.bio && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">About</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">{provider.bio}</p>
              </CardContent>
            </Card>
          )}

          <div>
            <h2 className="mb-4 text-xl font-semibold">Services</h2>
            <div className="grid gap-4 md:grid-cols-2">
              {providerServices.length === 0 ? (
                <p className="text-sm text-muted-foreground">No services listed yet.</p>
              ) : (
                providerServices.map((service) => (
                  <Card key={service.id}>
                    <CardHeader>
                      <CardTitle className="text-base">{service.title}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex items-center justify-between">
                      <span className="text-sm text-muted-foreground">
                        {service.duration_minutes} min
                      </span>
                      <div className="flex items-center gap-2">
                        <span className="font-semibold">
                          {formatCurrency(service.price)}
                        </span>
                        <Button size="sm" variant="outline" asChild>
                          <Link href={`/services/${service.id}`}>Details</Link>
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ))
              )}
            </div>
          </div>
        </div>

        <aside className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Profile details</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <p>{provider.experience_years} years experience</p>
              <div className="flex flex-wrap gap-2">
                {provider.skills.map((skill) => (
                  <Badge key={skill} variant="secondary">
                    {skill}
                  </Badge>
                ))}
              </div>
              <div className="flex flex-wrap gap-2">
                {provider.languages.map((lang) => (
                  <Badge key={lang} variant="outline">
                    {lang}
                  </Badge>
                ))}
              </div>
            </CardContent>
          </Card>
        </aside>
      </div>
    </div>
  );
}
