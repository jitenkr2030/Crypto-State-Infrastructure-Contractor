import { motion } from 'framer-motion';
import {
  Activity,
  Server,
  Shield,
  AlertTriangle,
  TrendingUp,
  TrendingDown,
  Zap,
  Globe,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
} from 'lucide-react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  BarChart,
  Bar,
  LineChart,
  Line,
} from 'recharts';

// Mock data for dashboard
const stats = [
  {
    name: 'Active Nodes',
    value: '24',
    change: '+2',
    changeType: 'positive',
    icon: Server,
    color: 'primary',
  },
  {
    name: 'Compliance Rate',
    value: '98.5%',
    change: '+0.5%',
    changeType: 'positive',
    icon: Shield,
    color: 'success',
  },
  {
    name: 'Open Violations',
    value: '12',
    change: '-3',
    changeType: 'positive',
    icon: AlertTriangle,
    color: 'warning',
  },
  {
    name: 'System Health',
    value: '99.9%',
    change: '+0.1%',
    changeType: 'positive',
    icon: Activity,
    color: 'info',
  },
];

const transactionData = [
  { time: '00:00', transactions: 145, compliance: 98.2 },
  { time: '04:00', transactions: 89, compliance: 97.8 },
  { time: '08:00', transactions: 234, compliance: 98.5 },
  { time: '12:00', transactions: 456, compliance: 99.1 },
  { time: '16:00', transactions: 521, compliance: 98.7 },
  { time: '20:00', transactions: 389, compliance: 98.3 },
  { time: '23:59', transactions: 267, compliance: 98.9 },
];

const complianceDistribution = [
  { name: 'Passed', value: 985, color: '#10b981' },
  { name: 'Warning', value: 12, color: '#f59e0b' },
  { name: 'Failed', value: 3, color: '#ef4444' },
];

const nodeStatusData = [
  { name: 'Running', value: 20, color: '#10b981' },
  { name: 'Syncing', value: 3, color: '#f59e0b' },
  { name: 'Error', value: 1, color: '#ef4444' },
];

const recentViolations = [
  {
    id: 1,
    rule: 'Maximum Amount Check',
    severity: 'HIGH',
    status: 'OPEN',
    timestamp: '5 minutes ago',
    transaction: 'TX-2024-001234',
  },
  {
    id: 2,
    rule: 'Sanctioned Country Check',
    severity: 'CRITICAL',
    status: 'OPEN',
    timestamp: '15 minutes ago',
    transaction: 'TX-2024-001235',
  },
  {
    id: 3,
    rule: 'Transaction Frequency Check',
    severity: 'MEDIUM',
    status: 'RESOLVED',
    timestamp: '1 hour ago',
    transaction: 'TX-2024-001230',
  },
  {
    id: 4,
    rule: 'Source KYC Verification',
    severity: 'HIGH',
    status: 'OPEN',
    timestamp: '2 hours ago',
    transaction: 'TX-2024-001228',
  },
];

const systemHealth = [
  { component: 'API Gateway', status: 'healthy', uptime: '99.99%' },
  { component: 'Audit Service', status: 'healthy', uptime: '99.95%' },
  { component: 'Compliance Service', status: 'healthy', uptime: '99.92%' },
  { component: 'Node Manager', status: 'degraded', uptime: '99.50%' },
  { component: 'Health Monitor', status: 'healthy', uptime: '99.98%' },
];

const containerVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};

