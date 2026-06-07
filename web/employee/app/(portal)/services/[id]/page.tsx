"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useState } from "react";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useCategories, useServiceMutations, useServices } from "@/hooks/use-services";

export default function ServiceDetailPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data: services, isLoading, isError, error } = useServices();
  const { data: categories } = useCategories();
  const { setStatus, remove } = useServiceMutations();
  const [confirmDelete, setConfirmDelete] = useState(false);

  const service = services?.find((s) => s.id === id);
  const categoryName =
    categories?.find((c) => c.id === service?.category_id)?.name ?? "—";

  if (isLoading) {
    return (
      <div className="p-6 md:p-8">
        <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">
          Loading service...
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-6 md:p-8">
        <div className="rounded-lg border border-destructive/40 p-8 text-center text-sm text-destructive">
          {(error as Error)?.message || "Failed to load service"}
        </div>
      </div>
    );
  }

  if (!service) {
    return (
      <div className="p-6 md:p-8">
        <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">
          Service not found.
        </div>
      </div>
    );
  }

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{service.title}</h1>
          <p className="text-sm text-muted-foreground">{categoryName}</p>
        </div>
        <Badge variant={service.is_active ? "default" : "secondary"}>
          {service.is_active ? "Active" : "Inactive"}
        </Badge>
      </div>

      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Details</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 text-sm">
          <div>
            <span className="text-muted-foreground">Description: </span>
            {service.description || "—"}
          </div>
          <div>
            <span className="text-muted-foreground">Price: </span>
            {service.price.toFixed(2)}
          </div>
          <div>
            <span className="text-muted-foreground">Duration: </span>
            {service.duration_minutes} min
          </div>
        </CardContent>
      </Card>

      <div className="mt-6 flex flex-wrap gap-2">
        <Button asChild variant="outline">
          <Link href={`/services/${id}/edit`}>Edit</Link>
        </Button>
        <Button
          variant={service.is_active ? "secondary" : "default"}
          disabled={setStatus.isPending}
          onClick={() =>
            setStatus.mutate({ id, isActive: !service.is_active })
          }
        >
          {service.is_active ? "Deactivate" : "Activate"}
        </Button>
        <Button variant="destructive" onClick={() => setConfirmDelete(true)}>
          Delete
        </Button>
      </div>

      {setStatus.isError && (
        <p className="mt-3 text-sm text-destructive">
          {(setStatus.error as Error)?.message || "Failed to update status"}
        </p>
      )}

      <ConfirmDialog
        open={confirmDelete}
        onOpenChange={setConfirmDelete}
        title="Delete service?"
        description="This permanently removes the service. This cannot be undone."
        confirmLabel="Delete"
        variant="destructive"
        loading={remove.isPending}
        onConfirm={async () => {
          await remove.mutateAsync(id);
          setConfirmDelete(false);
          router.push("/services");
        }}
      />
    </div>
  );
}
