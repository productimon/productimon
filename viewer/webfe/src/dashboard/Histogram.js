import React, { useEffect, useState } from "react";
import { useTheme } from "@material-ui/core/styles";
import {
  BarChart,
  Bar,
  CartesianGrid,
  XAxis,
  YAxis,
  ResponsiveContainer,
  Legend,
} from "recharts";
import Title from "./Title";

import { grpc } from "@improbable-eng/grpc-web";
import {
  DataAggregatorGetTimeRequest,
  DataAggregatorGetTimeResponse,
} from "productimon/proto/svc/aggregator_pb";
import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Interval, Timestamp } from "productimon/proto/common/common_pb";
import IconButton from "@material-ui/core/IconButton";
import DeleteIcon from "@material-ui/icons/Delete";
import { calculateDate } from "./utils";

// startDate and endDate are miliseconds.
function createIntervals(startDate, endDate, numIntervals) {
  const intervalDuration = (endDate - startDate) / numIntervals;
  var curr = endDate;
  curr = curr - (curr % intervalDuration) + intervalDuration;
  const ret = [];
  for (let i = 0; i < numIntervals; ++i) {
    const interval = new Interval();
    const start = new Timestamp();
    start.setNanos((curr - intervalDuration) * 10 ** 6);
    const end = new Timestamp();
    end.setNanos(curr * 10 ** 6);
    interval.setStart(start);
    interval.setEnd(end);
    curr -= intervalDuration;
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
      []
    )
    .reduce((result, arr) => [...result, ...arr], [])
    .filter((v, i, a) => a.indexOf(v) === i)
    .sort();
}

export default function Histogram(props) {
  const [data, setData] = useState([]);
  const [dataKeys, setDataKeys] = useState([]);
  const [title, setTitle] = useState(props.spec.graphTitle);

  useEffect(() => {
    const startDate = calculateDate(
      props.spec.startTimeUnit,
      props.spec.startTimeVal
    );
    const endDate = calculateDate(
      props.spec.endTimeUnit,
      props.spec.endTimeVal
    );
    const numIntervals = props.spec.intervals;
    const intervals = createIntervals(startDate, endDate, numIntervals);

    const request = new DataAggregatorGetTimeRequest();
    request.setDevicesList([]);
    request.setIntervalsList(intervals);
    request.setGroupBy(DataAggregatorGetTimeRequest.GroupBy.LABEL);

    const token = window.localStorage.getItem("token");
    grpc.unary(DataAggregator.GetTime, {
      host: "/rpc",
      metadata: new grpc.Metadata({ Authorization: token }),
      onEnd: ({ status, statusMessage, headers, message }) => {
        if (status != 0) {
          console.error(
            `Error getting res, status ${status}: ${statusMessage}`
          );
          return;
        }
        setDataKeys(getUniqLabels(message));
        setData(message.getDataList().map(transformRange).reverse());
      },
      request,
    });
    if (title === "")
      setTitle(
        "From " +
          props.spec.startTimeVal +
          " " +
          props.spec.startTimeUnit +
          " until " +
          props.spec.endTimeVal +
          " " +
          props.spec.endTimeUnit +
          " ago."
      );
  }, []);

  return (
    <React.Fragment>
      {/* Alignment is hard...
          <IconButton aria-label="delete">
          <DeleteIcon />
          </IconButton> */}
      <Title>{title}</Title>
      <ResponsiveContainer height="80%">
        <BarChart
          data={data}
          margin={{
            top: 16,
            right: 16,
            bottom: 0,
            left: 16,
          }}
        >
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="name" />
          <YAxis />
          <Legend />
          {dataKeys.map((label, index) => (
            <Bar
              key={index}
              dataKey={label}
              stackId="a"
              fill={props.getLabelColor(label)}
            />
          ))}
        </BarChart>
      </ResponsiveContainer>
    </React.Fragment>
  );
}
