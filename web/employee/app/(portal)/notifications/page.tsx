"use client";

import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useNotificationMutations, useNotifications } from "@/hooks/use-notifications";
import { cn } from "@/lib/utils";

export default function NotificationsPage() {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useNotifications(page);
  const { markRead, markAllRead } = useNotificationMutations();

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Notifications</h1>
          <p className="text-sm text-muted-foreground">
            Booking updates and platform alerts.
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          disabled={markAllRead.isPending}
          onClick={() => markAllRead.mutate()}
        >
          Mark all read
        </Button>
      </div>

      {isLoading && (
        <div className="space-y-3">
          <Skeleton className="h-20 w-full" />
          <Skeleton className="h-20 w-full" />
        </div>
      )}

      {!isLoading && (!data?.items.length) && (
        <Card>
          <CardContent className="py-10 text-center text-sm text-muted-foreground">
            No notifications yet.
          </CardContent>
        </Card>
      )}

      <div className="space-y-3">
        {data?.items.map((item) => (
          <Card
            key={item.id}
            className={cn(!item.read_at && "border-primary/40 bg-primary/5")}
          >
            <CardContent className="flex items-start justify-between gap-4 pt-6">
              <div>
                <div className="mb-1 flex items-center gap-2">
                  <p className="font-medium">{item.title}</p>
                  {!item.read_at && <Badge variant="secondary">New</Badge>}
                </div>
                <p className="text-sm text-muted-foreground">{item.body}</p>
                <p className="mt-2 text-xs text-muted-foreground">
                  {new Date(item.created_at).toLocaleString()}
                </p>
              </div>
              {!item.read_at && (
                <Button
                  size="sm"
                  variant="ghost"
                  disabled={markRead.isPending}
                  onClick={() => markRead.mutate(item.id)}
                >
                  Mark read
                </Button>
              )}
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
