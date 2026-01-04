import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import { useAuthStore } from '../store';

// API Configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor - add auth token
api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    // Add request ID
    config.headers['X-Request-ID'] = generateRequestId();

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor - handle errors
api.interceptors.response.use(
  (response) => {
    return response.data;
  },
  (error) => {
    // Handle specific error codes
    if (error.response) {
      const { status, data } = error.response;

      switch (status) {
        case 401:
          // Unauthorized - redirect to login
          useAuthStore.getState().logout();
          window.location.href = '/login';
          break;
        case 403:
          console.error('Access forbidden:', data.message);
          break;
        case 404:
          console.error('Resource not found:', data.message);
          break;
        case 429:
          console.error('Rate limit exceeded');
          break;
        case 500:
          console.error('Server error:', data.message);
          break;
      }

      return Promise.reject({
        status,
        message: data.message || data.error || 'An error occurred',
        code: data.code,
        details: data.details,
      });
    }

    if (error.request) {
      return Promise.reject({
        status: 0,
        message: 'Network error - please check your connection',
        code: 'NETWORK_ERROR',
      });
    }

    return Promise.reject({
      status: 0,
      message: error.message,
      code: 'UNKNOWN_ERROR',
    });
  }
);

// Generate unique request ID
function generateRequestId(): string {
  return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

// API Methods
export const apiClient = {
  // GET request
  get: <T = any>(url: string, config?: AxiosRequestConfig) =>
    api.get<T>(url, config),

  // POST request
  post: <T = any>(url: string, data?: any, config?: AxiosRequestConfig) =>
    api.post<T>(url, data, config),

  // PUT request
  put: <T = any>(url: string, data?: any, config?: AxiosRequestConfig) =>
    api.put<T>(url, data, config),

  // PATCH request
  patch: <T = any>(url: string, data?: any, config?: AxiosRequestConfig) =>
    api.patch<T>(url, data, config),

  // DELETE request
  delete: <T = any>(url: string, config?: AxiosRequestConfig) =>
    api.delete<T>(url, config),
};

// API Service - System
export const systemApi = {
  getStatus: () => apiClient.get('/status'),
  getHealth: () => apiClient.get('/health'),
  getReady: () => apiClient.get('/ready'),
  emergencyStop: (data: { stop_type: string; entity_id?: string; reason: string }) =>
    apiClient.post('/emergency/stop', data),
  emergencyResume: (data: { stop_id: string; resolution: string }) =>
    apiClient.post('/emergency/resume', data),
};

// API Service - Exchanges
export const exchangeApi = {
  getAll: (params?: { status?: string; limit?: number; offset?: number }) =>
    apiClient.get('/exchanges', { params }),
  getById: (id: string) => apiClient.get(`/exchanges/${id}`),
  freeze: (id: string, reason: string) =>
    apiClient.post(`/exchanges/${id}/freeze`, { reason }),
  thaw: (id: string) => apiClient.post(`/exchanges/${id}/thaw`),
  getMetrics: (id: string) => apiClient.get(`/exchanges/${id}/metrics`),
};

// API Service - Wallets
export const walletApi = {
  getAll: (params?: { limit?: number; offset?: number }) =>
    apiClient.get('/wallets', { params }),
  create: (data: any) => apiClient.post('/wallets', data),
  freeze: (id: string, reason: string) =>
    apiClient.post(`/wallets/${id}/freeze`, { reason }),
  unfreeze: (id: string) => apiClient.post(`/wallets/${id}/unfreeze`),
  transfer: (id: string, data: { to_address: string; amount: number; currency: string }) =>
    apiClient.post(`/wallets/${id}/transfer`, data),
};

// API Service - Transactions
export const transactionApi = {
  getAll: (params?: any) => apiClient.get('/transactions', { params }),
  getById: (id: string) => apiClient.get(`/transactions/${id}`),
  search: (params: any) => apiClient.get('/transactions/search', { params }),
  flag: (txId: string, reason: string) =>
    apiClient.post('/transactions/flag', { tx_id: txId, reason }),
  getRiskScore: (address: string) => apiClient.get(`/risk/score/${address}`),
};

// API Service - Licenses
export const licenseApi = {
  getAll: (params?: { status?: string; entity_type?: string; limit?: number; offset?: number }) =>
    apiClient.get('/licenses', { params }),
  create: (data: any) => apiClient.post('/licenses', data),
  getById: (id: string) => apiClient.get(`/licenses/${id}`),
  update: (id: string, data: any) => apiClient.put(`/licenses/${id}`, data),
  revoke: (id: string) => apiClient.post(`/licenses/${id}/revoke`),
};

// API Service - Mining
export const miningApi = {
  getAll: (params?: { status?: string; limit?: number; offset?: number }) =>
    apiClient.get('/miners', { params }),
  getById: (id: string) => apiClient.get(`/miners/${id}`),
  shutdown: (id: string, reason: string) =>
    apiClient.post(`/miners/${id}/shutdown`, { reason }),
  start: (id: string) => apiClient.post(`/miners/${id}/start`),
  getMetrics: () => apiClient.get('/mining/metrics'),
};

// API Service - Energy
export const energyApi = {
  getGridStatus: () => apiClient.get('/energy/grid'),
  getConsumption: (params?: { start_date?: string; end_date?: string }) =>
    apiClient.get('/energy/consumption', { params }),
  triggerLoadShedding: (data: { region_id?: string; percent: number }) =>
    apiClient.post('/energy/load-shedding', data),
};

// API Service - Reports
export const reportApi = {
  getAll: (params?: { status?: string; type?: string; limit?: number; offset?: number }) =>
    apiClient.get('/reports', { params }),
  generate: (data: { report_type: string; period_start?: string; period_end?: string }) =>
    apiClient.post('/reports/generate', data),
  download: (id: string) => apiClient.get(`/reports/${id}/download`, { responseType: 'blob' }),
};

// API Service - Audit & Security
export const auditApi = {
  getLogs: (params?: { user_id?: string; resource_type?: string; limit?: number; offset?: number }) =>
    apiClient.get('/audit/logs', { params }),
  getLogById: (id: string) => apiClient.get(`/audit/logs/${id}`),
  export: (params?: any) => apiClient.post('/audit/export', params),
  getHSMStatus: () => apiClient.get('/security/hsm/status'),
  rotateKeys: () => apiClient.post('/security/keys/rotate'),
};

// API Service - Alerts
export const alertApi = {
  getActive: (params?: { severity?: string; limit?: number; offset?: number }) =>
    apiClient.get('/alerts', { params }),
  acknowledge: (id: string, assignedTo: string) =>
    apiClient.post(`/alerts/${id}/acknowledge`, { assigned_to: assignedTo }),
  resolve: (id: string, resolution: string) =>
    apiClient.post(`/alerts/${id}/resolve`, { resolution }),
};

export default api;
