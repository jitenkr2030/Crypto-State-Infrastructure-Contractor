import { useState } from 'react';
import { motion } from 'framer-motion';
import {
  Server,
  Play,
  Square,
  RefreshCw,
  Settings,
  MoreVertical,
  Cpu,
  HardDrive,
  Wifi,
  Clock,
  CheckCircle,
  AlertTriangle,
  XCircle,
  Plus,
} from 'lucide-react';

const nodes = [
  {
    id: 'node-1',
    name: 'Node-01',
    type: 'Validator',
    network: 'Mainnet',
    status: 'running',
    version: '2.1.0',
    height: 15489234,
    peers: 156,
    uptime: '99.9%',
    cpuUsage: 45,
    memoryUsage: 62,
    diskUsage: 78,
    lastBlockTime: '2 seconds ago',
    region: 'us-east-1',
  },
  {
    id: 'node-2',
    name: 'Node-02',
    type: 'Validator',
    network: 'Mainnet',
    status: 'running',
    version: '2.1.0',
    height: 15489234,
    peers: 154,
    uptime: '99.8%',
    cpuUsage: 42,
    memoryUsage: 58,
    diskUsage: 75,
    lastBlockTime: '3 seconds ago',
    region: 'us-west-2',
  },
  {
    id: 'node-3',
    name: 'Node-03',
    type: 'RPC',
    network: 'Mainnet',
    status: 'syncing',
    version: '2.1.0',
    height: 15489230,
    peers: 89,
    uptime: '98.5%',
    cpuUsage: 78,
    memoryUsage: 71,
    diskUsage: 76,
    lastBlockTime: '45 seconds ago',
    region: 'eu-west-1',
  },
  {
    id: 'node-4',
    name: 'Node-04',
    type: 'Archive',
    network: 'Mainnet',
    status: 'running',
    version: '2.1.0',
    height: 15489234,
    peers: 45,
    uptime: '99.2%',
    cpuUsage: 55,
    memoryUsage: 85,
    diskUsage: 92,
    lastBlockTime: '1 second ago',
    region: 'ap-southeast-1',
  },
  {
    id: 'node-5',
    name: 'Node-05',
    type: 'Validator',
    network: 'Testnet',
    status: 'error',
    version: '2.1.0',
    height: 8934234,
    peers: 23,
    uptime: '95.0%',
    cpuUsage: 12,
    memoryUsage: 34,
    diskUsage: 45,
    lastBlockTime: '5 minutes ago',
    region: 'us-east-1',
  },
];

const containerVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'running':
      return <CheckCircle className="w-4 h-4 text-success" />;
    case 'syncing':
      return <RefreshCw className="w-4 h-4 text-warning animate-spin" />;
    case 'error':
      return <XCircle className="w-4 h-4 text-danger" />;
    default:
      return <AlertTriangle className="w-4 h-4 text-gray-400" />;
  }
};

const getStatusBadge = (status: string) => {
  switch (status) {
    case 'running':
      return 'badge-success';
    case 'syncing':
      return 'badge-warning';
    case 'error':
      return 'badge-danger';
    default:
      return 'badge-gray';
  }
};

