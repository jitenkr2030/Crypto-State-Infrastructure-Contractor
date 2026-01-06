import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { motion } from 'framer-motion';
import {
  Settings,
  User,
  Bell,
  Shield,
  Globe,
  Database,
  Key,
  Save,
  RefreshCw,
  Eye,
  EyeOff,
} from 'lucide-react';

interface GeneralSettings {
  theme: 'dark' | 'light' | 'system';
  language: string;
  timezone: string;
  dateFormat: string;
  timeFormat: string;
}

interface SecuritySettings {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
  mfaEnabled: boolean;
  sessionTimeout: number;
}

interface NotificationSettings {
  emailEnabled: boolean;
  slackEnabled: boolean;
  alertThreshold: string;
  digestFrequency: string;
}

interface IntegrationSettings {
  auditLogEndpoint: string;
  controlLayerEndpoint: string;
  healthMonitorEndpoint: string;
  blockchainIndexerEndpoint: string;
  apiKey: string;
}

const containerVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0 },
};

const languages = [
  { value: 'en', label: 'English' },
  { value: 'zh', label: 'Chinese' },
  { value: 'es', label: 'Spanish' },
  { value: 'fr', label: 'French' },
];

const timezones = [
  { value: 'UTC', label: 'UTC' },
  { value: 'America/New_York', label: 'Eastern Time' },
  { value: 'America/Los_Angeles', label: 'Pacific Time' },
  { value: 'Europe/London', label: 'London' },
  { value: 'Asia/Shanghai', label: 'Shanghai' },
];

