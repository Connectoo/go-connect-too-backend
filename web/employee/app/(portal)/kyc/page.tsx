"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useKYCMutations, useKYC } from "@/hooks/use-kyc";

const schema = z.object({
  id_proof_url: z.string().url("Enter a valid ID proof URL"),
  address_proof_url: z.string().url("Enter a valid address proof URL"),
});

type FormValues = z.infer<typeof schema>;

export default function KYCPage() {
  const { data, isLoading, isError } = useKYC();
  const { submit } = useKYCMutations();

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { id_proof_url: "", address_proof_url: "" },
  });

  const showForm = isError || !data || data.status === "rejected";
  const showPending = data?.status === "pending";
  const showApproved = data?.status === "approved";

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">KYC verification</h1>
          <p className="text-sm text-muted-foreground">
            Submit identity documents for admin review.
          </p>
        </div>
        {data && (
          <Badge
            variant={
              data.status === "approved"
                ? "default"
                : data.status === "rejected"
                  ? "outline"
                  : "secondary"
            }
            className={data.status === "rejected" ? "border-destructive text-destructive" : undefined}
          >
            {data.status}
          </Badge>
        )}
      </div>

      {isLoading && (
        <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">
          Loading KYC status...
        </div>
      )}

      {showApproved && (
        <Card className="max-w-xl">
          <CardContent className="py-8 text-center text-sm text-muted-foreground">
            Your KYC has been approved. You&apos;re all set to receive bookings.
          </CardContent>
        </Card>
      )}

      {showPending && (
        <Card className="max-w-xl">
          <CardContent className="py-8 text-center text-sm text-muted-foreground">
            Your documents are under review. We&apos;ll notify you once approved.
          </CardContent>
        </Card>
      )}

      {showForm && (
        <Card className="max-w-xl">
          <CardHeader>
            <CardTitle>
              {data?.status === "rejected" ? "Resubmit documents" : "Submit documents"}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {data?.status === "rejected" && data.rejection_reason && (
              <p className="mb-4 rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
                Rejected: {data.rejection_reason}
              </p>
            )}

            <form
              className="grid gap-4"
              onSubmit={form.handleSubmit((values) => submit.mutate(values))}
            >
              <div className="space-y-2">
                <Label htmlFor="id_proof_url">ID proof URL</Label>
                <Input
                  id="id_proof_url"
                  placeholder="https://..."
                  {...form.register("id_proof_url")}
                />
                {form.formState.errors.id_proof_url && (
                  <p className="text-sm text-destructive">
                    {form.formState.errors.id_proof_url.message}
                  </p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="address_proof_url">Address proof URL</Label>
                <Input
                  id="address_proof_url"
                  placeholder="https://..."
                  {...form.register("address_proof_url")}
                />
                {form.formState.errors.address_proof_url && (
                  <p className="text-sm text-destructive">
                    {form.formState.errors.address_proof_url.message}
                  </p>
                )}
              </div>

              <p className="text-xs text-muted-foreground">
                Upload files via presigned URL (when storage is configured) or paste
                publicly accessible document URLs.
              </p>

              {submit.isError && (
                <p className="text-sm text-destructive">
                  {(submit.error as Error)?.message || "Failed to submit KYC"}
                </p>
              )}
              {submit.isSuccess && (
                <p className="text-sm text-green-600">Documents submitted for review.</p>
              )}

              <Button type="submit" disabled={submit.isPending}>
                {submit.isPending ? "Submitting..." : "Submit for review"}
              </Button>
            </form>
          </CardContent>
        </Card>
      )}

      <div className="mt-6">
        <Button variant="ghost" asChild className="px-0">
          <Link href="/onboarding">← Back to onboarding</Link>
        </Button>
      </div>
    </div>
  );
}
