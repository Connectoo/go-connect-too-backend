export default function ContactPage() {
  return (
    <div className="container mx-auto max-w-3xl px-4 py-10">
      <h1 className="text-3xl font-bold">Contact us</h1>
      <p className="mt-4 text-muted-foreground">
        Reach our support team at{" "}
        <a href="mailto:support@goconnect.example" className="text-primary underline">
          support@goconnect.example
        </a>
        .
      </p>
    </div>
  );
}
