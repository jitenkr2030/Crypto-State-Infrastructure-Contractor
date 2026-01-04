// CSIC Platform - React Application Entry Point
// Main application component with routing and layout

import React, { useEffect, useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore, useSystemStore } from './store';
import Sidebar from './components/Sidebar';
import Header from './components/Header';
import Login from './pages/Login';
import Dashboard from './pages/admin-dashboard/Dashboard';
import AlertConsole from './pages/incident-console/AlertConsole';
import ExchangeList from './pages/regulator-view/ExchangeList';
import ExchangeDetail from './pages/regulator-view/ExchangeDetail';
import WalletMonitor from './pages/regulator-view/WalletMonitor';
import TransactionMonitor from './pages/regulator-view/TransactionMonitor';
import MinerRegistry from './pages/regulator-view/MinerRegistry';
import ComplianceDashboard from './pages/auditor-view/ComplianceDashboard';
import Reports from './pages/auditor-view/Reports';
import AuditLogs from './pages/auditor-view/AuditLogs';
import Settings from './pages/admin-dashboard/Settings';
import Users from './pages/admin-dashboard/Users';
import './App.css';

// Protected Route wrapper component
interface ProtectedRouteProps {
  children: React.ReactNode;
  requiredPermission?: string;
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children, requiredPermission }) => {
  const { isAuthenticated, isLoading, hasPermission } = useAuthStore();
  const [isVerifying, setIsVerifying] = useState(true);

  useEffect(() => {
    // Verify authentication on mount
    const verifyAuth = async () => {
      await new Promise((resolve) => setTimeout(resolve, 100));
      setIsVerifying(false);
    };
    verifyAuth();
  }, []);

  if (isLoading || isVerifying) {
    return (
      <div className="app-loading">
        <div className="loading-container">
          <div className="csic-logo">
            <svg viewBox="0 0 100 100" className="logo-icon">
              <circle cx="50" cy="50" r="45" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M30 50 L45 65 L70 35" fill="none" stroke="currentColor" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
          </div>
          <div className="loading-text">正在初始化安全连接...</div>
          <div className="loading-progress">
            <div className="progress-bar"></div>
          </div>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (requiredPermission && !hasPermission(requiredPermission)) {
    return (
      <div className="access-denied">
        <div className="access-denied-content">
          <svg viewBox="0 0 24 24" className="warning-icon">
            <path d="M12 2L2 7v10l10 5 10-5V7L12 2z" fill="none" stroke="currentColor" strokeWidth="2" />
            <path d="M12 8v4M12 16h.01" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
          </svg>
          <h1>访问被拒绝</h1>
          <p>您没有执行此操作的权限。</p>
          <button onClick={() => window.history.back()}>返回</button>
        </div>
      </div>
    );
  }

  return <>{children}</>;
};

// Main Layout component
interface MainLayoutProps {
  children: React.ReactNode;
}

const MainLayout: React.FC<MainLayoutProps> = ({ children }) => {
  const { sidebarCollapsed, toggleSidebar } = useAuthStore.getState();
  const { status } = useSystemStore.getState();

  return (
    <div className="app-layout">
      <Sidebar collapsed={sidebarCollapsed} onToggle={toggleSidebar} />
      <div className={`main-content ${sidebarCollapsed ? 'sidebar-collapsed' : ''}`}>
        <Header systemStatus={status.status} />
        <main className="page-content">
          {children}
        </main>
      </div>
    </div>
  );
};

// Error Boundary component
interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

interface ErrorBoundaryProps {
  children: React.ReactNode;
}

class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Application Error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="app-error">
          <div className="error-content">
            <svg viewBox="0 0 24 24" className="error-icon">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M12 8v4M12 16h.01" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
            </svg>
            <h1>系统错误</h1>
            <p>发生了意外错误，请刷新页面或联系系统管理员。</p>
            <div className="error-details">
              {this.state.error?.message}
            </div>
            <button onClick={() => window.location.reload()}>刷新页面</button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

// Main App component
const App: React.FC = () => {
  const { initialize, isInitialized } = useSystemStore();
  const { isAuthenticated } = useAuthStore();
  const [isStartingUp, setIsStartingUp] = useState(true);

  useEffect(() => {
    const startup = async () => {
      try {
        await initialize();
      } catch (error) {
        console.error('Failed to initialize system:', error);
      } finally {
        setIsStartingUp(false);
      }
    };

    if (!isInitialized) {
      startup();
    } else {
      setIsStartingUp(false);
    }
  }, [isInitialized, initialize]);

  if (isStartingUp) {
    return (
      <div className="app-loading">
        <div className="loading-container">
          <div className="csic-logo">
            <svg viewBox="0 0 100 100" className="logo-icon">
              <circle cx="50" cy="50" r="45" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M30 50 L45 65 L70 35" fill="none" stroke="currentColor" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
          </div>
          <div className="loading-title">加密货币国家基础设施承包商</div>
          <div className="loading-subtitle">Crypto State Infrastructure Contractor</div>
          <div className="loading-progress">
            <div className="progress-bar"></div>
          </div>
          <div className="loading-status">正在连接安全模块...</div>
        </div>
      </div>
    );
  }

  return (
    <ErrorBoundary>
      <Router>
        <Routes>
          {/* Public routes */}
          <Route
            path="/login"
            element={
              isAuthenticated ? <Navigate to="/" replace /> : <Login />
            }
          />

          {/* Protected routes */}
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <MainLayout>
                  <Dashboard />
                </MainLayout>
              </ProtectedRoute>
            }
          />

          {/* Alert Console */}
          <Route
            path="/alerts"
            element={
              <ProtectedRoute requiredPermission="view:alerts">
                <MainLayout>
                  <AlertConsole />
                </MainLayout>
              </ProtectedRoute>
            }
          />

          {/* Exchange Management */}
          <Route
            path="/exchanges"
            element={
              <ProtectedRoute requiredPermission="view:exchanges">
                <MainLayout>
                  <ExchangeList />
                </MainLayout>
              </ProtectedRoute>
            }
          />
          <Route
            path="/exchanges/:id"
            element={
              <ProtectedRoute requiredPermission="view:exchanges">
                <MainLayout>
                  <ExchangeDetail />
                </MainLayout>
              </ProtectedRoute>
            }
          />

          {/* Wallet Monitor */}
          <Route
            path="/wallets"
            element={
              <ProtectedRoute requiredPermission="view:wallets">
                <MainLayout>
                  <WalletMonitor />
                </MainLayout>
              </ProtectedRoute>
            }
          />

          {/* Transaction Monitor */}
          <Route
            path="/transactions"
            element={
              <ProtectedRoute requiredPermission="view:transactions">
                <MainLayout>
                  <TransactionMonitor />
                </MainLayout>
              </ProtectedRoute>
            }
          />

          {/* Miner Registry */}
          <Route
            path="/miners"
            element={
              <ProtectedRoute requiredPermission="view:miners">
                <MainLayout>
                  <MinerRegistry />
                </MainLayout>
              </ProtectedRoute>
            }
          />

          {/* Compliance Dashboard */}
          <Route
            path="/compliance"
            element={
              <ProtectedRoute requiredPermission="view:reports">
                <MainLayout>
                  <ComplianceDashboard />
                </MainLayout>
              </ProtectedRoute>
            }
          />

          {/* Reports */}
          <Route
            path="/reports"
            element={
              <ProtectedRoute requiredPermission="view:reports">
                <MainLayout>
                  <Reports />
                </MainLayout>
              </ProtectedRoute>
            }
          />

          {/* Audit Logs */}
          <Route
            path="/audit"
            element={
              <ProtectedRoute requiredPermission="view:audit">
                <MainLayout>
                  <AuditLogs />
                </MainLayout>
              </ProtectedRoute>
            }
          />

          {/* Settings */}
          <Route
            path="/settings"
            element={
              <ProtectedRoute requiredPermission="manage:settings">
                <MainLayout>
                  <Settings />
                </MainLayout>
              </ProtectedRoute>
            }
          />

          {/* User Management */}
          <Route
            path="/users"
            element={
              <ProtectedRoute requiredPermission="manage:users">
                <MainLayout>
                  <Users />
                </MainLayout>
              </ProtectedRoute>
            }
          />

          {/* Fallback route */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Router>
    </ErrorBoundary>
  );
};

export default App;
