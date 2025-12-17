import { RequestLog } from "@/types";
import { CheckCircle, XCircle, Database, Key, Clock, DollarSign, Globe, Image } from "lucide-react";

interface RequestHeaderProps {
  request: RequestLog;
}

export function RequestHeader({ request }: RequestHeaderProps) {
  return (
    <>
      <div className="flex items-start gap-3 flex-1 min-w-0">
        {request.Success ? (
          <CheckCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
        ) : (
          <XCircle className="w-4 h-4 flex-shrink-0 mt-0.5" />
        )}

        <div className="flex-1 min-w-0">
          <div className="font-medium truncate">{request.Model}</div>
          <div className="text-xs text-gray-500 mt-1">
            {new Date(request.Timestamp).toLocaleString()}
          </div>
        </div>
      </div>

      <div className="flex items-center gap-4 text-sm flex-shrink-0">
        {request.IsVision && (
          <div className="flex items-center gap-1.5 text-blue-600">
            <Image className="w-3.5 h-3.5" />
            <span>vision</span>
          </div>
        )}
        <div className="flex items-center gap-1.5">
          <Database className="w-3.5 h-3.5 text-gray-400" />
          <span>{request.CacheHit ? "Yes" : "No"}</span>
        </div>
        <div className="flex items-center gap-1.5">
          <Key className="w-3.5 h-3.5 text-gray-400" />
          <span>{request.KeySource}</span>
        </div>
        <div className="flex items-center gap-1.5">
          <Globe className="w-3.5 h-3.5 text-gray-400" />
          <span>{request.StatusCode}</span>
        </div>
        <div className="flex items-center gap-1">
          <DollarSign className="w-3.5 h-3.5 text-gray-400" />
          <span className={request.CacheHit ? "line-through" : ""}>
            {request.Cost.Total.toFixed(6)}
          </span>
        </div>
        <div className="flex items-center gap-1">
          <Clock className="w-3.5 h-3.5 text-gray-400" />
          <span>{request.DurationMs}ms</span>
        </div>
      </div>
    </>
  );
}
