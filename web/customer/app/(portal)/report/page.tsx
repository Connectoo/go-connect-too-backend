import { Suspense } from "react";
import { ReportForm } from "@/components/report/report-form";
import { Skeleton } from "@/components/ui/skeleton";

export default function ReportPage() {
  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Report</h1>
        <p className="text-sm text-muted-foreground">
          Flag inappropriate behaviour or content.
        </p>
      </div>
      <Suspense fallback={<Skeleton className="h-64 max-w-lg" />}>
        <ReportForm />
      </Suspense>
    </div>
  );
}
