// CSIC Platform - Users Management Page
// User administration and access control interface

import React, { useState, useEffect } from 'react';
import { useAuthStore } from '../../store';

interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  status: 'ACTIVE' | 'INACTIVE' | 'SUSPENDED';
  lastLogin: Date;
  mfaEnabled: boolean;
  createdAt: Date;
}

const Users: React.FC = () => {
  const { hasPermission } = useAuthStore();
  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [roleFilter, setRoleFilter] = useState('all');
  const [statusFilter, setStatusFilter] = useState('all');
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    setIsLoading(true);
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    const mockUsers: User[] = [
      {
        id: 'usr_001',
        username: 'admin',
        email: 'admin@csic.gov',
        role: 'ADMIN',
        status: 'ACTIVE',
        lastLogin: new Date(Date.now() - 3600000),
        mfaEnabled: true,
        createdAt: new Date('2024-01-01'),
      },
      {
        id: 'usr_002',
        username: 'regulator_1',
        email: 'regulator1@csic.gov',
        role: 'REGULATOR',
        status: 'ACTIVE',
        lastLogin: new Date(Date.now() - 86400000),
        mfaEnabled: true,
        createdAt: new Date('2024-02-15'),
      },
      {
        id: 'usr_003',
        username: 'regulator_2',
        email: 'regulator2@csic.gov',
        role: 'REGULATOR',
        status: 'ACTIVE',
        lastLogin: new Date(Date.now() - 172800000),
        mfaEnabled: false,
        createdAt: new Date('2024-03-01'),
      },
      {
        id: 'usr_004',
        username: 'auditor_1',
        email: 'auditor1@csic.gov',
        role: 'AUDITOR',
        status: 'ACTIVE',
        lastLogin: new Date(Date.now() - 259200000),
        mfaEnabled: true,
        createdAt: new Date('2024-02-20'),
      },
      {
        id: 'usr_005',
        username: 'analyst_1',
        email: 'analyst1@csic.gov',
        role: 'ANALYST',
        status: 'SUSPENDED',
        lastLogin: new Date(Date.now() - 604800000),
        mfaEnabled: true,
        createdAt: new Date('2024-03-15'),
      },
    ];
    
    setUsers(mockUsers);
    setIsLoading(false);
  };

  const filteredUsers = users.filter(user => {
    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      if (!user.username.toLowerCase().includes(query) &&
          !user.email.toLowerCase().includes(query)) {
        return false;
      }
    }

    // Role filter
    if (roleFilter !== 'all' && user.role !== roleFilter) {
      return false;
    }

    // Status filter
    if (statusFilter !== 'all' && user.status !== statusFilter) {
      return false;
    }

    return true;
  });

  const getRoleBadge = (role: string) => {
    switch (role) {
      case 'ADMIN': return 'role-admin';
      case 'REGULATOR': return 'role-regulator';
      case 'AUDITOR': return 'role-auditor';
      case 'ANALYST': return 'role-analyst';
      default: return 'role-default';
    }
  };

  const getRoleLabel = (role: string) => {
    const labels: Record<string, string> = {
      'ADMIN': '管理员',
      'REGULATOR': '监管员',
      'AUDITOR': '审计员',
      'ANALYST': '分析师',
    };
    return labels[role] || role;
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'ACTIVE': return 'badge-success';
      case 'INACTIVE': return 'badge-info';
      case 'SUSPENDED': return 'badge-warning';
      default: return 'badge-info';
    }
  };

  const getStatusLabel = (status: string) => {
    const labels: Record<string, string> = {
      'ACTIVE': '活跃',
      'INACTIVE': '非活跃',
      'SUSPENDED': '已暂停',
    };
    return labels[status] || status;
  };

  const handleToggleStatus = (user: User) => {
    const newStatus = user.status === 'ACTIVE' ? 'SUSPENDED' : 'ACTIVE';
    setUsers(prev => prev.map(u => 
      u.id === user.id ? { ...u, status: newStatus } : u
    ));
  };

  const handleDeleteUser = (userId: string) => {
    if (window.confirm('确定要删除该用户吗？此操作不可撤销。')) {
      setUsers(prev => prev.filter(u => u.id !== userId));
    }
  };

  const stats = {
    total: users.length,
    active: users.filter(u => u.status === 'ACTIVE').length,
    admins: users.filter(u => u.role === 'ADMIN').length,
    mfaEnabled: users.filter(u => u.mfaEnabled).length,
  };

  return (
    <div className="users-page">
      <div className="page-header">
        <div className="header-left">
          <h1>用户管理</h1>
          <p>管理系统用户和访问权限</p>
        </div>
        <div className="header-actions">
          <button className="btn btn-secondary">
            <svg viewBox="0 0 24 24" className="btn-icon">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="7 10 12 15 17 10" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="15" x2="12" y2="3" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            导出用户
          </button>
          <button className="btn btn-primary" onClick={() => setShowAddModal(true)}>
            <svg viewBox="0 0 24 24" className="btn-icon">
              <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="12" y1="8" x2="12" y2="16" fill="none" stroke="currentColor" strokeWidth="2" />
              <line x1="8" y1="12" x2="16" y2="12" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
            添加用户
          </button>
        </div>
      </div>

      <div className="stats-row">
        <div className="stat-card">
          <div className="stat-icon users">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" fill="none" stroke="currentColor" strokeWidth="2" />
              <circle cx="9" cy="7" r="4" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M23 21v-2a4 4 0 0 0-3-3.87" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M16 3.13a4 4 0 0 1 0 7.75" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.total}</span>
            <span className="stat-label">用户总数</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon active">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" fill="none" stroke="currentColor" strokeWidth="2" />
              <polyline points="22 4 12 14.01 9 11.01" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.active}</span>
            <span className="stat-label">活跃用户</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon admins">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <path d="M12 2L2 7l10 5 10-5-10-5z" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M2 17l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M2 12l10 5 10-5" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.admins}</span>
            <span className="stat-label">管理员</span>
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-icon mfa">
            <svg viewBox="0 0 24 24" className="icon-svg">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2" fill="none" stroke="currentColor" strokeWidth="2" />
              <path d="M7 11V7a5 5 0 0 1 10 0v4" fill="none" stroke="currentColor" strokeWidth="2" />
            </svg>
          </div>
          <div className="stat-content">
            <span className="stat-value">{stats.mfaEnabled}</span>
            <span className="stat-label">MFA 已启用</span>
          </div>
        </div>
      </div>

      <div className="content-card">
        <div className="card-header">
          <div className="filters">
            <div className="search-box">
              <svg viewBox="0 0 24 24" className="search-icon">
                <circle cx="11" cy="11" r="8" fill="none" stroke="currentColor" strokeWidth="2" />
                <path d="M21 21l-4.35-4.35" fill="none" stroke="currentColor" strokeWidth="2" />
              </svg>
              <input
                type="text"
                placeholder="搜索用户..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="search-input"
              />
            </div>
            <select
              className="filter-select"
              value={roleFilter}
              onChange={(e) => setRoleFilter(e.target.value)}
            >
              <option value="all">所有角色</option>
              <option value="ADMIN">管理员</option>
              <option value="REGULATOR">监管员</option>
              <option value="AUDITOR">审计员</option>
              <option value="ANALYST">分析师</option>
            </select>
            <select
              className="filter-select"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
            >
              <option value="all">所有状态</option>
              <option value="ACTIVE">活跃</option>
              <option value="INACTIVE">非活跃</option>
              <option value="SUSPENDED">已暂停</option>
            </select>
          </div>
          <span className="result-count">{filteredUsers.length} 个结果</span>
        </div>

        {isLoading ? (
          <div className="loading-state">
            <div className="loading-spinner large"></div>
            <p>加载用户数据...</p>
          </div>
        ) : (
          <div className="users-table-container">
            <table className="table users-table">
              <thead>
                <tr>
                  <th>用户名</th>
                  <th>邮箱</th>
                  <th>角色</th>
                  <th>MFA</th>
                  <th>状态</th>
                  <th>最后登录</th>
                  <th>创建时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {filteredUsers.map(user => (
                  <tr 
                    key={user.id} 
                    className={`user-row ${selectedUser?.id === user.id ? 'selected' : ''}`}
                    onClick={() => setSelectedUser(user)}
                  >
                    <td>
                      <div className="user-cell">
                        <div className="user-avatar">
                          {user.username.charAt(0).toUpperCase()}
                        </div>
                        <span className="username">{user.username}</span>
                      </div>
                    </td>
                    <td>
                      <span className="email">{user.email}</span>
                    </td>
                    <td>
                      <span className={`role-badge ${getRoleBadge(user.role)}`}>
                        {getRoleLabel(user.role)}
                      </span>
                    </td>
                    <td>
                      {user.mfaEnabled ? (
                        <span className="mfa-enabled">
                          <svg viewBox="0 0 24 24" className="check-icon">
                            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" fill="none" stroke="currentColor" strokeWidth="2" />
                            <polyline points="22 4 12 14.01 9 11.01" fill="none" stroke="currentColor" strokeWidth="2" />
                          </svg>
                          已启用
                        </span>
                      ) : (
                        <span className="mfa-disabled">未启用</span>
                      )}
                    </td>
                    <td>
                      <span className={`badge ${getStatusBadge(user.status)}`}>
                        {getStatusLabel(user.status)}
                      </span>
                    </td>
                    <td>
                      {user.lastLogin.toLocaleString('zh-CN')}
                    </td>
                    <td>
                      {user.createdAt.toLocaleDateString('zh-CN')}
                    </td>
                    <td>
                      <div className="action-buttons" onClick={(e) => e.stopPropagation()}>
                        <button 
                          className="action-btn" 
                          title="编辑"
                          onClick={() => setSelectedUser(user)}
                        >
                          <svg viewBox="0 0 24 24" className="action-icon">
                            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" fill="none" stroke="currentColor" strokeWidth="2" />
                            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" fill="none" stroke="currentColor" strokeWidth="2" />
                          </svg>
                        </button>
                        <button 
                          className={`action-btn ${user.status === 'ACTIVE' ? 'warning' : 'success'}`}
                          title={user.status === 'ACTIVE' ? '暂停用户' : '激活用户'}
                          onClick={() => handleToggleStatus(user)}
                        >
                          {user.status === 'ACTIVE' ? (
                            <svg viewBox="0 0 24 24" className="action-icon">
                              <rect x="6" y="4" width="4" height="16" fill="none" stroke="currentColor" strokeWidth="2" />
                              <rect x="14" y="4" width="4" height="16" fill="none" stroke="currentColor" strokeWidth="2" />
                            </svg>
                          ) : (
                            <svg viewBox="0 0 24 24" className="action-icon">
                              <polygon points="5 3 19 12 5 21 5 3" fill="none" stroke="currentColor" strokeWidth="2" />
                            </svg>
                          )}
                        </button>
                        <button 
                          className="action-btn danger" 
                          title="删除用户"
                          onClick={() => handleDeleteUser(user.id)}
                        >
                          <svg viewBox="0 0 24 24" className="action-icon">
                            <polyline points="3 6 5 6 21 6" fill="none" stroke="currentColor" strokeWidth="2" />
                            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" fill="none" stroke="currentColor" strokeWidth="2" />
                          </svg>
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Add User Modal */}
      {showAddModal && (
        <div className="modal-overlay" onClick={() => setShowAddModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">添加新用户</h3>
              <button className="modal-close" onClick={() => setShowAddModal(false)}>
                <svg viewBox="0 0 24 24" className="close-icon">
                  <path d="M18 6L6 18M6 6l12 12" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
                </svg>
              </button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label className="form-label">用户名 *</label>
                <input type="text" className="form-input" placeholder="输入用户名" />
              </div>
              <div className="form-group">
                <label className="form-label">邮箱 *</label>
                <input type="email" className="form-input" placeholder="输入邮箱" />
              </div>
              <div className="form-group">
                <label className="form-label">角色 *</label>
                <select className="form-select">
                  <option value="">选择角色</option>
                  <option value="REGULATOR">监管员</option>
                  <option value="AUDITOR">审计员</option>
                  <option value="ANALYST">分析师</option>
                </select>
              </div>
              <div className="form-group">
                <label className="form-label">初始密码 *</label>
                <input type="password" className="form-input" placeholder="输入初始密码" />
              </div>
              <div className="form-group">
                <label className="checkbox-label">
                  <input type="checkbox" />
                  <span className="checkbox-text">发送欢迎邮件</span>
                </label>
              </div>
            </div>
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setShowAddModal(false)}>取消</button>
              <button className="btn btn-primary">创建用户</button>
            </div>
          </div>
        </div>
      )}

      {/* User Detail Panel */}
      {selectedUser && !showAddModal && (
        <div className="detail-panel">
          <div className="detail-header">
            <h3>用户详情</h3>
            <button className="close-btn" onClick={() => setSelectedUser(null)}>
              <svg viewBox="0 0 24 24" className="close-icon">
                <path d="M18 6L6 18M6 6l12 12" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
              </svg>
            </button>
          </div>
          <div className="detail-content">
            <div className="user-profile">
              <div className="profile-avatar large">
                {selectedUser.username.charAt(0).toUpperCase()}
              </div>
              <h4>{selectedUser.username}</h4>
              <span className={`role-badge ${getRoleBadge(selectedUser.role)}`}>
                {getRoleLabel(selectedUser.role)}
              </span>
            </div>
            <div className="detail-section">
              <h4>账户信息</h4>
              <div className="detail-grid">
                <div className="detail-item">
                  <span className="item-label">用户ID</span>
                  <span className="item-value">{selectedUser.id}</span>
                </div>
                <div className="detail-item">
                  <span className="item-label">邮箱</span>
                  <span className="item-value">{selectedUser.email}</span>
                </div>
                <div className="detail-item">
                  <span className="item-label">状态</span>
                  <span className={`badge ${getStatusBadge(selectedUser.status)}`}>
                    {getStatusLabel(selectedUser.status)}
                  </span>
                </div>
                <div className="detail-item">
                  <span className="item-label">MFA</span>
                  <span className="item-value">{selectedUser.mfaEnabled ? '已启用' : '未启用'}</span>
                </div>
              </div>
            </div>
            <div className="detail-section">
              <h4>活动记录</h4>
              <div className="detail-grid">
                <div className="detail-item">
                  <span className="item-label">最后登录</span>
                  <span className="item-value">{selectedUser.lastLogin.toLocaleString('zh-CN')}</span>
                </div>
                <div className="detail-item">
                  <span className="item-label">创建时间</span>
                  <span className="item-value">{selectedUser.createdAt.toLocaleString('zh-CN')}</span>
                </div>
              </div>
            </div>
            <div className="detail-actions">
              <button className="btn btn-secondary">重置密码</button>
              <button className="btn btn-primary">保存更改</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Users;
