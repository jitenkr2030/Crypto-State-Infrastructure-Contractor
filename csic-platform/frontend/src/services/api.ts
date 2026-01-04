// CSIC Platform - Frontend API Client
// Centralized API communication layer for the regulator dashboard

import axios, { AxiosInstance, AxiosError, InternalAxiosRequestConfig } from 'axios';
import { useAuthStore } from '../store';

// API Configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';
const API_TIMEOUT = 30000;

// Custom error class for API errors
export class APIError extends Error {
  constructor(
    message: string,
    public statusCode: number,
    public code: string,
    public details?: Record<string, unknown>
  ) {
    super(message);
    this.name = 'APIError';
  }
}

// Response types
export interface ApiResponse<T> {
  data: T;
  meta?: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
  timestamp: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

// Create axios instance with default configuration
const createApiClient = (): AxiosInstance => {
  const client = axios.create({
    baseURL: API_BASE_URL,
    timeout: API_TIMEOUT,
    headers: {
      'Content-Type': 'application/json',
      'X-Request-ID': generateRequestId(),
    },
  });

  // Request interceptor - Add auth token
  client.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
      const token = useAuthStore.getState().token;
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      config.headers['X-Request-ID'] = generateRequestId();
      return config;
    },
    (error) => Promise.reject(error)
  );

  // Response interceptor - Handle errors and responses
  client.interceptors.response.use(
    (response) => response,
    (error: AxiosError) => {
      if (error.response) {
        const { status, data } = error.response;
        
        // Handle specific error codes
        if (status === 401) {
          useAuthStore.getState().logout();
          window.location.href = '/login';
        }
        
        const errorData = data as { message?: string; code?: string; details?: Record<string, unknown> };
        throw new APIError(
          errorData?.message || 'An error occurred',
          status,
          errorData?.code || 'UNKNOWN_ERROR',
          errorData?.details
        );
      } else if (error.request) {
        throw new APIError('Network error - please check your connection', 0, 'NETWORK_ERROR');
      } else {
        throw new APIError(error.message || 'An unexpected error occurred', 0, 'UNKNOWN_ERROR');
      }
    }
  );

  return client;
};

// Generate unique request ID
const generateRequestId = (): string => {
  return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
};

// API Client instance
export const apiClient = createApiClient();

// Dashboard API
export const DashboardAPI = {
  async getSystemStatus() {
    const response = await apiClient.get('/dashboard/system-status');
    return response.data;
  },

  async getMetrics() {
    const response = await apiClient.get('/dashboard/metrics');
    return response.data;
  },

  async getCharts(type: string, period: string) {
    const response = await apiClient.get(`/dashboard/charts/${type}`, {
      params: { period },
    });
    return response.data;
  },
};

// Alert API
export const AlertAPI = {
  async getAlerts(params?: {
    page?: number;
    pageSize?: number;
    severity?: string;
    status?: string;
    category?: string;
    startDate?: string;
    endDate?: string;
  }) {
    const response = await apiClient.get('/alerts', { params });
    return response.data as PaginatedResponse<{
      id: string;
      title: string;
      description: string;
      severity: string;
      status: string;
      category: string;
      source: string;
      createdAt: string;
      updatedAt: string;
    }>;
  },

  async getAlertById(id: string) {
    const response = await apiClient.get(`/alerts/${id}`);
    return response.data;
  },

  async acknowledgeAlert(id: string) {
    const response = await apiClient.post(`/alerts/${id}/acknowledge`);
    return response.data;
  },

  async resolveAlert(id: string, resolution: string) {
    const response = await apiClient.post(`/alerts/${id}/resolve`, { resolution });
    return response.data;
  },

  async dismissAlert(id: string, reason: string) {
    const response = await apiClient.post(`/alerts/${id}/dismiss`, { reason });
    return response.data;
  },

  async createAlert(data: {
    title: string;
    description: string;
    severity: string;
    category: string;
    source: string;
    metadata?: Record<string, unknown>;
  }) {
    const response = await apiClient.post('/alerts', data);
    return response.data;
  },

  async getAlertStats() {
    const response = await apiClient.get('/alerts/stats');
    return response.data;
  },
};

