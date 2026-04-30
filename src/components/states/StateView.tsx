import React from 'react';
import { AppStatus } from '../../types';
import { LoadingState } from './LoadingState';
import { EmptyState } from './EmptyState';
import { ErrorState } from './ErrorState';
import { ServerDownState } from './ServerDownState';
import { SuccessState } from './SuccessState';

interface Props {
  status: AppStatus;
  // loading
  loadingMessage?: string;
  // empty
  emptyTitle?: string;
  emptyDescription?: string;
  // error
  errorTitle?: string;
  errorMessage?: string;
  // success
  successTitle?: string;
  successMessage?: string;
  successAction?: { label: string; onPress: () => void };
  // shared
  onRetry?: () => void;
  fullScreen?: boolean;
  children: React.ReactNode;
}

export function StateView({
  status,
  loadingMessage,
  emptyTitle = 'Nothing here yet',
  emptyDescription,
  errorTitle,
  errorMessage,
  successTitle = 'Done!',
  successMessage,
  successAction,
  onRetry,
  fullScreen = true,
  children,
}: Props) {
  switch (status) {
    case 'loading':
      return <LoadingState message={loadingMessage} fullScreen={fullScreen} />;
    case 'empty':
      return (
        <EmptyState
          title={emptyTitle}
          description={emptyDescription}
          fullScreen={fullScreen}
          action={onRetry ? { label: 'Refresh', onPress: onRetry } : undefined}
        />
      );
    case 'error':
      return (
        <ErrorState
          title={errorTitle}
          message={errorMessage}
          onRetry={onRetry}
          fullScreen={fullScreen}
        />
      );
    case 'server_down':
      return <ServerDownState onRetry={onRetry} fullScreen={fullScreen} />;
    case 'success':
      return (
        <SuccessState
          title={successTitle}
          message={successMessage}
          action={successAction}
          fullScreen={fullScreen}
        />
      );
    default:
      return <>{children}</>;
  }
}
