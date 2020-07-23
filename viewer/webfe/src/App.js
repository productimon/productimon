import React, { useState } from "react";
import { BrowserRouter as Router, Switch, Route } from "react-router-dom";

import { makeStyles } from "@material-ui/core/styles";

import SignIn from "./account/SignIn";
import SignUp from "./account/SignUp";
import Dashboard from "./dashboard/Dashboard";
import Fixture from "./core/Fixture";

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
  const [loggedIn, setLoggedIn] = React.useState(
    localStorage.getItem("token") != null
  );
  const [graphs, setGraphs] = React.useState({});

  const classes = useStyles();
  return (
    <div className={classes.root}>
      <Router>
        <Fixture
          graphs={graphs}
          loggedIn={loggedIn}
          setLoggedIn={setLoggedIn}
        />
        <main className={classes.content}>
          <div className={classes.appBarSpacer} />
          <Switch>
            <Route path="/" exact>
              <SignIn setLoggedIn={setLoggedIn} />
            </Route>
            <Route path="/signup">
              <SignUp setLoggedIn={setLoggedIn} />
            </Route>
            <Route path="/dashboard">
              <Dashboard graphs={graphs} setGraphs={setGraphs} />
            </Route>
          </Switch>
        </main>
      </Router>
    </div>
  );
}