export default function Dashboard() {
  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="page-header">Dashboard</h1>
        <p className="page-description">
          Real-time overview of your CSIC Platform status and metrics
        </p>
      </div>

      {/* Stats Grid */}
      <motion.div
        variants={containerVariants}
        initial="hidden"
        animate="visible"
        className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6"
      >
        {stats.map((stat, index) => (
          <motion.div
            key={stat.name}
            variants={{
              hidden: { opacity: 0, y: 20 },
              visible: { opacity: 1, y: 0 },
            }}
            transition={{ delay: index * 0.1 }}
            className="stat-card hover-lift"
          >
            <div className="flex items-start justify-between">
              <div>
                <p className="stat-label">{stat.name}</p>
                <p className="stat-value mt-2">{stat.value}</p>
                <div className="flex items-center gap-1 mt-2">
                  {stat.changeType === 'positive' ? (
                    <TrendingUp className="w-4 h-4 text-success" />
                  ) : (
                    <TrendingDown className="w-4 h-4 text-danger" />
                  )}
                  <span
                    className={`text-sm ${
                      stat.changeType === 'positive' ? 'stat-change-positive' : 'stat-change-negative'
                    }`}
                  >
                    {stat.change}
                  </span>
                  <span className="text-xs text-gray-500">vs last 24h</span>
                </div>
              </div>
              <div
                className={`p-3 rounded-xl bg-${stat.color}/10`}
              >
                <stat.icon className={`w-6 h-6 text-${stat.color}`} />
              </div>
            </div>
          </motion.div>
        ))}
      </motion.div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Transaction Volume Chart */}
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="card"
        >
          <div className="card-header">
            <h3 className="section-title mb-0">Transaction Volume</h3>
          </div>
          <div className="card-body">
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={transactionData}>
                <defs>
                  <linearGradient id="colorTransactions" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="time" stroke="#64748b" fontSize={12} />
                <YAxis stroke="#64748b" fontSize={12} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#1e293b',
                    border: '1px solid #334155',
                    borderRadius: '8px',
                  }}
                  labelStyle={{ color: '#fff' }}
                />
                <Area
                  type="monotone"
                  dataKey="transactions"
                  stroke="#3b82f6"
                  strokeWidth={2}
                  fillOpacity={1}
                  fill="url(#colorTransactions)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </motion.div>

        {/* Compliance Distribution */}
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="card"
        >
          <div className="card-header">
            <h3 className="section-title mb-0">Compliance Distribution</h3>
          </div>
          <div className="card-body">
            <div className="flex items-center justify-center">
              <ResponsiveContainer width="50%" height={300}>
                <PieChart>
                  <Pie
                    data={complianceDistribution}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={100}
                    paddingAngle={5}
                    dataKey="value"
                  >
                    {complianceDistribution.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#1e293b',
                      border: '1px solid #334155',
                      borderRadius: '8px',
                    }}
                  />
                </PieChart>
              </ResponsiveContainer>
              <div className="space-y-4">
                {complianceDistribution.map((item) => (
                  <div key={item.name} className="flex items-center gap-3">
                    <div
                      className="w-3 h-3 rounded-full"
                      style={{ backgroundColor: item.color }}
                    />
                    <span className="text-gray-400">{item.name}</span>
                    <span className="font-semibold text-white">{item.value}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </motion.div>
      </div>

      {/* Bottom Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Recent Violations */}
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="lg:col-span-2 card"
        >
          <div className="card-header flex items-center justify-between">
            <h3 className="section-title mb-0">Recent Violations</h3>
            <button className="btn-ghost btn-sm">View all</button>
          </div>
          <div className="overflow-x-auto">
            <table className="table">
              <thead>
                <tr>
                  <th>Rule</th>
                  <th>Severity</th>
                  <th>Status</th>
                  <th>Transaction</th>
                  <th>Time</th>
                </tr>
              </thead>
              <tbody>
                {recentViolations.map((violation) => (
                  <tr key={violation.id}>
                    <td className="font-medium">{violation.rule}</td>
                    <td>
                      <span
                        className={`badge ${
                          violation.severity === 'CRITICAL'
                            ? 'badge-danger'
                            : violation.severity === 'HIGH'
                            ? 'badge-warning'
                            : 'badge-gray'
                        }`}
                      >
                        {violation.severity}
                      </span>
                    </td>
                    <td>
                      <span
                        className={`badge ${
                          violation.status === 'OPEN' ? 'badge-danger' : 'badge-success'
                        }`}
                      >
                        {violation.status}
                      </span>
                    </td>
                    <td className="font-mono text-sm">{violation.transaction}</td>
                    <td className="text-gray-400">{violation.timestamp}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </motion.div>

        {/* System Health */}
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="card"
        >
          <div className="card-header">
            <h3 className="section-title mb-0">System Health</h3>
          </div>
          <div className="card-body space-y-4">
            {systemHealth.map((system) => (
              <div key={system.component} className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  {system.status === 'healthy' ? (
                    <CheckCircle className="w-5 h-5 text-success" />
                  ) : system.status === 'degraded' ? (
                    <AlertCircle className="w-5 h-5 text-warning" />
                  ) : (
                    <XCircle className="w-5 h-5 text-danger" />
                  )}
                  <div>
                    <p className="text-sm font-medium text-white">{system.component}</p>
                    <p className="text-xs text-gray-400">Uptime: {system.uptime}</p>
                  </div>
                </div>
                <span
                  className={`badge ${
                    system.status === 'healthy'
                      ? 'badge-success'
                      : system.status === 'degraded'
                      ? 'badge-warning'
                      : 'badge-danger'
                  }`}
                >
                  {system.status}
                </span>
              </div>
            ))}
          </div>
        </motion.div>
      </div>
    </div>
  );
}
