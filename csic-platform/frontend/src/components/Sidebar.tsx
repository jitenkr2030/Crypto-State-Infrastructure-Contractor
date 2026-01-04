import React from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { useAuthStore } from '../../store';

interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
  alertCount: number;
}

interface NavItem {
  path: string;
  label: string;
  icon: React.ReactNode;
  badge?: number;
  roles?: string[];
}

const Sidebar: React.FC<SidebarProps> = ({ collapsed, onToggle, alertCount }) => {
  const location = useLocation();
  const { user } = useAuthStore();

  const navItems: NavItem[] = [
    {
      path: '/',
      label: '控制台',
      icon: <DashboardIcon />,
    },
    {
      path: '/exchanges',
      label: '交易所监管',
      icon: <ExchangeIcon />,
    },
    {
      path: '/wallets',
      label: '钱包治理',
      icon: <WalletIcon />,
    },
    {
      path: '/transactions',
      label: '交易监控',
      icon: <TransactionIcon />,
    },
    {
      path: '/licenses',
      label: '许可合规',
      icon: <LicenseIcon />,
    },
    {
      path: '/mining',
      label: '挖矿控制',
      icon: <MiningIcon />,
    },
    {
      path: '/energy',
      label: '能源集成',
      icon: <EnergyIcon />,
    },
    {
      path: '/reports',
      label: '报告情报',
      icon: <ReportIcon />,
    },
    {
      path: '/security',
      label: '安全审计',
      icon: <SecurityIcon />,
      badge: alertCount > 0 ? alertCount : undefined,
    },
    {
      path: '/settings',
      label: '系统设置',
      icon: <SettingsIcon />,
      roles: ['ADMIN'],
    },
  ];

  const isActive = (path: string) => {
    if (path === '/') {
      return location.pathname === '/';
    }
    return location.pathname.startsWith(path);
  };

  return (
    <aside className={`sidebar ${collapsed ? 'collapsed' : ''}`}>
      <div className="sidebar-header">
        <div className="logo">
          {!collapsed && (
            <>
              <div className="logo-icon">
                <ShieldIcon />
              </div>
              <div className="logo-text">
                <span className="logo-title">CSIC Platform</span>
                <span className="logo-subtitle">Crypto State Infrastructure</span>
              </div>
            </>
          )}
          {collapsed && <ShieldIcon />}
        </div>
        <button className="toggle-btn" onClick={onToggle}>
          <ChevronIcon direction={collapsed ? 'right' : 'left'} />
        </button>
      </div>

      <nav className="sidebar-nav">
        <ul className="nav-list">
          {navItems.map((item) => {
            // Check role permissions
            if (item.roles && user && !item.roles.includes(user.role)) {
              return null;
            }

            return (
              <li key={item.path} className="nav-item">
                <NavLink
                  to={item.path}
                  className={`nav-link ${isActive(item.path) ? 'active' : ''}`}
                >
                  <span className="nav-icon">{item.icon}</span>
                  {!collapsed && (
                    <>
                      <span className="nav-label">{item.label}</span>
                      {item.badge && (
                        <span className="nav-badge">{item.badge}</span>
                      )}
                    </>
                  )}
                </NavLink>
              </li>
            );
          })}
        </ul>
      </nav>

      <div className="sidebar-footer">
        {!collapsed && (
          <div className="system-info">
            <div className="info-item">
              <span className="info-label">版本</span>
              <span className="info-value">v1.0.0</span>
            </div>
            <div className="info-item">
              <span className="info-label">环境</span>
              <span className="info-value status-online">生产</span>
            </div>
          </div>
        )}
      </div>
    </aside>
  );
};

// Icon Components
const DashboardIcon: React.FC = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <rect x="3" y="3" width="7" height="9" rx="1" />
    <rect x="14" y="3" width="7" height="5" rx="1" />
    <rect x="14" y="12" width="7" height="9" rx="1" />
    <rect x="3" y="16" width="7" height="5" rx="1" />
  </svg>
);

const ExchangeIcon: React.FC = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M12 2L2 7l10 5 10-5-10-5z" />
    <path d="M2 17l10 5 10-5" />
    <path d="M2 12l10 5 10-5" />
  </svg>
);

const WalletIcon: React.FC = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M21 12V7a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h7" />
    <rect x="2" y="6" width="20" height="12" rx="2" />
    <path d="M2 10h20" />
  </svg>
);

const TransactionIcon: React.FC = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
  </svg>
);

const LicenseIcon: React.FC = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
    <polyline points="14,2 14,8 20,8" />
    <line x1="16" y1="13" x2="8" y2="13" />
    <line x1="16" y1="17" x2="8" y2="17" />
    <polyline points="10,9 9,9 8,9" />
  </svg>
);

const MiningIcon: React.FC = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M12 2L2 7l10 5 10-5-10-5z" />
    <path d="M2 17l10 5 10-5" />
    <path d="M2 12l10 5 10-5" />
    <circle cx="12" cy="12" r="3" />
  </svg>
);

const EnergyIcon: React.FC = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <polygon points="13,2 3,14 12,14 11,22 21,10 12,10 13,2" />
  </svg>
);

const ReportIcon: React.FC = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
    <polyline points="14,2 14,8 20,8" />
    <line x1="16" y1="13" x2="8" y2="13" />
    <line x1="16" y1="17" x2="8" y2="17" />
    <line x1="10" y1="9" x2="8" y2="9" />
  </svg>
);

const SecurityIcon: React.FC = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
    <path d="M9 12l2 2 4-4" />
  </svg>
);

const SettingsIcon: React.FC = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <circle cx="12" cy="12" r="3" />
    <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
  </svg>
);

const ShieldIcon: React.FC = () => (
  <svg width="32" height="32" viewBox="0 0 24 24" fill="currentColor">
    <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z" />
  </svg>
);

interface ChevronIconProps {
  direction: 'left' | 'right';
}

const ChevronIcon: React.FC<ChevronIconProps> = ({ direction }) => (
  <svg
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    style={{ transform: direction === 'right' ? 'rotate(180deg)' : 'none' }}
  >
    <polyline points="15,18 9,12 15,6" />
  </svg>
);

export default Sidebar;
