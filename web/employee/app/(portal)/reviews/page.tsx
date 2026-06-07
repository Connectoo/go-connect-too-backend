"use client";

import Link from "next/link";
import { Star } from "lucide-react";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useReviews } from "@/hooks/use-reviews";

export default function ReviewsPage() {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useReviews(page);

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Reviews</h1>
        <p className="text-sm text-muted-foreground">
          See what customers say about your services.
        </p>
      </div>

      {isLoading && (
        <div className="space-y-3">
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
        </div>
      )}

      {!isLoading && (!data?.items.length) && (
        <Card>
          <CardContent className="py-10 text-center text-sm text-muted-foreground">
            No reviews yet.
          </CardContent>
        </Card>
      )}

      <div className="space-y-3">
        {data?.items.map((review) => (
          <Card key={review.id}>
            <CardContent className="pt-6">
              <div className="mb-2 flex items-center justify-between">
                <div className="flex items-center gap-1">
                  <Star className="h-4 w-4 fill-amber-400 text-amber-400" />
                  <span className="font-medium">{review.rating}/5</span>
                </div>
                <Badge variant="secondary">{review.status}</Badge>
              </div>
              {review.comment && (
                <p className="text-sm text-muted-foreground">{review.comment}</p>
              )}
              <p className="mt-2 text-xs text-muted-foreground">
                {new Date(review.created_at).toLocaleDateString()}
              </p>
              <Button size="sm" variant="ghost" className="mt-2 px-0" asChild>
                <Link href={`/reviews/${review.id}`}>
                  {review.reply ? "View reply" : "Reply"}
                </Link>
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>

      {data && data.total > data.limit && (
        <div className="mt-4 flex gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={page <= 1}
            onClick={() => setPage((p) => p - 1)}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={page * data.limit >= data.total}
            onClick={() => setPage((p) => p + 1)}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
}
