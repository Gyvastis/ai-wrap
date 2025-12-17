"use client";

import { useState } from "react";
import { RequestLog } from "@/types";
import { ChevronDown, ChevronUp, Loader2 } from "lucide-react";
import { RequestHeader } from "./RequestHeader";
import { RequestDetails } from "./RequestDetails";

interface RequestItemProps {
  request: RequestLog;
  isExpanded: boolean;
  onToggle: () => void;
}

export function RequestItem({ request, isExpanded, onToggle }: RequestItemProps) {
  const [details, setDetails] = useState<RequestLog | null>(null);
  const [loading, setLoading] = useState(false);

  const handleToggle = async () => {
    if (!isExpanded && !details) {
      setLoading(true);
      try {
        const res = await fetch(`/api/admin/requests/${request.ID}`, { cache: "no-store" });
        if (res.ok) {
          const data = await res.json();
          setDetails(data);
        }
      } catch (error) {
        console.error("Failed to fetch request details:", error);
      } finally {
        setLoading(false);
      }
    }
    onToggle();
  };

  return (
    <div className="border border-gray-200 rounded-lg bg-white hover:border-gray-300 transition-colors">
      <div
        className="p-4 cursor-pointer flex items-start justify-between gap-4"
        onClick={handleToggle}
      >
        <RequestHeader request={request} />

        {loading ? (
          <Loader2 className="w-4 h-4 text-gray-400 flex-shrink-0 mt-1 animate-spin" />
        ) : isExpanded ? (
          <ChevronUp className="w-4 h-4 text-gray-400 flex-shrink-0 mt-1" />
        ) : (
          <ChevronDown className="w-4 h-4 text-gray-400 flex-shrink-0 mt-1" />
        )}
      </div>

      {isExpanded && details && <RequestDetails request={details} />}
    </div>
  );
}
