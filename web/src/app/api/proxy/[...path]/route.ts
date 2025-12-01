import { NextRequest, NextResponse } from 'next/server';

const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080/api';
const API_KEY = process.env.API_KEY ?? '';

export const dynamic = 'force-dynamic';

type ProxyParams = { path?: string[] };

async function proxyRequest(req: NextRequest, params: ProxyParams) {
  const targetPath = params.path?.join('/') || '';
  const targetUrl = new URL(`${API_BASE_URL.replace(/\/$/, '')}/${targetPath}`);
  req.nextUrl.searchParams.forEach((value, key) => {
    targetUrl.searchParams.append(key, value);
  });

  const headers = new Headers(req.headers);
  headers.set('X-API-Key', API_KEY);
  headers.delete('host');
  headers.delete('content-length');

  const init: RequestInit = {
    method: req.method,
    headers,
  };

  if (!['GET', 'HEAD'].includes(req.method)) {
    init.body = await req.arrayBuffer();
  }

  try {
    const upstreamResponse = await fetch(targetUrl, init);
    const responseHeaders = new Headers(upstreamResponse.headers);
    responseHeaders.delete('content-encoding');

    if (upstreamResponse.status === 204 || upstreamResponse.status === 304) {
      return new NextResponse(null, {
        status: upstreamResponse.status,
        headers: responseHeaders,
      });
    }

    const body = await upstreamResponse.arrayBuffer();

    return new NextResponse(body, {
      status: upstreamResponse.status,
      headers: responseHeaders,
    });
  } catch (error) {
    return NextResponse.json(
      { message: 'Upstream request failed', details: (error as Error).message },
      { status: 502 }
    );
  }
}

async function resolveParams(params: Promise<ProxyParams> | ProxyParams) {
  return Promise.resolve(params);
}

export async function GET(
  req: NextRequest,
  context: { params: Promise<ProxyParams> | ProxyParams }
) {
  const params = await resolveParams(context.params);
  return proxyRequest(req, params);
}

export async function POST(
  req: NextRequest,
  context: { params: Promise<ProxyParams> | ProxyParams }
) {
  const params = await resolveParams(context.params);
  return proxyRequest(req, params);
}

export async function PATCH(
  req: NextRequest,
  context: { params: Promise<ProxyParams> | ProxyParams }
) {
  const params = await resolveParams(context.params);
  return proxyRequest(req, params);
}

export async function DELETE(
  req: NextRequest,
  context: { params: Promise<ProxyParams> | ProxyParams }
) {
  const params = await resolveParams(context.params);
  return proxyRequest(req, params);
}
