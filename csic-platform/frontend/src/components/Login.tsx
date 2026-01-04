import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store';
import toast from 'react-hot-toast';

const Login: React.FC = () => {
  const navigate = useNavigate();
  const { login, isLoading } = useAuthStore();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [mfaCode, setMfaCode] = useState('');
  const [showMfa, setShowMfa] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!username || !password) {
      setError('请输入用户名和密码');
      return;
    }

    try {
      await login(username, password);
      toast.success('登录成功');
      navigate('/');
    } catch (err: any) {
      setError(err.message || '登录失败，请检查凭据');
    }
  };

  const handleDemoLogin = async (role: string) => {
    setError('');
    try {
      await login(role, 'demo');
      toast.success(`以${role}角色登录成功`);
      navigate('/');
    } catch (err: any) {
      setError(err.message || '登录失败');
    }
  };

  return (
    <div className="login-page">
      <div className="login-container">
        <div className="login-header">
          <div className="login-logo">
            <ShieldIcon />
          </div>
          <h1>CSIC Platform</h1>
          <p>Crypto State Infrastructure Contractor</p>
        </div>

        <form className="login-form" onSubmit={handleSubmit}>
          {error && <div className="login-error">{error}</div>}

          <div className="form-group">
            <label htmlFor="username">用户名</label>
            <input
              type="text"
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="请输入用户名"
              disabled={isLoading}
              autoComplete="username"
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">密码</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="请输入密码"
              disabled={isLoading}
              autoComplete="current-password"
            />
          </div>

          {showMfa && (
            <div className="form-group">
              <label htmlFor="mfa">多因素认证码</label>
              <input
                type="text"
                id="mfa"
                value={mfaCode}
                onChange={(e) => setMfaCode(e.target.value)}
                placeholder="请输入认证码"
                disabled={isLoading}
                maxLength={6}
              />
            </div>
          )}

          <button type="submit" className="login-btn" disabled={isLoading}>
            {isLoading ? (
              <span className="loading-spinner small"></span>
            ) : (
              '登录'
            )}
          </button>

          <button
            type="button"
            className="mfa-btn"
            onClick={() => setShowMfa(!showMfa)}
          >
            {showMfa ? '跳过MFA' : '使用MFA'}
          </button>
        </form>

        <div className="login-divider">
          <span>或使用演示账户</span>
        </div>

        <div className="demo-accounts">
          <button
            type="button"
            className="demo-btn admin"
            onClick={() => handleDemoLogin('admin')}
            disabled={isLoading}
          >
            <span className="demo-role">管理员</span>
            <span className="demo-desc">完全访问权限</span>
          </button>
          <button
            type="button"
            className="demo-btn operator"
            onClick={() => handleDemoLogin('operator')}
            disabled={isLoading}
          >
            <span className="demo-role">操作员</span>
            <span className="demo-desc">日常操作权限</span>
          </button>
          <button
            type="button"
            className="demo-btn auditor"
            onClick={() => handleDemoLogin('auditor')}
            disabled={isLoading}
          >
            <span className="demo-role">审计员</span>
            <span className="demo-desc">只读访问权限</span>
          </button>
        </div>

        <div className="login-footer">
          <p>安全警告: 此系统受政府法规保护</p>
          <p>所有操作都将被记录和审计</p>
        </div>
      </div>

      <div className="login-info">
        <h2>政府级加密货币监管平台</h2>
        <ul>
          <li>交易所实时监控</li>
          <li>交易风险评估</li>
          <li>钱包治理控制</li>
          <li>挖矿活动监管</li>
          <li>能源消耗追踪</li>
          <li>合规报告生成</li>
        </ul>
        <div className="security-badges">
          <span className="badge">HSM加密</span>
          <span className="badge">WORM存储</span>
          <span className="badge">审计追踪</span>
          <span className="badge">国产化</span>
        </div>
      </div>
    </div>
  );
};

const ShieldIcon: React.FC = () => (
  <svg width="48" height="48" viewBox="0 0 24 24" fill="currentColor">
    <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z" />
  </svg>
);

export default Login;
