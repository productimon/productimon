import React from "react";

import { makeStyles } from "@material-ui/core/styles";

import Container from "@material-ui/core/Container";
import Grid from "@material-ui/core/Grid";
import Paper from "@material-ui/core/Paper";
import Typography from "@material-ui/core/Typography";

import DevicesTwoToneIcon from "@material-ui/icons/DevicesTwoTone";
import PeopleAltIcon from "@material-ui/icons/PeopleAlt";

const useStyles = makeStyles((theme) => ({
  root: {
    flexGrow: 1,
  },
  paper: {
    padding: theme.spacing(2),
    margin: "auto",
  },
  container: {
    paddingTop: theme.spacing(4),
    paddingBottom: theme.spacing(4),
    height: "90%",
  },
  halfWidth: {
    width: "50%",
  },
}));

export default function AdminServerStatus() {
  const classes = useStyles();

  return (
    <Container maxWidth="lg" className={classes.container}>
      <div className={classes.root}>
        <Typography variant="h4" gutterBottom>
          Server Status
        </Typography>
        <Grid
          container
          spacing={3}
          justify="center"
          alignItems="center"
          direction="row"
        >
          <Grid item className={classes.halfWidth}>
            <Paper className={classes.paper}>
              <Grid
                container
                spacing={3}
                direction="column"
                alignItems="center"
                justify="center"
              >
                <Grid item>
                  <Typography variant="h5">Devices Connected</Typography>
                </Grid>
                <Grid
                  container
                  spacing={3}
                  direction="row"
                  alignItems="center"
                  justify="center"
                >
                  <Grid item>
                    <DevicesTwoToneIcon style={{ color: "#1E90FF" }} />
                  </Grid>
                  <Grid item>
                    {/* TODO insert data from backend */}
                    <Typography variant="h5">42</Typography>
                  </Grid>
                </Grid>
              </Grid>
            </Paper>
          </Grid>
          <Grid item className={classes.halfWidth}>
            <Paper className={classes.paper}>
              <Grid
                container
                spacing={3}
                direction="column"
                alignItems="center"
                justify="center"
              >
                <Grid item>
                  <Typography variant="h5">Users Online</Typography>
                </Grid>
                <Grid
                  container
                  spacing={3}
                  direction="row"
                  alignItems="center"
                  justify="center"
                >
                  <Grid item>
                    <PeopleAltIcon />
                  </Grid>
                  <Grid item>
                    {/* TODO insert data from backend */}
                    <Typography variant="h5">42</Typography>
                  </Grid>
                </Grid>
              </Grid>
            </Paper>
          </Grid>
        </Grid>
      </div>
    </Container>
  );
}