// Exchange API
export const ExchangeAPI = {
  async getExchanges(params?: {
    page?: number;
    pageSize?: number;
    status?: string;
    jurisdiction?: string;
    search?: string;
  }) {
    const response = await apiClient.get('/exchanges', { params });
    return response.data as PaginatedResponse<{
      id: string;
      name: string;
      licenseNumber: string;
      status: string;
      jurisdiction: string;
      complianceScore: number;
      riskLevel: string;
      registrationDate: string;
      lastAudit: string;
    }>;
  },

  async getExchangeById(id: string) {
    const response = await apiClient.get(`/exchanges/${id}`);
    return response.data;
  },

  async createExchange(data: {
    name: string;
    website: string;
    contactEmail: string;
    jurisdiction: string;
    businessType: string;
  }) {
    const response = await apiClient.post('/exchanges', data);
    return response.data;
  },

  async updateExchange(id: string, data: Partial<{
    name: string;
    website: string;
    contactEmail: string;
    jurisdiction: string;
    status: string;
  }>) {
    const response = await apiClient.patch(`/exchanges/${id}`, data);
    return response.data;
  },

  async suspendExchange(id: string, reason: string) {
    const response = await apiClient.post(`/exchanges/${id}/suspend`, { reason });
    return response.data;
  },

  async revokeLicense(id: string, reason: string) {
    const response = await apiClient.post(`/exchanges/${id}/revoke`, { reason });
    return response.data;
  },

  async getExchangeComplianceHistory(id: string) {
    const response = await apiClient.get(`/exchanges/${id}/compliance-history`);
    return response.data;
  },

  async getExchangeTransactions(id: string, params?: {
    page?: number;
    pageSize?: number;
    startDate?: string;
    endDate?: string;
  }) {
    const response = await apiClient.get(`/exchanges/${id}/transactions`, { params });
    return response.data;
  },

  async getExchangeAuditReport(id: string) {
    const response = await apiClient.get(`/exchanges/${id}/audit-report`);
    return response.data;
  },
};

// Wallet API
export const WalletAPI = {
  async getWallets(params?: {
    page?: number;
    pageSize?: number;
    status?: string;
    type?: string;
    search?: string;
  }) {
    const response = await apiClient.get('/wallets', { params });
    return response.data as PaginatedResponse<{
      id: string;
      address: string;
      label: string;
      type: string;
      status: string;
      riskScore: number;
      firstSeen: string;
      lastActivity: string;
    }>;
  },

  async getWalletById(id: string) {
    const response = await apiClient.get(`/wallets/${id}`);
    return response.data;
  },

  async searchWallets(query: string) {
    const response = await apiClient.get('/wallets/search', { params: { q: query } });
    return response.data;
  },

  async freezeWallet(id: string, reason: string) {
    const response = await apiClient.post(`/wallets/${id}/freeze`, { reason });
    return response.data;
  },

  async unfreezeWallet(id: string, reason: string) {
    const response = await apiClient.post(`/wallets/${id}/unfreeze`, { reason });
    return response.data;
  },

  async blacklistWallet(id: string, reason: string) {
    const response = await apiClient.post(`/wallets/${id}/blacklist`, { reason });
    return response.data;
  },

  async removeFromBlacklist(id: string, reason: string) {
    const response = await apiClient.post(`/wallets/${id}/remove-blacklist`, { reason });
    return response.data;
  },

  async getWalletTransactions(id: string, params?: {
    page?: number;
    pageSize?: number;
  }) {
    const response = await apiClient.get(`/wallets/${id}/transactions`, { params });
    return response.data;
  },

  async getWalletRiskAnalysis(id: string) {
    const response = await apiClient.get(`/wallets/${id}/risk-analysis`);
    return response.data;
  },

  async getWalletNetworkGraph(id: string) {
    const response = await apiClient.get(`/wallets/${id}/network-graph`);
    return response.data;
  },
};

