import React, { useEffect, useState } from "react";
import { useHistory } from "react-router-dom";
import moment from "moment";
import {
  BarChart,
  Bar,
  CartesianGrid,
  XAxis,
  YAxis,
  ResponsiveContainer,
  Tooltip,
  Legend,
  Label,
  ReferenceLine,
} from "recharts";

import { makeStyles } from "@material-ui/core/styles";
import FormGroup from "@material-ui/core/FormGroup";
import FormLabel from "@material-ui/core/FormLabel";
import FormControl from "@material-ui/core/FormControl";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import Switch from "@material-ui/core/Switch";
import Select from "@material-ui/core/Select";
import InputLabel from "@material-ui/core/InputLabel";
import MenuItem from "@material-ui/core/MenuItem";

import {
  DataAggregatorGetTimeRequest,
  DataAggregatorGetTimeResponse,
} from "productimon/proto/svc/aggregator_pb";
import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Interval, Timestamp } from "productimon/proto/common/common_pb";

import { getLabelColor, timeUnits, calculateDate, rpc } from "../Utils";

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
  let dateFmt = "DD-MM-YY";
  if (totalSpan <= timeUnits.Days) {
    dateFmt = "";
  } else if (totalSpan < timeUnits.Years) {
    dateFmt = "DD-MM";
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
 * getSymbol is the function to retrive the symbol out of
 * the data point passed to the returned mapping function
 */
function transformRange(allData, getSymbol) {
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

    return data.getDataList().reduce(
      (ret, point) => {
        const total = ret.Total + Math.floor(point.getTime() / 10 ** 6);
        const active = ret.Active + Math.floor(point.getActivetime() / 10 ** 6);
        return {
          ...ret,
          Total: total,
          [getSymbol(point)]: Math.floor(point.getTime() / 10 ** 6),
          Active: active,
          "Non-active": total - active,
        };
      },
      {
        label: formatNanoTS(
          data.getInterval().getStart().getNanos(),
          timeSpan,
          totalTimeSpan
        ),
        Total: 0,
        Active: 0,
      }
    );
  };
}

function getUniqSymbols(response, getSymbol) {
  return response
    .getDataList()
    .reduce(
      (result, range) => [
        ...result,
        range
          .getDataList()
          .reduce((labels, datapoint) => [...labels, getSymbol(datapoint)], []),
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
  const [unitLabel, setUnitLabel] = useState("");
  const history = useHistory();

  const classes = useStyles();

  useEffect(() => {
    const startDate = calculateDate(
      props.graphSpec.startTimeUnit,
      props.graphSpec.startTimeVal
    );
    const endDate = calculateDate(
      props.graphSpec.endTimeUnit,
      props.graphSpec.endTimeVal
    );
    const numIntervals = props.graphSpec.intervals;
    const intervals = createIntervals(startDate, endDate, numIntervals);

    const request = new DataAggregatorGetTimeRequest();
    request.setDevicesList([]);
    request.setIntervalsList(intervals);
    request.setGroupBy(
      props.graphSpec.stack === "application"
        ? DataAggregatorGetTimeRequest.GroupBy.APPLICATION
        : DataAggregatorGetTimeRequest.GroupBy.LABEL
    );

    rpc(DataAggregator.GetTime, history, {
      onEnd: ({ status, statusMessage, headers, message }) => {
        // get all symbols (label/application)
        const getSymbol =
          props.graphSpec.stack === "application"
            ? (point) => point.getApp()
            : (point) => point.getLabel();

        // decide what to stack on the bars
        let displayedKeys;
        switch (props.graphSpec.stack) {
          case "application":
          case "label":
            displayedKeys = getUniqSymbols(message, getSymbol);
            break;
          case "active":
            displayedKeys = ["Active", "Non-active"];
            break;
          default:
            displayedKeys = ["Total"];
            break;
        }
        setDataKeys(displayedKeys);

        // format and aggregate time values
        const data = message
          .getDataList()
          .map(transformRange(message.getDataList(), getSymbol))
          .reverse();

        // adaptive unit for y-axis
        // get the fisrt unit less than the largest unit that can cover our max timeval
        const [unit, factor] = Object.entries(timeUnits)
          .reverse()
          .find(
            ([unit, factor]) =>
              factor < Math.max(...data.map((ent) => ent.Total), 0) ||
              unit == "Seconds"
          );
        setUnitLabel(unit);

        // normalise the displayedKeys and Total (which is used for the Average overlay)
        setData(
          data.map((ent) => ({
            ...ent,
            ...["Total", ...displayedKeys].reduce(
              (obj, key) => ({
                ...obj,
                [key]: parseFloat((ent[key] / factor).toFixed(2)) || 0,
              }),
              {}
            ),
          }))
        );
      },
      request,
    });
  }, [props.graphSpec]);

  const handleCheckbox = (e) => {
    const newGraphSpec = {
      ...props.graphSpec,
      [e.target.name]: e.target.checked,
    };
    props.onUpdate(newGraphSpec);
  };
  const handleChange = (e) => {
    const newGraphSpec = {
      ...props.graphSpec,
      [e.target.name]: e.target.value,
    };
    props.onUpdate(newGraphSpec);
  };

  return (
    <React.Fragment>
      <ResponsiveContainer style={{ flexShrink: 1 }}>
        <BarChart
          data={data}
          margin={{
            top: 16,
            right: props.graphSpec.average ? 60 : 16,
            bottom: 0,
            left: 16,
          }}
          barCategoryGap={props.fullscreen ? "20%" : "10%"}
        >
          {/* TODO have adaptive unit for the tooltip label too */}
          {props.fullscreen && <Tooltip />}
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="label" />
          <YAxis>
            <Label value={unitLabel} angle={-90} position="insideLeft" />
          </YAxis>
          <Legend />
          {dataKeys.map((label, index) => (
            <Bar
              key={index}
              dataKey={label}
              stackId="a"
              fill={getLabelColor(label)}
            />
          ))}
          {props.graphSpec.average && data.length && (
            <ReferenceLine
              y={
                data.map((ent) => ent.Total).reduce((a, b) => a + b, 0) /
                data.length
              }
              stroke="rgba(0, 0, 0, 0.5)"
              strokeDasharray="3 3"
            >
              <Label value="Average" position="right" />
            </ReferenceLine>
          )}
        </BarChart>
      </ResponsiveContainer>
      {props.options && (
        <FormControl component="fieldset" className={classes.formControl}>
          <FormLabel className={classes.center} component="legend">
            Overlay options
          </FormLabel>
          <FormGroup className={classes.formBox} row>
            <FormControlLabel
              labelPlacement="start"
              control={
                <Switch
                  checked={Boolean(props.graphSpec.average)}
                  onChange={handleCheckbox}
                  name="average"
                  color="primary"
                />
              }
              label="Average overlay"
            />
            <FormControlLabel
              labelPlacement="start"
              control={
                <Select
                  value={props.graphSpec.stack || ""}
                  name="stack"
                  onChange={handleChange}
                  className={classes.select}
                >
                  <MenuItem value="">None</MenuItem>
                  <MenuItem value="label">Label</MenuItem>
                  <MenuItem value="application">Application</MenuItem>
                  <MenuItem value="active">Active time</MenuItem>
                </Select>
              }
              label="Stack"
            />
          </FormGroup>
        </FormControl>
      )}
    </React.Fragment>
  );
}
