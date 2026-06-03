"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { loginCustomer, registerCustomer, saveAuth } from "@/services/auth";

const loginSchema = z.object({
  email: z.string().email("Enter a valid email"),
  password: z.string().min(8, "Password must be at least 8 characters"),
});

const registerSchema = loginSchema.extend({
  name: z.string().min(2, "Name is required"),
  phone: z.string().optional(),
});

type AuthFormProps = {
  mode: "login" | "register";
};

export function AuthForm({ mode }: AuthFormProps) {
  const router = useRouter();

  const loginForm = useForm<z.infer<typeof loginSchema>>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  });

  const registerForm = useForm<z.infer<typeof registerSchema>>({
    resolver: zodResolver(registerSchema),
    defaultValues: { email: "", password: "", name: "", phone: "" },
  });

  const loginMutation = useMutation({
    mutationFn: loginCustomer,
    onSuccess: (data) => {
      if (data.user) saveAuth({ user: data.user, tokens: data.tokens });
      router.push("/");
    },
  });

  const registerMutation = useMutation({
    mutationFn: registerCustomer,
    onSuccess: (data) => {
      if (data.user) saveAuth({ user: data.user, tokens: data.tokens });
      router.push("/");
    },
  });

  if (mode === "login") {
    return (
      <form
        className="space-y-4"
        onSubmit={loginForm.handleSubmit((values) => loginMutation.mutate(values))}
      >
        <Field
          id="email"
          label="Email"
          error={loginForm.formState.errors.email?.message}
          input={<Input type="email" autoComplete="email" {...loginForm.register("email")} />}
        />
        <Field
          id="password"
          label="Password"
          error={loginForm.formState.errors.password?.message}
          input={
            <Input
              type="password"
              autoComplete="current-password"
              {...loginForm.register("password")}
            />
          }
        />
        <ErrorMessage error={loginMutation.error} />
        <Button type="submit" className="w-full" disabled={loginMutation.isPending}>
          {loginMutation.isPending ? "Please wait..." : "Sign in"}
        </Button>
        <Footer mode="login" />
      </form>
    );
  }

  return (
    <form
      className="space-y-4"
      onSubmit={registerForm.handleSubmit((values) =>
        registerMutation.mutate({
          name: values.name,
          email: values.email,
          password: values.password,
          phone: values.phone || undefined,
        }),
      )}
    >
      <Field
        id="name"
        label="Full name"
        error={registerForm.formState.errors.name?.message}
        input={<Input {...registerForm.register("name")} />}
      />
      <Field
        id="email"
        label="Email"
        error={registerForm.formState.errors.email?.message}
        input={<Input type="email" autoComplete="email" {...registerForm.register("email")} />}
      />
      <Field
        id="phone"
        label="Phone (optional)"
        input={<Input {...registerForm.register("phone")} />}
      />
      <Field
        id="password"
        label="Password"
        error={registerForm.formState.errors.password?.message}
        input={
          <Input
            type="password"
            autoComplete="new-password"
            {...registerForm.register("password")}
          />
        }
      />
      <ErrorMessage error={registerMutation.error} />
      <Button type="submit" className="w-full" disabled={registerMutation.isPending}>
        {registerMutation.isPending ? "Please wait..." : "Create account"}
      </Button>
      <Footer mode="register" />
    </form>
  );
}

function Field({
  id,
  label,
  input,
  error,
}: {
  id: string;
  label: string;
  input: React.ReactNode;
  error?: string;
}) {
  return (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      {input}
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  );
}

function ErrorMessage({ error }: { error: unknown }) {
  if (!error) return null;
  return (
    <p className="text-sm text-destructive">
      {(error as Error).message || "Authentication failed"}
    </p>
  );
}

function Footer({ mode }: { mode: "login" | "register" }) {
  return (
    <p className="text-center text-sm text-muted-foreground">
      {mode === "login" ? (
        <>
          No account?{" "}
          <Link href="/register" className="text-primary underline-offset-4 hover:underline">
            Register
          </Link>
        </>
      ) : (
        <>
          Already have an account?{" "}
          <Link href="/login" className="text-primary underline-offset-4 hover:underline">
            Sign in
          </Link>
        </>
      )}
    </p>
  );
}
