export interface Stats {
  total_requests: number;
  successful_requests: number;
  failed_requests: number;
  cache_hits: number;
  total_cost: number;
  avg_response_time_ms: number;
}

export interface RequestLog {
  ID: string;
  Timestamp: string;
  Model: string;
  Request: any;
  Response?: any;
  StatusCode: number;
  Success: boolean;
  Error: string;
  Cost: {
    Input: number;
    Output: number;
    Total: number;
  };
  Temperature: number;
  KeySource: string;
  CacheHit: boolean;
  RequestHash: string;
  DurationMs: number;
  PromptTokens: number;
  OutputTokens: number;
  TotalTokens: number;
  IsVision: boolean;
}

export interface RequestsResponse {
  requests: RequestLog[];
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface TimeSeriesData {
  timestamp: string;
  count: number;
}
