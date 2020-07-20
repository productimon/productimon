import React from "react";
import clsx from "clsx";
import { useSnackbar } from "notistack";
import { useHistory } from "react-router-dom";

import { makeStyles } from "@material-ui/core/styles";
import MenuItem from "@material-ui/core/MenuItem";
import TextField from "@material-ui/core/TextField";
import FormControl from "@material-ui/core/FormControl";
import Radio from "@material-ui/core/Radio";
import RadioGroup from "@material-ui/core/RadioGroup";
import Checkbox from "@material-ui/core/Checkbox";
import FormControlLabel from "@material-ui/core/FormControlLabel";
import FormLabel from "@material-ui/core/FormLabel";
import Box from "@material-ui/core/Box";
import Button from "@material-ui/core/Button";
import IconButton from "@material-ui/core/IconButton";
import Tooltip from "@material-ui/core/Tooltip";
import Select from "@material-ui/core/Select";
import Fab from "@material-ui/core/Fab";
import InfoIcon from "@material-ui/icons/Info";
import Typography from "@material-ui/core/Typography";
import Container from "@material-ui/core/Container";

import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Goal, Interval, Timestamp } from "productimon/proto/common/common_pb";

import { rpc, timeUnits, nowNano, parseDateTime } from "../../Utils";

const useStyles = makeStyles((theme) => ({
  root: {
    width: "100%",
    display: "flex",
    justifyContent: "center",
  },
  row: {
    display: "flex",
    flexDirection: "row",
    justifyContent: "left",
    marginTop: theme.spacing(1),
    marginBottom: theme.spacing(1),
  },
  indent: {
    marginLeft: theme.spacing(1),
  },
  form: {
    display: "flex",
    flexFlow: "column",
    backgroundColor: "AliceBlue",
    padding: "20px",
  },
  typography: {
    subtitle1: {
      fontSize: 4,
    },
  },
  textField: {
    marginLeft: theme.spacing(1),
    marginRight: theme.spacing(1),
    flex: "1 0 0",
  },
  graphBox: {
    flex: 1,
    marginLeft: "50px",
  },
  tooltip: {
    marginLeft: "10px",
  },
}));

function createInterval(startNano, endNano) {
  const interval = new Interval();
  const startTS = new Timestamp();
  const endTS = new Timestamp();
  startTS.setNanos(startNano);
  endTS.setNanos(endNano);
  interval.setStart(startTS);
  interval.setEnd(endTS);
  return interval;
}

function createGoalRequest(goalSpec) {
  const goalReq = new Goal();

  goalReq.setTitle(goalSpec.title);
  goalReq.setType(goalSpec.type);

  if (goalSpec.itemType === "app") goalReq.setApplication(goalSpec.item);
  else goalReq.setLabel(goalSpec.item);

  const sign = (function () {
    if (goalSpec.amountType === "absolute") return 1;
    if (goalSpec.deltaType === "reduce") return -1;
    return 1;
  })();

  if (goalSpec.amountType === "percent") {
    goalReq.setPercentamount((sign * goalSpec.amount) / 100);
  } else {
    goalReq.setFixedamount(
      sign * goalSpec.amount * timeUnits[goalSpec.amountUnit] * 10 ** 6
    );
  }

  const goalInterval = createInterval(
    parseDateTime(goalSpec.startDate, goalSpec.startTime),
    parseDateTime(goalSpec.endDate, goalSpec.endTime)
  );
  goalReq.setGoalinterval(goalInterval);

  if (goalSpec.amountType != "absolute") {
    let goalInterval;
    if (goalSpec.compareBy == "dates") {
      goalInterval = createInterval(
        parseDateTime(goalSpec.compareStartDate, goalSpec.compareStartTime),
        parseDateTime(goalSpec.compareEndDate, goalSpec.compareEndTime)
      );
    } else {
      const now = nowNano();
      goalInterval = createInterval(
        now -
          goalSpec.compareTimeLength *
            timeUnits[goalSpec.compareUnit] *
            10 ** 6,
        now
      );
    }
    goalReq.setCompareinterval(goalInterval);
  }

  goalReq.setCompareequalized(goalSpec.compareEqualized == "true");

  return goalReq;
}

