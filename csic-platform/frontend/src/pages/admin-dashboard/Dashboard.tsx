import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { useSystemStore, useAlertStore } from '../../store';

interface DashboardData {
  systemStatus: {
    status: 'ONLINE' | 'DEGRADED' | 'OFFLINE' | 'EMERGENCY';
    uptime: number;
    lastHeartbeat: Date;
  };
  metrics: {
    exchanges: {
      total: number;
      active: number;
      suspended: number;
    };
    transactions: {
      total24h: number;
      volume24h: number;
      flagged: number;
    };
    wallets: {
      total: number;
      frozen: number;
    };
    miners: {
      total: number;
      online: number;
      totalHashRate: number;
    };
  };
  charts: {
    transactionVolume: { time: string; volume: number }[];
    energyConsumption: { time: string; consumption: number }[];
    alertsTrend: { time: string; critical: number; warning: number; info: number }[];
  };
}

const Dashboard: React.FC = () => {
  const { status, setSystemStatus } = useSystemStore();
  const { alerts, unreadCount } = useAlertStore();
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadDashboardData();
    const interval = setInterval(loadDashboardData, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const loadDashboardData = async () => {
    try {
      // Simulate API call - replace with actual API call
      const mockData: DashboardData = {
        systemStatus: {
          status: 'ONLINE',
          uptime: 86400 * 15, // 15 days
          lastHeartbeat: new Date(),
        },
        metrics: {
          exchanges: {
            total: 24,
            active: 20,
            suspended: 2,
          },
          transactions: {
            total24h: 15420,
            volume24h: 2.5e9,
            flagged: 12,
          },
          wallets: {
            total: 1560,
            frozen: 23,
          },
          miners: {
            total: 156,
            online: 142,
            totalHashRate: 450,
          },
        },
        charts: {
          transactionVolume: generateVolumeData(),
          energyConsumption: generateEnergyData(),
          alertsTrend: generateAlertsData(),
        },
      };
      setData(mockData);
      setLoading(false);
    } catch (error) {
      console.error('Failed to load dashboard data:', error);
      setLoading(false);
    }
  };

  const getStatusColor = () => {
    switch (status) {
      case 'EMERGENCY': return 'status-critical';
      case 'DEGRADED': return 'status-warning';
      default: return 'status-normal';
    }
  };

  const formatNumber = (num: number) => {
    if (num >= 1e9) return (num / 1e9).toFixed(2) + 'B';
    if (num >= 1e6) return (num / 1e6).toFixed(2) + 'M';
    if (num >= 1e3) return (num / 1e3).toFixed(2) + 'K';
    return num.toString();
  };

  const formatDuration = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${days}d ${hours}h ${minutes}m`;
  };

  if (loading) {
    return (
      <div className="dashboard-loading">
        <div className="loading-spinner large"></div>
        <p>加载仪表板数据...</p>
      </div>
    );
  }

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <h1>系统控制台</h1>
        <div className={`system-status-badge ${getStatusColor()}`}>
          <span className="status-dot"></span>
          <span>{status === 'ONLINE' ? '系统正常运行' : status}</span>
        </div>
      </div>

      {/* Quick Stats Row */}
      <div className="stats-row">
        <div className="stat-card">
          <div className="stat-icon exchanges">
            <ExchangeIcon />
          </div>
          <div className="stat-content">
            <span className="stat-label">活跃交易所</span>
            <span className="stat-value">{data?.metrics.exchanges.active || 0}</span>
            <span className="stat-detail">/ {data?.metrics.exchanges.total || 0} 总计</span>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon transactions">
            <TransactionIcon />
          </div>
          <div className="stat-content">
            <span className="stat-label">24小时交易</span>
            <span className="stat-value">{formatNumber(data?.metrics.transactions.total24h || 0)}</span>
            <span className="stat-detail">${formatNumber(data?.metrics.transactions.volume24h || 0)} 交易量</span>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon wallets">
            <WalletIcon />
          </div>
          <div className="stat-content">
            <span className="stat-label">监控钱包</span>
            <span className="stat-value">{formatNumber(data?.metrics.wallets.total || 0)}</span>
            <span className="stat-detail">{data?.metrics.wallets.frozen || 0} 已冻结</span>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon miners">
            <MiningIcon />
          </div>
          <div className="stat-content">
            <span className="stat-label">在线矿工</span>
            <span className="stat-value">{data?.metrics.miners.online || 0}</span>
            <span className="stat-detail">{data?.metrics.miners.totalHashRate} PH/s 算力</span>
          </div>
        </div>

        <div className="stat-card alert-card">
          <div className="stat-icon alerts">
            <AlertIcon />
          </div>
          <div className="stat-content">
            <span className="stat-label">待处理警报</span>
            <span className="stat-value">{unreadCount}</span>
            <span className="stat-detail">{alerts.filter(a => a.status === 'ACTIVE').length} 活动警报</span>
          </div>
        </div>
      </div>

      {/* Charts Row */}
      <div className="charts-row">
        <div className="chart-card large">
          <div className="chart-header">
            <h3>交易量趋势</h3>
            <div className="chart-actions">
              <button className="chart-btn active">24小时</button>
              <button className="chart-btn">7天</button>
              <button className="chart-btn">30天</button>
            </div>
          </div>
          <div className="chart-body">
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={data?.charts.transactionVolume || []}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="time" stroke="#94a3b8" />
                <YAxis stroke="#94a3b8" tickFormatter={formatNumber} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155' }}
                  formatter={(value: number) => [formatNumber(value), '交易量']}
                />
                <Legend />
                <Line type="monotone" dataKey="volume" stroke="#3b82f6" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="chart-card">
          <div className="chart-header">
            <h3>能源消耗</h3>
          </div>
          <div className="chart-body">
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={data?.charts.energyConsumption || []}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="time" stroke="#94a3b8" />
                <YAxis stroke="#94a3b8" />
                <Tooltip
                  contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155' }}
                  formatter={(value: number) => [value.toFixed(1), 'MW']}
                />
                <Line type="monotone" dataKey="consumption" stroke="#10b981" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* Alerts Trend */}
      <div className="alerts-section">
        <div className="section-header">
          <h3>警报趋势</h3>
          <button className="view-all-btn">查看全部</button>
        </div>
        <div className="alerts-chart">
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={data?.charts.alertsTrend || []}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="time" stroke="#94a3b8" />
              <YAxis stroke="#94a3b8" />
              <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155' }} />
              <Legend />
              <Line type="monotone" dataKey="critical" stroke="#ef4444" strokeWidth={2} />
              <Line type="monotone" dataKey="warning" stroke="#f59e0b" strokeWidth={2} />
              <Line type="monotone" dataKey="info" stroke="#3b82f6" strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Recent Alerts */}
      <div className="recent-alerts-section">
        <div className="section-header">
          <h3>最新警报</h3>
          <button className="view-all-btn">查看全部</button>
        </div>
        <div className="alerts-list">
          {alerts.slice(0, 5).map((alert) => (
            <div key={alert.id} className={`alert-item severity-${alert.severity.toLowerCase()}`}>
              <div className="alert-severity">
                <span className={`severity-badge ${alert.severity.toLowerCase()}`}>
                  {alert.severity}
                </span>
              </div>
              <div className="alert-content">
                <div className="alert-title">{alert.title}</div>
                <div className="alert-meta">
                  <span className="alert-category">{alert.category}</span>
                  <span className="alert-time">
                    {new Date(alert.createdAt).toLocaleString('zh-CN')}
                  </span>
                </div>
              </div>
              <div className="alert-actions">
                <button className="action-btn">查看</button>
                <button className="action-btn">确认</button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* System Info Footer */}
      <div className="dashboard-footer">
        <div className="system-info">
          <div className="info-item">
            <span className="info-label">系统运行时间</span>
            <span className="info-value">{formatDuration(data?.systemStatus.uptime || 0)}</span>
          </div>
          <div className="info-item">
            <span className="info-label">最后心跳</span>
            <span className="info-value">
              {data?.systemStatus.lastHeartbeat.toLocaleString('zh-CN')}
            </span>
          </div>
          <div className="info-item">
            <span className="info-label">版本</span>
            <span className="info-value">v1.0.0</span>
          </div>
          <div className="info-item">
            <span className="info-label">HSM状态</span>
            <span className="info-value status-online">已连接</span>
          </div>
        </div>
      </div>
    </div>
  );
};

// Helper functions to generate mock chart data
function generateVolumeData() {
  const data = [];
  for (let i = 23; i >= 0; i--) {
    const time = new Date();
    time.setHours(time.getHours() - i);
    data.push({
      time: time.getHours().toString().padStart(2, '0') + ':00',
      volume: Math.floor(Math.random() * 50000000) + 10000000,
    });
  }
  return data;
}

function generateEnergyData() {
  const data = [];
  for (let i = 23; i >= 0; i--) {
    const time = new Date();
    time.setHours(time.getHours() - i);
    data.push({
      time: time.getHours().toString().padStart(2, '0') + ':00',
      consumption: Math.floor(Math.random() * 100) + 400,
    });
  }
  return data;
}

function generateAlertsData() {
  const data = [];
  for (let i = 6; i >= 0; i--) {
    const date = new Date();
    date.setDate(date.getDate() - i);
    data.push({
      time: date.toLocaleDateString('zh-CN'),
      critical: Math.floor(Math.random() * 3),
      warning: Math.floor(Math.random() * 10),
      info: Math.floor(Math.random() * 15),
    });
  }
  return data;
}

// Icon components
const ExchangeIcon: React.FC = () => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M12 2L2 7l10 5 10-5-10-5z" />
    <path d="M2 17l10 5 10-5" />
    <path d="M2 12l10 5 10-5" />
  </svg>
);

const TransactionIcon: React.FC = () => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
  </svg>
);

const WalletIcon: React.FC = () => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M21 12V7a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h7" />
    <rect x="2" y="6" width="20" height="12" rx="2" />
  </svg>
);

const MiningIcon: React.FC = () => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M12 2L2 7l10 5 10-5-10-5z" />
    <path d="M2 17l10 5 10-5" />
    <path d="M2 12l10 5 10-5" />
  </svg>
);

const AlertIcon: React.FC = () => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
    <path d="M13.73 21a2 2 0 0 1-3.46 0" />
  </svg>
);

export default Dashboard;
