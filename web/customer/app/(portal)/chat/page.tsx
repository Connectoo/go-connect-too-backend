"use client";

import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useConversations } from "@/hooks/use-chat";

export default function ChatPage() {
  const { data: conversations, isLoading } = useConversations();

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Messages</h1>
        <p className="text-sm text-muted-foreground">
          Chat with your service providers.
        </p>
      </div>

      {isLoading && (
        <div className="space-y-3">
          <Skeleton className="h-16 w-full" />
          <Skeleton className="h-16 w-full" />
        </div>
      )}

      {!isLoading && (!conversations || conversations.length === 0) && (
        <Card>
          <CardContent className="py-10 text-center text-sm text-muted-foreground">
            No conversations yet. Messages appear when you have an active booking.
          </CardContent>
        </Card>
      )}

      <div className="space-y-2">
        {conversations?.map((conv) => (
          <Link key={conv.id} href={`/chat/${conv.id}`}>
            <Card className="transition-colors hover:bg-muted/50">
              <CardContent className="flex items-center justify-between py-4">
                <div>
                  <p className="font-medium">Conversation</p>
                  <p className="text-xs text-muted-foreground">
                    Updated {new Date(conv.updated_at).toLocaleString()}
                  </p>
                </div>
                {conv.booking_id && (
                  <span className="text-xs text-muted-foreground">
                    Booking linked
                  </span>
                )}
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