// returns error message if not valid
function validateGoalSpec(goalSpec) {
  if (!goalSpec.title) return "Invalid title";
  if (!goalSpec.item) return "Invalid app/label";

  if (goalSpec.amountType != "absolute") {
    if (goalSpec.compareBy == "dates") {
      const compareStart = parseDateTime(
        goalSpec.compareStartDate,
        goalSpec.compareStartTime
      );
      const compareEnd = parseDateTime(
        goalSpec.compareEndDate,
        goalSpec.compareEndTime
      );
      if (!compareStart || !compareEnd || compareEnd <= compareStart)
        return "Invalid comparison start/end time";
      if (compareEnd >= nowNano() || compareEnd >= goalStart)
        return "Comparison interval must be prior to goal starting time and current time";
    } else if (!(goalSpec.compareTimeLength > 0)) {
      return "Invalid compare interval length";
    }
  }

  if (!(goalSpec.amount > 0)) return "Invalid amount";

  if (
    goalSpec.amountType == "percent" &&
    !(goalSpec.amount > 0 && goalSpec.amount <= 100)
  ) {
    return "Percentage must be betwen 1 and 100";
  }

  const goalStart = parseDateTime(goalSpec.startDate, goalSpec.startTime);
  const goalEnd = parseDateTime(goalSpec.endDate, goalSpec.endTime);
  if (!goalSpec || !goalEnd || goalEnd <= goalStart)
    return "Invalid goal start/end time";

  return null;
}

export default function AddGoal(props) {
  const classes = useStyles();
  const [goalSpec, setGoalSpec] = React.useState({
    title: "",
    amountType: "percent",
    amountUnit: "Minutes",
    type: "aspiring",
    itemType: "app",
    item: "",
    deltaType: "reduce",
    compareEqualized: "false",
    compareBy: "length",
    compareUnit: "Weeks",
    startTime: "00:00",
    endTime: "00:00",
    compareStartTime: "00:00",
    compareEndTime: "00:00",
  });

  return (
    <div className={classes.root}>
      <Form {...props} goalSpec={goalSpec} setGoalSpec={setGoalSpec} />
    </div>
  );
}

