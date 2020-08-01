import React, { useEffect, useState } from "react";
import { useHistory } from "react-router-dom";
import { BrowserRouter as Router, Switch, Route } from "react-router-dom";

import {
  MuiThemeProvider,
  createMuiTheme,
  makeStyles,
} from "@material-ui/core/styles";

import SignIn from "./account/SignIn";
import SignUp from "./account/SignUp";
import Settings from "./account/Settings";
import Dashboard from "./dashboard/Dashboard";
import Fixture from "./core/Fixture";

import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Empty } from "productimon/proto/common/common_pb";

import { rpc } from "./Utils";

const useStyles = makeStyles((theme) => ({
  root: {
    display: "flex",
  },
  content: {
    flexGrow: 1,
    height: "100vh",
    overflow: "auto",
  },
  appBarSpacer: theme.mixins.toolbar,
}));

export default function App() {
  const [loaded, setLoaded] = React.useState(false);
  const [userDetails, setUserDetails] = React.useState(null);
  const [graphs, setGraphs] = React.useState({});
  const theme = createMuiTheme();
  const classes = useStyles();
  const history = useHistory();

  if (window.localStorage.getItem("token")) {
    const request = new Empty();
    useEffect(() => {
      rpc(DataAggregator.UserDetails, history, {
        onEnd: ({ status, statusMessage, headers, message }) => {
          console.log(`Authenticated as ${message.getUser().getEmail()}`);
          setUserDetails(message.getUser());
          setLoaded(true);
        },
        request,
      });
    }, []);
  } else {
    useEffect(() => setLoaded(true), []);
  }

  return loaded ? (
    <MuiThemeProvider theme={theme}>
      <div className={classes.root}>
        <Router>
          <Fixture
            graphs={graphs}
            userDetails={userDetails}
            setUserDetails={setUserDetails}
          />
          <main className={classes.content}>
            <div className={classes.appBarSpacer} />
            <Switch>
              <Route path="/" exact>
                <SignIn
                  userDetails={userDetails}
                  setUserDetails={setUserDetails}
                />
              </Route>
              <Route path="/signup">
                <SignUp setUserDetails={setUserDetails} />
              </Route>
              <Route path="/dashboard">
                <Dashboard graphs={graphs} setGraphs={setGraphs} />
              </Route>
              <Route path="/settings">
                <Settings />
              </Route>
            </Switch>
          </main>
        </Router>
      </div>
    </MuiThemeProvider>
  ) : (
    <p>loading...</p>
  );
}