// Miner API
export const MinerAPI = {
  async getMiners(params?: {
    page?: number;
    pageSize?: number;
    status?: string;
    jurisdiction?: string;
    complianceStatus?: string;
  }) {
    const response = await apiClient.get('/miners', { params });
    return response.data as PaginatedResponse<{
      id: string;
      name: string;
      licenseNumber: string;
      status: string;
      jurisdiction: string;
      hashRate: number;
      energyConsumption: number;
      complianceStatus: string;
      registrationDate: string;
      lastInspection: string;
    }>;
  },

  async getMinerById(id: string) {
    const response = await apiClient.get(`/miners/${id}`);
    return response.data;
  },

  async createMiner(data: {
    name: string;
    jurisdiction: string;
    energySource: string;
    hashRate: number;
    energyConsumption: number;
  }) {
    const response = await apiClient.post('/miners', data);
    return response.data;
  },

  async updateMiner(id: string, data: Partial<{
    name: string;
    status: string;
    hashRate: number;
    energyConsumption: number;
    energySource: string;
  }>) {
    const response = await apiClient.patch(`/miners/${id}`, data);
    return response.data;
  },

  async suspendMiner(id: string, reason: string) {
    const response = await apiClient.post(`/miners/${id}/suspend`, { reason });
    return response.data;
  },

  async resumeMiner(id: string) {
    const response = await apiClient.post(`/miners/${id}/resume`);
    return response.data;
  },

  async getMinerEnergyReport(id: string, period: string) {
    const response = await apiClient.get(`/miners/${id}/energy-report`, { params: { period } });
    return response.data;
  },

  async getMinerComplianceReport(id: string) {
    const response = await apiClient.get(`/miners/${id}/compliance-report`);
    return response.data;
  },

  async scheduleInspection(id: string, date: string, notes: string) {
    const response = await apiClient.post(`/miners/${id}/schedule-inspection`, { date, notes });
    return response.data;
  },
};

// Transaction API
export const TransactionAPI = {
  async getTransactions(params?: {
    page?: number;
    pageSize?: number;
    status?: string;
    type?: string;
    currency?: string;
    exchangeId?: string;
    walletId?: string;
    startDate?: string;
    endDate?: string;
    minAmount?: number;
    maxAmount?: number;
  }) {
    const response = await apiClient.get('/transactions', { params });
    return response.data as PaginatedResponse<{
      id: string;
      txHash: string;
      type: string;
      amount: number;
      currency: string;
      fromAddress: string;
      toAddress: string;
      status: string;
      riskScore: number;
      timestamp: string;
    }>;
  },

  async getTransactionById(id: string) {
    const response = await apiClient.get(`/transactions/${id}`);
    return response.data;
  },

  async getTransactionByHash(hash: string) {
    const response = await apiClient.get(`/transactions/hash/${hash}`);
    return response.data;
  },

  async flagTransaction(id: string, reason: string) {
    const response = await apiClient.post(`/transactions/${id}/flag`, { reason });
    return response.data;
  },

  async unflagTransaction(id: string, reason: string) {
    const response = await apiClient.post(`/transactions/${id}/unflag`, { reason });
    return response.data;
  },

  async blockTransaction(id: string, reason: string) {
    const response = await apiClient.post(`/transactions/${id}/block`, { reason });
    return response.data;
  },

  async unblockTransaction(id: string, reason: string) {
    const response = await apiClient.post(`/transactions/${id}/unblock`, { reason });
    return response.data;
  },

  async getTransactionStats(params?: {
    startDate?: string;
    endDate?: string;
    currency?: string;
  }) {
    const response = await apiClient.get('/transactions/stats', { params });
    return response.data;
  },

  async getTransactionVolume(period: string) {
    const response = await apiClient.get('/transactions/volume', { params: { period } });
    return response.data;
  },
};

