// CSIC Platform - Exchange List Page
// Exchange management and monitoring interface for regulators

import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useExchangeStore, Exchange } from '../../store';

const ExchangeList: React.FC = () => {
  const navigate = useNavigate();
  const { 
    exchanges, 
    loadExchanges, 
    selectExchange, 
    suspendExchange,
    revokeLicense,
    isLoading,
    error 
  } = useExchangeStore();
  
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [sortBy, setSortBy] = useState<'name' | 'complianceScore' | 'riskLevel'>('complianceScore');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');

  useEffect(() => {
    loadExchanges();
  }, [loadExchanges]);

  const filteredExchanges = exchanges
    .filter(exchange => {
      // Search filter
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        if (!exchange.name.toLowerCase().includes(query) &&
            !exchange.licenseNumber.toLowerCase().includes(query) &&
            !exchange.jurisdiction.toLowerCase().includes(query)) {
          return false;
        }
      }

      // Status filter
      if (statusFilter !== 'all' && exchange.status !== statusFilter) {
        return false;
      }

      return true;
    })
    .sort((a, b) => {
      let comparison = 0;
      switch (sortBy) {
        case 'name':
          comparison = a.name.localeCompare(b.name);
          break;
        case 'complianceScore':
          comparison = a.complianceScore - b.complianceScore;
          break;
        case 'riskLevel':
          const riskOrder = { CRITICAL: 4, HIGH: 3, MEDIUM: 2, LOW: 1 };
          comparison = riskOrder[a.riskLevel] - riskOrder[b.riskLevel];
          break;
      }
      return sortOrder === 'asc' ? comparison : -comparison;
    });

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'ACTIVE': return 'badge-success';
      case 'SUSPENDED': return 'badge-warning';
      case 'REVOKED': return 'badge-error';
      case 'PENDING': return 'badge-info';
      default: return 'badge-info';
    }
  };

  const getRiskBadge = (riskLevel: string) => {
    switch (riskLevel) {
      case 'LOW': return 'risk-low';
      case 'MEDIUM': return 'risk-medium';
      case 'HIGH': return 'risk-high';
      case 'CRITICAL': return 'risk-critical';
      default: return 'risk-medium';
    }
  };

  const getComplianceColor = (score: number) => {
    if (score >= 80) return 'var(--color-success)';
    if (score >= 60) return 'var(--color-warning)';
    return 'var(--color-error)';
  };

  const handleExchangeClick = (exchange: Exchange) => {
    selectExchange(exchange.id);
    navigate(`/exchanges/${exchange.id}`);
  };

  const handleSuspend = (e: React.MouseEvent, exchangeId: string) => {
    e.stopPropagation();
    if (window.confirm('确定要暂停该交易所的运营吗？')) {
      suspendExchange(exchangeId);
    }
  };

  const handleRevoke = (e: React.MouseEvent, exchangeId: string) => {
    e.stopPropagation();
    if (window.confirm('确定要撤销该交易所的牌照吗？此操作不可撤销。')) {
      revokeLicense(exchangeId);
    }
  };

  const stats = {
    total: exchanges.length,
    active: exchanges.filter(e => e.status === 'ACTIVE').length,
    suspended: exchanges.filter(e => e.status === 'SUSPENDED').length,
    revoked: exchanges.filter(e => e.status === 'REVOKED').length,
    avgCompliance: exchanges.length > 0 
      ? Math.round(exchanges.reduce((sum, e) => sum + e.complianceScore, 0) / exchanges.length)
      : 0,
  };

  return (
    <div className="exchange-list-page">
      <div className="page-header">
        <div className="header-left">
          <h1>交易所管理</h1>
          <p>监管范围内的加密货币交易所列表</p>
        </div>
        <div className="header-actions">
          <button className="btn btn-secondary">
            <svg viewBox="0 0 24 24" className="btn-icon">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="7 10 12 15 17 10" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="15" x2="12" y2="3" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            导出列表
          </button>
          <button className="btn btn-primary">
            <svg viewBox="0 0 24 24" className="btn-icon">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="8" x2="12" y2="16" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="8" y1="12" x2="16" y2="12" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            注册新交易所
          </button>
        </div>
      </div>

      <div className="stats-row">
        <div className="stat-card">
          <div className="stat-icon exchanges">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M12 2L2 7l10 5 10-5-10-5z" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M2 17l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M2 12l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.total}</span>
            <span className="stat-label">注册交易所</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon active">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="22 4 12 14.01 9 11.01" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.active}</span>
            <span className="stat-label">正常运营</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon warning">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="9" x2="12" y2="13" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="17" x2="12.01" y2="17" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.suspended}</span>
            <span className="stat-label">已暂停</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon info">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M12 2L2 7l10 5 10-5-10-5z" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M2 17l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M2 12l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.avgCompliance}%</span>
            <span className="stat-label">平均合规分数</span>
          </div>
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
                placeholder="搜索交易所..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="search-input"
              />
            </div>
            <select 
              className="filter-select"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
            >
              <option value="all">所有状态</option>
              <option value="ACTIVE">正常运营</option>
              <option value="SUSPENDED">已暂停</option>
              <option value="REVOKED">已撤销</option>
            </select>
            <select
              className="sort-select"
              value={`${sortBy}-${sortOrder}`}
              onChange={(e) => {
                const [field, order] = e.target.value.split('-');
                setSortBy(field as 'name' | 'complianceScore' | 'riskLevel');
                setSortOrder(order as 'asc' | 'desc');
              }}
            >
              <option value="complianceScore-desc">合规分数（高到低）</option>
              <option value="complianceScore-asc">合规分数（低到高）</option>
              <option value="name-asc">名称（A-Z）</option>
              <option value="name-desc">名称（Z-A）</option>
              <option value="riskLevel-desc">风险等级（高到低）</option>
              <option value="riskLevel-asc">风险等级（低到高）</option>
            </select>
          </div>
          <span className="result-count">{filteredExchanges.length} 个结果</span>
        </div>

        {error && (
          <div className="error-banner">
            <svg viewBox="0 0 24 24" className="error-icon">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M12 8v4M12 16h.01" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
            </svg>
            <span>{error}</span>
          </div>
        )}

        {isLoading ? (
          <div className="loading-state">
            <div className="loading-spinner large"></div>
            <p>加载交易所数据...</p>
          </div>
        ) : (
          <div className="exchange-table-container">
            <table className="table exchange-table">
              <thead>
                <tr>
                  <th>交易所名称</th>
                  <th>牌照编号</th>
                  <th>司法管辖区</th>
                  <th>合规分数</th>
                  <th>风险等级</th>
                  <th>状态</th>
                  <th>注册日期</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {filteredExchanges.map(exchange => (
                  <tr 
                    key={exchange.id} 
                    onClick={() => handleExchangeClick(exchange)}
                    className="exchange-row"
                  >
                    <td>
                      <div className="exchange-name-cell">
                        <div className="exchange-avatar">
                          {exchange.name.charAt(0)}
                        </div>
                        <div className="exchange-info">
                          <span className="exchange-name">{exchange.name}</span>
                          <span className="exchange-website">{exchange.website}</span>
                        </div>
                      </div>
                    </td>
                    <td>
                      <code className="license-number">{exchange.licenseNumber}</code>
                    </td>
                    <td>{exchange.jurisdiction}</td>
                    <td>
                      <div className="compliance-score">
                        <div className="score-bar">
                          <div 
                            className="score-fill"
                            style={{ 
                              width: `${exchange.complianceScore}%`,
                              backgroundColor: getComplianceColor(exchange.complianceScore)
                            }}
                          ></div>
                        </div>
                        <span className="score-value" style={{ color: getComplianceColor(exchange.complianceScore) }}>
                          {exchange.complianceScore}%
                        </span>
                      </div>
                    </td>
                    <td>
                      <span className={`risk-badge ${getRiskBadge(exchange.riskLevel)}`}>
                        {exchange.riskLevel}
                      </span>
                    </td>
                    <td>
                      <span className={`badge ${getStatusBadge(exchange.status)}`}>
                        {exchange.status === 'ACTIVE' ? '正常运营' : 
                         exchange.status === 'SUSPENDED' ? '已暂停' : '已撤销'}
                      </span>
                    </td>
                    <td>
                      {new Date(exchange.registrationDate).toLocaleDateString('zh-CN')}
                    </td>
                    <td>
                      <div className="action-buttons" onClick={(e) => e.stopPropagation()}>
                        <button 
                          className="action-btn" 
                          title="查看详情"
                          onClick={() => handleExchangeClick(exchange)}
                        >
                          <svg viewBox="0 0 24 24" className="action-icon">
                            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" fill="none" stroke="currentColor" strokeWidth="2" />
                            <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" strokeWidth="2" />
                          </svg>
                        </button>
                        {exchange.status === 'ACTIVE' && (
                          <>
                            <button 
                              className="action-btn warning" 
                              title="暂停运营"
                              onClick={(e) => handleSuspend(e, exchange.id)}
                            >
                              <svg viewBox="0 0 24 24" className="action-icon">
                                <rect x="6" y="4" width="4" height="16" fill="none" stroke="currentColor" strokeWidth="2" />
                                <rect x="14" y="4" width="4" height="16" fill="none" stroke="currentColor" strokeWidth="2" />
                              </svg>
                            </button>
                            <button 
                              className="action-btn danger" 
                              title="撤销牌照"
                              onClick={(e) => handleRevoke(e, exchange.id)}
                            >
                              <svg viewBox="0 0 24 24" className="action-icon">
                                <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
                                <path d="M15 9l-6 6M9 9l6 6" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
                              </svg>
                            </button>
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};

export default ExchangeList;
