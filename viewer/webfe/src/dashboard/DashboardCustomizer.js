import React from "react";
import clsx from "clsx";
import { makeStyles } from "@material-ui/core/styles";
import ExpansionPanel from "@material-ui/core/ExpansionPanel";
import ExpansionPanelSummary from "@material-ui/core/ExpansionPanelSummary";
import ExpansionPanelDetails from "@material-ui/core/ExpansionPanelDetails";
import Typography from "@material-ui/core/Typography";
import ExpandMoreIcon from "@material-ui/icons/ExpandMore";
import InputLabel from "@material-ui/core/InputLabel";
import MenuItem from "@material-ui/core/MenuItem";
import FormHelperText from "@material-ui/core/FormHelperText";
import IconButton from "@material-ui/core/IconButton";
import Input from "@material-ui/core/Input";
import FilledInput from "@material-ui/core/FilledInput";
import OutlinedInput from "@material-ui/core/OutlinedInput";
import InputAdornment from "@material-ui/core/InputAdornment";
import TextField from "@material-ui/core/TextField";
import FormControl from "@material-ui/core/FormControl";
import Select from "@material-ui/core/Select";
import Box from "@material-ui/core/Box";
import Button from "@material-ui/core/Button";
import { timeUnits } from "./utils";

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
}));

export default function DashboardCustomizer(props) {
  const classes = useStyles();

  return (
    <div className={classes.root}>
      <ExpansionPanel>
        <ExpansionPanelSummary expandIcon={<ExpandMoreIcon />}>
          <Typography className={classes.heading}>Histogram</Typography>
        </ExpansionPanelSummary>
        <ExpansionPanelDetails>
          <SimpleForm
            intervals={true}
            onAdd={props.onAdd}
            graphType="histogram"
          />
        </ExpansionPanelDetails>
      </ExpansionPanel>
      <ExpansionPanel>
        <ExpansionPanelSummary expandIcon={<ExpandMoreIcon />}>
          <Typography className={classes.heading}>Piechart</Typography>
        </ExpansionPanelSummary>
        <ExpansionPanelDetails>
          <SimpleForm
            numItems={true}
            onAdd={props.onAdd}
            graphType="piechart"
          />
        </ExpansionPanelDetails>
      </ExpansionPanel>
      <ExpansionPanel>
        <ExpansionPanelSummary expandIcon={<ExpandMoreIcon />}>
          <Typography className={classes.heading}>Table</Typography>
        </ExpansionPanelSummary>
        <ExpansionPanelDetails>
          <SimpleForm onAdd={props.onAdd} graphType="table" />
        </ExpansionPanelDetails>
      </ExpansionPanel>
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
  const [graphSpec, setGraphSpec] = React.useState({
    graphType: props.graphType,
    graphTitle: "",
    startTimeUnit: "",
    startTimeVal: "",
    endTimeUnit: "",
    endTimeVal: "",
    intervals: "",
    numItems: "",
  });

  const handleInputChange = (event) => {
    setGraphSpec({ ...graphSpec, [event.target.name]: event.target.value });
  };

  return (
    <div className={classes.form}>
      <div className={classes.container}>
        <Box my={3}>Title:</Box>
        <TextField
          className={clsx(classes.margin, classes.wideTextField)}
          variant="filled"
          onChange={handleInputChange}
          name="graphTitle"
        />
      </div>
      <div className={classes.container}>
        <Box my={3}>Start:</Box>
        <TextField
          className={clsx(classes.margin, classes.textField)}
          variant="filled"
          onChange={handleInputChange}
          name="startTimeVal"
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
          variant="filled"
          onChange={handleInputChange}
          name="endTimeVal"
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
          <Select value="">
            <MenuItem value="">
              <em>All</em>
            </MenuItem>
          </Select>
        </FormControl>
      </div>
      <div className={classes.container}>
        <Box my={3}>Categories:</Box>
        <FormControl variant="filled" className={classes.formControl}>
          <Select value="">
            <MenuItem value="">
              <em>All</em>
            </MenuItem>
          </Select>
        </FormControl>
      </div>
      <Button
        variant="contained"
        style={{ marginTop: "20px" }}
        color="primary"
        onClick={() => {
          props.onAdd(graphSpec);
        }}
      >
        Add to Dashboard
      </Button>
    </div>
  );
}
