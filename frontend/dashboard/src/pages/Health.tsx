import { motion } from 'framer-motion';
import {
  Activity,
  Server,
  Database,
  Globe,
  Cpu,
  HardDrive,
  Wifi,
  Zap,
  Clock,
  CheckCircle,
  AlertTriangle,
  XCircle,
  TrendingUp,
} from 'lucide-react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  AreaChart,
  Area,
} from 'recharts';

const services = [
  {
    name: 'API Gateway',
    status: 'healthy',
    uptime: '99.99%',
    latency: '45ms',
    requests: '12.5K/min',
    errors: '0.01%',
  },
  {
    name: 'Audit Service',
    status: 'healthy',
    uptime: '99.95%',
    latency: '32ms',
    requests: '8.2K/min',
    errors: '0.00%',
  },
  {
    name: 'Compliance Service',
    status: 'healthy',
    uptime: '99.92%',
    latency: '156ms',
    requests: '5.1K/min',
    errors: '0.05%',
  },
  {
    name: 'Control Layer',
    status: 'degraded',
    uptime: '99.50%',
    latency: '89ms',
    requests: '3.8K/min',
    errors: '0.15%',
  },
  {
    name: 'Node Manager',
    status: 'healthy',
    uptime: '99.88%',
    latency: '67ms',
    requests: '1.2K/min',
    errors: '0.02%',
  },
  {
    name: 'Health Monitor',
    status: 'healthy',
    uptime: '99.98%',
    latency: '28ms',
    requests: '15.6K/min',
    errors: '0.00%',
  },
  {
    name: 'Blockchain Indexer',
    status: 'healthy',
    uptime: '99.85%',
    latency: '234ms',
    requests: '2.1K/min',
    errors: '0.08%',
  },
  {
    name: 'Reporting Service',
    status: 'unhealthy',
    uptime: '95.20%',
    latency: '567ms',
    requests: '0.5K/min',
    errors: '2.50%',
  },
];

const metricsData = [
  { time: '00:00', cpu: 45, memory: 62, network: 125, requests: 8500 },
  { time: '04:00', cpu: 38, memory: 58, network: 98, requests: 6200 },
  { time: '08:00', cpu: 65, memory: 72, network: 245, requests: 15200 },
  { time: '12:00', cpu: 78, memory: 85, network: 312, requests: 18500 },
  { time: '16:00', cpu: 72, memory: 80, network: 287, requests: 16800 },
  { time: '20:00', cpu: 58, memory: 70, network: 198, requests: 11200 },
  { time: '23:59', cpu: 48, memory: 65, network: 145, requests: 8900 },
];

const containerVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'healthy':
      return <CheckCircle className="w-5 h-5 text-success" />;
    case 'degraded':
      return <AlertTriangle className="w-5 h-5 text-warning" />;
    case 'unhealthy':
      return <XCircle className="w-5 h-5 text-danger" />;
    default:
      return <AlertTriangle className="w-5 h-5 text-gray-400" />;
  }
};

const getStatusBadge = (status: string) => {
  switch (status) {
    case 'healthy':
      return 'badge-success';
    case 'degraded':
      return 'badge-warning';
    case 'unhealthy':
      return 'badge-danger';
    default:
      return 'badge-gray';
  }
};

