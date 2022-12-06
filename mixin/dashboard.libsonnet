local grafana = import 'grafonnet/grafana.libsonnet';
local graphPanel = grafana.graphPanel;

local utils = import 'snmp-mixin/lib/utils.libsonnet';

local matcher = 'job=~"$job", instance=~"$instance"';

local queries = {
  duration_by_day: 'sum(increase(play_seconds_total{' + matcher + '}[24h]))',
  duration_by_hour: 'sum(increase(play_seconds_total{' + matcher + '}[1h]))',
  duration_by_month: 'sum(increase(play_seconds_total{' + matcher + '}[30d]))',
  duration_by_user: '',

  count: 'plays_total',
  count_by_day: '',
  count_by_hour: '',
  count_by_month: '',
  count_by_user: '',

  top_ten_plays_by_user: '',
  top_ten_duration_by_user: '',
  top_ten_plays_by_media_type: '',
};

local ds_template = {
  current: {
    text: 'default',
    value: 'default',
  },
  hide: 0,
  label: 'Data Source',
  name: 'datasource',
  options: [],
  query: 'prometheus',
  refresh: 1,
  regex: '',
  type: 'datasource',
};

local job_template =
  grafana.template.new(
    'job',
    '$datasource',
    'label_values(awx_system_info, job)',
    label='job',
    refresh='load',
    multi=true,
    includeAll=true,
    allValues='.+',
    sort=1,
  );

local instance_template =
  grafana.template.new(
    'instance',
    '$datasource',
    'label_values(awx_system_info{job=~"$job"}, instance)',
    label='instance',
    refresh='load',
    multi=true,
    includeAll=true,
    allValues='.+',
    sort=1,
  );

local durationPanel =
  graphPanel.new(
    'Duration',
    datasource='$datasource',
  )
  .addTarget(grafana.prometheus.target(queries.duration_by_day, interval='1d')) + utils.timeSeriesOverride(
    unit='s',
    fillOpacity=10,
    showPoints='never',
  ) { span: 12 };

local playback_dashboard =
  grafana.dashboard.new(
    'Playback', uid=std.md5('playback.json')
  )
  .addTemplates([ds_template, job_template, instance_template])
  .addRow(
    grafana.row.new('Duration')
    .addPanels([
      durationPanel,
    ])
  );

{
  grafanaDashboards+:: {
    'playback.json': playback_dashboard,
  },
}
