import React from 'react';

interface LoadingSpinnerProps {
  size?: 'small' | 'medium' | 'large';
  fullScreen?: boolean;
  message?: string;
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  size = 'medium',
  fullScreen = false,
  message,
}) => {
  const spinner = (
    <div className={`loading-spinner ${size}`}>
      <div className="spinner-ring"></div>
      <div className="spinner-ring"></div>
      <div className="spinner-ring"></div>
    </div>
  );

  if (fullScreen) {
    return (
      <div className="loading-overlay">
        <div className="loading-content">
          {spinner}
          {message && <p className="loading-message">{message}</p>}
        </div>
      </div>
    );
  }

  return (
    <div className="loading-container">
      {spinner}
      {message && <p className="loading-message">{message}</p>}
    </div>
  );
};

export default LoadingSpinner;
