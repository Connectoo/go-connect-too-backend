"use client";

import Link from "next/link";
import { CheckCircle2, Circle } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAvailability } from "@/hooks/use-availability";
import { useKYC } from "@/hooks/use-kyc";
import { useProfile } from "@/hooks/use-profile";
import { useServices } from "@/hooks/use-services";

type Step = {
  id: string;
  title: string;
  description: string;
  href: string;
  done: boolean;
};

export default function OnboardingPage() {
  const { data: profile } = useProfile();
  const { data: kyc } = useKYC();
  const { data: services } = useServices();
  const { data: availability } = useAvailability();

  const profileDone = Boolean(
    profile?.display_name && profile?.phone && profile?.bio,
  );
  const kycDone = kyc?.status === "approved";
  const servicesDone = (services?.length ?? 0) > 0;
  const availabilityDone = (availability?.length ?? 0) > 0;

  const steps: Step[] = [
    {
      id: "profile",
      title: "Complete your profile",
      description: "Add display name, phone, bio, and location.",
      href: "/profile",
      done: profileDone,
    },
    {
      id: "kyc",
      title: "Submit KYC documents",
      description: "Upload ID and address proof for verification.",
      href: "/kyc",
      done: kycDone,
    },
    {
      id: "services",
      title: "Create a service",
      description: "List at least one service you offer.",
      href: "/services/new",
      done: servicesDone,
    },
    {
      id: "availability",
      title: "Set your availability",
      description: "Add time slots when customers can book you.",
      href: "/availability",
      done: availabilityDone,
    },
  ];

  const completedCount = steps.filter((s) => s.done).length;
  const allDone = completedCount === steps.length;

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Provider onboarding</h1>
          <p className="text-sm text-muted-foreground">
            Complete these steps to start receiving bookings.
          </p>
        </div>
        <Badge variant={allDone ? "default" : "secondary"}>
          {completedCount}/{steps.length} complete
        </Badge>
      </div>

      {allDone && (
        <Card className="mb-6 border-green-200 bg-green-50">
          <CardContent className="py-4 text-sm text-green-800">
            You&apos;re ready to go! Head to your{" "}
            <Link href="/dashboard" className="font-medium underline">
              dashboard
            </Link>{" "}
            to view bookings.
          </CardContent>
        </Card>
      )}

      <div className="grid gap-4">
        {steps.map((step, index) => (
          <Card key={step.id}>
            <CardHeader className="flex flex-row items-start gap-4 space-y-0">
              <div className="mt-0.5">
                {step.done ? (
                  <CheckCircle2 className="h-5 w-5 text-green-600" />
                ) : (
                  <Circle className="h-5 w-5 text-muted-foreground" />
                )}
              </div>
              <div className="flex-1">
                <CardTitle className="text-base">
                  Step {index + 1}: {step.title}
                </CardTitle>
                <p className="mt-1 text-sm text-muted-foreground">{step.description}</p>
              </div>
              {!step.done && (
                <Button size="sm" asChild>
                  <Link href={step.href}>Start</Link>
                </Button>
              )}
            </CardHeader>
          </Card>
        ))}
      </div>
    </div>
  );
}
