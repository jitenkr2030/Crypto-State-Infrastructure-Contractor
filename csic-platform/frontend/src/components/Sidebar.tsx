// Sidebar Component - Navigation sidebar for CSIC Platform
import React from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { useAuthStore } from '../store';
import styles from './Sidebar.module.css';

interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
}

const Sidebar: React.FC<SidebarProps> = ({ collapsed, onToggle }) => {
  const location = useLocation();
  const { user, hasPermission } = useAuthStore();

  const navItems = [
    {
      section: '主面板',
      items: [
        { path: '/', label: '仪表盘', icon: 'dashboard', permission: null },
      ],
    },
    {
      section: '监管模块',
      items: [
        { path: '/exchanges', label: '交易所管理', icon: 'exchange', permission: 'view:exchanges' },
        { path: '/wallets', label: '钱包监控', icon: 'wallet', permission: 'view:wallets' },
        { path: '/transactions', label: '交易监控', icon: 'transaction', permission: 'view:transactions' },
        { path: '/miners', label: '矿工注册', icon: 'miner', permission: 'view:miners' },
      ],
    },
    {
      section: '审计模块',
      items: [
        { path: '/compliance', label: '合规仪表盘', icon: 'compliance', permission: 'view:reports' },
        { path: '/reports', label: '报告生成', icon: 'report', permission: 'view:reports' },
        { path: '/audit', label: '审计日志', icon: 'audit', permission: 'view:audit' },
      ],
    },
    {
      section: '事件响应',
      items: [
        { path: '/alerts', label: '告警控制台', icon: 'alert', permission: 'view:alerts' },
      ],
    },
    {
      section: '系统管理',
      items: [
        { path: '/users', label: '用户管理', icon: 'users', permission: 'manage:users' },
        { path: '/settings', label: '系统设置', icon: 'settings', permission: 'manage:settings' },
      ],
    },
  ];

  const renderIcon = (icon: string) => {
    const icons: Record<string, JSX.Element> = {
      dashboard: (
        <svg viewBox="0 0 24 24" width="20" height="20">
          <rect x="3" y="3" width="7" height="7" rx="1" fill="currentColor" />
          <rect x="14" y="3" width="7" height="7" rx="1" fill="currentColor" />
          <rect x="3" y="14" width="7" height="7" rx="1" fill="currentColor" />
          <rect x="14" y="14" width="7" height="7" rx="1" fill="currentColor" />
        </svg>
      ),
      exchange: (
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path d="M12 2L2 7l10 5 10-5-10-5z" fill="none" stroke="currentColor" strokeWidth="2" />
          <path d="M2 17l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
          <path d="M2 12l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
        </svg>
      ),
      wallet: (
        <svg viewBox="0 0 24 24" width="20" height="20">
          <rect x="1" y="4" width="22" height="16" rx="2" fill="none" stroke="currentColor" strokeWidth="2" />
          <path d="M1 10h22" fill="none" stroke="currentColor" strokeWidth="2" />
        </svg>
      ),
      transaction: (
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" fill="none" stroke="currentColor" strokeWidth="2" />
        </svg>
      ),
      miner: (
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" fill="none" stroke="currentColor" strokeWidth="2" />
          <path d="M3.27 6.96L12 12.01l8.73-5.05" fill="none" stroke="currentColor" strokeWidth="2" />
          <path d="M12 22.08V12" fill="none" stroke="currentColor" strokeWidth="2" />
        </svg>
      ),
      compliance: (
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" fill="none" stroke="currentColor" strokeWidth="2" />
          <path d="M9 12l2 2 4-4" fill="none" stroke="currentColor" strokeWidth="2" />
        </svg>
      ),
      report: (
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill="none" stroke="currentColor" strokeWidth="2" />
          <path d="M14 2v6h6M16 13H8M16 17H8M10 9H8" fill="none" stroke="currentColor" strokeWidth="2" />
        </svg>
      ),
      audit: (
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill="none" stroke="currentColor" strokeWidth="2" />
          <path d="M14 2v6h6" fill="none" stroke="currentColor" strokeWidth="2" />
          <line x1="16" y1="13" x2="8" y2="13" stroke="currentColor" strokeWidth="2" />
          <line x1="16" y1="17" x2="8" y2="17" stroke="currentColor" strokeWidth="2" />
          <line x1="10" y1="9" x2="8" y2="9" stroke="currentColor" strokeWidth="2" />
        </svg>
      ),
      alert: (
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" fill="none" stroke="currentColor" strokeWidth="2" />
          <line x1="12" y1="9" x2="12" y2="13" stroke="currentColor" strokeWidth="2" />
          <line x1="12" y1="17" x2="12.01" y2="17" stroke="currentColor" strokeWidth="2" />
        </svg>
      ),
      users: (
        <svg viewBox="0 0 24 24" width="20" height="20">
          <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" fill="none" stroke="currentColor" strokeWidth="2" />
          <circle cx="9" cy="7" r="4" fill="none" stroke="currentColor" strokeWidth="2" />
          <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" fill="none" stroke="currentColor" strokeWidth="2" />
        </svg>
      ),
      settings: (
        <svg viewBox="0 0 24 24" width="20" height="20">
          <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" strokeWidth="2" />
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" fill="none" stroke="currentColor" strokeWidth="2" />
        </svg>
      ),
    };
    return icons[icon] || icons.dashboard;
  };

  const shouldShowItem = (permission: string | null): boolean => {
    if (!permission) return true;
    return hasPermission(permission);
  };

  return (
    <aside className={`${styles.sidebar} ${collapsed ? styles.collapsed : ''}`}>
      <div className={styles.header}>
        <div className={styles.logo}>
          <svg viewBox="0 0 100 100" className={styles.logoIcon}>
            <circle cx="50" cy="50" r="45" fill="none" stroke="currentColor" strokeWidth="4" />
            <path d="M30 50 L45 65 L70 35" fill="none" stroke="currentColor" strokeWidth="6" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
          {!collapsed && (
            <div className={styles.logoText}>
              <span className={styles.logoTitle}>CSIC</span>
              <span className={styles.logoSubtitle}>国家基础设施</span>
            </div>
          )}
        </div>
        <button className={styles.toggleButton} onClick={onToggle}>
          <svg viewBox="0 0 24 24" width="20" height="20" style={{ transform: collapsed ? 'rotate(180deg)' : 'none' }}>
            <path d="M15 18l-6-6 6-6" fill="none" stroke="currentColor" strokeWidth="2" />
          </svg>
        </button>
      </div>

      <nav className={styles.nav}>
        {navItems.map((section) => {
          const visibleItems = section.items.filter(item => shouldShowItem(item.permission));
          if (visibleItems.length === 0) return null;

          return (
            <div key={section.section} className={styles.navSection}>
              {!collapsed && (
                <div className={styles.sectionTitle}>{section.section}</div>
              )}
              <ul className={styles.navList}>
                {visibleItems.map((item) => (
                  <li key={item.path}>
                    <NavLink
                      to={item.path}
                      className={({ isActive }) =>
                        `${styles.navItem} ${isActive ? styles.active : ''}`
                      }
                      title={collapsed ? item.label : undefined}
                    >
                      <span className={styles.navIcon}>{renderIcon(item.icon)}</span>
                      {!collapsed && <span className={styles.navLabel}>{item.label}</span>}
                    </NavLink>
                  </li>
                ))}
              </ul>
            </div>
          );
        })}
      </nav>

      <div className={styles.footer}>
        {!collapsed && user && (
          <div className={styles.userInfo}>
            <div className={styles.userAvatar}>
              {user.name?.charAt(0) || 'U'}
            </div>
            <div className={styles.userDetails}>
              <span className={styles.userName}>{user.name}</span>
              <span className={styles.userRole}>{user.role}</span>
            </div>
          </div>
        )}
      </div>
    </aside>
  );
};

export default Sidebar;
