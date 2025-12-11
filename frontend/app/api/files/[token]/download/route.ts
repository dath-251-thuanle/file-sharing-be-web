import { NextRequest, NextResponse } from "next/server";

const getBaseUrl = (): string => {
  const internal = process.env.BACKEND_INTERNAL_URL?.trim();
  if (internal) return internal.endsWith("/") ? internal.slice(0, -1) : internal;
  return "http://127.0.0.1:8080/api";
};

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ token: string }> }
) {
  try {
    const { token } = await params;
    const searchParams = request.nextUrl.searchParams;
    const passwordFromQuery = searchParams.get('password');
    
    const passwordFromHeader = request.headers.get('X-File-Password');
    const password = passwordFromHeader || passwordFromQuery;
    
    const authHeader = request.headers.get('Authorization');
    const tokenFromHeader = authHeader?.startsWith('Bearer ') 
      ? authHeader.substring(7) 
      : null;
    let tokenFromCookie: string | null = null;
    try {
      const cookieValue = request.cookies.get('fs_access_token');
      tokenFromCookie = cookieValue?.value || null;
    } catch (error) {
      // Fallback: try parsing from Cookie header manually
      const cookieHeader = request.headers.get('Cookie');
      if (cookieHeader) {
        const cookies = cookieHeader.split(';').reduce((acc, cookie) => {
          const [key, value] = cookie.trim().split('=');
          if (key && value) acc[key] = decodeURIComponent(value);
          return acc;
        }, {} as Record<string, string>);
        tokenFromCookie = cookies['fs_access_token'] || null;
      }
    }
    const accessToken = tokenFromHeader || tokenFromCookie;
    
    // Log all request headers for debugging
    console.log(`[Download Route] All request headers:`, Object.fromEntries(request.headers.entries()));
    console.log(`[Download Route] Authorization header from request:`, authHeader ? authHeader.substring(0, 30) + '...' : 'none');
    console.log(`[Download Route] Token from cookie:`, tokenFromCookie ? tokenFromCookie.substring(0, 20) + '...' : 'none');
    console.log(`[Download Route] Final access token:`, accessToken ? accessToken.substring(0, 20) + '...' : 'none');
    
    const userAgent = request.headers.get('User-Agent') || '';
    const acceptHeader = request.headers.get('Accept') || '';
    const referer = request.headers.get('Referer') || '';
    const secFetchMode = request.headers.get('sec-fetch-mode') || '';
    const isDirectBrowserNavigation = 
      secFetchMode === 'navigate' &&
      acceptHeader.includes('text/html');
    const isBrowserRequest = isDirectBrowserNavigation;
    
    console.log(`[Download Route] User-Agent:`, userAgent.substring(0, 50));
    console.log(`[Download Route] Accept:`, acceptHeader);
    console.log(`[Download Route] Referer:`, referer);
    console.log(`[Download Route] isDirectBrowserNavigation:`, isDirectBrowserNavigation);
    console.log(`[Download Route] Has accessToken:`, !!accessToken);
    if (isDirectBrowserNavigation && !accessToken) {
      const frontendUrl = `/f/${token}?autoDownload=true`;
      console.log(`[Download Route] Direct browser navigation without token, redirecting to: ${frontendUrl}`);
      
      // Get the correct origin from request headers (set by nginx proxy)
      const host = request.headers.get('host') || request.headers.get('x-forwarded-host') || 'localhost';
      const protocol = request.headers.get('x-forwarded-proto') || 'http';
      const redirectUrl = `${protocol}://${host}${frontendUrl}`;
      
      console.log(`[Download Route] Redirect URL: ${redirectUrl}`);
      return NextResponse.redirect(redirectUrl);
    }

    const baseUrl = getBaseUrl();
    const downloadUrl = `${baseUrl}/files/${token}/download`;

    console.log(`[Download] Attempting download for token: ${token}`);
    console.log(`[Download] Backend URL: ${downloadUrl}`);
    console.log(`[Download] Has auth token: ${!!accessToken}`);
    console.log(`[Download] Auth header: ${authHeader ? authHeader.substring(0, 20) + '...' : 'none'}`);
    console.log(`[Download] Token from header: ${tokenFromHeader ? tokenFromHeader.substring(0, 20) + '...' : 'none'}`);
    console.log(`[Download] Token from cookie: ${tokenFromCookie ? tokenFromCookie.substring(0, 20) + '...' : 'none'}`);
    console.log(`[Download] Final token used: ${accessToken ? accessToken.substring(0, 20) + '...' : 'none'}`);
    console.log(`[Download] Has password: ${!!password}`);

    let downloadResponse;
    
    // Try direct connection first (faster, no proxy overhead)
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 15000); // 15 second timeout for file download
    
    try {
      console.log(`[Download] Trying direct connection: ${downloadUrl}`);
      downloadResponse = await fetch(downloadUrl, {
        method: 'GET',
        headers: {
          ...(accessToken ? { 'Authorization': `Bearer ${accessToken}` } : {}),
          ...(password ? { 'X-File-Password': password } : {}),
        },
        signal: controller.signal,
      });
      clearTimeout(timeoutId);
      console.log(`[Download] Direct connection success, status: ${downloadResponse.status}`);
    } catch (fetchError: any) {
      clearTimeout(timeoutId);
      console.error('[Download] Direct connection failed, trying nginx proxy');
      console.error('[Download] Error:', fetchError?.message);
      
      // Fallback: try via nginx
      const nginxServiceName = process.env.NODE_ENV !== 'production' ? 'nginx-dev' : 'nginx';
      const nginxUrl = `http://${nginxServiceName}/api/files/${token}/download`;
      const nginxController = new AbortController();
      const nginxTimeoutId = setTimeout(() => nginxController.abort(), 10000);
      
      try {
        console.log(`[Download] Trying via nginx: ${nginxUrl}`);
        downloadResponse = await fetch(nginxUrl, {
          method: 'GET',
          headers: {
            ...(accessToken ? { 'Authorization': `Bearer ${accessToken}` } : {}),
            ...(password ? { 'X-File-Password': password } : {}),
          },
          signal: nginxController.signal,
        });
        clearTimeout(nginxTimeoutId);
        console.log(`[Download] Nginx proxy success, status: ${downloadResponse.status}`);
      } catch (nginxError: any) {
        clearTimeout(nginxTimeoutId);
        console.error('[Download] Both methods failed');
        return NextResponse.json(
          {
            error: 'Connection error',
            message: `Không thể kết nối đến backend server. Đã thử direct connection và nginx proxy.`,
          },
          { status: 500 }
        );
      }
    }

    // Handle different error responses
    if (!downloadResponse.ok) {
      const errorData = await downloadResponse.json().catch(() => ({}));
      const errorMessage = errorData.message || errorData.error || 'Không thể tải file';
      
      // Handle 401 - File requires authentication (whitelist)
      if (downloadResponse.status === 401) {
        if (errorMessage.includes('requires authentication') || 
            errorMessage.includes('Bearer token')) {
          // Check if request has token (from header or cookie)
          const hasTokenInRequest = !!accessToken;
          
          if (hasTokenInRequest) {
            // Has token but still 401 - token invalid or expired
            // Return JSON error, let frontend handle redirect
            return NextResponse.json(
              {
                error: 'Token invalid',
                message: 'Token không hợp lệ hoặc đã hết hạn. Vui lòng đăng nhập lại.',
                requiresLogin: true,
                tokenInvalid: true,
              },
              { status: 401 }
            );
          } else {
            // No token in request - API/fetch request from frontend
            // Return JSON error, frontend will handle login redirect
            return NextResponse.json(
              {
                error: 'Authentication required',
                message: 'File này yêu cầu đăng nhập. Vui lòng đăng nhập để truy cập.',
                requiresLogin: true,
                noTokenInRequest: true,
              },
              { status: 401 }
            );
          }
        }
      }
      
      // Handle 403 - Password required, incorrect password, or not whitelisted
      if (downloadResponse.status === 403) {
        // Check if it's a whitelist error
        if (errorMessage.includes('not in the shared list') || 
            errorMessage.includes('not allowed to download') ||
            errorMessage.includes('Access denied') && errorMessage.includes('shared list')) {
          // User is logged in but not in whitelist
          if (isBrowserRequest) {
            // Redirect to beautiful error page
            const errorUrl = `/error/forbidden?message=${encodeURIComponent(errorMessage)}&reason=not_whitelisted`;
            const host = request.headers.get('host') || request.headers.get('x-forwarded-host') || 'localhost';
            const protocol = request.headers.get('x-forwarded-proto') || 'http';
            const redirectUrl = `${protocol}://${host}${errorUrl}`;
            return NextResponse.redirect(redirectUrl);
          }
          
          // For API requests, return JSON
          return NextResponse.json(
            {
              error: 'Access denied',
              message: errorMessage,
              notWhitelisted: true,
            },
            { status: 403 }
          );
        }
        
        // Check if it's a password error
        if (errorMessage.includes('Password required') || 
            errorMessage.includes('password-protected') ||
            errorMessage.includes('password') ||
            errorMessage.includes('Incorrect password')) {
          // If browser request and no password provided, redirect to beautiful password page
          if (isBrowserRequest && !password) {
            const passwordUrl = `/error/password-required?token=${token}`;
            const host = request.headers.get('host') || request.headers.get('x-forwarded-host') || 'localhost';
            const protocol = request.headers.get('x-forwarded-proto') || 'http';
            const redirectUrl = `${protocol}://${host}${passwordUrl}`;
            return NextResponse.redirect(redirectUrl);
          }
          
          // For API requests or if password was wrong, return JSON
          return NextResponse.json(
            {
              error: errorData.error || 'Password required',
              message: errorMessage,
              requiresPassword: !password,
              token: token,
            },
            { status: 403 }
          );
        }
      }
      
      // Other errors (404, 410, 423, etc.)
      return NextResponse.json(
        {
          error: errorData.error || 'Download failed',
          message: errorMessage,
        },
        { status: downloadResponse.status }
      );
    }

    // Stream the file response directly without loading into memory
    const contentType = downloadResponse.headers.get('Content-Type') || 'application/octet-stream';
    const contentDisposition = downloadResponse.headers.get('Content-Disposition') || '';
    const contentLength = downloadResponse.headers.get('Content-Length');

    console.log('[Download Route] Content-Type:', contentType);
    console.log('[Download Route] Content-Disposition:', contentDisposition);
    console.log('[Download Route] Content-Length:', contentLength);
    console.log('[Download Route] Has body:', !!downloadResponse.body);

    // Create response headers
    const responseHeaders = new Headers();
    responseHeaders.set('Content-Type', contentType);
    if (contentDisposition) {
      responseHeaders.set('Content-Disposition', contentDisposition);
    }
    if (contentLength) {
      responseHeaders.set('Content-Length', contentLength);
    }

    // Stream the response body directly - this avoids loading entire file into memory
    if (!downloadResponse.body) {
      console.error('[Download Route] No response body!');
      return NextResponse.json(
        {
          error: 'No content',
          message: 'File response has no body',
        },
        { status: 500 }
      );
    }

    // For file downloads, convert to blob first (more reliable than streaming)
    console.log('[Download Route] Converting response to blob...');
    const blob = await downloadResponse.blob();
    console.log('[Download Route] Blob created, size:', blob.size);

    console.log('[Download Route] Returning blob response');
    return new NextResponse(blob, {
      status: 200,
      headers: responseHeaders,
    });
  } catch (error) {
    console.error('Download error:', error);
    return NextResponse.json(
      {
        error: 'Internal server error',
        message: 'Đã xảy ra lỗi khi xử lý yêu cầu tải file',
      },
      { status: 500 }
    );
  }
}

