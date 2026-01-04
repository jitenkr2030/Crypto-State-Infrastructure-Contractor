import React, { Suspense, lazy } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from './store/authStore';
import Layout from './components/Layout';
import LoadingSpinner from './components/LoadingSpinner';

// Lazy load all dashboard modules for better performance
const Dashboard = lazy(() => import('./modules/Dashboard'));
const ExchangeOversight = lazy(() => import('./modules/ExchangeOversight'));
const WalletGovernance = lazy(() => import('./modules/WalletGovernance'));
const TransactionMonitoring = lazy(() => import('./modules/TransactionMonitoring'));
const LicensingCompliance = lazy(() => import('./modules/LicensingCompliance'));
const MiningControl = lazy(() => import('./modules/MiningControl'));
const EnergyIntegration = lazy(() => import('./modules/EnergyIntegration'));
const ReportingIntelligence = lazy(() => import('./modules/ReportingIntelligence'));
const SecurityAudit = lazy(() => import('./modules/SecurityAudit'));
const Settings = lazy(() => import('./modules/Settings'));
const Login = lazy(() => import('./components/Login'));

// Protected Route wrapper
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuthStore();

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

// Public Route wrapper (redirect to dashboard if authenticated)
const PublicRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuthStore();

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
};

const App: React.FC = () => {
  return (
    <Suspense fallback={<LoadingSpinner fullScreen />}>
      <Routes>
        {/* Public Routes */}
        <Route
          path="/login"
          element={
            <PublicRoute>
              <Login />
            </PublicRoute>
          }
        />

        {/* Protected Routes */}
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Dashboard />} />
          <Route path="exchanges" element={<ExchangeOversight />} />
          <Route path="wallets" element={<WalletGovernance />} />
          <Route path="transactions" element={<TransactionMonitoring />} />
          <Route path="licenses" element={<LicensingCompliance />} />
          <Route path="mining" element={<MiningControl />} />
          <Route path="energy" element={<EnergyIntegration />} />
          <Route path="reports" element={<ReportingIntelligence />} />
          <Route path="security" element={<SecurityAudit />} />
          <Route path="settings" element={<Settings />} />
        </Route>

        {/* Catch all - redirect to dashboard */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Suspense>
  );
};

export default App;
