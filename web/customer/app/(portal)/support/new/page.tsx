"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useCreateSupportTicket } from "@/hooks/use-support";

type FormValues = {
  subject: string;
  message: string;
};

export default function NewSupportTicketPage() {
  const router = useRouter();
  const create = useCreateSupportTicket();
  const form = useForm<FormValues>({
    defaultValues: { subject: "", message: "" },
  });

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <Link
          href="/support"
          className="mb-2 inline-block text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to support
        </Link>
        <h1 className="text-2xl font-bold">New support ticket</h1>
      </div>

      <Card className="max-w-lg">
        <CardHeader>
          <CardTitle>Describe your issue</CardTitle>
        </CardHeader>
        <CardContent>
          <form
            className="space-y-4"
            onSubmit={form.handleSubmit((values) =>
              create.mutate(values, {
                onSuccess: (ticket) => router.push(`/support/${ticket.id}`),
              }),
            )}
          >
            <div className="space-y-2">
              <Label htmlFor="subject">Subject</Label>
              <Input id="subject" {...form.register("subject", { required: true })} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="message">Message</Label>
              <textarea
                id="message"
                className="flex min-h-32 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                {...form.register("message", { required: true })}
              />
            </div>
            {create.isError && (
              <p className="text-sm text-destructive">
                {(create.error as Error)?.message || "Failed to create ticket"}
              </p>
            )}
            <Button type="submit" disabled={create.isPending}>
              {create.isPending ? "Submitting..." : "Submit ticket"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
