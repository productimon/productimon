import React, { useEffect } from "react";
import { useHistory } from "react-router-dom";

import { makeStyles } from "@material-ui/core/styles";
import Button from "@material-ui/core/Button";
import Container from "@material-ui/core/Container";
import Grid from "@material-ui/core/Grid";
import MaterialTable from "material-table";
import Paper from "@material-ui/core/Paper";
import TableContainer from "@material-ui/core/TableContainer";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";

import ArrowUpwardIcon from "@material-ui/icons/ArrowUpward";
import BackspaceIcon from "@material-ui/icons/Backspace";
import Check from "@material-ui/icons/Check";
import ChevronLeft from "@material-ui/icons/ChevronLeft";
import ChevronRight from "@material-ui/icons/ChevronRight";
import ClearIcon from "@material-ui/icons/Clear";
import EditIcon from "@material-ui/icons/Edit";
import FirstPage from "@material-ui/icons/FirstPage";
import KeyboardArrowDownIcon from "@material-ui/icons/KeyboardArrowDown";
import LastPage from "@material-ui/icons/LastPage";
import Remove from "@material-ui/icons/Remove";
import Search from "@material-ui/icons/Search";

import { rpc } from "../Utils";
import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { User, Empty } from "productimon/proto/common/common_pb";

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
  fullWidth: {
    width: "100%",
  },
}));

// helper function to create dummy data
function createData(email) {
  return { email };
}

// Creating a list of fields to be used in the table
const columns = [{ title: "User Email", field: "email", editable: "never" }];

export default function AdminManagement() {
  const { useState } = React;
  const classes = useStyles();
  const history = useHistory();

  const [email, setEmail] = useState("");
  const [data, setData] = useState([]);
  useEffect(() => {
    const request = new Empty();
    rpc(DataAggregator.ListAdmins, history, {
      onEnd: ({ status, statusMessage, headers, message }) => {
        setData(message.getAdminsList().map((a) => createData(a.getEmail())));
      },
      request,
    });
  }, []);

  const promoteAdmin = () => {
    const request = new User();
    request.setEmail(email);
    rpc(DataAggregator.PromoteAccount, history, {
      onEnd: ({ status, statusMessage, headers, message }) => {
        setData([...data, createData(email)]);
        setEmail("");
      },
      request,
    });
  };

  return (
    <Container maxWidth="lg" className={classes.container}>
      <div className={classes.root}>
        <Grid
          container
          spacing={3}
          justify="center"
          alignItems="center"
          direction="column"
        >
          <Grid item className={classes.fullWidth}>
            <Paper className={classes.paper}>
              <Grid container spacing={3} direction="column">
                <Grid item>
                  <Typography variant="h5">Admin Promotion</Typography>
                </Grid>
                <Grid
                  container
                  justify="center"
                  alignItems="center"
                  spacing={3}
                >
                  <Grid item style={{ flexGrow: 1, marginLeft: 12 }}>
                    <TextField
                      label="Email"
                      variant="outlined"
                      fullWidth
                      margin="normal"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                    />
                  </Grid>
                  <Grid item style={{ marginRight: 12 }}>
                    <Button onClick={promoteAdmin} variant="contained">
                      Promote
                    </Button>
                  </Grid>
                </Grid>
              </Grid>
            </Paper>
          </Grid>
          <Grid item className={classes.fullWidth}>
            <TableContainer component={Paper}>
              <MaterialTable
                title="Server Admins"
                columns={columns}
                data={data}
                localization={{
                  header: {
                    actions: "Demote Admin",
                  },
                  body: {
                    editRow: {
                      deleteText: "Are you sure you want to demote this admin?",
                    },
                    emptyDataSourceMessage: "No current Admins",
                  },
                }}
                icons={{
                  Check: Check,
                  DetailPanel: ChevronRight,
                  FirstPage: FirstPage,
                  LastPage: LastPage,
                  NextPage: ChevronRight,
                  PreviousPage: ChevronLeft,
                  Search: Search,
                  ThirdStateCheck: Remove,
                  Edit: EditIcon,
                  Clear: ClearIcon,
                  ResetSearch: BackspaceIcon,
                  SortArrow: ArrowUpwardIcon,
                  Delete: KeyboardArrowDownIcon,
                }}
                options={{
                  actionsColumnIndex: -1,
                  actionsCellStyle: {},
                }}
                editable={{
                  onRowDelete: (oldData) =>
                    new Promise((resolve, reject) => {
                      const request = new User();
                      request.setEmail(oldData.email);
                      rpc(DataAggregator.DemoteAccount, history, {
                        onEnd: ({
                          status,
                          statusMessage,
                          headers,
                          message,
                        }) => {
                          const updatedData = data.filter(
                            (a) => a.email != oldData.email
                          );
                          setData(updatedData);
                          resolve();
                        },
                        request,
                      });
                    }),
                }}
              />
            </TableContainer>
          </Grid>
        </Grid>
      </div>
    </Container>
  );
}
