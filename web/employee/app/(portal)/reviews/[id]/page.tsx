"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useMemo, useState } from "react";
import { Star } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { useReviewReply, useReviews } from "@/hooks/use-reviews";

export default function ReviewDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const { data, isLoading } = useReviews(1, 100);
  const reply = useReviewReply(id);
  const [text, setText] = useState("");

  const review = useMemo(
    () => data?.items.find((item) => item.id === id),
    [data, id],
  );

  if (isLoading) {
    return (
      <div className="p-6 md:p-8">
        <Skeleton className="h-64 w-full max-w-xl" />
      </div>
    );
  }

  if (!review) {
    return (
      <div className="p-6 md:p-8">
        <p className="text-sm text-muted-foreground">Review not found.</p>
        <Link href="/reviews" className="mt-4 inline-block text-sm underline">
          Back to reviews
        </Link>
      </div>
    );
  }

  return (
    <div className="p-6 md:p-8">
      <Link
        href="/reviews"
        className="mb-4 inline-block text-sm text-muted-foreground hover:text-foreground"
      >
        ← Back to reviews
      </Link>

      <Card className="max-w-xl">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Star className="h-5 w-5 fill-amber-400 text-amber-400" />
            {review.rating}/5
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {review.comment ? (
            <p className="text-sm">{review.comment}</p>
          ) : (
            <p className="text-sm text-muted-foreground">No comment provided.</p>
          )}

          {review.reply ? (
            <div className="rounded-md border bg-muted/50 p-3">
              <p className="text-xs font-medium text-muted-foreground">Your reply</p>
              <p className="mt-1 text-sm">{review.reply.reply}</p>
            </div>
          ) : (
            <div className="space-y-2">
              <Label htmlFor="reply">Your reply</Label>
              <textarea
                id="reply"
                rows={4}
                value={text}
                onChange={(e) => setText(e.target.value)}
                className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                placeholder="Thank the customer..."
              />
              {reply.isError && (
                <p className="text-sm text-destructive">
                  {(reply.error as Error)?.message || "Failed to send reply"}
                </p>
              )}
              <Button
                disabled={!text.trim() || reply.isPending}
                onClick={() => reply.mutate({ reply: text.trim() })}
              >
                {reply.isPending ? "Sending..." : "Post reply"}
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
