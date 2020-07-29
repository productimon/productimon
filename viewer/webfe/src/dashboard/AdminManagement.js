import React from "react";

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
function createData(id, userEmail) {
  return { id, userEmail };
}


// dummy data for table
const rows = [
  createData(1, "a@x.com"),
  createData(0, "user@gmail.com"),
  createData(2, "admin@productimon.com"),
  createData(3, "basicAdmin@productimon.com"),
];

// Creating a list of fields to be used in the table
const cols = [
  { title: "User Email", field: "userEmail", editable: "never" },
];

export default function AdminManagement() {
  const { useState } = React;
  const classes = useStyles();
  const [columns, setColumns] = useState(cols);

  // TODO get data from backend

  const [data, setData] = useState(rows);

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
                      id="emailAddr"
                      label="Email"
                      variant="outlined"
                      fullWidth
                      margin="normal"
                    />
                  </Grid>
                  <Grid item style={{ marginRight: 12 }}>
                    <Button variant="contained">Promote</Button>
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
                      setTimeout(() => {
                        const dataDelete = [...data];
                        const index = oldData.tableData.id;
                        dataDelete.splice(index, 1);
                        setData([...dataDelete]);

                        // TODO demote user in backend

                        resolve();
                      }, 1000);
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
