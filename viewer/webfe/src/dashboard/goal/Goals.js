import React, { useState, useEffect } from "react";
import { Switch, Route, Link } from "react-router-dom";
import { useSnackbar } from "notistack";

import Container from "@material-ui/core/Container";
import Grid from "@material-ui/core/Grid";
import Paper from "@material-ui/core/Paper";
import Toolbar from "@material-ui/core/Toolbar";
import Box from "@material-ui/core/Box";
import FormControl from "@material-ui/core/FormControl";
import AddIcon from "@material-ui/icons/Add";
import ViewModuleIcon from "@material-ui/icons/ViewModule";
import MenuItem from "@material-ui/core/MenuItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";
import Button from "@material-ui/core/Button";
import EditIcon from "@material-ui/icons/Edit";
import SaveIcon from "@material-ui/icons/Save";
import {
  MuiThemeProvider,
  createMuiTheme,
  makeStyles,
} from "@material-ui/core/styles";

import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Goal, Empty } from "productimon/proto/common/common_pb";

import DisplayGoal from "./DisplayGoal";
import AddGoal from "./AddGoal";
import { rpc, nowNano } from "../../Utils";

const theme = createMuiTheme({
  palette: {
    secondary: { main: "#4caf50" },
  },
});

const useStyles = makeStyles((theme) => ({
  root: {
    display: "flex",
  },
  content: {
    flexGrow: 1,
    overflow: "auto",
  },
  toolbarSecondary: {
    justifyContent: "space-between",
    overflowX: "auto",
  },
  container: {
    paddingTop: theme.spacing(1),
    paddingBottom: theme.spacing(4),
    height: "90%",
  },
  paper: {
    padding: theme.spacing(2),
    display: "flex",
    overflow: "auto",
    flexDirection: "column",
    height: 290,
  },
  formControl: {
    marginBottom: theme.spacing(1),
    float: "right",
  },
}));

export default function Goals(props) {
  const classes = useStyles();
  const [state, setState] = useState(location.pathname);

  return (
    <div>
      <div className={classes.root}>
        <Box mx="auto" p={1}>
          <Toolbar className={classes.toolbarSecondary}>
            <MenuItem
              button
              color="inherit"
              component={Link}
              to="/dashboard/goals/view"
              onClick={() => setState("/dashboard/goals/view")}
              selected={state == "/dashboard/goals/view"}
            >
              <ListItemIcon>
                <ViewModuleIcon />
              </ListItemIcon>
              <ListItemText primary="View Goals" />
            </MenuItem>
            <MenuItem
              button
              color="inherit"
              component={Link}
              to="/dashboard/goals/add"
              onClick={() => setState("/dashboard/goals/add")}
              selected={state == "/dashboard/goals/add"}
            >
              <ListItemIcon>
                <AddIcon />
              </ListItemIcon>
              <ListItemText primary="Add Goals" />
            </MenuItem>
          </Toolbar>
        </Box>
      </div>
      <main className={classes.content}>
        <Display goals={props.goals} setGoals={props.setGoals} />
      </main>
    </div>
  );
}

function transformGoal(goal) {
  return {
    id: goal.getId(),
    isPercent: goal.hasPercentamount(),
    type: goal.getType(),
    title: goal.getTitle(),
    amount: goal.hasPercentamount()
      ? goal.getPercentamount()
      : goal.getFixedamount(),
    itemType: goal.hasLabel() ? "label" : "app",
    item: goal.hasLabel() ? goal.getLabel() : goal.getApplication(),
    start: goal.getGoalinterval().getStart().getNanos(),
    end: goal.getGoalinterval().getEnd().getNanos(),
    compareStart: goal.getCompareinterval()
      ? goal.getCompareinterval().getStart().getNanos()
      : 0,
    compareEnd: goal.getCompareinterval()
      ? goal.getCompareinterval().getEnd().getNanos()
      : 0,
    progress: goal.getProgress(),
  };
}

function Display(props) {
  const classes = useStyles();
  const [goals, setGoals] = React.useState([]);
  const [refresh, setRefresh] = React.useState(0);
  const [showOld, setShowOld] = React.useState(false);
  const { enqueueSnackbar } = useSnackbar();

  useEffect(() => {
    rpc(DataAggregator.GetGoals, new Empty())
      .then((res) => {
        setGoals(
          res
            .getGoalsList()
            .map(transformGoal)
            .filter((g) => (showOld ? g.end <= nowNano() : g.end > nowNano()))
        );
      })
      .catch((err) => {
        console.error(err);
        enqueueSnackbar(err, { variant: "error" });
      });
  }, [refresh, showOld]);

  const addGoal = () => {
    setRefresh(refresh + 1);
  };

  const removeGoal = (goalSpec) => {
    const goal = new Goal();
    goal.setId(goalSpec.id);
    rpc(DataAggregator.DeleteGoal, goal)
      .then((res) => {
        setRefresh(refresh + 1);
      })
      .catch((err) => {
        console.error(err);
        enqueueSnackbar(err, { variant: "error" });
      });

    // TODO
    const newGoals = goals.filter((g) => g.id != goalSpec.goalId);
    setGoals(newGoals);
  };

  const [editing, setEditing] = React.useState(false);

  const toggleEditingMode = (event) => {
    setEditing(!editing);
  };

  return (
    <Switch>
      <Route path="/dashboard/goals/add">
        <div>
          <Container maxWidth="lg" className={classes.container}>
            <AddGoal onAdd={addGoal} />
          </Container>
        </div>
      </Route>
      <Route path="/dashboard/goals/view">
        <Container maxWidth="lg" className={classes.container}>
          <FormControl className={classes.formControl}>
            <MuiThemeProvider theme={theme}>
              <Button
                variant="contained"
                color={editing ? "secondary" : "primary"}
                className={classes.button}
                startIcon={editing ? <SaveIcon /> : <EditIcon />}
                onClick={toggleEditingMode}
              >
                {editing ? "Done" : "Edit"}
              </Button>
            </MuiThemeProvider>
          </FormControl>
          <Grid container spacing={2}>
            {goals.length ? (
              goals.map((goal) => (
                <Grid item xs={12} md={4} lg={4} key={goal.id}>
                  <Paper className={classes.paper}>
                    <DisplayGoal
                      spec={goal}
                      removeButton={editing}
                      onRemove={removeGoal}
                    />
                  </Paper>
                </Grid>
              ))
            ) : (
              <p>You don't have any goals yet</p>
            )}
          </Grid>
        </Container>
      </Route>
    </Switch>
  );
}
