"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useProfile, useProfileMutations } from "@/hooks/use-profile";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  phone: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

export default function ProfilePage() {
  const { data, isLoading, isError, error } = useProfile();
  const { update } = useProfileMutations();

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", phone: "" },
  });

  const { reset } = form;
  useEffect(() => {
    if (!data) return;
    reset({
      name: data.name,
      phone: data.phone ?? "",
    });
  }, [data, reset]);

  if (isLoading) {
    return (
      <div className="p-6 md:p-8">
        <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">
          Loading profile...
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-6 md:p-8">
        <div className="rounded-lg border border-destructive/40 p-8 text-center text-sm text-destructive">
          {(error as Error)?.message || "Failed to load profile"}
        </div>
      </div>
    );
  }

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Profile</h1>
          <p className="text-sm text-muted-foreground">Update your account details.</p>
        </div>
        {data && <Badge variant="secondary">{data.status}</Badge>}
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Personal information</CardTitle>
          </CardHeader>
          <CardContent>
            <form
              className="grid max-w-md gap-4"
              onSubmit={form.handleSubmit((values) =>
                update.mutate({
                  name: values.name,
                  phone: values.phone || null,
                }),
              )}
            >
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input id="email" value={data?.email ?? ""} disabled />
              </div>
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input id="name" {...form.register("name")} />
                {form.formState.errors.name && (
                  <p className="text-sm text-destructive">
                    {form.formState.errors.name.message}
                  </p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="phone">Phone</Label>
                <Input id="phone" {...form.register("phone")} />
              </div>

              {update.isError && (
                <p className="text-sm text-destructive">
                  {(update.error as Error)?.message || "Failed to save profile"}
                </p>
              )}
              {update.isSuccess && (
                <p className="text-sm text-green-600">Profile saved.</p>
              )}

              <Button type="submit" disabled={update.isPending}>
                {update.isPending ? "Saving..." : "Save changes"}
              </Button>
            </form>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Settings</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <Button variant="outline" className="w-full justify-start" asChild>
              <Link href="/settings/security">Security</Link>
            </Button>
            <Button variant="outline" className="w-full justify-start" asChild>
              <Link href="/settings/account">Account</Link>
            </Button>
            <Button variant="outline" className="w-full justify-start" asChild>
              <Link href="/settings/notifications">Notifications</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
