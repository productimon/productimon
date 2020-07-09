import React, { useEffect } from "react";
import { useTheme } from "@material-ui/core/styles";
import {
  PieChart,
  Pie,
  Sector,
  Cell,
  ResponsiveContainer,
  Tooltip,
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
import moment from "moment";

// format a time in seconds to readable string
// TODO: refactor
function humanizeDuration(seconds) {
  const duration = moment.duration(seconds * 10 ** 3);
  if (seconds < 60) {
    return `${duration.seconds()} seconds`;
  }
  return duration.humanize();
}

// TODO: make time format customisable
function toSec(nanoseconds) {
  return nanoseconds / 10 ** 9;
}

function createData(program, time, label) {
  return { program, time, label };
}

// TODO: make color map univseral for all graphs
const labelColorMap = new Map();
const colors = ["#ef5350", "#d81b60", "#2196f3", "#4db6ac", "#9ccc65"];
var colorIdx = 0;
function getLabelColor(label) {
  if (!labelColorMap.has(label)) {
    labelColorMap.set(label, colors[colorIdx]);
    colorIdx += 1;
    colorIdx = colorIdx % colors.length;
  }
  return labelColorMap.get(label);
}

export default function DisplayPie() {
  const theme = useTheme();
  const [rows, setRows] = React.useState([createData("init", 1, 3)]);
  var data = [];

  useEffect(() => {
    const interval = new Interval();
    const start = new Timestamp();
    start.setNanos(0);
    const end = new Timestamp();
    end.setNanos(new Date().getTime() * 10 ** 6);
    interval.setStart(start);
    interval.setEnd(end);

    const request = new DataAggregatorGetTimeRequest();
    // Get time data for all device and all intervals
    request.setDevicesList([]);
    request.setIntervalsList([interval]);
    request.setGroupBy(DataAggregatorGetTimeRequest.GroupBy.APPLICATION);

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
        // Sort data by most used applications
        const sorted = message
          .getDataList()[0]
          .getDataList()
          .sort((a, b) => b.getTime() - a.getTime());

        // Display the 5 most used programs and the rest to a sector called 'other'
        // To change number of applications displayed change variable appsDisplay
        var appsDisplay = 5;
        // Cumulatively store the amount of time used in apps other than main displayed
        var othTime = 0;
        for (var i = 0; i < sorted.length; i++) {
          if (i < appsDisplay) {
            data.push(
              createData(
                sorted[i].getApp(),
                toSec(sorted[i].getTime()),
                sorted[i].getLabel()
              )
            );
          } else {
            othTime += sorted[i].getTime();
          }
        }
        if (othTime > 0) {
          data.push(createData("Other", toSec(othTime), "other"));
        }
        setRows(data);
      },
      request,
    });
  }, []);

  // TODO: add legend to the pie chart
  return (
    <React.Fragment>
      <Title>Five Most used Applications</Title>
      <ResponsiveContainer>
        <PieChart width={200} height={200}>
          <Pie
            innerRadius={38}
            outerRadius={76}
            data={rows}
            dataKey="time"
            nameKey="program"
            label={({ program, time }) =>
              `${program}: ${humanizeDuration(time)}`
            }
            labelLine={false}
          >
            {rows.map((label, index) => (
              <Cell key={label} fill={colors[index % colors.length]} />
            ))}
          </Pie>
          <Tooltip />
        </PieChart>
      </ResponsiveContainer>
    </React.Fragment>
  );
}