// Blockchain API
export const BlockchainAPI = {
  // Bitcoin
  async getBitcoinBlockHeight() {
    const response = await apiClient.get('/blockchain/btc/block-height');
    return response.data;
  },

  async getBitcoinBlock(hash: string) {
    const response = await apiClient.get(`/blockchain/btc/blocks/${hash}`);
    return response.data;
  },

  async getBitcoinTransaction(txHash: string) {
    const response = await apiClient.get(`/blockchain/btc/transactions/${txHash}`);
    return response.data;
  },

  async getBitcoinMempool() {
    const response = await apiClient.get('/blockchain/btc/mempool');
    return response.data;
  },

  async getBitcoinNetworkStats() {
    const response = await apiClient.get('/blockchain/btc/network-stats');
    return response.data;
  },

  // Ethereum
  async getEthereumBlockHeight() {
    const response = await apiClient.get('/blockchain/eth/block-height');
    return response.data;
  },

  async getEthereumBlock(hash: string | number) {
    const response = await apiClient.get(`/blockchain/eth/blocks/${hash}`);
    return response.data;
  },

  async getEthereumTransaction(txHash: string) {
    const response = await apiClient.get(`/blockchain/eth/transactions/${txHash}`);
    return response.data;
  },

  async getEthereumPendingTransactions() {
    const response = await apiClient.get('/blockchain/eth/pending');
    return response.data;
  },

  async getEthereumNetworkStats() {
    const response = await apiClient.get('/blockchain/eth/network-stats');
    return response.data;
  },

  // General
  async getBlockchainStatus() {
    const response = await apiClient.get('/blockchain/status');
    return response.data;
  },
};

// Compliance API
export const ComplianceAPI = {
  async getComplianceReports(params?: {
    page?: number;
    pageSize?: number;
    entityType?: string;
    status?: string;
  }) {
    const response = await apiClient.get('/compliance/reports', { params });
    return response.data;
  },

  async getComplianceReportById(id: string) {
    const response = await apiClient.get(`/compliance/reports/${id}`);
    return response.data;
  },

  async generateComplianceReport(entityType: string, entityId: string, period: string) {
    const response = await apiClient.post('/compliance/reports/generate', {
      entityType,
      entityId,
      period,
    });
    return response.data;
  },

  async getObligations(entityType: string, entityId: string) {
    const response = await apiClient.get(`/compliance/obligations`, {
      params: { entityType, entityId },
    });
    return response.data;
  },

  async checkObligationCompliance(obligationId: string) {
    const response = await apiClient.get(`/compliance/obligations/${obligationId}/check`);
    return response.data;
  },

  async getViolations(params?: {
    page?: number;
    pageSize?: number;
    severity?: string;
    status?: string;
  }) {
    const response = await apiClient.get('/compliance/violations', { params });
    return response.data;
  },

  async recordViolation(data: {
    entityType: string;
    entityId: string;
    violationType: string;
    description: string;
    severity: string;
    penalty?: string;
  }) {
    const response = await apiClient.post('/compliance/violations', data);
    return response.data;
  },

  async resolveViolation(id: string, resolution: string) {
    const response = await apiClient.post(`/compliance/violations/${id}/resolve`, { resolution });
    return response.data;
  },
};

// Reporting API
export const ReportingAPI = {
  async getReports(params?: {
    page?: number;
    pageSize?: number;
    type?: string;
    status?: string;
  }) {
    const response = await apiClient.get('/reports', { params });
    return response.data;
  },

  async getReportById(id: string) {
    const response = await apiClient.get(`/reports/${id}`);
    return response.data;
  },

  async generateReport(data: {
    type: string;
    title: string;
    parameters: Record<string, unknown>;
  }) {
    const response = await apiClient.post('/reports/generate', data);
    return response.data;
  },

  async exportReport(id: string, format: 'pdf' | 'xlsx' | 'csv') {
    const response = await apiClient.get(`/reports/${id}/export`, {
      params: { format },
      responseType: 'blob',
    });
    return response.data;
  },

  async scheduleReport(data: {
    type: string;
    title: string;
    schedule: string;
    recipients: string[];
    parameters: Record<string, unknown>;
  }) {
    const response = await apiClient.post('/reports/schedule', data);
    return response.data;
  },

  async getReportTemplates() {
    const response = await apiClient.get('/reports/templates');
    return response.data;
  },
};

