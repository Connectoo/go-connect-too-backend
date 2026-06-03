import Link from "next/link";
import { Star } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { Provider } from "@/types/public";

export function ProviderCard({ provider }: { provider: Provider }) {
  return (
    <Link href={`/providers/${provider.id}`}>
      <Card className="h-full transition-shadow hover:shadow-md">
        <CardHeader>
          <div className="flex items-start justify-between gap-2">
            <CardTitle className="text-base">
              {provider.display_name ?? "Service Provider"}
            </CardTitle>
            {provider.rating != null && (
              <span className="flex items-center gap-1 text-sm text-muted-foreground">
                <Star className="h-4 w-4 fill-amber-400 text-amber-400" />
                {provider.rating.toFixed(1)}
              </span>
            )}
          </div>
          {provider.location_text && (
            <p className="text-sm text-muted-foreground">{provider.location_text}</p>
          )}
        </CardHeader>
        <CardContent className="space-y-3">
          {provider.bio && (
            <p className="line-clamp-2 text-sm text-muted-foreground">{provider.bio}</p>
          )}
          <div className="flex flex-wrap gap-2">
            {provider.skills.slice(0, 3).map((skill) => (
              <Badge key={skill} variant="secondary">
                {skill}
              </Badge>
            ))}
          </div>
          <p className="text-xs text-muted-foreground">
            {provider.experience_years} yrs experience · {provider.total_reviews} reviews
          </p>
        </CardContent>
      </Card>
    </Link>
  );
}
