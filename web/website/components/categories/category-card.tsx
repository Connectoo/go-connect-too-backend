import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { Category } from "@/types/public";

export function CategoryCard({ category }: { category: Category }) {
  return (
    <Link href={`/providers?category=${category.id}`}>
      <Card className="h-full transition-shadow hover:shadow-md">
        <CardHeader>
          <CardTitle className="text-base">{category.name}</CardTitle>
        </CardHeader>
        {category.description && (
          <CardContent>
            <p className="line-clamp-2 text-sm text-muted-foreground">
              {category.description}
            </p>
          </CardContent>
        )}
      </Card>
    </Link>
  );
}
