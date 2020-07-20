import React, { useEffect, useState } from "react";
import {
  PieChart as RechartsPieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Tooltip,
  Legend,
} from "recharts";

import { grpc } from "@improbable-eng/grpc-web";
import { DataAggregatorGetTimeRequest } from "productimon/proto/svc/aggregator_pb";
import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Interval, Timestamp } from "productimon/proto/common/common_pb";

import { getLabelColor, humanizeDuration, toSec, calculateDate } from "../Utils";

function createData(program, time, label) {
  return { program, time, label };
}

export default function PieChart(props) {
  const [rows, setRows] = React.useState([]);
  var data = [];

  useEffect(() => {
    const interval = new Interval();
    const start = new Timestamp();
    const startDate = calculateDate(
      props.graphSpec.startTimeUnit,
      props.graphSpec.startTimeVal
    );
    const endDate = calculateDate(
      props.graphSpec.endTimeUnit,
      props.graphSpec.endTimeVal
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
        // Sort data by most used applications
        const sorted = message
          .getDataList()[0]
          .getDataList()
          .sort((a, b) => b.getTime() - a.getTime());

        // Cumulatively store the amount of time used in apps other than main displayed
        var othTime = 0;
        for (var i = 0; i < sorted.length; i++) {
          if (i < props.graphSpec.numItems) {
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
      <ResponsiveContainer height="100%">
        <RechartsPieChart width={200} height={200}>
          <Pie
            innerRadius={44}
            outerRadius={88}
            data={rows}
            dataKey="time"
            nameKey="label"
            label={({ label, time }) => `${label}: ${humanizeDuration(time)}`}
            labelLine={false}
          >
            {rows.map((data, index) => (
              <Cell key={index} fill={getLabelColor(data.label)} />
            ))}
          </Pie>
          { props.fullscreen && <Tooltip />}
          <Legend />
        </RechartsPieChart>
      </ResponsiveContainer>
    </React.Fragment>
  );
}
