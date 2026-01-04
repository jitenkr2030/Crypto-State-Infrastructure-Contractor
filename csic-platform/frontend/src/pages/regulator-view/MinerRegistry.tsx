// CSIC Platform - Mining Registry Page
// Mining operations monitoring and management interface

import React, { useState, useEffect } from 'react';
import { useMinerStore, Miner } from '../../store';

const MinerRegistry: React.FC = () => {
  const { miners, loadMiners, suspendMiner, isLoading, error } = useMinerStore();
  
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [complianceFilter, setComplianceFilter] = useState<string>('all');
  const [selectedMiner, setSelectedMiner] = useState<Miner | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);

  useEffect(() => {
    loadMiners();
  }, [loadMiners]);

  const filteredMiners = miners.filter(miner => {
    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      if (!miner.name.toLowerCase().includes(query) &&
          !miner.licenseNumber.toLowerCase().includes(query) &&
          !miner.jurisdiction.toLowerCase().includes(query)) {
        return false;
      }
    }

    // Status filter
    if (statusFilter !== 'all' && miner.status !== statusFilter) {
      return false;
    }

    // Compliance filter
    if (complianceFilter !== 'all' && miner.complianceStatus !== complianceFilter) {
      return false;
    }

    return true;
  });

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'ACTIVE': return 'badge-success';
      case 'SUSPENDED': return 'badge-warning';
      case 'OFFLINE': return 'badge-info';
      default: return 'badge-info';
    }
  };

  const getComplianceBadge = (status: string) => {
    switch (status) {
      case 'COMPLIANT': return 'compliance-compliant';
      case 'NON_COMPLIANT': return 'compliance-non-compliant';
      case 'UNDER_REVIEW': return 'compliance-under-review';
      default: return 'compliance-under-review';
    }
  };

  const formatHashRate = (hashRate: number) => {
    if (hashRate >= 1000) {
      return `${(hashRate / 1000).toFixed(2)} EH/s`;
    }
    return `${hashRate.toFixed(0)} PH/s`;
  };

  const formatEnergy = (energy: number) => {
    return `${energy.toLocaleString()} MW`;
  };

  const handleSuspend = (e: React.MouseEvent, minerId: string) => {
    e.stopPropagation();
    if (window.confirm('确定要暂停该矿工的运营吗？')) {
      suspendMiner(minerId);
    }
  };

  const stats = {
    total: miners.length,
    active: miners.filter(m => m.status === 'ACTIVE').length,
    suspended: miners.filter(m => m.status === 'SUSPENDED').length,
    totalHashRate: miners.reduce((sum, m) => sum + m.hashRate, 0),
    totalEnergy: miners.reduce((sum, m) => sum + m.energyConsumption, 0),
  };

  return (
    <div className="miner-registry-page">
      <div className="page-header">
        <div className="header-left">
          <h1>矿工注册管理</h1>
          <p>监管范围内的加密货币矿工列表</p>
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
          <button 
            className="btn btn-primary"
            onClick={() => setShowAddModal(true)}
          >
            <svg viewBox="0 0 24 24" className="btn-icon">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="8" x2="12" y2="16" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="8" y1="12" x2="16" y2="12" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            注册新矿工
          </button>
        </div>
      </div>

      <div className="stats-row">
        <div className="stat-card">
          <div className="stat-icon miners">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M12 2L2 7l10 5 10-5-10-5z" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M2 17l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M2 12l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.total}</span>
            <span className="stat-label">注册矿工</span>
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
            <span className="stat-label">活跃矿工</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon hashrate">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M12 6v6l4 2" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{formatHashRate(stats.totalHashRate)}</span>
            <span className="stat-label">总算力</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon energy">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{formatEnergy(stats.totalEnergy)}</span>
            <span className="stat-label">总能耗</span>
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
                placeholder="搜索矿工..."
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
              <option value="ACTIVE">活跃</option>
              <option value="SUSPENDED">已暂停</option>
              <option value="OFFLINE">离线</option>
            </select>
            <select
              className="filter-select"
              value={complianceFilter}
              onChange={(e) => setComplianceFilter(e.target.value)}
            >
              <option value="all">所有合规状态</option>
              <option value="COMPLIANT">合规</option>
              <option value="UNDER_REVIEW">审核中</option>
              <option value="NON_COMPLIANT">不合规</option>
            </select>
          </div>
          <span className="result-count">{filteredMiners.length} 个结果</span>
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
            <p>加载矿工数据...</p>
          </div>
        ) : (
          <div className="miner-grid">
            {filteredMiners.map(miner => (
              <div 
                key={miner.id} 
                className={`miner-card ${selectedMiner?.id === miner.id ? 'selected' : ''}`}
                onClick={() => setSelectedMiner(miner)}
              >
                <div className="miner-header">
                  <div className="miner-avatar">
                    {miner.name.charAt(0)}
                  </div>
                  <div className="miner-info">
                    <h3 className="miner-name">{miner.name}</h3>
                    <code className="miner-license">{miner.licenseNumber}</code>
                  </div>
                  <div className="miner-badges">
                    <span className={`badge ${getStatusBadge(miner.status)}`}>
                      {miner.status === 'ACTIVE' ? '活跃' : 
                       miner.status === 'SUSPENDED' ? '已暂停' : '离线'}
                    </span>
                  </div>
                </div>

                <div className="miner-details">
                  <div className="detail-row">
                    <span className="detail-label">司法管辖区</span>
                    <span className="detail-value">{miner.jurisdiction}</span>
                  </div>
                  <div className="detail-row">
                    <span className="detail-label">能源来源</span>
                    <span className="detail-value">{miner.energySource}</span>
                  </div>
                  <div className="detail-row">
                    <span className="detail-label">算力</span>
                    <span className="detail-value">{formatHashRate(miner.hashRate)}</span>
                  </div>
                  <div className="detail-row">
                    <span className="detail-label">能耗</span>
                    <span className="detail-value">{formatEnergy(miner.energyConsumption)}</span>
                  </div>
                  <div className="detail-row">
                    <span className="detail-label">合规状态</span>
                    <span className={`compliance-badge ${getComplianceBadge(miner.complianceStatus)}`}>
                      {miner.complianceStatus === 'COMPLIANT' ? '合规' : 
                       miner.complianceStatus === 'NON_COMPLIANT' ? '不合规' : '审核中'}
                    </span>
                  </div>
                  <div className="detail-row">
                    <span className="detail-label">上次检查</span>
                    <span className="detail-value">
                      {new Date(miner.lastInspection).toLocaleDateString('zh-CN')}
                    </span>
                  </div>
                </div>

                <div className="miner-actions" onClick={(e) => e.stopPropagation()}>
                  <button className="action-btn" title="查看详情">
                    <svg viewBox="0 0 24 24" className="action-icon">
                      <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" fill="none" stroke="currentColor" strokeWidth="2" />
                      <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" strokeWidth="2" />
                    </svg>
                  </button>
                  <button className="action-btn" title="安排检查">
                    <svg viewBox="0 0 24 24" className="action-icon">
                      <rect x="3" y="4" width="18" height="18" rx="2" ry="2" fill="none" stroke="currentColor" strokeWidth="2" />
                      <line x1="16" y1="2" x2="16" y2="6" fill="none" stroke="currentColor" strokeWidth="2" />
                      <line x1="8" y1="2" x2="8" y2="6" fill="none" stroke="currentColor" strokeWidth="2" />
                      <line x1="3" y1="10" x2="21" y2="10" fill="none" stroke="currentColor" strokeWidth="2" />
                    </svg>
                  </button>
                  {miner.status === 'ACTIVE' && (
                    <button 
                      className="action-btn warning" 
                      title="暂停运营"
                      onClick={(e) => handleSuspend(e, miner.id)}
                    >
                      <svg viewBox="0 0 24 24" className="action-icon">
                        <rect x="6" y="4" width="4" height="16" fill="none" stroke="currentColor" strokeWidth="2" />
                        <rect x="14" y="4" width="4" height="16" fill="none" stroke="currentColor" strokeWidth="2" />
                      </svg>
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Add Miner Modal */}
      {showAddModal && (
        <div className="modal-overlay" onClick={() => setShowAddModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">注册新矿工</h3>
              <button className="modal-close" onClick={() => setShowAddModal(false)}>
                <svg viewBox="0 0 24 24" className="close-icon">
                  <path d="M18 6L6 18M6 6l12 12" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
                </svg>
              </button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label className="form-label">矿工名称</label>
                <input type="text" className="form-input" placeholder="输入矿工名称" />
              </div>
              <div className="form-group">
                <label className="form-label">司法管辖区</label>
                <input type="text" className="form-input" placeholder="输入司法管辖区" />
              </div>
              <div className="form-group">
                <label className="form-label">能源来源</label>
                <select className="form-select">
                  <option value="">选择能源来源</option>
                  <option value="hydro">水电</option>
                  <option value="solar">太阳能</option>
                  <option value="wind">风能</option>
                  <option value="geothermal">地热能</option>
                  <option value="nuclear">核能</option>
                  <option value="fossil">化石能源</option>
                </select>
              </div>
              <div className="form-group">
                <label className="form-label">算力 (PH/s)</label>
                <input type="number" className="form-input" placeholder="输入算力" />
              </div>
              <div className="form-group">
                <label className="form-label">能耗 (MW)</label>
                <input type="number" className="form-input" placeholder="输入能耗" />
              </div>
            </div>
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setShowAddModal(false)}>取消</button>
              <button className="btn btn-primary">注册</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default MinerRegistry;
