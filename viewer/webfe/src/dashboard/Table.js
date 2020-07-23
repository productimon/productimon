import React, { useEffect, useState } from "react";
import { useHistory } from "react-router-dom";

import { makeStyles } from "@material-ui/core/styles";
import MaterialTable from "@material-ui/core/Table";
import TableBody from "@material-ui/core/TableBody";
import TableCell from "@material-ui/core/TableCell";
import TableContainer from "@material-ui/core/TableContainer";
import TableHead from "@material-ui/core/TableHead";
import TableRow from "@material-ui/core/TableRow";
import { TableSortLabel } from "@material-ui/core";
import Paper from "@material-ui/core/Paper";

import { rpc, humanizeDuration, calculateDate } from "../Utils";

import {
  DataAggregatorGetTimeRequest,
  DataAggregatorGetTimeResponse,
} from "productimon/proto/svc/aggregator_pb";
import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Interval, Timestamp } from "productimon/proto/common/common_pb";

const useStyles = makeStyles({
  table: {
    minWidth: 350,
  },
});

var id = 0;
function createData(program, hours, label) {
  id += 1;
  return { program, hours, label, id };
}

export default function Table(props) {
  const classes = useStyles();
  const history = useHistory();

  const [rows, setRows] = React.useState([createData("init", 1, 3)]);

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
    /* Get time data for all device and all interval */
    request.setDevicesList([]);
    request.setIntervalsList([interval]);
    request.setGroupBy(DataAggregatorGetTimeRequest.GroupBy.APPLICATION);

    rpc(DataAggregator.GetTime, history, {
      onEnd: ({ status, statusMessage, headers, message }) => {
        setRows(
          message
            .getDataList()[0]
            .getDataList()
            .sort((a, b) => b.getTime() - a.getTime())
            .map((data) =>
              createData(
                data.getApp(),
                humanizeDuration(data.getTime() / 10 ** 9),
                data.getLabel()
              )
            )
        );
      },
      request,
    });
  }, [props.graphSpec]);

  // TODO enable sort table by col
  return (
    <TableContainer component={Paper}>
      <MaterialTable className={classes.table}>
        <TableHead>
          <TableRow>
            <TableCell>
              <TableSortLabel>Program Name</TableSortLabel>
            </TableCell>
            <TableCell>Time Spent</TableCell>
            <TableCell>Labels</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.id}>
              <TableCell>{row.program}</TableCell>
              <TableCell>{row.hours}</TableCell>
              <TableCell>{row.label}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </MaterialTable>
    </TableContainer>
  );
}
