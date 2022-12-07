local grafana = import 'grafonnet/grafana.libsonnet';
local barGaugePanel = grafana.barGaugePanel;
local graphPanel = grafana.graphPanel;
local statPanel = grafana.statPanel;


local utils = import 'snmp-mixin/lib/utils.libsonnet';

local matcher = 'job=~"$job", instance=~"$instance", server=~"$server"';

local dow = [
  'Sunday',
  'Monday',
  'Tuesday',
  'Wednesday',
  'Thursday',
  'Friday',
  'Saturday',
];

local queries = {
  library_duration: 'sum(library_duration_total{' + matcher + '}) by (library)',
  library_storage: 'sum(library_storage_total{' + matcher + '}) by (library)',

  host_cpu: 'host_cpu_util{' + matcher + '}',
  host_mem: 'host_mem_util{' + matcher + '}',

  // duration_by_day_bc: '(sum(max_over_time(play_seconds_total{'+matcher+'}[24h])) and on() day_of_week(timestamp(play_seconds_total{'+matcher+'})) == %d) or vector(0)',
  duration_by_day_bc: 'play_seconds_total{' + matcher + '} and on() sum(max_over_time(play_seconds_total{' + matcher + '}[24h])) and on() day_of_week(timestamp(play_seconds_total{' + matcher + '})) == %d',
  duration_by_day_ts: 'sum(max_over_time(play_seconds_total{' + matcher + '}[24h])) by (library_type)',
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
    'label_values(plays_total, job)',
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
    'label_values(plays_total{job=~"$job"}, instance)',
    label='instance',
    refresh='load',
    multi=true,
    includeAll=true,
    allValues='.+',
    sort=1,
  );

local server_template =
  grafana.template.new(
    'server',
    '$datasource',
    'label_values(plays_total{job=~"$job", instance=~"$instance"}, server)',
    label='server',
    refresh='load',
    multi=true,
    includeAll=true,
    allValues='.+',
  );

local hostCpuStat =
  statPanel.new(
    'Host CPU Utilization by Server',
    description='Only available on Plex servers with PlexPass',
    datasource='$datasource',
    unit='percent',
    reducerFunction='last',
    graphMode='none',
  )
  .addThresholds([
    { color: 'green', value: 0},    
    { color: 'yellow', value: 70 },
    { color: 'red', value: 90 },    
  ])
  .addTarget(
    grafana.prometheus.target(
      queries.host_cpu,
      legendFormat='{{server}}'
    )
  ) + { span: 3 };

local hostMemStat =
  statPanel.new(
    'Host Memory Utilization by Server',
    description='Only available on Plex servers with PlexPass',
    datasource='$datasource',
    unit='percent',
    reducerFunction='last',
    graphMode='none',
  )
  .addThresholds([
    { color: 'green', value: 0},    
    { color: 'yellow', value: 70 },
    { color: 'red', value: 90 },    
  ])
  .addTarget(
    grafana.prometheus.target(
      queries.host_mem,
      legendFormat='{{server}}'
    )
  ) + { span: 3 };

local durationStat =
  statPanel.new(
    'Library Duration',
    datasource='$datasource',
    unit='ms',
    reducerFunction='max',
    graphMode='none',
    colorMode='background',
  )
  .addTarget(
    grafana.prometheus.target(
      queries.library_duration,
      legendFormat='{{library}}'
    )
  ) + { 
    span: 9,
    fieldConfig+: {
      defaults+: {
        color: {
          mode: 'fixed',
          fixedColor: 'light-blue',
        },
      },
    },
  };

local storageStat =
  statPanel.new(
    'Library Storage',
    datasource='$datasource',
    unit='bytes',
    reducerFunction='max',
    graphMode='none',
    colorMode='background',
  )
  .addTarget(
    grafana.prometheus.target(
      queries.library_storage,
      legendFormat='{{library}}'
    )
  ) + { 
    span: 9,
    fieldConfig+: {
      defaults+: {
        color: {
          mode: 'fixed',
          fixedColor: 'light-purple',
        },
      },
    },
  };

local durationGraph =
  graphPanel.new(
    'Duration',
    datasource='$datasource',
  )
  .addTarget(grafana.prometheus.target(queries.duration_by_day_ts, interval='1d', legendFormat='{{library_type}}', intervalFactor=1))
  + utils.timeSeriesOverride(
    unit='s',
    fillOpacity=10,
    showPoints='never',
  ) { span: 12 };

local durationDayBar =
  barGaugePanel.new(
    'Duration by day of week',
    datasource='$datasource',
    unit='s',
  )
  .addTargets(
    [
      grafana.prometheus.target(std.format(queries.duration_by_day_bc, day), legendFormat=dow[day],)
      for day in std.range(0, std.length(dow) - 1)
    ]
  ) {
    span: 6,
    options+: {
      reduceOptions+: {
        calcs: [
          'max',
        ],
      },
    },
    fieldConfig+: {
      defaults+: {
        color: {
          mode: 'continuous-BlPu',
        },
      },
    },
  };

local playback_dashboard =
  grafana.dashboard.new(
    'Media Server',
    uid=std.md5('mediaserver.json'),
    time_from='now-7d',
  )
  .addTemplates([
    ds_template,
    job_template,
    instance_template,
    server_template,
  ])
  .addRow(
    grafana.row.new('Overview')
    .addPanels([
      hostCpuStat,
      durationStat,
      hostMemStat,
      storageStat,
    ])
  )
  .addRow(
    grafana.row.new('Duration')
    .addPanels([
      durationGraph,
      durationDayBar,
    ])
  );

{
  grafanaDashboards+:: {
    'mediaserver.json': playback_dashboard,
  },
}
