import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';
import Title from './Title';

const useStyles = makeStyles({
  table: {
    minWidth: 650,
  },
});

function createData(platform, program, hours, label) {
  return { platform, program, hours, label };
}

const rows = [
  createData('Andriod', 'Facebook', 9, 'Social Media'),
  createData('Andriod', 'Spotify', 20, 'Entertainment'),
  createData('Linux', 'Terminal', 45, 'Productivity'),
  createData('Windows', 'Steam', 85, 'Entertainment'),
  createData('Windows', 'Adobe Illustrator', 23, 'Productivity'),
  createData('Windows', 'Microsoft Excel', 12, 'Productivity'),
];

export default function DisplayTable() {
  const classes = useStyles();

  return (
    <React.Fragment>
    <Title>Total Time Data</Title>
    <TableContainer component={Paper}>
      <Table className={classes.table} aria-label="simple table">
        <TableHead>
          <TableRow>
            <TableCell align="right">Platform</TableCell>
            <TableCell align="right">Program Name</TableCell>
            <TableCell align="right">Hours Spent</TableCell>
            <TableCell align="right">Labels</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {rows.map((row) => (
            <TableRow>
              <TableCell align="right">{row.platform}</TableCell>
              <TableCell align="right">{row.program}</TableCell>
              <TableCell align="right">{row.hours}</TableCell>
              <TableCell align="right">{row.label}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
    </React.Fragment>
  );
}

