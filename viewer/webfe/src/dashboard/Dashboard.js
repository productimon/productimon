import React from 'react';
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
import Link from '@material-ui/core/Link';
import MenuIcon from '@material-ui/icons/Menu';
import ChevronLeftIcon from '@material-ui/icons/ChevronLeft';
import NotificationsIcon from '@material-ui/icons/Notifications';
//import { mainListItems, secondaryListItems } from './listItems';
import Chart from './Chart';

import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import DashboardIcon from '@material-ui/icons/Dashboard';

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
  fixedHeight: {
    height: 240,
  },
}));

export default function Dashboard() {

  // state stores what page the main section displays is in, by default we start with dashboard
  const [state, setState] = React.useState('dashboard')

  const classes = useStyles();
  const [open, setOpen] = React.useState(true);
  const handleDrawerOpen = () => {
    setOpen(true);
  };
  const handleDrawerClose = () => {
    setOpen(false);
  };

  return (
    <div className={classes.root}>
      <CssBaseline />
        <AppBar position="absolute" className={clsx(classes.appBar, open && classes.appBarShift)} style={{backgroundColor: "brown"}}>
          <Toolbar className={classes.toolbar}>
            <IconButton
              edge="start"
              color="inherit"
              aria-label="open drawer"
              onClick={handleDrawerOpen}
              className={clsx(classes.menuButton, open && classes.menuButtonHidden)}
             >
           <MenuIcon />
           </IconButton>

           <Typography component="h1" variant="h6" color="inherit" noWrap className={classes.title}>
             Productimon
           </Typography>

           <IconButton color="inherit">
             <Badge badgeContent={4} color="secondary">
               <NotificationsIcon />
             </Badge>
           </IconButton>
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
            <ListItem button onClick={() => setState('dashboard')}>
               <ListItemIcon>
                  <DashboardIcon />
               </ListItemIcon>
               <ListItemText primary="Dashboard" />
            </ListItem>
            <ListItem button onClick={() => setState('goal')}>
               <ListItemIcon>
                 <DashboardIcon />
               </ListItemIcon>
               <ListItemText primary="Goals"/>
            </ListItem>
            <ListItem button onClick={() => setState('histogram')}>
               <ListItemIcon>
                 <DashboardIcon />
               </ListItemIcon>
               <ListItemText primary="Histogram" />
            </ListItem>  
          </List>
        <Divider />
      </Drawer>

      <main className={classes.content}>
        <div className={classes.appBarSpacer} />

        <Display page={state}/>

      </main>
    </div>
  );
}

// Based on what page we select on the list (props.page) display specific content
// I dont think this is how react is supposed to be used, it seems messy, but it'll do for now
function Display(props) {
     const classes = useStyles();
     const state = props.page;
     const fixedHeightPaper = clsx(classes.paper, classes.fixedHeight);
     if (state=="histogram") {
         return (<h1>Histogram</h1>);
     } else if (state=="dashboard") {
         return (
         <div>
<<<<<<< HEAD
=======
         <h1>Dashboard</h1>
>>>>>>> [PROD-104]viewer:init dashboard
         <Container maxWidth="lg" className={classes.container}>
          <Grid container spacing={3}>
            {/* Chart */}
            <Grid item xs={12} md={8} lg={9}>
              <Paper className={fixedHeightPaper}>
                <Chart />
              </Paper>
            </Grid>
          </Grid>
        </Container>
        </div>
        );
     } else if (state=="goal") {
         return (<h1>Goal</h1>);
     }
}
