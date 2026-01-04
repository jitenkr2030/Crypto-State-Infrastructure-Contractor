// CSIC Platform - Reports Page
// Report generation and management interface

import React, { useState, useEffect } from 'react';

interface Report {
  id: string;
  title: string;
  type: string;
  status: 'GENERATING' | 'COMPLETED' | 'FAILED';
  createdAt: Date;
  createdBy: string;
  period: string;
  format: string;
  size?: number;
}

const Reports: React.FC = () => {
  const [reports, setReports] = useState<Report[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedReport, setSelectedReport] = useState<Report | null>(null);
  const [showGenerateModal, setShowGenerateModal] = useState(false);
  const [activeTab, setActiveTab] = useState<'generated' | 'templates'>('generated');

  // Form state
  const [reportType, setReportType] = useState('compliance');
  const [reportPeriod, setReportPeriod] = useState('30d');
  const [reportFormat, setReportFormat] = useState('pdf');

  useEffect(() => {
    loadReports();
  }, []);

  const loadReports = async () => {
    setIsLoading(true);
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    const mockReports: Report[] = [
      {
        id: 'rpt_001',
        title: '2024年第四季度交易所合规报告',
        type: 'COMPLIANCE',
        status: 'COMPLETED',
        createdAt: new Date(Date.now() - 86400000),
        createdBy: 'admin',
        period: 'Q4 2024',
        format: 'pdf',
        size: 2456000,
      },
      {
        id: 'rpt_002',
        title: '2024年12月交易监控报告',
        type: 'TRANSACTION',
        status: 'COMPLETED',
        createdAt: new Date(Date.now() - 172800000),
        createdBy: 'regulator_1',
        period: 'Dec 2024',
        format: 'xlsx',
        size: 8540000,
      },
      {
        id: 'rpt_003',
        title: '矿工能源消耗年度报告',
        type: 'ENERGY',
        status: 'GENERATING',
        createdAt: new Date(Date.now() - 3600000),
        createdBy: 'auditor_1',
        period: '2024',
        format: 'pdf',
      },
      {
        id: 'rpt_004',
        title: '钱包风险评估报告',
        type: 'RISK',
        status: 'COMPLETED',
        createdAt: new Date(Date.now() - 259200000),
        createdBy: 'admin',
        period: 'Jan 2025',
        format: 'pdf',
        size: 1234000,
      },
    ];
    
    setReports(mockReports);
    setIsLoading(false);
  };

  const handleGenerateReport = async () => {
    const newReport: Report = {
      id: `rpt_${Date.now()}`,
      title: getReportTitle(reportType, reportPeriod),
      type: reportType.toUpperCase(),
      status: 'GENERATING',
      createdAt: new Date(),
      createdBy: 'current_user',
      period: reportPeriod,
      format: reportFormat,
    };

    setReports([newReport, ...reports]);
    setShowGenerateModal(false);

    // Simulate report generation
    setTimeout(() => {
      setReports(prev => prev.map(r => 
        r.id === newReport.id 
          ? { ...r, status: 'COMPLETED', size: Math.floor(Math.random() * 5000000) + 1000000 }
          : r
      ));
    }, 5000);
  };

  const getReportTitle = (type: string, period: string) => {
    const titles: Record<string, string> = {
      'compliance': '合规报告',
      'transaction': '交易监控报告',
      'energy': '能源消耗报告',
      'risk': '风险评估报告',
      'audit': '审计追踪报告',
    };
    const periodLabels: Record<string, string> = {
      '7d': '周报',
      '30d': '月报',
      '90d': '季报',
      '1y': '年报',
    };
    return `${period} - ${titles[type] || '报告'}`;
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'COMPLETED': return 'badge-success';
      case 'GENERATING': return 'badge-warning';
      case 'FAILED': return 'badge-error';
      default: return 'badge-info';
    }
  };

  const getTypeLabel = (type: string) => {
    const labels: Record<string, string> = {
      'COMPLIANCE': '合规报告',
      'TRANSACTION': '交易报告',
      'ENERGY': '能源报告',
      'RISK': '风险报告',
      'AUDIT': '审计报告',
    };
    return labels[type] || type;
  };

  const formatSize = (bytes?: number) => {
    if (!bytes) return '--';
    if (bytes >= 1000000) return `${(bytes / 1000000).toFixed(2)} MB`;
    if (bytes >= 1000) return `${(bytes / 1000).toFixed(2)} KB`;
    return `${bytes} B`;
  };

  const reportTemplates = [
    {
      id: 'tpl_001',
      name: '交易所合规月报',
      type: 'compliance',
      description: '包含所有注册交易所的合规状态和评分',
      frequency: '每月',
    },
    {
      id: 'tpl_002',
      name: '交易监控日报',
      type: 'transaction',
      description: '每日交易量、异常交易检测汇总',
      frequency: '每日',
    },
    {
      id: 'tpl_003',
      name: '矿工能源消耗报告',
      type: 'energy',
      description: '矿工能源使用效率和合规性分析',
      frequency: '每月',
    },
    {
      id: 'tpl_004',
      name: '钱包风险评估',
      type: 'risk',
      description: '高风险钱包和关联实体分析',
      frequency: '按需',
    },
    {
      id: 'tpl_005',
      name: '审计追踪报告',
      type: 'audit',
      description: '系统操作日志和安全事件记录',
      frequency: '按需',
    },
    {
      id: 'tpl_006',
      name: '季度监管汇总',
      type: 'compliance',
      description: '全面监管活动汇总和趋势分析',
      frequency: '每季度',
    },
  ];

  const stats = {
    total: reports.length,
    completed: reports.filter(r => r.status === 'COMPLETED').length,
    generating: reports.filter(r => r.status === 'GENERATING').length,
    totalSize: reports.reduce((sum, r) => sum + (r.size || 0), 0),
  };

  return (
    <div className="reports-page">
      <div className="page-header">
        <div className="header-left">
          <h1>报告管理</h1>
          <p>生成和管理监管报告</p>
        </div>
        <div className="header-actions">
          <button className="btn btn-primary" onClick={() => setShowGenerateModal(true)}>
            <svg viewBox="0 0 24 24" className="btn-icon">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="8" x2="12" y2="16" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="8" y1="12" x2="16" y2="12" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            生成新报告
          </button>
        </div>
      </div>

      <div className="stats-row">
        <div className="stat-card">
          <div className="stat-icon reports">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="14 2 14 8 20 8" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.total}</span>
            <span className="stat-label">报告总数</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon completed">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="22 4 12 14.01 9 11.01" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.completed}</span>
            <span className="stat-label">已完成</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon generating">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M12 6v6l4 2" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.generating}</span>
            <span className="stat-label">生成中</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon size">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="7 10 12 15 17 10" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="15" x2="12" y2="3" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{formatSize(stats.totalSize)}</span>
            <span className="stat-label">总大小</span>
          </div>
        </div>
      </div>

      <div className="tabs-container">
        <div className="tabs">
          <button 
            className={`tab ${activeTab === 'generated' ? 'active' : ''}`}
            onClick={() => setActiveTab('generated')}
          >
            已生成报告
          </button>
          <button 
            className={`tab ${activeTab === 'templates' ? 'active' : ''}`}
            onClick={() => setActiveTab('templates')}
          >
            报告模板
          </button>
        </div>
      </div>

      {activeTab === 'generated' && (
        <div className="content-card">
          {isLoading ? (
            <div className="loading-state">
              <div className="loading-spinner large"></div>
              <p>加载报告...</p>
            </div>
          ) : (
            <div className="reports-list">
              {reports.map(report => (
                <div 
                  key={report.id} 
                  className={`report-item ${selectedReport?.id === report.id ? 'selected' : ''}`}
                  onClick={() => setSelectedReport(report)}
                >
                  <div className="report-icon">
                    <svg viewBox="0 0 24 24" className="file-icon">
                      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill="none" stroke="currentColor" strokeWidth="2" />
                      <polyline points="14 2 14 8 20 8" fill="none" stroke="currentColor" strokeWidth="2" />
                    </svg>
                  </div>
                  <div className="report-info">
                    <h4 className="report-title">{report.title}</h4>
                    <div className="report-meta">
                      <span className="meta-item">{getTypeLabel(report.type)}</span>
                      <span className="meta-divider">•</span>
                      <span className="meta-item">{report.period}</span>
                      <span className="meta-divider">•</span>
                      <span className="meta-item">{report.format.toUpperCase()}</span>
                    </div>
                  </div>
                  <div className="report-status">
                    <span className={`badge ${getStatusBadge(report.status)}`}>
                      {report.status === 'COMPLETED' ? '已完成' : 
                       report.status === 'GENERATING' ? '生成中' : '失败'}
                    </span>
                    {report.status === 'GENERATING' && (
                      <div className="generating-indicator">
                        <div className="progress-dots">
                          <span></span>
                          <span></span>
                          <span></span>
                        </div>
                      </div>
                    )}
                  </div>
                  <div className="report-size">
                    {formatSize(report.size)}
                  </div>
                  <div className="report-date">
                    {report.createdAt.toLocaleDateString('zh-CN')}
                  </div>
                  <div className="report-actions">
                    {report.status === 'COMPLETED' && (
                      <button className="action-btn" title="下载">
                        <svg viewBox="0 0 24 24" className="action-icon">
                          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" fill="none" stroke="currentColor" strokeWidth="2" />
                          <polyline points="7 10 12 15 17 10" fill="none" stroke="currentColor" strokeWidth="2" />
                          <line x1="12" y1="15" x2="12" y2="3" fill="none" stroke="currentColor" strokeWidth="2" />
                        </svg>
                      </button>
                    )}
                    <button className="action-btn" title="查看">
                      <svg viewBox="0 0 24 24" className="action-icon">
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" fill="none" stroke="currentColor" strokeWidth="2" />
                        <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" strokeWidth="2" />
                      </svg>
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {activeTab === 'templates' && (
        <div className="content-card">
          <div className="templates-grid">
            {reportTemplates.map(template => (
              <div key={template.id} className="template-card">
                <div className="template-header">
                  <svg viewBox="0 0 24 24" className="template-icon">
                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill="none" stroke="currentColor" strokeWidth="2" />
                    <polyline points="14 2 14 8 20 8" fill="none" stroke="currentColor" strokeWidth="2" />
                    <line x1="16" y1="13" x2="8" y2="13" fill="none" stroke="currentColor" strokeWidth="2" />
                    <line x1="16" y1="17" x2="8" y2="17" fill="none" stroke="currentColor" strokeWidth="2" />
                  </svg>
                  <div className="template-info">
                    <h4>{template.name}</h4>
                    <span className="template-frequency">{template.frequency}</span>
                  </div>
                </div>
                <p className="template-description">{template.description}</p>
                <button 
                  className="btn btn-primary btn-sm"
                  onClick={() => {
                    setReportType(template.type);
                    setShowGenerateModal(true);
                  }}
                >
                  使用此模板
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Generate Report Modal */}
      {showGenerateModal && (
        <div className="modal-overlay" onClick={() => setShowGenerateModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">生成新报告</h3>
              <button className="modal-close" onClick={() => setShowGenerateModal(false)}>
                <svg viewBox="0 0 24 24" className="close-icon">
                  <path d="M18 6L6 18M6 6l12 12" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
                </svg>
              </button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label className="form-label">报告类型</label>
                <select 
                  className="form-select" 
                  value={reportType}
                  onChange={(e) => setReportType(e.target.value)}
                >
                  <option value="compliance">合规报告</option>
                  <option value="transaction">交易监控报告</option>
                  <option value="energy">能源消耗报告</option>
                  <option value="risk">风险评估报告</option>
                  <option value="audit">审计追踪报告</option>
                </select>
              </div>
              <div className="form-group">
                <label className="form-label">报告周期</label>
                <select 
                  className="form-select"
                  value={reportPeriod}
                  onChange={(e) => setReportPeriod(e.target.value)}
                >
                  <option value="7d">最近7天</option>
                  <option value="30d">最近30天</option>
                  <option value="90d">最近90天</option>
                  <option value="1y">最近一年</option>
                </select>
              </div>
              <div className="form-group">
                <label className="form-label">输出格式</label>
                <select 
                  className="form-select"
                  value={reportFormat}
                  onChange={(e) => setReportFormat(e.target.value)}
                >
                  <option value="pdf">PDF 文档</option>
                  <option value="xlsx">Excel 电子表格</option>
                  <option value="csv">CSV 数据</option>
                </select>
              </div>
              <div className="form-group">
                <label className="form-label">报告标题</label>
                <input 
                  type="text" 
                  className="form-input" 
                  value={getReportTitle(reportType, reportPeriod)}
                  readOnly
                />
              </div>
            </div>
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setShowGenerateModal(false)}>取消</button>
              <button className="btn btn-primary" onClick={handleGenerateReport}>
                开始生成
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Reports;
