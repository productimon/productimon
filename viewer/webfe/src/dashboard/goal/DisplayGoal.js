import React, { useEffect, useState } from "react";
import moment from "moment";
import { ResponsiveContainer, Cell, PieChart, Pie, Label } from "recharts";

import Typography from "@material-ui/core/Typography";
import IconButton from "@material-ui/core/IconButton";
import DeleteIcon from "@material-ui/icons/Delete";
import { makeStyles } from "@material-ui/core/styles";

import { formatNano, humanizeDuration } from "../../Utils";

const useStyles = makeStyles((theme) => ({
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

// just using this to change size now, idea is use this to format labels nicely
const Text = ({ tag = "span", size = 12, caps, ...props }) => {
  const Tag = tag;
  const sx = {
    fontSize: size,
    textTransform: caps ? "uppercase" : null,
  };
  return <Tag {...props} style={sx} />;
};

// convert progress which is a decimal to a percentage
function renderProgressLabel(progress) {
  return (progress * 100.0).toFixed(0).toString().concat("%");
}

function appendZero(number) {
  if (number < 10) {
    return "0" + number;
  }
  return number;
}

// convert Date() object to a date and time string
function dateTimeString(date) {
  return (
    appendZero(date.getDate()) +
    "/" +
    appendZero(date.getMonth() + 1) +
    "/" +
    appendZero(date.getFullYear()) +
    " " +
    appendZero(date.getHours()) +
    ":" +
    appendZero(date.getMinutes()) +
    ":" +
    appendZero(date.getSeconds())
  );
}

function GoalLabel({ goalSpec }) {
  const sign = goalSpec.amount > 0 ? "more" : "less";
  const amount = goalSpec.isPercent
    ? `${(goalSpec.amount * 100).toFixed()}%`
    : humanizeDuration(Math.abs(goalSpec.amount) / 10 ** 9);
  const target = goalSpec.compareStart == 0 ? amount : `${amount} ${sign}`;
  return (
    <p>
      {`${target} on ${goalSpec.item}`}
      <br />
      {`from ${formatNano(goalSpec.start)} to ${formatNano(goalSpec.end)}`}
      {goalSpec.compareStart != 0 && (
        <React.Fragment>
          <br />
          {`compared from ${formatNano(goalSpec.compareStart)} to ${formatNano(
            goalSpec.compareEnd
          )}`}
        </React.Fragment>
      )}
    </p>
  );
}

export default function DisplayGoal(props) {
  const classes = useStyles();
  const goalSpec = props.spec;
  const title = goalSpec.title || "No title given";

  const pieColor = goalSpec.type === "limiting" ? "red" : "green";

  const pieData = [
    { name: "completed", value: goalSpec.progress },
    { name: "remaining", value: 1 - goalSpec.progress },
  ];

  return (
    <div className={classes.box}>
      <div className={classes.titleAndButton}>
        <Typography variant="h6" color="primary">
          {title}
        </Typography>
        {props.removeButton && (
          <IconButton
            style={{ marginLeft: "auto" }}
            onClick={() => {
              props.onRemove(goalSpec);
            }}
            size="small"
          >
            <DeleteIcon />
          </IconButton>
        )}
      </div>
      <Typography variant="subtitle1" align="center">
        Progress
      </Typography>

      <ResponsiveContainer
        height="95%"
        margin={{ top: 0, left: 0, right: 0, bottom: 0 }}
      >
        <PieChart width={200} height={200}>
          <Pie
            innerRadius={30}
            outerRadius={45}
            data={pieData}
            dataKey="value"
            startAngle={90}
            endAngle={450}
          >
            <Label
              value={renderProgressLabel(goalSpec.progress)}
              position="center"
            />
            {pieData.map((entry, index) => (
              <Cell
                key={`cell-${index}`}
                fill={
                  entry.name === "completed" ? pieColor : "rgba(0, 0, 0, 0)"
                }
              />
            ))}
          </Pie>
        </PieChart>
      </ResponsiveContainer>
      <GoalLabel goalSpec={goalSpec} />
    </div>
  );
}
