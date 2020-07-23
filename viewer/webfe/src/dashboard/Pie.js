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

import {
  getLabelColor,
  humanizeDuration,
  toSec,
  calculateDate,
} from "../Utils";

const DISPLAY_LABEL_THRESHOLD = 0.03;

function createData(program, time, label) {
  return { program, time, label, humanizedTime: humanizeDuration(time) };
}

export default function PieChart({ graphSpec, fullscreen }) {
  const [sectors, setSectors] = React.useState([]);
  const [totalTime, setTotalTime] = React.useState(0);
  var data = [];

  const numItems = graphSpec.numItems || 5;

  useEffect(() => {
    const interval = new Interval();
    const start = new Timestamp();
    const startDate = calculateDate(
      graphSpec.startTimeUnit,
      graphSpec.startTimeVal
    );
    const endDate = calculateDate(graphSpec.endTimeUnit, graphSpec.endTimeVal);

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

        setTotalTime(
          toSec(sorted.reduce((sum, point) => sum + point.getTime(), 0))
        );

        const entries = sorted
          .slice(0, numItems)
          .map((data) =>
            createData(data.getApp(), toSec(data.getTime()), data.getLabel())
          );
        if (sorted.length > numItems) {
          const other = sorted
            .slice(numItems)
            .reduce(
              (ret, point) =>
                createData(
                  "Other",
                  (ret.time || 0) + toSec(point.getTime()),
                  "Other"
                ),
              {}
            );
          setSectors([...entries, other]);
        } else {
          setSectors(entries);
        }
      },
      request,
    });
  }, []);

  return (
    <React.Fragment>
      <ResponsiveContainer height="100%">
        <RechartsPieChart>
          <Pie
            innerRadius="50%"
            data={sectors}
            dataKey="time"
            nameKey="label"
            label={({ label, time, humanizedTime }) =>
              time / totalTime > DISPLAY_LABEL_THRESHOLD
                ? `${label}: ${humanizedTime}`
                : null
            }
            labelLine={false}
          >
            {sectors.map((data, index) => (
              <Cell key={index} fill={getLabelColor(data.label)} />
            ))}
          </Pie>
          {fullscreen && (
            <Tooltip
              formatter={(_, __, { payload: { humanizedTime } }) =>
                humanizedTime
              }
            />
          )}
          <Legend />
        </RechartsPieChart>
      </ResponsiveContainer>
    </React.Fragment>
  );
}
