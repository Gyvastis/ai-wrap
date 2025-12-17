import { RequestLog } from "@/types";
import { Thermometer, Hash, Database, Key, ArrowRight, ArrowLeft, AlertCircle } from "lucide-react";

interface RequestDetailsProps {
  request: RequestLog;
}

export function RequestDetails({ request }: RequestDetailsProps) {
  return (
    <div className="border-t border-gray-200 p-4 bg-gray-50 space-y-4">
      <div className="grid grid-cols-4 gap-4 text-sm">
        <div>
          <div className="flex items-center gap-1.5 text-gray-500 mb-1">
            <Thermometer className="w-3.5 h-3.5" />
            <span>Temp</span>
          </div>
          <div>{request.Temperature}</div>
        </div>
        <div>
          <div className="flex items-center gap-1.5 text-gray-500 mb-1">
            <Hash className="w-3.5 h-3.5" />
            <span>Tokens</span>
          </div>
          <div>{request.TotalTokens}</div>
        </div>
        <div>
          <div className="flex items-center gap-1.5 text-gray-500 mb-1">
            <Database className="w-3.5 h-3.5" />
            <span>Cache</span>
          </div>
          <div>{request.CacheHit ? "Yes" : "No"}</div>
        </div>
        <div>
          <div className="flex items-center gap-1.5 text-gray-500 mb-1">
            <Key className="w-3.5 h-3.5" />
            <span>Key</span>
          </div>
          <div>{request.KeySource}</div>
        </div>
      </div>

      <div>
        <div className="flex items-center gap-1.5 text-xs font-medium text-gray-500 mb-2">
          <ArrowRight className="w-3.5 h-3.5" />
          <span>Request</span>
        </div>
        <pre className="text-xs bg-white border border-gray-200 rounded p-3 whitespace-pre-wrap break-words max-h-96 overflow-y-auto">
          {JSON.stringify(request.Request, null, 2)}
        </pre>
      </div>

      {request.Response && (
        <div>
          <div className="flex items-center gap-1.5 text-xs font-medium text-gray-500 mb-2">
            <ArrowLeft className="w-3.5 h-3.5" />
            <span>Response</span>
          </div>
          <pre className="text-xs bg-white border border-gray-200 rounded p-3 whitespace-pre-wrap break-words max-h-96 overflow-y-auto">
            {JSON.stringify(request.Response, null, 2)}
          </pre>
        </div>
      )}

      {request.Error && (
        <div>
          <div className="flex items-center gap-1.5 text-xs font-medium text-gray-500 mb-2">
            <AlertCircle className="w-3.5 h-3.5" />
            <span>Error</span>
          </div>
          <pre className="text-xs bg-white border border-gray-200 rounded p-3 whitespace-pre-wrap break-words">
            {request.Error}
          </pre>
        </div>
      )}
    </div>
  );
}
