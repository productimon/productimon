import React, { useState } from "react";
import { useHistory } from "react-router-dom";

import Container from "@material-ui/core/Container";
import Grid from "@material-ui/core/Grid";
import Paper from "@material-ui/core/Paper";
import { useParams } from "react-router-dom";
import { makeStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import IconButton from "@material-ui/core/IconButton";
import DeleteIcon from "@material-ui/icons/Delete";

import Histogram from "./Histogram";
import Table from "./Table";
import Pie from "./Pie";

const useStyles = makeStyles((theme) => ({
  link: {
    cursor: "pointer",
  },
  box: {
    display: "flex",
    flexFlow: "column",
    overflow: "hidden",
    height: "100%",
  },
  titleAndButton: {
    display: "flex",
    flexFlow: "row",
  },
}));

export const graphTypes = {
  histogram: {
    heading: "Histogram",
    render: (props) => <Histogram {...props} />,
  },
  piechart: {
    heading: "Piechart",
    render: (props) => <Pie {...props} />,
  },
  table: {
    heading: "Table",
    render: (props) => <Table {...props} />,
  },
};

export function graphTitle(graphSpec) {
  // TODO could have a generic title generation here if title is not given
  const defaultTitle = graphTypes[graphSpec.graphType].heading;
  return graphSpec.graphTitle || defaultTitle;
}

export default function Graph(props) {
  const graphSpec = props.graphSpec;
  const classes = useStyles();
  const history = useHistory();

  if (!graphSpec) return <p>No such graph</p>;

  // disable title link when in preview or fullscreen
  const titleProps =
    props.preview || props.fullscreen
      ? {}
      : {
          onClick: () => history.push(`/dashboard/graph/${graphSpec.graphId}`),
          className: classes.link,
        };

  // TODO make graphs more interative in fullscreen mode?
  // check props.fullscreen
  return (
    <div className={classes.box}>
      <div className={classes.titleAndButton}>
        <Typography
          {...titleProps}
          component="h2"
          variant="h6"
          color="primary"
          gutterBottom
        >
          {props.preview ? "Preview" : graphTitle(graphSpec)}
        </Typography>
        {props.removeButton && (
          <IconButton
            style={{ marginLeft: "auto" }}
            onClick={() => {
              props.onRemove(graphSpec);
            }}
          >
            <DeleteIcon />
          </IconButton>
        )}
      </div>
      {graphTypes[graphSpec.graphType].render(props)}
    </div>
  );
}
