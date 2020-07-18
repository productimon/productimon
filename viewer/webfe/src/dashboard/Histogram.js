import React, { useEffect, useState } from "react";
import moment from "moment";
import { timeUnits } from "./utils.js";
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

/* use interval to deduce a appropriate format for the ts
 * span and totalSpan in miliseconds */
function formatNanoTS(ts, span, totalSpan) {
  // console.log(`formatting ${ts} with total span ${totalSpan} and span ${span}`);

  const m = moment(ts / 10 ** 9, "X");

  const components = [];

  /* only show date if the histogram spans over a year */
  let dateFmt = "YY-MM-DD";
  if (totalSpan < timeUnits.Days) {
    dateFmt = "";
  } else if (totalSpan < timeUnits.Years) {
    dateFmt = "MM-DD";
  }
  if (dateFmt) components.push(m.format(dateFmt));

  let timeFmt = "";
  if (span < timeUnits.Minutes) {
    timeFmt = "HH:mm:ss";
  } else if (span < timeUnits.Days) {
    timeFmt = "HH:mm";
  }
  if (timeFmt) components.push(m.format(timeFmt));

  return components.join(" ");
}

/* Returns a transform function
 * use closure to capture the total time span
 * of all the data to be used to deduce an appropriate datetime format
 * Assume allData in reverse chronological order
 */
function transformRange(allData) {
  /* Time span of the whole histogram */
  let totalTimeSpan = allData.length
    ? allData[0].getInterval().getEnd().getNanos() -
      allData[allData.length - 1].getInterval().getStart().getNanos()
    : 0;
  totalTimeSpan /= 10 ** 6;
  return (data) => {
    /* Time span of this perticular interval */
    let timeSpan =
      data.getInterval().getEnd().getNanos() -
      data.getInterval().getStart().getNanos();
    timeSpan /= 10 ** 6;
    const ret = {
      label: formatNanoTS(
        data.getInterval().getStart().getNanos(),
        timeSpan,
        totalTimeSpan
      ),
    };
    const dataPoints = data.getDataList();
    dataPoints.forEach((point) => {
      // TODO convert it to a proper unit
      // always using seconds for now

      ret[point.getLabel()] = point.getTime() / 10 ** 9;
    });
    return ret;
  };
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
        setData(
          message
            .getDataList()
            .map(transformRange(message.getDataList()))
            .reverse()
        );
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
          <IconButton>
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
          <XAxis dataKey="label" />} />
          <YAxis
            label={{ value: "seconds", angle: -90, position: "insideLeft" }}
          />
          {dataKeys.map((label, index) => (
            <Bar
              key={index}
              dataKey={label}
              stackId="a"
              fill={props.getLabelColor(label)}
            />
          ))}
          <Legend />
        </BarChart>
      </ResponsiveContainer>
    </React.Fragment>
  );
}
