"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { useAdminSettings, useSettingsMutation } from "@/hooks/use-ops";
import type { GeneralSettings, ProviderSettings } from "@/types/ops";

export default function SettingsPage() {
  const { data, isLoading } = useAdminSettings();
  const save = useSettingsMutation();
  const [general, setGeneral] = useState<GeneralSettings>({
    site_name: "",
    support_email: "",
    maintenance_mode: false,
  });
  const [provider, setProvider] = useState<ProviderSettings>({});

  useEffect(() => {
    if (data) {
      setGeneral(data.general);
      setProvider(data.provider ?? {});
    }
  }, [data]);

  if (isLoading) {
    return (
      <div className="p-6 md:p-8">
        <Skeleton className="h-64 max-w-xl" />
      </div>
    );
  }

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-sm text-muted-foreground">Platform configuration.</p>
      </div>

      <form
        className="grid max-w-2xl gap-6"
        onSubmit={(e) => {
          e.preventDefault();
          save.mutate({ general, provider });
        }}
      >
        <Card>
          <CardHeader>
            <CardTitle>General</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="site_name">Site name</Label>
              <Input
                id="site_name"
                value={general.site_name}
                onChange={(e) =>
                  setGeneral((g) => ({ ...g, site_name: e.target.value }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="support_email">Support email</Label>
              <Input
                id="support_email"
                type="email"
                value={general.support_email}
                onChange={(e) =>
                  setGeneral((g) => ({ ...g, support_email: e.target.value }))
                }
              />
            </div>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={general.maintenance_mode}
                onChange={(e) =>
                  setGeneral((g) => ({
                    ...g,
                    maintenance_mode: e.target.checked,
                  }))
                }
              />
              Maintenance mode
            </label>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Providers</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="razorpay_key_id">Razorpay key ID</Label>
              <Input
                id="razorpay_key_id"
                value={provider.razorpay_key_id ?? ""}
                onChange={(e) =>
                  setProvider((p) => ({ ...p, razorpay_key_id: e.target.value }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="smtp_host">SMTP host</Label>
              <Input
                id="smtp_host"
                value={provider.smtp_host ?? ""}
                onChange={(e) =>
                  setProvider((p) => ({ ...p, smtp_host: e.target.value }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="smtp_from">SMTP from</Label>
              <Input
                id="smtp_from"
                value={provider.smtp_from ?? ""}
                onChange={(e) =>
                  setProvider((p) => ({ ...p, smtp_from: e.target.value }))
                }
              />
            </div>
            <p className="text-xs text-muted-foreground">
              Secrets are redacted on load. Leave blank to keep existing values.
            </p>
          </CardContent>
        </Card>

        <Button type="submit" disabled={save.isPending}>
          {save.isPending ? "Saving…" : "Save settings"}
        </Button>
      </form>
    </div>
  );
}
