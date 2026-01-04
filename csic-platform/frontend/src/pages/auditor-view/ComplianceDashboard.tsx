// CSIC Platform - Compliance Dashboard Page
// Compliance monitoring and reporting interface for auditors

import React, { useState, useEffect } from 'react';
import { useExchangeStore, useWalletStore, useMinerStore } from '../../store';

const ComplianceDashboard: React.FC = () => {
  const { exchanges, loadExchanges: loadExchanges } = useExchangeStore();
  const { wallets, loadWallets } = useWalletStore();
  const { miners, loadMiners } = useMinerStore();
  
  const [selectedPeriod, setSelectedPeriod] = useState<'7d' | '30d' | '90d' | '1y'>('30d');
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadData = async () => {
      setIsLoading(true);
      await Promise.all([
        loadExchanges(),
        loadWallets(),
        loadMiners(),
      ]);
      setIsLoading(false);
    };
    loadData();
  }, [loadExchanges, loadWallets, loadMiners]);

  const complianceStats = {
    totalExchanges: exchanges.length,
    compliantExchanges: exchanges.filter(e => e.complianceScore >= 80).length,
    nonCompliantWallets: wallets.filter(w => w.status === 'BLACKLISTED').length,
    activeMiners: miners.filter(m => m.status === 'ACTIVE').length,
    avgComplianceScore: exchanges.length > 0 
      ? Math.round(exchanges.reduce((sum, e) => sum + e.complianceScore, 0) / exchanges.length)
      : 0,
  };

  const recentViolations = [
    { id: '1', type: 'KYC缺失', entity: 'CryptoExchange Pro', severity: 'MEDIUM', date: '2024-01-10' },
    { id: '2', type: '大额交易报告', entity: 'Digital Asset Hub', severity: 'LOW', date: '2024-01-08' },
    { id: '3', type: '审计延迟', entity: 'BlockTrade Global', severity: 'HIGH', date: '2024-01-05' },
  ];

  const upcomingObligations = [
    { id: '1', obligation: '季度合规报告', entity: 'CryptoExchange Pro', dueDate: '2024-01-15', status: 'PENDING' },
    { id: '2', obligation: '年度审计', entity: 'Digital Asset Hub', dueDate: '2024-01-20', status: 'IN_PROGRESS' },
    { id: '3', obligation: '风险评估更新', entity: 'Northern Mining Pool', dueDate: '2024-01-25', status: 'PENDING' },
  ];

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'HIGH': return 'var(--color-error)';
      case 'MEDIUM': return 'var(--color-warning)';
      case 'LOW': return 'var(--color-info)';
      default: return 'var(--color-text-muted)';
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'COMPLIANT': return 'badge-success';
      case 'NON_COMPLIANT': return 'badge-error';
      case 'UNDER_REVIEW': return 'badge-warning';
      default: return 'badge-info';
    }
  };

  if (isLoading) {
    return (
      <div className="compliance-dashboard loading">
        <div className="loading-container">
          <div className="loading-spinner large"></div>
          <p>加载合规数据...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="compliance-dashboard">
      <div className="page-header">
        <div className="header-left">
          <h1>合规仪表板</h1>
          <p>监管合规状态概览和风险管理</p>
        </div>
        <div className="header-actions">
          <div className="period-selector">
            <button 
              className={`period-btn ${selectedPeriod === '7d' ? 'active' : ''}`}
              onClick={() => setSelectedPeriod('7d')}
            >
              7天
            </button>
            <button 
              className={`period-btn ${selectedPeriod === '30d' ? 'active' : ''}`}
              onClick={() => setSelectedPeriod('30d')}
            >
              30天
            </button>
            <button 
              className={`period-btn ${selectedPeriod === '90d' ? 'active' : ''}`}
              onClick={() => setSelectedPeriod('90d')}
            >
              90天
            </button>
            <button 
              className={`period-btn ${selectedPeriod === '1y' ? 'active' : ''}`}
              onClick={() => setSelectedPeriod('1y')}
            >
              1年
            </button>
          </div>
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
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="14 2 14 8 20 8" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="18" x2="12" y2="12" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="9" y1="15" x2="15" y2="15" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            生成合规报告
          </button>
        </div>
      </div>

      <div className="stats-row">
        <div className="stat-card">
          <div className="stat-icon compliance">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="22 4 12 14.01 9 11.01" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{complianceStats.compliantExchanges}/{complianceStats.totalExchanges}</span>
            <span className="stat-label">合规交易所</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon score">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M12 6v6l4 2" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{complianceStats.avgComplianceScore}%</span>
            <span className="stat-label">平均合规分数</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon violations">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="9" x2="12" y2="13" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="17" x2="12.01" y2="17" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{complianceStats.nonCompliantWallets}</span>
            <span className="stat-label">黑名单钱包</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon miners">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M12 2L2 7l10 5 10-5-10-5z" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M2 17l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M2 12l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{complianceStats.activeMiners}</span>
            <span className="stat-label">活跃矿工</span>
          </div>
        </div>
      </div>

      <div className="dashboard-grid">
        <div className="card violations-card">
          <div className="card-header">
            <h3 className="card-title">最近违规</h3>
            <button className="view-all-btn">查看全部</button>
          </div>
          <div className="card-body">
            <div className="violations-list">
              {recentViolations.map(violation => (
                <div key={violation.id} className="violation-item">
                  <div className="violation-info">
                    <span className="violation-type">{violation.type}</span>
                    <span className="violation-entity">{violation.entity}</span>
                  </div>
                  <div className="violation-meta">
                    <span 
                      className="violation-severity"
                      style={{ color: getSeverityColor(violation.severity) }}
                    >
                      {violation.severity}
                    </span>
                    <span className="violation-date">{violation.date}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="card obligations-card">
          <div className="card-header">
            <h3 className="card-title">即将到期的义务</h3>
            <button className="view-all-btn">查看全部</button>
          </div>
          <div className="card-body">
            <div className="obligations-list">
              {upcomingObligations.map(obligation => (
                <div key={obligation.id} className="obligation-item">
                  <div className="obligation-info">
                    <span className="obligation-name">{obligation.obligation}</span>
                    <span className="obligation-entity">{obligation.entity}</span>
                  </div>
                  <div className="obligation-meta">
                    <span className="obligation-due">截止: {obligation.dueDate}</span>
                    <span className={`badge ${obligation.status === 'IN_PROGRESS' ? 'badge-warning' : 'badge-info'}`}>
                      {obligation.status === 'IN_PROGRESS' ? '进行中' : '待处理'}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="card exchange-compliance-card">
          <div className="card-header">
            <h3 className="card-title">交易所合规状态</h3>
          </div>
          <div className="card-body">
            <div className="compliance-list">
              {exchanges.map(exchange => (
                <div key={exchange.id} className="compliance-item">
                  <div className="compliance-info">
                    <span className="exchange-name">{exchange.name}</span>
                    <span className="exchange-jurisdiction">{exchange.jurisdiction}</span>
                  </div>
                  <div className="compliance-score-bar">
                    <div className="score-track">
                      <div 
                        className="score-fill"
                        style={{ 
                          width: `${exchange.complianceScore}%`,
                          backgroundColor: exchange.complianceScore >= 80 ? 'var(--color-success)' :
                                         exchange.complianceScore >= 60 ? 'var(--color-warning)' : 'var(--color-error)'
                        }}
                      ></div>
                    </div>
                    <span className="score-text">{exchange.complianceScore}%</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="card regulations-card">
          <div className="card-header">
            <h3 className="card-title">法规更新</h3>
          </div>
          <div className="card-body">
            <div className="regulations-list">
              <div className="regulation-item">
                <span className="regulation-badge new">新规</span>
                <div className="regulation-info">
                  <span className="regulation-title">反洗钱指令更新</span>
                  <span className="regulation-date">2024-01-15 生效</span>
                </div>
              </div>
              <div className="regulation-item">
                <span className="regulation-badge update">更新</span>
                <div className="regulation-info">
                  <span className="regulation-title">加密资产分类标准修订</span>
                  <span className="regulation-date">2024-02-01 生效</span>
                </div>
              </div>
              <div className="regulation-item">
                <span className="regulation-badge notice">通知</span>
                <div className="regulation-info">
                  <span className="regulation-title">年度报告要求变更</span>
                  <span className="regulation-date">2024-03-01 生效</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ComplianceDashboard;
