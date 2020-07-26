import React, { useState, useEffect } from "react";
import clsx from "clsx";
import { useHistory } from "react-router-dom";
import { makeStyles } from "@material-ui/core/styles";
import Container from "@material-ui/core/Container";
import Grid from "@material-ui/core/Grid";
import Button from "@material-ui/core/Button";
import TextField from "@material-ui/core/TextField";
import Checkbox from "@material-ui/core/Checkbox";
import { rpc, redirectToLogin } from "../Utils";
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
}));

export default function Settings() {
  const history = useHistory();

  // redirect user to login page if unable to get user details
  const request = new Empty();
  rpc(DataAggregator.UserDetails, history, {
    onEnd: ({ status, statusMessage, headers, message }) => {
      console.log(`Authenticated as ${message.getUser().getEmail()}`);
    },
    request,
  });

  return <SettingsForm />;
}

function SettingsForm() {
  const history = useHistory();

  const [confirmationInput, setConfirmationInput] = React.useState("");
  const handleSetConfirmationInput = (event) => {
    setConfirmationInput(event.target.value);
  };

  const deleteAccount = () => {
    const request = new Empty();
    rpc(DataAggregator.DeleteAccount, history, {
      onEnd: ({ status, statusMessage, headers, message }) => {
        redirectToLogin(history);
      },
      request,
    });
  };

  const classes = useStyles();

  return (
    <Container maxWidth="lg" className={classes.container}>
      <Grid
        container
        spacing={2}
        direction={"column"}
        justify="center"
        alignItems="center"
      >
        <Grid item xs={12} md={6} lg={6}>
          <Button
            variant="contained"
            disabled={!(confirmationInput == "CONFIRM")}
            onClick={() => {
              deleteAccount();
            }}
          >
            Delete Account
          </Button>
        </Grid>
        <Grid item xs={12} md={6} lg={6}>
          <TextField
            id="standard-basic"
            label="Type CONFIRM to continue"
            value={confirmationInput}
            onChange={handleSetConfirmationInput}
          />
        </Grid>
      </Grid>
    </Container>
  );
}
