import React, { useEffect } from "react";
import { Link as RouterLink, useHistory } from "react-router-dom";
import { useSnackbar } from "notistack";

import Avatar from "@material-ui/core/Avatar";
import Button from "@material-ui/core/Button";
import TextField from "@material-ui/core/TextField";
import Link from "@material-ui/core/Link";
import Grid from "@material-ui/core/Grid";
import LockOutlinedIcon from "@material-ui/icons/LockOutlined";
import Typography from "@material-ui/core/Typography";
import Container from "@material-ui/core/Container";

import { User } from "productimon/proto/common/common_pb";
import { DataAggregatorSignupRequest } from "productimon/proto/svc/aggregator_pb";
import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";

import { formUseStyles } from "./SignIn";
import { rpc } from "../Utils";

export default function SignUp(props) {
  const classes = formUseStyles();
  const { enqueueSnackbar, closeSnackbar } = useSnackbar();

  const [username, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");

  const handleChange = function (e, setter) {
    setter(e.target.value);
  };

  const history = useHistory();

  // TODO do in using a customised route with redirection
  // in App.js
  if (window.localStorage.getItem("token")) {
    useEffect(() => history.push("/dashboard"), []);
  } else if (!props.userDetails) {
    useEffect(() => {
      props.setUserDetails(null);
    }, []);
  }

  const doSignup = function (e) {
    e.preventDefault();

    const request = new DataAggregatorSignupRequest();
    const user = new User();
    user.setEmail(username);
    user.setPassword(password);
    request.setUser(user);
    rpc(DataAggregator.Signup, request)
      .then((res) => {
        window.localStorage.setItem("token", res.getToken());
        props.setUserDetails(res.getUser());
        enqueueSnackbar("Signed up successfully", { variant: "success" });
        history.push("/dashboard");
      })
      .catch((err) => {
        enqueueSnackbar(err, { variant: "error" });
      });
  };

  return (
    <Container className={classes.paper} maxWidth="xs">
      <Avatar className={classes.avatar}>
        <LockOutlinedIcon />
      </Avatar>
      <Typography component="h1" variant="h5">
        Sign up
      </Typography>

      <form className={classes.form} onSubmit={doSignup}>
        <TextField
          variant="outlined"
          margin="normal"
          required
          fullWidth
          label="Email Address"
          autoFocus
          onChange={(e) => handleChange(e, setEmail)}
        />
        <TextField
          variant="outlined"
          margin="normal"
          required
          fullWidth
          label="Password"
          type="password"
          onChange={(e) => handleChange(e, setPassword)}
        />
        <Button
          type="submit"
          fullWidth
          variant="contained"
          color="primary"
          className={classes.submit}
        >
          Sign Up
        </Button>

        <Grid container justify="flex-end">
          <Grid item>
            <RouterLink to="/" style={{ textDecoration: "none" }}>
              Already have an account? Sign in
            </RouterLink>
          </Grid>
        </Grid>
      </form>
    </Container>
  );
}
