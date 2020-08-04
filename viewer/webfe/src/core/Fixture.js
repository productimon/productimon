/**
 * Top menu bar and drawer
 */
import React from "react";
import clsx from "clsx";
import { useHistory, useLocation } from "react-router-dom";

import { makeStyles } from "@material-ui/core/styles";
import CssBaseline from "@material-ui/core/CssBaseline";
import Drawer from "@material-ui/core/Drawer";
import AppBar from "@material-ui/core/AppBar";
import Toolbar from "@material-ui/core/Toolbar";
import List from "@material-ui/core/List";
import Typography from "@material-ui/core/Typography";
import Divider from "@material-ui/core/Divider";
import IconButton from "@material-ui/core/IconButton";
import MenuIcon from "@material-ui/icons/Menu";
import ChevronLeftIcon from "@material-ui/icons/ChevronLeft";
import Button from "@material-ui/core/Button";
import Menu from "@material-ui/core/Menu";
import ListItem from "@material-ui/core/ListItem";
import ListItemIcon from "@material-ui/core/ListItemIcon";
import DashboardIcon from "@material-ui/icons/Dashboard";
import BrushIcon from "@material-ui/icons/Brush";
import BarChartIcon from "@material-ui/icons/BarChart";
import PieChartIcon from "@material-ui/icons/PieChart";
import TableChartIcon from "@material-ui/icons/TableChart";
import SupervisorAccountIcon from "@material-ui/icons/SupervisorAccount";
import StorageIcon from "@material-ui/icons/Storage";
import TextFieldsIcon from "@material-ui/icons/TextFields";
import ListItemText from "@material-ui/core/ListItemText";

import { redirectToLogin } from "../Utils";
import { graphTitle } from "../dashboard/Graph";

const drawerWidth = 240;

const useStyles = makeStyles((theme) => ({
  toolbar: {
    paddingRight: 24, // keep right padding when drawer closed
  },
  toolbarIcon: {
    display: "flex",
    alignItems: "center",
    justifyContent: "flex-end",
    padding: "0 8px",
    ...theme.mixins.toolbar,
  },
  appBar: {
    zIndex: theme.zIndex.drawer + 1,
    transition: theme.transitions.create(["width", "margin"], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.leavingScreen,
    }),
  },
  appBarShift: {
    marginLeft: drawerWidth,
    width: `calc(100% - ${drawerWidth}px)`,
    transition: theme.transitions.create(["width", "margin"], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  },
  menuButton: {
    marginRight: 36,
  },
  hide: {
    display: "none",
  },
  title: {
    flexGrow: 1,
  },
  drawer: {
    width: drawerWidth,
    flexShrink: 0,
    whiteSpace: "nowrap",
  },
  drawerPaper: {
    width: drawerWidth,
    transition: theme.transitions.create("width", {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  },
  drawerPaperClose: {
    overflowX: "hidden",
    transition: theme.transitions.create("width", {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.leavingScreen,
    }),
    width: theme.spacing(7),
    [theme.breakpoints.up("sm")]: {
      width: theme.spacing(9),
    },
  },
}));

const graphIconMap = {
  histogram: <BarChartIcon />,
  piechart: <PieChartIcon />,
  table: <TableChartIcon />,
};

export default function Fixture(props) {
  const history = useHistory();
  const location = useLocation();
  const [anchorEl, setAnchorEl] = React.useState(null);

  const handleClick = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = () => {
    handleClose();
    props.setUserDetails(null);
    window.localStorage.removeItem("token");
    history.push("/");
  };

  const [open, setOpen] = React.useState(true);
  const handleDrawerOpen = () => {
    setOpen(true);
  };
  const handleDrawerClose = () => {
    setOpen(false);
  };

  const gotoLink = (link) => {
    history.push(link);
  };

  const classes = useStyles();
  return (
    <React.Fragment>
      <CssBaseline />
      <AppBar
        position="fixed"
        className={clsx(classes.appBar, {
          [classes.appBarShift]: props.userDetails && open,
        })}
        style={{ backgroundColor: props.userDetails ? "#484848" : "brown" }}
      >
        <Toolbar className={classes.toolbar}>
          <IconButton
            edge="start"
            color="inherit"
            onClick={handleDrawerOpen}
            className={clsx(classes.menuButton, {
              [classes.hide]: !props.userDetails || open,
            })}
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
            <ListItem
              onClick={() => {
                gotoLink("/settings");
              }}
            >
              Settings
            </ListItem>
            <ListItem onClick={handleLogout}>Logout</ListItem>
          </Menu>

          <Typography
            onClick={handleClick}
            style={{ textAlign: "right", color: "white" }}
            className={clsx({
              [classes.hide]: !props.userDetails,
            })}
          >
            Account
          </Typography>
        </Toolbar>
      </AppBar>
      <Drawer
        variant="permanent"
        classes={{
          paper: clsx(open ? classes.drawerPaper : classes.drawerPaperClose),
        }}
        className={clsx(classes.drawer, {
          [classes.drawerPaper]: open,
          [classes.drawerPaperClose]: !open,
          [classes.hide]: !props.userDetails,
        })}
        open={open}
      >
        <div className={classes.toolbarIcon}>
          <IconButton onClick={handleDrawerClose}>
            <ChevronLeftIcon />
          </IconButton>
        </div>
        <Divider />
        <List>
          <ListItem
            button
            onClick={() => gotoLink("/dashboard")}
            selected={location.pathname == "/dashboard"}
          >
            <ListItemIcon>
              <DashboardIcon />
            </ListItemIcon>
            <ListItemText primary="Dashboard" />
          </ListItem>

          <ListItem
            button
            onClick={() => gotoLink("/dashboard/customize")}
            selected={location.pathname == "/dashboard/customize"}
          >
            <ListItemIcon>
              <BrushIcon />
            </ListItemIcon>
            <ListItemText primary="Customize" />
          </ListItem>

          <ListItem
            button
            onClick={() => gotoLink("/dashboard/labels")}
            selected={location.pathname == "/dashboard/labels"}
          >
            <ListItemIcon>
              <DashboardIcon />
            </ListItemIcon>
            <ListItemText primary="Labels" />
          </ListItem>

          <Divider />

          {Object.values(props.graphs).map((graph, idx) => (
            <ListItem
              button
              key={idx}
              onClick={() => gotoLink(`/dashboard/graph/${graph.graphId}`)}
              selected={
                location.pathname == `/dashboard/graph/${graph.graphId}`
              }
            >
              <ListItemIcon>{graphIconMap[graph.graphType]}</ListItemIcon>
              <ListItemText primary={graphTitle(graph)} />
            </ListItem>
          ))}
          <Divider />
          {props.userDetails && props.userDetails.getAdmin() && (
            <React.Fragment>
              <ListItem
                button
                onClick={() => gotoLink("/dashboard/adminLabels")}
                selected={location.pathname == "/dashboard/adminLabels"}
              >
                <ListItemIcon>
                  <TextFieldsIcon />
                </ListItemIcon>
                <ListItemText primary="Admin Labels" />
              </ListItem>

              <ListItem
                button
                onClick={() => gotoLink("/dashboard/adminManagement")}
                selected={location.pathname == "/dashboard/adminManagement"}
              >
                <ListItemIcon>
                  <SupervisorAccountIcon />
                </ListItemIcon>
                <ListItemText primary="Admin Management" />
              </ListItem>

              <Divider />
            </React.Fragment>
          )}
        </List>
      </Drawer>
    </React.Fragment>
  );
}
