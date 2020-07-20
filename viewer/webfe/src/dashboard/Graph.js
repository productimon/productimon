import React, { useState } from "react";
import clsx from "clsx";
import { useHistory } from "react-router-dom";

import Container from "@material-ui/core/Container";
import Grid from "@material-ui/core/Grid";
import Paper from "@material-ui/core/Paper";
import { useParams } from "react-router-dom";
import { makeStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
// import IconButton from "@material-ui/core/IconButton";
// import DeleteIcon from "@material-ui/icons/Delete";
//

import Histogram from "./Histogram";
import Table from "./Table";
import Pie from "./Pie";

const useStyles = makeStyles((theme) => ({
  link: {
    cursor: "pointer",
  },
  box: {
    height: "100%",
    display: "flex",
    flexFlow: "column",
    overflow: "hidden",
  },
}));

function renderGraph(props) {
  switch (props.graphSpec.graphType) {
    case "histogram":
      return <Histogram {...props} />;
    case "piechart":
      return <Pie {...props} />;
    case "table":
      return <Table {...props} />;
  }
  console.error(`Unknown graph type ${graphSpec.graphType}`);
  return null;
}

export default function Graph(props) {
  const graphSpec = props.graphSpec;
  const classes = useStyles();
  const history = useHistory();

  if (!graphSpec) return <p>No such graph</p>;
  // TODO could have a generic title generation here if title is not given
  const defaultTitle = graphSpec.graphType;

  // TODO make graphs more interative in fullscreen mode?
  // check props.fullscreen
  return (
    <div className={classes.box}>
      {/* Alignment is hard...
          <IconButton>
          <DeleteIcon />
          </IconButton> */}
      <Typography
        onClick={() => history.push(`/dashboard/graph/${graphSpec.graphId}`)}
        className={classes.link}
        component="h2"
        variant="h6"
        color="primary"
        gutterBottom
      >
        {graphSpec.graphTitle || defaultTitle}
      </Typography>
      {renderGraph(props)}
    </div>
  );
}
