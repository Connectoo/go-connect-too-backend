"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useParams, useRouter } from "next/navigation";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useCategories, useServiceMutations, useServices } from "@/hooks/use-services";
import type { EmployeeServiceRequest } from "@/types/service";

const schema = z.object({
  category_id: z.string().uuid("Select a category"),
  title: z.string().min(1, "Title is required").max(150),
  description: z.string().max(1000).optional(),
  price: z.coerce.number().min(0.01, "Price must be greater than 0"),
  duration_minutes: z.coerce.number().int().min(1, "Duration is required"),
  is_active: z.boolean().optional(),
});

type FormValues = z.input<typeof schema>;

export default function EditServicePage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data: services, isLoading } = useServices();
  const { data: categories, isLoading: loadingCategories } = useCategories();
  const { update } = useServiceMutations();

  const service = services?.find((s) => s.id === id);

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      category_id: "",
      title: "",
      description: "",
      price: undefined,
      duration_minutes: undefined,
      is_active: false,
    },
  });

  const { reset } = form;
  useEffect(() => {
    if (!service) return;
    reset({
      category_id: service.category_id,
      title: service.title,
      description: service.description ?? "",
      price: service.price,
      duration_minutes: service.duration_minutes,
      is_active: service.is_active,
    });
  }, [service, reset]);

  if (isLoading) {
    return (
      <div className="p-6 md:p-8">
        <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">
          Loading service...
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

  const onSubmit = (values: FormValues) => {
    const parsed = schema.parse(values);
    const body: EmployeeServiceRequest = {
      category_id: parsed.category_id,
      title: parsed.title,
      description: parsed.description || null,
      price: parsed.price,
      duration_minutes: parsed.duration_minutes,
      is_active: parsed.is_active ?? false,
    };
    update.mutate(
      { id, body },
      {
        onSuccess: () => router.push(`/services/${id}`),
      },
    );
  };

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Edit service</h1>
        <p className="text-sm text-muted-foreground">Update your service.</p>
      </div>

      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Service details</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="grid gap-4" onSubmit={form.handleSubmit(onSubmit)}>
            <div className="space-y-2">
              <Label htmlFor="category_id">Category</Label>
              <select
                id="category_id"
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                disabled={loadingCategories}
                {...form.register("category_id")}
              >
                <option value="">Select a category</option>
                {categories?.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
              {form.formState.errors.category_id && (
                <p className="text-sm text-destructive">
                  {form.formState.errors.category_id.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="title">Title</Label>
              <Input id="title" {...form.register("title")} />
              {form.formState.errors.title && (
                <p className="text-sm text-destructive">
                  {form.formState.errors.title.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Input id="description" {...form.register("description")} />
            </div>

            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="price">Price</Label>
                <Input
                  id="price"
                  type="number"
                  min={0}
                  step="0.01"
                  {...form.register("price")}
                />
                {form.formState.errors.price && (
                  <p className="text-sm text-destructive">
                    {form.formState.errors.price.message}
                  </p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="duration_minutes">Duration (minutes)</Label>
                <Input
                  id="duration_minutes"
                  type="number"
                  min={1}
                  {...form.register("duration_minutes")}
                />
                {form.formState.errors.duration_minutes && (
                  <p className="text-sm text-destructive">
                    {form.formState.errors.duration_minutes.message}
                  </p>
                )}
              </div>
            </div>

            <label className="flex items-center gap-2 text-sm">
              <input type="checkbox" {...form.register("is_active")} />
              Active (visible to customers)
            </label>

            {update.isError && (
              <p className="text-sm text-destructive">
                {(update.error as Error)?.message || "Failed to update service"}
              </p>
            )}

            <div className="flex gap-2">
              <Button type="submit" disabled={update.isPending}>
                {update.isPending ? "Saving..." : "Save changes"}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => router.push(`/services/${id}`)}
              >
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
