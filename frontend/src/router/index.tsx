import { createBrowserRouter, Navigate } from 'react-router-dom';
import Shell from '../components/Shell';
import Login from '../pages/Login';
import Dashboard from '../pages/Dashboard';
import PackageManage from '../pages/PackageManage';
import RegistrationManage from '../pages/RegistrationManage';
import ResultEntry from '../pages/ResultEntry';
import ReportManage from '../pages/ReportManage';
import AbnormalMetricTrack from '../pages/AbnormalMetricTrack';
import GroupOrderManage from '../pages/GroupOrderManage';
import Profile from '../pages/Profile';
import { RequireAuth } from './guards';

export const router = createBrowserRouter([
  { path: '/login', element: <Login /> },
  {
    path: '/',
    element: (<RequireAuth><Shell /></RequireAuth>),
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: 'dashboard', element: <Dashboard /> },
      { path: 'packages', element: <PackageManage /> },
      { path: 'registrations', element: <RegistrationManage /> },
      { path: 'results', element: <ResultEntry /> },
      { path: 'reports', element: <ReportManage /> },
      { path: 'abnormal-metrics', element: <AbnormalMetricTrack /> },
      { path: 'group-orders', element: <GroupOrderManage /> },
      { path: 'profile', element: <Profile /> },
    ],
  },
  { path: '*', element: <Navigate to="/" replace /> },
]);
