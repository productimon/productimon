import React, { useState, useEffect } from "react";
import { Switch, Route, useHistory } from "react-router-dom";
import clsx from "clsx";
import { SortableContainer, SortableElement } from "react-sortable-hoc";
import { useSnackbar } from "notistack";

import {
  MuiThemeProvider,
  createMuiTheme,
  makeStyles,
} from "@material-ui/core/styles";
import { palette } from "@material-ui/system";
import Container from "@material-ui/core/Container";
import Grid from "@material-ui/core/Grid";
import Paper from "@material-ui/core/Paper";
import FormControl from "@material-ui/core/FormControl";
import Button from "@material-ui/core/Button";
import EditIcon from "@material-ui/icons/Edit";
import SaveIcon from "@material-ui/icons/Save";

import DashboardCustomizer from "./DashboardCustomizer";
import Graph from "./Graph";
import FullScreenGraph from "./FullScreenGraph";
import AdminLabelManagement from "./AdminLabelManagement";
import AdminManagement from "./AdminManagement";
import AdminServerStatus from "./AdminServerStatus";

const useStyles = makeStyles((theme) => ({
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
  },
  dashboardGraph: {
    height: 375,
  },
  formControl: {
    marginBottom: theme.spacing(1),
    float: "right",
  },
}));

const theme = createMuiTheme({
  palette: {
    secondary: { main: "#4caf50" },
  },
});

const initialGraphs = {
  0: {
    graphId: 0,
    graphOrder: 0,
    graphType: "histogram",
    graphTitle: "Last ten minutes",
    startTimeUnit: "Minutes",
    startTimeVal: "10",
    endTimeUnit: "Seconds",
    endTimeVal: "0",
    intervals: "10",
    device: "all",
  },
  1: {
    graphId: 1,
    graphOrder: 1,
    graphType: "piechart",
    graphTitle: "Top 5 most used",
    startTimeUnit: "Years",
    startTimeVal: "10",
    endTimeUnit: "Seconds",
    endTimeVal: "0",
    numItems: "5",
    device: "all",
  },
  2: {
    graphId: 2,
    graphOrder: 2,
    graphType: "table",
    graphTitle: "Total use",
    startTimeUnit: "Years",
    startTimeVal: "10",
    endTimeUnit: "Seconds",
    endTimeVal: "0",
    device: "all",
  },
};

