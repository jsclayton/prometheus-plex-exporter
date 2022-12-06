local grafana = import 'grafonnet/grafana.libsonnet';
local graphPanel = grafana.graphPanel;
local barGaugePanel = grafana.barGaugePanel;

local utils = import 'snmp-mixin/lib/utils.libsonnet';

local matcher = 'job=~"$job", instance=~"$instance"';

local dow = {
  Sunday: 0,
  Monday: 1,
  Tuesday: 2,
  Wednesday: 3,
  Thursday: 4,
  Friday: 5,
  Saturday: 6,
};

local queries = {
  duration_by_day_bc: '(sum(increase(play_seconds_total[24h])) and on() day_of_week(timestamp(play_seconds_total)) == %d) or vector(0)',
  duration_by_day_ts: 'sum(increase(play_seconds_total{' + matcher + '}[24h]))',
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
    'label_values(plays_total{job=~"$job", instance=~"$instance"})',
    label='server',
    refresh='load',
    multi=true,
    includeAll=true,
    allValues='.+',
  );

local durationGraph =
  graphPanel.new(
    'Duration',
    datasource='$datasource',
  )
  .addTarget(grafana.prometheus.target(queries.duration_by_day_ts, interval='1d', legendFormat='Total'))
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
      grafana.prometheus.target(std.format(queries.duration_by_day_bc, dow[day]), legendFormat=day)
      for day in std.objectFields(dow)
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
    'Playback', uid=std.md5('playback.json')
  )
  .addTemplates([
    ds_template,
    job_template,
    instance_template,
    server_template,
  ])
  .addRow(
    grafana.row.new('Duration')
    .addPanels([
      durationGraph,
      durationDayBar,
    ])
  );

{
  grafanaDashboards+:: {
    'playback.json': playback_dashboard,
  },
}

/*
{
  id: 5,
  gridPos: {
    x: 0,
    y: 8,
    w: 12,
    h: 8,
  },
  type: 'bargauge',
  title: 'By Day of Week',
  targets: [
    {
      refId: 'A',
      datasource: {
        type: 'prometheus',
        uid: 'grafanacloud-prom',
      },
      editorMode: 'code',
      expr: 'sum((increase(play_seconds_total[24h])) and on() day_of_week(timestamp(play_seconds_total)) == 0) or vector(0)',
      legendFormat: 'Sunday',
      range: true,
      interval: '',
    },
    {
      datasource: {
        uid: 'grafanacloud-prom',
        type: 'prometheus',
      },
      refId: 'B',
      hide: false,
      editorMode: 'code',
      expr: 'sum((increase(play_seconds_total[1d])) and on() day_of_week(timestamp(play_seconds_total)) == 1) or vector(0)',
      legendFormat: 'Monday',
      range: true,
      interval: '',
    },
    {
      datasource: {
        uid: 'grafanacloud-prom',
        type: 'prometheus',
      },
      refId: 'C',
      hide: false,
      editorMode: 'code',
      expr: 'sum((increase(play_seconds_total[24h])) and on() day_of_week(timestamp(play_seconds_total)) == 2) or vector(0)',
      legendFormat: 'Tuesday',
      range: true,
      interval: '',
    },
    {
      refId: 'D',
      datasource: {
        type: 'prometheus',
        uid: 'grafanacloud-prom',
      },
      editorMode: 'code',
      expr: 'sum((increase(play_seconds_total[24h])) and on() day_of_week() == 3) or vector(0)',
      legendFormat: 'Wednesday',
      range: true,
      interval: '1d',
      hide: false,
    },
    {
      refId: 'E',
      datasource: {
        type: 'prometheus',
        uid: 'grafanacloud-prom',
      },
      editorMode: 'code',
      expr: 'sum((increase(play_seconds_total[24h])) and on() day_of_week() == 4) or vector(0)',
      legendFormat: 'Thursday',
      range: true,
      interval: '1d',
      hide: false,
    },
    {
      refId: 'F',
      datasource: {
        type: 'prometheus',
        uid: 'grafanacloud-prom',
      },
      editorMode: 'code',
      expr: 'sum((increase(play_seconds_total[24h])) and on() day_of_week() == 5) or vector(0)',
      legendFormat: 'Friday',
      range: true,
      interval: '1d',
      hide: false,
    },
    {
      refId: 'G',
      datasource: {
        type: 'prometheus',
        uid: 'grafanacloud-prom',
      },
      editorMode: 'code',
      expr: 'sum((increase(play_seconds_total[24h])) and on() day_of_week() == 6) or vector(0)',
      legendFormat: 'Saturday',
      range: true,
      interval: '1d',
      hide: false,
    },
  ],
  options: {
    reduceOptions: {
      values: false,
      calcs: [
        'max',
      ],
      fields: '',
    },
    orientation: 'auto',
    displayMode: 'gradient',
    showUnfilled: true,
    minVizWidth: 0,
    minVizHeight: 10,
  },
  fieldConfig: {
    defaults: {
      mappings: [],
      thresholds: {
        mode: 'absolute',
        steps: [
          {
            value: null,
            color: 'green',
          },
          {
            value: 80,
            color: 'red',
          },
        ],
      },
      color: {
        mode: 'fixed',
        fixedColor: 'orange',
      },
      unit: 's',
      noValue: '0',
    },
    overrides: [],
  },
  datasource: {
    uid: 'grafanacloud-prom',
    type: 'prometheus',
  },
  pluginVersion: '9.2.7-8da65d62',
  transformations: [],
}
*/



