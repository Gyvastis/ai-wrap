import { TimeSeriesData } from "@/types";
import { useMemo } from "react";

interface RequestsChartProps {
  data: TimeSeriesData[];
  duration?: "24h" | "7d";
}

function fillMissingBuckets(data: TimeSeriesData[], duration: "24h" | "7d"): TimeSeriesData[] {
  const dataMap = new Map((data || []).map((d) => [d.timestamp, d.count]));
  const filled: TimeSeriesData[] = [];

  const now = new Date();
  const bucketCount = duration === "7d" ? 7 : 24;
  const halfBuckets = Math.floor(bucketCount / 2);

  for (let i = -halfBuckets; i < bucketCount - halfBuckets; i++) {
    let timestamp: string;
    if (duration === "7d") {
      const date = new Date(now);
      date.setDate(date.getDate() + i);
      timestamp = date.toISOString().split("T")[0];
    } else {
      const date = new Date(now);
      date.setHours(date.getHours() + i, 0, 0, 0);
      timestamp = `${date.toISOString().split("T")[0]} ${String(date.getHours()).padStart(2, "0")}:00`;
    }

    filled.push({
      timestamp,
      count: dataMap.get(timestamp) || 0,
    });
  }

  return filled;
}

export function RequestsChart({ data, duration = "24h" }: RequestsChartProps) {
  const filledData = useMemo(() => fillMissingBuckets(data, duration), [data, duration]);
  const maxCount = Math.max(...filledData.map((d) => d.count), 1);
  const chartHeight = 192; // h-48 = 12rem = 192px

  return (
    <div className="border border-gray-200 rounded-lg p-4 bg-white">
      <div className="text-sm font-medium mb-4">Requests over time</div>
      <div className="flex items-end gap-1 h-48">
        {filledData.map((item, idx) => {
          const heightPx = item.count === 0
            ? 4
            : Math.max((item.count / maxCount) * chartHeight, 8);

          return (
            <div key={idx} className="flex-1 flex flex-col justify-end group relative min-w-0">
              <div className="absolute -top-8 left-1/2 -translate-x-1/2 opacity-0 group-hover:opacity-100 text-xs bg-black text-white px-2 py-1 rounded whitespace-nowrap z-10">
                {item.count} requests
              </div>
              <div
                className={`hover:bg-gray-600 transition-colors rounded-t ${
                  item.count === 0 ? "bg-gray-300" : "bg-black"
                }`}
                style={{ height: `${heightPx}px` }}
              />
              <div className="text-xs text-gray-500 mt-2 text-center truncate">
                {item.timestamp.includes(":")
                  ? item.timestamp.split(" ")[1]
                  : new Date(item.timestamp).toLocaleDateString(undefined, {
                      month: "short",
                      day: "numeric",
                    })
                }
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
