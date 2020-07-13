import React, { useEffect, useState } from "react";
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
import { humanizeDuration, toSec, google_colors, calculateDate } from "./utils";
import { grpc } from "@improbable-eng/grpc-web";
import { DataAggregatorGetTimeRequest } from "productimon/proto/svc/aggregator_pb";
import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Interval, Timestamp } from "productimon/proto/common/common_pb";

function createData(program, time, label) {
  return { program, time, label };
}

const colors = google_colors;

export default function DisplayPie(props) {
  const theme = useTheme();
  const [rows, setRows] = React.useState([]);
  const [title, setTitle] = useState(props.spec.graphTitle);
  var data = [];

  useEffect(() => {
    const interval = new Interval();
    const start = new Timestamp();
    const startDate = calculateDate(
      props.spec.startTimeUnit,
      props.spec.startTimeVal
    );
    const endDate = calculateDate(
      props.spec.endTimeUnit,
      props.spec.endTimeVal
    );

    start.setNanos(startDate * 10 ** 6);
    const end = new Timestamp();
    end.setNanos(endDate * 10 ** 6);
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

        // Cumulatively store the amount of time used in apps other than main displayed
        var othTime = 0;
        for (var i = 0; i < sorted.length; i++) {
          if (i < props.spec.numItems) {
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
    if (props.spec.graphTitle === "")
      setTitle(props.spec.numItems + " most used");
  }, []);

  // TODO: add legend to the pie chart
  return (
    <React.Fragment>
      <Title>{title}</Title>
      <ResponsiveContainer height="80%">
        <PieChart width={200} height={200}>
          <Pie
            innerRadius={44}
            outerRadius={88}
            data={rows}
            dataKey="time"
            nameKey="program"
            label={({ program, time }) =>
              `${program}: ${humanizeDuration(time)}`
            }
            labelLine={false}
          >
            {rows.map((data, index) => (
              <Cell key={index} fill={props.getLabelColor(data.program)} />
            ))}
          </Pie>
          <Tooltip />
        </PieChart>
      </ResponsiveContainer>
    </React.Fragment>
  );
}
