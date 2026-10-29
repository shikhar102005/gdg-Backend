import React, { Component, ErrorInfo, ReactNode } from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';

interface Props {
  children: ReactNode;
  fallbackTitle?: string;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
  }

  private handleReset = () => {
    this.setState({ hasError: false, error: null });
    window.location.reload();
  };

  public render() {
    if (this.state.hasError) {
      return (
        <div style={{
          minHeight: '400px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '2rem',
        }}>
          <div style={{
            background: '#fef08a',
            border: '3px solid #000',
            borderRadius: '16px',
            boxShadow: '6px 6px 0px #000',
            padding: '2rem',
            maxWidth: '520px',
            width: '100%',
            textAlign: 'center',
          }}>
            <div style={{
              display: 'inline-flex',
              padding: '0.75rem',
              borderRadius: '9999px',
              background: '#fca5a5',
              border: '2px solid #000',
              boxShadow: '2px 2px 0px #000',
              marginBottom: '1rem',
            }}>
              <AlertTriangle size={32} color="#000" />
            </div>
            <h2 style={{ fontSize: '1.25rem', fontWeight: 800, marginBottom: '0.5rem', color: '#000' }}>
              {this.props.fallbackTitle || 'Component Encountered an Issue'}
            </h2>
            <p style={{ fontSize: '0.85rem', color: '#334155', marginBottom: '1.5rem', fontFamily: 'monospace', wordBreak: 'break-word' }}>
              {this.state.error?.message || 'An unexpected rendering error occurred.'}
            </p>
            <button
              onClick={this.handleReset}
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '0.5rem',
                background: '#86efac',
                border: '2px solid #000',
                borderRadius: '9999px',
                padding: '0.6rem 1.5rem',
                fontSize: '0.9rem',
                fontWeight: 700,
                cursor: 'pointer',
                boxShadow: '3px 3px 0px #000',
              }}
            >
              <RefreshCw size={16} /> Recover & Reload
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
