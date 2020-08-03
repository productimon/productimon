import React, { useEffect, useState } from "react";
import {
  PieChart as RechartsPieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Tooltip,
  Legend,
} from "recharts";
import { useSnackbar } from "notistack";

import { makeStyles } from "@material-ui/core/styles";
import FormGroup from "@material-ui/core/FormGroup";
import FormLabel from "@material-ui/core/FormLabel";
import FormControl from "@material-ui/core/FormControl";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import Select from "@material-ui/core/Select";
import MenuItem from "@material-ui/core/MenuItem";

import { DataAggregatorGetTimeRequest } from "productimon/proto/svc/aggregator_pb";
import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Interval, Timestamp } from "productimon/proto/common/common_pb";

import {
  rpc,
  getLabelColor,
  humanizeDuration,
  toSec,
  calculateDate,
} from "../Utils";

const DISPLAY_LABEL_THRESHOLD = 0.03;

// TODO refactor, samething in histogram
// TODO groupBy is likely a generic options to all of the graphs
// put it in the customise form and make it so
const useStyles = makeStyles((theme) => ({
  formBox: {
    justifyContent: "center",
  },
  select: {
    margin: theme.spacing(1),
    minWidth: 120,
  },
  formControl: {
    margin: theme.spacing(3),
  },
  center: {
    textAlign: "center",
  },
}));

function createData(symbol, time) {
  return { symbol, time, humanizedTime: humanizeDuration(time) };
}

export default function PieChart({ graphSpec, options, fullscreen, onUpdate }) {
  const [sectors, setSectors] = React.useState([]);
  const [totalTime, setTotalTime] = React.useState(0);

  const classes = useStyles();
  const { enqueueSnackbar, closeSnackbar } = useSnackbar();

  const localGroupBy = graphSpec.groupBy || "label";
  const numItems = graphSpec.numItems || 5;

  const handleChange = (e) => {
    const newGraphSpec = {
      ...graphSpec,
      [e.target.name]: e.target.value,
    };
    onUpdate(newGraphSpec);
  };

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
    request.setGroupBy(
      localGroupBy === "application"
        ? DataAggregatorGetTimeRequest.GroupBy.APPLICATION
        : DataAggregatorGetTimeRequest.GroupBy.LABEL
    );

    rpc(DataAggregator.GetTime, request)
      .then((res) => {
        const getSymbol =
          graphSpec.groupBy === "application"
            ? (point) => point.getApp()
            : (point) => point.getLabel();

        // Sort data by most used applications
        const sorted = res
          .getDataList()[0]
          .getDataList()
          .sort((a, b) => b.getTime() - a.getTime());

        setTotalTime(
          toSec(sorted.reduce((sum, point) => sum + point.getTime(), 0))
        );

        const entries = sorted
          .slice(0, numItems)
          .map((data) => createData(getSymbol(data), toSec(data.getTime())));
        if (sorted.length > numItems) {
          const other = sorted
            .slice(numItems)
            .reduce(
              (ret, point) =>
                createData("Other", (ret.time || 0) + toSec(point.getTime())),
              {}
            );
          setSectors([...entries, other]);
        } else {
          setSectors(entries);
        }
      })
      .catch((err) => {
        enqueueSnackbar(err, { variant: "error" });
      });
  }, [graphSpec]);

  return (
    <React.Fragment>
      <ResponsiveContainer height="100%">
        <RechartsPieChart>
          <Pie
            innerRadius="40%"
            outerRadius="60%"
            data={sectors}
            dataKey="time"
            nameKey="symbol"
            label={({ symbol, time, humanizedTime }) =>
              time / totalTime > DISPLAY_LABEL_THRESHOLD
                ? `${symbol}: ${humanizedTime}`
                : null
            }
            labelLine={false}
          >
            {sectors.map((data, index) => (
              <Cell key={index} fill={getLabelColor(data.symbol)} />
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
      {options && (
        <FormControl component="fieldset" className={classes.formControl}>
          <FormLabel className={classes.center} component="legend">
            Piechart Options
          </FormLabel>
          <FormGroup className={classes.formBox} row>
            <FormControlLabel
              labelPlacement="start"
              control={
                <Select
                  value={localGroupBy || ""}
                  name="groupBy"
                  onChange={handleChange}
                  className={classes.select}
                >
                  <MenuItem value="label">Label</MenuItem>
                  <MenuItem value="application">Application</MenuItem>
                </Select>
              }
              label="Group by"
            />
          </FormGroup>
        </FormControl>
      )}
    </React.Fragment>
  );
}
