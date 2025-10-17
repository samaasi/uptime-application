/**
 * API Configuration
 * Centralized configuration for API endpoints
 */

// Base API URL - connects directly to api-services (no api-gateway)
export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8082';

// API Endpoints
export const API_ENDPOINTS = {
  // Health check
  health: `${API_BASE_URL}/health`,
  
  // Authentication
  auth: {
    login: `${API_BASE_URL}/auth/login`,
    logout: `${API_BASE_URL}/auth/logout`,
    register: `${API_BASE_URL}/auth/register`,
    refresh: `${API_BASE_URL}/auth/refresh`,
  },
  
  // User management
  users: {
    profile: `${API_BASE_URL}/users/profile`,
    update: `${API_BASE_URL}/users/profile`,
  },
  
  // Uptime monitoring
  monitors: {
    list: `${API_BASE_URL}/monitors`,
    create: `${API_BASE_URL}/monitors`,
    get: (id: string) => `${API_BASE_URL}/monitors/${id}`,
    update: (id: string) => `${API_BASE_URL}/monitors/${id}`,
    delete: (id: string) => `${API_BASE_URL}/monitors/${id}`,
    status: (id: string) => `${API_BASE_URL}/monitors/${id}/status`,
  },
  
  // Incidents
  incidents: {
    list: `${API_BASE_URL}/incidents`,
    get: (id: string) => `${API_BASE_URL}/incidents/${id}`,
  },
  
  // Analytics
  analytics: {
    overview: `${API_BASE_URL}/analytics/overview`,
    uptime: (monitorId: string) => `${API_BASE_URL}/analytics/uptime/${monitorId}`,
    response_times: (monitorId: string) => `${API_BASE_URL}/analytics/response-times/${monitorId}`,
  },
} as const;

// API Configuration
export const API_CONFIG = {
  timeout: 10000, // 10 seconds
  retries: 3,
  headers: {
    'Content-Type': 'application/json',
  },
} as const;

// Helper function to build API URLs
export function buildApiUrl(endpoint: string, params?: Record<string, string>): string {
  let url = endpoint;
  
  if (params) {
    const searchParams = new URLSearchParams(params);
    url += `?${searchParams.toString()}`;
  }
  
  return url;
}