export default function Nodes() {
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  const [filter, setFilter] = useState('all');

  const filteredNodes = nodes.filter(
    (node) => filter === 'all' || node.status === filter
  );

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="page-header">Node Management</h1>
          <p className="page-description">
            Monitor and manage your blockchain nodes
          </p>
        </div>
        <button className="btn-primary">
          <Plus className="w-4 h-4 mr-2" />
          Deploy Node
        </button>
      </div>

      {/* Stats */}
      <motion.div
        variants={containerVariants}
        initial="hidden"
        animate="visible"
        className="grid grid-cols-1 md:grid-cols-4 gap-6"
      >
        <div className="stat-card">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-xl bg-primary/10">
              <Server className="w-6 h-6 text-primary" />
            </div>
            <div>
              <p className="stat-label">Total Nodes</p>
              <p className="stat-value">{nodes.length}</p>
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-xl bg-success/10">
              <CheckCircle className="w-6 h-6 text-success" />
            </div>
            <div>
              <p className="stat-label">Running</p>
              <p className="stat-value">
                {nodes.filter((n) => n.status === 'running').length}
              </p>
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-xl bg-warning/10">
              <RefreshCw className="w-6 h-6 text-warning" />
            </div>
            <div>
              <p className="stat-label">Syncing</p>
              <p className="stat-value">
                {nodes.filter((n) => n.status === 'syncing').length}
              </p>
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-xl bg-danger/10">
              <XCircle className="w-6 h-6 text-danger" />
            </div>
            <div>
              <p className="stat-label">Error</p>
              <p className="stat-value">
                {nodes.filter((n) => n.status === 'error').length}
              </p>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Filters */}
      <div className="flex gap-2">
        {['all', 'running', 'syncing', 'error'].map((status) => (
          <button
            key={status}
            onClick={() => setFilter(status)}
            className={`btn ${
              filter === status ? 'btn-primary' : 'btn-secondary'
            } btn-sm capitalize`}
          >
            {status}
          </button>
        ))}
      </div>

      {/* Node List */}
      <motion.div
        variants={containerVariants}
        initial="hidden"
        animate="visible"
        className="grid gap-4"
      >
        {filteredNodes.map((node) => (
          <motion.div
            key={node.id}
            variants={{
              hidden: { opacity: 0, y: 20 },
              visible: { opacity: 1, y: 0 },
            }}
            className={`card p-6 cursor-pointer transition-all hover:border-primary ${
              selectedNode === node.id ? 'border-primary' : ''
            }`}
            onClick={() => setSelectedNode(node.id)}
          >
            <div className="flex items-start justify-between">
              <div className="flex items-start gap-4">
                <div className="p-3 rounded-xl bg-background-lighter">
                  <Server className="w-6 h-6 text-gray-400" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-lg font-semibold text-white">{node.name}</h3>
                    {getStatusIcon(node.status)}
                    <span className={`badge ${getStatusBadge(node.status)} capitalize`}>
                      {node.status}
                    </span>
                  </div>
                  <p className="text-sm text-gray-400">
                    {node.type} • {node.network} • {node.region}
                  </p>
                  <div className="flex items-center gap-4 mt-3 text-sm text-gray-400">
                    <span className="flex items-center gap-1">
                      <Cpu className="w-4 h-4" />
                      {node.cpuUsage}% CPU
                    </span>
                    <span className="flex items-center gap-1">
                      <HardDrive className="w-4 h-4" />
                      {node.memoryUsage}% RAM
                    </span>
                    <span className="flex items-center gap-1">
                      <Wifi className="w-4 h-4" />
                      {node.peers} peers
                    </span>
                    <span className="flex items-center gap-1">
                      <Clock className="w-4 h-4" />
                      {node.uptime} uptime
                    </span>
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                {node.status === 'running' ? (
                  <button className="btn-ghost btn-sm">
                    <Square className="w-4 h-4" />
                  </button>
                ) : node.status === 'error' ? (
                  <button className="btn-ghost btn-sm">
                    <RefreshCw className="w-4 h-4" />
                  </button>
                ) : null}
                <button className="btn-ghost btn-sm">
                  <Settings className="w-4 h-4" />
                </button>
                <button className="btn-ghost btn-sm">
                  <MoreVertical className="w-4 h-4" />
                </button>
              </div>
            </div>

            {/* Expanded Details */}
            {selectedNode === node.id && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: 'auto' }}
                className="mt-6 pt-6 border-t border-gray-700"
              >
                <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
                  <div>
                    <p className="text-sm text-gray-400">Block Height</p>
                    <p className="text-lg font-mono text-white">{node.height.toLocaleString()}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-400">Version</p>
                    <p className="text-lg text-white">{node.version}</p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-400">Disk Usage</p>
                    <div className="flex items-center gap-2">
                      <div className="flex-1 h-2 bg-background-lighter rounded-full overflow-hidden">
                        <div
                          className={`h-full rounded-full ${
                            node.diskUsage > 90
                              ? 'bg-danger'
                              : node.diskUsage > 75
                              ? 'bg-warning'
                              : 'bg-success'
                          }`}
                          style={{ width: `${node.diskUsage}%` }}
                        />
                      </div>
                      <span className="text-sm text-white">{node.diskUsage}%</span>
                    </div>
                  </div>
                  <div>
                    <p className="text-sm text-gray-400">Last Block</p>
                    <p className="text-sm text-white">{node.lastBlockTime}</p>
                  </div>
                </div>
                <div className="flex gap-2 mt-6">
                  <button className="btn-primary btn-sm">
                    <Play className="w-4 h-4 mr-1" />
                    Start
                  </button>
                  <button className="btn-secondary btn-sm">
                    <Square className="w-4 h-4 mr-1" />
                    Stop
                  </button>
                  <button className="btn-secondary btn-sm">
                    <RefreshCw className="w-4 h-4 mr-1" />
                    Restart
                  </button>
                </div>
              </motion.div>
            )}
          </motion.div>
        ))}
      </motion.div>
    </div>
  );
}
