import React from "react";
import clsx from "clsx";
import { useHistory } from "react-router-dom";

import { makeStyles } from "@material-ui/core/styles";
import ExpansionPanel from "@material-ui/core/ExpansionPanel";
import ExpansionPanelSummary from "@material-ui/core/ExpansionPanelSummary";
import ExpansionPanelDetails from "@material-ui/core/ExpansionPanelDetails";
import Typography from "@material-ui/core/Typography";
import ExpandMoreIcon from "@material-ui/icons/ExpandMore";
import MenuItem from "@material-ui/core/MenuItem";
import TextField from "@material-ui/core/TextField";
import FormControl from "@material-ui/core/FormControl";
import Select from "@material-ui/core/Select";
import Box from "@material-ui/core/Box";
import Button from "@material-ui/core/Button";

import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Empty } from "productimon/proto/common/common_pb";

import { rpc, timeUnits, calculateDate } from "../Utils";
import Graph, { graphTypes } from "./Graph";

const useStyles = makeStyles((theme) => ({
  root: {
    width: "100%",
  },
  heading: {
    fontSize: theme.typography.pxToRem(15),
    fontWeight: theme.typography.fontWeightRegular,
  },
  formControl: {
    margin: theme.spacing(1),
    width: "15ch",
  },
  margin: {
    margin: theme.spacing(1),
  },
  textField: {
    width: "10ch",
  },
  wideTextField: {
    width: "25ch",
  },
  container: {
    display: "flex",
    flexDirection: "row",
    justifyContent: "center",
  },
  form: {
    display: "grid",
    backgroundColor: "AliceBlue",
    padding: "20px",
  },
  graphBox: {
    flex: 1,
    marginLeft: "50px",
  },
  fixedHeight: {
    height: 600, // TODO should be less when we remove the graph-specific options from the form
  },
}));

function validateGraphSpec(graphSpec) {
  return (
    graphSpec.startTimeUnit &&
    graphSpec.endTimeUnit &&
    graphSpec.startTimeVal >= 0 &&
    graphSpec.endTimeVal >= 0 &&
    calculateDate(graphSpec.startTimeUnit, graphSpec.startTimeVal) <
      calculateDate(graphSpec.endTimeUnit, graphSpec.endTimeVal) &&
    (!graphSpec.intervals ||
      (graphSpec.intervals > 0 && graphSpec.intervals <= 500)) &&
    (!graphSpec.numItems ||
      (graphSpec.numItems > 0 && graphSpec.numItems <= 10))
  );
}

export default function DashboardCustomizer(props) {
  const classes = useStyles();
  // TODO extract graph-specific options to their own component
  // to avoid these props
  // can auto deduce an appropriate interval for histogram by checking
  // the units
  const graphSpecificProps = {
    histogram: { intervals: true },
    piechart: { numItems: true },
    table: {},
  };

  // TODO change expansionpanel to tabs so that we don't have to deal with multiple graphSpec at once
  // currently there's only one copy of graphSpec shared by multiple forms
  const [graphSpec, setGraphSpec] = React.useState({
    graphType: "",
    graphTitle: "New graph",
    startTimeUnit: "Minutes",
    startTimeVal: "1",
    endTimeUnit: "Seconds",
    endTimeVal: "0",
    intervals: "5",
    numItems: "3",
    device: "all",
  });

  const updateGraph = (newGraphSpec) => {
    setGraphSpec(newGraphSpec);
  };

  return (
    <div className={classes.root}>
      {Object.entries(graphTypes).map(([graphType, { heading, render }]) => (
        <ExpansionPanel key={graphType}>
          <ExpansionPanelSummary expandIcon={<ExpandMoreIcon />}>
            <Typography className={classes.heading}>{heading}</Typography>
          </ExpansionPanelSummary>
          <ExpansionPanelDetails className={classes.fixedHeight}>
            <SimpleForm
              {...graphSpecificProps[graphType]}
              onAdd={props.onAdd}
              graphSpec={{ ...graphSpec, graphType: graphType }}
              setGraphSpec={setGraphSpec}
            />
            <div className={classes.graphBox}>
              {validateGraphSpec(graphSpec) ? (
                <Graph
                  graphSpec={{ ...graphSpec, graphType: graphType }}
                  onUpdate={updateGraph}
                  preview
                  options
                />
              ) : (
                <Typography>
                  No preview at this point, check your input options
                </Typography>
              )}
            </div>
          </ExpansionPanelDetails>
        </ExpansionPanel>
      ))}
    </div>
  );
}

