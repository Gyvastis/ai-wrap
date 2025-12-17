import { NextRequest } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL || "http://api:8089";

async function proxyRequest(req: NextRequest, path: string) {
  const url = new URL(req.url);
  const targetUrl = `${BACKEND_URL}/admin/${path}${url.search}`;

  const res = await fetch(targetUrl, {
    method: req.method,
    headers: {
      "Content-Type": "application/json",
    },
    body: req.method !== "GET" ? await req.text() : undefined,
    cache: "no-store",
  });

  const data = await res.text();

  return new Response(data, {
    status: res.status,
    headers: {
      "Content-Type": res.headers.get("Content-Type") || "application/json",
      "Cache-Control": "no-store, no-cache, must-revalidate",
      "Pragma": "no-cache",
    },
  });
}

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { path } = await params;
  return proxyRequest(req, path.join("/"));
}

export async function POST(
  req: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { path } = await params;
  return proxyRequest(req, path.join("/"));
}