export default function Settings() {
  const [activeTab, setActiveTab] = useState('general');
  const [showApiKey, setShowApiKey] = useState(false);

  const {
    register: registerGeneral,
    handleSubmit: handleSubmitGeneral,
    formState: { errors: generalErrors },
  } = useForm<GeneralSettings>({
    defaultValues: {
      theme: 'dark',
      language: 'en',
      timezone: 'UTC',
      dateFormat: 'YYYY-MM-DD',
      timeFormat: '24h',
    },
  });

  const {
    register: registerSecurity,
    handleSubmit: handleSubmitSecurity,
    formState: { errors: securityErrors },
  } = useForm<SecuritySettings>();

  const {
    register: registerNotifications,
    handleSubmit: handleSubmitNotifications,
    formState: { errors: notificationErrors },
  } = useForm<NotificationSettings>({
    defaultValues: {
      emailEnabled: true,
      slackEnabled: false,
      alertThreshold: 'HIGH',
      digestFrequency: 'daily',
    },
  });

  const {
    register: registerIntegrations,
    handleSubmit: handleSubmitIntegrations,
    formState: { errors: integrationErrors },
  } = useForm<IntegrationSettings>({
    defaultValues: {
      auditLogEndpoint: 'http://localhost:8080',
      controlLayerEndpoint: 'http://localhost:8081',
      healthMonitorEndpoint: 'http://localhost:8083',
      blockchainIndexerEndpoint: 'http://localhost:8087',
      apiKey: 'sk_live_xxxxxxxxxxxxxxxxxxxxxxxx',
    },
  });

  const onSubmitGeneral = (data: GeneralSettings) => {
    console.log('General settings:', data);
  };

  const onSubmitSecurity = (data: SecuritySettings) => {
    console.log('Security settings:', data);
  };

  const onSubmitNotifications = (data: NotificationSettings) => {
    console.log('Notification settings:', data);
  };

  const onSubmitIntegrations = (data: IntegrationSettings) => {
    console.log('Integration settings:', data);
  };

  const tabs = [
    { id: 'general', label: 'General', icon: Globe },
    { id: 'security', label: 'Security', icon: Shield },
    { id: 'notifications', label: 'Notifications', icon: Bell },
    { id: 'integrations', label: 'Integrations', icon: Database },
  ];

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="page-header">Settings</h1>
        <p className="page-description">
          Manage your account and platform configuration
        </p>
      </div>

      <div className="flex flex-col lg:flex-row gap-8">
        {/* Sidebar */}
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="lg:w-64 flex-shrink-0"
        >
          <nav className="card p-2 space-y-1">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg transition-all ${
                  activeTab === tab.id
                    ? 'bg-primary/10 text-primary'
                    : 'text-gray-400 hover:text-white hover:bg-background-light'
                }`}
              >
                <tab.icon className="w-5 h-5" />
                <span className="font-medium">{tab.label}</span>
              </button>
            ))}
          </nav>
        </motion.div>

        {/* Content */}
        <motion.div
          variants={containerVariants}
          initial="hidden"
          animate="visible"
          className="flex-1"
        >
          {/* General Settings */}
          {activeTab === 'general' && (
            <form
              onSubmit={handleSubmitGeneral(onSubmitGeneral)}
              className="card"
            >
              <div className="card-header">
                <h3 className="section-title mb-0">General Settings</h3>
              </div>
              <div className="card-body space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label className="label">Theme</label>
                    <select
                      {...registerGeneral('theme')}
                      className="input"
                    >
                      <option value="dark">Dark</option>
                      <option value="light">Light</option>
                      <option value="system">System</option>
                    </select>
                  </div>
                  <div>
                    <label className="label">Language</label>
                    <select
                      {...registerGeneral('language')}
                      className="input"
                    >
                      {languages.map((lang) => (
                        <option key={lang.value} value={lang.value}>
                          {lang.label}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="label">Timezone</label>
                    <select
                      {...registerGeneral('timezone')}
                      className="input"
                    >
                      {timezones.map((tz) => (
                        <option key={tz.value} value={tz.value}>
                          {tz.label}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="label">Date Format</label>
                    <select
                      {...registerGeneral('dateFormat')}
                      className="input"
                    >
                      <option value="YYYY-MM-DD">YYYY-MM-DD</option>
                      <option value="MM/DD/YYYY">MM/DD/YYYY</option>
                      <option value="DD/MM/YYYY">DD/MM/YYYY</option>
                    </select>
                  </div>
                </div>
                <div className="flex justify-end pt-4 border-t border-gray-700">
                  <button type="submit" className="btn-primary">
                    <Save className="w-4 h-4 mr-2" />
                    Save Changes
                  </button>
                </div>
              </div>
            </form>
          )}

          {/* Security Settings */}
          {activeTab === 'security' && (
            <form
              onSubmit={handleSubmitSecurity(onSubmitSecurity)}
              className="card"
            >
              <div className="card-header">
                <h3 className="section-title mb-0">Security Settings</h3>
              </div>
              <div className="card-body space-y-6">
                <div>
                  <h4 className="text-sm font-medium text-white mb-4">Change Password</h4>
                  <div className="space-y-4">
                    <div>
                      <label className="label">Current Password</label>
                      <input
                        type="password"
                        {...registerSecurity('currentPassword', {
                          required: 'Current password is required',
                        })}
                        className={`input ${securityErrors.currentPassword ? 'input-error' : ''}`}
                      />
                      {securityErrors.currentPassword && (
                        <p className="mt-1 text-sm text-danger">
                          {securityErrors.currentPassword.message}
                        </p>
                      )}
                    </div>
                    <div>
                      <label className="label">New Password</label>
                      <input
                        type="password"
                        {...registerSecurity('newPassword', {
                          required: 'New password is required',
                          minLength: {
                            value: 8,
                            message: 'Password must be at least 8 characters',
                          },
                        })}
                        className={`input ${securityErrors.newPassword ? 'input-error' : ''}`}
                      />
                      {securityErrors.newPassword && (
                        <p className="mt-1 text-sm text-danger">
                          {securityErrors.newPassword.message}
                        </p>
                      )}
                    </div>
                    <div>
                      <label className="label">Confirm New Password</label>
                      <input
                        type="password"
                        {...registerSecurity('confirmPassword', {
                          required: 'Please confirm your password',
                        })}
                        className={`input ${securityErrors.confirmPassword ? 'input-error' : ''}`}
                      />
                      {securityErrors.confirmPassword && (
                        <p className="mt-1 text-sm text-danger">
                          {securityErrors.confirmPassword.message}
                        </p>
                      )}
                    </div>
                  </div>
                </div>

                <div className="pt-6 border-t border-gray-700">
                  <h4 className="text-sm font-medium text-white mb-4">Two-Factor Authentication</h4>
                  <div className="flex items-center justify-between p-4 bg-background-lighter rounded-lg">
                    <div className="flex items-center gap-3">
                      <Shield className="w-5 h-5 text-primary" />
                      <div>
                        <p className="font-medium text-white">2FA Authentication</p>
                        <p className="text-sm text-gray-400">
                          Add an extra layer of security to your account
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        {...registerSecurity('mfaEnabled')}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-gray-600 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
                    </label>
                  </div>
                </div>

                <div className="flex justify-end pt-4 border-t border-gray-700">
                  <button type="submit" className="btn-primary">
                    <Save className="w-4 h-4 mr-2" />
                    Save Changes
                  </button>
                </div>
              </div>
            </form>
          )}

          {/* Notification Settings */}
          {activeTab === 'notifications' && (
            <form
              onSubmit={handleSubmitNotifications(onSubmitNotifications)}
              className="card"
            >
              <div className="card-header">
                <h3 className="section-title mb-0">Notification Settings</h3>
              </div>
              <div className="card-body space-y-6">
                <div className="space-y-4">
                  <div className="flex items-center justify-between p-4 bg-background-lighter rounded-lg">
                    <div className="flex items-center gap-3">
                      <User className="w-5 h-5 text-primary" />
                      <div>
                        <p className="font-medium text-white">Email Notifications</p>
                        <p className="text-sm text-gray-400">
                          Receive notifications via email
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        {...registerNotifications('emailEnabled')}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-gray-600 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
                    </label>
                  </div>

                  <div className="flex items-center justify-between p-4 bg-background-lighter rounded-lg">
                    <div className="flex items-center gap-3">
                      <Bell className="w-5 h-5 text-warning" />
                      <div>
                        <p className="font-medium text-white">Slack Notifications</p>
                        <p className="text-sm text-gray-400">
                          Receive notifications in Slack
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        {...registerNotifications('slackEnabled')}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-gray-600 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
                    </label>
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div>
                    <label className="label">Alert Threshold</label>
                    <select
                      {...registerNotifications('alertThreshold')}
                      className="input"
                    >
                      <option value="LOW">Low and above</option>
                      <option value="MEDIUM">Medium and above</option>
                      <option value="HIGH">High only</option>
                      <option value="CRITICAL">Critical only</option>
                    </select>
                  </div>
                  <div>
                    <label className="label">Digest Frequency</label>
                    <select
                      {...registerNotifications('digestFrequency')}
                      className="input"
                    >
                      <option value="realtime">Real-time</option>
                      <option value="hourly">Hourly</option>
                      <option value="daily">Daily</option>
                      <option value="weekly">Weekly</option>
                    </select>
                  </div>
                </div>

                <div className="flex justify-end pt-4 border-t border-gray-700">
                  <button type="submit" className="btn-primary">
                    <Save className="w-4 h-4 mr-2" />
                    Save Changes
                  </button>
                </div>
              </div>
            </form>
          )}

          {/* Integration Settings */}
          {activeTab === 'integrations' && (
            <form
              onSubmit={handleSubmitIntegrations(onSubmitIntegrations)}
              className="card"
            >
              <div className="card-header">
                <h3 className="section-title mb-0">Integration Settings</h3>
              </div>
              <div className="card-body space-y-6">
                <div className="grid grid-cols-1 gap-6">
                  <div>
                    <label className="label">Audit Log Service Endpoint</label>
                    <input
                      type="url"
                      {...registerIntegrations('auditLogEndpoint', {
                        required: 'Endpoint is required',
                      })}
                      className={`input ${integrationErrors.auditLogEndpoint ? 'input-error' : ''}`}
                      placeholder="http://localhost:8080"
                    />
                  </div>
                  <div>
                    <label className="label">Control Layer Endpoint</label>
                    <input
                      type="url"
                      {...registerIntegrations('controlLayerEndpoint', {
                        required: 'Endpoint is required',
                      })}
                      className={`input ${integrationErrors.controlLayerEndpoint ? 'input-error' : ''}`}
                      placeholder="http://localhost:8081"
                    />
                  </div>
                  <div>
                    <label className="label">Health Monitor Endpoint</label>
                    <input
                      type="url"
                      {...registerIntegrations('healthMonitorEndpoint', {
                        required: 'Endpoint is required',
                      })}
                      className={`input ${integrationErrors.healthMonitorEndpoint ? 'input-error' : ''}`}
                      placeholder="http://localhost:8083"
                    />
                  </div>
                  <div>
                    <label className="label">Blockchain Indexer Endpoint</label>
                    <input
                      type="url"
                      {...registerIntegrations('blockchainIndexerEndpoint', {
                        required: 'Endpoint is required',
                      })}
                      className={`input ${integrationErrors.blockchainIndexerEndpoint ? 'input-error' : ''}`}
                      placeholder="http://localhost:8087"
                    />
                  </div>
                  <div>
                    <label className="label">API Key</label>
                    <div className="relative">
                      <input
                        type={showApiKey ? 'text' : 'password'}
                        {...registerIntegrations('apiKey', {
                          required: 'API key is required',
                        })}
                        className={`input pr-10 ${integrationErrors.apiKey ? 'input-error' : ''}`}
                        placeholder="sk_live_xxxxxxxxxxxxxxxxxxxxxxxx"
                      />
                      <button
                        type="button"
                        onClick={() => setShowApiKey(!showApiKey)}
                        className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-white"
                      >
                        {showApiKey ? (
                          <EyeOff className="w-5 h-5" />
                        ) : (
                          <Eye className="w-5 h-5" />
                        )}
                      </button>
                    </div>
                  </div>
                </div>

                <div className="flex justify-between pt-4 border-t border-gray-700">
                  <button type="button" className="btn-secondary">
                    <RefreshCw className="w-4 h-4 mr-2" />
                    Test Connections
                  </button>
                  <button type="submit" className="btn-primary">
                    <Save className="w-4 h-4 mr-2" />
                    Save Changes
                  </button>
                </div>
              </div>
            </form>
          )}
        </motion.div>
      </div>
    </div>
  );
}
