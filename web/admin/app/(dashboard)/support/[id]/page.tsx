"use client";

import { use, useState } from "react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { useAdminSupportTicket, useSupportTicketActions } from "@/hooks/use-ops";

export default function SupportTicketPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const { data: ticket, isLoading } = useAdminSupportTicket(id);
  const actions = useSupportTicketActions(id);
  const [reply, setReply] = useState("");

  if (isLoading) {
    return (
      <div className="p-6 md:p-8">
        <Skeleton className="h-64 w-full max-w-2xl" />
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
      <div className="mb-6">
        <Button asChild variant="ghost" size="sm">
          <Link href="/support">← Back</Link>
        </Button>
        <h1 className="mt-2 text-2xl font-bold">{ticket.subject}</h1>
        <div className="mt-2 flex gap-2">
          <Badge variant="secondary">{ticket.status}</Badge>
          <Badge variant="outline">{ticket.priority}</Badge>
        </div>
      </div>

      <div className="grid max-w-2xl gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Update ticket</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-wrap gap-2">
            <select
              className="h-10 rounded-md border border-input bg-background px-3 text-sm"
              value={ticket.status}
              onChange={(e) => actions.update.mutate({ status: e.target.value })}
            >
              <option value="open">Open</option>
              <option value="in_progress">In progress</option>
              <option value="resolved">Resolved</option>
              <option value="closed">Closed</option>
            </select>
            <select
              className="h-10 rounded-md border border-input bg-background px-3 text-sm"
              value={ticket.priority}
              onChange={(e) => actions.update.mutate({ priority: e.target.value })}
            >
              <option value="low">Low</option>
              <option value="normal">Normal</option>
              <option value="high">High</option>
              <option value="urgent">Urgent</option>
            </select>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Messages</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {ticket.messages.map((msg) => (
              <div
                key={msg.id}
                className={`rounded-md border p-3 text-sm ${
                  msg.is_staff ? "bg-muted" : ""
                }`}
              >
                <p className="mb-1 text-xs text-muted-foreground">
                  {msg.is_staff ? "Staff" : "Customer"} ·{" "}
                  {new Date(msg.created_at).toLocaleString()}
                </p>
                <p>{msg.message}</p>
              </div>
            ))}

            <form
              className="flex gap-2"
              onSubmit={(e) => {
                e.preventDefault();
                if (!reply.trim()) return;
                actions.reply.mutate(reply.trim(), {
                  onSuccess: () => setReply(""),
                });
              }}
            >
              <Input
                value={reply}
                onChange={(e) => setReply(e.target.value)}
                placeholder="Staff reply…"
                className="flex-1"
              />
              <Button type="submit" disabled={actions.reply.isPending}>
                Send
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
