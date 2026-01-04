import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// Types
interface User {
  id: string;
  username: string;
  email: string;
  role: 'ADMIN' | 'OPERATOR' | 'AUDITOR' | 'VIEWER';
  department: string;
  mfaEnabled: boolean;
}

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  setUser: (user: User) => void;
  setLoading: (loading: boolean) => void;
}

// Auth Store
export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,

      login: async (username: string, password: string) => {
        set({ isLoading: true });

        try {
          // Simulate API call - in production, this would call the actual API
          await new Promise(resolve => setTimeout(resolve, 1000));

          // Mock user data - replace with actual API response
          const mockUser: User = {
            id: 'usr_001',
            username,
            email: `${username}@csic.gov`,
            role: username === 'admin' ? 'ADMIN' : 'OPERATOR',
            department: 'Financial Regulation',
            mfaEnabled: true,
          };

          const mockToken = 'eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.mock_token_for_demo';

          set({
            user: mockUser,
            token: mockToken,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      logout: () => {
        set({
          user: null,
          token: null,
          isAuthenticated: false,
        });
      },

      setUser: (user: User) => {
        set({ user });
      },

      setLoading: (isLoading: boolean) => {
        set({ isLoading });
      },
    }),
    {
      name: 'csic-auth',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

// System Store - for global system state
interface SystemState {
  status: 'ONLINE' | 'DEGRADED' | 'OFFLINE' | 'EMERGENCY';
  lastHeartbeat: Date;
  activeExchanges: number;
  monitoredWallets: number;
  pendingAlerts: number;
  hsmConnected: boolean;
  setSystemStatus: (status: Partial<SystemState>) => void;
}

export const useSystemStore = create<SystemState>((set) => ({
  status: 'ONLINE',
  lastHeartbeat: new Date(),
  activeExchanges: 12,
  monitoredWallets: 156,
  pendingAlerts: 3,
  hsmConnected: true,

  setSystemStatus: (status) => {
    set((prev) => ({ ...prev, ...status }));
  },
}));

// Alert Store
interface Alert {
  id: string;
  severity: 'INFO' | 'WARNING' | 'CRITICAL' | 'EMERGENCY';
  category: string;
  title: string;
  description: string;
  source: string;
  status: 'ACTIVE' | 'ACKNOWLEDGED' | 'RESOLVED';
  createdAt: Date;
}

interface AlertState {
  alerts: Alert[];
  unreadCount: number;
  addAlert: (alert: Alert) => void;
  acknowledgeAlert: (id: string) => void;
  resolveAlert: (id: string) => void;
  setAlerts: (alerts: Alert[]) => void;
}

export const useAlertStore = create<AlertState>((set) => ({
  alerts: [
    {
      id: 'alt_001',
      severity: 'WARNING',
      category: 'EXCHANGE',
      title: '交易所交易量异常',
      description: '检测到某交易所的交易量在1小时内增长了200%',
      source: 'Surveillance System',
      status: 'ACTIVE',
      createdAt: new Date(),
    },
    {
      id: 'alt_002',
      severity: 'CRITICAL',
      category: 'TRANSACTION',
      title: '大额可疑交易',
      description: '检测到一笔1000万USDT的大额转账，来自被标记地址',
      source: 'Risk Engine',
      status: 'ACTIVE',
      createdAt: new Date(),
    },
  ],
  unreadCount: 2,

  addAlert: (alert) => {
    set((state) => ({
      alerts: [alert, ...state.alerts],
      unreadCount: state.unreadCount + 1,
    }));
  },

  acknowledgeAlert: (id) => {
    set((state) => ({
      alerts: state.alerts.map((a) =>
        a.id === id ? { ...a, status: 'ACKNOWLEDGED' } : a
      ),
    }));
  },

  resolveAlert: (id) => {
    set((state) => ({
      alerts: state.alerts.map((a) =>
        a.id === id ? { ...a, status: 'RESOLVED' } : a
      ),
      unreadCount: Math.max(0, state.unreadCount - 1),
    }));
  },

  setAlerts: (alerts) => {
    set({ alerts, unreadCount: alerts.filter((a) => a.status === 'ACTIVE').length });
  },
}));

// Theme Store
interface ThemeState {
  theme: 'light' | 'dark';
  toggleTheme: () => void;
  setTheme: (theme: 'light' | 'dark') => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set) => ({
      theme: 'dark',

      toggleTheme: () => {
        set((state) => ({
          theme: state.theme === 'light' ? 'dark' : 'light',
        }));
      },

      setTheme: (theme) => {
        set({ theme });
      },
    }),
    {
      name: 'csic-theme',
    }
  )
);
