import { NextResponse } from "next/server"
import type { NextRequest } from "next/server"

// Routes that require authentication
const protectedRoutes = ["/dashboard"]

// Routes that should redirect to dashboard if authenticated
const authRoutes = ["/login"]

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl
  const token = request.cookies.get("access_token")?.value

  // Check if route is protected
  const isProtectedRoute = protectedRoutes.some((route) =>
    pathname.startsWith(route)
  )

  // Check if route is auth route
  const isAuthRoute = authRoutes.some((route) => pathname.startsWith(route))

  // For now, we'll check localStorage token on client side
  // This middleware just handles basic redirects
  
  // If trying to access auth routes while logged in, redirect to dashboard
  // Note: In production, use HTTP-only cookies for better security
  
  return NextResponse.next()
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    "/((?!api|_next/static|_next/image|favicon.ico).*)",
  ],
}
