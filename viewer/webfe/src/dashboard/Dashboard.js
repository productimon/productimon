import React from 'react';
import {
  BrowserRouter as Router,
  Switch,
  Route,
  Link,
  useRouteMatch,
  useHistory,
} from 'react-router-dom';
import clsx from 'clsx';
import { makeStyles } from '@material-ui/core/styles';
import CssBaseline from '@material-ui/core/CssBaseline';
import Drawer from '@material-ui/core/Drawer';
import Box from '@material-ui/core/Box';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import List from '@material-ui/core/List';
import Typography from '@material-ui/core/Typography';
import Divider from '@material-ui/core/Divider';
import IconButton from '@material-ui/core/IconButton';
import Badge from '@material-ui/core/Badge';
import Container from '@material-ui/core/Container';
import Grid from '@material-ui/core/Grid';
import Paper from '@material-ui/core/Paper';
import MaterialLink from '@material-ui/core/Link';
import MenuIcon from '@material-ui/icons/Menu';
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft';
import NotificationsIcon from '@material-ui/icons/Notifications';
import ExitToAppIcon from '@material-ui/icons/ExitToApp';

import Histogram from './Histogram';
import DisplayTable from './Table';
import DisplayPie from './Pie'

//import { mainListItems, secondaryListItems } from './listItems';
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import DashboardIcon from '@material-ui/icons/Dashboard';
import BarChartIcon from '@material-ui/icons/BarChart';
import PieChartIcon from '@material-ui/icons/PieChart';
import TableChartIcon from '@material-ui/icons/TableChart';

import { grpc } from '@improbable-eng/grpc-web';
import { DataAggregator } from 'productimon/proto/svc/aggregator_pb_service';
import { Empty } from 'productimon/proto/common/common_pb';

const drawerWidth = 240;

const useStyles = makeStyles((theme) => ({
  root: {
    display: 'flex',
  },
  toolbar: {
    paddingRight: 24, // keep right padding when drawer closed
  },
  toolbarIcon: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'flex-end',
    padding: '0 8px',
    ...theme.mixins.toolbar,
  },
  appBar: {
    zIndex: theme.zIndex.drawer + 1,
    transition: theme.transitions.create(['width', 'margin'], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.leavingScreen,
    }),
  },
  appBarShift: {
    marginLeft: drawerWidth,
    width: `calc(100% - ${drawerWidth}px)`,
    transition: theme.transitions.create(['width', 'margin'], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  },
  menuButton: {
    marginRight: 36,
  },
  menuButtonHidden: {
    display: 'none',
  },
  title: {
    flexGrow: 1,
  },
  drawerPaper: {
    position: 'relative',
    whiteSpace: 'nowrap',
    width: drawerWidth,
    transition: theme.transitions.create('width', {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  },
  drawerPaperClose: {
    overflowX: 'hidden',
    transition: theme.transitions.create('width', {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.leavingScreen,
    }),
    width: theme.spacing(7),
    [theme.breakpoints.up('sm')]: {
      width: theme.spacing(9),
    },
  },
  appBarSpacer: theme.mixins.toolbar,
  content: {
    flexGrow: 1,
    height: '100vh',
    overflow: 'auto',
  },
  container: {
    paddingTop: theme.spacing(4),
    paddingBottom: theme.spacing(4),
  },
  paper: {
    padding: theme.spacing(2),
    display: 'flex',
    overflow: 'auto',
    flexDirection: 'column',
  },
  fixedHeightTable: {
    height: 600,
  },
  fixedHeightHistogram: {
    height: 300,
  },
  fixedHeightPie: {
    height: 300,
  },
}));

export default function Dashboard() {
  const history = useHistory();

  const redirectToLogin = () => {
    window.localStorage.clear();
    history.push('/');
  };

  // redirect user to login page if unable to get user details
  const token = window.localStorage.getItem('token');
  const request = new Empty();
  if (!token) {
    redirectToLogin();
    // this return is needed to stop teh execution of the following code
    return <p>Redirecting to login...</p>;
  }
  grpc.unary(DataAggregator.UserDetails, {
    host: '/rpc',
    metadata: new grpc.Metadata({ Authorization: token }),
    onEnd: ({ status, statusMessage, headers, message }) => {
      if (status != 0) {
        console.error('response ', status, statusMessage, headers, message);
        redirectToLogin();
        return;
      }
      console.log(`Authenticated as ${message.getUser().getEmail()}`);
    },
    request,
  });

  // state stores what page the main section displays is in, by default we start with dashboard
  const [state, setState] = React.useState('dashboard');
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
        style={{ backgroundColor: '#484848' }}
      >
        <Toolbar className={classes.toolbar}>
          <IconButton
            edge="start"
            color="inherit"
            aria-label="open drawer"
            onClick={handleDrawerOpen}
            className={clsx(
              classes.menuButton,
              open && classes.menuButtonHidden,
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
            id="account-menu"
            anchorEl={anchorEl}
            keepMounted
            open={Boolean(anchorEl)}
            onClose={handleClose}
          >
            <MenuItem onClick={handleClose}>Settings</MenuItem>
            <MenuItem onClick={redirectToLogin}>Logout</MenuItem>
          </Menu>

          <Typography
            aria-controls="account-menu"
            aria-haspopup="true"
            onClick={handleClick}
            style={{ textAlign: 'right', color: 'white' }}
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
            onClick={() => setState('dashboard')}
            selected={state == 'dashboard'}
          >
            <ListItemIcon>
              <DashboardIcon />
            </ListItemIcon>
            <ListItemText primary="Dashboard" />
          </MenuItem>
          <MenuItem
            button
            component={Link}
            to="/dashboard/histogram"
            onClick={() => setState('histogram')}
            selected={state == 'histogram'}
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
            onClick={() => setState('pie')}
            selected={state == 'pie'}
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
            onClick={() => setState('table')}
            selected={state == 'table'}
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

function Display() {
  const classes = useStyles();

  const fixedHeightPaperTable = clsx(classes.paper, classes.fixedHeightTable);
  const fixedHeightPaperHistogram = clsx(
    classes.paper,
    classes.fixedHeightHistogram,
  );
  const fixedHeightPaperPie = clsx(classes.paper, classes.fixedHeightPie);

  let match = useRouteMatch();

  return (
    <Switch>
      <Route path="/dashboard/histogram">
        <div>
          <Container maxWidth="lg" className={classes.container}>
            <Grid container spacing={2}>
              <Grid item xs={12} md={12} lg={12}>
                <Paper className={fixedHeightPaperHistogram}>
                  <Histogram />
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
                  <DisplayTable />
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
                  <DisplayPie />
                </Paper>
              </Grid>
            </Grid>
          </Container>
        </div>
      </Route>
      <Route path="/">
        <div>
          <Container maxWidth="lg" className={classes.container}>
            <Grid container spacing={2}>
              <Grid item xs={12} md={6} lg={6}>
                <Paper className={fixedHeightPaperHistogram}>
                  <Histogram />
                </Paper>
              </Grid>
              <Grid item xs={12} md={6} lg={6}>
                <Paper className={fixedHeightPaperPie}>
                  <DisplayPie />
                </Paper>
              </Grid>
              <Grid item xs={12} md={12} lg={12}>
                <Paper className={classes.paper}>
                  <DisplayTable />
                </Paper>
              </Grid>
            </Grid>
          </Container>
        </div>
      </Route>
    </Switch>
  );
}
