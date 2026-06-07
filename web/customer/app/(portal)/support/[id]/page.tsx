"use client";

import Link from "next/link";
import { use } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useSupportTickets } from "@/hooks/use-support";

export default function SupportTicketPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const { data: tickets, isLoading } = useSupportTickets();
  const ticket = tickets?.find((t) => t.id === id);

  if (isLoading) {
    return (
      <div className="p-6 md:p-8">
        <Skeleton className="h-48 max-w-lg" />
      </div>
    );
  }

  if (!ticket) {
    return (
      <div className="p-6 md:p-8">
        <p className="text-sm text-muted-foreground">Ticket not found.</p>
        <Button asChild className="mt-4" variant="outline">
          <Link href="/support">Back to support</Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="p-6 md:p-8">
      <Button asChild variant="ghost" size="sm">
        <Link href="/support">← Back</Link>
      </Button>
      <h1 className="mt-2 text-2xl font-bold">{ticket.subject}</h1>
      <div className="mt-2 flex gap-2">
        <Badge variant="secondary">{ticket.status}</Badge>
        <Badge variant="outline">{ticket.priority}</Badge>
      </div>

      <Card className="mt-6 max-w-lg">
        <CardHeader>
          <CardTitle>Ticket details</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <p>
            <span className="text-muted-foreground">Created:</span>{" "}
            {new Date(ticket.created_at).toLocaleString()}
          </p>
          <p>
            <span className="text-muted-foreground">Last updated:</span>{" "}
            {new Date(ticket.updated_at).toLocaleString()}
          </p>
          <p className="pt-2 text-muted-foreground">
            Our team will respond to your ticket by email. Check your inbox for
            updates from support.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
