import React, { useEffect, useState } from 'react';
import { useTheme } from '@material-ui/core/styles';
import {
  BarChart,
  Bar,
  CartesianGrid,
  XAxis,
  YAxis,
  Label,
  ResponsiveContainer,
  Tooltip,
  Legend,
} from 'recharts';
import Title from './Title';

import { grpc } from '@improbable-eng/grpc-web';
import {
  DataAggregatorGetTimeRequest,
  DataAggregatorGetTimeResponse,
} from 'productimon/proto/svc/aggregator_pb';
import { DataAggregator } from 'productimon/proto/svc/aggregator_pb_service';
import { Interval, Timestamp } from 'productimon/proto/common/common_pb';

function last10MinIntervals() {
  const gap = 60 * 10 ** 9;
  var curr = new Date().getTime() * 10 ** 6;
  curr = curr - (curr % gap) + gap;
  const ret = [];
  for (let i = 0; i < 10; i++) {
    const interval = new Interval();
    const start = new Timestamp();
    start.setNanos(curr - gap);
    const end = new Timestamp();
    end.setNanos(curr);
    interval.setStart(start);
    interval.setEnd(end);
    curr -= gap;
    ret.push(interval);
  }
  return ret;
}

function formatNanoTS(ts) {
  const date = new Date(ts / 10 ** 6);
  return `${date.getHours()}:${date.getMinutes()}`;
}

function transformRange(data) {
  const dataPoints = data.getDataList();
  const ret = {
    time: formatNanoTS(data.getInterval().getStart()),
  };
  dataPoints.forEach((point) => {
    // TODO convert it to a proper unit
    // always using seconds for now
    ret[point.getLabel()] = point.getTime() / 10 ** 9;
  });
  return ret;
}

function getUniqLabels(response) {
  return response
    .getDataList()
    .reduce(
      (result, range) => [
        ...result,
        range
          .getDataList()
          .reduce((labels, datapoint) => [...labels, datapoint.getLabel()], []),
      ],
      [],
    )
    .reduce((result, arr) => [...result, ...arr], [])
    .filter((v, i, a) => a.indexOf(v) === i)
    .sort();
}

const labelColorMap = new Map();
const colors = ['#ef5350', '#d81b60', '#2196f3', '#4db6ac', '#9ccc65'];
var colorIdx = 0;
function getLabelColor(label) {
  if (!labelColorMap.has(label)) {
    labelColorMap.set(label, colors[colorIdx]);
    colorIdx += 1;
    colorIdx = colorIdx % colors.length;
  }
  return labelColorMap.get(label);
}

export default function Histogram() {
  const theme = useTheme();
  const [data, setData] = useState([]);
  const [dataKeys, setDataKeys] = useState([]);

  useEffect(() => {
    const intervals = last10MinIntervals();

    /* Get time data for all device for last 10 minutes */
    const request = new DataAggregatorGetTimeRequest();
    request.setDevicesList([]);
    request.setIntervalsList(intervals);
    request.setGroupBy(DataAggregatorGetTimeRequest.GroupBy.LABEL);

    const token = window.localStorage.getItem('token');
    grpc.unary(DataAggregator.GetTime, {
      host: '/rpc',
      metadata: new grpc.Metadata({ Authorization: token }),
      onEnd: ({ status, statusMessage, headers, message }) => {
        if (status != 0) {
          console.error(
            `Error getting res, status ${status}: ${statusMessage}`,
          );
          return;
        }
        setDataKeys(getUniqLabels(message));
        setData(message.getDataList().map(transformRange).reverse());
      },
      request,
    });
  }, []);

  return (
    <React.Fragment>
      <Title>Activity in last 10 minutes</Title>
      <ResponsiveContainer>
        <BarChart
          data={data}
          margin={{
            top: 16,
            right: 16,
            bottom: 0,
            left: 24,
          }}
        >
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="name" />
          <YAxis />
          {/* <Tooltip /> */}
          <Legend />
          {dataKeys.map((label, index) => (
            <Bar
              key={index}
              dataKey={label}
              stackId="a"
              fill={getLabelColor(label)}
            />
          ))}
        </BarChart>
      </ResponsiveContainer>
    </React.Fragment>
  );
}
