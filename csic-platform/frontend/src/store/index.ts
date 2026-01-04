// CSIC Platform - Frontend State Management Store
// Global state management using Zustand for the regulator dashboard

import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { v4 as uuidv4 } from 'uuid';

// Types
export interface User {
  id: string;
  username: string;
  role: 'ADMIN' | 'REGULATOR' | 'AUDITOR' | 'ANALYST';
  permissions: string[];
  lastLogin: Date;
  mfaEnabled: boolean;
}

export interface Alert {
  id: string;
  title: string;
  description: string;
  severity: 'CRITICAL' | 'WARNING' | 'INFO';
  status: 'ACTIVE' | 'ACKNOWLEDGED' | 'RESOLVED' | 'DISMISSED';
  category: string;
  source: string;
  createdAt: Date;
  updatedAt: Date;
  acknowledgedBy?: string;
  resolvedBy?: string;
  metadata?: Record<string, unknown>;
}

export interface SystemStatus {
  status: 'ONLINE' | 'DEGRADED' | 'OFFLINE' | 'EMERGENCY';
  uptime: number;
  lastHeartbeat: Date;
  components: {
    name: string;
    status: 'healthy' | 'degraded' | 'down';
    latency?: number;
  }[];
}

export interface Exchange {
  id: string;
  name: string;
  licenseNumber: string;
  status: 'ACTIVE' | 'SUSPENDED' | 'REVOKED' | 'PENDING';
  registrationDate: Date;
  lastAudit: Date;
  jurisdiction: string;
  website: string;
  contactEmail: string;
  complianceScore: number;
  riskLevel: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
}

export interface Transaction {
  id: string;
  txHash: string;
  type: 'TRANSFER' | 'EXCHANGE' | 'WALLET' | 'CONTRACT';
  amount: number;
  currency: string;
  fromAddress: string;
  toAddress: string;
  timestamp: Date;
  status: 'PENDING' | 'CONFIRMED' | 'FLAGGED' | 'BLOCKED';
  riskScore: number;
  exchangeId?: string;
}

export interface Wallet {
  id: string;
  address: string;
  label: string;
  type: 'CUSTODIAL' | 'NON_CUSTODIAL' | 'EXCHANGE' | 'MIXER' | 'DARKNET';
  status: 'ACTIVE' | 'FROZEN' | 'BLACKLISTED';
  riskScore: number;
  associatedEntities: string[];
  firstSeen: Date;
  lastActivity: Date;
  balance?: number;
  currency?: string;
}

export interface Miner {
  id: string;
  name: string;
  licenseNumber: string;
  status: 'ACTIVE' | 'SUSPENDED' | 'OFFLINE';
  jurisdiction: string;
  hashRate: number;
  energyConsumption: number;
  energySource: string;
  registrationDate: Date;
  lastInspection: Date;
  complianceStatus: 'COMPLIANT' | 'NON_COMPLIANT' | 'UNDER_REVIEW';
}

// System Store - Core system state
interface SystemState {
  status: SystemStatus;
  isInitialized: boolean;
  isLoading: boolean;
  error: string | null;
  setStatus: (status: SystemStatus) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  initialize: () => Promise<void>;
  updateHeartbeat: () => void;
}

export const useSystemStore = create<SystemState>()(
  devtools(
    (set, get) => ({
      status: {
        status: 'OFFLINE',
        uptime: 0,
        lastHeartbeat: new Date(),
        components: [],
      },
      isInitialized: false,
      isLoading: false,
      error: null,

      setStatus: (status) => set({ status }),
      setLoading: (isLoading) => set({ isLoading }),
      setError: (error) => set({ error }),

      initialize: async () => {
        set({ isLoading: true });
        try {
          // Simulate initialization
          await new Promise((resolve) => setTimeout(resolve, 1000));
          
          set({
            status: {
              status: 'ONLINE',
              uptime: 86400 * 15,
              lastHeartbeat: new Date(),
              components: [
                { name: 'API Gateway', status: 'healthy', latency: 12 },
                { name: 'Database', status: 'healthy', latency: 5 },
                { name: 'Redis Cache', status: 'healthy', latency: 1 },
                { name: 'Bitcoin Node', status: 'healthy', latency: 45 },
                { name: 'Ethereum Node', status: 'healthy', latency: 120 },
                { name: 'Exchange Surveillance', status: 'healthy', latency: 25 },
                { name: 'Risk Engine', status: 'healthy', latency: 8 },
                { name: 'Alert System', status: 'healthy', latency: 3 },
              ],
            },
            isInitialized: true,
            isLoading: false,
            error: null,
          });
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : 'Initialization failed',
            isLoading: false,
          });
        }
      },

      updateHeartbeat: () => {
        const current = get().status;
        set({
          status: {
            ...current,
            lastHeartbeat: new Date(),
          },
        });
      },
    }),
    { name: 'system-store' }
  )
);

