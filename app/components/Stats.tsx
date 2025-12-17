import { Stats as StatsType } from "@/types";
import { Activity, CheckCircle, XCircle, Database, DollarSign, Clock } from "lucide-react";

interface StatsProps {
  stats: StatsType;
}

export function Stats({ stats }: StatsProps) {
  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3 mb-6">
      <StatCard icon={Activity} label="Total" value={stats.total_requests.toLocaleString()} />
      <StatCard icon={CheckCircle} label="Success" value={stats.successful_requests.toLocaleString()} />
      <StatCard icon={XCircle} label="Failed" value={stats.failed_requests.toLocaleString()} />
      <StatCard icon={Database} label="Cached" value={stats.cache_hits.toLocaleString()} />
      <StatCard icon={DollarSign} label="Cost" value={`$${stats.total_cost.toFixed(6)}`} />
      <StatCard icon={Clock} label="Latency" value={`${stats.avg_response_time_ms}ms`} />
    </div>
  );
}

function StatCard({ icon: Icon, label, value }: { icon: any; label: string; value: string }) {
  return (
    <div className="border border-gray-200 rounded-lg p-4 bg-white">
      <div className="flex items-center gap-2 text-xs text-gray-500 mb-1">
        <Icon className="w-3.5 h-3.5" />
        <span>{label}</span>
      </div>
      <div className="text-xl font-semibold">{value}</div>
    </div>
  );
}