function Form(props) {
  const classes = useStyles();
  const { goalSpec, setGoalSpec } = props;

  const handleInputChange = (event) => {
    setGoalSpec({ ...goalSpec, [event.target.name]: event.target.value });
  };

  const { enqueueSnackbar } = useSnackbar();
  const history = useHistory();

  return (
    <div className={classes.form}>
      <FormControl className={classes.row} component="fieldset">
        <FormLabel component="legend">Select Goal Type</FormLabel>
        <RadioGroup
          row
          name="type"
          value={goalSpec.type}
          onChange={handleInputChange}
        >
          <FormControlLabel
            value="aspiring"
            control={<Radio color="primary" />}
            label="Aspiring"
            labelPlacement="start"
          />
          <FormControlLabel
            value="limiting"
            control={<Radio color="primary" />}
            label="Limiting"
            labelPlacement="start"
          />
          <Tooltip
            className={classes.tooltip}
            title="An aspiring goal is one which the user desires to reach, a limiting goal is one which user desires to limit usage"
            placement="right"
          >
            <IconButton>
              <InfoIcon />
            </IconButton>
          </Tooltip>
        </RadioGroup>
      </FormControl>
      <div className={classes.row}>
        <TextField
          className={classes.textField}
          onChange={handleInputChange}
          name="title"
          label="Goal Title"
          variant="outlined"
        />
      </div>
      <FormControl className={classes.row} component="fieldset">
        <FormLabel component="legend">Item Type</FormLabel>
        <RadioGroup
          row
          name="itemType"
          value={goalSpec.itemType}
          onChange={handleInputChange}
        >
          <FormControlLabel
            value="app"
            control={<Radio color="primary" />}
            label="Application"
            labelPlacement="start"
          />
          <FormControlLabel
            value="label"
            control={<Radio color="primary" />}
            label="Label"
            labelPlacement="start"
          />
          <Tooltip
            className={classes.tooltip}
            title="Whether this goals is related to an application or a label"
            placement="right"
          >
            <IconButton>
              <InfoIcon />
            </IconButton>
          </Tooltip>
        </RadioGroup>
      </FormControl>
      <div className={classes.row}>
        <TextField
          className={classes.textField}
          onChange={handleInputChange}
          name="item"
          label={
            goalSpec.itemType === "app" ? "Goal Application" : "Goal Label"
          }
          variant="outlined"
        />
      </div>

      <FormControl className={classes.row} component="fieldset">
        <FormLabel component="legend">Amount type</FormLabel>
        <RadioGroup
          row
          name="amountType"
          value={goalSpec.amountType}
          onChange={handleInputChange}
        >
          <FormControlLabel
            value="percent"
            control={<Radio color="primary" />}
            label="Compare Percent"
            labelPlacement="start"
          />
          <FormControlLabel
            value="amount"
            control={<Radio color="primary" />}
            label="Compare Amount"
            labelPlacement="start"
          />
          <FormControlLabel
            value="absolute"
            control={<Radio color="primary" />}
            label="Absolute"
            labelPlacement="start"
          />
          <Tooltip
            className={classes.tooltip}
            title="In a Compare Goal, usage is compared to an interval from the past. In an absolute goal user sets a specific amount to reach/limit"
            placement="right"
          >
            <IconButton>
              <InfoIcon />
            </IconButton>
          </Tooltip>
        </RadioGroup>
      </FormControl>

      {goalSpec.amountType != "absolute" && (
        <React.Fragment>
          <FormControl className={classes.row} component="fieldset">
            <FormLabel component="legend">Set Goal Compare Interval</FormLabel>
            <RadioGroup
              row
              name="compareBy"
              value={goalSpec.compareBy}
              onChange={handleInputChange}
            >
              <FormControlLabel
                value="length"
                control={<Radio color="primary" />}
                label="By Length"
                labelPlacement="start"
              />
              <FormControlLabel
                value="dates"
                control={<Radio color="primary" />}
                label="By Dates"
                labelPlacement="start"
              />
              <Tooltip
                title="In length mode, comparison interval is the set period right before the goal's starting time"
                placement="right"
              >
                <IconButton>
                  <InfoIcon />
                </IconButton>
              </Tooltip>
            </RadioGroup>
          </FormControl>
          {goalSpec.compareBy == "dates" && (
            <div>
              <div className={classes.row}>
                <TextField
                  label="Goal Compare Start"
                  type="date"
                  className={classes.textField}
                  InputLabelProps={{
                    shrink: true,
                  }}
                  name="compareStartDate"
                  value={goalSpec.compareStartDate}
                  onChange={handleInputChange}
                  variant="outlined"
                />
                <TextField
                  label="Time"
                  type="time"
                  className={classes.textField}
                  InputLabelProps={{
                    shrink: true,
                  }}
                  inputProps={{
                    step: 300, // 5 min
                  }}
                  name="compareStartTime"
                  value={goalSpec.compareStartTime}
                  onChange={handleInputChange}
                  variant="outlined"
                />
              </div>
              <div className={classes.row}>
                <TextField
                  label="Goal Compare End"
                  type="date"
                  className={classes.textField}
                  InputLabelProps={{
                    shrink: true,
                  }}
                  name="compareEndDate"
                  value={goalSpec.compareEndDate}
                  onChange={handleInputChange}
                  variant="outlined"
                />
                <TextField
                  label="Time"
                  type="time"
                  className={classes.textField}
                  InputLabelProps={{
                    shrink: true,
                  }}
                  inputProps={{
                    step: 300, // 5 min
                  }}
                  name="compareEndTime"
                  value={goalSpec.compareEndTime}
                  onChange={handleInputChange}
                  variant="outlined"
                />
              </div>
            </div>
          )}
          {goalSpec.compareBy == "length" && (
            <div className={classes.row}>
              <TextField
                className={classes.textField}
                onChange={handleInputChange}
                label="Compare Interval Length"
                variant="outlined"
                name="compareTimeLength"
                value={goalSpec.compareTimeLength || ""}
              />
              <TextField
                select
                className={classes.textField}
                label="Select Time Units"
                onChange={handleInputChange}
                variant="outlined"
                value={goalSpec.compareUnit}
                name="compareUnit"
              >
                {Object.entries(timeUnits).map(([unit, _], index) => (
                  <MenuItem value={unit} key={index}>
                    {unit}
                  </MenuItem>
                ))}
              </TextField>
            </div>
          )}
          <div className={clsx(classes.row, classes.indent)}>
            <FormControlLabel
              control={
                <Checkbox
                  checked={goalSpec.compareEqualized == "true"}
                  onChange={handleInputChange}
                  name="compareEqualized"
                  value={!(goalSpec.compareEqualized == "true")}
                />
              }
              label="Average Compare Data"
            />
            <Tooltip
              title="Average compare interval data to ensure we compare goals with equal time periods"
              placement="right"
            >
              <IconButton>
                <InfoIcon />
              </IconButton>
            </Tooltip>
          </div>
        </React.Fragment>
      )}

      <div className={classes.row}>
        {goalSpec.amountType != "absolute" && (
          <TextField
            select
            className={classes.textField}
            value={goalSpec.deltaType}
            onChange={handleInputChange}
            name="deltaType"
            variant="outlined"
          >
            <MenuItem value="reduce">Reduce</MenuItem>
            <MenuItem value="increase">Increase</MenuItem>
          </TextField>
        )}
        <TextField
          className={classes.textField}
          onChange={handleInputChange}
          name="amount"
          label={(function () {
            switch (goalSpec.amountType) {
              case "percent":
                return "By Percent";
              case "amount":
                return "By Amount";
              case "absolute":
                return "Amount";
            }
          })()}
          variant="outlined"
        />
        {goalSpec.amountType != "percent" && (
          <TextField
            select
            className={classes.textField}
            label="Select Time Units"
            onChange={handleInputChange}
            variant="outlined"
            value={goalSpec.amountUnit}
            name="amountUnit"
          >
            {Object.entries(timeUnits).map(([unit, _], index) => (
              <MenuItem value={unit} key={index}>
                {unit}
              </MenuItem>
            ))}
          </TextField>
        )}
      </div>
      <div className={classes.row}>
        <TextField
          label="Goal Start"
          type="date"
          className={classes.textField}
          InputLabelProps={{
            shrink: true,
          }}
          name="startDate"
          onChange={handleInputChange}
          variant="outlined"
        />
        <TextField
          label="Time"
          type="time"
          className={classes.textField}
          InputLabelProps={{
            shrink: true,
          }}
          inputProps={{
            step: 300, // 5 min
          }}
          name="startTime"
          value={goalSpec.startTime}
          onChange={handleInputChange}
          variant="outlined"
        />
      </div>
      <div className={classes.row}>
        <TextField
          label="Goal End"
          type="date"
          defaultValue=""
          className={classes.textField}
          InputLabelProps={{
            shrink: true,
          }}
          name="endDate"
          onChange={handleInputChange}
          variant="outlined"
        />
        <TextField
          label="Time"
          type="time"
          className={classes.textField}
          InputLabelProps={{
            shrink: true,
          }}
          inputProps={{
            step: 300, // 5 min
          }}
          name="endTime"
          value={goalSpec.endTime}
          onChange={handleInputChange}
          variant="outlined"
        />
      </div>
      <Button
        variant="contained"
        style={{ marginTop: "20px" }}
        color="primary"
        onClick={() => {
          const err = validateGoalSpec(goalSpec);
          if (err) {
            enqueueSnackbar(err, { variant: "error" });
            return;
          }
          console.log(goalSpec);
          const request = createGoalRequest(goalSpec);
          console.log(request);
          rpc(DataAggregator.AddGoal, request)
            .then((res) => {
              console.log(res);
              props.onAdd(goalSpec);
              history.push("/dashboard/goals/view");
            })
            .catch((err) => {
              console.error(err);
              enqueueSnackbar(err, { variant: "error" });
            });
        }}
      >
        Add Goal
      </Button>
    </div>
  );
}
