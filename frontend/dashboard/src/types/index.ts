// User and Authentication Types
export interface User {
  id: string;
  email: string;
  name: string;
  role: 'admin' | 'operator' | 'viewer';
  avatar?: string;
  createdAt: string;
  lastLogin?: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

// Dashboard Types
export interface DashboardStats {
  totalNodes: number;
  activeNodes: number;
  totalTransactions: number;
  complianceRate: number;
  openViolations: number;
  systemHealth: number;
}

export interface ChartData {
  timestamp: string;
  value: number;
  label?: string;
}

// Node Types
export interface Node {
  id: string;
  name: string;
  type: string;
  status: 'running' | 'stopped' | 'syncing' | 'error' | 'maintenance';
  network: string;
  version: string;
  height: number;
  peers: number;
  lastBlockTime: string;
  uptime: number;
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  createdAt: string;
  updatedAt: string;
}

export interface NodeMetrics {
  nodeId: string;
  cpuUsage: number[];
  memoryUsage: number[];
  networkIn: number[];
  networkOut: number[];
  blockHeight: number[];
  timestamps: string[];
}

// Compliance Types
export interface ComplianceRule {
  id: string;
  name: string;
  description: string;
  type: string;
  severity: 'INFO' | 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  enabled: boolean;
  version: number;
  expression?: string;
  parameters: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
  expiresAt?: string;
}

export interface ComplianceResult {
  transactionId: string;
  overallStatus: 'PASS' | 'FAIL' | 'WARN' | 'PENDING' | 'ERROR';
  riskScore: number;
  checks: ComplianceCheck[];
  violations: Violation[];
  summary: ComplianceSummary;
  checkedAt: string;
  processingTime: number;
}

export interface ComplianceCheck {
  id: string;
  ruleId: string;
  ruleName: string;
  ruleType: string;
  status: 'PASS' | 'FAIL' | 'WARN' | 'PENDING' | 'ERROR';
  severity: string;
  message: string;
  details?: Record<string, unknown>;
  checkedAt: string;
  duration: number;
}

export interface Violation {
  id: string;
  transactionId: string;
  ruleId: string;
  ruleName: string;
  severity: 'INFO' | 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  status: 'OPEN' | 'RESOLVED' | 'IGNORED';
  resolution?: string;
  resolvedBy?: string;
  resolvedAt?: string;
  createdAt: string;
  details?: Record<string, unknown>;
}

export interface ComplianceSummary {
  totalChecks: number;
  passedChecks: number;
  failedChecks: number;
  warningChecks: number;
  criticalCount: number;
  highCount: number;
  mediumCount: number;
  lowCount: number;
}

// Audit Log Types
export interface AuditLog {
  id: string;
  timestamp: string;
  eventType: string;
  service: string;
  action: string;
  actor: string;
  actorType: string;
  resource: string;
  resourceType: string;
  details: Record<string, unknown>;
  outcome: 'success' | 'failure' | 'partial';
  severity: string;
  ipAddress?: string;
  userAgent?: string;
}

export interface AuditFilter {
  startDate?: string;
  endDate?: string;
  eventType?: string;
  service?: string;
  actor?: string;
  severity?: string;
  outcome?: string;
}

// Health Types
export interface HealthStatus {
  component: string;
  status: 'healthy' | 'degraded' | 'unhealthy' | 'unknown';
  lastCheck: string;
  uptime: number;
  version?: string;
  details?: Record<string, unknown>;
  checks: HealthCheck[];
}

export interface HealthCheck {
  name: string;
  status: 'healthy' | 'unhealthy' | 'degraded';
  message?: string;
  duration?: number;
  metadata?: Record<string, unknown>;
}

export interface HealthMetrics {
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  networkIn: number;
  networkOut: number;
  activeConnections: number;
  requestRate: number;
  errorRate: number;
  latencyP50: number;
  latencyP95: number;
  latencyP99: number;
}

// Report Types
export interface Report {
  id: string;
  name: string;
  type: string;
  status: 'pending' | 'generating' | 'completed' | 'failed';
  format: 'pdf' | 'csv' | 'json' | 'xlsx';
  createdAt: string;
  completedAt?: string;
  size?: number;
  downloadUrl?: string;
  parameters: Record<string, unknown>;
}

export interface ReportTemplate {
  id: string;
  name: string;
  description: string;
  type: string;
  parameters: ReportParameter[];
  createdAt: string;
}

export interface ReportParameter {
  name: string;
  type: 'string' | 'number' | 'date' | 'select' | 'multiselect';
  label: string;
  required: boolean;
  defaultValue?: unknown;
  options?: { value: string; label: string }[];
}

// API Response Types
export interface ApiResponse<T> {
  data: T;
  success: boolean;
  message?: string;
  pagination?: Pagination;
}

export interface Pagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface ApiError {
  message: string;
  code?: string;
  details?: Record<string, unknown>;
}

// Notification Types
export interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  timestamp: string;
  read: boolean;
  actionUrl?: string;
}

// Settings Types
export interface Settings {
  general: GeneralSettings;
  security: SecuritySettings;
  notifications: NotificationSettings;
  integrations: IntegrationSettings;
}

export interface GeneralSettings {
  theme: 'light' | 'dark' | 'system';
  language: string;
  timezone: string;
  dateFormat: string;
  timeFormat: string;
}

export interface SecuritySettings {
  mfaEnabled: boolean;
  sessionTimeout: number;
  passwordExpiry: number;
  ipWhitelist: string[];
  auditRetention: number;
}

export interface NotificationSettings {
  emailEnabled: boolean;
  slackEnabled: boolean;
  alertThreshold: string;
  digestFrequency: 'realtime' | 'hourly' | 'daily' | 'weekly';
}

export interface IntegrationSettings {
  auditLogEndpoint: string;
  controlLayerEndpoint: string;
  healthMonitorEndpoint: string;
  blockchainIndexerEndpoint: string;
  apiKeys: Record<string, string>;
}