export default function Dashboard(props) {
  const history = useHistory();
  const { enqueueSnackbar, closeSnackbar } = useSnackbar();

  const { graphs, setGraphs } = props;
  useEffect(() => {
    const graphJson = window.localStorage.getItem("graphs");
    const localGraphs = graphJson ? JSON.parse(graphJson) : initialGraphs;
    setGraphs(localGraphs);
  }, []);

  const classes = useStyles();

  // This is passed as a prop to the DashboardCustomizer.
  // Right now this just updates a list of graphs that are rendered.
  // In the future this will send the graph to the aggregator to save it to the account.
  const addGraph = (graphSpec) => {
    const newId =
      Math.max(...Object.values(graphs).map((o) => o.graphId), -1) + 1;
    const newOrder =
      Math.max(...Object.values(graphs).map((o) => o.graphOrder), -1) + 1;
    const newGraphs = {
      [newId]: { graphId: newId, graphOrder: newOrder, ...graphSpec },
      ...graphs,
    };

    setGraphs(newGraphs);
    window.localStorage.setItem("graphs", JSON.stringify(newGraphs));
    enqueueSnackbar("Graph added", { variant: "success" });
  };

  const updateGraph = (graphSpec) => {
    const newGraphs = {
      ...graphs,
      [graphSpec.graphId]: { ...graphSpec },
    };
    setGraphs(newGraphs);
    window.localStorage.setItem("graphs", JSON.stringify(newGraphs));
    enqueueSnackbar("Graph updated", { variant: "success" });
  };

  const removeGraph = (graphSpec) => {
    const newGraphs = Object.keys(graphs)
      .filter((id) => id != graphSpec.graphId)
      .reduce((ret, graphId) => ({ ...ret, [graphId]: graphs[graphId] }), {});
    setGraphs(newGraphs);
    window.localStorage.setItem("graphs", JSON.stringify(newGraphs));
    enqueueSnackbar("Graph deleted", { variant: "success" });
    history.push("/dashboard");
  };

  const SortableItem = SortableElement((props) => {
    return (
      <Grid item xs={12} md={6} lg={6}>
        <Paper className={clsx(classes.paper, classes.dashboardGraph)}>
          <Graph
            graphSpec={props.graphSpec}
            onRemove={removeGraph}
            removeButton={props.removeButton}
          />
        </Paper>
      </Grid>
    );
  });

  const SortableList = SortableContainer((props) => {
    const tmp = Object.values(graphs).sort(
      (a, b) => a.graphOrder - b.graphOrder
    );
    return (
      <Grid container spacing={2}>
        {Object.values(graphs)
          .sort((a, b) => a.graphOrder - b.graphOrder)
          .map((graph, index) => (
            <SortableItem
              index={graph.graphOrder}
              key={index}
              graphSpec={graph}
              disabled={props.disabled}
              removeButton={!props.disabled}
            />
          ))}
      </Grid>
    );
  });

  const onSortEnd = ({ oldIndex, newIndex }) => {
    if (oldIndex === newIndex) return;
    const delta = oldIndex > newIndex ? 1 : -1;
    const [lo, hi] =
      oldIndex > newIndex ? [newIndex, oldIndex - 1] : [oldIndex + 1, newIndex];

    const shiftedGraphs = Object.entries(graphs)
      .filter(([_, g]) => g.graphOrder >= lo && g.graphOrder <= hi)
      .reduce(
        (ret, [id, g]) => ({
          ...ret,
          [id]: { ...g, graphOrder: g.graphOrder + delta },
        }),
        {}
      );
    const movedGraph = Object.values(graphs).find(
      (g) => g.graphOrder == oldIndex
    );

    const newGraphs = {
      ...graphs,
      ...shiftedGraphs,
      [movedGraph.graphId]: { ...movedGraph, graphOrder: newIndex },
    };

    setGraphs(newGraphs);
    window.localStorage.setItem("graphs", JSON.stringify(newGraphs));
  };
  const [editing, setEditing] = React.useState(false);

  const toggleEditingMode = (event) => {
    setEditing(!editing);
  };
  return (
    <Switch>
      <Route path="/dashboard/customize">
        <div>
          <Container maxWidth="lg" className={classes.container}>
            <Grid container spacing={2}>
              <Grid item xs={12} md={12} lg={12}>
                <DashboardCustomizer
                  onAdd={(graphSpec) => addGraph(graphSpec)}
                />
              </Grid>
            </Grid>
          </Container>
        </div>
      </Route>
      <Route path="/dashboard/labels">
        <div>
          <Container maxWidth="lg" className={classes.container}>
            <Grid container spacing={2}>
              <Grid item xs={12} md={12} lg={12}>
                <AdminLabelManagement />
              </Grid>
            </Grid>
          </Container>
        </div>
      </Route>
      <Route path="/dashboard/adminLabels">
        <div>
          <Container maxWidth="lg" className={classes.container}>
            <Grid container spacing={2}>
              <Grid item xs={12} md={12} lg={12}>
                <AdminLabelManagement admin />
              </Grid>
            </Grid>
          </Container>
        </div>
      </Route>
      <Route path="/dashboard/adminManagement">
        <div>
          <Container maxWidth="lg" className={classes.container}>
            <Grid container spacing={2}>
              <Grid item xs={12} md={12} lg={12}>
                <AdminManagement />
              </Grid>
            </Grid>
          </Container>
        </div>
      </Route>
      <Route path="/dashboard/adminServerStatus">
        <div>
          <Container maxWidth="lg" className={classes.container}>
            <Grid container spacing={2}>
              <Grid item xs={12} md={12} lg={12}>
                <AdminServerStatus />
              </Grid>
            </Grid>
          </Container>
        </div>
      </Route>
      <Route path="/dashboard/graph/:graphId">
        <FullScreenGraph
          graphs={graphs}
          onUpdate={updateGraph}
          onRemove={removeGraph}
        />
      </Route>
      <Route path="/">
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
          <SortableList
            axis="xy"
            onSortEnd={onSortEnd}
            distance={5}
            lockToContainerEdges={true}
            disabled={!editing}
          />
        </Container>
      </Route>
    </Switch>
  );
}
