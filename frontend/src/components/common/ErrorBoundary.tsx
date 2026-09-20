import { Component, type ErrorInfo, type ReactNode } from 'react';
import { Alert } from 'antd';

interface Props { children: ReactNode }
interface State { hasError: boolean; message: string }

export default class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, message: '' };
  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, message: error.message };
  }
  componentDidCatch(error: Error, info: ErrorInfo): void {
    console.error('ErrorBoundary caught', error, info);
  }
  render() {
    if (this.state.hasError) {
      return <Alert type="error" showIcon message="页面渲染出错" description={this.state.message} />;
    }
    return this.props.children;
  }
}
