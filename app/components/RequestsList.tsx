"use client";

import { useState } from "react";
import { RequestLog } from "@/types";
import { RequestItem } from "./RequestItem";

interface RequestsListProps {
  requests: RequestLog[];
}

export function RequestsList({ requests }: RequestsListProps) {
  const [expandedId, setExpandedId] = useState<string | null>(null);

  return (
    <div className="space-y-2">
      {requests.map((req) => (
        <RequestItem
          key={req.ID}
          request={req}
          isExpanded={expandedId === req.ID}
          onToggle={() => setExpandedId(expandedId === req.ID ? null : req.ID)}
        />
      ))}
    </div>
  );
}