// Audit API
export const AuditAPI = {
  async getAuditLogs(params?: {
    page?: number;
    pageSize?: number;
    userId?: string;
    action?: string;
    resourceType?: string;
    startDate?: string;
    endDate?: string;
  }) {
    const response = await apiClient.get('/audit/logs', { params });
    return response.data;
  },

  async getAuditLogById(id: string) {
    const response = await apiClient.get(`/audit/logs/${id}`);
    return response.data;
  },

  async getAuditTrail(resourceType: string, resourceId: string) {
    const response = await apiClient.get('/audit/trail', {
      params: { resourceType, resourceId },
    });
    return response.data;
  },

  async verifyAuditLog(logId: string) {
    const response = await apiClient.get(`/audit/logs/${logId}/verify`);
    return response.data;
  },

  async exportAuditLogs(params: {
    startDate: string;
    endDate: string;
    format: 'pdf' | 'xlsx' | 'csv';
  }) {
    const response = await apiClient.get('/audit/export', {
      params,
      responseType: 'blob',
    });
    return response.data;
  },
};

// User & Security API
export const UserAPI = {
  async getCurrentUser() {
    const response = await apiClient.get('/users/me');
    return response.data;
  },

  async updateProfile(data: {
    name?: string;
    email?: string;
    phone?: string;
  }) {
    const response = await apiClient.patch('/users/me/profile', data);
    return response.data;
  },

  async changePassword(currentPassword: string, newPassword: string) {
    const response = await apiClient.post('/users/me/change-password', {
      currentPassword,
      newPassword,
    });
    return response.data;
  },

  async enableMFA() {
    const response = await apiClient.post('/users/me/mfa/enable');
    return response.data;
  },

  async disableMFA(code: string) {
    const response = await apiClient.post('/users/me/mfa/disable', { code });
    return response.data;
  },

  async getUsers(params?: {
    page?: number;
    pageSize?: number;
    role?: string;
    status?: string;
  }) {
    const response = await apiClient.get('/users', { params });
    return response.data;
  },

  async createUser(data: {
    username: string;
    email: string;
    role: string;
    permissions?: string[];
  }) {
    const response = await apiClient.post('/users', data);
    return response.data;
  },

  async updateUser(id: string, data: Partial<{
    username: string;
    email: string;
    role: string;
    permissions: string[];
    status: string;
  }>) {
    const response = await apiClient.patch(`/users/${id}`, data);
    return response.data;
  },

  async deactivateUser(id: string) {
    const response = await apiClient.post(`/users/${id}/deactivate`);
    return response.data;
  },

  async activateUser(id: string) {
    const response = await apiClient.post(`/users/${id}/activate`);
    return response.data;
  },
};

// Health API
export const HealthAPI = {
  async getSystemHealth() {
    const response = await apiClient.get('/health');
    return response.data;
  },

  async getComponentHealth(component: string) {
    const response = await apiClient.get(`/health/${component}`);
    return response.data;
  },

  async getSystemMetrics() {
    const response = await apiClient.get('/health/metrics');
    return response.data;
  },
};

// Export all API modules
export const api = {
  dashboard: DashboardAPI,
  alerts: AlertAPI,
  exchanges: ExchangeAPI,
  wallets: WalletAPI,
  miners: MinerAPI,
  transactions: TransactionAPI,
  blockchain: BlockchainAPI,
  compliance: ComplianceAPI,
  reports: ReportingAPI,
  audit: AuditAPI,
  users: UserAPI,
  health: HealthAPI,
};

export default api;
