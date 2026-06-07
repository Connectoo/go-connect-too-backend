"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { resetPassword } from "@/services/auth";

const schema = z
  .object({
    new_password: z.string().min(8),
    confirm_password: z.string().min(8),
  })
  .refine((d) => d.new_password === d.confirm_password, {
    message: "Passwords do not match",
    path: ["confirm_password"],
  });

export function ResetPasswordForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";

  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { new_password: "", confirm_password: "" },
  });

  const mutation = useMutation({
    mutationFn: (password: string) => resetPassword(token, password),
    onSuccess: () => router.push("/login"),
  });

  if (!token) {
    return (
      <Card className="w-full max-w-md">
        <CardContent className="pt-6">
          <p className="text-sm text-destructive">Invalid or missing reset token.</p>
          <Button asChild className="mt-4" variant="outline">
            <Link href="/forgot-password">Request a new link</Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Set a new password</CardTitle>
      </CardHeader>
      <CardContent>
        <form
          className="space-y-4"
          onSubmit={form.handleSubmit((v) => mutation.mutate(v.new_password))}
        >
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
            <Label htmlFor="confirm_password">Confirm password</Label>
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
          {mutation.isError && (
            <p className="text-sm text-destructive">
              {(mutation.error as Error)?.message || "Reset failed"}
            </p>
          )}
          <Button type="submit" className="w-full" disabled={mutation.isPending}>
            {mutation.isPending ? "Updating..." : "Update password"}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
