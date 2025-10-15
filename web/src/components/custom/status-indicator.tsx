'use client';

import * as React from "react";

interface StatusIndicatorProps {
  status: string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export const StatusIndicator: React.FC<StatusIndicatorProps> = ({ 
  status, 
  size = 'sm', 
  className = '' 
}) => {
  const sizeClasses = {
    sm: 'w-2 h-2',
    md: 'w-3 h-3',
    lg: 'w-4 h-4'
  };

  // Random Tailwind colors for unmatched statuses
  const randomColors = [
    'bg-blue-500', 'bg-purple-500', 'bg-pink-500', 'bg-indigo-500',
    'bg-cyan-500', 'bg-teal-500', 'bg-lime-500', 'bg-orange-500',
    'bg-rose-500', 'bg-violet-500', 'bg-sky-500', 'bg-emerald-500'
  ];

  const getStatusColor = (status: string): string => {
    const normalizedStatus = status.toLowerCase();
    
    switch (normalizedStatus) {
      case 'development':
        return 'bg-amber-700';
      case 'staging':
        return 'bg-gray-500';
      case 'production':
        return 'bg-green-500';
      case 'online':
        return 'bg-green-500';
      case 'warning':
        return 'bg-yellow-500';
      case 'offline':
        return 'bg-red-500';
      default:
        const hash = status.split('').reduce((a, b) => {
          a = ((a << 5) - a) + b.charCodeAt(0);
          return a & a;
        }, 0);
        const index = Math.abs(hash) % randomColors.length;
        return randomColors[index];
    }
  };

  return (
    <div 
      className={`rounded-full ${sizeClasses[size]} ${getStatusColor(status)} ${className}`}
    />
  );
};