// Alert Store - Alert management
interface AlertState {
  alerts: Alert[];
  filters: {
    severity: string[];
    status: string[];
    category: string[];
    dateRange: { start: Date | null; end: Date | null };
  };
  selectedAlertId: string | null;
  unreadCount: number;
  
  addAlert: (alert: Omit<Alert, 'id' | 'createdAt' | 'updatedAt'>) => void;
  acknowledgeAlert: (id: string, userId: string) => void;
  resolveAlert: (id: string, userId: string) => void;
  dismissAlert: (id: string) => void;
  setSelectedAlert: (id: string | null) => void;
  setFilters: (filters: Partial<AlertState['filters']>) => void;
  loadAlerts: () => Promise<void>;
  markAllAsRead: () => void;
}

export const useAlertStore = create<AlertState>()(
  devtools(
    persist(
      (set, get) => ({
        alerts: [
          {
            id: uuidv4(),
            title: '高风险交易模式检测',
            description: '检测到与已知混币服务相关的异常交易模式，涉及地址 1A2B3C4D5E',
            severity: 'CRITICAL',
            status: 'ACTIVE',
            category: '交易监控',
            source: '风险引擎',
            createdAt: new Date(Date.now() - 3600000),
            updatedAt: new Date(Date.now() - 3600000),
          },
          {
            id: uuidv4(),
            title: '交易所合规分数下降',
            description: 'CryptoExchange Pro 的合规分数从 92 下降到 78，需要关注',
            severity: 'WARNING',
            status: 'ACTIVE',
            category: '合规监控',
            source: '合规系统',
            createdAt: new Date(Date.now() - 7200000),
            updatedAt: new Date(Date.now() - 7200000),
          },
          {
            id: uuidv4(),
            title: '矿池能源消耗异常',
            description: 'Northern Mining Pool 能源消耗超出预期阈值 15%',
            severity: 'INFO',
            status: 'ACKNOWLEDGED',
            category: '能源监控',
            source: '能源监测',
            createdAt: new Date(Date.now() - 86400000),
            updatedAt: new Date(Date.now() - 43200000),
            acknowledgedBy: 'admin',
          },
        ],
        filters: {
          severity: [],
          status: [],
          category: [],
          dateRange: { start: null, end: null },
        },
        selectedAlertId: null,
        unreadCount: 2,

        addAlert: (alertData) => {
          const alert: Alert = {
            ...alertData,
            id: uuidv4(),
            createdAt: new Date(),
            updatedAt: new Date(),
          };
          set((state) => ({
            alerts: [alert, ...state.alerts],
            unreadCount: state.unreadCount + 1,
          }));
        },

        acknowledgeAlert: (id, userId) => {
          set((state) => ({
            alerts: state.alerts.map((alert) =>
              alert.id === id
                ? { ...alert, status: 'ACKNOWLEDGED', acknowledgedBy: userId, updatedAt: new Date() }
                : alert
            ),
          }));
        },

        resolveAlert: (id, userId) => {
          set((state) => ({
            alerts: state.alerts.map((alert) =>
              alert.id === id
                ? { ...alert, status: 'RESOLVED', resolvedBy: userId, updatedAt: new Date() }
                : alert
            ),
          }));
        },

        dismissAlert: (id) => {
          set((state) => ({
            alerts: state.alerts.map((alert) =>
              alert.id === id ? { ...alert, status: 'DISMISSED', updatedAt: new Date() } : alert
            ),
          }));
        },

        setSelectedAlert: (id) => set({ selectedAlertId: id }),
        setFilters: (filters) =>
          set((state) => ({ filters: { ...state.filters, ...filters } })),

        loadAlerts: async () => {
          // Simulate API call
          await new Promise((resolve) => setTimeout(resolve, 500));
          // In production, this would fetch from the API
        },

        markAllAsRead: () => set({ unreadCount: 0 }),
      }),
      { name: 'alert-store' }
    ),
    { name: 'alert-store' }
  )
);

