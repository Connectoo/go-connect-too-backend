"use client";

import { CategoryCard } from "@/components/categories/category-card";
import { Skeleton } from "@/components/ui/skeleton";
import { useCategories } from "@/hooks/use-public";

export default function CategoriesPage() {
  const { data, isLoading, isError } = useCategories();

  return (
    <div className="container mx-auto px-4 py-10">
      <div className="mb-8 max-w-2xl">
        <h1 className="text-3xl font-bold">Service categories</h1>
        <p className="mt-2 text-muted-foreground">
          Browse services by category and find the right professional for your needs.
        </p>
      </div>

      {isError && (
        <p className="text-sm text-destructive">Could not load categories. Showing cached data if available.</p>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {isLoading
          ? Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-32" />
            ))
          : data?.map((category) => (
              <CategoryCard key={category.id} category={category} />
            ))}
      </div>
    </div>
  );
}
