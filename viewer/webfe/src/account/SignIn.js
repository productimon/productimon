import React from 'react';
import Avatar from '@material-ui/core/Avatar';
import Button from '@material-ui/core/Button';
import TextField from '@material-ui/core/TextField';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import Checkbox from '@material-ui/core/Checkbox';
import Link from '@material-ui/core/Link';
import Grid from '@material-ui/core/Grid';
import LockOutlinedIcon from '@material-ui/icons/LockOutlined';
import Typography from '@material-ui/core/Typography';
import { makeStyles } from '@material-ui/core/styles';
import Container from '@material-ui/core/Container';
import {
  BrowserRouter as Router,
  Switch,
  Route,
  Link as RouterLink,
  useHistory,
} from 'react-router-dom';

import { grpc } from '@improbable-eng/grpc-web';
import { DataAggregatorLoginRequest } from 'productimon/proto/svc/aggregator_pb';
import { DataAggregator } from 'productimon/proto/svc/aggregator_pb_service';

import ReactDOM from 'react-dom';
import TopMenu from '../core/TopMenu';
import SignUp from './SignUp';

const useStyles = makeStyles((theme) => ({
  paper: {
    marginTop: theme.spacing(8),
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
  },
  avatar: {
    margin: theme.spacing(1),
    backgroundColor: theme.palette.secondary.main,
  },
  form: {
    width: '100%', // Fix IE 11 issue.
    marginTop: theme.spacing(1),
  },
  submit: {
    margin: theme.spacing(3, 0, 2),
  },
}));

export default function SignIn() {
  const classes = useStyles();

  const [username, setEmail] = React.useState('');
  const [password, setPassword] = React.useState('');

  const handleChange = function (e, setter) {
    setter(e.target.value);
  };

  const history = useHistory();

  const doLogin = function (e) {
    e.preventDefault();

    const request = new DataAggregatorLoginRequest();
    request.setEmail(username);
    request.setPassword(password);
    grpc.unary(DataAggregator.Login, {
      host: '/rpc',
      onEnd: ({ status, statusMessage, headers, message }) => {
        if (status != 0) {
          alert(statusMessage);
          console.error('response ', status, statusMessage, headers, message);
          return;
        }
        window.localStorage.setItem('token', message.getToken());
        history.push('/dashboard');
      },
      request,
    });
  };
  return (
    <Container component="main" maxWidth="xs">
      <TopMenu />

      <div className={classes.paper}>
        <Avatar className={classes.avatar}>
          <LockOutlinedIcon />
        </Avatar>
        <Typography component="h1" variant="h5">
          Sign in
        </Typography>

        <form className={classes.form} onSubmit={doLogin}>
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
          <FormControlLabel
            control={<Checkbox value="remember" color="primary" />}
            label="Remember me"
          />
          <Button
            type="submit"
            fullWidth
            variant="contained"
            color="primary"
            className={classes.submit}
          >
            Sign In
          </Button>

          <Grid container>
            <Grid item xs>
              <Link href="#">Forgot password?</Link>
            </Grid>
            <Grid item>
              <RouterLink to="/signup" style={{ textDecoration: 'none' }}>
                Don't have an account? Sign Up
              </RouterLink>
            </Grid>
          </Grid>
        </form>
      </div>
    </Container>
  );
}
