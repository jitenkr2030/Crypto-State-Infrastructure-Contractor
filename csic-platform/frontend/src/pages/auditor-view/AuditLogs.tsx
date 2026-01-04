// CSIC Platform - Audit Logs Page
// System audit trail and compliance monitoring interface

import React, { useState, useEffect } from 'react';
import { useAuthStore } from '../../store';

interface AuditLog {
  id: string;
  timestamp: Date;
  userId: string;
  username: string;
  action: string;
  resourceType: string;
  resourceId: string;
  details: string;
  ipAddress: string;
  userAgent: string;
  status: 'SUCCESS' | 'FAILURE';
}

const AuditLogs: React.FC = () => {
  const { user: currentUser } = useAuthStore();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [actionFilter, setActionFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);

  useEffect(() => {
    loadAuditLogs();
  }, []);

  const loadAuditLogs = async () => {
    setIsLoading(true);
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    const mockLogs: AuditLog[] = [
      {
        id: '1',
        timestamp: new Date(Date.now() - 3600000),
        userId: 'u1',
        username: 'admin',
        action: 'LOGIN',
        resourceType: 'session',
        resourceId: 'sess_123',
        details: '用户登录成功',
        ipAddress: '192.168.1.100',
        userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        status: 'SUCCESS',
      },
      {
        id: '2',
        timestamp: new Date(Date.now() - 7200000),
        userId: 'u2',
        username: 'regulator_1',
        action: 'SUSPEND_EXCHANGE',
        resourceType: 'exchange',
        resourceId: 'ex_456',
        details: '暂停交易所运营: BlockTrade Global',
        ipAddress: '192.168.1.101',
        userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
        status: 'SUCCESS',
      },
      {
        id: '3',
        timestamp: new Date(Date.now() - 10800000),
        userId: 'u1',
        username: 'admin',
        action: 'FREEZE_WALLET',
        resourceType: 'wallet',
        resourceId: 'w_789',
        details: '冻结钱包: 1A2B3C4D5E',
        ipAddress: '192.168.1.100',
        userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        status: 'SUCCESS',
      },
      {
        id: '4',
        timestamp: new Date(Date.now() - 14400000),
        userId: 'u3',
        username: 'auditor_1',
        action: 'EXPORT_REPORT',
        resourceType: 'report',
        resourceId: 'rpt_101',
        details: '导出合规报告',
        ipAddress: '192.168.1.102',
        userAgent: 'Mozilla/5.0 (X11; Linux x86_64)',
        status: 'SUCCESS',
      },
      {
        id: '5',
        timestamp: new Date(Date.now() - 18000000),
        userId: 'u2',
        username: 'regulator_1',
        action: 'UPDATE_SETTINGS',
        resourceType: 'system',
        resourceId: 'sys_cfg',
        details: '更新系统配置: rate_limit',
        ipAddress: '192.168.1.101',
        userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
        status: 'SUCCESS',
      },
      {
        id: '6',
        timestamp: new Date(Date.now() - 21600000),
        userId: 'u4',
        username: 'analyst_1',
        action: 'VIEW_ALERTS',
        resourceType: 'alert',
        resourceId: 'alert_all',
        details: '查看警报列表',
        ipAddress: '192.168.1.103',
        userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        status: 'SUCCESS',
      },
      {
        id: '7',
        timestamp: new Date(Date.now() - 25200000),
        userId: 'u1',
        username: 'admin',
        action: 'CREATE_USER',
        resourceType: 'user',
        resourceId: 'u5',
        details: '创建新用户: new_regulator',
        ipAddress: '192.168.1.100',
        userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        status: 'SUCCESS',
      },
      {
        id: '8',
        timestamp: new Date(Date.now() - 28800000),
        userId: 'u3',
        username: 'auditor_1',
        action: 'LOGIN_FAILED',
        resourceType: 'session',
        resourceId: 'sess_124',
        details: '登录失败: 密码错误',
        ipAddress: '192.168.1.102',
        userAgent: 'Mozilla/5.0 (X11; Linux x86_64)',
        status: 'FAILURE',
      },
    ];
    
    setLogs(mockLogs);
    setIsLoading(false);
  };

  const filteredLogs = logs.filter(log => {
    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      if (!log.username.toLowerCase().includes(query) &&
          !log.action.toLowerCase().includes(query) &&
          !log.details.toLowerCase().includes(query) &&
          !log.resourceType.toLowerCase().includes(query)) {
        return false;
      }
    }

    // Action filter
    if (actionFilter !== 'all' && log.action !== actionFilter) {
      return false;
    }

    // Status filter
    if (statusFilter !== 'all' && log.status !== statusFilter) {
      return false;
    }

    return true;
  });

  const getActionColor = (action: string) => {
    if (action.includes('LOGIN')) return 'action-login';
    if (action.includes('CREATE') || action.includes('UPDATE') || action.includes('DELETE')) return 'action-modify';
    if (action.includes('SUSPEND') || action.includes('FREEZE') || action.includes('REVOKE')) return 'action-critical';
    if (action.includes('FAILED')) return 'action-failure';
    return 'action-view';
  };

  const getActionLabel = (action: string) => {
    const labels: Record<string, string> = {
      LOGIN: '登录',
      LOGOUT: '登出',
      LOGIN_FAILED: '登录失败',
      CREATE_USER: '创建用户',
      UPDATE_SETTINGS: '更新设置',
      SUSPEND_EXCHANGE: '暂停交易所',
      FREEZE_WALLET: '冻结钱包',
      EXPORT_REPORT: '导出报告',
      VIEW_ALERTS: '查看警报',
    };
    return labels[action] || action;
  };

  const formatTimestamp = (date: Date) => {
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  const stats = {
    totalLogs: logs.length,
    todayLogs: logs.filter(l => {
      const today = new Date();
      return l.timestamp.toDateString() === today.toDateString();
    }).length,
    failures: logs.filter(l => l.status === 'FAILURE').length,
    criticalActions: logs.filter(l => 
      l.action.includes('SUSPEND') || 
      l.action.includes('FREEZE') || 
      l.action.includes('REVOKE')
    ).length,
  };

  return (
    <div className="audit-logs-page">
      <div className="page-header">
        <div className="header-left">
          <h1>审计日志</h1>
          <p>系统操作审计追踪和安全事件记录</p>
        </div>
        <div className="header-actions">
          <button className="btn btn-secondary">
            <svg viewBox="0 0 24 24" className="btn-icon">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="7 10 12 15 17 10" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="15" x2="12" y2="3" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            导出日志
          </button>
          <button className="btn btn-secondary">
            <svg viewBox="0 0 24 24" className="btn-icon">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            验证完整性
          </button>
          <button className="btn btn-primary">
            <svg viewBox="0 0 24 24" className="btn-icon">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="14 2 14 8 20 8" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            生成审计报告
          </button>
        </div>
      </div>

      <div className="stats-row">
        <div className="stat-card">
          <span className="stat-value">{stats.totalLogs}</span>
          <span className="stat-label">总日志数</span>
        </div>
        <div className="stat-card">
          <span className="stat-value">{stats.todayLogs}</span>
          <span className="stat-label">今日操作</span>
        </div>
        <div className="stat-card warning">
          <span className="stat-value">{stats.failures}</span>
          <span className="stat-label">失败操作</span>
        </div>
        <div className="stat-card danger">
          <span className="stat-value">{stats.criticalActions}</span>
          <span className="stat-label">关键操作</span>
        </div>
      </div>

      <div className="content-card">
        <div className="card-header">
          <div className="filters">
            <div className="search-box">
              <svg viewBox="0 0 24 24" className="search-icon">
                <circle cx="11" cy="11" r="8" fill="none" stroke="currentColor" strokeWidth="2" />
                <path d="M21 21l-4.35-4.35" fill="none" stroke="currentColor" strokeWidth="2" />
              </svg>
              <input
                type="text"
                placeholder="搜索日志..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="search-input"
              />
            </div>
            <select
              className="filter-select"
              value={actionFilter}
              onChange={(e) => setActionFilter(e.target.value)}
            >
              <option value="all">所有操作</option>
              <option value="LOGIN">登录操作</option>
              <option value="CREATE_USER">创建用户</option>
              <option value="SUSPEND_EXCHANGE">暂停交易所</option>
              <option value="FREEZE_WALLET">冻结钱包</option>
              <option value="EXPORT_REPORT">导出报告</option>
            </select>
            <select
              className="filter-select"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
            >
              <option value="all">所有状态</option>
              <option value="SUCCESS">成功</option>
              <option value="FAILURE">失败</option>
            </select>
          </div>
          <span className="result-count">{filteredLogs.length} 条记录</span>
        </div>

        {isLoading ? (
          <div className="loading-state">
            <div className="loading-spinner large"></div>
            <p>加载审计日志...</p>
          </div>
        ) : (
          <div className="logs-table-container">
            <table className="table logs-table">
              <thead>
                <tr>
                  <th>时间戳</th>
                  <th>用户</th>
                  <th>操作</th>
                  <th>资源类型</th>
                  <th>详情</th>
                  <th>IP地址</th>
                  <th>状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {filteredLogs.map(log => (
                  <tr 
                    key={log.id} 
                    className={`log-row ${selectedLog?.id === log.id ? 'selected' : ''}`}
                    onClick={() => setSelectedLog(log)}
                  >
                    <td>
                      <code className="timestamp">{formatTimestamp(log.timestamp)}</code>
                    </td>
                    <td>
                      <div className="user-cell">
                        <span className="username">{log.username}</span>
                        <span className="user-id">{log.userId}</span>
                      </div>
                    </td>
                    <td>
                      <span className={`action-badge ${getActionColor(log.action)}`}>
                        {getActionLabel(log.action)}
                      </span>
                    </td>
                    <td>
                      <code className="resource-type">{log.resourceType}</code>
                    </td>
                    <td>
                      <span className="details" title={log.details}>{log.details}</span>
                    </td>
                    <td>
                      <code className="ip-address">{log.ipAddress}</code>
                    </td>
                    <td>
                      <span className={`status-badge ${log.status === 'SUCCESS' ? 'status-success' : 'status-failure'}`}>
                        {log.status === 'SUCCESS' ? '成功' : '失败'}
                      </span>
                    </td>
                    <td>
                      <button className="action-btn" title="查看详情">
                        <svg viewBox="0 0 24 24" className="action-icon">
                          <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" fill="none" stroke="currentColor" strokeWidth="2" />
                          <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" strokeWidth="2" />
                        </svg>
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {selectedLog && (
        <div className="log-detail-panel">
          <div className="detail-header">
            <h3>日志详情</h3>
            <button className="close-btn" onClick={() => setSelectedLog(null)}>
              <svg viewBox="0 0 24 24" className="close-icon">
                <path d="M18 6L6 18M6 6l12 12" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
              </svg>
            </button>
          </div>
          <div className="detail-content">
            <div className="detail-row">
              <span className="detail-label">日志ID</span>
              <code className="detail-value">{selectedLog.id}</code>
            </div>
            <div className="detail-row">
              <span className="detail-label">时间戳</span>
              <code className="detail-value">{formatTimestamp(selectedLog.timestamp)}</code>
            </div>
            <div className="detail-row">
              <span className="detail-label">用户</span>
              <span className="detail-value">{selectedLog.username} ({selectedLog.userId})</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">操作</span>
              <span className={`action-badge ${getActionColor(selectedLog.action)}`}>
                {getActionLabel(selectedLog.action)}
              </span>
            </div>
            <div className="detail-row">
              <span className="detail-label">资源类型</span>
              <code className="detail-value">{selectedLog.resourceType}</code>
            </div>
            <div className="detail-row">
              <span className="detail-label">资源ID</span>
              <code className="detail-value">{selectedLog.resourceId}</code>
            </div>
            <div className="detail-row">
              <span className="detail-label">详情</span>
              <span className="detail-value">{selectedLog.details}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">IP地址</span>
              <code className="detail-value">{selectedLog.ipAddress}</code>
            </div>
            <div className="detail-row">
              <span className="detail-label">用户代理</span>
              <span className="detail-value user-agent">{selectedLog.userAgent}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">状态</span>
              <span className={`status-badge ${selectedLog.status === 'SUCCESS' ? 'status-success' : 'status-failure'}`}>
                {selectedLog.status === 'SUCCESS' ? '成功' : '失败'}
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default AuditLogs;
