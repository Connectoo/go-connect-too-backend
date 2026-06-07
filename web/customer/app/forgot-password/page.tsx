"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { forgotPassword } from "@/services/auth";

const schema = z.object({
  email: z.string().email(),
});

export default function ForgotPasswordPage() {
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { email: "" },
  });

  const mutation = useMutation({
    mutationFn: (email: string) => forgotPassword(email),
  });

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Reset your password</CardTitle>
        </CardHeader>
        <CardContent>
          {mutation.isSuccess ? (
            <div className="space-y-4">
              <p className="text-sm text-muted-foreground">
                If an account exists for that email, we sent reset instructions.
              </p>
              <Button asChild variant="outline" className="w-full">
                <Link href="/login">Back to sign in</Link>
              </Button>
            </div>
          ) : (
            <form
              className="space-y-4"
              onSubmit={form.handleSubmit((v) => mutation.mutate(v.email))}
            >
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input id="email" type="email" {...form.register("email")} />
              </div>
              {mutation.isError && (
                <p className="text-sm text-destructive">
                  {(mutation.error as Error)?.message || "Request failed"}
                </p>
              )}
              <Button type="submit" className="w-full" disabled={mutation.isPending}>
                {mutation.isPending ? "Sending..." : "Send reset link"}
              </Button>
              <Button asChild variant="ghost" className="w-full">
                <Link href="/login">Back to sign in</Link>
              </Button>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
