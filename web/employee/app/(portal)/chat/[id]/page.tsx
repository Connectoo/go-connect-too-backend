"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { FormEvent, useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { useChatMutations, useMessages } from "@/hooks/use-chat";
import { loadEmployeeAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";

export default function ChatThreadPage() {
  const params = useParams<{ id: string }>();
  const conversationId = params.id;
  const { data, isLoading } = useMessages(conversationId);
  const { send } = useChatMutations(conversationId);
  const [text, setText] = useState("");
  const bottomRef = useRef<HTMLDivElement>(null);
  const userId = loadEmployeeAuth()?.user.id;

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [data?.items]);

  const onSubmit = (e: FormEvent) => {
    e.preventDefault();
    const message = text.trim();
    if (!message) return;
    send.mutate({ message }, { onSuccess: () => setText("") });
  };

  return (
    <div className="flex h-[calc(100vh-4rem)] flex-col p-6 md:p-8">
      <Link
        href="/chat"
        className="mb-4 inline-block text-sm text-muted-foreground hover:text-foreground"
      >
        ← Back to messages
      </Link>

      <div className="flex-1 space-y-3 overflow-y-auto rounded-lg border p-4">
        {isLoading && <Skeleton className="h-12 w-2/3" />}

        {data?.items.map((msg) => {
          const isMine = msg.sender_id === userId;
          return (
            <div
              key={msg.id}
              className={cn("flex", isMine ? "justify-end" : "justify-start")}
            >
              <div
                className={cn(
                  "max-w-[75%] rounded-lg px-3 py-2 text-sm",
                  isMine
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-foreground",
                )}
              >
                <p>{msg.message}</p>
                <p className="mt-1 text-[10px] opacity-70">
                  {new Date(msg.created_at).toLocaleTimeString()}
                </p>
              </div>
            </div>
          );
        })}
        <div ref={bottomRef} />
      </div>

      <form onSubmit={onSubmit} className="mt-4 flex gap-2">
        <Input
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Type a message..."
          disabled={send.isPending}
        />
        <Button type="submit" disabled={send.isPending || !text.trim()}>
          Send
        </Button>
      </form>
    </div>
  );
}
