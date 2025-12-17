"use client";

import { useEffect, useState } from "react";
import { Stats as StatsType, RequestsResponse, TimeSeriesData } from "@/types";
import { Stats } from "@/components/Stats";
import { RequestsList } from "@/components/RequestsList";
import { Pagination } from "@/components/Pagination";
import { DurationToggle } from "@/components/DurationToggle";
import { RequestsChart } from "@/components/RequestsChart";
import { RefreshCw } from "lucide-react";

export default function Home() {
  const [duration, setDuration] = useState<"24h" | "7d">("24h");
  const [stats, setStats] = useState<StatsType | null>(null);
  const [requestsData, setRequestsData] = useState<RequestsResponse | null>(null);
  const [timeSeriesData, setTimeSeriesData] = useState<TimeSeriesData[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  useEffect(() => {
    fetchStats();
    fetchTimeSeries();
  }, [duration]);

  useEffect(() => {
    fetchRequests(currentPage);
  }, [currentPage]);

  const fetchStats = async () => {
    try {
      const res = await fetch(`/api/admin/stats?duration=${duration}`, { cache: "no-store" });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setStats(data);
    } catch (error) {
      console.error("Failed to fetch stats:", error);
      setStats(null);
    }
  };

  const fetchTimeSeries = async () => {
    try {
      const res = await fetch(`/api/admin/timeseries?duration=${duration}`, { cache: "no-store" });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setTimeSeriesData(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error("Failed to fetch timeseries:", error);
      setTimeSeriesData([]);
    }
  };

  const fetchRequests = async (page: number) => {
    setLoading(true);
    try {
      const res = await fetch(`/api/admin/requests?page=${page}&per_page=20`, { cache: "no-store" });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setRequestsData(data);
    } catch (error) {
      console.error("Failed to fetch requests:", error);
      setRequestsData(null);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    await Promise.all([fetchStats(), fetchTimeSeries(), fetchRequests(currentPage)]);
    setRefreshing(false);
  };

  return (
    <div className="min-h-screen bg-white">
      <div className="border-b border-gray-200">
        <div className="max-w-6xl mx-auto px-6 py-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold">AI Wrap Admin</h1>
            <p className="text-sm text-gray-600 mt-1">Gemini proxy analytics</p>
          </div>
          <div className="flex items-center gap-4">
            <DurationToggle duration={duration} onChange={setDuration} />
            <button
              onClick={handleRefresh}
              disabled={refreshing}
              className="flex items-center gap-2 px-3 py-1.5 text-sm border border-gray-200 rounded hover:bg-gray-50 disabled:opacity-50"
            >
              <RefreshCw className={`w-4 h-4 ${refreshing ? "animate-spin" : ""}`} />
              Refresh
            </button>
          </div>
        </div>
      </div>

      <div className="max-w-6xl mx-auto px-6 py-6 space-y-6">
        {stats && <Stats stats={stats} />}

        <RequestsChart data={timeSeriesData} duration={duration} />

        <div>
          <h2 className="text-lg font-semibold mb-4">Requests</h2>

          {loading ? (
            <div className="text-center py-12 text-gray-500">Loading...</div>
          ) : requestsData && requestsData.requests.length > 0 ? (
            <>
              <RequestsList requests={requestsData.requests} />
              <Pagination
                currentPage={requestsData.page}
                totalPages={requestsData.total_pages}
                onPageChange={setCurrentPage}
              />
            </>
          ) : (
            <div className="text-center py-12 text-gray-500">No requests found</div>
          )}
        </div>
      </div>
    </div>
  );
}
