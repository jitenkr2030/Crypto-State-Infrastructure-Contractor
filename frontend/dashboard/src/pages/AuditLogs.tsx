import { useState } from 'react';
import { motion } from 'framer-motion';
import {
  FileText,
  Search,
  Filter,
  Download,
  Calendar,
  User,
  Globe,
  Server,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Clock,
} from 'lucide-react';

const auditLogs = [
  {
    id: '1',
    timestamp: '2024-01-15T10:30:00Z',
    eventType: 'USER_LOGIN',
    service: 'Auth Service',
    action: 'login',
    actor: 'admin@csic.com',
    actorType: 'user',
    resource: '/api/v1/auth/login',
    resourceType: 'endpoint',
    outcome: 'success',
    severity: 'INFO',
    ipAddress: '192.168.1.100',
  },
  {
    id: '2',
    timestamp: '2024-01-15T10:31:15Z',
    eventType: 'COMPLIANCE_CHECK',
    service: 'Compliance Service',
    action: 'check',
    actor: 'system',
    actorType: 'service',
    resource: 'TX-2024-001234',
    resourceType: 'transaction',
    outcome: 'success',
    severity: 'INFO',
    ipAddress: 'internal',
  },
  {
    id: '3',
    timestamp: '2024-01-15T10:32:30Z',
    eventType: 'VIOLATION_DETECTED',
    service: 'Compliance Service',
    action: 'violation',
    actor: 'system',
    actorType: 'service',
    resource: 'TX-2024-001235',
    resourceType: 'transaction',
    outcome: 'failure',
    severity: 'HIGH',
    ipAddress: 'internal',
  },
  {
    id: '4',
    timestamp: '2024-01-15T10:33:45Z',
    eventType: 'NODE_START',
    service: 'Node Manager',
    action: 'start',
    actor: 'admin@csic.com',
    actorType: 'user',
    resource: 'node-01',
    resourceType: 'node',
    outcome: 'success',
    severity: 'INFO',
    ipAddress: '192.168.1.100',
  },
  {
    id: '5',
    timestamp: '2024-01-15T10:34:00Z',
    eventType: 'RULE_UPDATE',
    service: 'Compliance Service',
    action: 'update',
    actor: 'admin@csic.com',
    actorType: 'user',
    resource: 'rule-001',
    resourceType: 'rule',
    outcome: 'success',
    severity: 'INFO',
    ipAddress: '192.168.1.100',
  },
  {
    id: '6',
    timestamp: '2024-01-15T10:35:15Z',
    eventType: 'UNAUTHORIZED_ACCESS',
    service: 'API Gateway',
    action: 'access_denied',
    actor: 'unknown',
    actorType: 'user',
    resource: '/api/v1/admin',
    resourceType: 'endpoint',
    outcome: 'failure',
    severity: 'HIGH',
    ipAddress: '203.0.113.50',
  },
];

const containerVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};

const getOutcomeIcon = (outcome: string) => {
  switch (outcome) {
    case 'success':
      return <CheckCircle className="w-4 h-4 text-success" />;
    case 'failure':
      return <XCircle className="w-4 h-4 text-danger" />;
    default:
      return <AlertTriangle className="w-4 h-4 text-warning" />;
  }
};

const getSeverityBadge = (severity: string) => {
  switch (severity) {
    case 'HIGH':
      return 'badge-danger';
    case 'MEDIUM':
      return 'badge-warning';
    default:
      return 'badge-gray';
  }
};

const formatTimestamp = (timestamp: string) => {
  const date = new Date(timestamp);
  return date.toLocaleString();
};

