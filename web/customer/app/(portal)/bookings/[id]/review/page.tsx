"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { useBooking } from "@/hooks/use-bookings";
import { useCreateReview } from "@/hooks/use-reviews";
import { cn } from "@/lib/utils";

export default function ReviewBookingPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const router = useRouter();
  const { data: booking } = useBooking(id);
  const createReview = useCreateReview(id);

  const [rating, setRating] = useState(5);
  const [comment, setComment] = useState("");

  if (booking && booking.status !== "completed") {
    return (
      <div className="p-6 md:p-8">
        <p className="text-sm text-muted-foreground">
          Reviews can only be left for completed bookings.
        </p>
        <Link href={`/bookings/${id}`} className="mt-4 inline-block text-sm underline">
          Back to booking
        </Link>
      </div>
    );
  }

  return (
    <div className="p-6 md:p-8">
      <Link
        href={`/bookings/${id}`}
        className="mb-4 inline-block text-sm text-muted-foreground hover:text-foreground"
      >
        ← Back to booking
      </Link>

      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>Leave a review</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label>Rating</Label>
            <div className="flex gap-2">
              {[1, 2, 3, 4, 5].map((value) => (
                <button
                  key={value}
                  type="button"
                  onClick={() => setRating(value)}
                  className={cn(
                    "flex h-10 w-10 items-center justify-center rounded-md border text-sm font-medium",
                    rating >= value
                      ? "border-primary bg-primary text-primary-foreground"
                      : "hover:bg-muted",
                  )}
                >
                  {value}
                </button>
              ))}
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="comment">Comment (optional)</Label>
            <textarea
              id="comment"
              rows={4}
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              placeholder="Share your experience..."
            />
          </div>

          {createReview.isError && (
            <p className="text-sm text-destructive">
              {(createReview.error as Error)?.message || "Failed to submit review"}
            </p>
          )}

          <Button
            disabled={createReview.isPending}
            onClick={() =>
              createReview.mutate(
                { rating, comment: comment.trim() || undefined },
                { onSuccess: () => router.push(`/bookings/${id}`) },
              )
            }
          >
            {createReview.isPending ? "Submitting..." : "Submit review"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
