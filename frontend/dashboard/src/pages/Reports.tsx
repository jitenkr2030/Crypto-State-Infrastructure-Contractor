import { useState } from 'react';
import { motion } from 'framer-motion';
import {
  BarChart3,
  FileText,
  Download,
  Calendar,
  Clock,
  CheckCircle,
  XCircle,
  Loader2,
  Plus,
  Filter,
} from 'lucide-react';

const reports = [
  {
    id: '1',
    name: 'Monthly Compliance Report',
    type: 'COMPLIANCE',
    status: 'completed',
    format: 'PDF',
    size: '2.4 MB',
    createdAt: '2024-01-15T10:00:00Z',
    completedAt: '2024-01-15T10:05:00Z',
  },
  {
    id: '2',
    name: 'Transaction Audit Log',
    type: 'AUDIT',
    status: 'completed',
    format: 'CSV',
    size: '15.8 MB',
    createdAt: '2024-01-14T09:00:00Z',
    completedAt: '2024-01-14T09:02:00Z',
  },
  {
    id: '3',
    name: 'Node Performance Summary',
    type: 'PERFORMANCE',
    status: 'completed',
    format: 'PDF',
    size: '1.2 MB',
    createdAt: '2024-01-13T08:00:00Z',
    completedAt: '2024-01-13T08:03:00Z',
  },
  {
    id: '4',
    name: 'Violation Analysis Report',
    type: 'COMPLIANCE',
    status: 'generating',
    format: 'PDF',
    size: null,
    createdAt: '2024-01-15T11:00:00Z',
    completedAt: null,
  },
  {
    id: '5',
    name: 'Annual Regulatory Report',
    type: 'REGULATORY',
    status: 'failed',
    format: 'PDF',
    size: null,
    createdAt: '2024-01-12T14:00:00Z',
    completedAt: null,
    error: 'Data extraction timeout',
  },
];

const templates = [
  {
    id: '1',
    name: 'Monthly Compliance Summary',
    description: 'Overview of compliance checks and violations for the month',
    type: 'COMPLIANCE',
  },
  {
    id: '2',
    name: 'Transaction Activity Report',
    description: 'Detailed breakdown of all transactions',
    type: 'AUDIT',
  },
  {
    id: '3',
    name: 'Node Health Report',
    description: 'Performance and uptime metrics for all nodes',
    type: 'PERFORMANCE',
  },
  {
    id: '4',
    name: 'SAR (Suspicious Activity Report)',
    description: 'Generate SAR for regulatory compliance',
    type: 'REGULATORY',
  },
];

const containerVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};

const getStatusBadge = (status: string) => {
  switch (status) {
    case 'completed':
      return 'badge-success';
    case 'generating':
      return 'badge-warning';
    case 'failed':
      return 'badge-danger';
    default:
      return 'badge-gray';
  }
};

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'completed':
      return <CheckCircle className="w-4 h-4 text-success" />;
    case 'generating':
      return <Loader2 className="w-4 h-4 text-warning animate-spin" />;
    case 'failed':
      return <XCircle className="w-4 h-4 text-danger" />;
    default:
      return <Clock className="w-4 h-4 text-gray-400" />;
  }
};

