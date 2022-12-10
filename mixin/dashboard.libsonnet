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

// Queries
local queries = {
  library_duration: 'sum(library_duration_total{' + matcher + '}) by (library)',
  library_storage: 'sum(library_storage_total{' + matcher + '}) by (library)',

  server_info: 'server_info{' + matcher + '}',
  server_network: 'rate(transmit_bytes_total{' + matcher + '}[$__rate_interval])',
  server_network_est: 'rate(estimated_transmit_bytes_total{' + matcher + '}[$__rate_interval])',

  host_cpu: 'host_cpu_util{' + matcher + '}',
  host_mem: 'host_mem_util{' + matcher + '}',

  // duration_by_day_bc: '(sum(max_over_time(play_seconds_total{'+matcher+'}[24h])) and on() day_of_week(timestamp(play_seconds_total{'+matcher+'})) == %d) or vector(0)',
  // duration_by_day_bc: 'play_seconds_total{' + matcher + '} and on() sum(max_over_time(play_seconds_total{' + matcher + '}[24h])) and on() day_of_week(timestamp(play_seconds_total{' + matcher + '})) == %d',
  duration_by_day_bc: std.format(|||
    sum(increase(play_seconds_total{%s}[$__interval])) by (media_type) * ignoring(day,dow) group_right

     label_replace(   
     label_replace(   
     label_replace(   
     label_replace(   
     label_replace(   
     label_replace(      
     label_replace(   
     count_values without() ("day", day_of_week(timestamp(
          sum(increase(          
             play_seconds_total{%s}          
          [$__interval]))  by (media_type)
        )
      ))
      ,"dow","Sunday","day","0")
      ,"dow","Monday","day","1")
      ,"dow","Tuesday","day","2")
      ,"dow","Wednesday","day","3")
      ,"dow","Thursday","day","4")
      ,"dow","Friday","day","5")
      ,"dow","Saturday","day","6")
  |||, [matcher, matcher]),
  duration_by_day_ts: 'sum(max_over_time(play_seconds_total{' + matcher + '}[24h])) by (library_type)',
  duration_by_hour: std.format(|||
    sum(increase(play_seconds_total{%s}[$__interval])) by (media_type) * ignoring(hour) group_right
        count_values without() ("hour", hour(timestamp(
          sum(increase(play_seconds_total{%s}[$__interval]))  by (media_type)
        )
      )
    )
  |||, [matcher, matcher]),
  duration_by_title: 'sum(increase(play_seconds_total{' + matcher + '}[$__interval])) by (media_type, title)',
  duration_by_user: 'sum(increase(play_seconds_total{' + matcher + '}[$__interval])) by (media_type, user)',
  duration_by_platform: 'sum(increase(play_seconds_total{' + matcher + '}[$__interval])) by (media_type, device_type)',

  duration_by_resolution: std.format(|||
    sum(increase(
      label_replace(
      label_replace(
          play_seconds_total{%s,stream_type!=""}
      , "res", "$1", "stream_resolution", "(.*)")
      , "res", "${1}p", "stream_resolution", "^([0-9]+)$")
      
    [$__interval:])) by (stream_type, res)
  |||, matcher),

  duration_by_file_resolution: std.format(|||
    sum(increase(
      label_replace(
      label_replace(
          play_seconds_total{%s,stream_type!=""}
      , "res", "$1", "stream_file_resolution", "(.*)")
      , "res", "${1}p", "stream_file_resolution", "^([0-9]+)$")
      
    [$__interval:])) by (stream_type, res)
  |||, matcher),
};

// Templates
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

// Per-Server visualizations
local platVerStat =
  statPanel.new(
    'Platform Version',
    datasource='$datasource',
  )
  .addTarget(
    grafana.prometheus.target(
      queries.server_info,
      legendFormat='{{platform}} - {{platform_version}}',
    )
  ) {
    options+: {
      textMode: 'name',
      graphMode: 'none',
      colorMode: 'background',
    },
  };

local plexVerStat =
  statPanel.new(
    'Plex Version',
    datasource='$datasource',
  )
  // TODO: Re-use the previous query, rather than making it again.
  .addTarget(
    grafana.prometheus.target(
      queries.server_info,
      legendFormat='{{version}}',
    )
  ) {
    options+: {
      textMode: 'name',
      graphMode: 'none',
      colorMode: 'background',
    },
  };

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
    { color: 'green', value: 0 },
    { color: 'yellow', value: 70 },
    { color: 'red', value: 90 },
  ])
  .addTarget(
    grafana.prometheus.target(
      queries.host_cpu,
      legendFormat='{{server}}'
    )
  );

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
    { color: 'green', value: 0 },
    { color: 'yellow', value: 70 },
    { color: 'red', value: 90 },
  ])
  .addTarget(
    grafana.prometheus.target(
      queries.host_mem,
      legendFormat='{{server}}'
    )
  );