function TimeUnitSelect(props) {
  return (
    <Select value={props.value} onChange={props.onChange} name={props.name}>
      <MenuItem value="">
        <em>None</em>
      </MenuItem>
      {Object.entries(timeUnits).map(([unit, _], index) => (
        <MenuItem value={unit} key={index}>
          {unit}
        </MenuItem>
      ))}
    </Select>
  );
}

function SimpleForm(props) {
  const classes = useStyles();
  const { graphSpec, setGraphSpec } = props;
  const history = useHistory();

  const [deviceList, setDeviceList] = React.useState([]);

  React.useEffect(() => {
    rpc(DataAggregator.GetDevices, new Empty()).then((res) => {
      setDeviceList(res.getDevicesList());
    });
  }, []);

  const handleInputChange = (event) => {
    setGraphSpec({ ...graphSpec, [event.target.name]: event.target.value });
  };

  return (
    <div className={classes.form}>
      <div className={classes.container}>
        <Box my={3}>Title:</Box>
        <TextField
          className={clsx(classes.margin, classes.wideTextField)}
          value={graphSpec.graphTitle}
          variant="filled"
          onChange={handleInputChange}
          name="graphTitle"
          value={graphSpec.graphTitle}
        />
      </div>
      <div className={classes.container}>
        <Box my={3}>Start:</Box>
        <TextField
          className={clsx(classes.margin, classes.textField)}
          value={graphSpec.startTimeVal}
          variant="filled"
          onChange={handleInputChange}
          name="startTimeVal"
          value={graphSpec.startTimeVal}
        />
        <FormControl variant="filled" className={classes.formControl}>
          <TimeUnitSelect
            value={graphSpec.startTimeUnit}
            onChange={handleInputChange}
            name="startTimeUnit"
          />
        </FormControl>
        <Box my={3}>Ago</Box>
      </div>
      <div className={classes.container}>
        <Box my={3}>End:</Box>
        <TextField
          className={clsx(classes.margin, classes.textField)}
          value={graphSpec.endTimeVal}
          variant="filled"
          onChange={handleInputChange}
          name="endTimeVal"
          value={graphSpec.endTimeVal}
        />
        <FormControl variant="filled" className={classes.formControl}>
          <TimeUnitSelect
            value={graphSpec.endTimeUnit}
            onChange={handleInputChange}
            name="endTimeUnit"
          />
        </FormControl>
        <Box my={3}>Ago</Box>
      </div>
      {props.intervals && (
        <div className={classes.container}>
          <Box my={3}>Intervals:</Box>
          <TextField
            className={clsx(classes.margin, classes.textField)}
            variant="filled"
            value={graphSpec.intervals}
            onChange={handleInputChange}
            name="intervals"
          />
        </div>
      )}
      {props.numItems && (
        <div className={classes.container}>
          <Box my={3}>Num items:</Box>
          <TextField
            className={clsx(classes.margin, classes.textField)}
            variant="filled"
            value={graphSpec.numItems}
            onChange={handleInputChange}
            name="numItems"
          />
        </div>
      )}
      <div className={classes.container}>
        <Box my={3}>Devices:</Box>
        <FormControl variant="filled" className={classes.formControl}>
          <Select
            name="device"
            value={graphSpec.device || ""}
            onChange={handleInputChange}
          >
            <MenuItem value="all">
              <em>All</em>
            </MenuItem>
            {deviceList.map((d) => (
              <MenuItem key={d.getId()} value={`device-${d.getId()}`}>
                <em>{d.getName()}</em>
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </div>
      <Button
        variant="contained"
        style={{ marginTop: "20px" }}
        disabled={!validateGraphSpec(graphSpec)}
        color="primary"
        onClick={() => {
          props.onAdd(graphSpec);
          history.push("/dashboard");
        }}
      >
        Add to Dashboard
      </Button>
    </div>
  );
}
