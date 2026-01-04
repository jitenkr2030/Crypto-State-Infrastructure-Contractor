// Exchange Detail View - Detailed exchange monitoring page
// Component for viewing detailed information about a specific cryptocurrency exchange

import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useSystemStore, useAlertStore } from '../../store';
import { api } from '../../services/api';
import styles from './ExchangeDetail.module.css';

// Types for exchange data
interface ExchangeDetailData {
  id: string;
  name: string;
  registrationNumber: string;
  jurisdiction: string;
  website: string;
  status: 'operational' | 'suspended' | 'revoked' | 'pending';
  licenseType: string;
  foundingDate: string;
  headquarterAddress: string;
  contactEmail: string;
  phoneNumber: string;
  complianceOfficer: {
    name: string;
    email: string;
    phone: string;
  };
  tradingVolume24h: number;
  tradingVolumeChange24h: number;
  registeredUsers: number;
  activeUsers24h: number;
  listedAssets: number;
  tradingPairs: number;
  dailyWithdrawalLimit: number;
  kycCompliantUsers: number;
  amlRiskRating: 'low' | 'medium' | 'high';
  lastInspectionDate: string;
  nextInspectionDate: string;
  regulatoryFilings: RegulatoryFiling[];
  enforcementActions: EnforcementAction[];
}

interface RegulatoryFiling {
  id: string;
  type: string;
  filingDate: string;
  status: 'pending' | 'approved' | 'rejected';
  description: string;
}

interface EnforcementAction {
  id: string;
  date: string;
  type: string;
  description: string;
  status: 'open' | 'resolved' | 'appealed';
}

const ExchangeDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { status } = useSystemStore();
  const { addAlert } = useAlertStore();
  const [exchange, setExchange] = useState<ExchangeDetailData | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'overview' | 'compliance' | 'filings' | 'enforcement'>('overview');

  useEffect(() => {
    const fetchExchangeDetails = async () => {
      if (!id) {
        navigate('/exchanges');
        return;
      }

      try {
        setLoading(true);
        // Simulated API call - replace with actual API call
        await new Promise(resolve => setTimeout(resolve, 500));
        
        // Mock data for demonstration
        const mockExchange: ExchangeDetailData = {
          id: id,
          name: 'CryptoTrade Pro',
          registrationNumber: 'CRYPTO-2024-78901',
          jurisdiction: 'United States',
          website: 'https://cryptotradepro.com',
          status: 'operational',
          licenseType: 'MTL (Money Transmitter License)',
          foundingDate: '2019-03-15',
          headquarterAddress: '450 Financial District, New York, NY 10005',
          contactEmail: 'compliance@cryptotradepro.com',
          phoneNumber: '+1-212-555-0142',
          complianceOfficer: {
            name: 'Sarah Mitchell',
            email: 'sarah.mitchell@cryptotradepro.com',
            phone: '+1-212-555-0143'
          },
          tradingVolume24h: 4523000000,
          tradingVolumeChange24h: 12.5,
          registeredUsers: 2450000,
          activeUsers24h: 125000,
          listedAssets: 342,
          tradingPairs: 520,
          dailyWithdrawalLimit: 10000,
          kycCompliantUsers: 98.5,
          amlRiskRating: 'low',
          lastInspectionDate: '2024-09-15',
          nextInspectionDate: '2025-03-15',
          regulatoryFilings: [
            {
              id: 'RF-001',
              type: 'Annual Audit Report',
              filingDate: '2024-01-15',
              status: 'approved',
              description: 'Annual comprehensive audit report submitted and approved'
            },
            {
              id: 'RF-002',
              type: 'AML/KYC Compliance Report',
              filingDate: '2024-07-01',
              status: 'approved',
              description: 'Quarterly AML/KYC compliance review'
            },
            {
              id: 'RF-003',
              type: 'Capital Reserve Certification',
              filingDate: '2024-10-01',
              status: 'pending',
              description: 'Capital reserve certification for Q4 2024'
            }
          ],
          enforcementActions: []
        };

        setExchange(mockExchange);
      } catch (error) {
        console.error('Failed to fetch exchange details:', error);
        addAlert({
          type: 'error',
          title: '数据加载失败',
          message: '无法加载交易所详细信息'
        });
      } finally {
        setLoading(false);
      }
    };

    fetchExchangeDetails();
  }, [id, navigate, addAlert]);

  const formatCurrency = (value: number): string => {
    if (value >= 1000000000) {
      return `$${(value / 1000000000).toFixed(2)}B`;
    } else if (value >= 1000000) {
      return `$${(value / 1000000).toFixed(2)}M`;
    } else if (value >= 1000) {
      return `$${(value / 1000).toFixed(2)}K`;
    }
    return `$${value.toFixed(2)}`;
  };

  const formatNumber = (value: number): string => {
    return new Intl.NumberFormat('en-US').format(value);
  };

  const getStatusBadgeClass = (status: string): string => {
    switch (status) {
      case 'operational':
        return styles.statusOperational;
      case 'suspended':
        return styles.statusSuspended;
      case 'revoked':
        return styles.statusRevoked;
      case 'pending':
        return styles.statusPending;
      default:
        return '';
    }
  };

  if (loading) {
    return (
      <div className={styles.loadingContainer}>
        <div className={styles.loadingSpinner}></div>
        <p>正在加载交易所详细信息...</p>
      </div>
    );
  }

  if (!exchange) {
    return (
      <div className={styles.errorContainer}>
        <h2>交易所未找到</h2>
        <p>无法找到指定的交易所信息。</p>
        <button onClick={() => navigate('/exchanges')} className={styles.backButton}>
          返回交易所列表
        </button>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      {/* Header Section */}
      <div className={styles.header}>
        <div className={styles.headerContent}>
          <button onClick={() => navigate('/exchanges')} className={styles.backButton}>
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path d="M19 12H5M12 19l-7-7 7-7" stroke="currentColor" strokeWidth="2" fill="none" />
            </svg>
            返回
          </button>
          <div className={styles.exchangeTitle}>
            <h1>{exchange.name}</h1>
            <span className={`${styles.statusBadge} ${getStatusBadgeClass(exchange.status)}`}>
              {exchange.status === 'operational' ? '运营中' : 
               exchange.status === 'suspended' ? '已暂停' : 
               exchange.status === 'revoked' ? '已吊销' : '待审核'}
            </span>
          </div>
          <div className={styles.exchangeMeta}>
            <span className={styles.metaItem}>
              <svg viewBox="0 0 24 24" width="16" height="16">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z" fill="none" stroke="currentColor" strokeWidth="2"/>
                <path d="M12 6v6l4 2" fill="none" stroke="currentColor" strokeWidth="2"/>
              </svg>
              注册号: {exchange.registrationNumber}
            </span>
            <span className={styles.metaItem}>
              <svg viewBox="0 0 24 24" width="16" height="16">
                <path d="M12 22s-8-4.5-8-11.8A8 8 0 0 1 12 2a8 8 0 0 1 8 8.2c0 7.3-8 11.8-8 11.8z" fill="none" stroke="currentColor" strokeWidth="2"/>
                <circle cx="12" cy="10" r="3" fill="none" stroke="currentColor" strokeWidth="2"/>
              </svg>
              {exchange.jurisdiction}
            </span>
          </div>
        </div>
        <div className={styles.headerActions}>
          <button className={styles.actionButton}>
            <svg viewBox="0 0 24 24" width="18" height="18">
              <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" fill="none" stroke="currentColor" strokeWidth="2"/>
              <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" fill="none" stroke="currentColor" strokeWidth="2"/>
            </svg>
            编辑信息
          </button>
          <button className={styles.actionButton}>
            <svg viewBox="0 0 24 24" width="18" height="18">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill="none" stroke="currentColor" strokeWidth="2"/>
              <path d="M14 2v6h6M16 13H8M16 17H8M10 9H8" fill="none" stroke="currentColor" strokeWidth="2"/>
            </svg>
            生成报告
          </button>
        </div>
      </div>

      {/* Quick Stats */}
      <div className={styles.quickStats}>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>
            <svg viewBox="0 0 24 24" width="24" height="24">
              <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" fill="none" stroke="currentColor" strokeWidth="2"/>
            </svg>
          </div>
          <div className={styles.statInfo}>
            <span className={styles.statLabel}>24h 交易量</span>
            <span className={styles.statValue}>{formatCurrency(exchange.tradingVolume24h)}</span>
            <span className={`${styles.statChange} ${exchange.tradingVolumeChange24h >= 0 ? styles.positive : styles.negative}`}>
              {exchange.tradingVolumeChange24h >= 0 ? '+' : ''}{exchange.tradingVolumeChange24h.toFixed(2)}%
            </span>
          </div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>
            <svg viewBox="0 0 24 24" width="24" height="24">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" fill="none" stroke="currentColor" strokeWidth="2"/>
              <circle cx="9" cy="7" r="4" fill="none" stroke="currentColor" strokeWidth="2"/>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" fill="none" stroke="currentColor" strokeWidth="2"/>
            </svg>
          </div>
          <div className={styles.statInfo}>
            <span className={styles.statLabel}>注册用户</span>
            <span className={styles.statValue}>{formatNumber(exchange.registeredUsers)}</span>
            <span className={styles.statSubtext}>{formatNumber(exchange.activeUsers24h)} 活跃 (24h)</span>
          </div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>
            <svg viewBox="0 0 24 24" width="24" height="24">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" fill="none" stroke="currentColor" strokeWidth="2"/>
              <path d="M3.27 6.96L12 12.01l8.73-5.05M12 22.08V12" fill="none" stroke="currentColor" strokeWidth="2"/>
            </svg>
          </div>
          <div className={styles.statInfo}>
            <span className={styles.statLabel}>上架资产</span>
            <span className={styles.statValue}>{formatNumber(exchange.listedAssets)}</span>
            <span className={styles.statSubtext}>{formatNumber(exchange.tradingPairs)} 交易对</span>
          </div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>
            <svg viewBox="0 0 24 24" width="24" height="24">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" fill="none" stroke="currentColor" strokeWidth="2"/>
              <path d="M9 12l2 2 4-4" fill="none" stroke="currentColor" strokeWidth="2"/>
            </svg>
          </div>
          <div className={styles.statInfo}>
            <span className={styles.statLabel}>KYC 认证率</span>
            <span className={styles.statValue}>{exchange.kycCompliantUsers}%</span>
            <span className={`${styles.riskRating} ${exchange.amlRiskRating === 'low' ? styles.riskLow : exchange.amlRiskRating === 'medium' ? styles.riskMedium : styles.riskHigh}`}>
              AML 风险: {exchange.amlRiskRating === 'low' ? '低' : exchange.amlRiskRating === 'medium' ? '中' : '高'}
            </span>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className={styles.tabs}>
        <button 
          className={`${styles.tab} ${activeTab === 'overview' ? styles.activeTab : ''}`}
          onClick={() => setActiveTab('overview')}
        >
          概览
        </button>
        <button 
          className={`${styles.tab} ${activeTab === 'compliance' ? styles.activeTab : ''}`}
          onClick={() => setActiveTab('compliance')}
        >
          合规信息
        </button>
        <button 
          className={`${styles.tab} ${activeTab === 'filings' ? styles.activeTab : ''}`}
          onClick={() => setActiveTab('filings')}
        >
          监管文件
        </button>
        <button 
          className={`${styles.tab} ${activeTab === 'enforcement' ? styles.activeTab : ''}`}
          onClick={() => setActiveTab('enforcement')}
        >
          执法行动
        </button>
      </div>

      {/* Tab Content */}
      <div className={styles.tabContent}>
        {activeTab === 'overview' && (
          <div className={styles.overviewGrid}>
            {/* Basic Information */}
            <div className={styles.infoCard}>
              <h3>基本信息</h3>
              <div className={styles.infoGrid}>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>官方网站</span>
                  <a href={exchange.website} target="_blank" rel="noopener noreferrer" className={styles.infoLink}>
                    {exchange.website}
                  </a>
                </div>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>许可证类型</span>
                  <span className={styles.infoValue}>{exchange.licenseType}</span>
                </div>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>成立日期</span>
                  <span className={styles.infoValue}>{exchange.foundingDate}</span>
                </div>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>总部地址</span>
                  <span className={styles.infoValue}>{exchange.headquarterAddress}</span>
                </div>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>日提现限额</span>
                  <span className={styles.infoValue}>{formatCurrency(exchange.dailyWithdrawalLimit)}</span>
                </div>
              </div>
            </div>

            {/* Contact Information */}
            <div className={styles.infoCard}>
              <h3>联系信息</h3>
              <div className={styles.infoGrid}>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>联系邮箱</span>
                  <span className={styles.infoValue}>{exchange.contactEmail}</span>
                </div>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>联系电话</span>
                  <span className={styles.infoValue}>{exchange.phoneNumber}</span>
                </div>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>合规负责人</span>
                  <span className={styles.infoValue}>{exchange.complianceOfficer.name}</span>
                </div>
                <div className={styles.infoItem}>
                  <span className={styles.infoLabel}>负责人邮箱</span>
                  <span className={styles.infoValue}>{exchange.complianceOfficer.email}</span>
                </div>
              </div>
            </div>

            {/* Inspection Schedule */}
            <div className={styles.infoCard}>
              <h3>检查安排</h3>
              <div className={styles.inspectionSchedule}>
                <div className={styles.inspectionItem}>
                  <div className={styles.inspectionIcon}>
                    <svg viewBox="0 0 24 24" width="20" height="20">
                      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill="none" stroke="currentColor" strokeWidth="2"/>
                      <path d="M14 2v6h6M16 13H8M16 17H8M10 9H8" fill="none" stroke="currentColor" strokeWidth="2"/>
                    </svg>
                  </div>
                  <div className={styles.inspectionInfo}>
                    <span className={styles.inspectionLabel}>上次检查</span>
                    <span className={styles.inspectionDate}>{exchange.lastInspectionDate}</span>
                  </div>
                </div>
                <div className={styles.inspectionItem}>
                  <div className={styles.inspectionIcon}>
                    <svg viewBox="0 0 24 24" width="20" height="20">
                      <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" fill="none" stroke="currentColor" strokeWidth="2"/>
                    </svg>
                  </div>
                  <div className={styles.inspectionInfo}>
                    <span className={styles.inspectionLabel}>下次检查</span>
                    <span className={`${styles.inspectionDate} ${styles.upcoming}`}>{exchange.nextInspectionDate}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'compliance' && (
          <div className={styles.complianceContent}>
            <div className={styles.complianceGrid}>
              <div className={styles.complianceCard}>
                <h4>KYC 合规性</h4>
                <div className={styles.complianceProgress}>
                  <div className={styles.progressBar}>
                    <div 
                      className={styles.progressFill} 
                      style={{ width: `${exchange.kycCompliantUsers}%` }}
                    ></div>
                  </div>
                  <span className={styles.progressLabel}>{exchange.kycCompliantUsers}%</span>
                </div>
                <p className={styles.complianceDesc}>
                  {formatNumber(exchange.registeredUsers)} 名注册用户中，{formatNumber(Math.floor(exchange.registeredUsers * exchange.kycCompliantUsers / 100))} 名已完成 KYC 认证
                </p>
              </div>
              <div className={styles.complianceCard}>
                <h4>AML 风险评级</h4>
                <div className={`${styles.riskIndicator} ${styles[`risk${exchange.amlRiskRating.charAt(0).toUpperCase() + exchange.amlRiskRating.slice(1)}`]}`}>
                  <svg viewBox="0 0 24 24" width="32" height="32">
                    <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" fill="none" stroke="currentColor" strokeWidth="2"/>
                    <path d="M9 12l2 2 4-4" fill="none" stroke="currentColor" strokeWidth="2"/>
                  </svg>
                  <span>{exchange.amlRiskRating === 'low' ? '低风险' : exchange.amlRiskRating === 'medium' ? '中等风险' : '高风险'}</span>
                </div>
                <p className={styles.complianceDesc}>
                  基于交易模式、地理位置和用户行为的综合风险评估
                </p>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'filings' && (
          <div className={styles.filingsContent}>
            <div className={styles.filingsHeader}>
              <h3>监管文件列表</h3>
              <button className={styles.addButton}>
                <svg viewBox="0 0 24 24" width="18" height="18">
                  <path d="M12 5v14M5 12h14" stroke="currentColor" strokeWidth="2" fill="none"/>
                </svg>
                新增文件
              </button>
            </div>
            <div className={styles.filingsTable}>
              <div className={styles.tableHeader}>
                <span>文件编号</span>
                <span>文件类型</span>
                <span>提交日期</span>
                <span>状态</span>
                <span>描述</span>
                <span>操作</span>
              </div>
              {exchange.regulatoryFilings.map((filing) => (
                <div key={filing.id} className={styles.tableRow}>
                  <span className={styles.filingId}>{filing.id}</span>
                  <span className={styles.filingType}>{filing.type}</span>
                  <span className={styles.filingDate}>{filing.filingDate}</span>
                  <span className={`${styles.filingStatus} ${styles[`status${filing.status.charAt(0).toUpperCase() + filing.status.slice(1)}`]}`}>
                    {filing.status === 'approved' ? '已批准' : filing.status === 'pending' ? '待审核' : '已拒绝'}
                  </span>
                  <span className={styles.filingDesc}>{filing.description}</span>
                  <span className={styles.filingActions}>
                    <button className={styles.viewButton}>查看</button>
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'enforcement' && (
          <div className={styles.enforcementContent}>
            {exchange.enforcementActions.length === 0 ? (
              <div className={styles.emptyState}>
                <svg viewBox="0 0 24 24" width="48" height="48">
                  <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" fill="none" stroke="currentColor" strokeWidth="2"/>
                  <path d="M9 12l2 2 4-4" fill="none" stroke="currentColor" strokeWidth="2"/>
                </svg>
                <h4>暂无执法行动</h4>
                <p>该交易所目前没有正在进行的执法行动记录。</p>
              </div>
            ) : (
              <div className={styles.enforcementList}>
                {exchange.enforcementActions.map((action) => (
                  <div key={action.id} className={styles.enforcementCard}>
                    <div className={styles.enforcementHeader}>
                      <span className={styles.enforcementType}>{action.type}</span>
                      <span className={`${styles.enforcementStatus} ${styles[`enforcement${action.status.charAt(0).toUpperCase() + action.status.slice(1)}`]}`}>
                        {action.status === 'open' ? '进行中' : action.status === 'resolved' ? '已解决' : '上诉中'}
                      </span>
                    </div>
                    <p className={styles.enforcementDesc}>{action.description}</p>
                    <span className={styles.enforcementDate}>{action.date}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default ExchangeDetail;
