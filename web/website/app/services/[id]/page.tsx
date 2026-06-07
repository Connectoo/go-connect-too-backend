"use client";

import Link from "next/link";
import { use } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useService } from "@/hooks/use-public";
import { customerBookUrl } from "@/lib/portal-url";
import { formatCurrency } from "@/lib/utils";

export default function ServiceDetailsPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const { data: service, isLoading } = useService(id);

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-10">
        <Skeleton className="h-64 w-full max-w-2xl" />
      </div>
    );
  }

  if (!service) {
    return (
      <div className="container mx-auto px-4 py-10">
        <p>Service not found.</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto max-w-2xl px-4 py-10">
      <Card>
        <CardHeader>
          <CardTitle>{service.title}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {service.description && (
            <p className="text-muted-foreground">{service.description}</p>
          )}
          <div className="flex items-center justify-between text-sm">
            <span>Duration: {service.duration_minutes} minutes</span>
            <span className="text-xl font-bold">{formatCurrency(service.price)}</span>
          </div>
          <div className="flex gap-3">
            <Button asChild>
              <Link href={`/providers/${service.employee_id}`}>View provider</Link>
            </Button>
            <Button asChild>
              <a href={customerBookUrl(service.id)}>Book now</a>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
