import React from "react";
import { useParams } from "react-router-dom";

import Container from "@material-ui/core/Container";
import Paper from "@material-ui/core/Paper";
import { makeStyles } from "@material-ui/core/styles";

import Graph from "./Graph.js";

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
    height: "100%",
  },
}));

export default function FullScreenGraph(props) {
  const graphId = useParams().graphId;
  const graphSpec = props.graphs[graphId];
  const classes = useStyles();
  return (
    <Container key={graphId} maxWidth="lg" className={classes.container}>
      <Paper className={classes.paper}>
        <Graph {...props} graphSpec={graphSpec} fullscreen />
      </Paper>
    </Container>
  );
}