const formatDate = (dateString: string) => {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

export default function Reports() {
  const [activeTab, setActiveTab] = useState('history');
  const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null);

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="page-header">Reports</h1>
          <p className="page-description">
            Generate and download compliance and audit reports
          </p>
        </div>
        <button className="btn-primary">
          <Plus className="w-4 h-4 mr-2" />
          Generate Report
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
              <p className="stat-label">Total Reports</p>
              <p className="stat-value">{reports.length}</p>
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-xl bg-success/10">
              <CheckCircle className="w-6 h-6 text-success" />
            </div>
            <div>
              <p className="stat-label">Completed</p>
              <p className="stat-value">
                {reports.filter((r) => r.status === 'completed').length}
              </p>
            </div>
          </div>
        </div>
        <div className="stat-card">
          <div className="flex items-center gap-4">
            <div className="p-3 rounded-xl bg-warning/10">
              <Loader2 className="w-6 h-6 text-warning" />
            </div>
            <div>
              <p className="stat-label">Generating</p>
              <p className="stat-value">
                {reports.filter((r) => r.status === 'generating').length}
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
              <p className="stat-label">Failed</p>
              <p className="stat-value">
                {reports.filter((r) => r.status === 'failed').length}
              </p>
            </div>
          </div>
        </div>
      </motion.div>

      {/* Tabs */}
      <div className="flex gap-2 border-b border-gray-700">
        {['history', 'templates'].map((tab) => (
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

      {/* History Tab */}
      {activeTab === 'history' && (
        <>
          {/* Filters */}
          <div className="flex gap-4">
            <div className="relative flex-1">
              <Filter className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <select className="input pl-10 appearance-none cursor-pointer">
                <option value="">All Types</option>
                <option value="COMPLIANCE">Compliance</option>
                <option value="AUDIT">Audit</option>
                <option value="PERFORMANCE">Performance</option>
                <option value="REGULATORY">Regulatory</option>
              </select>
            </div>
            <button className="btn-secondary">
              <Calendar className="w-4 h-4 mr-2" />
              Date Range
            </button>
          </div>

          {/* Reports Table */}
          <motion.div
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            className="card overflow-hidden"
          >
            <table className="table">
              <thead>
                <tr>
                  <th>Report Name</th>
                  <th>Type</th>
                  <th>Status</th>
                  <th>Format</th>
                  <th>Size</th>
                  <th>Created</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {reports.map((report) => (
                  <tr key={report.id}>
                    <td>
                      <div className="flex items-center gap-3">
                        <div className="p-2 rounded-lg bg-background-lighter">
                          <FileText className="w-5 h-5 text-gray-400" />
                        </div>
                        <span className="font-medium text-white">{report.name}</span>
                      </div>
                    </td>
                    <td>
                      <span className="badge badge-gray">{report.type}</span>
                    </td>
                    <td>
                      <span className={`badge ${getStatusBadge(report.status)} flex items-center gap-1 w-fit`}>
                        {getStatusIcon(report.status)}
                        {report.status}
                      </span>
                    </td>
                    <td className="font-mono text-sm">{report.format}</td>
                    <td className="text-gray-400">{report.size || '-'}</td>
                    <td className="text-sm text-gray-400">
                      {formatDate(report.createdAt)}
                    </td>
                    <td>
                      <div className="flex items-center gap-2">
                        {report.status === 'completed' && (
                          <button className="btn-ghost btn-sm">
                            <Download className="w-4 h-4" />
                          </button>
                        )}
                        {report.status === 'failed' && (
                          <button className="btn-ghost btn-sm text-danger">
                            Retry
                          </button>
                        )}
                        <button className="btn-ghost btn-sm">View</button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </motion.div>
        </>
      )}

      {/* Templates Tab */}
      {activeTab === 'templates' && (
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="grid grid-cols-1 md:grid-cols-2 gap-6"
        >
          {templates.map((template) => (
            <motion.div
              key={template.id}
              variants={{
                hidden: { opacity: 0, y: 20 },
                visible: { opacity: 1, y: 0 },
              }}
              className={`card p-6 cursor-pointer transition-all hover:border-primary ${
                selectedTemplate === template.id ? 'border-primary' : ''
              }`}
              onClick={() => setSelectedTemplate(template.id)}
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-4">
                  <div className="p-3 rounded-xl bg-primary/10">
                    <BarChart3 className="w-6 h-6 text-primary" />
                  </div>
                  <div>
                    <h3 className="font-semibold text-white">{template.name}</h3>
                    <p className="text-sm text-gray-400 mt-1">{template.description}</p>
                    <span className="badge badge-gray mt-3">{template.type}</span>
                  </div>
                </div>
              </div>
              <div className="flex gap-2 mt-4 pt-4 border-t border-gray-700">
                <button className="btn-primary btn-sm flex-1">
                  Generate Now
                </button>
                <button className="btn-secondary btn-sm flex-1">
                  Schedule
                </button>
              </div>
            </motion.div>
          ))}
        </motion.div>
      )}
    </div>
  );
}
