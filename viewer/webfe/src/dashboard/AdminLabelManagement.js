import React, { useEffect, useState } from "react";
import { useHistory } from "react-router-dom";

import MaterialTable from "material-table";
import TableContainer from "@material-ui/core/TableContainer";
import Paper from "@material-ui/core/Paper";

import Search from "@material-ui/icons/Search";
import ChevronLeft from "@material-ui/icons/ChevronLeft";
import ChevronRight from "@material-ui/icons/ChevronRight";
import FirstPage from "@material-ui/icons/FirstPage";
import LastPage from "@material-ui/icons/LastPage";
import Check from "@material-ui/icons/Check";
import Remove from "@material-ui/icons/Remove";
import EditIcon from "@material-ui/icons/Edit";
import ClearIcon from "@material-ui/icons/Clear";
import BackspaceIcon from "@material-ui/icons/Backspace";
import ArrowUpwardIcon from "@material-ui/icons/ArrowUpward";

import { rpc } from "../Utils";

import {
  DataAggregatorGetLabelsRequest,
  DataAggregatorUpdateLabelRequest,
} from "productimon/proto/svc/aggregator_pb";
import { DataAggregator } from "productimon/proto/svc/aggregator_pb_service";
import { Label } from "productimon/proto/common/common_pb";

function createData(program, label, numUsers) {
  return { program, label, numUsers };
}

export default function AdminLabelManagement(props) {
  const [data, setData] = useState([]);
  const history = useHistory();

  let columns = [
    { title: "Program", field: "program", editable: "never" },
    { title: "Label", field: "label" },
  ];
  if (props.admin)
    columns.push({
      title: "#Users",
      field: "numUsers",
      type: "numeric",
      editable: "never",
    });

  useEffect(() => {
    const request = new DataAggregatorGetLabelsRequest();
    request.setAllLabels(props.admin);

    rpc(DataAggregator.GetLabels, history, {
      onEnd: ({ status, statusMessage, headers, message }) => {
        if (status != 0) {
          console.error(statusMessage);
          return;
        }
        setData(
          message
            .getLabelsList()
            .map((l) => createData(l.getApp(), l.getLabel(), l.getUsedBy()))
        );
      },
      request,
    });
  }, [props.admin]);

  return (
    <TableContainer component={Paper}>
      <MaterialTable
        title={props.admin ? "Server Labels" : "My Labels"}
        columns={columns}
        data={data}
        localization={{
          header: {
            actions: "Edit Label",
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
        }}
        options={{
          actionsColumnIndex: -1,
          pageSizeOptions: [10, 20, 50],
          pageSize: 10,
        }}
        editable={{
          onRowUpdate: (newData, oldData) => {
            if (newData.label == "") newData.label = "Unknown";
            const label = new Label();
            label.setApp(newData.program);
            label.setLabel(newData.label);
            const request = new DataAggregatorUpdateLabelRequest();
            request.setAllLabels(props.admin);
            request.setLabel(label);
            return new Promise((resolve, reject) => {
              rpc(DataAggregator.UpdateLabel, history, {
                onEnd: ({ status, statusMessage, headers, message }) => {
                  if (status != 0) {
                    reject();
                    return;
                  }
                  setData(
                    data.map((d) =>
                      d.program == newData.program ? newData : d
                    )
                  );
                  resolve();
                },
                request,
              });
            });
          },
        }}
      />
    </TableContainer>
  );
}
