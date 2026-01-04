import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore, useAlertStore, useSystemStore } from '../../store';

interface HeaderProps {
  alertCount: number;
}

const Header: React.FC<HeaderProps> = ({ alertCount }) => {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();
  const { alerts, acknowledgeAlert } = useAlertStore();
  const { status, hsmConnected } = useSystemStore();
  const [showNotifications, setShowNotifications] = useState(false);
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [currentTime, setCurrentTime] = useState(new Date());

  // Update time every second
  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentTime(new Date());
    }, 1000);
    return () => clearInterval(timer);
  }, []);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const getStatusColor = () => {
    switch (status) {
      case 'EMERGENCY':
        return 'status-critical';
      case 'DEGRADED':
        return 'status-warning';
      default:
        return 'status-normal';
    }
  };

  const activeAlerts = alerts.filter((a) => a.status === 'ACTIVE');

  return (
    <header className="header">
      <div className="header-left">
        <div className={`system-status ${getStatusColor()}`}>
          <span className="status-indicator"></span>
          <span className="status-text">
            {status === 'ONLINE' && '系统正常运行'}
            {status === 'DEGRADED' && '系统降级运行'}
            {status === 'EMERGENCY' && '紧急停止状态'}
            {status === 'OFFLINE' && '系统离线'}
          </span>
        </div>
      </div>

      <div className="header-center">
        <div className="header-time">
          {currentTime.toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            hour12: false,
          })}
          <span className="timezone">UTC</span>
        </div>
      </div>

      <div className="header-right">
        {/* HSM Status */}
        <div className={`hsm-status ${hsmConnected ? 'connected' : 'disconnected'}`}>
          <span className="hsm-icon">
            <KeyIcon />
          </span>
          <span className="hsm-text">
            {hsmConnected ? 'HSM已连接' : 'HSM断开'}
          </span>
        </div>

        {/* Notifications */}
        <div className="notification-wrapper">
          <button
            className="notification-btn"
            onClick={() => setShowNotifications(!showNotifications)}
          >
            <BellIcon />
            {alertCount > 0 && <span className="notification-badge">{alertCount}</span>}
          </button>

          {showNotifications && (
            <div className="notification-dropdown">
              <div className="dropdown-header">
                <h4>系统警报</h4>
                <span className="alert-count">{activeAlerts.length}个活动警报</span>
              </div>
              <div className="notification-list">
                {activeAlerts.slice(0, 5).map((alert) => (
                  <div key={alert.id} className={`notification-item severity-${alert.severity.toLowerCase()}`}>
                    <div className="notification-content">
                      <div className="notification-title">{alert.title}</div>
                      <div className="notification-desc">{alert.description}</div>
                      <div className="notification-meta">
                        <span className="alert-time">
                          {new Date(alert.createdAt).toLocaleString('zh-CN')}
                        </span>
                        <span className={`severity-badge ${alert.severity.toLowerCase()}`}>
                          {alert.severity}
                        </span>
                      </div>
                    </div>
                    <button
                      className="acknowledge-btn"
                      onClick={() => acknowledgeAlert(alert.id)}
                    >
                      确认
                    </button>
                  </div>
                ))}
              </div>
              <div className="dropdown-footer">
                <button onClick={() => navigate('/security')}>查看全部警报</button>
              </div>
            </div>
          )}
        </div>

        {/* User Menu */}
        <div className="user-menu-wrapper">
          <button
            className="user-btn"
            onClick={() => setShowUserMenu(!showUserMenu)}
          >
            <div className="user-avatar">
              {user?.username?.charAt(0).toUpperCase() || 'U'}
            </div>
            <div className="user-info">
              <span className="user-name">{user?.username || '用户'}</span>
              <span className="user-role">{user?.role || 'VIEWER'}</span>
            </div>
          </button>

          {showUserMenu && (
            <div className="user-dropdown">
              <div className="dropdown-header">
                <div className="user-full-info">
                  <div className="user-avatar large">
                    {user?.username?.charAt(0).toUpperCase() || 'U'}
                  </div>
                  <div className="user-details">
                    <span className="user-name">{user?.username}</span>
                    <span className="user-email">{user?.email}</span>
                    <span className="user-department">{user?.department}</span>
                  </div>
                </div>
              </div>
              <div className="dropdown-divider"></div>
              <ul className="dropdown-menu">
                <li>
                  <button onClick={() => navigate('/settings')}>
                    <SettingsIcon />
                    系统设置
                  </button>
                </li>
                <li>
                  <button onClick={handleLogout}>
                    <LogoutIcon />
                    退出登录
                  </button>
                </li>
              </ul>
            </div>
          )}
        </div>
      </div>
    </header>
  );
};

// Icon Components
const BellIcon: React.FC = () => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
    <path d="M13.73 21a2 2 0 0 1-3.46 0" />
  </svg>
);

const KeyIcon: React.FC = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" />
  </svg>
);

const SettingsIcon: React.FC = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <circle cx="12" cy="12" r="3" />
    <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
  </svg>
);

const LogoutIcon: React.FC = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
    <polyline points="16,17 21,12 16,7" />
    <line x1="21" y1="12" x2="9" y2="12" />
  </svg>
);

export default Header;
