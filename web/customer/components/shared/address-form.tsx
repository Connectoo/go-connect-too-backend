"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { AddressInput } from "@/types/address";

const schema = z.object({
  label: z.string().min(1, "Label is required"),
  address_line: z.string().min(1, "Address is required"),
  city: z.string().min(1, "City is required"),
  state: z.string().min(1, "State is required"),
  country: z.string().min(1, "Country is required"),
  pincode: z.string().min(1, "Pincode is required"),
  is_default: z.boolean(),
});

type FormValues = z.infer<typeof schema>;

type AddressFormProps = {
  defaultValues?: Partial<FormValues>;
  submitLabel: string;
  loading?: boolean;
  error?: string | null;
  onSubmit: (values: AddressInput) => void;
  cancelHref?: string;
};

export function AddressForm({
  defaultValues,
  submitLabel,
  loading,
  error,
  onSubmit,
  cancelHref = "/addresses",
}: AddressFormProps) {
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      label: "",
      address_line: "",
      city: "",
      state: "",
      country: "India",
      pincode: "",
      is_default: false,
      ...defaultValues,
    },
  });

  return (
    <form
      className="grid max-w-xl gap-4"
      onSubmit={form.handleSubmit((values) => onSubmit(values))}
    >
      <div className="space-y-2">
        <Label htmlFor="label">Label</Label>
        <Input id="label" placeholder="Home, Office..." {...form.register("label")} />
        {form.formState.errors.label && (
          <p className="text-sm text-destructive">{form.formState.errors.label.message}</p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="address_line">Address</Label>
        <Input id="address_line" {...form.register("address_line")} />
        {form.formState.errors.address_line && (
          <p className="text-sm text-destructive">
            {form.formState.errors.address_line.message}
          </p>
        )}
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="city">City</Label>
          <Input id="city" {...form.register("city")} />
        </div>
        <div className="space-y-2">
          <Label htmlFor="state">State</Label>
          <Input id="state" {...form.register("state")} />
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="country">Country</Label>
          <Input id="country" {...form.register("country")} />
        </div>
        <div className="space-y-2">
          <Label htmlFor="pincode">Pincode</Label>
          <Input id="pincode" {...form.register("pincode")} />
        </div>
      </div>

      <label className="flex items-center gap-2 text-sm">
        <input type="checkbox" {...form.register("is_default")} />
        Set as default address
      </label>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <div className="flex gap-3">
        <Button type="submit" disabled={loading}>
          {loading ? "Saving..." : submitLabel}
        </Button>
        <Button type="button" variant="outline" asChild>
          <Link href={cancelHref}>Cancel</Link>
        </Button>
      </div>
    </form>
  );
}
