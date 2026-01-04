// CSIC Platform - Alert Console Page
// Incident management and alert monitoring interface

import React, { useState, useEffect } from 'react';
import { useAlertStore, useSystemStore, Alert } from '../store';

const AlertConsole: React.FC = () => {
  const { 
    alerts, 
    filters, 
    selectedAlertId, 
    unreadCount,
    acknowledgeAlert, 
    resolveAlert, 
    dismissAlert,
    setSelectedAlert,
    setFilters,
    loadAlerts,
  } = useAlertStore();
  
  const { status: systemStatus } = useSystemStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [activeTab, setActiveTab] = useState<'all' | 'active' | 'acknowledged' | 'resolved'>('all');

  useEffect(() => {
    loadAlerts();
  }, [loadAlerts]);

  const filteredAlerts = alerts.filter(alert => {
    // Filter by search query
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      if (!alert.title.toLowerCase().includes(query) && 
          !alert.description.toLowerCase().includes(query)) {
        return false;
      }
    }

    // Filter by tab
    switch (activeTab) {
      case 'active':
        return alert.status === 'ACTIVE';
      case 'acknowledged':
        return alert.status === 'ACKNOWLEDGED';
      case 'resolved':
        return alert.status === 'RESOLVED';
      default:
        return true;
    }
  });

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'CRITICAL': return 'severity-critical';
      case 'WARNING': return 'severity-warning';
      case 'INFO': return 'severity-info';
      default: return '';
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'ACTIVE': return 'badge-error';
      case 'ACKNOWLEDGED': return 'badge-warning';
      case 'RESOLVED': return 'badge-success';
      case 'DISMISSED': return 'badge-info';
      default: return 'badge-info';
    }
  };

  const selectedAlert = alerts.find(a => a.id === selectedAlertId);

  return (
    <div className="alert-console">
      <div className="console-header">
        <div className="header-left">
          <h1>警报控制台</h1>
          <span className="alert-count">{unreadCount} 个未处理警报</span>
        </div>
        <div className="header-actions">
          <button className="btn btn-secondary">
            <svg viewBox="0 0 24 24" className="btn-icon">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="7 10 12 15 17 10" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="15" x2="12" y2="3" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            导出报告
          </button>
          <button className="btn btn-primary">
            <svg viewBox="0 0 24 24" className="btn-icon">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="22 4 12 14.01 9 11.01" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            批量处理
          </button>
        </div>
      </div>

      <div className="console-stats">
        <div className="stat-card">
          <span className="stat-value">{alerts.filter(a => a.status === 'ACTIVE').length}</span>
          <span className="stat-label">活动警报</span>
        </div>
        <div className="stat-card warning">
          <span className="stat-value">{alerts.filter(a => a.severity === 'CRITICAL' && a.status === 'ACTIVE').length}</span>
          <span className="stat-label">严重警报</span>
        </div>
        <div className="stat-card success">
          <span className="stat-value">{alerts.filter(a => a.status === 'RESOLVED').length}</span>
          <span className="stat-label">已解决</span>
        </div>
        <div className="stat-card">
          <span className="stat-value">{Math.round((alerts.filter(a => a.status === 'RESOLVED').length / Math.max(alerts.length, 1)) * 100)}%</span>
          <span className="stat-label">解决率</span>
        </div>
      </div>

      <div className="console-content">
        <div className="alert-list-panel">
          <div className="list-header">
            <div className="search-box">
              <svg viewBox="0 0 24 24" className="search-icon">
                <circle cx="11" cy="11" r="8" fill="none" stroke="currentColor" strokeWidth="2" />
                <path d="M21 21l-4.35-4.35" fill="none" stroke="currentColor" strokeWidth="2" />
              </svg>
              <input
                type="text"
                placeholder="搜索警报..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="search-input"
              />
            </div>
            <div className="filter-tabs">
              <button 
                className={`tab ${activeTab === 'all' ? 'active' : ''}`}
                onClick={() => setActiveTab('all')}
              >
                全部
              </button>
              <button 
                className={`tab ${activeTab === 'active' ? 'active' : ''}`}
                onClick={() => setActiveTab('active')}
              >
                活动
              </button>
              <button 
                className={`tab ${activeTab === 'acknowledged' ? 'active' : ''}`}
                onClick={() => setActiveTab('acknowledged')}
              >
                已确认
              </button>
              <button 
                className={`tab ${activeTab === 'resolved' ? 'active' : ''}`}
                onClick={() => setActiveTab('resolved')}
              >
                已解决
              </button>
            </div>
          </div>

          <div className="alert-list">
            {filteredAlerts.length === 0 ? (
              <div className="empty-state">
                <svg viewBox="0 0 24 24" className="empty-icon">
                  <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" fill="none" stroke="currentColor" strokeWidth="2" />
                  <polyline points="22 4 12 14.01 9 11.01" fill="none" stroke="currentColor" strokeWidth="2" />
                </svg>
                <p>没有找到警报</p>
              </div>
            ) : (
              filteredAlerts.map(alert => (
                <div 
                  key={alert.id} 
                  className={`alert-item ${selectedAlertId === alert.id ? 'selected' : ''} ${alert.status === 'ACTIVE' ? 'unread' : ''}`}
                  onClick={() => setSelectedAlert(alert.id)}
                >
                  <div className="alert-severity-indicator">
                    <span className={`severity-dot ${getSeverityColor(alert.severity)}`}></span>
                  </div>
                  <div className="alert-item-content">
                    <div className="alert-item-header">
                      <span className="alert-title">{alert.title}</span>
                      <span className={`badge ${getStatusBadge(alert.status)}`}>{alert.status}</span>
                    </div>
                    <div className="alert-item-meta">
                      <span className="alert-category">{alert.category}</span>
                      <span className="alert-time">
                        {new Date(alert.createdAt).toLocaleString('zh-CN')}
                      </span>
                    </div>
                  </div>
                  {alert.status === 'ACTIVE' && (
                    <div className="alert-unread-badge"></div>
                  )}
                </div>
              ))
            )}
          </div>
        </div>

        <div className="alert-detail-panel">
          {selectedAlert ? (
            <>
              <div className="detail-header">
                <div className="detail-title-section">
                  <span className={`severity-badge-large ${getSeverityColor(selectedAlert.severity)}`}>
                    {selectedAlert.severity}
                  </span>
                  <h2>{selectedAlert.title}</h2>
                </div>
                <div className="detail-meta">
                  <span className="meta-item">
                    <svg viewBox="0 0 24 24" className="meta-icon">
                      <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
                      <path d="M12 6v6l4 2" fill="none" stroke="currentColor" strokeWidth="2" />
                    </svg>
                    {new Date(selectedAlert.createdAt).toLocaleString('zh-CN')}
                  </span>
                  <span className="meta-item">
                    <svg viewBox="0 0 24 24" className="meta-icon">
                      <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" fill="none" stroke="currentColor" strokeWidth="2" />
                      <circle cx="12" cy="10" r="3" fill="none" stroke="currentColor" strokeWidth="2" />
                    </svg>
                    {selectedAlert.source}
                  </span>
                  <span className="meta-item">
                    <svg viewBox="0 0 24 24" className="meta-icon">
                      <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" fill="none" stroke="currentColor" strokeWidth="2" />
                      <circle cx="12" cy="7" r="4" fill="none" stroke="currentColor" strokeWidth="2" />
                    </svg>
                    {selectedAlert.acknowledgedBy || '未分配'}
                  </span>
                </div>
              </div>

              <div className="detail-body">
                <div className="detail-section">
                  <h3>描述</h3>
                  <p className="description">{selectedAlert.description}</p>
                </div>

                {selectedAlert.metadata && (
                  <div className="detail-section">
                    <h3>元数据</h3>
                    <div className="metadata-grid">
                      {Object.entries(selectedAlert.metadata).map(([key, value]) => (
                        <div key={key} className="metadata-item">
                          <span className="metadata-key">{key}</span>
                          <span className="metadata-value">{String(value)}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                <div className="detail-section">
                  <h3>证据</h3>
                  <div className="evidence-list">
                    {selectedAlert.evidence.map((evidence, index) => (
                      <div key={index} className="evidence-item">
                        <span className="evidence-type">{evidence.type}</span>
                        <span className="evidence-value">{evidence.value}</span>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              <div className="detail-actions">
                {selectedAlert.status === 'ACTIVE' && (
                  <>
                    <button 
                      className="btn btn-primary"
                      onClick={() => acknowledgeAlert(selectedAlert.id, 'current-user')}
                    >
                      确认警报
                    </button>
                    <button className="btn btn-danger">阻止交易</button>
                    <button className="btn btn-outline">冻结钱包</button>
                  </>
                )}
                {selectedAlert.status === 'ACKNOWLEDGGED' && (
                  <>
                    <button 
                      className="btn btn-success"
                      onClick={() => resolveAlert(selectedAlert.id, 'current-user')}
                    >
                      标记已解决
                    </button>
                    <button 
                      className="btn btn-secondary"
                      onClick={() => dismissAlert(selectedAlert.id)}
                    >
                      标记为误报
                    </button>
                  </>
                )}
                <button className="btn btn-outline">
                  <svg viewBox="0 0 24 24" className="btn-icon">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" fill="none" stroke="currentColor" strokeWidth="2" />
                    <polyline points="7 10 12 15 17 10" fill="none" stroke="currentColor" strokeWidth="2" />
                    <line x1="12" y1="15" x2="12" y2="3" fill="none" stroke="currentColor" strokeWidth="2" />
                  </svg>
                  导出证据
                </button>
              </div>
            </>
          ) : (
            <div className="no-selection">
              <svg viewBox="0 0 24 24" className="selection-icon">
                <path d="M22 12h-4l-3 9L9 3l-3 9H2" fill="none" stroke="currentColor" strokeWidth="2" />
              </svg>
              <p>选择一个警报查看详情</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default AlertConsole;