local networkTs =
  graphPanel.new(
    'Network Utilization',
    datasource='$datasource',
  )
  .addTargets([
    grafana.prometheus.target(queries.server_network, legendFormat='Network'),
    grafana.prometheus.target(queries.server_network_est, legendFormat='Network Est.'),
  ])
  + utils.timeSeriesOverride(
    unit='bps',
    fillOpacity=10,
  );

// Duration stuff
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
  );

local durationDayBar =
  barGaugePanel.new(
    'Duration by day of week',
    datasource='$datasource',
    unit='s',
  )
  .addTarget(grafana.prometheus.target(queries.duration_by_day_bc, interval='5m')) {
    type: 'barchart',
    options+: {
      reduceOptions+: {
        calcs: [
          'max',
        ],
      },
      xField: 'dow\\media_type',
      stacking: 'normal',
      showValue: 'never',
    },
    fieldConfig+: {
      defaults+: {
        color: {
          mode: 'palette-classic',
        },
      },
    },
    transformations: [
      {
        id: 'reduce',
        options: {
          labelsToFields: true,
          reducers: [
            'sum',
          ],
        },
      },
      {
        id: 'groupingToMatrix',
        options: {
          columnField: 'media_type',
          rowField: 'dow',
          valueField: 'Total',
        },
      },
    ],
  };

local durationHourBar =
  barGaugePanel.new(
    'Duration by hour of day',
    datasource='$datasource',
    unit='s',
  )
  .addTarget(grafana.prometheus.target(queries.duration_by_hour, interval='5m')) {
    type: 'barchart',
    options+: {
      reduceOptions+: {
        calcs: [
          'max',
        ],
      },
      xField: 'hour\\media_type',
      stacking: 'normal',
    },
    fieldConfig+: {
      defaults+: {
        color: {
          mode: 'palette-classic',
        },
      },
    },
    transformations: [
      {
        id: 'labelsToFields',
        options: {
          mode: 'columns',
        },
      },
      {
        id: 'merge',
        options: {},
      },
      {
        id: 'groupBy',
        options: {
          fields: {
            Value: {
              aggregations: [
                'sum',
              ],
              operation: 'aggregate',
            },
            day: {
              aggregations: [],
              operation: 'groupby',
            },
            hour: {
              aggregations: [],
              operation: 'groupby',
            },
            media_type: {
              aggregations: [],
              operation: 'groupby',
            },
          },
        },
      },
      {
        id: 'groupingToMatrix',
        options: {
          columnField: 'media_type',
          rowField: 'hour',
          valueField: 'Value (sum)',
        },
      },
      {
        id: 'convertFieldType',
        options: {
          conversions: [
            {
              dateFormat: 'hh',
              destinationType: 'time',
              targetField: 'hour\\media_type',
            },
          ],
          fields: {},
        },
      },
    ],
  };

local topTitlesBar =
  barGaugePanel.new(
    'Top 10 Titles by duration',
    datasource='$datasource',
    unit='s',
  )
  .addTarget(grafana.prometheus.target(queries.duration_by_title, interval='5m', legendFormat='{{title}}')) {
    type: 'barchart',
    options+: {
      reduceOptions+: {
        calcs: [
          'max',
        ],
      },
      xField: 'title\\media_type',
      stacking: 'normal',
      orientation: 'horizontal',
      showValue: 'never',
    },
    fieldConfig+: {
      defaults+: {
        color: {
          mode: 'palette-classic',
        },
      },
    },
    transformations: [
      {
        id: 'reduce',
        options: {
          labelsToFields: true,
          reducers: [
            'sum',
          ],
        },
      },
      {
        id: 'sortBy',
        options: {
          fields: {},
          sort: [
            {
              desc: true,
              field: 'Total',
            },
          ],
        },
      },
      {
        id: 'limit',
        options: {},
      },
      {
        id: 'groupingToMatrix',
        options: {
          columnField: 'media_type',
          rowField: 'title',
          valueField: 'Total',
        },
      },
    ],
  };

local topUsersBar =
  barGaugePanel.new(
    'Top 10 Users by duration',
    datasource='$datasource',
    unit='s',
  )
  .addTarget(grafana.prometheus.target(queries.duration_by_user, interval='5m', legendFormat='{{user}}')) {
    type: 'barchart',
    options+: {
      reduceOptions+: {
        calcs: [
          'max',
        ],
      },
      xField: 'user\\media_type',
      stacking: 'normal',
      orientation: 'horizontal',
      showValue: 'never',
    },
    fieldConfig+: {
      defaults+: {
        color: {
          mode: 'palette-classic',
        },
      },
    },
    transformations: [
      {
        id: 'reduce',
        options: {
          labelsToFields: true,
          reducers: [
            'sum',
          ],
        },
      },
      {
        id: 'sortBy',
        options: {
          fields: {},
          sort: [
            {
              desc: true,
              field: 'Total',
            },
          ],
        },
      },
      {
        id: 'limit',
        options: {},
      },
      {
        id: 'groupingToMatrix',
        options: {
          columnField: 'media_type',
          rowField: 'user',
          valueField: 'Total',
        },
      },
    ],
  };