export default function AuditLogs() {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedLog, setSelectedLog] = useState<string | null>(null);

  const filteredLogs = auditLogs.filter(
    (log) =>
      log.eventType.toLowerCase().includes(searchQuery.toLowerCase()) ||
      log.actor.toLowerCase().includes(searchQuery.toLowerCase()) ||
      log.service.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="page-header">Audit Logs</h1>
          <p className="page-description">
            View and search platform audit trail
          </p>
        </div>
        <button className="btn-secondary">
          <Download className="w-4 h-4 mr-2" />
          Export Logs
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
              <FileText className="w-6 h-6 text-primary" />
            </div>
            <div>
              <p className="stat-label">Total Events</p>
              <p className="stat-value">{auditLogs.length * 100}</p>
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-xl bg-success/10">
              <CheckCircle className="w-6 h-6 text-success" />
            </div>
            <div>
              <p className="stat-label">Successful</p>
              <p className="stat-value">{auditLogs.filter((l) => l.outcome === 'success').length * 100}</p>
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-xl bg-danger/10">
              <XCircle className="w-6 h-6 text-danger" />
            </div>
            <div>
              <p className="stat-label">Failed</p>
              <p className="stat-value">{auditLogs.filter((l) => l.outcome === 'failure').length * 100}</p>
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-xl bg-warning/10">
              <AlertTriangle className="w-6 h-6 text-warning" />
            </div>
            <div>
              <p className="stat-label">High Severity</p>
              <p className="stat-value">{auditLogs.filter((l) => l.severity === 'HIGH').length * 10}</p>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Search and Filters */}
      <div className="flex gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search audit logs..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="input pl-10"
          />
        </div>
        <button className="btn-secondary">
          <Filter className="w-4 h-4 mr-2" />
          Filter
        </button>
        <button className="btn-secondary">
          <Calendar className="w-4 h-4 mr-2" />
          Date Range
        </button>
      </div>

      {/* Logs Table */}
      <motion.div
        variants={containerVariants}
        initial="hidden"
        animate="visible"
        className="card overflow-hidden"
      >
        <table className="table">
          <thead>
            <tr>
              <th>Timestamp</th>
              <th>Event Type</th>
              <th>Service</th>
              <th>Actor</th>
              <th>Outcome</th>
              <th>Severity</th>
              <th>Details</th>
            </tr>
          </thead>
          <tbody>
            {filteredLogs.map((log) => (
              <tr
                key={log.id}
                className={`cursor-pointer ${
                  selectedLog === log.id ? 'bg-background-lighter/50' : ''
                }`}
                onClick={() => setSelectedLog(log.id)}
              >
                <td className="text-sm text-gray-400">
                  {formatTimestamp(log.timestamp)}
                </td>
                <td>
                  <span className="font-medium text-white">{log.eventType}</span>
                </td>
                <td className="text-gray-400">{log.service}</td>
                <td>
                  <div className="flex items-center gap-2">
                    <User className="w-4 h-4 text-gray-400" />
                    <span className="text-sm">{log.actor}</span>
                  </div>
                </td>
                <td>
                  <div className="flex items-center gap-1">
                    {getOutcomeIcon(log.outcome)}
                    <span className="capitalize text-sm">{log.outcome}</span>
                  </div>
                </td>
                <td>
                  <span className={`badge ${getSeverityBadge(log.severity)}`}>
                    {log.severity}
                  </span>
                </td>
                <td>
                  <button className="btn-ghost btn-sm">View</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </motion.div>

      {/* Log Details Panel */}
      {selectedLog && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="card p-6"
        >
          <h3 className="section-title">Log Details</h3>
          {(() => {
            const log = auditLogs.find((l) => l.id === selectedLog);
            if (!log) return null;
            return (
              <div className="grid grid-cols-2 md:grid-cols-3 gap-6 mt-4">
                <div>
                  <p className="text-sm text-gray-400">Event Type</p>
                  <p className="font-medium text-white">{log.eventType}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-400">Service</p>
                  <p className="font-medium text-white">{log.service}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-400">Action</p>
                  <p className="font-medium text-white">{log.action}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-400">Actor</p>
                  <p className="font-medium text-white">{log.actor}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-400">Resource</p>
                  <p className="font-medium text-white">{log.resource}</p>
                </div>
                <div>
                  <p className="text-sm text-gray-400">IP Address</p>
                  <p className="font-medium text-white">{log.ipAddress}</p>
                </div>
              </div>
            );
          })()}
        </motion.div>
      )}
    </div>
  );
}
