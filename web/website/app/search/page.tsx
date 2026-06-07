import { Suspense } from "react";
import { SearchPageContent } from "@/components/search/search-page-content";
import { Skeleton } from "@/components/ui/skeleton";

export default function SearchPage() {
  return (
    <Suspense
      fallback={
        <div className="container mx-auto px-4 py-10">
          <Skeleton className="mb-4 h-10 w-64" />
          <Skeleton className="h-32 w-full max-w-xl" />
        </div>
      }
    >
      <SearchPageContent />
    </Suspense>
  );
}
