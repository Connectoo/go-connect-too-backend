"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useProfile, useProfileMutations } from "@/hooks/use-profile";
import type { UpdateEmployeeProfileRequest } from "@/types/profile";

const schema = z.object({
  display_name: z.string().min(1, "Display name is required"),
  phone: z.string().min(1, "Phone is required"),
  bio: z.string().optional(),
  experience_years: z.coerce.number().int().min(0).optional(),
  profile_photo_url: z.string().optional(),
  location_text: z.string().optional(),
  service_area_radius_km: z.coerce.number().min(0).optional(),
  languages: z.string().optional(),
  skills: z.string().optional(),
});

type FormValues = z.input<typeof schema>;

function splitList(value?: string): string[] {
  if (!value) return [];
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

export default function ProfilePage() {
  const { data, isLoading, isError, error } = useProfile();
  const { update } = useProfileMutations();

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      display_name: "",
      phone: "",
      bio: "",
      experience_years: 0,
      profile_photo_url: "",
      location_text: "",
      service_area_radius_km: undefined,
      languages: "",
      skills: "",
    },
  });

  const { reset } = form;
  useEffect(() => {
    if (!data) return;
    reset({
      display_name: data.display_name ?? "",
      phone: data.phone ?? "",
      bio: data.bio ?? "",
      experience_years: data.experience_years ?? 0,
      profile_photo_url: data.profile_photo_url ?? "",
      location_text: data.location_text ?? "",
      service_area_radius_km: data.service_area_radius_km ?? undefined,
      languages: data.languages?.join(", ") ?? "",
      skills: data.skills?.join(", ") ?? "",
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

  const onSubmit = (values: FormValues) => {
    const parsed = schema.parse(values);
    const body: UpdateEmployeeProfileRequest = {
      display_name: parsed.display_name,
      phone: parsed.phone,
      bio: parsed.bio || null,
      experience_years: parsed.experience_years ?? 0,
      profile_photo_url: parsed.profile_photo_url || null,
      location_text: parsed.location_text || null,
      languages: splitList(parsed.languages),
      skills: splitList(parsed.skills),
    };
    if (parsed.service_area_radius_km !== undefined) {
      body.service_area_radius_km = parsed.service_area_radius_km;
    }
    update.mutate(body);
  };

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Business profile</h1>
          <p className="text-sm text-muted-foreground">
            Manage how customers see you.
          </p>
        </div>
        {data && (
          <Badge
            variant={
              data.verification_status === "approved" ? "default" : "secondary"
            }
          >
            {data.verification_status}
          </Badge>
        )}
      </div>

      <Card className="max-w-3xl">
        <CardHeader>
          <CardTitle>Profile details</CardTitle>
        </CardHeader>
        <CardContent>
          <form
            className="grid gap-4 md:grid-cols-2"
            onSubmit={form.handleSubmit(onSubmit)}
          >
            <div className="space-y-2">
              <Label htmlFor="display_name">Display name</Label>
              <Input id="display_name" {...form.register("display_name")} />
              {form.formState.errors.display_name && (
                <p className="text-sm text-destructive">
                  {form.formState.errors.display_name.message}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="phone">Phone</Label>
              <Input id="phone" {...form.register("phone")} />
              {form.formState.errors.phone && (
                <p className="text-sm text-destructive">
                  {form.formState.errors.phone.message}
                </p>
              )}
            </div>

            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="bio">Bio</Label>
              <Input id="bio" {...form.register("bio")} />
            </div>

            <div className="space-y-2">
              <Label htmlFor="experience_years">Experience (years)</Label>
              <Input
                id="experience_years"
                type="number"
                min={0}
                {...form.register("experience_years")}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="service_area_radius_km">
                Service radius (km)
              </Label>
              <Input
                id="service_area_radius_km"
                type="number"
                min={0}
                step="0.1"
                {...form.register("service_area_radius_km")}
              />
            </div>

            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="location_text">Location</Label>
              <Input id="location_text" {...form.register("location_text")} />
            </div>

            <div className="space-y-2">
              <Label htmlFor="skills">Skills (comma separated)</Label>
              <Input id="skills" {...form.register("skills")} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="languages">Languages (comma separated)</Label>
              <Input id="languages" {...form.register("languages")} />
            </div>

            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="profile_photo_url">Photo URL</Label>
              <Input
                id="profile_photo_url"
                placeholder="https://..."
                {...form.register("profile_photo_url")}
              />
            </div>

            {update.isError && (
              <p className="text-sm text-destructive md:col-span-2">
                {(update.error as Error)?.message || "Failed to save profile"}
              </p>
            )}
            {update.isSuccess && (
              <p className="text-sm text-green-600 md:col-span-2">
                Profile saved.
              </p>
            )}

            <div className="md:col-span-2">
              <Button type="submit" disabled={update.isPending}>
                {update.isPending ? "Saving..." : "Save profile"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
