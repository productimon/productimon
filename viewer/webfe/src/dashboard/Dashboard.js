import React, { useState } from "react";
import {
  Switch,
  Route,
  Link,
  useRouteMatch,
  useHistory,
} from "react-router-dom";
import clsx from "clsx";
import { makeStyles } from "@material-ui/core/styles";
import CssBaseline from "@material-ui/core/CssBaseline";
import Drawer from "@material-ui/core/Drawer";
import Box from "@material-ui/core/Box";
import AppBar from "@material-ui/core/AppBar";
import Toolbar from "@material-ui/core/Toolbar";
import List from "@material-ui/core/List";
import Typography from "@material-ui/core/Typography";
import Divider from "@material-ui/core/Divider";
import IconButton from "@material-ui/core/IconButton";
import Badge from "@material-ui/core/Badge";
import Container from "@material-ui/core/Container";
import Grid from "@material-ui/core/Grid";
import Paper from "@material-ui/core/Paper";
import MaterialLink from "@material-ui/core/Link";
import MenuIcon from "@material-ui/icons/Menu";
import ChevronLeftIcon from "@material-ui/icons/ChevronLeft";
import NotificationsIcon from "@material-ui/icons/Notifications";
import ExitToAppIcon from "@material-ui/icons/ExitToApp";

import { useStyles } from "./Theme";
import DashboardCustomizer from "./DashboardCustomizer";
import Histogram from "./Histogram";
import DisplayTable from "./Table";
import DisplayPie from "./Pie";
import { google_colors } from "./utils";

import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";
import ListItem from "@material-ui/core/ListItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import ListItemText from "@material-ui/core/ListItemText";
import DashboardIcon from "@material-ui/icons/Dashboard";

import BrushIcon from "@material-ui/icons/Brush";
import BarChartIcon from "@material-ui/icons/BarChart";
import PieChartIcon from "@material-ui/icons/PieChart";
import TableChartIcon from "@material-ui/icons/TableChart";

import { grpc } from "@improbable-eng/grpc-web";
import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Empty } from "productimon/proto/common/common_pb";

export default function Dashboard() {
  const history = useHistory();

  const redirectToLogin = () => {
    window.localStorage.clear();
    history.push("/");
  };

  // redirect user to login page if unable to get user details
  const token = window.localStorage.getItem("token");
  const request = new Empty();
  if (!token) {
    redirectToLogin();
    return <p>Redirecting to login...</p>;
  }
  grpc.unary(DataAggregator.UserDetails, {
    host: "/rpc",
    metadata: new grpc.Metadata({ Authorization: token }),
    onEnd: ({ status, statusMessage, headers, message }) => {
      if (status != 0) {
        console.error("response ", status, statusMessage, headers, message);
        redirectToLogin();
        return;
      }
      console.log(`Authenticated as ${message.getUser().getEmail()}`);
    },
    request,
  });

  // state stores what page the main section displays is in, by default we start with dashboard
  const [state, setState] = React.useState("dashboard");
  const classes = useStyles();
  const [open, setOpen] = React.useState(true);
  const handleDrawerOpen = () => {
    setOpen(true);
  };
  const handleDrawerClose = () => {
    setOpen(false);
  };
  const [anchorEl, setAnchorEl] = React.useState(null);

  const handleClick = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  return (
    <div className={classes.root}>
      <CssBaseline />
      <AppBar
        position="absolute"
        className={clsx(classes.appBar, open && classes.appBarShift)}
        style={{ backgroundColor: "#484848" }}
      >
        <Toolbar className={classes.toolbar}>
          <IconButton
            edge="start"
            color="inherit"
            onClick={handleDrawerOpen}
            className={clsx(
              classes.menuButton,
              open && classes.menuButtonHidden
            )}
          >
            <MenuIcon />
          </IconButton>
          <Typography
            component="h1"
            variant="h6"
            color="inherit"
            noWrap
            className={classes.title}
          >
            Productimon
          </Typography>
          <Menu
            anchorEl={anchorEl}
            keepMounted
            open={Boolean(anchorEl)}
            onClose={handleClose}
          >
            <MenuItem onClick={handleClose}>Settings</MenuItem>
            <MenuItem onClick={redirectToLogin}>Logout</MenuItem>
          </Menu>

          <Typography
            onClick={handleClick}
            style={{ textAlign: "right", color: "white" }}
          >
            Account
          </Typography>
        </Toolbar>
      </AppBar>

      <Drawer
        variant="permanent"
        classes={{
          paper: clsx(classes.drawerPaper, !open && classes.drawerPaperClose),
        }}
        open={open}
      >
        <div className={classes.toolbarIcon}>
          <IconButton onClick={handleDrawerClose}>
            <ChevronLeftIcon />
          </IconButton>
        </div>
        <Divider />
        <List>
          <MenuItem
            button
            component={Link}
            to="/dashboard"
            onClick={() => setState("dashboard")}
            selected={state == "dashboard"}
          >
            <ListItemIcon>
              <DashboardIcon />
            </ListItemIcon>
            <ListItemText primary="Dashboard" />
          </MenuItem>

          <MenuItem
            button
            component={Link}
            to="/dashboard/customize"
            onClick={() => setState("customize")}
            selected={state == "customize"}
          >
            <ListItemIcon>
              <BrushIcon />
            </ListItemIcon>
            <ListItemText primary="Customize" />
          </MenuItem>

          <MenuItem
            button
            component={Link}
            to="/dashboard/histogram"
            onClick={() => setState("histogram")}
            selected={state == "histogram"}
          >
            <ListItemIcon>
              <BarChartIcon />
            </ListItemIcon>
            <ListItemText primary="Histogram" />
          </MenuItem>
          <MenuItem
            button
            component={Link}
            to="/dashboard/pie"
            onClick={() => setState("pie")}
            selected={state == "pie"}
          >
            <ListItemIcon>
              <PieChartIcon />
            </ListItemIcon>
            <ListItemText primary="Pie Chart" />
          </MenuItem>
          <MenuItem
            button
            component={Link}
            to="/dashboard/table"
            onClick={() => setState("table")}
            selected={state == "table"}
          >
            <ListItemIcon>
              <TableChartIcon />
            </ListItemIcon>
            <ListItemText primary="Table" />
          </MenuItem>
        </List>
        <Divider />
      </Drawer>

      <main className={classes.content}>
        <div className={classes.appBarSpacer} />
        <Display />
      </main>
    </div>
  );
}

