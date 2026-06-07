"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { isCustomerUser, saveCustomerAuth } from "@/lib/auth";
import { loginCustomer } from "@/services/auth";

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
});

function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { email: "", password: "" },
  });

  const mutation = useMutation({
    mutationFn: loginCustomer,
    onSuccess: (data) => {
      if (!data.user || !isCustomerUser(data.user)) {
        form.setError("root", {
          message: "Access denied. Customer account required.",
        });
        return;
      }
      saveCustomerAuth({ user: data.user, tokens: data.tokens });
      router.push(searchParams.get("from") || "/dashboard");
    },
  });

  return (
    <Card className="w-full max-w-md">
      <CardHeader>
        <CardTitle>Sign in</CardTitle>
      </CardHeader>
      <CardContent>
        <form
          className="space-y-4"
          onSubmit={form.handleSubmit((v) => mutation.mutate(v))}
        >
          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input id="email" type="email" {...form.register("email")} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>
            <Input id="password" type="password" {...form.register("password")} />
          </div>
          {(form.formState.errors.root || mutation.isError) && (
            <p className="text-sm text-destructive">
              {form.formState.errors.root?.message ||
                (mutation.error as Error)?.message ||
                "Login failed"}
            </p>
          )}
          <Button type="submit" className="w-full" disabled={mutation.isPending}>
            {mutation.isPending ? "Signing in..." : "Sign in"}
          </Button>
        </form>
        <div className="mt-4 flex items-center justify-between text-sm">
          <Link href="/forgot-password" className="text-muted-foreground hover:underline">
            Forgot password?
          </Link>
          <Link href="/register" className="font-medium hover:underline">
            Create account
          </Link>
        </div>
      </CardContent>
    </Card>
  );
}

export default function CustomerLoginPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 p-4">
      <Suspense fallback={<div className="text-sm text-muted-foreground">Loading...</div>}>
        <LoginForm />
      </Suspense>
    </div>
  );
}
