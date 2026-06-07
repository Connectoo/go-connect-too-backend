"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useSubmitReport } from "@/hooks/use-report";

type FormValues = {
  reported_user_id: string;
  booking_id: string;
  reason: string;
  description: string;
};

export function ReportForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const submit = useSubmitReport();

  const form = useForm<FormValues>({
    defaultValues: {
      reported_user_id: searchParams.get("userId") ?? "",
      booking_id: searchParams.get("bookingId") ?? "",
      reason: "",
      description: "",
    },
  });

  return (
    <Card className="max-w-lg">
      <CardHeader>
        <CardTitle>Report a user or issue</CardTitle>
      </CardHeader>
      <CardContent>
        {submit.isSuccess ? (
          <div className="space-y-4">
            <p className="text-sm text-green-600">Report submitted. Thank you.</p>
            <Button asChild variant="outline">
              <Link href="/dashboard">Back to dashboard</Link>
            </Button>
          </div>
        ) : (
          <form
            className="space-y-4"
            onSubmit={form.handleSubmit((values) =>
              submit.mutate(
                {
                  reported_user_id: values.reported_user_id,
                  booking_id: values.booking_id || undefined,
                  reason: values.reason,
                  description: values.description || undefined,
                },
                { onSuccess: () => router.refresh() },
              ),
            )}
          >
            <div className="space-y-2">
              <Label htmlFor="reported_user_id">Reported user ID</Label>
              <Input
                id="reported_user_id"
                {...form.register("reported_user_id", { required: true })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="booking_id">Booking ID (optional)</Label>
              <Input id="booking_id" {...form.register("booking_id")} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="reason">Reason</Label>
              <Input id="reason" {...form.register("reason", { required: true })} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">Details (optional)</Label>
              <textarea
                id="description"
                className="flex min-h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                {...form.register("description")}
              />
            </div>
            {submit.isError && (
              <p className="text-sm text-destructive">
                {(submit.error as Error)?.message || "Failed to submit report"}
              </p>
            )}
            <Button type="submit" disabled={submit.isPending}>
              {submit.isPending ? "Submitting..." : "Submit report"}
            </Button>
          </form>
        )}
      </CardContent>
    </Card>
  );
}