export default function Health() {
  const overallHealth =
    (services.filter((s) => s.status === 'healthy').length / services.length) * 100;

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="page-header">System Health</h1>
        <p className="page-description">
          Real-time monitoring of platform services and infrastructure
        </p>
      </div>

      {/* Overall Health */}
      <motion.div
        variants={containerVariants}
        initial="hidden"
        animate="visible"
        className="card p-8"
      >
        <div className="flex items-center justify-between">
          <div>
            <p className="text-gray-400">Overall Platform Health</p>
            <div className="flex items-center gap-4 mt-2">
              <h2 className="text-4xl font-bold text-white">
                {overallHealth.toFixed(1)}%
              </h2>
              <div className="flex items-center gap-2 text-success">
                <TrendingUp className="w-5 h-5" />
                <span>+0.5% from last hour</span>
              </div>
            </div>
          </div>
          <div className="flex items-center gap-8">
            <div className="text-center">
              <p className="text-3xl font-bold text-success">
                {services.filter((s) => s.status === 'healthy').length}
              </p>
              <p className="text-sm text-gray-400">Healthy</p>
            </div>
            <div className="text-center">
              <p className="text-3xl font-bold text-warning">1</p>
              <p className="text-sm text-gray-400">Degraded</p>
            </div>
            <div className="text-center">
              <p className="text-3xl font-bold text-danger">1</p>
              <p className="text-sm text-gray-400">Unhealthy</p>
            </div>
          </div>
        </div>

        {/* Health bar */}
        <div className="mt-6">
          <div className="flex h-4 rounded-full overflow-hidden bg-background-lighter">
            <div
              className="bg-success transition-all duration-500"
              style={{
                width: `${(services.filter((s) => s.status === 'healthy').length / services.length) * 100}%`,
              }}
            />
            <div
              className="bg-warning transition-all duration-500"
              style={{
                width: `${(services.filter((s) => s.status === 'degraded').length / services.length) * 100}%`,
              }}
            />
            <div
              className="bg-danger transition-all duration-500"
              style={{
                width: `${(services.filter((s) => s.status === 'unhealthy').length / services.length) * 100}%`,
              }}
            />
          </div>
          <div className="flex justify-between mt-2 text-sm text-gray-400">
            <span>0%</span>
            <span>50%</span>
            <span>100%</span>
          </div>
        </div>
      </motion.div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="card"
        >
          <div className="card-header">
            <h3 className="section-title mb-0">CPU & Memory Usage</h3>
          </div>
          <div className="card-body">
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={metricsData}>
                <defs>
                  <linearGradient id="colorCpu" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="colorMemory" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
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
                />
                <Area
                  type="monotone"
                  dataKey="cpu"
                  stroke="#3b82f6"
                  strokeWidth={2}
                  fillOpacity={1}
                  fill="url(#colorCpu)"
                  name="CPU %"
                />
                <Area
                  type="monotone"
                  dataKey="memory"
                  stroke="#10b981"
                  strokeWidth={2}
                  fillOpacity={1}
                  fill="url(#colorMemory)"
                  name="Memory %"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </motion.div>

        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="card"
        >
          <div className="card-header">
            <h3 className="section-title mb-0">Request Rate</h3>
          </div>
          <div className="card-body">
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={metricsData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="time" stroke="#64748b" fontSize={12} />
                <YAxis stroke="#64748b" fontSize={12} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: '#1e293b',
                    border: '1px solid #334155',
                    borderRadius: '8px',
                  }}
                />
                <Line
                  type="monotone"
                  dataKey="requests"
                  stroke="#f59e0b"
                  strokeWidth={2}
                  dot={false}
                  name="Requests/min"
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </motion.div>
      </div>

      {/* Services Grid */}
      <motion.div
        variants={containerVariants}
        initial="hidden"
        animate="visible"
      >
        <h3 className="section-title">Service Status</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {services.map((service, index) => (
            <motion.div
              key={service.name}
              variants={{
                hidden: { opacity: 0, y: 20 },
                visible: { opacity: 1, y: 0 },
              }}
              transition={{ delay: index * 0.1 }}
              className={`card p-4 border ${
                service.status === 'healthy'
                  ? 'border-success/20'
                  : service.status === 'degraded'
                  ? 'border-warning/20'
                  : 'border-danger/20'
              }`}
            >
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  {getStatusIcon(service.status)}
                  <span className="font-medium text-white">{service.name}</span>
                </div>
                <span className={`badge ${getStatusBadge(service.status)}`}>
                  {service.status}
                </span>
              </div>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-400">Uptime</span>
                  <span className="text-white">{service.uptime}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">Latency</span>
                  <span className="text-white">{service.latency}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">Requests</span>
                  <span className="text-white">{service.requests}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">Errors</span>
                  <span className={service.errors.startsWith('0') ? 'text-success' : 'text-danger'}>
                    {service.errors}
                  </span>
                </div>
              </div>
            </motion.div>
          ))}
        </div>
      </motion.div>
    </div>
  );
}
