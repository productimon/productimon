import React, { useState, useEffect } from "react";
import { Switch, Route, useHistory } from "react-router-dom";
import clsx from "clsx";

import { makeStyles } from "@material-ui/core/styles";
import Container from "@material-ui/core/Container";
import Grid from "@material-ui/core/Grid";
import Paper from "@material-ui/core/Paper";

import DashboardCustomizer from "./DashboardCustomizer";
import { rpc } from "../Utils";
import Graph from "./Graph";
import FullScreenGraph from "./FullScreenGraph";

import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Empty } from "productimon/proto/common/common_pb";

const useStyles = makeStyles((theme) => ({
  container: {
    paddingTop: theme.spacing(4),
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
}));

const initialGraphs = {
  0: {
    graphId: 0,
    graphType: "histogram",
    graphTitle: "Last ten minutes",
    startTimeUnit: "Minutes",
    startTimeVal: "10",
    endTimeUnit: "Seconds",
    endTimeVal: "0",
    intervals: "10",
  },
  1: {
    graphId: 1,
    graphType: "piechart",
    graphTitle: "Top 5 most used",
    startTimeUnit: "Years",
    startTimeVal: "10",
    endTimeUnit: "Seconds",
    endTimeVal: "0",
    numItems: "5",
  },
  2: {
    graphId: 2,
    graphType: "table",
    graphTitle: "Total use",
    startTimeUnit: "Years",
    startTimeVal: "10",
    endTimeUnit: "Seconds",
    endTimeVal: "0",
  },
};

export default function Dashboard(props) {
  const history = useHistory();

  // redirect user to login page if unable to get user details
  const request = new Empty();

  rpc(DataAggregator.UserDetails, history, {
    onEnd: ({ status, statusMessage, headers, message }) => {
      console.log(`Authenticated as ${message.getUser().getEmail()}`);
    },
    request,
  });

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
    const newId = Math.max(...Object.keys(graphs), -1) + 1;
    const newGraphs = {
      [newId]: { graphId: newId, ...graphSpec },
      ...graphs,
    };
    setGraphs(newGraphs);
    window.localStorage.setItem("graphs", JSON.stringify(newGraphs));
  };

  const updateGraph = (graphSpec) => {
    const newGraphs = {
      ...graphs,
      [graphSpec.graphId]: { ...graphSpec },
    };
    setGraphs(newGraphs);
    window.localStorage.setItem("graphs", JSON.stringify(newGraphs));
  };

  const removeGraph = (graphSpec) => {
    const newGraphs = Object.keys(graphs)
      .filter((id) => id != graphSpec.graphId)
      .reduce((ret, graphId) => ({ ...ret, [graphId]: graphs[graphId] }), {});
    setGraphs(newGraphs);
    window.localStorage.setItem("graphs", JSON.stringify(newGraphs));
    history.push("/dashboard");
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
      <Route path="/dashboard/graph/:graphId">
        <FullScreenGraph
          graphs={graphs}
          onUpdate={updateGraph}
          onRemove={removeGraph}
        />
      </Route>
      <Route path="/">
        <Container maxWidth="lg" className={classes.container}>
          <Grid container spacing={2}>
            {Object.values(graphs).map((graph, index) => (
              <Grid item xs={12} md={6} lg={6} key={index}>
                <Paper
                  className={clsx(classes.paper, classes.dashboardGraph)}
                  key={index}
                >
                  <Graph graphSpec={graph} onRemove={removeGraph} />
                </Paper>
              </Grid>
            ))}
          </Grid>
        </Container>
      </Route>
    </Switch>
  );
}
