"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useProfileMutations } from "@/hooks/use-profile";

const schema = z
  .object({
    current_password: z.string().min(1, "Current password is required"),
    new_password: z.string().min(8, "Password must be at least 8 characters"),
    confirm_password: z.string().min(1, "Please confirm your password"),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: "Passwords do not match",
    path: ["confirm_password"],
  });

type FormValues = z.infer<typeof schema>;

export default function SecuritySettingsPage() {
  const { changePassword } = useProfileMutations();

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      current_password: "",
      new_password: "",
      confirm_password: "",
    },
  });

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <Link
          href="/profile"
          className="mb-2 inline-block text-sm text-muted-foreground hover:text-foreground"
        >
          ← Back to profile
        </Link>
        <h1 className="text-2xl font-bold">Security</h1>
        <p className="text-sm text-muted-foreground">Change your password.</p>
      </div>

      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>Change password</CardTitle>
        </CardHeader>
        <CardContent>
          <form
            className="grid gap-4"
            onSubmit={form.handleSubmit((values) =>
              changePassword.mutate(
                {
                  current_password: values.current_password,
                  new_password: values.new_password,
                },
                {
                  onSuccess: () => form.reset(),
                },
              ),
            )}
          >
            <div className="space-y-2">
              <Label htmlFor="current_password">Current password</Label>
              <Input
                id="current_password"
                type="password"
                autoComplete="current-password"
                {...form.register("current_password")}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="new_password">New password</Label>
              <Input
                id="new_password"
                type="password"
                autoComplete="new-password"
                {...form.register("new_password")}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirm_password">Confirm new password</Label>
              <Input
                id="confirm_password"
                type="password"
                autoComplete="new-password"
                {...form.register("confirm_password")}
              />
              {form.formState.errors.confirm_password && (
                <p className="text-sm text-destructive">
                  {form.formState.errors.confirm_password.message}
                </p>
              )}
            </div>

            {changePassword.isError && (
              <p className="text-sm text-destructive">
                {(changePassword.error as Error)?.message || "Failed to change password"}
              </p>
            )}
            {changePassword.isSuccess && (
              <p className="text-sm text-green-600">Password updated.</p>
            )}

            <Button type="submit" disabled={changePassword.isPending}>
              {changePassword.isPending ? "Updating..." : "Update password"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