// Exchange Store - Exchange management
interface ExchangeState {
  exchanges: Exchange[];
  selectedExchangeId: string | null;
  isLoading: boolean;
  error: string | null;
  
  loadExchanges: () => Promise<void>;
  selectExchange: (id: string | null) => void;
  suspendExchange: (id: string) => void;
  revokeLicense: (id: string) => void;
  updateComplianceScore: (id: string, score: number) => void;
}

export const useExchangeStore = create<ExchangeState>()(
  devtools(
    (set, get) => ({
      exchanges: [
        {
          id: '1',
          name: 'CryptoExchange Pro',
          licenseNumber: 'CSIC-2024-001',
          status: 'ACTIVE',
          registrationDate: new Date('2024-01-15'),
          lastAudit: new Date('2024-11-01'),
          jurisdiction: 'Singapore',
          website: 'https://cryptopro.exchange',
          contactEmail: 'compliance@cryptopro.exchange',
          complianceScore: 92,
          riskLevel: 'LOW',
        },
        {
          id: '2',
          name: 'Digital Asset Hub',
          licenseNumber: 'CSIC-2024-002',
          status: 'ACTIVE',
          registrationDate: new Date('2024-02-20'),
          lastAudit: new Date('2024-10-15'),
          jurisdiction: 'Switzerland',
          website: 'https://dah.io',
          contactEmail: 'regulatory@dah.io',
          complianceScore: 88,
          riskLevel: 'LOW',
        },
        {
          id: '3',
          name: 'BlockTrade Global',
          licenseNumber: 'CSIC-2024-003',
          status: 'SUSPENDED',
          registrationDate: new Date('2024-03-10'),
          lastAudit: new Date('2024-09-20'),
          jurisdiction: 'British Virgin Islands',
          website: 'https://blocktrade.com',
          contactEmail: 'compliance@blocktrade.com',
          complianceScore: 45,
          riskLevel: 'HIGH',
        },
      ],
      selectedExchangeId: null,
      isLoading: false,
      error: null,

      loadExchanges: async () => {
        set({ isLoading: true });
        try {
          await new Promise((resolve) => setTimeout(resolve, 500));
          // In production, fetch from API
          set({ isLoading: false });
        } catch (error) {
          set({
            error: error instanceof Error ? error.message : 'Failed to load exchanges',
            isLoading: false,
          });
        }
      },

      selectExchange: (id) => set({ selectedExchangeId: id }),

      suspendExchange: (id) => {
        set((state) => ({
          exchanges: state.exchanges.map((exchange) =>
            exchange.id === id ? { ...exchange, status: 'SUSPENDED' } : exchange
          ),
        }));
      },

      revokeLicense: (id) => {
        set((state) => ({
          exchanges: state.exchanges.map((exchange) =>
            exchange.id === id ? { ...exchange, status: 'REVOKED' } : exchange
          ),
        }));
      },

      updateComplianceScore: (id, score) => {
        set((state) => ({
          exchanges: state.exchanges.map((exchange) =>
            exchange.id === id
              ? {
                  ...exchange,
                  complianceScore: score,
                  riskLevel: score >= 80 ? 'LOW' : score >= 60 ? 'MEDIUM' : score >= 40 ? 'HIGH' : 'CRITICAL',
                }
              : exchange
          ),
        }));
      },
    }),
    { name: 'exchange-store' }
  )
);

// Wallet Store - Wallet monitoring
interface WalletState {
  wallets: Wallet[];
  isLoading: boolean;
  error: string | null;
  freezeWallet: (id: string) => void;
  unfreezeWallet: (id: string) => void;
  blacklistWallet: (id: string) => void;
  loadWallets: () => Promise<void>;
}

