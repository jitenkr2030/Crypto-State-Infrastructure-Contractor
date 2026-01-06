import { useState } from 'react';
import { motion } from 'framer-motion';
import {
  Shield,
  Plus,
  Search,
  Filter,
  CheckCircle,
  XCircle,
  AlertTriangle,
  AlertOctagon,
  Info,
  Edit,
  Trash2,
  Play,
  Pause,
  MoreVertical,
} from 'lucide-react';

const rules = [
  {
    id: '1',
    name: 'Blacklisted Entity Check',
    description: 'Verify source and target are not on any blacklist',
    type: 'AML',
    severity: 'CRITICAL',
    enabled: true,
    version: 1,
    checks: 15423,
    failures: 12,
  },
  {
    id: '2',
    name: 'Watchlist Match Check',
    description: 'Check against known watchlists',
    type: 'AML',
    severity: 'HIGH',
    enabled: true,
    version: 2,
    checks: 15423,
    failures: 45,
  },
  {
    id: '3',
    name: 'Source KYC Verification',
    description: 'Verify source entity has completed KYC',
    type: 'KYC',
    severity: 'HIGH',
    enabled: true,
    version: 1,
    checks: 15423,
    failures: 234,
  },
  {
    id: '4',
    name: 'Sanctioned Country Check',
    description: 'Check source and target countries against sanctions list',
    type: 'SANCTIONS',
    severity: 'CRITICAL',
    enabled: true,
    version: 3,
    checks: 15423,
    failures: 3,
  },
  {
    id: '5',
    name: 'Required Fields Check',
    description: 'Verify all required transaction fields are present',
    type: 'TRANSACTION',
    severity: 'MEDIUM',
    enabled: true,
    version: 1,
    checks: 15423,
    failures: 56,
  },
  {
    id: '6',
    name: 'Maximum Amount Check',
    description: 'Flag transactions exceeding maximum amount',
    type: 'AMOUNT',
    severity: 'HIGH',
    enabled: false,
    version: 1,
    checks: 15423,
    failures: 89,
  },
];

const recentResults = [
  {
    id: 'tx-1',
    transactionId: 'TX-2024-001234',
    status: 'PASS',
    riskScore: 0.12,
    checksPassed: 5,
    checksFailed: 0,
    timestamp: '5 minutes ago',
  },
  {
    id: 'tx-2',
    transactionId: 'TX-2024-001235',
    status: 'FAIL',
    riskScore: 0.85,
    checksPassed: 3,
    checksFailed: 2,
    timestamp: '15 minutes ago',
  },
  {
    id: 'tx-3',
    transactionId: 'TX-2024-001236',
    status: 'WARN',
    riskScore: 0.52,
    checksPassed: 4,
    checksFailed: 1,
    timestamp: '30 minutes ago',
  },
  {
    id: 'tx-4',
    transactionId: 'TX-2024-001237',
    status: 'PASS',
    riskScore: 0.08,
    checksPassed: 5,
    checksFailed: 0,
    timestamp: '1 hour ago',
  },
];

const containerVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};

const getSeverityIcon = (severity: string) => {
  switch (severity) {
    case 'CRITICAL':
      return <AlertOctagon className="w-4 h-4 text-danger" />;
    case 'HIGH':
      return <AlertTriangle className="w-4 h-4 text-warning" />;
    case 'MEDIUM':
      return <Info className="w-4 h-4 text-info" />;
    default:
      return <Info className="w-4 h-4 text-gray-400" />;
  }
};

const getSeverityBadge = (severity: string) => {
  switch (severity) {
    case 'CRITICAL':
      return 'badge-danger';
    case 'HIGH':
      return 'badge-warning';
    case 'MEDIUM':
      return 'badge-info';
    default:
      return 'badge-gray';
  }
};

const getStatusBadge = (status: string) => {
  switch (status) {
    case 'PASS':
      return 'badge-success';
    case 'FAIL':
      return 'badge-danger';
    case 'WARN':
      return 'badge-warning';
    default:
      return 'badge-gray';
  }
};

