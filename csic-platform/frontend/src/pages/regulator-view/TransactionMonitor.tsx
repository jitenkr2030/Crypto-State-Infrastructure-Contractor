// CSIC Platform - Transaction Monitor Page
// Real-time transaction monitoring and management interface

import React, { useState, useEffect } from 'react';

interface Transaction {
  id: string;
  txHash: string;
  type: string;
  amount: number;
  currency: string;
  fromAddress: string;
  toAddress: string;
  status: 'PENDING' | 'CONFIRMED' | 'FLAGGED' | 'BLOCKED';
  riskScore: number;
  timestamp: Date;
  exchangeId?: string;
}

const TransactionMonitor: React.FC = () => {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [currencyFilter, setCurrencyFilter] = useState('all');
  const [selectedTx, setSelectedTx] = useState<Transaction | null>(null);

  useEffect(() => {
    loadTransactions();
    
    // Simulate real-time updates
    const interval = setInterval(() => {
      addNewTransaction();
    }, 10000);
    
    return () => clearInterval(interval);
  }, []);

  const loadTransactions = async () => {
    setIsLoading(true);
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    const mockTxs: Transaction[] = [
      {
        id: 'tx_001',
        txHash: '0x1a2b3c4d5e6f7890abcdef1234567890abcdef1234567890abcdef1234567890',
        type: 'TRANSFER',
        amount: 150.5,
        currency: 'ETH',
        fromAddress: '0x742d35Cc6634C0532925a3b844Bc9e7595f7f679',
        toAddress: '0x8626f6940E2eb28930eFb4CeF49B2d1F2c9C1199',
        status: 'CONFIRMED',
        riskScore: 15,
        timestamp: new Date(Date.now() - 3600000),
        exchangeId: 'ex_001',
      },
      {
        id: 'tx_002',
        txHash: '2b3c4d5e6f7890abcdef1234567890abcdef1234567890abcdef1234567890ab',
        type: 'EXCHANGE',
        amount: 2500.0,
        currency: 'BTC',
        fromAddress: '1A2B3C4D5E6F7890ABCDEF1234567890',
        toAddress: '3J98t1WpEZ73CNmQviecrnyiWrnqRhWNLy',
        status: 'FLAGGED',
        riskScore: 85,
        timestamp: new Date(Date.now() - 7200000),
        exchangeId: 'ex_002',
      },
      {
        id: 'tx_003',
        txHash: '3c4d5e6f7890abcdef1234567890abcdef1234567890abcdef1234567890abcd',
        type: 'WALLET',
        amount: 50.0,
        currency: 'ETH',
        fromAddress: '0xdD2FD4581271e230360230F9337D5c0430Bf44C0',
        toAddress: '0x2546BcD3c84621e976D8185a91A922aE77ECEc30',
        status: 'PENDING',
        riskScore: 45,
        timestamp: new Date(Date.now() - 1800000),
      },
      {
        id: 'tx_004',
        txHash: '4d5e6f7890abcdef1234567890abcdef1234567890abcdef1234567890abcde',
        type: 'CONTRACT',
        amount: 100.0,
        currency: 'ETH',
        fromAddress: '0x742d35Cc6634C0532925a3b844Bc9e7595f7f679',
        toAddress: '0x4Ed2Ae858fA3F7D8EeA99b9f3F7C7A8b2cD4e6f8',
        status: 'CONFIRMED',
        riskScore: 25,
        timestamp: new Date(Date.now() - 14400000),
        exchangeId: 'ex_001',
      },
      {
        id: 'tx_005',
        txHash: '5e6f7890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12',
        type: 'TRANSFER',
        amount: 50000.0,
        currency: 'USDT',
        fromAddress: '0x1234567890abcdef1234567890abcdef12345678',
        toAddress: '0xabcdef1234567890abcdef1234567890abcdef12',
        status: 'BLOCKED',
        riskScore: 95,
        timestamp: new Date(Date.now() - 10800000),
      },
    ];
    
    setTransactions(mockTxs);
    setIsLoading(false);
  };

  const addNewTransaction = () => {
    const currencies = ['ETH', 'BTC', 'USDT'];
    const types = ['TRANSFER', 'EXCHANGE', 'WALLET', 'CONTRACT'];
    const statuses = ['PENDING', 'CONFIRMED', 'FLAGGED'];
    
    const newTx: Transaction = {
      id: `tx_${Date.now()}`,
      txHash: generateRandomHash(),
      type: types[Math.floor(Math.random() * types.length)],
      amount: Math.random() * 1000 + 10,
      currency: currencies[Math.floor(Math.random() * currencies.length)],
      fromAddress: generateRandomAddress(),
      toAddress: generateRandomAddress(),
      status: statuses[Math.floor(Math.random() * statuses.length)],
      riskScore: Math.floor(Math.random() * 100),
      timestamp: new Date(),
    };
    
    setTransactions(prev => [newTx, ...prev.slice(0, 49)]);
  };

  const generateRandomHash = () => {
    const chars = '0123456789abcdef';
    let hash = '';
    for (let i = 0; i < 64; i++) {
      hash += chars[Math.floor(Math.random() * chars.length)];
    }
    return '0x' + hash;
  };

  const generateRandomAddress = () => {
    const chars = '0123456789abcdef';
    let address = '0x';
    for (let i = 0; i < 40; i++) {
      address += chars[Math.floor(Math.random() * chars.length)];
    }
    return address;
  };

  const filteredTransactions = transactions.filter(tx => {
    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      if (!tx.txHash.toLowerCase().includes(query) &&
          !tx.fromAddress.toLowerCase().includes(query) &&
          !tx.toAddress.toLowerCase().includes(query)) {
        return false;
      }
    }

    // Status filter
    if (statusFilter !== 'all' && tx.status !== statusFilter) {
      return false;
    }

    // Currency filter
    if (currencyFilter !== 'all' && tx.currency !== currencyFilter) {
      return false;
    }

    return true;
  });

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'CONFIRMED': return 'badge-success';
      case 'PENDING': return 'badge-warning';
      case 'FLAGGED': return 'badge-error';
      case 'BLOCKED': return 'badge-error';
      default: return 'badge-info';
    }
  };

  const getTypeBadge = (type: string) => {
    switch (type) {
      case 'TRANSFER': return 'type-transfer';
      case 'EXCHANGE': return 'type-exchange';
      case 'WALLET': return 'type-wallet';
      case 'CONTRACT': return 'type-contract';
      default: return 'type-default';
    }
  };

  const getRiskColor = (score: number) => {
    if (score >= 80) return 'var(--color-error)';
    if (score >= 60) return 'var(--color-warning)';
    return 'var(--color-success)';
  };

  const formatAddress = (address: string) => {
    if (address.length <= 16) return address;
    return `${address.substring(0, 8)}...${address.substring(address.length - 6)}`;
  };

  const formatHash = (hash: string) => {
    return `${hash.substring(0, 10)}...${hash.substring(hash.length - 8)}`;
  };

  const stats = {
    total24h: transactions.length,
    volume24h: transactions.reduce((sum, tx) => sum + tx.amount, 0),
    flagged: transactions.filter(tx => tx.status === 'FLAGGED').length,
    blocked: transactions.filter(tx => tx.status === 'BLOCKED').length,
    avgRisk: Math.round(transactions.reduce((sum, tx) => sum + tx.riskScore, 0) / Math.max(transactions.length, 1)),
  };

  const formatVolume = (volume: number) => {
    if (volume >= 1000000) return `$${(volume / 1000000).toFixed(2)}M`;
    if (volume >= 1000) return `$${(volume / 1000).toFixed(2)}K`;
    return `$${volume.toFixed(2)}`;
  };

  return (
    <div className="transaction-monitor-page">
      <div className="page-header">
        <div className="header-left">
          <h1>交易监控</h1>
          <p>实时监控和分析区块链交易</p>
        </div>
        <div className="header-actions">
          <div className="live-indicator">
            <span className="pulse-dot"></span>
            实时监控中
          </div>
          <button className="btn btn-secondary">
            <svg viewBox="0 0 24 24" className="btn-icon">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="7 10 12 15 17 10" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="15" x2="12" y2="3" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            导出
          </button>
          <button className="btn btn-primary">
            <svg viewBox="0 0 24 24" className="btn-icon">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            设置阈值
          </button>
        </div>
      </div>

      <div className="stats-row">
        <div className="stat-card">
          <div className="stat-icon transactions">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.total24h}</span>
            <span className="stat-label">24小时交易</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon volume">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <line x1="12" y1="1" x2="12" y2="23" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{formatVolume(stats.volume24h)}</span>
            <span className="stat-label">24小时交易量</span>
          </div>
        </div>
        <div className="stat-card warning">
          <div className="stat-icon flagged">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.flagged}</span>
            <span className="stat-label">已标记交易</span>
          </div>
        </div>
        <div className="stat-card danger">
          <div className="stat-icon blocked">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M15 9l-6 6M9 9l6 6" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.blocked}</span>
            <span className="stat-label">已阻止交易</span>
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
                placeholder="搜索交易哈希或地址..."
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
              <option value="CONFIRMED">已确认</option>
              <option value="PENDING">待确认</option>
              <option value="FLAGGED">已标记</option>
              <option value="BLOCKED">已阻止</option>
            </select>
            <select
              className="filter-select"
              value={currencyFilter}
              onChange={(e) => setCurrencyFilter(e.target.value)}
            >
              <option value="all">所有币种</option>
              <option value="BTC">BTC</option>
              <option value="ETH">ETH</option>
              <option value="USDT">USDT</option>
            </select>
          </div>
          <span className="result-count">{filteredTransactions.length} 个结果</span>
        </div>

        {isLoading ? (
          <div className="loading-state">
            <div className="loading-spinner large"></div>
            <p>加载交易数据...</p>
          </div>
        ) : (
          <div className="transactions-table-container">
            <table className="table transactions-table">
              <thead>
                <tr>
                  <th>交易哈希</th>
                  <th>类型</th>
                  <th>金额</th>
                  <th>发送方</th>
                  <th>接收方</th>
                  <th>风险评分</th>
                  <th>状态</th>
                  <th>时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {filteredTransactions.map(tx => (
                  <tr 
                    key={tx.id} 
                    className={`transaction-row ${selectedTx?.id === tx.id ? 'selected' : ''} ${tx.status === 'BLOCKED' ? 'row-blocked' : ''} ${tx.status === 'FLAGGED' ? 'row-flagged' : ''}`}
                    onClick={() => setSelectedTx(tx)}
                  >
                    <td>
                      <div className="hash-cell">
                        <code className="tx-hash" title={tx.txHash}>
                          {formatHash(tx.txHash)}
                        </code>
                        <span className={`currency-badge ${tx.currency.toLowerCase()}`}>
                          {tx.currency}
                        </span>
                      </div>
                    </td>
                    <td>
                      <span className={`type-badge ${getTypeBadge(tx.type)}`}>
                        {tx.type}
                      </span>
                    </td>
                    <td>
                      <span className="amount">
                        {tx.amount.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 4 })}
                      </span>
                    </td>
                    <td>
                      <code className="address" title={tx.fromAddress}>
                        {formatAddress(tx.fromAddress)}
                      </code>
                    </td>
                    <td>
                      <code className="address" title={tx.toAddress}>
                        {formatAddress(tx.toAddress)}
                      </code>
                    </td>
                    <td>
                      <div className="risk-score">
                        <div className="score-bar">
                          <div 
                            className="score-fill"
                            style={{ 
                              width: `${tx.riskScore}%`,
                              backgroundColor: getRiskColor(tx.riskScore)
                            }}
                          ></div>
                        </div>
                        <span 
                          className="score-value"
                          style={{ color: getRiskColor(tx.riskScore) }}
                        >
                          {tx.riskScore}
                        </span>
                      </div>
                    </td>
                    <td>
                      <span className={`badge ${getStatusBadge(tx.status)}`}>
                        {tx.status === 'CONFIRMED' ? '已确认' : 
                         tx.status === 'PENDING' ? '待确认' :
                         tx.status === 'FLAGGED' ? '已标记' : '已阻止'}
                      </span>
                    </td>
                    <td>
                      <span className="timestamp">
                        {tx.timestamp.toLocaleTimeString('zh-CN')}
                      </span>
                    </td>
                    <td>
                      <div className="action-buttons" onClick={(e) => e.stopPropagation()}>
                        <button className="action-btn" title="查看详情">
                          <svg viewBox="0 0 24 24" className="action-icon">
                            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" fill="none" stroke="currentColor" strokeWidth="2" />
                            <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" strokeWidth="2" />
                          </svg>
                        </button>
                        {tx.status === 'PENDING' && (
                          <>
                            <button className="action-btn warning" title="标记交易">
                              <svg viewBox="0 0 24 24" className="action-icon">
                                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" fill="none" stroke="currentColor" strokeWidth="2" />
                              </svg>
                            </button>
                            <button className="action-btn danger" title="阻止交易">
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

      {/* Transaction Detail Panel */}
      {selectedTx && (
        <div className="detail-panel large">
          <div className="detail-header">
            <h3>交易详情</h3>
            <button className="close-btn" onClick={() => setSelectedTx(null)}>
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
                  <span className="item-label">交易哈希</span>
                  <code className="item-value">{selectedTx.txHash}</code>
                </div>
                <div className="detail-item">
                  <span className="item-label">类型</span>
                  <span className={`type-badge ${getTypeBadge(selectedTx.type)}`}>
                    {selectedTx.type}
                  </span>
                </div>
                <div className="detail-item">
                  <span className="item-label">金额</span>
                  <span className="item-value">
                    {selectedTx.amount.toLocaleString()} {selectedTx.currency}
                  </span>
                </div>
                <div className="detail-item">
                  <span className="item-label">状态</span>
                  <span className={`badge ${getStatusBadge(selectedTx.status)}`}>
                    {selectedTx.status}
                  </span>
                </div>
              </div>
            </div>
            <div className="detail-section">
              <h4>交易详情</h4>
              <div className="address-info">
                <div className="address-row">
                  <span className="address-label">发送方</span>
                  <code className="address-value">{selectedTx.fromAddress}</code>
                </div>
                <div className="arrow">
                  <svg viewBox="0 0 24 24">
                    <path d="M5 12h14M12 5l7 7-7 7" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                  </svg>
                </div>
                <div className="address-row">
                  <span className="address-label">接收方</span>
                  <code className="address-value">{selectedTx.toAddress}</code>
                </div>
              </div>
            </div>
            <div className="detail-section">
              <h4>风险评估</h4>
              <div className="risk-assessment">
                <div className="risk-meter">
                  <div className="risk-label">风险评分</div>
                  <div className="risk-bar large">
                    <div 
                      className="risk-fill"
                      style={{ 
                        width: `${selectedTx.riskScore}%`,
                        backgroundColor: getRiskColor(selectedTx.riskScore)
                      }}
                    ></div>
                  </div>
                  <div className="risk-value large" style={{ color: getRiskColor(selectedTx.riskScore) }}>
                    {selectedTx.riskScore}/100
                  </div>
                </div>
              </div>
            </div>
            <div className="detail-actions">
              <button className="btn btn-secondary">查看区块链</button>
              <button className="btn btn-secondary">标记为可疑</button>
              <button className="btn btn-danger">阻止交易</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default TransactionMonitor;