export const useWalletStore = create<WalletState>()(
  devtools(
    (set) => ({
      wallets: [
        {
          id: '1',
          address: '1A2B3C4D5E6F7890ABCDEF1234567890',
          label: '可疑地址 #4521',
          type: 'MIXER',
          status: 'BLACKLISTED',
          riskScore: 95,
          associatedEntities: ['DarkWeb Market A', 'Ransomware Group B'],
          firstSeen: new Date('2023-06-15'),
          lastActivity: new Date(Date.now() - 86400000),
        },
        {
          id: '2',
          address: '0x1234567890ABCDEF1234567890ABCDEF12345678',
          label: '交易所热钱包 B',
          type: 'EXCHANGE',
          status: 'ACTIVE',
          riskScore: 15,
          associatedEntities: ['Digital Asset Hub'],
          firstSeen: new Date('2024-01-20'),
          lastActivity: new Date(),
          balance: 15420.5,
          currency: 'ETH',
        },
      ],
      isLoading: false,
      error: null,

      freezeWallet: (id) => {
        set((state) => ({
          wallets: state.wallets.map((wallet) =>
            wallet.id === id ? { ...wallet, status: 'FROZEN' } : wallet
          ),
        }));
      },

      unfreezeWallet: (id) => {
        set((state) => ({
          wallets: state.wallets.map((wallet) =>
            wallet.id === id ? { ...wallet, status: 'ACTIVE' } : wallet
          ),
        }));
      },

      blacklistWallet: (id) => {
        set((state) => ({
          wallets: state.wallets.map((wallet) =>
            wallet.id === id ? { ...wallet, status: 'BLACKLISTED' } : wallet
          ),
        }));
      },

      loadWallets: async () => {
        set({ isLoading: true });
        await new Promise((resolve) => setTimeout(resolve, 500));
        set({ isLoading: false });
      },
    }),
    { name: 'wallet-store' }
  )
);

// Miner Store - Mining operations
interface MinerState {
  miners: Miner[];
  isLoading: boolean;
  error: string | null;
  loadMiners: () => Promise<void>;
  suspendMiner: (id: string) => void;
  updateEnergyConsumption: (id: string, consumption: number) => void;
}

export const useMinerStore = create<MinerState>()(
  devtools(
    (set) => ({
      miners: [
        {
          id: '1',
          name: 'Northern Mining Pool',
          licenseNumber: 'CSIC-MIN-2024-001',
          status: 'ACTIVE',
          jurisdiction: 'Canada',
          hashRate: 450,
          energyConsumption: 520,
          energySource: 'Hydroelectric',
          registrationDate: new Date('2023-08-10'),
          lastInspection: new Date('2024-10-01'),
          complianceStatus: 'COMPLIANT',
        },
        {
          id: '2',
          name: 'GreenHash Energy',
          licenseNumber: 'CSIC-MIN-2024-002',
          status: 'ACTIVE',
          jurisdiction: 'Iceland',
          hashRate: 320,
          energyConsumption: 380,
          energySource: 'Geothermal',
          registrationDate: new Date('2023-12-05'),
          lastInspection: new Date('2024-09-15'),
          complianceStatus: 'COMPLIANT',
        },
      ],
      isLoading: false,
      error: null,

      loadMiners: async () => {
        set({ isLoading: true });
        await new Promise((resolve) => setTimeout(resolve, 500));
        set({ isLoading: false });
      },

      suspendMiner: (id) => {
        set((state) => ({
          miners: state.miners.map((miner) =>
            miner.id === id ? { ...miner, status: 'SUSPENDED' } : miner
          ),
        }));
      },

      updateEnergyConsumption: (id, consumption) => {
        set((state) => ({
          miners: state.miners.map((miner) =>
            miner.id === id ? { ...miner, energyConsumption: consumption } : miner
          ),
        }));
      },
    }),
    { name: 'miner-store' }
  )
);

// Transaction Store - Transaction monitoring
interface TransactionState {
  transactions: Transaction[];
  isLoading: boolean;
  error: string | null;
  loadTransactions: () => Promise<void>;
  flagTransaction: (id: string) => void;
  blockTransaction: (id: string) => void;
}

