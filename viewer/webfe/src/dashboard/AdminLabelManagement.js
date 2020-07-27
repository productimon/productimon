import React from 'react';

import MaterialTable from "material-table";
import TableContainer from "@material-ui/core/TableContainer";
import Paper from "@material-ui/core/Paper";

import Search from '@material-ui/icons/Search';
import ChevronLeft from '@material-ui/icons/ChevronLeft';
import ChevronRight from '@material-ui/icons/ChevronRight';
import FirstPage from '@material-ui/icons/FirstPage';
import LastPage from '@material-ui/icons/LastPage';
import Check from '@material-ui/icons/Check';
import Remove from '@material-ui/icons/Remove';
import EditIcon from '@material-ui/icons/Edit';
import ClearIcon from '@material-ui/icons/Clear';
import BackspaceIcon from '@material-ui/icons/Backspace';
import ArrowUpwardIcon from '@material-ui/icons/ArrowUpward';


function createData(id, program, numUsers, label) {
  return { id, program, numUsers, label };
}

const rows = [
  createData(1, "Safari", 2, "Web Browser"),
  createData(2, "Discord", 10, "VOiP"),
  createData(3, "Play Store", 11, "Online Shopping"),
  createData(4, "Brave", 2, "Web Browser"),
  createData(5, "Internet Explorer", 10, "Web Browser"),
  createData(6, "Facebook", 11, "Social Media")
];

export default function AdminLabelManagement() {

  const { useState } = React;

  const [columns, setColumns] = useState([
    { title: 'Program', field: 'program', editable: 'never' },
    { title: '#Users', field: 'numUsers', type: 'numeric', editable: 'never' },
    { title: 'Label', field: 'label' },
  ]);

  // TODO get data from backend
  const [data, setData] = useState(rows);

  return (
    <TableContainer component={Paper}>
      <MaterialTable
        title="Server Labels"
        columns={columns}
        data={data}

        localization={{
          header: {
            actions: 'Edit Label'
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
          ResetSearch : BackspaceIcon,
          SortArrow : ArrowUpwardIcon,
        }}

        options={{
          actionsColumnIndex: -1,
          pageSizeOptions: [10, 20, 50],
          pageSize: 10,
        }}

        editable={{
          onRowUpdate: (newData, oldData) =>
            new Promise((resolve, reject) => {
              setTimeout(() => {

                if (newData.label == "") newData.label = "Unknown";

                const dataUpdate = [...data];
                const index = oldData.tableData.id;
                dataUpdate[index] = newData;
                setData([...dataUpdate]);

                // TODO send updated row data to backend

                resolve();
              }, 1000)
            }),
        }}
      />
    </TableContainer>
  );
}