export default function Compliance() {
  const [activeTab, setActiveTab] = useState('rules');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedRule, setSelectedRule] = useState<string | null>(null);

  const filteredRules = rules.filter(
    (rule) =>
      rule.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      rule.type.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="page-header">Compliance Management</h1>
          <p className="page-description">
            Configure and monitor compliance rules and check results
          </p>
        </div>
        <button className="btn-primary">
          <Plus className="w-4 h-4 mr-2" />
          Add Rule
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
              <Shield className="w-6 h-6 text-primary" />
            </div>
            <div>
              <p className="stat-label">Total Rules</p>
              <p className="stat-value">{rules.length}</p>
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-xl bg-success/10">
              <CheckCircle className="w-6 h-6 text-success" />
            </div>
            <div>
              <p className="stat-label">Active Rules</p>
              <p className="stat-value">
                {rules.filter((r) => r.enabled).length}
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
              <p className="stat-label">Critical Rules</p>
              <p className="stat-value">
                {rules.filter((r) => r.severity === 'CRITICAL').length}
              </p>
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-xl bg-warning/10">
              <AlertTriangle className="w-6 h-6 text-warning" />
            </div>
            <div>
              <p className="stat-label">Total Failures</p>
              <p className="stat-value">
                {rules.reduce((acc, r) => acc + r.failures, 0)}
              </p>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Tabs */}
      <div className="flex gap-2 border-b border-gray-700">
        {['rules', 'results'].map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors capitalize ${
              activeTab === tab
                ? 'text-primary border-primary'
                : 'text-gray-400 border-transparent hover:text-white'
            }`}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Rules Tab */}
      {activeTab === 'rules' && (
        <>
          {/* Search and Filters */}
          <div className="flex gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="text"
                placeholder="Search rules..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="input pl-10"
              />
            </div>
            <button className="btn-secondary">
              <Filter className="w-4 h-4 mr-2" />
              Filter
            </button>
          </div>

          {/* Rules Table */}
          <motion.div
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            className="card overflow-hidden"
          >
            <table className="table">
              <thead>
                <tr>
                  <th>Rule</th>
                  <th>Type</th>
                  <th>Severity</th>
                  <th>Status</th>
                  <th>Checks</th>
                  <th>Failures</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredRules.map((rule) => (
                  <tr
                    key={rule.id}
                    className={`cursor-pointer ${
                      selectedRule === rule.id ? 'bg-background-lighter/50' : ''
                    }`}
                    onClick={() => setSelectedRule(rule.id)}
                  >
                    <td>
                      <div>
                        <p className="font-medium text-white">{rule.name}</p>
                        <p className="text-xs text-gray-400">{rule.description}</p>
                      </div>
                    </td>
                    <td>
                      <span className="badge badge-gray">{rule.type}</span>
                    </td>
                    <td>
                      <span className={`badge ${getSeverityBadge(rule.severity)} flex items-center gap-1 w-fit`}>
                        {getSeverityIcon(rule.severity)}
                        {rule.severity}
                      </span>
                    </td>
                    <td>
                      <button
                        onClick={(e) => e.stopPropagation()}
                        className={`p-1 rounded ${
                          rule.enabled ? 'text-success' : 'text-gray-400'
                        }`}
                      >
                        {rule.enabled ? (
                          <Play className="w-4 h-4" />
                        ) : (
                          <Pause className="w-4 h-4" />
                        )}
                      </button>
                    </td>
                    <td className="font-mono text-sm">
                      {rule.checks.toLocaleString()}
                    </td>
                    <td className="font-mono text-sm text-danger">
                      {rule.failures}
                    </td>
                    <td>
                      <div className="flex items-center gap-1">
                        <button className="p-1.5 rounded hover:bg-background-lighter text-gray-400 hover:text-white">
                          <Edit className="w-4 h-4" />
                        </button>
                        <button className="p-1.5 rounded hover:bg-background-lighter text-gray-400 hover:text-danger">
                          <Trash2 className="w-4 h-4" />
                        </button>
                        <button className="p-1.5 rounded hover:bg-background-lighter text-gray-400 hover:text-white">
                          <MoreVertical className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </motion.div>
        </>
      )}

      {/* Results Tab */}
      {activeTab === 'results' && (
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="card overflow-hidden"
        >
          <table className="table">
            <thead>
              <tr>
                <th>Transaction ID</th>
                <th>Status</th>
                <th>Risk Score</th>
                <th>Checks</th>
                <th>Timestamp</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {recentResults.map((result) => (
                <tr key={result.id}>
                  <td className="font-mono font-medium">{result.transactionId}</td>
                  <td>
                    <span className={`badge ${getStatusBadge(result.status)}`}>
                      {result.status}
                    </span>
                  </td>
                  <td>
                    <div className="flex items-center gap-2">
                      <div className="w-16 h-2 bg-background-lighter rounded-full overflow-hidden">
                        <div
                          className={`h-full rounded-full ${
                            result.riskScore > 0.7
                              ? 'bg-danger'
                              : result.riskScore > 0.4
                              ? 'bg-warning'
                              : 'bg-success'
                          }`}
                          style={{ width: `${result.riskScore * 100}%` }}
                        />
                      </div>
                      <span className="text-sm font-mono">{result.riskScore.toFixed(2)}</span>
                    </div>
                  </td>
                  <td className="text-sm">
                    <span className="text-success">{result.checksPassed}</span>/
                    <span className="text-gray-400">{result.checksPassed + result.checksFailed}</span>
                  </td>
                  <td className="text-gray-400">{result.timestamp}</td>
                  <td>
                    <button className="btn-ghost btn-sm">View Details</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </motion.div>
      )}
    </div>
  );
}
