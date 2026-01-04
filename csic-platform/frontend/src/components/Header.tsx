// Header Component - Top header bar for CSIC Platform
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore, useAlertStore } from '../store';
import styles from './Header.module.css';

interface HeaderProps {
  systemStatus: {
    status: string;
    lastChecked: string;
    uptime: number;
  };
}

const Header: React.FC<HeaderProps> = ({ systemStatus }) => {
  const navigate = useNavigate();
  const { user, logout, isAuthenticated } = useAuthStore();
  const { alerts } = useAlertStore();
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [showNotifications, setShowNotifications] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  const unreadAlerts = alerts.filter(alert => !alert.read).length;

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const formatUptime = (seconds: number): string => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    
    if (days > 0) {
      return `${days}d ${hours}h ${mins}m`;
    } else if (hours > 0) {
      return `${hours}h ${mins}m`;
    }
    return `${mins}m`;
  };

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'operational':
        return '#22c55e';
      case 'degraded':
        return '#f59e0b';
      case 'outage':
        return '#ef4444';
      default:
        return '#64748b';
    }
  };

  return (
    <header className={styles.header}>
      <div className={styles.left}>
        <div className={styles.searchContainer}>
          <svg viewBox="0 0 24 24" width="18" height="18" className={styles.searchIcon}>
            <circle cx="11" cy="11" r="8" fill="none" stroke="currentColor" strokeWidth="2" />
            <path d="M21 21l-4.35-4.35" fill="none" stroke="currentColor" strokeWidth="2" />
          </svg>
          <input
            type="text"
            placeholder="搜索交易所、钱包、交易..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className={styles.searchInput}
          />
          <span className={styles.searchShortcut}>⌘K</span>
        </div>
      </div>

      <div className={styles.right}>
        {/* System Status */}
        <div className={styles.statusIndicator}>
          <span 
            className={styles.statusDot} 
            style={{ backgroundColor: getStatusColor(systemStatus.status) }}
          />
          <span className={styles.statusText}>
            {systemStatus.status === 'operational' ? '系统正常运行' : 
             systemStatus.status === 'degraded' ? '系统性能下降' : 
             systemStatus.status === 'outage' ? '系统故障' : '未知状态'}
          </span>
          <span className={styles.uptime}>
            运行时间: {formatUptime(systemStatus.uptime)}
          </span>
        </div>

        <div className={styles.divider} />

        {/* Notifications */}
        <div className={styles.notificationWrapper}>
          <button 
            className={styles.iconButton}
            onClick={() => setShowNotifications(!showNotifications)}
          >
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M13.73 21a2 2 0 0 1-3.46 0" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            {unreadAlerts > 0 && (
              <span className={styles.badge}>{unreadAlerts > 99 ? '99+' : unreadAlerts}</span>
            )}
          </button>

          {showNotifications && (
            <div className={styles.dropdown}>
              <div className={styles.dropdownHeader}>
                <h3>通知</h3>
                <button className={styles.markAllRead}>全部已读</button>
              </div>
              <div className={styles.notificationList}>
                {alerts.slice(0, 5).map((alert) => (
                  <div key={alert.id} className={`${styles.notificationItem} ${!alert.read ? styles.unread : ''}`}>
                    <div className={`${styles.notificationIcon} ${styles[`severity${alert.severity}`]}`}>
                      <svg viewBox="0 0 24 24" width="16" height="16">
                        {alert.severity === 'critical' ? (
                          <path d="M12 2L2 7v10l10 5 10-5V7L12 2z" fill="none" stroke="currentColor" strokeWidth="2" />
                        ) : alert.severity === 'warning' ? (
                          <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" fill="none" stroke="currentColor" strokeWidth="2" />
                        ) : (
                          <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
                        )}
                      </svg>
                    </div>
                    <div className={styles.notificationContent}>
                      <span className={styles.notificationTitle}>{alert.title}</span>
                      <span className={styles.notificationTime}>
                        {new Date(alert.timestamp).toLocaleString('zh-CN')}
                      </span>
                    </div>
                  </div>
                ))}
                {alerts.length === 0 && (
                  <div className={styles.emptyState}>
                    <p>暂无通知</p>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>

        {/* User Menu */}
        <div className={styles.userWrapper}>
          <button 
            className={styles.userButton}
            onClick={() => setShowUserMenu(!showUserMenu)}
          >
            <div className={styles.userAvatar}>
              {user?.name?.charAt(0) || 'U'}
            </div>
            <span className={styles.userName}>{user?.name || '用户'}</span>
            <svg viewBox="0 0 24 24" width="16" height="16" className={styles.chevron}>
              <path d="M6 9l6 6 6-6" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </button>

          {showUserMenu && (
            <div className={styles.dropdown}>
              <div className={styles.dropdownHeader}>
                <div className={styles.userInfo}>
                  <span className={styles.dropdownName}>{user?.name}</span>
                  <span className={styles.dropdownEmail}>{user?.email}</span>
                </div>
              </div>
              <div className={styles.dropdownMenu}>
                <button className={styles.menuItem}>
                  <svg viewBox="0 0 24 24" width="18" height="18">
                    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" fill="none" stroke="currentColor" strokeWidth="2" />
                    <circle cx="12" cy="7" r="4" fill="none" stroke="currentColor" strokeWidth="2" />
                  </svg>
                  个人信息
                </button>
                <button className={styles.menuItem}>
                  <svg viewBox="0 0 24 24" width="18" height="18">
                    <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" strokeWidth="2" />
                    <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" fill="none" stroke="currentColor" strokeWidth="2" />
                  </svg>
                  设置
                </button>
                <div className={styles.divider} />
                <button className={`${styles.menuItem} ${styles.danger}`} onClick={handleLogout}>
                  <svg viewBox="0 0 24 24" width="18" height="18">
                    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" fill="none" stroke="currentColor" strokeWidth="2" />
                    <polyline points="16 17 21 12 16 7" fill="none" stroke="currentColor" strokeWidth="2" />
                    <line x1="21" y1="12" x2="9" y2="12" fill="none" stroke="currentColor" strokeWidth="2" />
                  </svg>
                  退出登录
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </header>
  );
};

export default Header;
