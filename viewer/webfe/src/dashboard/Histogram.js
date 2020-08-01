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
import TextField from "@material-ui/core/TextField";
import MenuItem from "@material-ui/core/MenuItem";

import {
  DataAggregatorGetTimeRequest,
  DataAggregatorGetTimeResponse,
} from "productimon/proto/svc/aggregator_pb";
import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Interval, Timestamp } from "productimon/proto/common/common_pb";

import {
  rpc,
  getLabelColor,
  timeUnits,
  calculateDate,
  humanizeDuration,
} from "../Utils";

const useStyles = makeStyles((theme) => ({
  formBox: {
    justifyContent: "center",
  },
  select: {
    margin: theme.spacing(1),
    minWidth: 120,
  },
  input: {
    margin: theme.spacing(1),
    maxWidth: 60,
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
function formatMiliTS(ts, span, totalSpan) {
  const m = moment(ts / timeUnits.Seconds, "X");

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
 * getSymbol is the function to retrive the symbol out of
 * the data point passed to the returned mapping function
 */
function transformRange(getSymbol, displayedSymbols) {
  return (data) => {
    const startMili = data.getInterval().getStart().getNanos() / 10 ** 6;
    const endMili = data.getInterval().getEnd().getNanos() / 10 ** 6;

    return data.getDataList().reduce(
      (ret, point) => {
        const time = Math.floor(point.getTime() / 10 ** 6);
        const total = ret.Total + time;
        const active = ret.Active + Math.floor(point.getActivetime() / 10 ** 6);
        const other = {
          Other:
            (ret.Other || 0) +
            (displayedSymbols.includes(getSymbol(point)) ? 0 : time),
        };
        return {
          ...ret,
          Total: total,
          [getSymbol(point)]: time,
          Active: active,
          Inactive: total - active,
          ...other,
        };
      },
      {
        // we need both val for formatting the label in the legend
        // to show the interval in full datetime format
        // but rechart does not allow pass of object using dataKey
        interval: `${startMili}-${endMili}`,
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

function getSortedSymbols(response, getSymbol, maxReturn) {
  const symbolTimes = response.getDataList().reduce(
    (result, range) => ({
      ...result,
      ...range.getDataList().reduce(
        (rangeRet, datapoint) => ({
          ...rangeRet,
          [getSymbol(datapoint)]:
            (result[getSymbol(datapoint)] || 0) + datapoint.getTime(),
        }),
        {}
      ),
    }),
    []
  );

  return Object.entries(symbolTimes)
    .sort(([_, a], [__, b]) => b - a)
    .slice(0, maxReturn)
    .map(([symbol, _]) => symbol);
}

export default function Histogram(props) {
  const [data, setData] = useState([]);
  const [dataKeys, setDataKeys] = useState([]);
  const [unitLabel, setUnitLabel] = useState("");
  const history = useHistory();
  const [totalTimeSpan, setTotalTimeSpan] = useState(0);

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

        const nSymbols = props.fullscreen
          ? props.graphSpec.maxItem || 10
          : props.graphSpec.maxItem && props.graphSpec.maxItem < 5
          ? props.graphSpec.maxItem
          : 5;

        const symbols = getUniqSymbols(message, getSymbol);
        const displayedSymbols = getSortedSymbols(message, getSymbol, nSymbols);

        // decide what to stack on the bars
        let displayedKeys;
        switch (props.graphSpec.stack) {
          case "application":
          case "label":
            displayedKeys =
              symbols.length > nSymbols
                ? [...displayedSymbols, "Other"]
                : displayedSymbols;
            break;
          case "active":
            displayedKeys = ["Active", "Inactive"];
            break;
          default:
            displayedKeys = ["Total"];
            break;
        }
        setDataKeys(displayedKeys);

        const allData = message.getDataList();

        // format and aggregate time values
        const data = allData
          .map(transformRange(getSymbol, displayedSymbols))
          .reverse();

        /* Time span of the whole histogram, assuming backend returns data in reverse chronological order */
        let totalTimeSpan = allData.length
          ? allData[0].getInterval().getEnd().getNanos() -
            allData[allData.length - 1].getInterval().getStart().getNanos()
          : 0;
        totalTimeSpan /= 10 ** 6;
        setTotalTimeSpan(totalTimeSpan);

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
                // for each field we normalise, we preserve the raw value to be used in a formatter for the tooltip
                [`${key}-raw`]: ent[key] || null,
                [key]: parseFloat((ent[key] / factor).toFixed(2)) || null,
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

  // validator is a function that takes in the new value and returns if the new value is valid or not
  const handleChange = (e, validator) => {
    if (validator && !validator(e.target.value)) return;
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
          {props.fullscreen && (
            <Tooltip
              formatter={(_, key, props) => {
                const timeInMili = props.payload[`${key}-raw`];
                return humanizeDuration(timeInMili / timeUnits.Seconds);
              }}
              labelFormatter={(interval) => {
                const [start, end] = interval.split("-");
                const startFull = moment(start / timeUnits.Seconds, "X").format(
                  "LLLL"
                );
                const endFull = moment(end / timeUnits.Seconds, "X").format(
                  "LLLL"
                );
                return `${startFull} to ${endFull}`;
              }}
            />
          )}
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis
            dataKey="interval"
            tickFormatter={(interval) => {
              const [start, end] = interval.split("-");
              return formatMiliTS(start, end - start, totalTimeSpan);
            }}
          />
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
                  onChange={(e) => handleChange(e)}
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
            <FormControlLabel
              labelPlacement="start"
              control={
                <TextField
                  value={props.graphSpec.maxItem || 10}
                  onChange={(e) =>
                    handleChange(e, (val) => val >= 1 && val <= 50)
                  }
                  name="maxItem"
                  className={classes.input}
                />
              }
              label="Max Item"
            />
          </FormGroup>
        </FormControl>
      )}
    </React.Fragment>
  );
}