local topPlatformsBar =
  barGaugePanel.new(
    'Top 10 Platforms by duration',
    datasource='$datasource',
    unit='s',
  )
  .addTarget(grafana.prometheus.target(queries.duration_by_platform, interval='5m', legendFormat='{{device_type}}')) {
    type: 'barchart',
    options+: {
      reduceOptions+: {
        calcs: [
          'max',
        ],
      },
      xField: 'device_type\\media_type',
      stacking: 'normal',
      orientation: 'horizontal',
      showValue: 'never',
    },
    fieldConfig+: {
      defaults+: {
        color: {
          mode: 'palette-classic',
        },
      },
    },
    transformations: [
      {
        id: 'reduce',
        options: {
          labelsToFields: true,
          reducers: [
            'sum',
          ],
        },
      },
      {
        id: 'sortBy',
        options: {
          fields: {},
          sort: [
            {
              desc: true,
              field: 'Total',
            },
          ],
        },
      },
      {
        id: 'limit',
        options: {},
      },
      {
        id: 'groupingToMatrix',
        options: {
          columnField: 'media_type',
          rowField: 'device_type',
          valueField: 'Total',
        },
      },
    ],
  };

// Streaming stats
local sourceResBar =
  barGaugePanel.new(
    'Duration by source resolution',
    datasource='$datasource',
    unit='s',
  )
  .addTarget(grafana.prometheus.target(queries.duration_by_file_resolution, interval='5m', legendFormat='{{res}}')) {
    type: 'barchart',
    options+: {
      reduceOptions+: {
        calcs: [
          'max',
        ],
      },
      stacking: 'normal',
      showValue: 'never',
    },
    fieldConfig+: {
      defaults+: {
        color: {
          mode: 'palette-classic',
        },
      },
    },
    transformations: [
      {
        id: 'reduce',
        options: {
          labelsToFields: true,
          reducers: [
            'sum',
          ],
        },
      },
      {
        id: 'sortBy',
        options: {
          fields: {},
          sort: [
            {
              desc: true,
              field: 'Total',
            },
          ],
        },
      },
      {
        id: 'groupingToMatrix',
        options: {
          columnField: 'stream_type',
          rowField: 'res',
          valueField: 'Total',
        },
      },
    ],
  };

local streamResBar =
  barGaugePanel.new(
    'Duration by streamed resolution',
    datasource='$datasource',
    unit='s',
  )
  .addTarget(grafana.prometheus.target(queries.duration_by_resolution, interval='5m', legendFormat='{{res}}')) {
    type: 'barchart',
    options+: {
      reduceOptions+: {
        calcs: [
          'max',
        ],
      },
      stacking: 'normal',
      showValue: 'never',
    },
    fieldConfig+: {
      defaults+: {
        color: {
          mode: 'palette-classic',
        },
      },
    },
    transformations: [
      {
        id: 'reduce',
        options: {
          labelsToFields: true,
          reducers: [
            'sum',
          ],
        },
      },
      {
        id: 'sortBy',
        options: {
          fields: {},
          sort: [
            {
              desc: true,
              field: 'Total',
            },
          ],
        },
      },
      {
        id: 'groupingToMatrix',
        options: {
          columnField: 'stream_type',
          rowField: 'res',
          valueField: 'Total',
        },
      },
    ],
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
  .addPanels([
    grafana.row.new('$server', repeat='server') { gridPos: { h: 1, w: 24, x: 0, y: 0 } },
    platVerStat { gridPos: { h: 4, w: 6, x: 0, y: 1 } },
    hostCpuStat { gridPos: { h: 4, w: 3, x: 6, y: 1 } },
    durationStat { gridPos: { h: 4, w: 15, x: 9, y: 1 } },
    plexVerStat { gridPos: { h: 4, w: 6, x: 0, y: 5 } },
    hostMemStat { gridPos: { h: 4, w: 3, x: 6, y: 5 } },
    storageStat { gridPos: { h: 4, w: 15, x: 9, y: 5 } },
    networkTs { gridPos: { h: 7, w: 24, x: 0, y: 9 } },
    grafana.row.new('Duration') { gridPos: { h: 1, w: 24, x: 0, y: 16 } },
    durationGraph { gridPos: { h: 7, w: 24, x: 0, y: 17 } },
    durationDayBar { gridPos: { h: 7, w: 12, x: 0, y: 24 } },
    durationHourBar { gridPos: { h: 7, w: 12, x: 12, y: 24 } },
    topTitlesBar { gridPos: { h: 7, w: 8, x: 0, y: 31 } },
    topUsersBar { gridPos: { h: 7, w: 8, x: 8, y: 31 } },
    topPlatformsBar { gridPos: { h: 7, w: 8, x: 16, y: 31 } },
    grafana.row.new('Streaming') { gridPos: { h: 1, w: 24, x: 0, y: 32 } },
    sourceResBar { gridPos: { h: 7, w: 12, x: 0, y: 33 } },
    streamResBar { gridPos: { h: 7, w: 12, x: 16, y: 33 } },
  ]);

{
  grafanaDashboards+:: {
    'mediaserver.json': playback_dashboard,
  },
}
