import React, { useEffect } from 'react';
import { makeStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import { TableSortLabel } from '@material-ui/core';
import Paper from '@material-ui/core/Paper';
import Title from './Title';

import { grpc } from '@improbable-eng/grpc-web';
import { DataAggregatorGetTimeRequest, DataAggregatorGetTimeResponse } from 'productimon/proto/svc/aggregator_pb'
import { DataAggregator } from 'productimon/proto/svc/aggregator_pb_service'
import { Interval, Timestamp } from 'productimon/proto/common/common_pb'

import moment from 'moment'

/* format a timediff in nanoseconds to readable string */
function humanizeDuration(nanoseconds) {
  const duration = moment.duration(nanoseconds / 10 ** 6);
  if (nanoseconds < 60 * 10 ** 9) {
    return `${duration.seconds()} seconds`;
  }
  return duration.humanize();
}

const useStyles = makeStyles({
  table: {
    minWidth: 650,
  },
});

var id = 0;
function createData(program, hours, label) {
  id += 1;
  return { program, hours, label, id };
}

export default function DisplayTable() {
  const classes = useStyles();

  var request = new DataAggregatorGetTimeRequest();
  const [rows, setRows] = React.useState([createData("init",1 ,3)]);


  useEffect(() => {
    const interval = new Interval();
    const start = new Timestamp();
    start.setNanos(0);
    const end = new Timestamp();
    end.setNanos(new Date().getTime() * 10 ** 6);
    interval.setStart(start);
    interval.setEnd(end);

    /* Get time data for all device and all interval */
    request.setDevicesList([]);
    request.setIntervalsList([interval]);
    request.setGroupBy(DataAggregatorGetTimeRequest.GroupBy.APPLICATION);

    const token = window.localStorage.getItem("token");
    grpc.unary(DataAggregator.GetTime, {
      host: '/rpc',
      metadata: new grpc.Metadata({"Authorization": token}),
      onEnd: ({status, statusMessage, headers, message}) => {
        if (status != 0) {
          console.error(`Error getting res, status ${status}: ${statusMessage}`);
          return;
        }
        setRows(message.getDataList()[0].getDataList()
          .sort((a, b) => b.getTime() - a.getTime())
          .map(data =>
          createData(data.getApp(), humanizeDuration(data.getTime()), data.getLabel())));
      },
      request,
    });
  }, []);


  // TODO enable sort table by col
  return (
    <React.Fragment>
      <Title>Total Time Data</Title>
      <TableContainer component={Paper}>
        <Table className={classes.table}>
          <TableHead>
            <TableRow>
              <TableCell>
                <TableSortLabel>
                  Program Name
                </TableSortLabel>
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
        </Table>
      </TableContainer>
    </React.Fragment>
  );
}

