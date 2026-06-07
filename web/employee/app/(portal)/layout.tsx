import { PortalSidebar } from "@/components/layout/portal-sidebar";

export default function PortalLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen">
      <PortalSidebar />
      <div className="flex-1 overflow-auto">{children}</div>
    </div>
  );
}
