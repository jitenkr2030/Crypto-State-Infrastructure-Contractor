// CSIC Platform - Wallet Monitor Page
// Wallet monitoring and blacklist management interface

import React, { useState, useEffect } from 'react';
import { useWalletStore, Wallet } from '../../store';

const WalletMonitor: React.FC = () => {
  const { wallets, loadWallets, freezeWallet, unfreezeWallet, blacklistWallet, isLoading, error } = useWalletStore();
  
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [typeFilter, setTypeFilter] = useState<string>('all');
  const [selectedWallet, setSelectedWallet] = useState<Wallet | null>(null);
  const [showFreezeModal, setShowFreezeModal] = useState(false);
  const [freezeReason, setFreezeReason] = useState('');

  useEffect(() => {
    loadWallets();
  }, [loadWallets]);

  const filteredWallets = wallets.filter(wallet => {
    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      if (!wallet.address.toLowerCase().includes(query) &&
          !wallet.label.toLowerCase().includes(query)) {
        return false;
      }
    }

    // Status filter
    if (statusFilter !== 'all' && wallet.status !== statusFilter) {
      return false;
    }

    // Type filter
    if (typeFilter !== 'all' && wallet.type !== typeFilter) {
      return false;
    }

    return true;
  });

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'ACTIVE': return 'badge-success';
      case 'FROZEN': return 'badge-warning';
      case 'BLACKLISTED': return 'badge-error';
      default: return 'badge-info';
    }
  };

  const getTypeBadge = (type: string) => {
    switch (type) {
      case 'CUSTODIAL': return 'type-custodial';
      case 'NON_CUSTODIAL': return 'type-non-custodial';
      case 'EXCHANGE': return 'type-exchange';
      case 'MIXER': return 'type-mixer';
      case 'DARKNET': return 'type-darknet';
      default: return 'type-default';
    }
  };

  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'CUSTODIAL': return '托管';
      case 'NON_CUSTODIAL': return '非托管';
      case 'EXCHANGE': return '交易所';
      case 'MIXER': return '混币';
      case 'DARKNET': return '暗网';
      default: return type;
    }
  };

  const getRiskColor = (score: number) => {
    if (score >= 80) return 'var(--color-error)';
    if (score >= 60) return 'var(--color-warning)';
    return 'var(--color-success)';
  };

  const formatAddress = (address: string) => {
    if (address.length <= 16) return address;
    return `${address.substring(0, 8)}...${address.substring(address.length - 8)}`;
  };

  const handleFreeze = (wallet: Wallet) => {
    setSelectedWallet(wallet);
    setShowFreezeModal(true);
  };

  const confirmFreeze = () => {
    if (selectedWallet && freezeReason) {
      freezeWallet(selectedWallet.id);
      setShowFreezeModal(false);
      setFreezeReason('');
      setSelectedWallet(null);
    }
  };

  const handleBlacklist = (e: React.MouseEvent, walletId: string) => {
    e.stopPropagation();
    if (window.confirm('确定要将该钱包加入黑名单吗？')) {
      blacklistWallet(walletId);
    }
  };

  const stats = {
    total: wallets.length,
    active: wallets.filter(w => w.status === 'ACTIVE').length,
    frozen: wallets.filter(w => w.status === 'FROZEN').length,
    blacklisted: wallets.filter(w => w.status === 'BLACKLISTED').length,
    highRisk: wallets.filter(w => w.riskScore >= 80).length,
  };

  return (
    <div className="wallet-monitor-page">
      <div className="page-header">
        <div className="header-left">
          <h1>钱包监控</h1>
          <p>监控和管理的加密货币钱包列表</p>
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
              <circle cx="11" cy="11" r="8" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M21 21l-4.35-4.35" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            批量查询
          </button>
        </div>
      </div>

      <div className="stats-row">
        <div className="stat-card">
          <div className="stat-icon wallets">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <rect x="2" y="6" width="20" height="12" rx="2" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M12 6v6l4 2" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.total}</span>
            <span className="stat-label">监控钱包</span>
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
            <span className="stat-label">活跃钱包</span>
          </div>
        </div>
        <div className="stat-card warning">
          <div className="stat-icon frozen">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M12 2L2 7l10 5 10-5-10-5z" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M2 17l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M2 12l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.frozen}</span>
            <span className="stat-label">已冻结</span>
          </div>
        </div>
        <div className="stat-card danger">
          <div className="stat-icon blacklisted">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M15 9l-6 6M9 9l6 6" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.blacklisted}</span>
            <span className="stat-label">黑名单</span>
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
                placeholder="搜索钱包地址或标签..."
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
              <option value="FROZEN">已冻结</option>
              <option value="BLACKLISTED">黑名单</option>
            </select>
            <select
              className="filter-select"
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value)}
            >
              <option value="all">所有类型</option>
              <option value="CUSTODIAL">托管</option>
              <option value="NON_CUSTODIAL">非托管</option>
              <option value="EXCHANGE">交易所</option>
              <option value="MIXER">混币</option>
              <option value="DARKNET">暗网</option>
            </select>
          </div>
          <span className="result-count">{filteredWallets.length} 个结果</span>
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
            <p>加载钱包数据...</p>
          </div>
        ) : (
          <div className="wallet-table-container">
            <table className="table wallet-table">
              <thead>
                <tr>
                  <th>地址</th>
                  <th>标签</th>
                  <th>类型</th>
                  <th>风险评分</th>
                  <th>状态</th>
                  <th>首次发现</th>
                  <th>最后活跃</th>
                  <th>关联实体</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {filteredWallets.map(wallet => (
                  <tr 
                    key={wallet.id} 
                    className={`wallet-row ${selectedWallet?.id === wallet.id ? 'selected' : ''}`}
                    onClick={() => setSelectedWallet(wallet)}
                  >
                    <td>
                      <div className="address-cell">
                        <code className="wallet-address" title={wallet.address}>
                          {formatAddress(wallet.address)}
                        </code>
                        <button className="copy-btn" title="复制地址">
                          <svg viewBox="0 0 24 24" className="copy-icon">
                            <rect x="9" y="9" width="13" height="13" rx="2" ry="2" fill="none" stroke="currentColor" strokeWidth="2" />
                            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" fill="none" stroke="currentColor" strokeWidth="2" />
                          </svg>
                        </button>
                      </div>
                    </td>
                    <td>
                      <span className="wallet-label">{wallet.label}</span>
                    </td>
                    <td>
                      <span className={`type-badge ${getTypeBadge(wallet.type)}`}>
                        {getTypeLabel(wallet.type)}
                      </span>
                    </td>
                    <td>
                      <div className="risk-score">
                        <div className="score-bar">
                          <div 
                            className="score-fill"
                            style={{ 
                              width: `${wallet.riskScore}%`,
                              backgroundColor: getRiskColor(wallet.riskScore)
                            }}
                          ></div>
                        </div>
                        <span 
                          className="score-value" 
                          style={{ color: getRiskColor(wallet.riskScore) }}
                        >
                          {wallet.riskScore}
                        </span>
                      </div>
                    </td>
                    <td>
                      <span className={`badge ${getStatusBadge(wallet.status)}`}>
                        {wallet.status === 'ACTIVE' ? '活跃' : 
                         wallet.status === 'FROZEN' ? '已冻结' : '黑名单'}
                      </span>
                    </td>
                    <td>
                      {new Date(wallet.firstSeen).toLocaleDateString('zh-CN')}
                    </td>
                    <td>
                      {new Date(wallet.lastActivity).toLocaleDateString('zh-CN')}
                    </td>
                    <td>
                      <div className="entities-cell">
                        {wallet.associatedEntities.slice(0, 2).map((entity, idx) => (
                          <span key={idx} className="entity-tag">{entity}</span>
                        ))}
                        {wallet.associatedEntities.length > 2 && (
                          <span className="entity-more">+{wallet.associatedEntities.length - 2}</span>
                        )}
                      </div>
                    </td>
                    <td>
                      <div className="action-buttons" onClick={(e) => e.stopPropagation()}>
                        <button 
                          className="action-btn" 
                          title="查看详情"
                          onClick={() => setSelectedWallet(wallet)}
                        >
                          <svg viewBox="0 0 24 24" className="action-icon">
                            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" fill="none" stroke="currentColor" strokeWidth="2" />
                            <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" strokeWidth="2" />
                          </svg>
                        </button>
                        {wallet.status === 'ACTIVE' && (
                          <>
                            <button 
                              className="action-btn warning" 
                              title="冻结钱包"
                              onClick={() => handleFreeze(wallet)}
                            >
                              <svg viewBox="0 0 24 24" className="action-icon">
                                <path d="M12 2L2 7l10 5 10-5-10-5z" fill="none" stroke="currentColor" strokeWidth="2" />
                                <path d="M2 17l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
                                <path d="M2 12l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
                              </svg>
                            </button>
                            <button 
                              className="action-btn danger" 
                              title="加入黑名单"
                              onClick={(e) => handleBlacklist(e, wallet.id)}
                            >
                              <svg viewBox="0 0 24 24" className="action-icon">
                                <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
                                <path d="M15 9l-6 6M9 9l6 6" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
                              </svg>
                            </button>
                          </>
                        )}
                        {wallet.status === 'FROZEN' && (
                          <button 
                            className="action-btn success" 
                            title="解冻钱包"
                            onClick={() => unfreezeWallet(wallet.id)}
                          >
                            <svg viewBox="0 0 24 24" className="action-icon">
                              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" fill="none" stroke="currentColor" strokeWidth="2" />
                              <polyline points="22 4 12 14.01 9 11.01" fill="none" stroke="currentColor" strokeWidth="2" />
                            </svg>
                          </button>
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

      {/* Freeze Modal */}
      {showFreezeModal && selectedWallet && (
        <div className="modal-overlay" onClick={() => setShowFreezeModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">冻结钱包</h3>
              <button className="modal-close" onClick={() => setShowFreezeModal(false)}>
                <svg viewBox="0 0 24 24" className="close-icon">
                  <path d="M18 6L6 18M6 6l12 12" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
                </svg>
              </button>
            </div>
            <div className="modal-body">
              <div className="wallet-info-preview">
                <div className="info-row">
                  <span className="info-label">地址</span>
                  <code className="info-value">{selectedWallet.address}</code>
                </div>
                <div className="info-row">
                  <span className="info-label">标签</span>
                  <span className="info-value">{selectedWallet.label}</span>
                </div>
                <div className="info-row">
                  <span className="info-label">风险评分</span>
                  <span className="info-value" style={{ color: getRiskColor(selectedWallet.riskScore) }}>
                    {selectedWallet.riskScore}
                  </span>
                </div>
              </div>
              <div className="form-group">
                <label className="form-label">冻结原因 *</label>
                <textarea
                  className="form-textarea"
                  placeholder="输入冻结原因..."
                  value={freezeReason}
                  onChange={(e) => setFreezeReason(e.target.value)}
                  rows={3}
                />
              </div>
            </div>
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setShowFreezeModal(false)}>取消</button>
              <button 
                className="btn btn-danger" 
                onClick={confirmFreeze}
                disabled={!freezeReason}
              >
                确认冻结
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Wallet Detail Panel */}
      {selectedWallet && !showFreezeModal && (
        <div className="detail-panel">
          <div className="detail-header">
            <h3>钱包详情</h3>
            <button className="close-btn" onClick={() => setSelectedWallet(null)}>
              <svg viewBox="0 0 24 24" className="close-icon">
                <path d="M18 6L6 18M6 6l12 12" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
              </svg>
            </button>
          </div>
          <div className="detail-content">
            <div className="detail-section">
              <h4>基本信息</h4>
              <div className="detail-grid">
                <div className="detail-item">
                  <span className="item-label">地址</span>
                  <code className="item-value">{selectedWallet.address}</code>
                </div>
                <div className="detail-item">
                  <span className="item-label">标签</span>
                  <span className="item-value">{selectedWallet.label}</span>
                </div>
                <div className="detail-item">
                  <span className="item-label">类型</span>
                  <span className={`type-badge ${getTypeBadge(selectedWallet.type)}`}>
                    {getTypeLabel(selectedWallet.type)}
                  </span>
                </div>
                <div className="detail-item">
                  <span className="item-label">状态</span>
                  <span className={`badge ${getStatusBadge(selectedWallet.status)}`}>
                    {selectedWallet.status}
                  </span>
                </div>
              </div>
            </div>
            <div className="detail-section">
              <h4>风险评估</h4>
              <div className="risk-assessment">
                <div className="risk-meter">
                  <div className="risk-label">风险评分</div>
                  <div className="risk-bar">
                    <div 
                      className="risk-fill"
                      style={{ 
                        width: `${selectedWallet.riskScore}%`,
                        backgroundColor: getRiskColor(selectedWallet.riskScore)
                      }}
                    ></div>
                  </div>
                  <div className="risk-value" style={{ color: getRiskColor(selectedWallet.riskScore) }}>
                    {selectedWallet.riskScore}/100
                  </div>
                </div>
              </div>
            </div>
            <div className="detail-section">
              <h4>关联实体</h4>
              <div className="entities-list">
                {selectedWallet.associatedEntities.map((entity, idx) => (
                  <div key={idx} className="entity-item">
                    <span className="entity-name">{entity}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default WalletMonitor;