// colorMap is a universal mapping of label -> display color
const colorMap = new Map();
var colorIdx = 0;

function getLabelColor(label) {
  if (!colorMap.has(label)) {
    colorMap.set(label, google_colors[colorIdx]);
    colorIdx++;
    colorIdx = colorIdx % google_colors.length;
  }
  return colorMap.get(label);
}

function Display() {
  const classes = useStyles();

  const fixedHeightPaperTable = clsx(classes.paper, classes.fixedHeightTable);
  const fixedHeightPaperHistogram = clsx(
    classes.paper,
    classes.fixedHeightHistogram
  );
  const fixedHeightPaperPie = clsx(classes.paper, classes.fixedHeightPie);

  const initialGraphs = [
    {
      graphType: "histogram",
      graphTitle: "Last ten minutes",
      startTimeUnit: "Minutes",
      startTimeVal: "10",
      endTimeUnit: "Seconds",
      endTimeVal: "0",
      intervals: "10",
    },
    {
      graphType: "piechart",
      graphTitle: "Top 5 most used",
      startTimeUnit: "Years",
      startTimeVal: "10",
      endTimeUnit: "Seconds",
      endTimeVal: "0",
      numItems: "5",
    },
    {
      graphType: "table",
      graphTitle: "Total use",
      startTimeUnit: "Years",
      startTimeVal: "10",
      endTimeUnit: "Seconds",
      endTimeVal: "0",
    },
  ];
  const [graphs, setGraphs] = useState(initialGraphs);

  // This is passed as a prop to the DashboardCustomizer. Right now this just updates a list of graphs that are rendered. In the future this will send the graph to the aggregator to save it to the account.
  const addGraph = (graphSpec) => {
    setGraphs(graphs.concat([graphSpec]));
  };

  const gmap = {
    histogram: (graph) => {
      return <Histogram spec={graph} getLabelColor={getLabelColor} />;
    },
    piechart: (graph) => {
      return <DisplayPie spec={graph} getLabelColor={getLabelColor} />;
    },
    table: (graph) => {
      return <DisplayTable spec={graph} />;
    },
  };

  let match = useRouteMatch();
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
      <Route path="/dashboard/histogram">
        <div>
          <Container maxWidth="lg" className={classes.container}>
            <Grid container spacing={2}>
              <Grid item xs={12} md={12} lg={12}>
                <Paper className={fixedHeightPaperHistogram}>
                  <Histogram
                    spec={initialGraphs[0]}
                    getLabelColor={getLabelColor}
                  />
                </Paper>
              </Grid>
            </Grid>
          </Container>
        </div>
      </Route>
      <Route path="/dashboard/pie">
        <div>
          <Container maxWidth="lg" className={classes.container}>
            <Grid container spacing={2}>
              <Grid item xs={12} md={12} lg={12}>
                <Paper className={fixedHeightPaperPie}>
                  <DisplayPie
                    spec={initialGraphs[1]}
                    getLabelColor={getLabelColor}
                  />
                </Paper>
              </Grid>
            </Grid>
          </Container>
        </div>
      </Route>
      <Route path="/dashboard/table">
        <div>
          <Container maxWidth="lg" className={classes.container}>
            <Grid container spacing={2}>
              <Grid item xs={12} md={12} lg={12}>
                <Paper className={classes.paper}>
                  <DisplayTable spec={initialGraphs[2]} />
                </Paper>
              </Grid>
            </Grid>
          </Container>
        </div>
      </Route>
      <Route path="/">
        <Container maxWidth="lg" className={classes.container}>
          <Grid container spacing={2}>
            {graphs.map((graph, index) => (
              <Grid item xs={12} md={6} lg={6} key={index}>
                <Paper className={fixedHeightPaperHistogram} key={index}>
                  {gmap[graph.graphType](graph)}
                </Paper>
              </Grid>
            ))}
          </Grid>
        </Container>
      </Route>
    </Switch>
  );
}
