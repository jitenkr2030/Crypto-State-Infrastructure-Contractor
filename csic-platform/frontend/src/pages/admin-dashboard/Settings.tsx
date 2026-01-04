// CSIC Platform - Settings Page
// System configuration and settings management interface

import React, { useState } from 'react';
import { useAuthStore } from '../../store';

const Settings: React.FC = () => {
  const { user, hasPermission } = useAuthStore();
  const [activeSection, setActiveSection] = useState('general');
  const [saved, setSaved] = useState(false);

  // Settings state
  const [settings, setSettings] = useState({
    // General
    systemName: '加密货币国家基础设施承包商',
    timezone: 'Asia/Shanghai',
    language: 'zh-CN',
    
    // Security
    sessionTimeout: 30,
    maxLoginAttempts: 5,
    mfaRequired: true,
    passwordExpiry: 90,
    
    // Monitoring
    alertCooldown: 15,
    autoFlagThreshold: 80,
    monitoringInterval: 5,
    
    // Notifications
    emailNotifications: true,
    slackWebhook: '',
    alertSeverity: 'WARNING',
    
    // Compliance
    complianceCheckInterval: 24,
    autoAuditReminders: true,
    reportRetention: 365,
  });

  const handleSave = () => {
    // Simulate saving
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  const handleInputChange = (field: string, value: string | number | boolean) => {
    setSettings(prev => ({ ...prev, [field]: value }));
    setSaved(false);
  };

  const sections = [
    { id: 'general', name: '常规设置', icon: 'settings' },
    { id: 'security', name: '安全设置', icon: 'security' },
    { id: 'monitoring', name: '监控设置', icon: 'monitor' },
    { id: 'notifications', name: '通知设置', icon: 'bell' },
    { id: 'compliance', name: '合规设置', icon: 'compliance' },
    { id: 'api', name: 'API 配置', icon: 'api' },
  ];

  const renderSection = () => {
    switch (activeSection) {
      case 'general':
        return (
          <div className="settings-section">
            <h3>常规设置</h3>
            <div className="form-group">
              <label className="form-label">系统名称</label>
              <input
                type="text"
                className="form-input"
                value={settings.systemName}
                onChange={(e) => handleInputChange('systemName', e.target.value)}
              />
            </div>
            <div className="form-group">
              <label className="form-label">时区</label>
              <select
                className="form-select"
                value={settings.timezone}
                onChange={(e) => handleInputChange('timezone', e.target.value)}
              >
                <option value="Asia/Shanghai">亚洲/上海 (UTC+8)</option>
                <option value="UTC">UTC (UTC+0)</option>
                <option value="America/New_York">美国/纽约 (UTC-5)</option>
                <option value="Europe/London">欧洲/伦敦 (UTC+0)</option>
              </select>
            </div>
            <div className="form-group">
              <label className="form-label">界面语言</label>
              <select
                className="form-select"
                value={settings.language}
                onChange={(e) => handleInputChange('language', e.target.value)}
              >
                <option value="zh-CN">简体中文</option>
                <option value="en-US">English (US)</option>
              </select>
            </div>
          </div>
        );

      case 'security':
        return (
          <div className="settings-section">
            <h3>安全设置</h3>
            <div className="form-group">
              <label className="form-label">会话超时（分钟）</label>
              <input
                type="number"
                className="form-input"
                value={settings.sessionTimeout}
                onChange={(e) => handleInputChange('sessionTimeout', parseInt(e.target.value))}
              />
              <span className="form-hint">用户空闲多长时间后自动登出</span>
            </div>
            <div className="form-group">
              <label className="form-label">最大登录尝试次数</label>
              <input
                type="number"
                className="form-input"
                value={settings.maxLoginAttempts}
                onChange={(e) => handleInputChange('maxLoginAttempts', parseInt(e.target.value))}
              />
            </div>
            <div className="form-group">
              <label className="form-label">密码过期天数</label>
              <input
                type="number"
                className="form-input"
                value={settings.passwordExpiry}
                onChange={(e) => handleInputChange('passwordExpiry', parseInt(e.target.value))}
              />
            </div>
            <div className="form-group">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={settings.mfaRequired}
                  onChange={(e) => handleInputChange('mfaRequired', e.target.checked)}
                />
                <span className="checkbox-text">强制多因素认证</span>
              </label>
              <span className="form-hint">所有用户登录时必须使用 MFA</span>
            </div>
          </div>
        );

      case 'monitoring':
        return (
          <div className="settings-section">
            <h3>监控设置</h3>
            <div className="form-group">
              <label className="form-label">警报冷却时间（分钟）</label>
              <input
                type="number"
                className="form-input"
                value={settings.alertCooldown}
                onChange={(e) => handleInputChange('alertCooldown', parseInt(e.target.value))}
              />
              <span className="form-hint">相同类型警报之间的最小间隔</span>
            </div>
            <div className="form-group">
              <label className="form-label">自动标记阈值</label>
              <input
                type="number"
                className="form-input"
                value={settings.autoFlagThreshold}
                onChange={(e) => handleInputChange('autoFlagThreshold', parseInt(e.target.value))}
              />
              <span className="form-hint">风险评分达到此值时自动标记交易 (0-100)</span>
            </div>
            <div className="form-group">
              <label className="form-label">监控间隔（秒）</label>
              <input
                type="number"
                className="form-input"
                value={settings.monitoringInterval}
                onChange={(e) => handleInputChange('monitoringInterval', parseInt(e.target.value))}
              />
              <span className="form-hint">数据采集和监控检查的频率</span>
            </div>
          </div>
        );

      case 'notifications':
        return (
          <div className="settings-section">
            <h3>通知设置</h3>
            <div className="form-group">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={settings.emailNotifications}
                  onChange={(e) => handleInputChange('emailNotifications', e.target.checked)}
                />
                <span className="checkbox-text">启用邮件通知</span>
              </label>
            </div>
            <div className="form-group">
              <label className="form-label">Slack Webhook URL</label>
              <input
                type="url"
                className="form-input"
                value={settings.slackWebhook}
                onChange={(e) => handleInputChange('slackWebhook', e.target.value)}
                placeholder="https://hooks.slack.com/services/..."
              />
            </div>
            <div className="form-group">
              <label className="form-label">最低警报级别</label>
              <select
                className="form-select"
                value={settings.alertSeverity}
                onChange={(e) => handleInputChange('alertSeverity', e.target.value)}
              >
                <option value="INFO">信息</option>
                <option value="WARNING">警告</option>
                <option value="CRITICAL">严重</option>
              </select>
            </div>
          </div>
        );

      case 'compliance':
        return (
          <div className="settings-section">
            <h3>合规设置</h3>
            <div className="form-group">
              <label className="form-label">合规检查间隔（小时）</label>
              <input
                type="number"
                className="form-input"
                value={settings.complianceCheckInterval}
                onChange={(e) => handleInputChange('complianceCheckInterval', parseInt(e.target.value))}
              />
            </div>
            <div className="form-group">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={settings.autoAuditReminders}
                  onChange={(e) => handleInputChange('autoAuditReminders', e.target.checked)}
                />
                <span className="checkbox-text">自动发送审计提醒</span>
              </label>
            </div>
            <div className="form-group">
              <label className="form-label">报告保留天数</label>
              <input
                type="number"
                className="form-input"
                value={settings.reportRetention}
                onChange={(e) => handleInputChange('reportRetention', parseInt(e.target.value))}
              />
              <span className="form-hint">自动删除超过此天数的报告</span>
            </div>
          </div>
        );

      case 'api':
        return (
          <div className="settings-section">
            <h3>API 配置</h3>
            <div className="info-card">
              <div className="info-header">
                <svg viewBox="0 0 24 24" className="info-icon">
                  <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
                  <path d="M12 16v-4M12 8h.01" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
                </svg>
                <h4>API 访问信息</h4>
              </div>
              <div className="info-content">
                <div className="info-row">
                  <span className="info-label">API 端点</span>
                  <code className="info-value">https://api.csic.gov.local/api/v1</code>
                </div>
                <div className="info-row">
                  <span className="info-label">API 版本</span>
                  <span className="info-value">v1.0.0</span>
                </div>
                <div className="info-row">
                  <span className="info-label">速率限制</span>
                  <span className="info-value">1000 请求/分钟</span>
                </div>
              </div>
            </div>
            <div className="form-group">
              <label className="form-label">API 密钥</label>
              <div className="api-key-field">
                <input
                  type="password"
                  className="form-input"
                  value="••••••••••••••••••••"
                  readOnly
                />
                <button className="btn btn-secondary btn-sm">重新生成</button>
              </div>
              <span className="form-hint warning">重新生成将立即使当前密钥失效</span>
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="settings-page">
      <div className="page-header">
        <div className="header-left">
          <h1>系统设置</h1>
          <p>配置和管理系统参数</p>
        </div>
        <div className="header-actions">
          {saved && (
            <span className="save-status success">
              <svg viewBox="0 0 24 24" className="status-icon">
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" fill="none" stroke="currentColor" strokeWidth="2" />
                <polyline points="22 4 12 14.01 9 11.01" fill="none" stroke="currentColor" strokeWidth="2" />
              </svg>
              已保存
            </span>
          )}
          <button className="btn btn-primary" onClick={handleSave}>
            <svg viewBox="0 0 24 24" className="btn-icon">
              <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="17 21 17 13 7 13 7 21" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="7 3 7 8 15 8" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            保存更改
          </button>
        </div>
      </div>

      <div className="settings-layout">
        <div className="settings-sidebar">
          <nav className="settings-nav">
            {sections.map(section => (
              <button
                key={section.id}
                className={`nav-item ${activeSection === section.id ? 'active' : ''}`}
                onClick={() => setActiveSection(section.id)}
              >
                <span className="nav-icon">
                  {section.icon === 'settings' && (
                    <svg viewBox="0 0 24 24">
                      <circle cx="12" cy="12" r="3" fill="none" stroke="currentColor" strokeWidth="2" />
                      <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" fill="none" stroke="currentColor" strokeWidth="2" />
                    </svg>
                  )}
                  {section.icon === 'security' && (
                    <svg viewBox="0 0 24 24">
                      <rect x="3" y="11" width="18" height="11" rx="2" ry="2" fill="none" stroke="currentColor" strokeWidth="2" />
                      <path d="M7 11V7a5 5 0 0 1 10 0v4" fill="none" stroke="currentColor" strokeWidth="2" />
                    </svg>
                  )}
                  {section.icon === 'monitor' && (
                    <svg viewBox="0 0 24 24">
                      <rect x="2" y="3" width="20" height="14" rx="2" ry="2" fill="none" stroke="currentColor" strokeWidth="2" />
                      <line x1="8" y1="21" x2="16" y2="21" fill="none" stroke="currentColor" strokeWidth="2" />
                      <line x1="12" y1="17" x2="12" y2="21" fill="none" stroke="currentColor" strokeWidth="2" />
                    </svg>
                  )}
                  {section.icon === 'bell' && (
                    <svg viewBox="0 0 24 24">
                      <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" fill="none" stroke="currentColor" strokeWidth="2" />
                      <path d="M13.73 21a2 2 0 0 1-3.46 0" fill="none" stroke="currentColor" strokeWidth="2" />
                    </svg>
                  )}
                  {section.icon === 'compliance' && (
                    <svg viewBox="0 0 24 24">
                      <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" fill="none" stroke="currentColor" strokeWidth="2" />
                      <polyline points="22 4 12 14.01 9 11.01" fill="none" stroke="currentColor" strokeWidth="2" />
                    </svg>
                  )}
                  {section.icon === 'api' && (
                    <svg viewBox="0 0 24 24">
                      <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" fill="none" stroke="currentColor" strokeWidth="2" />
                    </svg>
                  )}
                </span>
                <span className="nav-text">{section.name}</span>
              </button>
            ))}
          </nav>
        </div>

        <div className="settings-content">
          <div className="settings-card">
            {renderSection()}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Settings;