export const useTransactionStore = create<TransactionState>()(
  devtools(
    (set) => ({
      transactions: [],
      isLoading: false,
      error: null,

      loadTransactions: async () => {
        set({ isLoading: true });
        await new Promise((resolve) => setTimeout(resolve, 500));
        set({ isLoading: false });
      },

      flagTransaction: (id) => {
        set((state) => ({
          transactions: state.transactions.map((tx) =>
            tx.id === id ? { ...tx, status: 'FLAGGED' } : tx
          ),
        }));
      },

      blockTransaction: (id) => {
        set((state) => ({
          transactions: state.transactions.map((tx) =>
            tx.id === id ? { ...tx, status: 'BLOCKED' } : tx
          ),
        }));
      },
    }),
    { name: 'transaction-store' }
  )
);

// Auth Store - Authentication state
interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (username: string, password: string) => Promise<boolean>;
  logout: () => void;
  refreshToken: () => Promise<void>;
  hasPermission: (permission: string) => boolean;
}

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      (set, get) => ({
        user: null,
        token: null,
        isAuthenticated: false,
        isLoading: false,

        login: async (username, password) => {
          set({ isLoading: true });
          try {
            // Simulate API call
            await new Promise((resolve) => setTimeout(resolve, 1000));
            
            // Mock successful login
            const user: User = {
              id: uuidv4(),
              username,
              role: 'ADMIN',
              permissions: [
                'view:dashboard',
                'view:alerts',
                'manage:alerts',
                'view:exchanges',
                'manage:exchanges',
                'view:wallets',
                'manage:wallets',
                'view:miners',
                'manage:miners',
                'view:transactions',
                'manage:transactions',
                'view:reports',
                'create:reports',
                'manage:settings',
                'manage:users',
              ],
              lastLogin: new Date(),
              mfaEnabled: true,
            };
            
            const token = 'mock-jwt-token-' + uuidv4();
            
            set({
              user,
              token,
              isAuthenticated: true,
              isLoading: false,
            });
            
            return true;
          } catch {
            set({ isLoading: false });
            return false;
          }
        },

        logout: () => {
          set({
            user: null,
            token: null,
            isAuthenticated: false,
          });
        },

        refreshToken: async () => {
          const { user } = get();
          if (!user) return;
          
          // Simulate token refresh
          await new Promise((resolve) => setTimeout(resolve, 500));
          set({ token: 'mock-refreshed-token-' + uuidv4() });
        },

        hasPermission: (permission) => {
          const { user } = get();
          return user?.permissions.includes(permission) ?? false;
        },
      }),
      { name: 'auth-store' }
    ),
    { name: 'auth-store' }
  )
);

// UI Store - UI state
interface UIState {
  sidebarCollapsed: boolean;
  theme: 'dark' | 'light';
  notifications: {
    id: string;
    type: 'success' | 'error' | 'warning' | 'info';
    message: string;
    timestamp: Date;
  }[];
  
  toggleSidebar: () => void;
  setTheme: (theme: 'dark' | 'light') => void;
  addNotification: (type: UIState['notifications'][0]['type'], message: string) => void;
  removeNotification: (id: string) => void;
  clearNotifications: () => void;
}

export const useUIStore = create<UIState>()(
  devtools(
    persist(
      (set) => ({
        sidebarCollapsed: false,
        theme: 'dark',
        notifications: [],

        toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),
        setTheme: (theme) => set({ theme }),

        addNotification: (type, message) => {
          const notification = {
            id: uuidv4(),
            type,
            message,
            timestamp: new Date(),
          };
          set((state) => ({
            notifications: [...state.notifications, notification],
          }));
          
          // Auto-remove after 5 seconds
          setTimeout(() => {
            set((state) => ({
              notifications: state.notifications.filter((n) => n.id !== notification.id),
            }));
          }, 5000);
        },

        removeNotification: (id) =>
          set((state) => ({
            notifications: state.notifications.filter((n) => n.id !== id),
          })),

        clearNotifications: () => set({ notifications: [] }),
      }),
      { name: 'ui-store' }
    ),
    { name: 'ui-store' }
  )
);
