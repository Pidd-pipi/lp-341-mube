import { Card, Col, Row, Spin, Statistic } from 'antd';
import { GiftOutlined, FileAddOutlined, FileTextOutlined, WarningOutlined, MoneyCollectOutlined } from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import { useReportStats } from '../hooks/useReportStats';

export default function Dashboard() {
  const { stats, loading } = useReportStats();

  const packagePie = {
    tooltip: { trigger: 'item' },
    legend: { bottom: 0 },
    series: [{ type: 'pie', radius: ['35%', '65%'], data: (stats?.package_sold ?? []).map((r) => ({ name: r.name, value: r.count })) }],
  };
  const deptBar = {
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: (stats?.dept_workload ?? []).map((r) => r.name) },
    yAxis: { type: 'value' },
    series: [{ type: 'bar', data: (stats?.dept_workload ?? []).map((r) => r.count), itemStyle: { color: '#1677ff' } }],
  };
  const abnormalBar = {
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: (stats?.abnormal_top ?? []).map((r) => r.name) },
    yAxis: { type: 'value' },
    series: [{ type: 'bar', data: (stats?.abnormal_top ?? []).map((r) => r.count), itemStyle: { color: '#fa541c' } }],
  };

  return (
    <Spin spinning={loading}>
      <Row gutter={16}>
        <Col span={4}><Card><Statistic title="套餐数" value={stats?.package_count ?? 0} prefix={<GiftOutlined />} /></Card></Col>
        <Col span={4}><Card><Statistic title="登记数" value={stats?.registration_count ?? 0} prefix={<FileAddOutlined />} /></Card></Col>
        <Col span={4}><Card><Statistic title="报告数" value={stats?.report_count ?? 0} prefix={<FileTextOutlined />} /></Card></Col>
        <Col span={4}><Card><Statistic title="异常指标" value={stats?.abnormal_count ?? 0} prefix={<WarningOutlined />} valueStyle={{ color: '#fa541c' }} /></Card></Col>
        <Col span={8}><Card><Statistic title="累计收入（元）" value={stats?.revenue ?? 0} precision={2} prefix={<MoneyCollectOutlined />} valueStyle={{ color: '#3f8600' }} /></Card></Col>
      </Row>
      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={8}><Card title="各套餐成交量" size="small"><ReactECharts option={packagePie} style={{ height: 280 }} /></Card></Col>
        <Col span={8}><Card title="科室工作量" size="small"><ReactECharts option={deptBar} style={{ height: 280 }} /></Card></Col>
        <Col span={8}><Card title="异常检出率 TOP" size="small"><ReactECharts option={abnormalBar} style={{ height: 280 }} /></Card></Col>
      </Row>
      <Card title="月度收入趋势" size="small" style={{ marginTop: 16 }}>
        <ReactECharts option={{
          tooltip: { trigger: 'axis' },
          xAxis: { type: 'category', data: (stats?.monthly_revenue ?? []).map((r) => r.month) },
          yAxis: { type: 'value' },
          series: [{ type: 'line', smooth: true, data: (stats?.monthly_revenue ?? []).map((r) => r.amount), areaStyle: {} }],
        }} style={{ height: 280 }} />
      </Card>
    </Spin>
  );
}
